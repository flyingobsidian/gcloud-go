package cli

import "testing"

// gcloud-python 589.0.0 removed the telco-automation group (#1974).
func TestTelcoAutomationRemoved(t *testing.T) {
	for _, c := range newRootCmd(legacyTestInvocation()).Commands() {
		if c.Name() == "telco-automation" {
			t.Fatal("telco-automation is still registered")
		}
	}
}
