# Changelog

## 2026-05-24

- Initial workspace created


## 2026-05-24

Created externalization analysis and intern-facing implementation guide for moving Report 794 OCR policy out of scraper and into 2026-05-20--book-ocr.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-EXTERNALIZE-001--move-book-specific-ocr-policy-out-of-scraper-into-book-ocr-repository/design-doc/01-externalizing-book-ocr-policy-from-scraper-design-and-implementation-guide.md — Design guide
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-EXTERNALIZE-001--move-book-specific-ocr-policy-out-of-scraper-into-book-ocr-repository/reference/01-diary.md — Diary


## 2026-05-24

Uploaded BOOK-OCR-EXTERNALIZE-001 design guide and diary to reMarkable under /ai/2026/05/24/BOOK-OCR-EXTERNALIZE-001.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-EXTERNALIZE-001--move-book-specific-ocr-policy-out-of-scraper-into-book-ocr-repository/design-doc/01-externalizing-book-ocr-policy-from-scraper-design-and-implementation-guide.md — Uploaded design guide


## 2026-05-24

Corrected scope: move all OCR/book-OCR functionality out of scraper, leaving scraper as workflow/job runtime only.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-EXTERNALIZE-001--move-book-specific-ocr-policy-out-of-scraper-into-book-ocr-repository/design-doc/01-externalizing-book-ocr-policy-from-scraper-design-and-implementation-guide.md — Corrected full OCR extraction design
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-EXTERNALIZE-001--move-book-specific-ocr-policy-out-of-scraper-into-book-ocr-repository/reference/01-diary.md — Diary Step 2


## 2026-05-25

Moved OCR workflows, quality workers, book profile code, and OCR CLI into 2026-05-20--book-ocr; removed OCR packages from scraper after external smoke test passed (book-ocr commit 04785a5, scraper commit cd01992).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go — External CLI
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrquality/package.go — External quality workflow
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/runtime.go — Remaining generic runtime

