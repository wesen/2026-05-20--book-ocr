---
Title: Glossary of terms, coinages, and writing patterns
Ticket: BOOK-OCR-PRODUCT-001
Status: active
Topics:
    - book-ocr
    - productization
    - workflow
    - ocr
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: A glossary of the vocabulary used across this ticket's documents — borrowed terms with their source literatures, project-specific coinages with the reasoning behind each choice, and the rhetorical constructions that compress arguments — so a developer new to these docs can decode them and understand how the terms were chosen.
LastUpdated: 2026-07-03T17:43:01.032097336-04:00
WhatFor: Decode the lingo in design docs 01-03 and the diaries; understand where each term comes from and why it was picked.
WhenToUse: Keep beside the design docs on first read; consult when a term reads as jargon.
---

# Glossary of terms, coinages, and writing patterns

## How this vocabulary arises (read this first)

The terms below come from three habits, and knowing the habits explains most individual entries:

1. **Borrowing from adjacent literatures.** When a concept already has a precise name in some engineering literature — testing theory, distributed systems, capability security, project management — I use that name rather than inventing one, because the borrowed term carries its literature's precision with it. The cost is that the reader must know the source literature; this glossary supplies it.
2. **Coining names to make things referable.** Arguments get shorter when a recurring idea has a handle. "Findings F1–F9", "source-derived", "money bug" exist so later sentences can point at a whole paragraph of reasoning with two words. These coinages are built from ordinary words chosen for the *distinction they encode* — the entries below spell out which distinction.
3. **Compression under length pressure.** Long documents reward dense sentences. Constructions like "the DAG — not caution — dictates" or "survives intact and gets sharper" pack a claim and its counter-claim into one clause. When the compression fails for a reader, it reads as oracular; the "writing patterns" section unpacks the recurring ones.

None of these words were chosen in reference to anything secret or private; where a term has a known origin (a book, a subculture, a codebase), the entry names it. Where I coined it, the entry says so and reconstructs the choice.

## Part I — Borrowed terms and their source literatures

**seam** — An extension point: a place in the pipeline where behavior can be changed without editing the code at that place. Borrowed directly from Michael Feathers, *Working Effectively with Legacy Code* (2004), where a seam is "a place where you can alter behavior in your program without editing in that place." I chose it over "extension point" or "hook site" because Feathers' term emphasizes that the surrounding fabric stays intact — which is precisely the design rule here ("plugins replace strategies, never invariants"). The S1–S8 numbering in design doc 02 exists so seams can be referenced like findings.

**invariant** — A property the system guarantees under all extensions: exactly one target-page image per OCR call; the model returns JSON and Go renders deterministically; the 01–07 artifact sequence exists for every page. Standard mathematics/CS usage (a condition preserved by every operation). In these docs it carries product weight: an invariant is something a user may *rely on*, which is why the seams deliberately exclude them.

**strategy** — The replaceable half of the strategy/invariant pair: *how* a page becomes structured JSON, as opposed to *what is guaranteed about the result*. From the Strategy pattern (Gamma et al., *Design Patterns*, 1994): interchangeable algorithms behind a stable interface. `StructuredOCRClient` is literally a Strategy interface, which made the vocabulary natural.

**surface** — The set of things exposed to some audience: "extension surface" (what plugin/script authors can touch), "agent surface" (what an LLM driver can call), "public API surface". Generalized from "API surface area", common in API-design writing (e.g., .NET framework design guidelines). I use it because "interface" is overloaded in Go.

**golden files / goldens** — Test fixtures that pin exact expected output; the test fails on any byte difference and a `-update` flag regenerates them. Standard term from snapshot/approval testing; widespread in Google's codebases and in the Go community (`testdata/*.golden` is an idiom `gofmt`'s own tests use). Chosen because the repo's refactors needed *proof of non-change*, which is exactly what goldens provide.

**oracle (validation oracle, anchor oracle)** — A mechanism that decides whether output is correct without a human. From software-testing theory (the "test oracle problem", Weyuker 1982). This one was inherited: the repo's May vlm-separation work already used "anchor oracles" for its expected/forbidden phrase lists; my "text layer as a free validation oracle" extends the repo's own usage — an independent transcription that adjudicates OCR output.

**capability / capability allowlist / "granted by construction, not restricted by confinement"** — From capability-based security (object-capability literature: Dennis & Van Horn 1966; E language; Capsicum): a program can only do what it holds an explicit handle to. Applied to goja in design doc 03: a script can only `require` what the host linked in, so the sandbox is the *absence of code paths*, not a kernel fence around existing ones. "Confinement" is the contrasting term for OS-level sandboxing (namespaces, seccomp) used for the NDJSON plugins.

**fail-closed / fail-open, fail-fast** — Security-engineering terms: a control that denies by default when its assumptions break (closed) versus permits (open); and the practice of erroring at startup rather than misbehaving later. Used for the allowlist-vs-denylist decision ("denylists fail open when upstream adds a module") and the plugin manager's startup seam validation.

**fencing / fenced completion** — From distributed systems (fencing tokens; see Kleppmann, *Designing Data-Intensive Applications*, ch. 8): a write is accepted only if it carries the still-valid token, so a worker that lost its lease cannot commit stale results. Used in the runtime-hardening ticket for making the lease token a precondition of `CompleteOp`.

**lease / drain / fan-out** — Worker-coordination vocabulary, all standard in distributed/queueing systems: a lease is time-bounded exclusive ownership of a work item; draining means letting in-flight work finish while admitting no new work; fan-out is one step emitting many parallel children. The scraper engine's own schema uses "leases", so these were partly inherited.

**provenance** — Where an artifact came from: which strategy, which plugin source hash, which model. From data-lineage/archival usage (and the art world before that). Chosen over "metadata" because provenance specifically answers "who produced this and can I trust/reproduce it".

**pinning (version pinning, source pinning, "pin the bytes")** — Fixing something that could drift to an exact recorded value: a dependency version, a plugin's source hash, a golden file's bytes. From dependency-management usage ("pin your dependencies"); I generalized it to any drift-prevention-by-recording because the underlying move is identical.

**shim** — A small compatibility piece that fills a gap between two components without being a real feature (carpentry origin: a thin wedge). Used for interim constructs slated for deletion, e.g. the schema guard that dies when `RequeueSteps` ships.

**smoke test / smoke** — The cheapest end-to-end check that the system basically works, run before anything expensive. Hardware-era origin (power it on; if smoke comes out, stop). "The dry-run smoke" = the 3-page pipeline check in CI.

**blast radius** — How much damage a failure can cause. SRE/ops vocabulary (popularized by AWS and Google SRE writing). Used to argue that host-side invariants "bound the blast radius" of arbitrary plugin behavior to wrong-content-on-inspectable-pages.

**posture ("hardened posture")** — A configuration stance toward risk, from infosec ("security posture"). Used for the goja builder settings that disable implicit module loading.

**policy / mechanism, policy knob** — The classic systems distinction (Hydra OS papers, later X Window System): mechanism is what the system *can* do, policy is what it *chooses* to do, and the two should be separated. The whole Phase-2 refactor is this distinction executed: `bookprofile` holds policy, the pipeline holds mechanism. A "knob" is systems slang for a tunable policy parameter.

**long pole** — The schedule-critical item everything else waits on (from tent-raising: the longest pole determines the tent's height). Project-management idiom. Used for the cross-repo scraper work.

**park / parked** — To deliberately set a topic aside in a place where it will be found again (meeting-facilitation "parking lot"). The runtime-hardening ticket "parks" the scraper-side items: not dropped, not now.

**land / landing** — To merge/commit a change into the mainline. Commit-culture vocabulary from Mozilla/Chromium ("land a patch"). Used interchangeably with "ship" for smaller units.

**retire (a risk, a hack)** — To eliminate permanently, not merely mitigate. From risk-management usage ("retire the risk") and backlog culture. "The prototype retired the transport risk" = after the experiment, that uncertainty no longer exists.

**amortize** — To spread a one-time cost across many uses until it is negligible per use. Finance term, standard in algorithm analysis ("amortized complexity"). "The frozen factory amortizes everything shareable" = build cost is paid once, runtimes are cheap.

**long tail** — The many small remaining items after the important few (statistics; popularized by Chris Anderson, 2004). "Plugin P3 is the long tail."

**exit criterion** — The testable condition that defines a phase as finished, chosen *before* the work. Project-management/agile usage. E.g. Phase 2's "a second book runs with zero Go changes."

**A/B** — Comparing two variants on the same input to attribute the difference to the variant. From controlled experimentation (web experimentation culture). The pilot's three-variant run is an A/B/C.

**drift (schema drift, API drift, "agents drift on field shapes")** — Gradual unintended divergence from a contract. From data engineering ("schema drift") and configuration management ("configuration drift"). Applied to models and agents improvising field names: the same phenomenon, so the same word.

**identity plugin / identity test** — A pass-through implementation used to prove a seam is transparent when it does nothing: output must be byte-identical to the built-in path. From the mathematical identity function. Coined-by-application here; the concept is common in adapter testing.

**adversarial fixtures** — Test inputs designed to misbehave (contaminated stdout, never-responds, lying page numbers). "Adversarial" from security/ML usage; the practice was directly inspired by devctl's own `testdata/plugins/` misbehaving-plugin suite, which the report credits.

**undersampled** — Statistics: a region of the input space with too few observations to support conclusions. Used honestly against my own pilot ("the 24-page range is front-matter-heavy; structure handling is undersampled").

**field study** — Observing behavior in its real setting rather than a lab. Social-science term. "The pilot was a field study of the agent-first mode" — meaning the frictions were observed during genuine work, not constructed.

## Part II — Coinages specific to this project

Each entry has three parts: the definition, **what I was actually thinking or doing at the moment the term appeared** (reconstructed from memory of this session — where I am reconstructing rather than replaying, I say so), and the exact first written use.

**source-derived vs run-derived** — *Source-derived*: data fully determined by the input artifact (the PDF, the profile) before any OCR runs. *Run-derived*: data produced by this run's execution. The distinction is the invention; the words are ordinary.

*What I was thinking:* The owner had just asked "can we do better than a page context? we have context for the entire book/project, no?" and my immediate internal objection was a race: page steps run in parallel, so during `ocrPage(42)`, page 41 may be unfinished — and a targeted rerun months later sees a completely different neighborhood, which breaks the promise that reruns reproduce pages. I was about to write "scripts can only get book context in post-stages" when the exception surfaced: the text layer doesn't have this problem, because it comes from the PDF, which exists *before* the run. That flipped the design question from "when is data available?" to "what does the data derive from?" — availability is racy, derivation is causal. I considered "static vs dynamic" (rejected: both words are saturated in programming), and "pre-run vs post-run" (rejected: temporal framing invites "well page 41 IS done by now" arguments), and chose derivation because it makes the determinism argument inherent to the noun: if it derives from the source, no execution order can change it.

*First written use* (design doc 03, "The context model: page, book, run"):

> **`book`** — *source-derived* context, immutable from the moment discovery completes and therefore safe everywhere: the compiled profile view (lexicon, policies), the ingest manifest (source hash, DPI, page count), and crucially `book.textLayer(n)` / `book.imageInfo(n)` for **any** page — the text layer comes from the source PDF, not from the run, so reading page 41's text layer while OCRing page 42 is deterministic regardless of execution order.

**three-scope / layered context (page, book, run)** — The `bookocr` module's context as three nested visibility scopes, each hook seeing exactly the scopes deterministic at its DAG position.

*What I was thinking:* While structuring the same amendment, the three context kinds arranged themselves like lexical scopes — an inner scope sees the outer ones, and visibility is a property of *where you stand*, which is exactly how the hook table works (where you stand in the DAG determines what you may read). "Layered" went into the doc; "three-scope" appeared only afterwards, while writing the chat summary, where I needed the whole design as a two-word noun phrase. It is a summary artifact, not a design term — which is itself worth knowing about my summaries: they mint compounds the documents never used.

*First written use* (chat, reporting the amendment):

> Doc 03 is amended (uploaded, pushed) with a three-scope context model:

**money bug** — A correctness bug whose consequences become financial once usage is metered.

*What I was thinking:* The owner asked what a credits-based product would need, and I was re-reading the runtime-hardening list I had written hours earlier through that new lens. Lease expiry causing duplicate execution had been filed as hygiene — annoying, mitigated by idempotent writes. Under metering, a duplicated model call is a duplicated charge: either the user pays twice or the margin eats it. I needed to express that the *bug was unchanged but its category moved*, and a category needs a name. "Billing bug" was the first candidate; I rejected it as too narrow (it suggests defects in invoicing code, not in the scheduler). "Money bug" is blunter and covers both directions of the loss. The "X bug" template (heisenbug, security bug) made it instantly parseable.

*First written use* (chat, credits-MVP analysis; the vault article later echoes it as "change category if the product ever meters usage"):

> Two items from WORKFLOW-RUNTIME-HARDENING-001 get promoted from "hygiene" to **money bugs**: 1. **Lease heartbeats** — a slow page step getting re-leased means duplicate model calls. Today that's your `/tmp` dir and your API key; in a credits product it's double-billing users or eating margin.

**structure-blind** — The text-layer strategy's specific incapacity: perfect prose reproduction, zero perception of layout.

*What I was thinking:* I was staring at the structure-sample comparison — the textlayer column showing `paragraph(466), paragraph(346)…` on a page whose VLM column showed headings, a code block, a figure. The finding wasn't "the text layer is worse"; on prose it had just tied the model byte-for-byte. The deficit was *dimensional*: one capability missing, everything else intact. "Color-blind" is the everyday word for exactly that shape of deficit, so the template transferred directly. I wanted the word to protect the strategy from being dismissed ("it's bad at structure" reads as general weakness; "structure-blind" reads as a known scope).

*First written use* (pilot design doc 01, structure-sample addendum):

> The textlayer variant emitted paragraphs on all three pages — structure-blind exactly as W4 predicted, confirming the routing story: prose pages to the free strategy, structure pages to vision.

**decline-to-builtin** — The `response.parse` fallback semantics: a plugin answers `E_DECLINED` and the built-in parser takes over.

*What I was thinking:* Designing the parse seam, I kept hitting the same wording problem: every verb I reached for ("fails over", "rejects", "errors out") implied malfunction, but the whole point of `E_DECLINED` is that declining is a *successful* outcome — the plugin correctly recognized a format as not-its-business. When P2 shipped and I was compressing it into the task list, I needed the mechanism as a modifier and built the hyphenated compound on the "fail-open/fail-closed" pattern: direction-of-fallback encoded in the term.

*First written use* (BOOK-OCR-PRODUCT-001 tasks.md, P2 completion entry):

> Plugin track P2: response.parse (decline-to-builtin), validate.page/book (tagged, additive), page.classify with per-page strategy routing, plugin retryable-hint classification

**routing economics** — The cost consequence of per-page strategy routing: routing pages between free and paid strategies is a spending decision.

*What I was thinking:* Writing the structure-sample chat summary, I was connecting two results produced hours apart: the credits analysis (cost per page is the metering unit) and the fresh evidence that the free strategy ties the paid one on prose. The connection *is* the product feature — classify-then-route decides where money goes — and I wanted the reader to see the economics and the routing as one object, not two observations. The doc version says "routing story"; "economics" is the variant the money-focused summary kept, and it is the better word because "story" claims narrative while "economics" claims arithmetic.

*First written use* (chat, structure-sample report; the doc's parallel sentence is quoted under structure-blind above):

> The text-layer variant emitted plain paragraphs on all three pages — structure-blind exactly as predicted, which confirms the routing economics: prose pages to the free strategy, structure pages to vision.

**driver / LLM driver** — The agent operating the product: reading outputs, acting on errors, writing extension code.

*What I was thinking:* Writing the agent-first doc's opening, I had just *been* the thing I was naming — I had spent the pilot parsing zerolog lines, hand-writing a wrapper script and a comparison harness. "User" was wrong (implies a human at a terminal); "agent" was already taken twice over (the product's own plugins, and geppetto's `agent()` API). The verb was already in my draft sentence — "operated end to end by an LLM agent *driving* the CLI" — and nominalizing a verb I had already committed to felt more honest than importing a new noun. The device-driver resonance (a component operating another through a defined interface) was noticed after the fact and kept because it was apt, not because it was the source.

*First written use* (pilot design doc 02, Executive Summary):

> The Wilensky pilot was operated end to end by an LLM agent driving the CLI — which makes it a field study of exactly the product mode this document designs for. Every friction the operator hit is a friction any driver, human or agent, will hit…

**agent-first** — A product whose primary operator is an LLM driver, humans second. **Not my coinage** — it is the repository owner's term, from the request that started the design: "imagine we are going to make this product 'agent-first' as well, so that most of the interactions with it will actually happen through an LLM."

*What I was thinking on adoption:* The term slotted into the established "-first" template (mobile-first, API-first, offline-first), which carries a precise meaning — not "supports X" but "designs for X first and lets the other audiences inherit" — and that was exactly the claim the document needed to defend, so I kept the owner's word rather than translating it.

*First written use by me* (pilot design doc 02, frontmatter Summary):

> …followed by the agent-first product design — machine-readable surfaces, a run manifest, a plugin authoring loop for LLM drivers, and the sandboxing/provenance model for executing agent-written plugins.

**free (as in "free strategy", "free validation oracle")** — Zero *marginal model cost*, not zero effort.

*What I was thinking:* Writing finding W1 I caught myself about to overclaim. "Free" is rhetorically powerful and therefore suspect — the text layer costs ingest time, cleanup code, and its own defects (the backslash broke the PDF build). What is genuinely zero is inference tokens, which happens to be the exact resource a credits product meters. So I kept the strong word but spent the sentence scoping it, and flagged it in this glossary because a reader who catches an unscoped "free" is right to distrust the surrounding argument.

*First written use in this scoped sense* (pilot design doc 01, finding W1):

> **W1 — The text layer is a competitive free strategy for prose.** For prose-dominant scanned books with IA-quality text layers, `ocr.page=textlayer` produces model-equivalent body text at zero cost.

**DAG-determinism ("staged by DAG determinism")** — Shorthand: a hook's readable context is fixed by its workflow-graph position such that every read is order-independent.

*What I was thinking:* Commit messages force the harshest compression in the whole workflow — the full argument was two paragraphs in the doc, and the commit subject-body format wanted it in one clause that a future `git log` reader could act on. "Staged" carries the DAG-position half; "determinism" carries the why. This is the same summary-mints-compounds behavior as "three-scope": the term exists because a *pointer* to the argument was needed, and it should always be read as a pointer, never as the argument.

*First written use* (commit 19f96b6):

> The script context is staged by DAG determinism: source-derived book context (profile, manifest, any page's text layer) is safe in every hook because it precedes the run; run-derived cross-page output is readable only in post-assembly hooks, keeping page steps and targeted reruns order-independent.

**ink-band, context bleed, hard-cut** — Inherited, not coined by me: *ink-band* is the repo's May name for its pixel-row figure segmentation (`ink-band-v1`, `ocrquality/figures.go`); *context bleed* is the May tickets' name for neighbor-page images leaking content into a target page's OCR; *hard-cut* is geppetto's own name for its deliberately narrowed JS API (`module_hardcut_test.go`). When a codebase already names a thing, I keep its name — renaming inherited concepts costs every future reader a translation table.

## Part III — Writing patterns (the compressed constructions)

Same three parts per entry: what the pattern compresses, what I was thinking when it first did its work, and the sentence itself.

**"a new seam falls out"** — Mathematics idiom: a result arriving as a corollary, without additional machinery, once the right structure exists. I use it to mark design elements that are *consequences* rather than *decisions*, because reviewers argue with those differently.

*What I was thinking:* `postProcessBook` genuinely was not premeditated. I was filling in the hook-versus-scope table for the doc-03 amendment and noticed the `run` row had validators as consumers but no *producer-of-output* seam — and simultaneously remembered that "second-pass cleanup workflow" had been sitting in the repo's future-work lists since the May HQ-001 ticket. The seam wasn't designed; it was the intersection of a scope that now existed and a need that had always existed. "Falls out" was chosen to report exactly that: I wanted credit assigned to the structure, not to me, because a reviewer should probe the structure (is the `run` scope sound?) rather than the seam (which follows if it is).

*First written use* (chat, answering "can we do better than a page context?"; the doc states it as "the `run` scope motivates a seam the plugin design never had"):

> **The payoff: a new seam falls out.** Giving post-stages the `run` scope makes `postProcessBook` the natural home for the second-pass cleanup that's been sitting in future-work since HQ-001…

**"the DAG — not caution — dictates" / the "X, not Y" contrast** — Names the actual constraint while explicitly displacing the one the reader was about to assume.

*What I was thinking:* The owner's question ("can we do better than a page context?") carried a gentle implication that the page-only design had been unambitious. The honest answer was "yes, much better — except for one restriction that will look like the same unambition unless I get ahead of it." Restricting page-stage hooks from run-derived data *looks* like safety-culture reflex; it is actually forced by parallel execution and rerun reproducibility. The em-dash interruption exists so the displacement physically interrupts the sentence — a subordinate clause ("which is not merely caution") can be skimmed; a dash pair cannot.

*First written use* (design doc 03, opening the context-model section):

> A page-only context undersells what the host knows. The system holds context at three scopes, and the workflow DAG — not caution — dictates which scope each hook may see:

**"survives intact and gets sharper"** — Two claims about a constraint under a new design: nothing weakened, boundary more precisely drawn.

*What I was thinking:* While writing the `book.textLayer(n)` design I had an actual moment of doubt: does letting an OCR script read *any* page's text violate the single-image invariant — the empirically-justified rule the whole May redesign rests on? I went back through it: the rule, and the vlm-separation benchmark behind it, concerned neighbor *images*; text context was a separate benchmark scenario that behaved. So the new design didn't erode the invariant — it revealed the invariant's true boundary (images, not context in general) and made the text-context decision explicit instead of incidental. "Survives intact" answers the doubt I actually had; "gets sharper" reports the bonus. The pairing exists because I wanted the reader to traverse the same doubt-then-resolution in six words.

*First written use* (design doc 03, same section):

> First, the May invariant survives intact and gets sharper: the single-image rule concerned neighbor *images*, whose bleed the vlm-separation benchmark measured; neighbor *text layers* are source material, benchmarked separately (the `target-plus-text-context` scenario), and their use in a prompt is now an explicit, profile-visible script decision rather than a hidden host behavior.

**"two consequences are worth naming" / "worth naming"** — A curation marker: of everything downstream of the design, these are the non-obvious items a reviewer would want surfaced; the list is chosen, not exhaustive.

*What I was thinking:* The context-model section had many consequences (module API shape, rerun behavior, profile visibility, the new seam…), and enumerating all of them would bury the two that could change a reviewer's verdict: the invariant question and the accidental seam. "Worth naming" is my standing signal that a selection judgment happened — it comes from code-review habit ("this deserves a comment") and it deliberately leaves a hook for the reader to ask "what did you judge *not* worth naming?", which is the right challenge to invite.

*First written use* (design doc 03):

> Two consequences are worth naming. First, the May invariant survives intact and gets sharper… Second, the `run` scope motivates a seam the plugin design never had…

**"too thin" (the host contract is too thin)** — Thin/fat as how much a boundary specifies; standard vernacular (thin client, thin wrapper).

*What I was thinking:* Diagnosing the hybrid's empty title page, I had two artifacts open side by side: the raw response with its invented `"lines"` field, and the two prompts — the built-in one, sixty lines of field-by-field instruction with a worked example, versus the appended contract, six lines naming block *types* only. The word arrived from that literal visual comparison: one prompt was physically thick with specification, the other thin. The fix verb ("fatten the contract") was chosen the same second, because a spatial metaphor you commit to should conjugate.

*First written use* (pilot design doc 01, after the chat diagnosis "a `prompt.render` plugin replaces the *detailed* block contract and the host only appends the compact one"):

> **W3 — The host contract is too thin for prompt.render experiments.**

**"the honest X"** — A marker that the statement includes the unflattering part deliberately.

*What I was thinking:* Writing the first assessment, I was aware of the structural pressure on the document: it was going to recommend investing in this codebase, and assessments that end in "invest" tend to retro-fit their evidence. The section admitting the May sprint's own acceptance criteria were unfinished needed protection from later editing-for-advocacy — labeling it "honest" makes deleting it feel like the lie it would be. The word recurs wherever I hand the reader ammunition against my own recommendation (the goja memory gap, the plugin surface's remaining advantages).

*First written use* (design doc 01, section heading):

> ### Honest quality assessment

**"named, not solved"** — Risk identified and characterized; no mitigation claimed.

*What I was thinking:* The goja memory-cap gap had no fix I could offer — goja simply lacks a heap limit, and every "mitigation" I listed (watchdogs, cgroups) was containment, not solution. The design-doc failure mode I was steering away from is the risks section where every entry quietly ends "…but it's fine". Three words that refuse the reassurance felt more trustworthy than a paragraph of hedged mitigation — and they create an honest TODO: this risk is *open*, and anyone re-reading the doc later should still treat it as open.

*First written use* (design doc 03, risks):

> **No memory cap** is the structural weakness; scripts are bounded in time but not heap. Local single-user mode accepts it (same trust as running the CLI); hosted agent-first mode either wraps the whole worker in cgroup limits or routes fully-untrusted strategies to the plugin surface. Named, not solved.

**"load-bearing"** — A fact others silently rest on, whose removal collapses things; structural engineering via programming slang ("load-bearing comment").

*What I was thinking:* Writing the architecture overview for an intern, I wanted one sentence to carry a warning: the dependency direction (book-ocr imports the runtime, never the reverse) is not a stylistic preference — the externalization sprint, the productization plan, and the publish-versus-vendor decision all assume it. An intern refactoring casually could invert it locally and compile fine. "Load-bearing" says: this wall looks like the others, and it is not.

*First written use* (design doc 01):

> The dependency direction is strict and load-bearing: `book-ocr` imports `scraper/pkg/workflow`; scraper contains zero OCR knowledge.

**"sharp edge"** — Intended behavior, unintended injury; distinct from a bug.

*What I was thinking:* Categorizing `--dry-run` defaulting to true: not a bug (deliberate, documented, useful during development), yet guaranteed to hurt the first real user, whose maiden live run would silently produce fake output. I needed a word for "correct but wounding", and the tool-safety vocabulary supplies it — a sharp edge is a manufacturing choice that the user's hand discovers.

*First written use* (design doc 01, the review-loop section; then Phase 1: "Fix F6's sharp edge."):

> The manual validation workflow in `README.md:337-363` is the actual product core and survives unchanged — the phases just automate its sharp edges.

**Residual metaphors: "fossil record", "engine room", "pressure valve"** — Three metaphors that escaped the no-analogies discipline; decoration, not argument.

*What I was thinking, honestly:* Each was reached for at a moment of narrative rather than analysis — introducing the repo's history (fossil record: layers preserving the failure that caused each), summarizing test results in a diary (engine room: internals versus packaging), and selling a sequencing choice in chat (pressure valve: prompt experiments relieving demand on the Go refactor). They came from the general metaphor stock, not from anywhere specific, and none survives scrutiny as an argument — which is why they are confessed here rather than defended. First uses:

> The repository's shape is the fossil record of one week of iteration, preserved in `ttmp/`. *(design doc 01)*

> The repo's engine room is healthy; the productization gaps are packaging and configuration, not correctness. *(diary step 3)*

> `prompt.render` plugins give you a pressure valve: you can run prompt experiments per book type *while* the Go-side generalization is still in flight, instead of the two competing for the same files. *(chat, the work-ordering answer)*

**Findings and letters (F1–F9, W1–W8, D1–D4, S1–S8, G1–G4, P1–P3)** — Numbering findings, decisions, seams, and phases so any sentence can cite one in two characters; RFC/requirements-engineering style.

*What I was thinking:* F1 was assigned in the diary minutes after the first build failure — *before* the design document existed. That ordering was deliberate: I knew the eventual report would need to say things like "Phase 1 fixes this" about discoveries that hadn't finished happening, so the findings needed stable names from the moment of discovery. The letters are nothing deeper than mnemonic initials; the practice is the point, not the alphabet.

*First written use* (diary step 2, the session's first failed experiment):

> This is finding F1 of the productization report: the repo has an unpublished, path-coupled dependency on the workflow runtime and does not build from a clean clone.

## Terms from the original query that do not appear in this project's writing

For completeness: *hardware ladder*, *offload hypothesis*, *graduate a policy*, *bridge files*, *thin bridge*, and *crucial product idea* do not occur in this ticket's documents, diaries, or reports and are not part of this project's vocabulary. ("Thin" and "product" occur, but in the constructions glossed above.)
