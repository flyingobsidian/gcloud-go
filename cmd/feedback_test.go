package cmd

import "testing"

func TestFeedbackCommandRegistered(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "feedback" {
			if c.Flags().Lookup("log-file") == nil {
				t.Error("--log-file flag not registered on feedback")
			}
			return
		}
	}
	t.Fatal("feedback command not registered on rootCmd")
}
