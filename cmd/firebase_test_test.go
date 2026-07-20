package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func firebaseTestSubgroup(name string) *cobra.Command {
	for _, c := range firebaseTestCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestFirebaseTestAndroidSubgroups(t *testing.T) {
	g := firebaseTestSubgroup("android")
	if g == nil {
		t.Fatal("firebase test android missing")
	}
	assertSubcommands(t, g, []string{"locales", "models", "run", "versions"})
}

func TestFirebaseTestIosSubgroups(t *testing.T) {
	g := firebaseTestSubgroup("ios")
	if g == nil {
		t.Fatal("firebase test ios missing")
	}
	assertSubcommands(t, g, []string{"locales", "models", "run", "versions"})
}

func TestFirebaseTestNetworkProfilesSubcommands(t *testing.T) {
	g := firebaseTestSubgroup("network-profiles")
	if g == nil {
		t.Fatal("firebase test network-profiles missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}
