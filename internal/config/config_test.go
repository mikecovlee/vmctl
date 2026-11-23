package config

import (
	"strings"
	"testing"
)

func TestParseRejectsInvalidContracts(t *testing.T) {
	// Given malformed or unsupported public configuration.
	for _, input := range []string{`{}`, `{"apiVersion":"v1alpha1","name":"../escape"}`, `{"unknown":true}`, `{} {}`} {
		t.Run(input, func(t *testing.T) {
			// When parsing the boundary.
			_, err := Parse(strings.NewReader(input))
			// Then invalid input cannot enter the domain.
			if err == nil {
				t.Fatal("accepted invalid configuration")
			}
		})
	}
}

func TestParseAppliesIsolatedDefaults(t *testing.T) {
	// Given a minimal supported manual installation.
	input := `{"apiVersion":"v1alpha1","name":"demo","resources":{"cpu":2,"memoryMiB":2048,"diskGiB":20},"media":{"iso":"installer.iso"}}`
	// When parsing the configuration.
	spec, err := Parse(strings.NewReader(input))
	// Then networking and console access default to isolation and loopback.
	if err != nil {
		t.Fatal(err)
	}
	if spec.Network.Mode != "isolated" || spec.Console.VNCBind != "127.0.0.1" {
		t.Fatalf("unsafe defaults: %+v", spec)
	}
}
