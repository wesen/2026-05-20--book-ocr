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

# OCR quality v3 list diplomatic prompt

```go
func renderQualityV3ListDiplomaticPrompt(input PageOCRInput, version string) string {
	return fmt.Sprintf(`Transcribe this scanned technical-report page into faithful, clean markdown.

Output contract:
1. Output only the transcription markdown. Do not explain your work.
2. Transcribe only visible page content. Do not add summaries, comments, or inferred missing text.
3. Exclude standalone running page numbers, footer folios, scanner borders, and scanner artifacts.
4. Preserve original spelling and historical terminology, including "data base" when visible.
5. Preserve visible text order from top to bottom.

Global normalization policy:
- Prefer readable markdown over visual line wrapping for normal prose.
- Join wrapped title lines when they are clearly one title phrase, unless the line break is semantically meaningful.
- Do not duplicate a line just because it appears as both a visual heading and a list row. If the same visible line is repeated on the page, transcribe it once per visible occurrence; otherwise do not invent duplicates.

Page-type rules:
- Blank or intentionally blank page: output exactly [BLANK PAGE].
- Title/front-matter page: transcribe visible report number, title, author, institution, date, copyright, and notes as text. Do not use an image marker for title pages. Normalize the main title to one readable line when it is a single phrase.
- Abstract/acknowledgments/body page: preserve headings and paragraphs. If a paragraph begins or ends mid-sentence because of the page boundary, transcribe exactly the visible fragment without ellipses.
- Table of Contents pages: use a diplomatic plain-text list, not markdown bullets and not markdown headings. Preserve chapter titles, section numbers, section titles, dot leaders or spacing when visible, and final page numbers. Continuation pages must use the same plain-text style as the first Table of Contents page. Never duplicate a chapter title line.
- Table of Figures pages: use a diplomatic plain-text list, not markdown bullets and not markdown headings. Preserve each entry as "Figure N-M: Title ... page" or the closest visible punctuation. Preserve dot leaders or spacing when visible and final page numbers. Continuation pages must use the same style as the first Table of Figures page.
- Figures/diagrams outside list pages: transcribe any visible caption as text, then add exactly one marker on the next line: [FIGURE: concise description]. Do not use this marker for title pages, contents pages, or table-of-figures pages.
- Tables: preserve rows and columns as markdown tables when readable; otherwise use aligned plain text.

List-page checklist:
- If this page is a Table of Contents or Table of Figures page, do not use #, ##, ###, -, *, or numbered markdown-list syntax for the list.
- Keep one entry per visible row.
- Keep page numbers at the end of entries.
- Exclude the page's own footer number.
- If a page continues a list without a repeated heading, do not invent a heading. If the heading is visibly repeated, transcribe it once.

Quality checklist before final answer:
- No duplicated lines.
- No invented continuation notes.
- No footer page number included.
- Title page text is readable and not split by decorative visual wrapping.
- Contents/list pages are plain text and internally consistent.

Book ID: %s
Page number: %03d
Prompt version: %s
`, input.BookID, input.PageNumber, version)
}
```
