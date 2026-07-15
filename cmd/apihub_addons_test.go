package cmd

import "testing"

func TestApihubHasAddonsSubgroup(t *testing.T) {
	if apihubSubgroup("addons") == nil {
		t.Fatal("apihub missing addons subgroup")
	}
}

func TestApihubAddonsSubcommands(t *testing.T) {
	g := apihubSubgroup("addons")
	if g == nil {
		t.Fatal("addons missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "manage-config"})
}
