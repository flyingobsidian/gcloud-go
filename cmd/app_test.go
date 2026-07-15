package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func appSubgroup(name string) *cobra.Command {
	for _, c := range appCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestAppDomainMappingsSubcommands(t *testing.T) {
	g := appSubgroup("domain-mappings")
	if g == nil {
		t.Fatal("app domain-mappings missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestAppFirewallRulesSubcommands(t *testing.T) {
	g := appSubgroup("firewall-rules")
	if g == nil {
		t.Fatal("app firewall-rules missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "test-ip", "update"})
}

func TestAppInstancesSubcommands(t *testing.T) {
	g := appSubgroup("instances")
	if g == nil {
		t.Fatal("app instances missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "disable-debug", "enable-debug", "list"})
}

func TestAppLogsSubcommands(t *testing.T) {
	g := appSubgroup("logs")
	if g == nil {
		t.Fatal("app logs missing")
	}
	assertSubcommands(t, g, []string{"read", "tail"})
}

func TestAppOperationsSubcommands(t *testing.T) {
	g := appSubgroup("operations")
	if g == nil {
		t.Fatal("app operations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "wait"})
}

func TestAppRegionsSubcommands(t *testing.T) {
	g := appSubgroup("regions")
	if g == nil {
		t.Fatal("app regions missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestAppRuntimesSubcommands(t *testing.T) {
	g := appSubgroup("runtimes")
	if g == nil {
		t.Fatal("app runtimes missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestAppServicesSubcommands(t *testing.T) {
	g := appSubgroup("services")
	if g == nil {
		t.Fatal("app services missing")
	}
	assertSubcommands(t, g, []string{"browse", "delete", "describe", "list", "set-traffic", "update"})
}

func TestAppSSLCertificatesSubcommands(t *testing.T) {
	g := appSubgroup("ssl-certificates")
	if g == nil {
		t.Fatal("app ssl-certificates missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestAppVersionsSubcommands(t *testing.T) {
	g := appSubgroup("versions")
	if g == nil {
		t.Fatal("app versions missing")
	}
	assertSubcommands(t, g, []string{"browse", "delete", "describe", "list", "migrate", "start", "stop"})
}

func TestAppServiceURL(t *testing.T) {
	cases := []struct {
		host    string
		service string
		version string
		want    string
	}{
		{"my-app.appspot.com", "", "", "https://my-app.appspot.com"},
		{"my-app.appspot.com", "default", "", "https://my-app.appspot.com"},
		{"my-app.appspot.com", "svc1", "", "https://svc1-dot-my-app.appspot.com"},
		{"my-app.appspot.com", "svc1", "v1", "https://v1-dot-svc1-dot-my-app.appspot.com"},
		{"my-app.appspot.com", "default", "v1", "https://v1-dot-my-app.appspot.com"},
	}
	for _, c := range cases {
		got := serviceURL(c.host, c.service, c.version)
		if got != c.want {
			t.Errorf("serviceURL(%q,%q,%q) = %q, want %q", c.host, c.service, c.version, got, c.want)
		}
	}
}

func TestAppParsePriority(t *testing.T) {
	if p, err := parsePriority("42"); err != nil || p != 42 {
		t.Errorf("parsePriority(42) = %d, %v", p, err)
	}
	if p, err := parsePriority("default"); err != nil || p != 2147483647 {
		t.Errorf("parsePriority(default) = %d, %v", p, err)
	}
	if _, err := parsePriority("notanumber"); err == nil {
		t.Error("expected error for invalid priority")
	}
}
