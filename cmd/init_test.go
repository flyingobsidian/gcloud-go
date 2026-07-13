package cmd

import "testing"

func TestInitCommandRegistered(t *testing.T) {
	for _, c := range rootCmd.Commands() {
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
	prev := flagQuiet
	flagQuiet = true
	t.Cleanup(func() { flagQuiet = prev })
	if err := runInit(initCmd, nil); err == nil {
		t.Fatal("expected --quiet to reject the interactive init flow")
	}
}
