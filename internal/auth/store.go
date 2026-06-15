package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	_ "modernc.org/sqlite"
)

// CredentialStore reads credentials from gcloud's native SQLite credentials.db
// and can also store service account JSON files for auth login --cred-file.
type CredentialStore struct {
	configDir  string
	dbPathOnce sync.Once
	dbPath     string
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
	if strings.ContainsAny(account, "/\\") {
		return "", fmt.Errorf("account name %q contains path separators", account)
	}

	// Store in SQLite credentials.db (best-effort for gcloud compatibility).
	_ = s.storeToSQLite(account, data)

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

// List returns all stored account identifiers from both the SQLite DB
// and the JSON credential directory.
func (s *CredentialStore) List() ([]string, error) {
	seen := make(map[string]bool)

	// Try SQLite credentials.db.
	if accts, err := s.listFromSQLite(); err == nil {
		for _, a := range accts {
			seen[a] = true
		}
	}

	// Also check JSON credential directory.
	jsonDir := filepath.Join(s.configDir, "credentials")
	entries, _ := os.ReadDir(jsonDir)
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".json") {
			seen[strings.TrimSuffix(name, ".json")] = true
		}
	}

	accounts := make([]string, 0, len(seen))
	for a := range seen {
		accounts = append(accounts, a)
	}
	sort.Strings(accounts)
	return accounts, nil
}

func (s *CredentialStore) listFromSQLite() ([]string, error) {
	dbPath := s.sqliteDBPath()
	if dbPath == "" {
		return nil, fmt.Errorf("no writable credentials.db path")
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT account_id FROM credentials`)
	if err != nil {
		return nil, err
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
	dbPath := s.sqliteDBPath()
	if dbPath == "" {
		return nil, fmt.Errorf("no credentials.db available")
	}
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
	path := filepath.Join(s.configDir, "credentials", filepath.Base(account)+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading credential for %s: %w", account, err)
	}
	return data, nil
}

// sqliteDBPath returns a writable path for credentials.db.
// It tries the config directory first, falling back to a temporary directory.
func (s *CredentialStore) sqliteDBPath() string {
	s.dbPathOnce.Do(func() {
		primary := filepath.Join(s.configDir, "credentials.db")
		if db, err := sql.Open("sqlite", primary); err == nil {
			if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS credentials (account_id TEXT PRIMARY KEY, value BLOB)`); err == nil {
				db.Close()
				os.Chmod(primary, 0600)
				s.dbPath = primary
				return
			}
			db.Close()
		}
		fallback, err := os.MkdirTemp("", "gcloud-go-*")
		if err == nil {
			s.dbPath = filepath.Join(fallback, "credentials.db")
		}
	})
	return s.dbPath
}

func (s *CredentialStore) storeToSQLite(account string, data []byte) error {
	dbPath := s.sqliteDBPath()
	if dbPath == "" {
		return fmt.Errorf("no writable credentials.db path")
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS credentials (account_id TEXT PRIMARY KEY, value BLOB)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`REPLACE INTO credentials (account_id, value) VALUES (?, ?)`, account, string(data))
	if err != nil {
		return err
	}
	os.Chmod(dbPath, 0600)
	return nil
}

// StoreRaw saves raw credential JSON for the given account identifier.
func (s *CredentialStore) StoreRaw(account string, data []byte) error {
	if strings.ContainsAny(account, "/\\") {
		return fmt.Errorf("account name %q contains path separators", account)
	}

	_ = s.storeToSQLite(account, data)

	jsonDir := filepath.Join(s.configDir, "credentials")
	if err := os.MkdirAll(jsonDir, 0700); err != nil {
		return fmt.Errorf("creating credentials directory: %w", err)
	}
	dest := filepath.Join(jsonDir, account+".json")
	if err := os.WriteFile(dest, data, 0600); err != nil {
		return fmt.Errorf("writing credential file: %w", err)
	}
	return nil
}

// Revoke removes stored credentials for the given account from both
// the SQLite DB and the JSON credential directory.
func (s *CredentialStore) Revoke(account string) error {
	if strings.ContainsAny(account, "/\\") {
		return fmt.Errorf("account name %q contains path separators", account)
	}

	var errs []error

	// Remove from SQLite.
	if err := s.deleteFromSQLite(account); err != nil {
		errs = append(errs, fmt.Errorf("sqlite: %w", err))
	}

	// Remove JSON file.
	jsonPath := filepath.Join(s.configDir, "credentials", account+".json")
	if err := os.Remove(jsonPath); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("json file: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("revoking %s: %w", account, errors.Join(errs...))
	}
	return nil
}

func (s *CredentialStore) deleteFromSQLite(account string) error {
	dbPath := s.sqliteDBPath()
	if dbPath == "" {
		return nil // no DB to delete from
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`DELETE FROM credentials WHERE account_id = ?`, account)
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
