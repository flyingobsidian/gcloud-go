package cmd

import "testing"

func TestIamWorkforcePoolsSubcommands(t *testing.T) {
	g := iamSubgroup("workforce-pools")
	if g == nil {
		t.Fatal("iam workforce-pools missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update", "undelete",
		"providers", "subjects", "create-cred-config", "create-login-config",
	})
	providers := findSub(g, "providers")
	if providers == nil {
		t.Fatal("iam workforce-pools providers missing")
	}
	assertSubcommands(t, providers, []string{
		"create", "delete", "describe", "list", "update", "undelete",
	})
}

func TestIamWorkforcePoolsSubjectsSubcommands(t *testing.T) {
	subjects := findSub(iamSubgroup("workforce-pools"), "subjects")
	if subjects == nil {
		t.Fatal("iam workforce-pools subjects missing")
	}
	assertSubcommands(t, subjects, []string{"delete", "undelete", "revoke-sessions"})
}

func TestIamWpSubjectName(t *testing.T) {
	name, id, err := iamWpSubjectName("alice", "global", "my-pool")
	if err != nil || name != "locations/global/workforcePools/my-pool/subjects/alice" || id != "alice" {
		t.Errorf("got %q %q %v", name, id, err)
	}
	full := "locations/global/workforcePools/p/subjects/bob"
	name, id, err = iamWpSubjectName(full, "", "")
	if err != nil || name != full || id != "bob" {
		t.Errorf("got %q %q %v", name, id, err)
	}
	for _, tc := range [][3]string{
		{"alice", "", "my-pool"},
		{"alice", "global", ""},
		{"locations/global/workforcePools/p", "", ""},
	} {
		if _, _, err := iamWpSubjectName(tc[0], tc[1], tc[2]); err == nil {
			t.Errorf("iamWpSubjectName(%q, %q, %q): expected error", tc[0], tc[1], tc[2])
		}
	}
}
