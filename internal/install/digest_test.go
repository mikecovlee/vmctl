package install

import (
	"strings"
	"testing"
)

func TestDigestChangesWhenMediaChanges(t *testing.T) {
	a, err := Digest(strings.NewReader("first ISO"), "template")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Digest(strings.NewReader("second ISO"), "template")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("changed media reused the same installation digest")
	}
}
