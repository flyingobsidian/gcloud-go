package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	_ "modernc.org/sqlite"
)

// CredentialStore reads credentials from gcloud's native SQLite credentials.db
// and can also store service account JSON files for auth login --cred-file.
type CredentialStore struct {
	configDir string
}

// NewStore creates a credential store using the gcloud config directory.
func NewStore() (*CredentialStore, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return nil, err
	}
	return &CredentialStore{configDir: dir}, nil
}

// Load retrieves a stored credential by account identifier.
// It first checks gcloud's credentials.db (SQLite), then falls back to
// the JSON file store used by auth login --cred-file.
func (s *CredentialStore) Load(account string) ([]byte, error) {
	// Try SQLite credentials.db first.
	data, err := s.loadFromSQLite(account)
	if err == nil {
		return data, nil
	}

	// Fall back to JSON file store.
	return s.loadFromJSON(account)
}

// Store saves a service account credential file and returns the account identifier.
// It writes to both the SQLite DB (for gcloud compatibility) and a JSON file (fallback).
func (s *CredentialStore) Store(credFile string) (string, error) {
	data, err := os.ReadFile(credFile)
	if err != nil {
		return "", fmt.Errorf("reading credential file: %w", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("parsing credential file: %w", err)
	}

	// Determine the account identifier.
	account := credAccountID(parsed)
	if account == "" {
		return "", fmt.Errorf("cannot determine account from credential file (no client_email or service_account_impersonation_url)")
	}

	// Store in SQLite credentials.db.
	if err := s.storeToSQLite(account, data); err != nil {
		// Non-fatal: fall through to JSON store.
		fmt.Fprintf(os.Stderr, "warning: could not write to credentials.db: %v\n", err)
	}

	// Also store as JSON file for fallback.
	jsonDir := filepath.Join(s.configDir, "credentials")
	if err := os.MkdirAll(jsonDir, 0700); err != nil {
		return "", fmt.Errorf("creating credentials directory: %w", err)
	}
	dest := filepath.Join(jsonDir, account+".json")
	if err := os.WriteFile(dest, data, 0600); err != nil {
		return "", fmt.Errorf("writing credential file: %w", err)
	}

	return account, nil
}

// List returns all stored account identifiers from the SQLite DB.
func (s *CredentialStore) List() ([]string, error) {
	dbPath := filepath.Join(s.configDir, "credentials.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening credentials.db: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT account_id FROM credentials`)
	if err != nil {
		return nil, fmt.Errorf("querying credentials: %w", err)
	}
	defer rows.Close()

	var accounts []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		accounts = append(accounts, id)
	}
	return accounts, rows.Err()
}

func (s *CredentialStore) loadFromSQLite(account string) ([]byte, error) {
	dbPath := filepath.Join(s.configDir, "credentials.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var value string
	err = db.QueryRow(`SELECT value FROM credentials WHERE account_id = ?`, account).Scan(&value)
	if err != nil {
		return nil, fmt.Errorf("credential not found in db for %s: %w", account, err)
	}
	return []byte(value), nil
}

func (s *CredentialStore) loadFromJSON(account string) ([]byte, error) {
	path := filepath.Join(s.configDir, "credentials", account+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading credential for %s: %w", account, err)
	}
	return data, nil
}

func (s *CredentialStore) storeToSQLite(account string, data []byte) error {
	dbPath := filepath.Join(s.configDir, "credentials.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Ensure table exists.
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS credentials (account_id TEXT PRIMARY KEY, value BLOB)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`REPLACE INTO credentials (account_id, value) VALUES (?, ?)`, account, string(data))
	return err
}

// credAccountID returns the account identifier from a credential JSON.
func credAccountID(cred map[string]any) string {
	// Service account: use client_email.
	if email, ok := cred["client_email"].(string); ok && email != "" {
		return email
	}
	// Authorized user: use account field if present.
	if acct, ok := cred["account"].(string); ok && acct != "" {
		return acct
	}
	// External account (workload identity federation): extract SA email
	// from service_account_impersonation_url.
	if url, ok := cred["service_account_impersonation_url"].(string); ok && url != "" {
		return extractSAFromImpersonationURL(url)
	}
	return ""
}

// extractSAFromImpersonationURL extracts the service account email from a URL like:
// https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/SA@PROJECT.iam.gserviceaccount.com:generateAccessToken
func extractSAFromImpersonationURL(url string) string {
	const marker = "/serviceAccounts/"
	idx := strings.Index(url, marker)
	if idx < 0 {
		return ""
	}
	sa := url[idx+len(marker):]
	if colon := strings.IndexByte(sa, ':'); colon >= 0 {
		sa = sa[:colon]
	}
	return sa
}
