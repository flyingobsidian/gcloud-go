package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// errNoBillingProject reports that api needs a billing / quota project but
// none is configured, naming every way to set one. Shared so the pre-flight
// checks and restClient.do speak with one voice.
func errNoBillingProject(api string) error {
	return fmt.Errorf(
		"a quota (billing) project is required for %s but none was found. "+
			"Set one via --billing-project=PROJECT_ID, "+
			"`gcloud config set billing/quota_project PROJECT_ID`, or "+
			"`gcloud auth application-default set-quota-project PROJECT_ID`", api)
}

// endpointHost returns the host of a REST endpoint for use in messages, or
// the endpoint unchanged when it cannot be parsed.
func endpointHost(endpoint string) string {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" {
		return strings.TrimRight(endpoint, "/")
	}
	return u.Host
}

// quotaProjectFromADC returns the quota_project_id field of the Application
// Default Credentials file at path, or "" if missing/unreadable.
func quotaProjectFromADC(path string) string {
	data, err := os.ReadFile(path)
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
