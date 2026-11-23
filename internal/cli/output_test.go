package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestOutputUsesReadableFormattingByDefault(t *testing.T) {
	command := New()
	var output bytes.Buffer
	command.SetOut(&output)
	if err := emit(command, struct {
		Name string `json:"name"`
	}{"demo"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "\n  \"name\"") {
		t.Fatalf("expected readable indentation: %q", output.String())
	}
}

func TestJSONOutputUsesOneRecordPerLine(t *testing.T) {
	command := New()
	if err := command.PersistentFlags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	command.SetOut(&output)
	if err := emit(command, struct {
		Name string `json:"name"`
	}{"demo"}); err != nil {
		t.Fatal(err)
	}
	if output.String() != "{\"name\":\"demo\"}\n" {
		t.Fatalf("unexpected JSON record: %q", output.String())
	}
}
