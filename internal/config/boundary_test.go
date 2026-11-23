package config

import (
	"strings"
	"testing"
)

func TestParseDefaultsForwardBind(t *testing.T) {
	input := `{"apiVersion":"v1alpha1","name":"demo","resources":{"cpu":2,"memoryMiB":2048,"diskGiB":20},"media":{"iso":"installer.iso"},"network":{"forwards":[{"host":2222,"guest":22}]}}`
	spec, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if spec.Network.Forwards[0].Bind != "127.0.0.1" {
		t.Fatal("forward must default to loopback")
	}
}

func TestParseRejectsOversizedInput(t *testing.T) {
	input := `{"apiVersion":"v1alpha1","name":"demo","resources":{"cpu":2,"memoryMiB":2048,"diskGiB":20},"media":{"iso":"installer.iso"}}` + strings.Repeat(" ", 1<<20)
	if _, err := Parse(strings.NewReader(input)); err == nil {
		t.Fatal("oversized configuration accepted")
	}
}
