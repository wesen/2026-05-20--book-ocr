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
- Use type "code" for visible program code, command transcripts, formal grammar, equations laid out as code, or monospaced/listing-like text where line breaks and indentation are semantically important.
- Use type "figure" only for page-local visual material that cannot be represented primarily as text, code, list, or table; include caption, description, and optional diagram_text.
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
- If a captioned figure region is primarily a visible table/grid/spreadsheet, emit the caption as a heading or paragraph and then an adjacent table block for the grid contents; do not emit a figure block for the same table unless there is non-tabular visual content that must remain as an image.

Code/listing rules:
- Program code, command listings, formal text examples, UI text transcripts, quoted text shown inside screenshots, and monospaced blocks are OCR text, not figure images.
- Preserve code/listing line breaks, indentation, punctuation, arrows, braces, quotes, and identifiers exactly in a code block.
- A code block must include its recognized content in the text field. Never emit an empty code block for a non-empty visible listing.
- If a screenshot or boxed region is mostly readable text/code/table content, transcribe that content as paragraph/list/code/table blocks instead of hiding it in a figure description.
- Do not use a figure block merely because the source is a screenshot, boxed display, or page image; classify by the content that must be readable in the final Markdown.

Figure rules:
- Create a figure block only for a figure visibly present on the target page and only for visual content that remains meaningful as an image after all readable text/code/tables have been transcribed.
- If a visible figure has a caption, copy that caption exactly into the figure block's caption field.
- Do not emit an empty figure block when a visible caption exists; caption is required for captioned figures.
- Prose references such as "as shown in Figure 1-1" are not figure blocks.
- Table of Figures entries are list/text entries, not figure blocks.
- Put long non-tabular diagram label transcriptions in diagram_text only when the region is truly diagrammatic.
- Do not use diagram_text as a substitute for paragraph, code, list, or table blocks.
- A page containing a captioned screenshot may have both: a figure block for the screenshot/caption and separate paragraph/list/code/table blocks for readable text inside the screenshot.
- If a screenshot contains a readable scrollable text window, transcribe the visible text as paragraph/list/code blocks. Do not replace the readable window with only a screenshot description.
- If you cannot read small screenshot text reliably, include a warning on the figure block instead of inventing text.

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
