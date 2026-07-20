package cmd

import "testing"

func TestStorageFoldersSubcommands(t *testing.T) {
	g := storageSubgroup("folders")
	if g == nil {
		t.Fatal("storage folders missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "rename"})
}

func TestStorageHmacSubcommands(t *testing.T) {
	g := storageSubgroup("hmac")
	if g == nil {
		t.Fatal("storage hmac missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestStorageManagedFoldersSubcommands(t *testing.T) {
	g := storageSubgroup("managed-folders")
	if g == nil {
		t.Fatal("storage managed-folders missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list",
		"get-iam-policy", "set-iam-policy",
	})
}
