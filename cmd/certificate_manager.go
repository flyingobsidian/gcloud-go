package cmd

import "github.com/spf13/cobra"

// --- gcloud certificate-manager (#313) ---

func (s *certificateManagerCmds) initCertificateManager() {
	// All subgroups (certificates, dns-authorizations, issuance-configs, maps
	// (with entries), operations, trust-configs) are implemented in
	// certificate_manager_all.go.

}

// certificateManagerCmds holds the commands and flag values of one invocation's "certificateManager"
// command group, so invocations don't share state (#2070).
type certificateManagerCmds struct {
	inv *invocation

	certificateManagerCmd *cobra.Command
	flagCMLocation        string
	flagCMConfigFile      string
	flagCMUpdateMask      string
	flagCMFormat          string
	flagCMAsync           bool
	// Resource-manager tag bindings to attach on create (gcloud-python 580.0.0
	// `--tags` flag). Bindings win over any `tags` field in --config-file so
	// the flag has predictable precedence.
	flagCMTags           map[string]string
	cmCertificatesCmd    *cobra.Command
	cmCertCreateCmd      *cobra.Command
	cmCertDeleteCmd      *cobra.Command
	cmCertDescribeCmd    *cobra.Command
	cmCertListCmd        *cobra.Command
	cmCertUpdateCmd      *cobra.Command
	cmDnsAuthCmd         *cobra.Command
	cmDACreateCmd        *cobra.Command
	cmDADeleteCmd        *cobra.Command
	cmDADescribeCmd      *cobra.Command
	cmDAListCmd          *cobra.Command
	cmDAUpdateCmd        *cobra.Command
	cmIssuanceConfigsCmd *cobra.Command
	cmICCreateCmd        *cobra.Command
	cmICDeleteCmd        *cobra.Command
	cmICDescribeCmd      *cobra.Command
	cmICListCmd          *cobra.Command
	cmICUpdateCmd        *cobra.Command
	cmMapsCmd            *cobra.Command
	cmMapCreateCmd       *cobra.Command
	cmMapDeleteCmd       *cobra.Command
	cmMapDescribeCmd     *cobra.Command
	cmMapListCmd         *cobra.Command
	cmMapUpdateCmd       *cobra.Command
	cmMapEntriesCmd      *cobra.Command
	cmMECreateCmd        *cobra.Command
	cmMEDeleteCmd        *cobra.Command
	cmMEDescribeCmd      *cobra.Command
	cmMEListCmd          *cobra.Command
	cmMEUpdateCmd        *cobra.Command
	flagCMMEMap          string
	cmTrustConfigsCmd    *cobra.Command
	cmTCCreateCmd        *cobra.Command
	cmTCDeleteCmd        *cobra.Command
	cmTCDescribeCmd      *cobra.Command
	cmTCListCmd          *cobra.Command
	cmTCUpdateCmd        *cobra.Command
	cmOperationsCmd      *cobra.Command
	cmOpDescribeCmd      *cobra.Command
	cmOpListCmd          *cobra.Command
}

// newCertificateManagerCmd returns the "certificateManager" command group for inv.
func newCertificateManagerCmd(inv *invocation) *cobra.Command {
	return newCertificateManagerCmds(inv).certificateManagerCmd
}

// newCertificateManagerCmds builds the "certificateManager" command group for inv and returns its state.
func newCertificateManagerCmds(inv *invocation) *certificateManagerCmds {
	s := &certificateManagerCmds{inv: inv}
	s.certificateManagerCmd = &cobra.Command{Use: "certificate-manager", Short: "Manage Certificate Manager"}
	s.cmCertificatesCmd = &cobra.Command{Use: "certificates", Short: "Manage Certificate Manager certificates"}
	s.cmCertCreateCmd = &cobra.Command{
		Use: "create CERT", Short: "Create a certificate from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runCMCertCreate,
	}
	s.cmCertDeleteCmd = &cobra.Command{
		Use: "delete CERT", Short: "Delete a certificate",
		Args: cobra.ExactArgs(1), RunE: s.runCMCertDelete,
	}
	s.cmCertDescribeCmd = &cobra.Command{
		Use: "describe CERT", Short: "Describe a certificate",
		Args: cobra.ExactArgs(1), RunE: s.runCMCertDescribe,
	}
	s.cmCertListCmd = &cobra.Command{
		Use: "list", Short: "List certificates",
		Args: cobra.NoArgs, RunE: s.runCMCertList,
	}
	s.cmCertUpdateCmd = &cobra.Command{
		Use: "update CERT", Short: "Update a certificate from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runCMCertUpdate,
	}
	s.cmDnsAuthCmd = &cobra.Command{Use: "dns-authorizations", Short: "Manage DNS authorizations"}
	s.cmDACreateCmd = &cobra.Command{
		Use: "create DA", Short: "Create a DNS authorization from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runCMDACreate,
	}
	s.cmDADeleteCmd = &cobra.Command{
		Use: "delete DA", Short: "Delete a DNS authorization",
		Args: cobra.ExactArgs(1), RunE: s.runCMDADelete,
	}
	s.cmDADescribeCmd = &cobra.Command{
		Use: "describe DA", Short: "Describe a DNS authorization",
		Args: cobra.ExactArgs(1), RunE: s.runCMDADescribe,
	}
	s.cmDAListCmd = &cobra.Command{
		Use: "list", Short: "List DNS authorizations",
		Args: cobra.NoArgs, RunE: s.runCMDAList,
	}
	s.cmDAUpdateCmd = &cobra.Command{
		Use: "update DA", Short: "Update a DNS authorization from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runCMDAUpdate,
	}
	s.cmIssuanceConfigsCmd = &cobra.Command{Use: "issuance-configs", Short: "Manage certificate issuance configs"}
	s.cmICCreateCmd = &cobra.Command{
		Use: "create IC", Short: "Create an issuance config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runCMICCreate,
	}
	s.cmICDeleteCmd = &cobra.Command{
		Use: "delete IC", Short: "Delete an issuance config",
		Args: cobra.ExactArgs(1), RunE: s.runCMICDelete,
	}
	s.cmICDescribeCmd = &cobra.Command{
		Use: "describe IC", Short: "Describe an issuance config",
		Args: cobra.ExactArgs(1), RunE: s.runCMICDescribe,
	}
	s.cmICListCmd = &cobra.Command{
		Use: "list", Short: "List issuance configs",
		Args: cobra.NoArgs, RunE: s.runCMICList,
	}
	s.cmICUpdateCmd = &cobra.Command{
		Use: "update IC", Short: "Update an issuance config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runCMICUpdate,
	}
	s.cmMapsCmd = &cobra.Command{Use: "maps", Short: "Manage certificate maps"}
	s.cmMapCreateCmd = &cobra.Command{
		Use: "create MAP", Short: "Create a certificate map from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runCMMapCreate,
	}
	s.cmMapDeleteCmd = &cobra.Command{
		Use: "delete MAP", Short: "Delete a certificate map",
		Args: cobra.ExactArgs(1), RunE: s.runCMMapDelete,
	}
	s.cmMapDescribeCmd = &cobra.Command{
		Use: "describe MAP", Short: "Describe a certificate map",
		Args: cobra.ExactArgs(1), RunE: s.runCMMapDescribe,
	}
	s.cmMapListCmd = &cobra.Command{
		Use: "list", Short: "List certificate maps",
		Args: cobra.NoArgs, RunE: s.runCMMapList,
	}
	s.cmMapUpdateCmd = &cobra.Command{
		Use: "update MAP", Short: "Update a certificate map from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runCMMapUpdate,
	}
	s.cmMapEntriesCmd = &cobra.Command{Use: "entries", Short: "Manage certificate map entries"}
	s.cmMECreateCmd = &cobra.Command{
		Use: "create ENTRY", Short: "Create a map entry from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runCMMECreate,
	}
	s.cmMEDeleteCmd = &cobra.Command{
		Use: "delete ENTRY", Short: "Delete a map entry",
		Args: cobra.ExactArgs(1), RunE: s.runCMMEDelete,
	}
	s.cmMEDescribeCmd = &cobra.Command{
		Use: "describe ENTRY", Short: "Describe a map entry",
		Args: cobra.ExactArgs(1), RunE: s.runCMMEDescribe,
	}
	s.cmMEListCmd = &cobra.Command{
		Use: "list", Short: "List map entries",
		Args: cobra.NoArgs, RunE: s.runCMMEList,
	}
	s.cmMEUpdateCmd = &cobra.Command{
		Use: "update ENTRY", Short: "Update a map entry from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runCMMEUpdate,
	}
	s.cmTrustConfigsCmd = &cobra.Command{Use: "trust-configs", Short: "Manage trust configs"}
	s.cmTCCreateCmd = &cobra.Command{
		Use: "create TC", Short: "Create a trust config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runCMTCCreate,
	}
	s.cmTCDeleteCmd = &cobra.Command{
		Use: "delete TC", Short: "Delete a trust config",
		Args: cobra.ExactArgs(1), RunE: s.runCMTCDelete,
	}
	s.cmTCDescribeCmd = &cobra.Command{
		Use: "describe TC", Short: "Describe a trust config",
		Args: cobra.ExactArgs(1), RunE: s.runCMTCDescribe,
	}
	s.cmTCListCmd = &cobra.Command{
		Use: "list", Short: "List trust configs",
		Args: cobra.NoArgs, RunE: s.runCMTCList,
	}
	s.cmTCUpdateCmd = &cobra.Command{
		Use: "update TC", Short: "Update a trust config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runCMTCUpdate,
	}
	s.cmOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Certificate Manager operations"}
	s.cmOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Certificate Manager operation",
		Args: cobra.ExactArgs(1), RunE: s.runCMOpDescribe,
	}
	s.cmOpListCmd = &cobra.Command{
		Use: "list", Short: "List Certificate Manager operations in a location",
		Args: cobra.NoArgs, RunE: s.runCMOpList,
	}
	s.initCertificateManager()
	s.initCertificateManagerAll()
	s.initCertificateManagerAll2()
	s.initCertificateManagerAll3()
	s.initCertificateManagerAll4()
	s.initCertificateManagerAll5()
	s.initCertificateManagerAll6()
	return s
}
