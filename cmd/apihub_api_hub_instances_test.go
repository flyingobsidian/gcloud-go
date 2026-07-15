package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func apihubSubgroup(name string) *cobra.Command {
	for _, c := range apihubCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestApihubHasApiHubInstancesSubgroup(t *testing.T) {
	if apihubSubgroup("api-hub-instances") == nil {
		t.Fatal("apihub missing api-hub-instances subgroup")
	}
}

func TestApihubApiHubInstancesSubcommands(t *testing.T) {
	g := apihubSubgroup("api-hub-instances")
	if g == nil {
		t.Fatal("api-hub-instances missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "lookup"})
}

func TestAhiEncryptionEnum(t *testing.T) {
	cases := map[string]string{
		"":     "",
		"gmek": "GMEK",
		"CMEK": "CMEK",
	}
	for in, want := range cases {
		got, err := ahiEncryptionEnum(in)
		if err != nil {
			t.Errorf("ahiEncryptionEnum(%q) unexpected error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ahiEncryptionEnum(%q) = %q, want %q", in, got, want)
		}
	}
	if _, err := ahiEncryptionEnum("aes256"); err == nil {
		t.Errorf("ahiEncryptionEnum(\"aes256\") expected an error")
	}
}
