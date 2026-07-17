package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud netapp (#360) ---

var netappCmd = &cobra.Command{Use: "netapp", Short: "Manage Cloud NetApp Files"}

func init() {
	rootCmd.AddCommand(netappCmd)
}

func netappLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func netappChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}
