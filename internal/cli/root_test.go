package cli

import (
	"bytes"
	"context"
	"testing"
)

func TestValidateRejectsMissingFile(t *testing.T) {
	var output bytes.Buffer
	command := New()
	command.SetOut(&output)
	command.SetArgs([]string{"validate", "/nonexistent/vmctl-config.json"})
	if err := command.ExecuteContext(context.Background()); err == nil {
		t.Fatal("missing config accepted")
	}
}

func TestListEmptyStateProducesJSONArray(t *testing.T) {
	var output bytes.Buffer
	command := New()
	command.SetOut(&output)
	command.SetArgs([]string{"--state-dir", t.TempDir(), "--json", "list"})
	if err := command.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if output.String() != "[]\n" {
		t.Fatalf("unexpected list: %s", output.String())
	}
}
