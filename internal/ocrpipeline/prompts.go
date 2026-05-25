package ocrpipeline

import "fmt"

const StructuredOCRSchemaVersion = "structured-ocr/v1"

const StructuredOCRSystemPrompt = `You are a precise structured OCR engine for scanned technical books.
Return strict JSON only. Do not return Markdown, commentary, or code fences.`

func RenderStructuredOCRPrompt(input StructuredOCRInput) string {
	return fmt.Sprintf(`Transcribe exactly one target page image into structured OCR JSON.

Output contract:
- Return only strict JSON matching schema_version %q.
- The root object must include schema_version, book_id, page_number, page_type, and blocks.
- Transcribe only visible content from the target page image.
- Do not infer text from neighboring pages or from general knowledge.
- Exclude standalone running page numbers, scanner borders, and footer folios unless they are semantically part of the page.
- Preserve historical terminology exactly when visible, including "data base", "PSBase", "PPSCalc", "Dired", "Steamer", "Zmacs", and "Xerox Star".

Block contract:
- Use type "heading" for visible headings and set level from 1 to 6.
- Use type "paragraph" for prose.
- Use type "list" with items for bullet/numbered lists.
- Use type "table" with table.headers and table.rows for visible tabular content.
- Use type "figure" for page-local figures, diagrams, charts, or screenshots; include caption, description, and optional diagram_text.
- Use type "footnote" for footnotes.
- Use type "page_footer" only for visible footer/running-page artifacts that should be tracked but not rendered by default.
- Use type "blank" only when the page is blank.

Table rules:
- Every visible table or spreadsheet-like grid must be represented as a table block.
- Preserve row and column structure.
- If column labels are visible, put them in table.headers.
- If no column labels are visible, leave headers empty and put all visible rows in table.rows.
- Preserve formulas, identifiers, punctuation, and numeric values exactly.
- Do not render aligned plain text tables inside paragraph blocks.
- Do not put table rows only in diagram_text. A grid with rows/columns must produce a table block.
- If a figure contains a visible table/grid, emit a figure block for the caption/description and then an adjacent table block for the grid contents.

Figure rules:
- Create a figure block only for a figure visibly present on the target page.
- If a visible figure has a caption, copy that caption exactly into the figure block's caption field.
- Do not emit an empty figure block when a visible caption exists; caption is required for captioned figures.
- Prose references such as "as shown in Figure 1-1" are not figure blocks.
- Table of Figures entries are list/text entries, not figure blocks.
- Put long non-tabular diagram label transcriptions in diagram_text, not in paragraph text.
- Do not use diagram_text as a substitute for table blocks when a diagram is a spreadsheet/table/grid.

Example table block for a spreadsheet figure:
{
  "id": "p032-t001",
  "type": "table",
  "table": {
    "headers": ["", "A", "B", "C"],
    "rows": [["1", "100", "20", "A1*B1"], ["2", "75", "5", "A2*B2"], ["3", "", "", "C1+C2"]]
  }
}

Target metadata:
book_id: %s
page_number: %03d
schema_version: %s
`, StructuredOCRSchemaVersion, input.BookID, input.PageNumber, StructuredOCRSchemaVersion)
}
