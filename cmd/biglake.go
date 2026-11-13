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

func biglakeIcebergRest() *restClient {
	return newRESTClient("https://biglake.googleapis.com/iceberg/v1/restcatalog/extensions")
}

func biglakeStandardRest() *restClient {
	return newRESTClient("https://biglake.googleapis.com/v1")
}

func (s *biglakeCmds) initBiglake() {
	s.biglakeCmd.AddCommand(s.biglakeIcebergCmd)

}

// biglakeCmds holds the commands and flag values of one invocation's "biglake"
// command group, so invocations don't share state (#2070).
type biglakeCmds struct {
	inv *invocation

	biglakeCmd                *cobra.Command
	biglakeIcebergCmd         *cobra.Command
	biglakeIcebergCatalogsCmd *cobra.Command
	flagBLICLocation          string
	flagBLICFormat            string
	flagBLICConfigFile        string
	flagBLICUpdateMask        string
	blicCreateCmd             *cobra.Command
	blicDeleteCmd             *cobra.Command
	blicDescribeCmd           *cobra.Command
	blicListCmd               *cobra.Command
	blicUpdateCmd             *cobra.Command
	blicFailoverCmd           *cobra.Command
	blicGetIamCmd             *cobra.Command
	blicSetIamCmd             *cobra.Command
}

// newBiglakeCmd returns the "biglake" command group for inv.
func newBiglakeCmd(inv *invocation) *cobra.Command {
	return newBiglakeCmds(inv).biglakeCmd
}

// newBiglakeCmds builds the "biglake" command group for inv and returns its state.
func newBiglakeCmds(inv *invocation) *biglakeCmds {
	s := &biglakeCmds{inv: inv}
	s.biglakeCmd = &cobra.Command{
		Use:   "biglake",
		Short: "Manage BigLake resources",
	}
	s.biglakeIcebergCmd = &cobra.Command{
		Use:   "iceberg",
		Short: "BigLake Iceberg REST catalogs",
	}
	s.biglakeIcebergCatalogsCmd = &cobra.Command{Use: "catalogs", Short: "Manage BigLake Iceberg REST catalogs"}
	s.blicCreateCmd = &cobra.Command{
		Use: "create CATALOG", Short: "Create a BigLake Iceberg REST catalog",
		Args: cobra.ExactArgs(1), RunE: s.runBLICCreate,
	}
	s.blicDeleteCmd = &cobra.Command{
		Use: "delete CATALOG", Short: "Delete a BigLake Iceberg REST catalog",
		Args: cobra.ExactArgs(1), RunE: s.runBLICDelete,
	}
	s.blicDescribeCmd = &cobra.Command{
		Use: "describe CATALOG", Short: "Describe a BigLake Iceberg REST catalog",
		Args: cobra.ExactArgs(1), RunE: s.runBLICDescribe,
	}
	s.blicListCmd = &cobra.Command{
		Use: "list", Short: "List BigLake Iceberg REST catalogs",
		Args: cobra.NoArgs, RunE: s.runBLICList,
	}
	s.blicUpdateCmd = &cobra.Command{
		Use: "update CATALOG", Short: "Update a BigLake Iceberg REST catalog",
		Args: cobra.ExactArgs(1), RunE: s.runBLICUpdate,
	}
	s.blicFailoverCmd = &cobra.Command{
		Use: "failover CATALOG", Short: "Fail over a BigLake Iceberg REST catalog",
		Args: cobra.ExactArgs(1), RunE: s.runBLICFailover,
	}
	s.blicGetIamCmd = &cobra.Command{
		Use: "get-iam-policy CATALOG", Short: "Get the IAM policy for a BigLake Iceberg catalog",
		Args: cobra.ExactArgs(1), RunE: s.runBLICGetIam,
	}
	s.blicSetIamCmd = &cobra.Command{
		Use: "set-iam-policy CATALOG POLICY_FILE", Short: "Set the IAM policy for a BigLake Iceberg catalog",
		Args: cobra.ExactArgs(2), RunE: s.runBLICSetIam,
	}
	s.initBiglake()
	s.initBiglakeIcebergCatalogs()
	return s
}
