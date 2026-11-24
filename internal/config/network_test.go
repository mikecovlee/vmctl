package config

import (
	"strings"
	"testing"
)

func TestNetworkModeCompatibility(t *testing.T) {
	for _, mode := range []string{"nat", "egress", "isolated", "invalid"} {
		t.Run(mode, func(t *testing.T) {
			s, err := Parse(strings.NewReader(`{"apiVersion":"v1alpha1","name":"demo","resources":{"cpu":1,"memoryMiB":512,"diskGiB":2},"media":{"iso":"x.iso"},"network":{"mode":"` + mode + `"}}`))
			if mode == "invalid" {
				if err == nil {
					t.Fatal("invalid mode accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := mode
			if mode == "egress" {
				want = "nat"
			}
			if s.Network.Mode != want {
				t.Fatalf("got %s want %s", s.Network.Mode, want)
			}
		})
	}
}
