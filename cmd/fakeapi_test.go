package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/api/option"
)

// newFakeAPIInvocation returns an invocation whose Google API clients talk
// to api and authenticate as user@example.com with tokens from a fake token
// endpoint. It is for testing migrated command groups end to end.
func newFakeAPIInvocation(t *testing.T, api http.Handler) (*invocation, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"fake-token","token_type":"Bearer","expires_in":3600}`))
	}))
	t.Cleanup(tokenSrv.Close)
	apiSrv := httptest.NewServer(api)
	t.Cleanup(apiSrv.Close)

	dir := t.TempDir()
	cred, err := json.Marshal(map[string]string{
		"type":          "authorized_user",
		"client_id":     "id",
		"client_secret": "secret",
		"refresh_token": "refresh",
		"token_uri":     tokenSrv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	for path, data := range map[string][]byte{
		"credentials/user@example.com.json": cred,
		"configurations/config_default":     []byte("[core]\naccount = user@example.com\n"),
	} {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, data, 0600); err != nil {
			t.Fatal(err)
		}
	}

	inv, stdout, stderr := newTestInvocation(t, map[string]string{"CLOUDSDK_CONFIG": dir}, "")
	inv.clientOptions = []option.ClientOption{option.WithEndpoint(apiSrv.URL + "/")}
	return inv, stdout, stderr
}

// runInvocation runs args on a root command for inv without the legacy
// command groups, as RunInProcess does.
func runInvocation(inv *invocation, args ...string) error {
	root := newRootCmd(inv, false)
	root.SetArgs(args)
	inv.ctx = context.Background()
	return root.ExecuteContext(inv.ctx)
}
