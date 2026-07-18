package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud design-center (#329) ---
//
// The Design Center API (`designcenter.googleapis.com`, v1alpha) is not
// exposed by google.golang.org/api. Uses the shared restClient from
// rest_helpers.go with a per-service endpoint.

var designCenterCmd = &cobra.Command{Use: "design-center", Short: "Manage Google Cloud Design Center"}

var designCenterRest = newRESTClient("https://designcenter.googleapis.com/v1alpha")

func init() {
	rootCmd.AddCommand(designCenterCmd)
}

func dcLocationName(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func dcJoin(parts ...string) string {
	out := parts[0]
	for _, p := range parts[1:] {
		out = out + "/" + strings.TrimPrefix(p, "/")
	}
	return out
}
