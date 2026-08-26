package cmd

import (
	"encoding/json"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/config"
)

// resolveBillingProject returns the effective billing / quota project
// identifier that should be sent in the `x-goog-user-project` header for
// REST calls. Resolution order (matches gcloud-python):
//
//  1. --billing-project CLI flag (flagBillingProject)
//  2. billing/quota_project from the active gcloud configuration
//  3. quota_project_id inside the Application Default Credentials file
//  4. "" if none of the above are set (callers decide whether that's fatal)
//
// Errors along the way are swallowed -- an unreadable config or ADC file
// simply falls through to the next source.
func resolveBillingProject() string {
	if flagBillingProject != "" {
		return flagBillingProject
	}
	if props, err := config.Load(); err == nil && props != nil {
		if props.Billing.QuotaProject != "" {
			return props.Billing.QuotaProject
		}
	}
	if q := adcQuotaProject(); q != "" {
		return q
	}
	return ""
}

// adcQuotaProject returns the quota_project_id field of the Application
// Default Credentials file at adcFilePath(), or "" if missing/unreadable.
func adcQuotaProject() string {
	data, err := os.ReadFile(adcFilePath())
	if err != nil {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	if v, ok := m["quota_project_id"].(string); ok {
		return v
	}
	return ""
}
