package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRejectsInvalidState(t *testing.T) {
	for _, data := range []string{
		`{"name":"demo","phase":"unexpected","mac":"52:54:00:12:34:56"}`,
		`{"name":"demo","phase":"created","mac":"52:54:00:12:34:56"} {}`,
	} {
		t.Run(data, func(t *testing.T) {
			s := Store{Root: t.TempDir()}
			p, err := s.Path("demo")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(p, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(p, "state.json"), []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Load("demo"); err == nil {
				t.Fatal("invalid persisted state accepted")
			}
		})
	}
}
