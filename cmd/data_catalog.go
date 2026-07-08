package cmd

import "github.com/spf13/cobra"

// --- gcloud data-catalog (#321) ---

var dataCatalogCmd = &cobra.Command{Use: "data-catalog", Short: "Manage Data Catalog (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(dataCatalogCmd, "entries", "Manage entries", append(crud, "lookup")...)
	registerStubGroup(dataCatalogCmd, "entry-groups", "(DEPRECATED) Manage entry groups", crud...)
	registerStubGroup(dataCatalogCmd, "tag-templates", "(DEPRECATED) Manage tag templates", append(crud, "fields")...)
	registerStubGroup(dataCatalogCmd, "tags", "(DEPRECATED) Manage tags", crud...)
	registerStubGroup(dataCatalogCmd, "taxonomies", "Manage taxonomies", append(crud, "policy-tags", "export", "import")...)
	registerStubCommand(dataCatalogCmd, "search", "(DEPRECATED) Search Data Catalog")
	rootCmd.AddCommand(dataCatalogCmd)
}
