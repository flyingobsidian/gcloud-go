package cli

import "testing"

func TestSurveyCommandRegistered(t *testing.T) {
	for _, c := range newRootCmd(legacyTestInvocation()).Commands() {
		if c.Name() == "survey" {
			return
		}
	}
	t.Fatal("survey command not registered on rootCmd")
}
