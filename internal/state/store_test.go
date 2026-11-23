package state

import (
	"errors"
	"os"
	"testing"
)

func TestStoreRejectsTraversal(t *testing.T) {
	s := Store{Root: t.TempDir()}
	if _, err := s.Path("../escape"); err == nil {
		t.Fatal("accepted traversal")
	}
}

func TestLockExcludesSecondWriter(t *testing.T) {
	s := Store{Root: t.TempDir()}
	lock, err := s.Lock("demo")
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	if _, err := s.Lock("demo"); !errors.Is(err, ErrBusy) {
		t.Fatalf("want busy, got %v", err)
	}
}

func TestSavePersistsPrivateState(t *testing.T) {
	s := Store{Root: t.TempDir()}
	want := Record{Name: "demo", Phase: Created, MAC: "52:54:00:12:34:56"}
	if err := s.Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load("demo")
	if err != nil || got != want {
		t.Fatalf("got %+v, %v", got, err)
	}
	p, err := s.Path("demo")
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p + "/state.json")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("permissions %v", info.Mode())
	}
}
