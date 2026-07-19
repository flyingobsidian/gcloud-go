package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func bmsSubgroup(name string) *cobra.Command {
	for _, c := range bmsCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestBmsInstancesSubcommands(t *testing.T) {
	g := bmsSubgroup("instances")
	if g == nil {
		t.Fatal("bms instances missing")
	}
	assertSubcommands(t, g, []string{
		"describe", "list", "update", "reset", "start", "stop",
		"enable-interactive-serial-console", "disable-interactive-serial-console", "rename",
	})
}

func TestBmsNetworksSubcommands(t *testing.T) {
	g := bmsSubgroup("networks")
	if g == nil {
		t.Fatal("bms networks missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "update", "rename"})
}

func TestBmsNfsSharesSubcommands(t *testing.T) {
	g := bmsSubgroup("nfs-shares")
	if g == nil {
		t.Fatal("bms nfs-shares missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "update", "rename"})
}

func TestBmsOperationsSubcommands(t *testing.T) {
	g := bmsSubgroup("operations")
	if g == nil {
		t.Fatal("bms operations missing")
	}
	assertSubcommands(t, g, []string{"describe"})
}

func TestBmsOsImagesSubcommands(t *testing.T) {
	g := bmsSubgroup("os-images")
	if g == nil {
		t.Fatal("bms os-images missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestBmsSshKeysSubcommands(t *testing.T) {
	g := bmsSubgroup("ssh-keys")
	if g == nil {
		t.Fatal("bms ssh-keys missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "list"})
}

func TestBmsVolumesSubcommands(t *testing.T) {
	g := bmsSubgroup("volumes")
	if g == nil {
		t.Fatal("bms volumes missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "update", "rename", "resize"})
}
