package cmd

import "testing"

func TestIamWorkforcePoolsSubcommands(t *testing.T) {
	g := iamSubgroup("workforce-pools")
	if g == nil {
		t.Fatal("iam workforce-pools missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update", "undelete",
		"providers",
	})
	providers := findSub(g, "providers")
	if providers == nil {
		t.Fatal("iam workforce-pools providers missing")
	}
	assertSubcommands(t, providers, []string{
		"create", "delete", "describe", "list", "update", "undelete",
	})
}
