---
Title: Workflow runtime hardening analysis, design, and implementation guide
Ticket: WORKFLOW-RUNTIME-HARDENING-001
Status: active
Topics:
    - workflow
    - scraper
    - book-ocr
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-golems/scraper/pkg/engine/scheduler/scheduler.go
      Note: RunOnce cycle; heartbeat goroutine wraps executeLeasedOp
    - Path: ../../../../../../../go-go-golems/scraper/pkg/engine/store/sqlite/result_store.go
      Note: CompleteOp/FailOp — fencing target
    - Path: ../../../../../../../go-go-golems/scraper/pkg/services/engineview/workflow_mutation_service.go
      Note: Home of RetryOp/CancelWorkflow; RequeueSteps lands here
    - Path: cmd/book-ocr/main.go
      Note: requeueStructuredPages SQL this ticket retires; isTerminal enum coupling
    - Path: internal/ocrpipeline/workflow_retry_test.go
      Note: Test pattern to extend for rerun/heartbeat scenarios
ExternalSources: []
Summary: Intern-ready guide to the scraper workflow runtime (engine schema, scheduler, leases, operator services) and the four hardening items it needs — a first-class RequeueSteps operator API, lease heartbeats, cooperative cancellation, and optional module extraction — plus the interim schema-version guard on the book-ocr side.
LastUpdated: 2026-07-03T11:52:26.165031243-07:00
WhatFor: Onboard an engineer onto the workflow runtime internals and give them an executable plan for the four cross-repo hardening items.
WhenToUse: Read before changing scraper/pkg/workflow or the engine, or before touching book-ocr's structured-rerun-pages.
---


# Workflow runtime hardening analysis, design, and implementation guide

## Executive Summary

book-ocr runs on a durable workflow runtime that lives in the scraper repository (`github.com/go-go-golems/scraper`, consumed at `v0.0.4` since 2026-07-03). The runtime is sound — it carried a 202-page OCR book through crashes, retries, and resumes — but four gaps remain, all known since the original design ticket (SCRAPER-JOBS-001) and all requiring changes in the **scraper repo**, then a release, then a version bump in book-ocr:

1. **`RequeueSteps` operator API (highest priority).** book-ocr's targeted page-repair command (`structured-rerun-pages`) currently performs raw SQL surgery on the engine's private schema. Now that book-ocr consumes *published* scraper versions, a schema change arrives silently with a version bump and can break rerun mid-repair. The engine should own step requeueing the way it already owns `RetryOp` and `CancelWorkflow`.
2. **Lease heartbeats.** A step that runs longer than its lease gets recovered and re-leased while still executing — duplicate execution, duplicate model spend. Long VLM calls make this a live risk, not a theoretical one.
3. **Cooperative cancellation.** `CancelRun` flips database state but does not interrupt in-flight executors; canceling a live OCR run lets running model calls finish and burn tokens.
4. **Module extraction (optional).** `pkg/workflow` + `pkg/engine` are generic but live inside scraper, dragging its dependency tree (grpc, envoy control-plane, goja modules) into every consumer.

There is also one **book-ocr-side interim mitigation** that does not wait for any release: a schema-version guard that makes `structured-rerun-pages` refuse to touch an `engine.db` whose migrations it does not recognize.

This document explains the runtime internals an implementer needs (Part I), then specifies each item with API sketches, pseudocode, and file-level guidance (Part II), ordered RequeueSteps → heartbeats → cancellation → extraction.

## Problem Statement and Scope

The four items above were identified during the BOOK-OCR-PRODUCT-001 productization analysis (findings F5 and the "runtime hardening requests upstream" of Phase 4; see that ticket's design doc 01, Part IV and Part VII). They are parked here because they are cross-repo: the code that must change is in `~/code/wesen/go-go-golems/scraper` (read-only from the book-ocr workspace), and each lands via a scraper release consumed by a version bump in book-ocr.

In scope: the design and implementation guide for all four items plus the interim guard. Out of scope: performing the scraper-side changes from the book-ocr workspace (repo boundary), and any Postgres/River backend work.

## Part I — The Runtime, Explained for an Implementer

### What it is

`scraper/pkg/workflow` (~1,830 lines) is an embeddable durable-execution facade over `scraper/pkg/engine/{model,runner,scheduler,store}`. An application (book-ocr) registers **packages** (named workflow definitions) and **executors** (business logic per step kind), then starts **runs**. All state — the step graph, dependencies, retry counters, leases, results, artifacts — lives in one SQLite file per work directory (`engine.db`). Workers are stateless: any process that opens the same `engine.db` and calls `RunOnce` participates safely, coordinated purely through the database.

```text
            ┌────────────────────────────────────────────────────────────┐
            │ Runtime (pkg/workflow/runtime.go:80-124)                   │
            │   RegisterPackage / RegisterExecutor / StartRun / RunOnce  │
            │   RetryStep / CancelRun          (operators.go:21-36)      │
            └───────┬────────────────────────────────────┬───────────────┘
                    ▼                                    ▼
   ┌──────────────────────────────┐      ┌────────────────────────────────┐
   │ scheduler (engine/scheduler) │      │ engineview.Service             │
   │  RunOnce cycle:              │      │  RetryOp / CancelWorkflow      │
   │  1 recover expired leases    │      │  (workflow_mutation_service.go)│
   │  2 cancel-propagate failures │      └────────────────────────────────┘
   │  3 promote pending→ready     │                    │
   │  4 lease + execute           │                    ▼
   └──────────────┬───────────────┘      ┌────────────────────────────────┐
                  ▼                      │ engine.db (SQLite)             │
   ┌──────────────────────────────┐      │  workflows · ops ·             │
   │ executors (app code, e.g.    │─────▶│  op_dependencies · leases ·    │
   │ book-ocr StructuredPage...)  │      │  queue_limit_state · results · │
   └──────────────────────────────┘      │  artifacts · schema_migrations │
                                         └────────────────────────────────┘
```

### The tables that matter here

From migrations `001_engine_core.sql` / `002_engine_runtime.sql` (`scraper/pkg/engine/store/sqlite/migrations/`):

- **`ops`** — one row per step: `status` (pending/ready/running/succeeded/failed/canceled), `retry_json` (the policy), `retry_state_json` (live attempt counter + last error), `next_attempt_at`, `input_json`, `queue_key`, `dedup_key`.
- **`op_dependencies`** — `(op_id, depends_on_op_id, required)`; `required=1` failures cancel dependents, optional deps merely delay readiness.
- **`leases`** — `(op_id PK, worker_id, token, acquired_at, expires_at)`. A lease is the *only* thing preventing two workers from executing the same op.
- **`results`** — per-op output data, projection record writes, emitted child specs, error JSON.
- **`schema_migrations`** — `(version PK, name, applied_at)`; the anchor for the interim guard.

### The scheduler cycle (where every gap lives)

`Scheduler.RunOnce` (`scraper/pkg/engine/scheduler/scheduler.go:197-294`) does, per cycle:

```text
RunOnce(now):
  RefreshRunnableOps(now)               # op_store.go:104-210
    a. leases with expires_at <= now  →  delete lease, op running→ready   ← GAP: heartbeats
    b. pending ops with a failed/canceled required dep → canceled (fixpoint)
    c. pending ops whose deps are done → ready
  for (site, queue) in ListQueueCandidates(now):
    lease up to min(queue.MaxInFlight, free slots) ready ops   # lease_store.go
    for each leased op: executeLeasedOp(op)                    # synchronous
      run executor → OpResult
      success: CompleteOp (results + artifacts + emitted + status, one tx)
      failure: nextRetryState (scheduler.go:472-492)
               retryable & attempts<max → status ready, next_attempt_at = backoff
               else → status failed
  refreshWorkflowStatus(touched workflows)
```

Three properties an implementer must internalize:

1. **Lease expiry is the crash-recovery mechanism** (step a). It cannot distinguish "worker died" from "executor is slow" — that is precisely the heartbeat gap.
2. **Execution is synchronous inside the cycle**; the executor holds the lease but nothing renews it.
3. **All mutations are idempotent SQL against `engine.db`** — which is why book-ocr's rerun-by-SQL *works*, and why it is fragile: it re-implements step-state invariants (b/c above) by hand outside the engine.

### The operator surface today

`Runtime.RetryStep` / `Runtime.CancelRun` (`pkg/workflow/operators.go:21-36`) delegate to `engineview.Service` (`scraper/pkg/services/engineview/workflow_mutation_service.go`):

- **`RetryOp`** (`:11-38`): failed op → ready, clear `retry_state_json` + `next_attempt_at`, workflow → running. One op, must currently be `failed`.
- **`CancelWorkflow`** (`:40-78`): ops in (pending, ready, running) → canceled, leases deleted, workflow → canceled unless succeeded. **No signal reaches a running executor** — the op row says canceled while the process keeps executing (acknowledged in `operators.go:29-35`).

### What book-ocr does instead (the code this ticket retires)

`requeueStructuredPages` (`book-ocr cmd/book-ocr/main.go:590-665`), invoked by `structured-rerun-pages`:

1. `UPDATE workflows SET status='running'` for the run.
2. For each selected page: `DELETE FROM leases WHERE op_id=?`, then `UPDATE ops SET status='ready', retry_state_json='{"attempt":0}', next_attempt_at=NULL`.
3. Downstream assemble/validate ops → `status='pending'` (not ready — readiness must be re-derived after the pages finish; comment at `main.go:627-629`).
4. If `--render-pdf`: mutate the assemble op's `input_json` in place to set `render_pdf`/`pdf_path`/`pandoc_path` (`main.go:634-658`).
5. Resume workers.

Every line of that is scheduler business executed from outside the scheduler, including a **hardcoded `retry_state_json` literal** that assumes the engine's current serialization. This worked while book-ocr built against the sibling checkout (schema drift was visible locally); with published-version consumption it is a silent-breakage channel.

## Part II — The Four Items

### Item 1: `RequeueSteps` operator API (scraper-side; highest priority)

**Requirement.** An engine-owned, transactional operation: "reset these steps of this run so they execute again, and put everything downstream of them back into dependency-derived readiness."

**Proposed API** (in `engineview.Service`, exposed through `pkg/workflow`):

```go
// pkg/workflow/operators.go
type RequeueOptions struct {
    // ResetDownstream: ops that transitively depend on the requeued steps are
    // set to pending (readiness re-derived by the scheduler). Default true.
    ResetDownstream bool
    // InputPatch optionally shallow-merges keys into a step's input_json,
    // keyed by op ID. Replaces book-ocr's in-place JSON mutation for
    // toggling render_pdf on the assemble step.
    InputPatch map[model.OpID]map[string]any
}

func (rt *Runtime) RequeueSteps(runID model.WorkflowID,
    stepIDs []model.OpID, opts RequeueOptions) error
```

**Semantics (single transaction):**

```text
RequeueSteps(run, steps, opts):
  tx:
    validate: run exists; every step belongs to run;
              no step is currently leased with expires_at > now
              (refuse — a live executor may still write its result)
    for s in steps:
      DELETE FROM leases WHERE op_id = s          # stale leases only, see above
      ops[s]: status='ready', retry_state_json=engine-owned zero value,
              next_attempt_at=NULL
      apply opts.InputPatch[s] via json_patch on input_json (if present)
    if opts.ResetDownstream:
      downstream = transitive closure over op_dependencies where required=1
                   ∪ optional dependents            # both must wait again
      ops[downstream where status in (succeeded, failed, canceled)]:
              status='pending', retry_state cleared, results row deleted?  → NO:
              keep old results row until overwritten (CompleteOp uses
              INSERT OR REPLACE, result_store.go:14-102) — cheaper and preserves
              history until the new result lands
    UPDATE workflows SET status='running'
  emit operator event (observer hook, scheduler.go observer pattern)
```

Decisions embedded there, called out for review: refusing to requeue *actively leased* steps (book-ocr's version silently deletes live leases — a running page executor would then double-write); keeping stale `results` rows because `CompleteOp` replaces them atomically; and letting the engine own the `retry_state_json` zero value instead of clients hardcoding it.

**book-ocr adoption** (after a scraper release): `requeueStructuredPages` (`main.go:590-665`) collapses to

```go
err := rt.RequeueSteps(runID, pageStepIDs, workflow.RequeueOptions{
    ResetDownstream: true,
    InputPatch: map[model.OpID]map[string]any{
        assembleOpID: {"render_pdf": renderPDF, "pdf_path": pdfPath, "pandoc_path": pandocPath},
    },
})
```

**Tests (scraper-side):** requeue a mid-graph step → downstream flips to pending and re-derives readiness only after the requeued step succeeds; requeue with a live lease → error; InputPatch round-trip; requeue of an op in every terminal status; concurrent RunOnce during requeue (tx isolation).

### Item 2: Lease heartbeats (scraper-side)

**The failure, concretely.** `Config.LeaseDuration` (`pkg/workflow/runtime.go:60-71`) bounds how long a lease lives. `executeLeasedOp` runs the executor synchronously without renewal. A structured-page step whose model call takes longer than the lease duration is recovered by step (a) of the next `RunOnce` on any worker: lease deleted, op → ready, second worker leases and executes it *while the first is still running*. Both eventually `CompleteOp`; the second silently overwrites the first (`INSERT OR REPLACE`, `result_store.go:14-102`). Cost: duplicate model spend, racing artifact writes (benign today only because page artifacts are idempotent), confusing projections.

**Design: renew from the execution wrapper, verify at completion.**

```text
# scheduler-side, around the executor call
executeLeasedOp(op, lease):
  hbCtx, stop = context.WithCancel(ctx)
  go every LeaseDuration/3 until hbCtx done:
      ok = store.RenewLease(op.ID, lease.Token, now+LeaseDuration)
      if !ok: cancel executor ctx        # we lost the lease; stop work
  result = executor.Execute(ctx, step)
  stop()
  CompleteOp / FailOp ... WHERE lease token still ours   # fencing write
```

Two engine changes carry the weight:

1. `RenewLease(opID, token, newExpiry)` in the lease store — `UPDATE leases SET expires_at=? WHERE op_id=? AND token=?`; rows-affected 0 means the lease was recovered and the executor must abort.
2. **Fenced completion**: `CompleteOp`/`FailOp` gain the lease token in their WHERE clause, so a worker that lost its lease cannot commit results. (Today the lease is deleted by token at completion, `result_store.go` — the fencing is almost there; it must become a precondition, not a cleanup.)

The heartbeat goroutine cancels the executor's context on renewal failure — which is only effective once executors honor context cancellation, i.e. Item 3. The two items are independent to land but compound in value.

**Compatibility:** heartbeats are transparent to executors and to book-ocr; no API change, only behavior. Default `LeaseDuration` can then be *shortened* (faster crash recovery) since slow-but-alive executors are no longer at risk.

**Tests:** executor sleeping 3× lease duration completes exactly once with heartbeats on; kill -9 the worker mid-execution → lease expires (no heartbeat) → recovery still works; renewal failure cancels the executor context; fenced CompleteOp rejects a stale token.

### Item 3: Cooperative cancellation (scraper-side)

**The failure, concretely.** `book-ocr cancel` → `CancelWorkflow` marks rows canceled and deletes leases; in-flight `eng.RunInference` calls (up to `--max-workers` of them, each potentially minutes long) run to completion, spending tokens after the operator said stop. The UI/status also cannot distinguish "we asked" from "it stopped".

**Design: status poll + context cancel, two-phase status.**

1. New op/workflow status value `cancel_requested` (or a `cancel_requested_at` column — less enum churn, keeps status transitions simple; **preferred**).
2. `CancelWorkflow` sets `cancel_requested_at` on running ops instead of force-flipping them, still hard-cancels pending/ready ops immediately.
3. The heartbeat goroutine from Item 2 doubles as the poll: each renewal also reads `cancel_requested_at`; if set, cancel the executor's context.
4. Executor contract documented: honor `ctx.Done()`; geppetto's `RunInference(ctx, …)` already takes the context through to the provider HTTP call, so book-ocr page steps become cancelable for free.
5. When the executor returns after cancellation, `FailOp` records code `E_CANCELED` (non-retryable) and status → canceled; `refreshWorkflowStatus` completes the workflow-level transition. A grace timeout (e.g. 2× heartbeat interval) after which the operator can force-cancel covers executors that ignore contexts.

```text
CancelWorkflow(run):                      # revised
  tx: ops pending/ready → canceled, leases deleted (unchanged)
      ops running       → cancel_requested_at = now      # NEW: leases kept!
      workflow → canceling (derived)
heartbeat tick (per running op):
  renew lease; if cancel_requested_at set → cancel executor ctx
executor returns ctx.Canceled → FailOp(E_CANCELED) → op canceled
all ops terminal → workflow canceled
```

Note the inversion from today: running ops *keep* their leases during cancellation so exactly one worker owns the wind-down; today's lease deletion is what makes the zombie-execution window possible.

**Tests:** cancel with a context-honoring executor stops within one heartbeat interval; cancel with an ignoring executor force-cancels after grace; `status` surfaces canceling vs canceled; canceled steps are requeue-able via Item 1.

### Item 4: Module extraction (scraper-side; optional)

**Motivation.** `go mod tidy` in book-ocr pulls scraper's full tree — grpc, envoy go-control-plane, goja modules — for ~2,400 lines of actually-used runtime. A standalone `github.com/go-go-golems/workflow` module containing today's `pkg/workflow` + `pkg/engine` (and the `engineview` mutation service, which is runtime-generic) would cut consumer dependency weight and give the runtime its own release cadence, decoupled from scraper features.

**Shape:** move, don't fork — scraper then imports the new module too, so there is one implementation. Go module boilerplate aside, the work is mechanical (the BOOK-OCR-EXTERNALIZE-001 ticket did the mirror-image move for OCR code in a day). The go-go-golems project scaffolding skill/template handles repo setup, CI, and goreleaser.

**Sequencing:** last. Items 1–3 should land first *inside scraper* — extraction mid-flight would double the release coordination. If extraction happens, items 1–3 travel with the code.

### Item 5 (book-ocr-side, no release needed): schema-version guard for rerun

Until Item 1 ships, `structured-rerun-pages` keeps its SQL — but it must stop assuming the schema silently. The engine records applied migrations in `schema_migrations`; the guard pins what the SQL was written against:

```go
// cmd/book-ocr/main.go, before requeueStructuredPages touches anything
var knownEngineMigrations = []string{"001_engine_core", "002_engine_runtime"} // scraper v0.0.4

func guardEngineSchema(db *sql.DB) error {
    rows := query(db, `SELECT name FROM schema_migrations ORDER BY version`)
    if !equal(rows, knownEngineMigrations) {
        return errors.Errorf(
            "engine.db schema (%v) differs from the schema structured-rerun-pages "+
            "was written against (%v); refusing direct SQL requeue. "+
            "Update book-ocr or use a matching scraper version.",
            rows, knownEngineMigrations)
    }
    return nil
}
```

Fail-closed and loud: a schema drift becomes an actionable error before any mutation, instead of a corrupted run after. This lands in book-ocr immediately (tracked in BOOK-OCR-PRODUCT-001's implementation work) and is deleted together with the SQL when Item 1 ships.

## Decision Records

### Decision: Order of work — RequeueSteps, then heartbeats, then cancellation, then extraction

- **Context:** Four items, one owner, limited attention; items 2 and 3 share machinery.
- **Options considered:** (1) requeue → heartbeat → cancellation → extraction; (2) heartbeat first (money risk); (3) extraction first (clean-room for the rest).
- **Decision:** Option 1.
- **Rationale:** Requeue guards the *correctness of the operator feature book-ocr uses most*, and its risk grew categorically when book-ocr switched to published versions (drift is now invisible until it bites). Heartbeats' cost is bounded (duplicate spend, idempotent artifacts) and partially mitigated by generous lease durations. Cancellation builds on the heartbeat loop, so it naturally comes after. Extraction mid-stream doubles coordination for zero user-visible value.
- **Consequences:** book-ocr carries the Item-5 guard in the interim; heartbeat-dependent cancellation waits for Item 2.
- **Status:** proposed

### Decision: `cancel_requested_at` column instead of a new status enum value

- **Context:** Two-phase cancel needs to represent "asked but not yet stopped".
- **Options considered:** (1) new `OpStatus` value `cancel_requested`; (2) nullable `cancel_requested_at` timestamp alongside the existing status.
- **Decision:** Option 2.
- **Rationale:** Every status consumer (scheduler SQL in `RefreshRunnableOps`, projections, book-ocr's `isTerminal` at `main.go:988-997`, UI filters) switches on the enum; a new value is a breaking sweep. A timestamp is additive, self-documenting for "how long has cancel been pending", and the derived workflow-level "canceling" can be computed without storing a new state.
- **Consequences:** "Is this op being canceled?" is a two-column read; migration is one `ALTER TABLE ADD COLUMN` (backward compatible for older readers, which ignore it).
- **Status:** proposed

### Decision: Refuse to requeue actively-leased steps

- **Context:** book-ocr's current SQL deletes leases unconditionally; a still-running executor would later double-complete.
- **Options considered:** (1) refuse with an actionable error ("step structured-page-084 is running; cancel it or wait"); (2) delete the lease and rely on fenced completion (Item 2) to reject the stale worker.
- **Decision:** Option 1 now; revisit option 2 once fencing exists.
- **Rationale:** Without fencing, option 2 is exactly the double-write bug. Even with fencing, silently discarding an in-flight expensive computation is operator-hostile — an explicit error tells the operator what is actually happening.
- **Consequences:** Requeue of a hung-but-leased step requires waiting out the lease or canceling first; acceptable, and Item 3 makes the cancel path real.
- **Status:** proposed

## Implementation Plan

Scraper-side work happens in the scraper repo (not from the book-ocr workspace); each numbered step is releasable independently.

1. **scraper: `RequeueSteps`** — engineview mutation (`workflow_mutation_service.go`, new method beside `RetryOp:11-38`), facade method in `pkg/workflow/operators.go`, transitive-downstream query over `op_dependencies`, tests as specified in Item 1. Release (v0.0.5).
2. **book-ocr: adopt** — bump scraper, replace `requeueStructuredPages` body with the API call, delete the schema guard, keep the CLI surface identical.
3. **scraper: heartbeats** — `RenewLease` in the lease store, heartbeat goroutine in `executeLeasedOp` (`scheduler.go:197-294` region), fenced `CompleteOp`/`FailOp` (`result_store.go:14-102`), config knob `HeartbeatInterval` defaulting to `LeaseDuration/3`. Release.
4. **scraper: cooperative cancellation** — `cancel_requested_at` migration, revised `CancelWorkflow` (`workflow_mutation_service.go:40-78`), heartbeat-driven ctx cancel, `E_CANCELED` fail path, grace-timeout force cancel. Release. book-ocr: no code change; document that `cancel` now stops in-flight model calls.
5. **scraper (optional): extraction** to `go-go-golems/workflow`; scraper and book-ocr both re-point imports.

Interim, tracked in BOOK-OCR-PRODUCT-001 (implementation ticket): the Item-5 schema guard.

## Testing Strategy

- Engine-level tests live in scraper next to the mutated stores (`lease_store`, `result_store`, `workflow_mutation_service`): every scenario listed per item above.
- book-ocr integration proof after each adoption: the existing structured workflow retry test (`internal/ocrpipeline/workflow_retry_test.go`) pattern extended with (a) a rerun-after-success test through the new API and (b) a slow-executor test that only passes with heartbeats (executor sleep > lease duration, assert single completion).
- The schema guard gets a unit test with a doctored `schema_migrations` table (extra row → refusal with the exact error text).

## Risks and Open Questions

- **Release coupling:** each scraper release must be consumed promptly by book-ocr or drift accumulates again; mitigate by bumping in the same working session.
- **Heartbeat write volume:** one UPDATE per running op per interval; negligible for OCR-scale runs (≤ max-workers concurrent), worth a benchmark note for scraper's own high-fan-out crawls.
- **Cancellation depends on executor discipline:** the grace-timeout force path bounds the damage of context-ignoring executors, but the guarantee is only "best effort within grace".
- **Open:** should `RequeueSteps` support requeueing *succeeded* control steps (e.g. re-discover pages after adding scans)? Current design allows any terminal status; discovery re-emission semantics (duplicate `Emit` dedup via `dedup_key`) need a test before recommending it.
- **Open:** extraction naming — `go-go-golems/workflow` vs folding into an existing infra module; owner's call, no technical constraint.

## References

**scraper (read-only from here, `~/code/wesen/go-go-golems/scraper`):** `pkg/workflow/runtime.go:60-124` (Config, LeaseDuration), `pkg/workflow/operators.go:15-52` (operator facade + known cancellation caveat at `:29-35`), `pkg/engine/scheduler/scheduler.go:197-294` (RunOnce), `:472-529` (retry state/backoff), `pkg/engine/store/sqlite/op_store.go:104-210` (RefreshRunnableOps — recovery/cancel-propagation/promotion), `pkg/engine/store/sqlite/result_store.go:14-102` (CompleteOp; fencing target), `pkg/engine/store/sqlite/lease_store.go` (LeaseReadyOp; RenewLease home), `pkg/services/engineview/workflow_mutation_service.go:11-78` (RetryOp/CancelWorkflow; RequeueSteps home), `pkg/engine/store/sqlite/migrations/00{1,2}_*.sql` (schema).

**book-ocr:** `cmd/book-ocr/main.go:590-665` (requeueStructuredPages — the SQL this ticket retires), `:627-629` (downstream-pending rationale), `:988-997` (isTerminal enum coupling), `internal/ocrpipeline/workflow_retry_test.go` (test pattern to extend).

**History:** SCRAPER-JOBS-001 (original gap list: heartbeats, cooperative cancel, artifact storage), BOOK-OCR-STRUCTURED-WORKFLOW-001 (rerun operator introduction), BOOK-OCR-PRODUCT-001 design doc 01 (findings F1/F5, Phase 4).
