package cmd

import (
	"testing"
)

func TestIamOauthClientsSubcommands(t *testing.T) {
	g := iamSubgroup("oauth-clients")
	if g == nil {
		t.Fatal("iam oauth-clients missing")
	}
	assertSubcommands(t, g, []string{
		"create", "credentials", "delete", "describe", "list", "undelete", "update",
	})
}

func TestIamOauthClientCredentialsSubcommands(t *testing.T) {
	g := iamSubgroup("oauth-clients")
	if g == nil {
		t.Fatal("iam oauth-clients missing")
	}
	creds := findSub(g, "credentials")
	if creds == nil {
		t.Fatal("iam oauth-clients credentials missing")
	}
	assertSubcommands(t, creds, []string{"create", "delete", "describe", "list", "update"})
}
