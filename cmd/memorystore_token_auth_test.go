package cmd

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2"
)

func TestMemorystoreTokenAuthCommands(t *testing.T) {
	if findSub(memstoreInstCmd, "create-token-auth-user") == nil {
		t.Error("memorystore instances create-token-auth-user missing")
	}
	tau := findSub(memstoreInstCmd, "token-auth-users")
	if tau == nil {
		t.Fatal("memorystore instances token-auth-users missing")
	}
	assertSubcommands(t, tau, []string{"create-auth-token", "delete", "describe", "list", "auth-tokens"})
	at := findSub(tau, "auth-tokens")
	if at == nil {
		t.Fatal("token-auth-users auth-tokens missing")
	}
	assertSubcommands(t, at, []string{"delete", "describe", "list"})
}

func TestMemstoreTAUResource(t *testing.T) {
	name, id, err := memstoreTAUResource("p", "us-central1", "u1",
		[]string{"instances", "tokenAuthUsers"}, []string{"instance"}, []string{"inst"})
	if err != nil || name != "projects/p/locations/us-central1/instances/inst/tokenAuthUsers/u1" || id != "u1" {
		t.Errorf("got %q %q %v", name, id, err)
	}
	full := "projects/p/locations/l/instances/i/tokenAuthUsers/u/authTokens/t"
	name, id, err = memstoreTAUResource("", "", full,
		[]string{"instances", "tokenAuthUsers", "authTokens"}, []string{"instance", "token-auth-user"}, []string{"", ""})
	if err != nil || name != full || id != "t" {
		t.Errorf("got %q %q %v", name, id, err)
	}
	if _, _, err := memstoreTAUResource("p", "l", "u1",
		[]string{"instances", "tokenAuthUsers"}, []string{"instance"}, []string{""}); err == nil {
		t.Error("expected error without --instance")
	}
	if _, _, err := memstoreTAUResource("p", "", "i", []string{"instances"}, nil, nil); err == nil {
		t.Error("expected error without --location")
	}
	if _, _, err := memstoreTAUResource("", "", "projects/p/locations/l/clusters/c",
		[]string{"instances"}, nil, nil); err == nil {
		t.Error("expected error for wrong collection")
	}
}

func TestMemstoreCreateTAUAsync(t *testing.T) {
	var gotPath, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.Method + " " + r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"name": "projects/p/locations/l/operations/op1"})
	}))
	defer server.Close()
	orig := memorystoreRest
	memorystoreRest = newRESTClient(server.URL + "/v1beta")
	defer func() { memorystoreRest = orig }()
	defer SetRestTokenSourceForTest(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "t"}))()
	defer SetRestUserProjectForTest("qp")()

	flagMemstoreTAUUser, flagMemstoreTAUAsync, flagMemstoreTAUFormat = "alice", true, "json"
	defer func() { flagMemstoreTAUUser, flagMemstoreTAUAsync, flagMemstoreTAUFormat = "", false, "" }()
	var op map[string]any
	if err := memorystoreRest.do(context.Background(), http.MethodPost,
		"/projects/p/locations/l/instances/i:addTokenAuthUser", nil, map[string]any{"tokenAuthUser": flagMemstoreTAUUser}, &op); err != nil {
		t.Fatal(err)
	}
	if gotPath != "POST /v1beta/projects/p/locations/l/instances/i:addTokenAuthUser" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody != `{"tokenAuthUser":"alice"}` {
		t.Errorf("body = %q", gotBody)
	}
	final, err := memstoreTAUWait(context.Background(), op, "Request issued for: [i]")
	if err != nil || final != nil {
		t.Errorf("async wait: %v %v", final, err)
	}
}
