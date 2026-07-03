// Experiment 02: minimal host-side NDJSON-stdio plugin client for book-ocr.
//
// This is the shape of the code that would live in internal/plugin (or
// pkg/bookocr/plugin) — spawn a plugin process, validate its handshake,
// correlate request/response frames by request_id, surface event frames,
// and map protocol errors. Stdlib only, runnable standalone:
//
//	go run host.go <plugin-command> <page-image-path>
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type frame struct {
	Type            string          `json:"type"`
	ProtocolVersion string          `json:"protocol_version,omitempty"`
	PluginName      string          `json:"plugin_name,omitempty"`
	Capabilities    *capabilities   `json:"capabilities,omitempty"`
	RequestID       string          `json:"request_id,omitempty"`
	OK              *bool           `json:"ok,omitempty"`
	Op              string          `json:"op,omitempty"`
	Event           string          `json:"event,omitempty"`
	Data            json.RawMessage `json:"data,omitempty"`
	Output          json.RawMessage `json:"output,omitempty"`
	Error           *pluginError    `json:"error,omitempty"`
}

type capabilities struct {
	Ops []string `json:"ops"`
}

type pluginError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

type request struct {
	Type      string         `json:"type"`
	RequestID string         `json:"request_id"`
	Op        string         `json:"op"`
	Ctx       map[string]any `json:"ctx"`
	Input     map[string]any `json:"input"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: host <plugin-command> <page-image>")
		os.Exit(2)
	}
	pluginCmd, imagePath := os.Args[1], os.Args[2]

	cmd := exec.Command(pluginCmd)
	cmd.Stderr = os.Stderr // plugin logs pass through
	stdin, err := cmd.StdinPipe()
	must(err)
	stdout, err := cmd.StdoutPipe()
	must(err)
	must(cmd.Start())

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024) // structured pages can be large

	// 1. Handshake must be the first frame.
	hs := readFrame(scanner)
	if hs.Type != "handshake" || hs.ProtocolVersion != "v2" {
		fatal("bad handshake: %+v", hs)
	}
	fmt.Printf("handshake ok: plugin=%s ops=%v\n", hs.PluginName, hs.Capabilities.Ops)

	// 2. Capability check before dispatch (host decides fallback if absent).
	if !supports(hs.Capabilities, "ocr.page") {
		fatal("plugin does not support ocr.page")
	}

	send := func(req request) {
		b, err := json.Marshal(req)
		must(err)
		_, err = stdin.Write(append(b, '\n'))
		must(err)
	}

	// 3. prompt.render round-trip.
	send(request{Type: "request", RequestID: "r-001", Op: "prompt.render",
		Ctx:   map[string]any{"deadline_ms": time.Now().Add(10 * time.Second).UnixMilli(), "dry_run": false},
		Input: map[string]any{"book_id": "demo-book", "page_number": 12}})
	resp := awaitResponse(scanner, "r-001")
	fmt.Printf("prompt.render ok: %s\n", compact(resp.Output))

	// 4. ocr.page round-trip (events may interleave before the response).
	send(request{Type: "request", RequestID: "r-002", Op: "ocr.page",
		Ctx: map[string]any{"deadline_ms": time.Now().Add(60 * time.Second).UnixMilli(), "dry_run": false},
		Input: map[string]any{"book_id": "demo-book", "page_number": 12,
			"image_path": imagePath}})
	resp = awaitResponse(scanner, "r-002")

	var page struct {
		SchemaVersion string `json:"schema_version"`
		PageNumber    int    `json:"page_number"`
		Blocks        []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"blocks"`
	}
	must(json.Unmarshal(resp.Output, &page))
	if page.SchemaVersion != "structured-ocr/v1" || len(page.Blocks) == 0 {
		fatal("plugin output failed host-side contract check: %s", compact(resp.Output))
	}
	fmt.Printf("ocr.page ok: page=%d blocks=%d text=%q\n",
		page.PageNumber, len(page.Blocks), page.Blocks[0].Text)

	// 5. Unsupported op must produce E_UNSUPPORTED, not a hang.
	send(request{Type: "request", RequestID: "r-003", Op: "figures.segment",
		Ctx: map[string]any{}, Input: map[string]any{}})
	fr := awaitTerminal(scanner, "r-003")
	if fr.Error == nil || fr.Error.Code != "E_UNSUPPORTED" {
		fatal("expected E_UNSUPPORTED, got %+v", fr)
	}
	fmt.Println("unsupported-op handling ok: E_UNSUPPORTED")

	must(stdin.Close())
	must(cmd.Wait())
	fmt.Println("PLUGIN_PROTOCOL_DEMO_OK")
}

func supports(c *capabilities, op string) bool {
	if c == nil {
		return false
	}
	for _, o := range c.Ops {
		if o == op {
			return true
		}
	}
	return false
}

func readFrame(s *bufio.Scanner) frame {
	for s.Scan() {
		line := s.Bytes()
		if len(line) == 0 {
			continue
		}
		var f frame
		if err := json.Unmarshal(line, &f); err != nil {
			fatal("stdout contamination (non-JSON line): %q", line)
		}
		return f
	}
	fatal("plugin closed stdout: %v", s.Err())
	return frame{}
}

// awaitResponse drains event frames and returns the ok response for id.
func awaitResponse(s *bufio.Scanner, id string) frame {
	f := awaitTerminal(s, id)
	if f.OK == nil || !*f.OK {
		fatal("request %s failed: %+v", id, f.Error)
	}
	return f
}

func awaitTerminal(s *bufio.Scanner, id string) frame {
	for {
		f := readFrame(s)
		switch f.Type {
		case "event":
			fmt.Printf("  event(%s): %s %s\n", f.RequestID, f.Event, compact(f.Data))
		case "response":
			if f.RequestID != id {
				fatal("response correlation mismatch: want %s got %s", id, f.RequestID)
			}
			return f
		default:
			fatal("unexpected frame type %q", f.Type)
		}
	}
}

func compact(r json.RawMessage) string {
	if len(r) > 120 {
		return string(r[:120]) + "…"
	}
	return string(r)
}

func must(err error) {
	if err != nil {
		fatal("%v", err)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "host: "+format+"\n", args...)
	os.Exit(1)
}
