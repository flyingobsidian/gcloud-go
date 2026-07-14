package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func pubsubSubgroup(name string) *cobra.Command {
	for _, c := range pubsubCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestPubsubTopicsSubcommands(t *testing.T) {
	g := pubsubSubgroup("topics")
	if g == nil {
		t.Fatal("pubsub topics missing")
	}
	assertSubcommands(t, g, []string{
		"add-iam-policy-binding", "create", "delete", "describe",
		"detach-subscription", "get-iam-policy", "list", "list-subscriptions",
		"publish", "remove-iam-policy-binding", "set-iam-policy", "update",
	})
}

func TestPubsubSchemasSubcommands(t *testing.T) {
	g := pubsubSubgroup("schemas")
	if g == nil {
		t.Fatal("pubsub schemas missing")
	}
	assertSubcommands(t, g, []string{
		"commit", "create", "delete", "delete-revision", "describe",
		"list", "list-revisions", "rollback", "validate-message", "validate-schema",
	})
}

func TestPubsubSnapshotsSubcommands(t *testing.T) {
	g := pubsubSubgroup("snapshots")
	if g == nil {
		t.Fatal("pubsub snapshots missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestPubsubSubscriptionsSubcommands(t *testing.T) {
	g := pubsubSubgroup("subscriptions")
	if g == nil {
		t.Fatal("pubsub subscriptions missing")
	}
	assertSubcommands(t, g, []string{
		"ack", "add-iam-policy-binding", "create", "delete", "describe",
		"get-iam-policy", "list", "modify-message-ack-deadline", "modify-push-config",
		"pull", "remove-iam-policy-binding", "seek", "set-iam-policy", "update",
	})
}
