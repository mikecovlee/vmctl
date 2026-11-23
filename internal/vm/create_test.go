package vm

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"vmctl/internal/config"
	"vmctl/internal/docker"
	"vmctl/internal/state"
)

func TestCreatePersistsIndependentIdentity(t *testing.T) {
	// Given a private store and a local installation medium.
	root := t.TempDir()
	iso := filepath.Join(root, "os.iso")
	if err := os.WriteFile(iso, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	m := Manager{Store: state.Store{Root: filepath.Join(root, "state")}, Docker: docker.Client{Binary: "/bin/true"}}
	s := config.Spec{Name: "demo", Media: config.Media{ISO: iso}}
	// When an instance is registered without starting a VM.
	r, err := m.Create(context.Background(), s)
	// Then identity and private credentials survive reload.
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := m.Store.Load("demo")
	if err != nil || loaded.MAC != r.MAC || loaded.Phase != state.Created {
		t.Fatalf("record: %+v, %v", loaded, err)
	}
	dir, err := m.Store.Path("demo")
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dir, "credentials.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatal("credentials must be private")
	}
}

func TestCreateSeparatesInstanceCredentials(t *testing.T) {
	root := t.TempDir()
	iso := filepath.Join(root, "os.iso")
	if err := os.WriteFile(iso, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	m := Manager{Store: state.Store{Root: filepath.Join(root, "state")}, Docker: docker.Client{Binary: "/bin/true"}}
	identities := make(map[string]bool)
	passwords := make(map[string]bool)
	for _, name := range []string{"first", "second"} {
		record, err := m.Create(context.Background(), config.Spec{Name: name, Media: config.Media{ISO: iso}})
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(m.Store.Root, name, "credentials.json"))
		if err != nil {
			t.Fatal(err)
		}
		var credentials Credentials
		if err := json.Unmarshal(data, &credentials); err != nil {
			t.Fatal(err)
		}
		if record.MAC == "" || identities[record.MAC] {
			t.Fatal("instances share or lack a MAC identity")
		}
		if credentials.VNCPassword == "" || passwords[credentials.VNCPassword] {
			t.Fatal("instances share or lack a console credential")
		}
		identities[record.MAC] = true
		passwords[credentials.VNCPassword] = true
	}
}

func TestCreateRefusesExistingInstance(t *testing.T) {
	// Given an existing registered instance.
	root := t.TempDir()
	iso := filepath.Join(root, "os.iso")
	if err := os.WriteFile(iso, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	m := Manager{Store: state.Store{Root: filepath.Join(root, "state")}, Docker: docker.Client{Binary: "/bin/true"}}
	s := config.Spec{Name: "demo", Media: config.Media{ISO: iso}}
	if _, err := m.Create(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	// When the same name is created again.
	_, err := m.Create(context.Background(), s)
	// Then existing data is not replaced.
	if err == nil {
		t.Fatal("duplicate accepted")
	}
}

func TestCreateRollsBackWhenDiskCreationFails(t *testing.T) {
	root := t.TempDir()
	iso := filepath.Join(root, "os.iso")
	if err := os.WriteFile(iso, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	m := Manager{Store: state.Store{Root: filepath.Join(root, "state")}, Docker: docker.Client{Binary: "/bin/false"}}
	_, err := m.Create(context.Background(), config.Spec{Name: "demo", Media: config.Media{ISO: iso}})
	if err == nil {
		t.Fatal("disk failure accepted as successful creation")
	}
	if _, err := os.Stat(filepath.Join(m.Store.Root, "demo")); !os.IsNotExist(err) {
		t.Fatalf("partial instance survived failed creation: %v", err)
	}
}
