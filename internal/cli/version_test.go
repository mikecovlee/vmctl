package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func TestVersionReportsBuildIdentityWithoutRuntime(t *testing.T) {
	// Given a CLI without Docker available.
	t.Setenv("PATH", t.TempDir())
	var output bytes.Buffer
	command := New()
	command.SetOut(&output)
	command.SetArgs([]string{"version", "--json"})
	// When querying its build identity.
	if err := command.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Then the public result identifies the build and supported contracts.
	var result struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
		Config  string `json:"configVersion"`
		Bundle  int    `json:"bundleVersion"`
	}
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Version == "" || result.Commit == "" || result.Config != "v1alpha1" || result.Bundle != 1 {
		t.Fatalf("incomplete build identity: %+v", result)
	}
}
