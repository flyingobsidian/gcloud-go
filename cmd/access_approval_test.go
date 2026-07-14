package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func aaSubgroup(name string) *cobra.Command {
	for _, c := range accessApprovalCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestAccessApprovalRequestsSubcommands(t *testing.T) {
	g := aaSubgroup("requests")
	if g == nil {
		t.Fatal("access-approval requests missing")
	}
	assertSubcommands(t, g, []string{"approve", "dismiss", "get", "invalidate", "list"})
}

func TestAccessApprovalServiceAccountSubcommands(t *testing.T) {
	g := aaSubgroup("service-account")
	if g == nil {
		t.Fatal("access-approval service-account missing")
	}
	assertSubcommands(t, g, []string{"get"})
}

func TestAccessApprovalSettingsSubcommands(t *testing.T) {
	g := aaSubgroup("settings")
	if g == nil {
		t.Fatal("access-approval settings missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "update"})
}
