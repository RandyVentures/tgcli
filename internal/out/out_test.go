package out

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestWriteErrorText(t *testing.T) {
	var buf bytes.Buffer
	err := WriteError(&buf, false, errors.New("something failed"))
	if err != nil {
		t.Fatalf("WriteError: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "something failed") {
		t.Fatalf("expected error message in output, got %q", output)
	}
}

func TestWriteErrorJSON(t *testing.T) {
	var buf bytes.Buffer
	err := WriteError(&buf, true, errors.New("something failed"))
	if err != nil {
		t.Fatalf("WriteError: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON unmarshal: %v", err)
	}

	if result["error"] != "something failed" {
		t.Fatalf("expected error in JSON, got %v", result)
	}
}

func TestWriteErrorNil(t *testing.T) {
	var buf bytes.Buffer
	err := WriteError(&buf, false, nil)
	if err != nil {
		t.Fatalf("WriteError with nil: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no output for nil error")
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"name": "test",
		"count": 42,
	}

	err := WriteJSON(&buf, data)
	if err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON unmarshal: %v", err)
	}

	if result["name"] != "test" {
		t.Fatalf("expected name=test, got %v", result["name"])
	}
	if result["count"] != float64(42) {
		t.Fatalf("expected count=42, got %v", result["count"])
	}
}
