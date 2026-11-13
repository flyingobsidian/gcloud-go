package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// `auth login --cred-file` saves the account into the invocation's gcloud
// configuration, not the process's: two invocations with different
// CLOUDSDK_CONFIG directories end up with different accounts.
func TestAuthLoginCredFileSavesInvocationConfig(t *testing.T) {
	t.Setenv("CLOUDSDK_CONFIG", t.TempDir()) // the process's, which must stay untouched

	for _, account := range []string{"a@p.iam.gserviceaccount.com", "b@p.iam.gserviceaccount.com"} {
		dir := t.TempDir()
		key := filepath.Join(t.TempDir(), "key.json")
		if err := os.WriteFile(key, []byte(`{"type":"service_account","client_email":"`+account+`"}`), 0600); err != nil {
			t.Fatal(err)
		}
		inv, _, _ := newTestInvocation(t, map[string]string{"CLOUDSDK_CONFIG": dir}, "")
		if err := runInvocation(inv, "auth", "login", "--cred-file", key, "--brief"); err != nil {
			t.Fatalf("auth login for %s: %v", account, err)
		}
		got, err := inv.env.Property("core", "account")
		if err != nil {
			t.Fatal(err)
		}
		if got != account {
			t.Errorf("core/account in the invocation's config = %q, want %q", got, account)
		}
	}

	if entries, _ := os.ReadDir(os.Getenv("CLOUDSDK_CONFIG")); len(entries) != 0 {
		t.Errorf("the process's config directory was written to: %v", entries)
	}
}
