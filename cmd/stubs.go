package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// registerStubGroup attaches a subgroup with the given short description and
// N stub subcommands (each returning a "not yet implemented" error) to parent.
// Used by command surfaces that are registered ahead of a full API-backed
// implementation.
func registerStubGroup(parent *cobra.Command, name, short string, subs ...string) *cobra.Command {
	g := &cobra.Command{Use: name, Short: short}
	for _, s := range subs {
		s := s
		g.AddCommand(&cobra.Command{
			Use:   s,
			Short: "Not yet implemented",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("%s %s %s: not yet implemented", parent.Name(), name, s)
			},
		})
	}
	parent.AddCommand(g)
	return g
}

// registerStubCommand attaches a single stub command to parent.
func registerStubCommand(parent *cobra.Command, name, short string) *cobra.Command {
	c := &cobra.Command{
		Use:   name,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("%s %s: not yet implemented", parent.Name(), name)
		},
	}
	parent.AddCommand(c)
	return c
}
