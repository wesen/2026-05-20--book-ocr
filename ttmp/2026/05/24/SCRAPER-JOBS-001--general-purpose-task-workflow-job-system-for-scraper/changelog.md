# Changelog

## 2026-05-24

- Initial workspace created


## 2026-05-24

Created evidence-backed design guide and diary for generalizing scraper into a task/workflow/job system with book OCR mapping and River backend analysis.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/01-scraper-task-workflow-job-system-design-and-book-ocr-mapping.md — Primary design guide
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/reference/01-diary.md — Investigation diary


## 2026-05-24

Validated SCRAPER-JOBS-001 with docmgr doctor and uploaded design guide plus diary bundle to reMarkable at /ai/2026/05/24/SCRAPER-JOBS-001.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/01-scraper-task-workflow-job-system-design-and-book-ocr-mapping.md — Uploaded primary design guide
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/reference/01-diary.md — Uploaded diary


## 2026-05-24

Updated design guide with concrete AITR-794 OCR evidence: prompt versions, page attempts, subset reprocessing, dashboard needs, and SQLite queue race lessons.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/01-scraper-task-workflow-job-system-design-and-book-ocr-mapping.md — Updated design guide
- /home/manuel/workspaces/2026-05-20/book-ocr/claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/reference/01-diary.md — Concrete OCR evidence source


## 2026-05-24

Uploaded refreshed AITR-updated SCRAPER-JOBS-001 bundle to reMarkable at /ai/2026/05/24/SCRAPER-JOBS-001.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/01-scraper-task-workflow-job-system-design-and-book-ocr-mapping.md — Refreshed upload includes AITR-794 OCR evidence update


## 2026-05-24

Added second design document for an embeddable River-like Go API, typed job handlers, scraper-vs-River tradeoffs, and River-as-backend architecture.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/02-embeddable-scraper-job-api-and-river-backend-design-considerations.md — New embeddable API and backend design guide


## 2026-05-24

Uploaded embeddable API and River backend design bundle to reMarkable at /ai/2026/05/24/SCRAPER-JOBS-001.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/02-embeddable-scraper-job-api-and-river-backend-design-considerations.md — Uploaded second design document


## 2026-05-24

Rewrote the second design document from the ground up around a workflow-native embeddable runtime API, removing queue-library framing and foregrounding Runtime, WorkflowPackage, Run, Step, Executor, ArtifactStore, ProjectionStore, EventSink, and operator actions.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/02-embeddable-workflow-runtime-api-design.md — Ground-up rewritten API design


## 2026-05-24

Uploaded rewritten workflow-native API design bundle to reMarkable at /ai/2026/05/24/SCRAPER-JOBS-001.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/02-embeddable-workflow-runtime-api-design.md — Uploaded rewritten workflow-native API design


## 2026-05-24

Implemented Phase 1 workflow executor facade in scraper (commit bc6baa26d8fb7eb3a78b8d9e32ab544ee6deaf43) with StepContext helpers, typed executor adapter, tests, and full go test validation.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/context.go — StepContext input/result/artifact/record/emit helpers
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/errors.go — Retryable/permanent workflow errors
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/executor.go — Workflow-native Executor facade and runner adapter
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/executor_test.go — Phase 1 facade tests


## 2026-05-24

Implemented Phase 2 workflow runtime skeleton in scraper (commit 4dd78466d8d1faa70df96df5aa59805ad831441d) with Runtime, SQLiteStore, package entrypoints, RunBuilder, StartRun, RunOnce, and tests.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/package.go — Workflow package
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/runtime.go — Runtime skeleton and SQLite-backed scheduler/store wiring
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/runtime_test.go — Embedded runtime integration tests

