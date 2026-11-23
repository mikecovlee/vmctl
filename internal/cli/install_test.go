package cli

import "testing"

func TestInstallExposesWaitOption(t *testing.T) {
	root := New()
	command, _, err := root.Find([]string{"install"})
	if err != nil {
		t.Fatal(err)
	}
	if command.Flags().Lookup("wait") == nil {
		t.Fatal("missing bounded installation wait option")
	}
}
