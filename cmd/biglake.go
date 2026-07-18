package cmd

import "github.com/spf13/cobra"

// --- gcloud biglake (#307) ---
//
// The BigLake Iceberg REST catalog surface uses a bespoke URL shape not
// modelled by the biglake/v1 Go client, so we talk to it directly through the
// shared REST helper. `biglakeIcebergRest` is rooted at the extensions prefix
// so subgroup files can just append `/projects/<PROJ>/locations/<LOC>/...`.
// `biglakeStandardRest` is the standard biglake/v1 endpoint used for IAM
// operations that live under the normal REST tree.

var biglakeCmd = &cobra.Command{
	Use:   "biglake",
	Short: "Manage BigLake resources",
}

var biglakeIcebergCmd = &cobra.Command{
	Use:   "iceberg",
	Short: "BigLake Iceberg REST catalogs",
}

var (
	biglakeIcebergRest  = newRESTClient("https://biglake.googleapis.com/iceberg/v1/restcatalog/extensions")
	biglakeStandardRest = newRESTClient("https://biglake.googleapis.com/v1")
)

func init() {
	biglakeCmd.AddCommand(biglakeIcebergCmd)
	rootCmd.AddCommand(biglakeCmd)
}
