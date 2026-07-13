package cmd

import "testing"

func TestSurveyCommandRegistered(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "survey" {
			return
		}
	}
	t.Fatal("survey command not registered on rootCmd")
}
