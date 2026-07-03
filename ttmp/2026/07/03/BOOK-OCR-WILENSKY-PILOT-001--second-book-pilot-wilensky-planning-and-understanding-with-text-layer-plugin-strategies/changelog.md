# Changelog

## 2026-07-03

- Initial workspace created


## 2026-07-03

Ran the second-book pilot: Wilensky pages 1-24 through three strategies (textlayer plugin 0 calls, pure VLM, hybrid draft-correction; all 24/24). Findings W1-W5: text layer competitive on prose for free; renderer lacks markdown escaping (pandoc failure); prompt.render contract too thin (hybrid schema drift with lines field); heading heuristics belong in profile; --plugin lacks args. Intern guide written.


## 2026-07-03

Added design doc 02: user-facing hardening from the pilot's observed frictions (trustworthy quality signals incl. text-layer oracle, cause-and-remedy errors, workspace defaults, compare command, progress/cost) and the agent-first design (JSON mode then MCP, run manifest, plugin authoring loop with plugin_check + op JSON Schemas, sandbox + provenance pinning for agent-written plugins).


## 2026-07-03

Structure sample (book pp.35/37/46 -> PDF 55/59/68): PANDORA code and explanation flowchart handled ideally by the VLM incl. full figure->crop->embed chain; task-network diagram misclassified as code losing its image (W6); running header leaked as heading (W7); textlayer structure-blind as predicted. Addendum + findings W6-W8 added to design doc 01.

