package auth

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestStoreAndLoadJSON(t *testing.T) {
	dir := t.TempDir()
	store := &CredentialStore{configDir: dir}

	// Create a fake service account JSON file.
	cred := map[string]any{
		"type":            "service_account",
		"project_id":      "test-project",
		"private_key_id":  "key123",
		"private_key":     "-----BEGIN RSA PRIVATE KEY-----\nfake\n-----END RSA PRIVATE KEY-----\n",
		"client_email":    "test@test-project.iam.gserviceaccount.com",
		"client_id":       "12345",
		"token_uri":       "https://oauth2.googleapis.com/token",
	}
	data, err := json.Marshal(cred)
	if err != nil {
		t.Fatal(err)
	}

	credFile := filepath.Join(t.TempDir(), "sa.json")
	if err := os.WriteFile(credFile, data, 0600); err != nil {
		t.Fatal(err)
	}

	// Store.
	account, err := store.Store(credFile)
	if err != nil {
		t.Fatalf("Store() error: %v", err)
	}
	if account != "test@test-project.iam.gserviceaccount.com" {
		t.Errorf("Store() account = %q, want %q", account, "test@test-project.iam.gserviceaccount.com")
	}

	// Load (should find it in SQLite or JSON fallback).
	loaded, err := store.Load(account)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	var loadedCred map[string]any
	if err := json.Unmarshal(loaded, &loadedCred); err != nil {
		t.Fatal(err)
	}
	if loadedCred["client_email"] != cred["client_email"] {
		t.Errorf("loaded client_email = %v, want %v", loadedCred["client_email"], cred["client_email"])
	}
}

func TestLoadFromSQLite(t *testing.T) {
	dir := t.TempDir()
	store := &CredentialStore{configDir: dir}

	// Create a SQLite credentials.db like gcloud does.
	dbPath := filepath.Join(dir, "credentials.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE credentials (account_id TEXT PRIMARY KEY, value BLOB)`)
	if err != nil {
		t.Fatal(err)
	}

	credJSON := `{"type":"authorized_user","client_id":"abc.apps.googleusercontent.com","client_secret":"secret","refresh_token":"1//refresh","token_uri":"https://oauth2.googleapis.com/token"}`
	_, err = db.Exec(`INSERT INTO credentials (account_id, value) VALUES (?, ?)`,
		"user@example.com", credJSON)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	// Load should find the credential in SQLite.
	loaded, err := store.Load("user@example.com")
	if err != nil {
		t.Fatalf("Load() from SQLite error: %v", err)
	}

	var cred map[string]any
	if err := json.Unmarshal(loaded, &cred); err != nil {
		t.Fatal(err)
	}
	if cred["type"] != "authorized_user" {
		t.Errorf("type = %v, want authorized_user", cred["type"])
	}
	if cred["client_id"] != "abc.apps.googleusercontent.com" {
		t.Errorf("client_id = %v", cred["client_id"])
	}
}

func TestLoadFallbackToJSON(t *testing.T) {
	dir := t.TempDir()
	store := &CredentialStore{configDir: dir}

	// No SQLite DB, just a JSON file.
	jsonDir := filepath.Join(dir, "credentials")
	if err := os.MkdirAll(jsonDir, 0700); err != nil {
		t.Fatal(err)
	}
	credJSON := `{"type":"service_account","client_email":"sa@project.iam.gserviceaccount.com"}`
	if err := os.WriteFile(filepath.Join(jsonDir, "sa@project.iam.gserviceaccount.com.json"), []byte(credJSON), 0600); err != nil {
		t.Fatal(err)
	}

	loaded, err := store.Load("sa@project.iam.gserviceaccount.com")
	if err != nil {
		t.Fatalf("Load() from JSON error: %v", err)
	}

	var cred map[string]any
	if err := json.Unmarshal(loaded, &cred); err != nil {
		t.Fatal(err)
	}
	if cred["type"] != "service_account" {
		t.Errorf("type = %v, want service_account", cred["type"])
	}
}

func TestStoreInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	store := &CredentialStore{configDir: dir}

	credFile := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(credFile, []byte("not json"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := store.Store(credFile)
	if err == nil {
		t.Fatal("Store() expected error for invalid JSON")
	}
}

func TestStoreExternalAccount(t *testing.T) {
	dir := t.TempDir()
	store := &CredentialStore{configDir: dir}

	cred := map[string]any{
		"type":                              "external_account",
		"audience":                          "//iam.googleapis.com/projects/123/locations/global/workloadIdentityPools/pool/providers/provider",
		"subject_token_type":                "urn:ietf:params:oauth:token-type:jwt",
		"token_url":                         "https://sts.googleapis.com/v1/token",
		"service_account_impersonation_url": "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/ci-sa@my-project.iam.gserviceaccount.com:generateAccessToken",
		"credential_source":                 map[string]any{"file": ".ci_job_jwt_file"},
	}
	data, err := json.Marshal(cred)
	if err != nil {
		t.Fatal(err)
	}

	credFile := filepath.Join(t.TempDir(), "external.json")
	if err := os.WriteFile(credFile, data, 0600); err != nil {
		t.Fatal(err)
	}

	account, err := store.Store(credFile)
	if err != nil {
		t.Fatalf("Store() error: %v", err)
	}
	if account != "ci-sa@my-project.iam.gserviceaccount.com" {
		t.Errorf("Store() account = %q, want %q", account, "ci-sa@my-project.iam.gserviceaccount.com")
	}
}

func TestStoreMissingClientEmail(t *testing.T) {
	dir := t.TempDir()
	store := &CredentialStore{configDir: dir}

	credFile := filepath.Join(t.TempDir(), "no-email.json")
	if err := os.WriteFile(credFile, []byte(`{"type":"service_account"}`), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := store.Store(credFile)
	if err == nil {
		t.Fatal("Store() expected error for missing client_email")
	}
}

func TestListFromSQLite(t *testing.T) {
	dir := t.TempDir()
	store := &CredentialStore{configDir: dir}

	dbPath := filepath.Join(dir, "credentials.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE credentials (account_id TEXT PRIMARY KEY, value BLOB)`)
	if err != nil {
		t.Fatal(err)
	}
	for _, acct := range []string{"a@test.com", "b@test.com"} {
		_, err = db.Exec(`INSERT INTO credentials (account_id, value) VALUES (?, ?)`, acct, "{}")
		if err != nil {
			t.Fatal(err)
		}
	}
	db.Close()

	accounts, err := store.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(accounts) != 2 {
		t.Errorf("List() returned %d accounts, want 2", len(accounts))
	}
}

func TestLoadNonexistent(t *testing.T) {
	dir := t.TempDir()
	store := &CredentialStore{configDir: dir}

	_, err := store.Load("nonexistent@test.com")
	if err == nil {
		t.Fatal("Load() expected error for nonexistent account")
	}
}

func TestCredAccountID(t *testing.T) {
	tests := []struct {
		name string
		cred map[string]any
		want string
	}{
		{"service_account", map[string]any{"client_email": "sa@proj.iam.gserviceaccount.com"}, "sa@proj.iam.gserviceaccount.com"},
		{"authorized_user with account", map[string]any{"account": "user@gmail.com"}, "user@gmail.com"},
		{"external_account with impersonation", map[string]any{
			"type":                                "external_account",
			"service_account_impersonation_url":   "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/gitlab-ci@my-project.iam.gserviceaccount.com:generateAccessToken",
		}, "gitlab-ci@my-project.iam.gserviceaccount.com"},
		{"external_account without impersonation", map[string]any{
			"type":     "external_account",
			"audience": "//iam.googleapis.com/projects/123/locations/global/workloadIdentityPools/pool/providers/provider",
		}, ""},
		{"empty", map[string]any{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := credAccountID(tt.cred)
			if got != tt.want {
				t.Errorf("credAccountID() = %q, want %q", got, tt.want)
			}
		})
	}
}
