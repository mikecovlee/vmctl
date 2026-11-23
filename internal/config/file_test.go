package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadResolvesMediaRelativeToConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vm.json")
	input := `{"apiVersion":"v1alpha1","name":"demo","resources":{"cpu":2,"memoryMiB":2048,"diskGiB":20},"media":{"iso":"install.iso"}}`
	if err := os.WriteFile(path, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	spec, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Media.ISO != filepath.Join(dir, "install.iso") {
		t.Fatalf("relative path unresolved: %s", spec.Media.ISO)
	}
}
