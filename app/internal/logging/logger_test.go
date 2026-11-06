package logging

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestNewAddsBuildMetadataToEveryRecord(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := New(Options{
		Service: "kargo-pattern",
		Version: "1.2.3",
		Commit:  "abc1234",
		Level:   "debug",
		Output:  &output,
	})

	logger.Info("hello", "answer", 42)

	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("decode log record: %v", err)
	}

	if got := record["service"]; got != "kargo-pattern" {
		t.Fatalf("service = %v, want kargo-pattern", got)
	}
	if got := record["version"]; got != "1.2.3" {
		t.Fatalf("version = %v, want 1.2.3", got)
	}
	if got := record["commit"]; got != "abc1234" {
		t.Fatalf("commit = %v, want abc1234", got)
	}
	if got := record["msg"]; got != "hello" {
		t.Fatalf("msg = %v, want hello", got)
	}
	if got := record["answer"]; got != float64(42) {
		t.Fatalf("answer = %v, want 42", got)
	}
}
