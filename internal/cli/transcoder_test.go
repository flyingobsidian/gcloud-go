package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func transcoderSubgroup(name string) *cobra.Command {
	grp := newTranscoderCmds(legacyTestInvocation())
	for _, c := range grp.transcoderCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestTranscoderJobsSubcommands(t *testing.T) {
	g := transcoderSubgroup("jobs")
	if g == nil {
		t.Fatal("transcoder jobs missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestTranscoderTemplatesSubcommands(t *testing.T) {
	g := transcoderSubgroup("templates")
	if g == nil {
		t.Fatal("transcoder templates missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}
