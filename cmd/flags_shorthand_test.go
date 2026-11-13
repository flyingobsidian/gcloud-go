package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// Guards #2084: cobra adds -h for --help when a command runs, and panics if
// the command already uses -h. Adding the help flag to every command finds
// such clashes without running anything.
func TestHelpFlagFitsEveryCommand(t *testing.T) {
	inv, _, _ := newTestInvocation(t, nil, "")
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s: adding --help panics: %v", c.CommandPath(), r)
				}
			}()
			c.InitDefaultHelpFlag()
		}()
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(newRootCmd(inv))
}
