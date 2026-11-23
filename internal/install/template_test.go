package install

import (
	"strings"
	"testing"
)

func TestRenderRejectsMissingTemplateField(t *testing.T) {
	_, err := Render("{{.Unknown}}", Identity{Hostname: "demo"})
	if err == nil {
		t.Fatal("missing template field accepted")
	}
}

func TestRenderInjectsInstanceIdentity(t *testing.T) {
	result, err := Render("host={{.Hostname}} mac={{.MAC}} key={{.PublicKey}}", Identity{Hostname: "demo", MAC: "52:54:00:11:22:33", PublicKey: "ssh-ed25519 fixture"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "host=demo mac=52:54:00:11:22:33 key=ssh-ed25519 fixture") {
		t.Fatal(result)
	}
}
