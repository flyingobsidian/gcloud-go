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
