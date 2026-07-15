package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func domainsSubgroup(name string) *cobra.Command {
	for _, c := range domainsCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestDomainsHasRegistrations(t *testing.T) {
	if domainsSubgroup("registrations") == nil {
		t.Fatal("domains registrations missing")
	}
}

func TestDomainsRegistrationsSubcommands(t *testing.T) {
	g := domainsSubgroup("registrations")
	if g == nil {
		t.Fatal("registrations missing")
	}
	assertSubcommands(t, g, []string{
		"configure", "delete", "describe", "export", "import",
		"initiate-push-transfer", "list", "register", "renew-domain",
		"reset-authorization-code", "retrieve-authorization-code",
		"retrieve-google-domains-dns-records",
		"retrieve-google-domains-forwarding-config",
		"retrieve-import-transfer-parameters",
		"retrieve-register-parameters", "retrieve-transfer-parameters",
		"search-domains", "transfer", "update",
	})
}

func TestDomainsConfigureSubcommands(t *testing.T) {
	g := domainsSubgroup("registrations")
	if g == nil {
		t.Fatal("registrations missing")
	}
	cfg := findSub(g, "configure")
	if cfg == nil {
		t.Fatal("configure missing")
	}
	assertSubcommands(t, cfg, []string{"contacts", "dns", "management"})
}

func TestParseDomainsMoney(t *testing.T) {
	m, err := parseDomainsMoney("12.00.USD")
	if err != nil {
		t.Fatalf("parseDomainsMoney error: %v", err)
	}
	if m.CurrencyCode != "USD" || m.Units != 12 {
		t.Errorf("got %+v, want units=12 currency=USD", m)
	}
}
