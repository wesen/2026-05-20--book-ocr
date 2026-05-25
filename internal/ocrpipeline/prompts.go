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
- Every visible table must be represented as a table block.
- Preserve row and column structure.
- If column labels are visible, put them in table.headers.
- If no column labels are visible, leave headers empty and put all visible rows in table.rows.
- Preserve formulas, identifiers, punctuation, and numeric values exactly.
- Do not render aligned plain text tables inside paragraph blocks.

Figure rules:
- Create a figure block only for a figure visibly present on the target page.
- Prose references such as "as shown in Figure 1-1" are not figure blocks.
- Table of Figures entries are list/text entries, not figure blocks.
- Put long diagram label transcriptions in diagram_text, not in paragraph text.

Target metadata:
book_id: %s
page_number: %03d
schema_version: %s
`, StructuredOCRSchemaVersion, input.BookID, input.PageNumber, StructuredOCRSchemaVersion)
}
