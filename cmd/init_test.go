package cmd

import "testing"

func TestInitCommandRegistered(t *testing.T) {
	for _, c := range newRootCmd(legacyTestInvocation()).Commands() {
		if c.Name() == "init" {
			for _, name := range []string{"no-browser", "console-only", "no-launch-browser", "skip-diagnostics"} {
				if c.Flags().Lookup(name) == nil {
					t.Errorf("init: --%s flag not registered", name)
				}
			}
			return
		}
	}
	t.Fatal("init command not registered on rootCmd")
}

func TestInitQuietRejects(t *testing.T) {
	grp := newInitCmds(legacyTestInvocation())
	prev := grp.inv.flags.quiet
	grp.inv.flags.quiet = true
	t.Cleanup(func() { grp.inv.flags.quiet = prev })
	if err := grp.runInit(grp.initCmd, nil); err == nil {
		t.Fatal("expected --quiet to reject the interactive init flow")
	}
}
