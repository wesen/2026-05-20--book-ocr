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

**source-derived vs run-derived** — The load-bearing pair in the script-context design. *Source-derived*: data fully determined by the input artifact (the PDF, the profile) before any OCR runs — e.g. any page's text layer. *Run-derived*: data produced by this run's execution — e.g. page 41's rendered Markdown. I coined the pair because the sandbox question ("which context may a hook read?") reduces to a determinism question, and I wanted the names themselves to carry the answer: source-derived data is order-independent and therefore safe everywhere; run-derived data is only stable after the stage that produces it. The words are ordinary; the *distinction* is the invention, and naming it let one table replace a page of case-by-case argument.

**three-scope / layered context (page, book, run)** — The `bookocr` module's context object, organized as three nested visibility scopes. "Scope" is borrowed from lexical scoping in programming languages (a name is visible in some region and not others); "layered" from layered architecture. Coined as a structure, not a metaphor: each hook sees exactly the scopes whose contents are deterministic at its DAG position.

**money bug** — A correctness bug whose consequences become financial once usage is metered: lease expiry causing duplicate model calls, cancellation not stopping in-flight inference. Coined during the credits-MVP analysis to explain a *priority inversion*: these were known, tolerable hygiene items for a personal tool, and the coinage marks the exact moment they change category. Built on the plain-English pattern of "X bug" (heisenbug, security bug).

**structure-blind** — Description of the text-layer strategy: it reproduces prose but cannot perceive layout (headings, tables, figures) because the PDF text layer contains none. Coined on the model of "color-blind": capable in general, insensitive to one specific dimension.

**decline-to-builtin** — The `response.parse` seam's fallback semantics: a plugin answers `E_DECLINED` and the built-in parser takes over, so plugins handle only formats they recognize. Coined to name the *protocol outcome that is not an error* — ordinary error vocabulary ("fail", "reject") would wrongly imply something went wrong.

**routing economics** — The cost consequence of `page.classify` strategy routing: prose pages to the free text-layer strategy, structure pages to the paid vision model. Coined by compounding; the underlying observation (per-page marginal cost differs by strategy, so routing is a cost decision) needed a two-word handle for the credits discussion.

**driver / LLM driver** — The agent operating the product: reading outputs, acting on errors, writing plugins/scripts. Chosen over "user" (which implies a human) and "agent" alone (overloaded). The image is of *driving* the CLI — the term came from describing this session itself ("an LLM drove the CLI end to end"), then hardened into a role name. Not a reference to device drivers, though the resonance (a component that operates another through a defined interface) is apt and welcome.

**agent-first** — A product designed so that the primary operator is an LLM driver, with humans second. Coined on the established "-first" template (mobile-first, API-first, offline-first): the suffix signals a design-priority inversion, which is exactly the claim.

**free (as in "free validation oracle", "free baseline")** — Zero *marginal model cost*, not zero effort: the text layer costs no inference tokens because the source PDF already contains it. The economic sense from "free as in beer", scoped deliberately to the metered resource. Flagged here because the word does real argumentative work in the pilot findings and could read as sloppy if taken to mean "no work at all".

**DAG-determinism ("staged by DAG determinism")** — Shorthand for the rule that a hook's readable context is determined by its position in the workflow graph such that every read is order-independent. A compression of the full argument in doc 03's context-model section; the phrase exists so tables and summaries can cite the rule without restating it.

**ink-band, context bleed, hard-cut** — Inherited, not coined by me, but glossed here because they confuse newcomers equally: *ink-band* is the repo's own name for its pixel-row figure-segmentation heuristic (`figures.go`, method tag `ink-band-v1`); *context bleed* is the May tickets' name for the defect where neighboring page images leaked content into a target page's OCR; *hard-cut* is geppetto's own name for its deliberately narrowed JS API (their test files use the term). When a codebase already names a thing, I keep its name.

## Part III — Writing patterns (the compressed constructions)

These are rhetorical habits, not technical terms. Each entry shows the pattern, what it compresses, and why I reach for it.

**"a new seam falls out"** — "Falls out" is a mathematics idiom: a result that arrives as a corollary, without additional machinery, once the right structure is in place. The sentence compresses: *we did not set out to design `postProcessBook`; once the run scope existed for validators, the post-pass seam required no further invention*. I use "falls out" specifically to mark design elements that are consequences rather than decisions — the distinction matters for review, because you argue with decisions differently than with corollaries.

**"the DAG — not caution — dictates" / the "X, not Y" contrast** — A pre-emptive strike against the most likely objection. When a design looks conservative (scripts can't read neighbor OCR output), a reader's first hypothesis is timidity. The construction names the actual constraint (parallel execution order, rerun reproducibility) and explicitly displaces the assumed one. I use it whenever the *reason class* for a restriction is likely to be misattributed. The em-dash interruption is deliberate: it makes the displacement impossible to skim past.

**"survives intact and gets sharper"** — Compresses two separate claims about the single-image invariant under the new context model: (1) *intact* — no part of the old rule is weakened; (2) *sharper* — the rule's boundary is now more precisely drawn (it was always about images, and the design now makes text context an explicit, visible decision instead of an undifferentiated prohibition). The pairing exists because refactors are usually suspected of eroding old guarantees; the sentence asserts the opposite in both directions at once.

**"two consequences are worth naming" / "worth a comment" / "worth naming"** — A discourse marker announcing deliberate selection: out of everything that follows from the preceding design, these are the non-obvious items a reviewer would want surfaced. It comes from code-review culture ("this deserves a comment") and from mathematical writing ("we remark that…"). The function is honesty about curation — signaling that the list is chosen, not exhaustive, and inviting the reader to trust that omitted consequences were judged routine.

**"too thin" (the host contract is too thin)** — Thin/fat as a measure of how much a boundary specifies or does: a *thin* contract names categories but not shapes; a *thin wrapper* adds no behavior. Standard software vernacular (thin client, thin wrapper), applied to the appended prompt contract whose thinness caused the W3 schema drift. The fix is correspondingly to "fatten" it — same axis, opposite direction.

**"the honest X" ("the honest division", "the honest gap", "honest quality assessment")** — A marker that the following statement includes the unflattering part deliberately: goja's missing memory cap, the pilot's undersampling, the division of labor where plugins keep real advantages. The habit exists because design documents drift toward advocacy; tagging the counter-inventory as "honest" holds a slot for it that advocacy cannot quietly delete.

**"named, not solved"** — The explicit admission that a risk has been identified and characterized but no mitigation is being claimed. Used for the memory-cap gap. It protects against the most common design-doc failure: a risks section whose entries all secretly end "…but it's fine".

**"load-bearing"** — From structural engineering via programming slang (a "load-bearing comment" is one whose removal breaks something surprisingly). A fact or dependency is load-bearing when other parts of the argument rest on it: "the dependency direction is load-bearing" means several later guarantees silently assume it. I use it to warn future editors what not to casually change.

**"sharp edge"** — A property of a tool that injures users who merely brush against it — e.g. `--dry-run` silently defaulting to fake output on a user's first real run. Usability vernacular (common in developer-experience writing). Distinct from a bug: the behavior is intended, the injury is not.

**"the engine room is healthy" / "fossil record" / "pressure valve"** — Residual metaphors (ship's engine room = the internal machinery as opposed to the packaging; fossil record = the code's shape preserving its history of failures; pressure valve = a mechanism that relieves demand on a contended resource). The textbook style used for the vault article bans analogies, and these three predate or escaped that rule in ticket prose. They are decoration, not argument; nothing depends on them, and this glossary is their apology.

**Findings and letters (F1–F9, W1–W8, D1–D4, S1–S8, G1–G4, P1–P3)** — Not vocabulary but the habit that generates the most cross-references: numbering findings, decisions, seams, and phases so any later sentence can cite one in two characters. The letters are mnemonic initials (F=finding, W=Wilensky, D=decision, S=seam, G=goja, P=plugin-phase). The practice is borrowed from requirements engineering and RFC style; its value is that traceability ("Phase 2 fixes F2–F4") becomes checkable.

## Terms from the original query that do not appear in this project's writing

For completeness: *hardware ladder*, *offload hypothesis*, *graduate a policy*, *bridge files*, *thin bridge*, and *crucial product idea* do not occur in this ticket's documents, diaries, or reports and are not part of this project's vocabulary. ("Thin" and "product" occur, but in the constructions glossed above.)
