---
docType: reference
status: active
intent: short-term
topics:
  - ocr
  - experiments
created: 2026-05-24
updated: 2026-05-24
---

# Experiment 003 run capture summary

## Command

["bash", "-lc", "cd '/home/manuel/workspaces/2026-05-20/book-ocr/scraper' && go run ./cmd/ocr-mvp run --book-id presentation-based-uis-hq-003-quality-v3 --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages --work-dir /tmp/book-ocr-hq-001/003-quality-v3-list-diplomatic --start-page 1 --end-page 9 --profile gpt-5-nano-low --profile-registries /tmp/book-ocr-hq-001/profiles-clean.yaml --prompt-version ocr-quality-v3-list-diplomatic --max-workers 2 --log-level warn"]

## Counts

- Return code: 0
- Captured lines: 10
- Parsed JSON lines: 0
- Trace lines: 0
- Warning/error lines: 0

## Non-trace/status timeline

```text
line_no	captured_at	level	event	op_id	attempt	workflow_status	message
1	1779664592.19808						/home/manuel/.hishtory/config.sh: line 72: bind: warning: line editing not enabled
2	1779664593.4466						started run ocr-mvp-b66cef03-7588-4cfe-8f57-a941fedd2e1e in /tmp/book-ocr-hq-001/003-quality-v3-list-diplomatic
3	1779664593.4585						status=running processed=1 succeeded=0 failed=0 retried=0
4	1779664597.67117						status=running processed=2 succeeded=0 failed=0 retried=0
5	1779664608.39257						status=running processed=2 succeeded=0 failed=0 retried=0
6	1779664622.12241						status=running processed=2 succeeded=0 failed=0 retried=0
7	1779664633.43768						status=running processed=2 succeeded=0 failed=0 retried=0
8	1779664651.77185						status=running processed=2 succeeded=0 failed=0 retried=0
9	1779664652.02938						status=succeeded processed=1 succeeded=0 failed=0 retried=0
10	1779664652.02943						assemble result: {"book_id":"presentation-based-uis-hq-003-quality-v3","page_count":9,"markdown_ref_id":"assemble-markdown:artifact:001","markdown_uri":"file:///tmp/book-ocr-hq-001/003-quality-v3-list-diplomatic/artifacts/assemble-markdown/artifact/001","char_count":8587}
```
