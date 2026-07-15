package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func idsSubgroup(name string) *cobra.Command {
	for _, c := range idsCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestIDSHasEndpointsSubgroup(t *testing.T) {
	if idsSubgroup("endpoints") == nil {
		t.Fatal("ids missing endpoints subgroup")
	}
}

func TestIDSEndpointsSubcommands(t *testing.T) {
	g := idsSubgroup("endpoints")
	if g == nil {
		t.Fatal("endpoints missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestIDSSeverityEnum(t *testing.T) {
	cases := map[string]string{
		"":              "",
		"low":           "LOW",
		"HIGH":          "HIGH",
		"informational": "INFORMATIONAL",
	}
	for in, want := range cases {
		got, err := idsSeverityEnum(in)
		if err != nil {
			t.Errorf("idsSeverityEnum(%q) unexpected error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("idsSeverityEnum(%q) = %q, want %q", in, got, want)
		}
	}
	if _, err := idsSeverityEnum("catastrophic"); err == nil {
		t.Errorf("idsSeverityEnum(\"catastrophic\") expected error")
	}
}
