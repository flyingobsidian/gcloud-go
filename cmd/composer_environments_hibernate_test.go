package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	composer "google.golang.org/api/composer/v1"
	composerbeta "google.golang.org/api/composer/v1beta1"
	"google.golang.org/api/option"
)

func TestComposerHibernateResumeRegistered(t *testing.T) {
	for _, name := range []string{"hibernate", "resume"} {
		c, _, err := composerEnvCmd.Find([]string{name})
		if err != nil || c.Name() != name {
			t.Fatalf("environments %s not registered: %v", name, err)
		}
		for _, f := range []string{"location", "async"} {
			if c.Flags().Lookup(f) == nil {
				t.Errorf("%s missing --%s", name, f)
			}
		}
	}
}

func TestComposerBetaWaitOp(t *testing.T) {
	defer func(d time.Duration) { composerOpPollInterval = d }(composerOpPollInterval)
	composerOpPollInterval = time.Millisecond

	polls, failNext := 0, false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		polls++
		if !strings.HasSuffix(r.URL.Path, "/operations/op1") {
			http.NotFound(w, r)
			return
		}
		if polls < 2 {
			fmt.Fprint(w, `{"name":"projects/p/locations/l/operations/op1"}`)
			return
		}
		if failNext {
			fmt.Fprint(w, `{"name":"projects/p/locations/l/operations/op1","done":true,"error":{"message":"boom"}}`)
			return
		}
		fmt.Fprint(w, `{"name":"projects/p/locations/l/operations/op1","done":true}`)
	}))
	defer srv.Close()
	svc, err := composerbeta.NewService(context.Background(),
		option.WithEndpoint(srv.URL), option.WithoutAuthentication())
	if err != nil {
		t.Fatal(err)
	}
	op := &composerbeta.Operation{Name: "projects/p/locations/l/operations/op1"}
	if err := composerBetaWaitOp(context.Background(), svc, op); err != nil {
		t.Fatalf("wait: %v", err)
	}
	if polls != 2 {
		t.Errorf("polls = %d, want 2", polls)
	}

	polls, failNext = 2, true
	err = composerBetaWaitOp(context.Background(), svc, op)
	want := "Operation [projects/p/locations/l/operations/op1] failed: boom"
	if err == nil || err.Error() != want {
		t.Errorf("got %v, want %q", err, want)
	}
}

func TestApplyComposerDevelopmentMode(t *testing.T) {
	if composerEnvCreateCmd.Flags().Lookup("enable-development-mode") == nil {
		t.Fatal("create missing --enable-development-mode")
	}
	env := &composer.Environment{Mode: "MODE_UNSPECIFIED"}
	applyComposerDevelopmentMode(env, false)
	if env.Mode != "MODE_UNSPECIFIED" {
		t.Errorf("unset flag changed mode to %q", env.Mode)
	}
	applyComposerDevelopmentMode(env, true)
	if env.Mode != "DEVELOPMENT" {
		t.Errorf("mode = %q, want DEVELOPMENT", env.Mode)
	}
}
