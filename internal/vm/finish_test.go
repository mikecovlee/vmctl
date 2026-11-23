package vm

import (
	"context"
	"errors"
	"testing"

	"vmctl/internal/state"
)

func TestFinishRejectsUnstartedInstallation(t *testing.T) {
	m := Manager{Store: state.Store{Root: t.TempDir()}}
	if err := m.Store.Save(state.Record{Name: "demo", Phase: state.Created}); err != nil {
		t.Fatal(err)
	}
	if err := m.FinishInstall(context.Background(), "demo"); !errors.Is(err, ErrState) {
		t.Fatalf("got %v, want invalid transition", err)
	}
}

func TestFinishRecordsUserDeclarationWithoutVerification(t *testing.T) {
	m := Manager{Store: state.Store{Root: t.TempDir()}}
	if err := m.Store.Save(state.Record{Name: "demo", Phase: state.Installing}); err != nil {
		t.Fatal(err)
	}
	if err := m.FinishInstall(context.Background(), "demo"); err != nil {
		t.Fatal(err)
	}
	record, err := m.Store.Load("demo")
	if err != nil {
		t.Fatal(err)
	}
	if record.Phase != state.Installed {
		t.Fatalf("got phase %s", record.Phase)
	}
}
