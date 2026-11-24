package config

import (
	"strings"
	"testing"
)

func TestRejectsIPv6ManagementBindings(t *testing.T) {
	for _, binding := range []string{`"network":{"forwards":[{"bind":"::1","host":2222,"guest":22}]}`, `"console":{"vncBind":"::1"}`, `"network":{"forwards":[{"bind":"::ffff:127.0.0.1","host":2222,"guest":22}]}`, `"console":{"vncBind":"::ffff:127.0.0.1"}`} {
		t.Run(binding, func(t *testing.T) {
			input := `{"apiVersion":"v1alpha1","name":"demo","resources":{"cpu":1,"memoryMiB":512,"diskGiB":2},"media":{"iso":"test.iso"},` + binding + `}`
			if _, err := Parse(strings.NewReader(input)); err == nil {
				t.Fatal("accepted IPv6 management binding")
			}
		})
	}
}
