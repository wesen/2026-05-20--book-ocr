# Tasks

## Scraper-side (each lands via a scraper release, then a version bump in book-ocr)

- [ ] Item 1: RequeueSteps operator API — engineview mutation + pkg/workflow facade + transitive-downstream reset + InputPatch + tests; release
- [ ] Item 2: Lease heartbeats — RenewLease store op, heartbeat goroutine in executeLeasedOp, fenced CompleteOp/FailOp, HeartbeatInterval config; release
- [ ] Item 3: Cooperative cancellation — cancel_requested_at column, revised CancelWorkflow, heartbeat-driven ctx cancel, E_CANCELED path, grace force-cancel; release
- [ ] Item 4 (optional): extract pkg/workflow + pkg/engine (+ engineview) into go-go-golems/workflow; re-point scraper and book-ocr

## book-ocr-side

- [ ] Item 5 (interim, no release needed): schema-version guard in structured-rerun-pages (refuse unknown schema_migrations) — implemented under BOOK-OCR-PRODUCT-001
- [ ] Adopt RequeueSteps after Item 1 ships: replace requeueStructuredPages SQL, delete guard
- [ ] Extend workflow_retry_test with rerun-through-API and slow-executor (heartbeat) scenarios as each item ships
