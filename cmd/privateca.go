package cmd

import "github.com/spf13/cobra"

// --- gcloud privateca (#374) ---

func (s *privatecaCmds) initPrivateca() {
	// All subgroups (certificates, locations, operations, pools, roots,
	// subordinates, templates) are implemented in dedicated privateca_*.go
	// files.

}

// privatecaCmds holds the commands and flag values of one invocation's "privateca"
// command group, so invocations don't share state (#2070).
type privatecaCmds struct {
	inv *invocation

	privatecaCmd             *cobra.Command
	privatecaLocationsCmd    *cobra.Command
	pcaLocDescribeCmd        *cobra.Command
	pcaLocListCmd            *cobra.Command
	flagPCAFormat            string
	privatecaOperationsCmd   *cobra.Command
	pcaOpCancelCmd           *cobra.Command
	pcaOpDeleteCmd           *cobra.Command
	pcaOpDescribeCmd         *cobra.Command
	pcaOpListCmd             *cobra.Command
	flagPCAOpLocation        string
	privatecaPoolsCmd        *cobra.Command
	pcaPoolCreateCmd         *cobra.Command
	pcaPoolDeleteCmd         *cobra.Command
	pcaPoolDescribeCmd       *cobra.Command
	pcaPoolListCmd           *cobra.Command
	pcaPoolUpdateCmd         *cobra.Command
	pcaPoolGetCaCertsCmd     *cobra.Command
	pcaPoolGetIamCmd         *cobra.Command
	pcaPoolSetIamCmd         *cobra.Command
	pcaPoolAddIamCmd         *cobra.Command
	pcaPoolRemoveIamCmd      *cobra.Command
	flagPCAPoolLocation      string
	flagPCAPoolConfigFile    string
	flagPCAPoolUpdateMask    string
	flagPCAPoolAsync         bool
	flagPCAPoolIamMember     string
	flagPCAPoolIamRole       string
	privatecaCertificatesCmd *cobra.Command
	pcaCertCreateCmd         *cobra.Command
	pcaCertDescribeCmd       *cobra.Command
	pcaCertListCmd           *cobra.Command
	pcaCertUpdateCmd         *cobra.Command
	pcaCertRevokeCmd         *cobra.Command
	pcaCertExportCmd         *cobra.Command
	flagPCACertLocation      string
	flagPCACertIssuerPool    string
	flagPCACertConfigFile    string
	flagPCACertUpdateMask    string
	flagPCACertRevokeReason  string
	flagPCACertOutputFile    string
	// Certificate.RequestedNotBeforeTime (gcloud-python 569.0.0). Only valid
	// when the issuing CA pool's issuancePolicy has
	// `allowRequesterSpecifiedNotBeforeTime: true`.
	flagPCACertNotBefore string
	// --generate-key / --key-output-file / --key-algorithm on
	// `certificates create`: generate a key pair locally, write the private
	// key to --key-output-file and send the public key as
	// config.publicKey (gcloud-python 588.0.0 added --key-algorithm).
	flagPCACertGenerateKey   bool
	flagPCACertKeyOutputFile string
	flagPCACertKeyAlgorithm  string
	privatecaRootsCmd        *cobra.Command
	privatecaSubordinatesCmd *cobra.Command
	flagPCACAPool            string
	flagPCACALocation        string
	flagPCACAConfigFile      string
	flagPCACAUpdateMask      string
	flagPCACAAsync           bool
	flagPCACAIgnoreActive    bool
	flagPCACAPemCert         string
	flagPCACASkipGrace       bool
	flagPCACAIgnoreDeps      bool
	// --- first-party activation on `subordinates activate` (580.0.0). ---
	flagPCACAIssuerPool     string
	flagPCACAIssuerLocation string
	flagPCACAIssuerCA       string
	privatecaTemplatesCmd   *cobra.Command
	pcaTplCreateCmd         *cobra.Command
	pcaTplDeleteCmd         *cobra.Command
	pcaTplDescribeCmd       *cobra.Command
	pcaTplListCmd           *cobra.Command
	pcaTplUpdateCmd         *cobra.Command
	pcaTplReplicateCmd      *cobra.Command
	pcaTplGetIamCmd         *cobra.Command
	pcaTplSetIamCmd         *cobra.Command
	pcaTplAddIamCmd         *cobra.Command
	pcaTplRemoveIamCmd      *cobra.Command
	flagPCATplLocation      string
	flagPCATplConfigFile    string
	flagPCATplUpdateMask    string
	flagPCATplAsync         bool
	flagPCATplReplicateLoc  string
	flagPCATplIamMember     string
	flagPCATplIamRole       string
}

// newPrivatecaCmd returns the "privateca" command group for inv.
func newPrivatecaCmd(inv *invocation) *cobra.Command {
	return newPrivatecaCmds(inv).privatecaCmd
}

// newPrivatecaCmds builds the "privateca" command group for inv and returns its state.
func newPrivatecaCmds(inv *invocation) *privatecaCmds {
	s := &privatecaCmds{inv: inv}
	s.privatecaCmd = &cobra.Command{Use: "privateca", Short: "Manage Private CA"}
	s.privatecaLocationsCmd = &cobra.Command{
		Use:   "locations",
		Short: "Explore Private CA locations",
	}
	s.pcaLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Private CA location",
		Args: cobra.ExactArgs(1), RunE: s.runPCALocDescribe,
	}
	s.pcaLocListCmd = &cobra.Command{
		Use: "list", Short: "List Private CA locations",
		Args: cobra.NoArgs, RunE: s.runPCALocList,
	}
	s.privatecaOperationsCmd = &cobra.Command{
		Use:   "operations",
		Short: "Manage Private CA operations",
	}
	s.pcaOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Private CA operation",
		Args: cobra.ExactArgs(1), RunE: s.runPCAOpCancel,
	}
	s.pcaOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Private CA operation",
		Args: cobra.ExactArgs(1), RunE: s.runPCAOpDelete,
	}
	s.pcaOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Private CA operation",
		Args: cobra.ExactArgs(1), RunE: s.runPCAOpDescribe,
	}
	s.pcaOpListCmd = &cobra.Command{
		Use: "list", Short: "List Private CA operations in a location",
		Args: cobra.NoArgs, RunE: s.runPCAOpList,
	}
	s.privatecaPoolsCmd = &cobra.Command{
		Use:   "pools",
		Short: "Manage Private CA pools",
	}
	s.pcaPoolCreateCmd = &cobra.Command{
		Use: "create POOL", Short: "Create a CA pool from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runPCAPoolCreate,
	}
	s.pcaPoolDeleteCmd = &cobra.Command{
		Use: "delete POOL", Short: "Delete a CA pool",
		Args: cobra.ExactArgs(1), RunE: s.runPCAPoolDelete,
	}
	s.pcaPoolDescribeCmd = &cobra.Command{
		Use: "describe POOL", Short: "Describe a CA pool",
		Args: cobra.ExactArgs(1), RunE: s.runPCAPoolDescribe,
	}
	s.pcaPoolListCmd = &cobra.Command{
		Use: "list", Short: "List CA pools in a location",
		Args: cobra.NoArgs, RunE: s.runPCAPoolList,
	}
	s.pcaPoolUpdateCmd = &cobra.Command{
		Use: "update POOL", Short: "Update a CA pool from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runPCAPoolUpdate,
	}
	s.pcaPoolGetCaCertsCmd = &cobra.Command{
		Use: "get-ca-certs POOL", Short: "Print PEM-encoded CA certificates for a pool",
		Args: cobra.ExactArgs(1), RunE: s.runPCAPoolGetCaCerts,
	}
	s.pcaPoolGetIamCmd = &cobra.Command{
		Use: "get-iam-policy POOL", Short: "Print the IAM policy for a CA pool",
		Args: cobra.ExactArgs(1), RunE: s.runPCAPoolGetIam,
	}
	s.pcaPoolSetIamCmd = &cobra.Command{
		Use: "set-iam-policy POOL POLICY_FILE", Short: "Replace the IAM policy for a CA pool",
		Args: cobra.ExactArgs(2), RunE: s.runPCAPoolSetIam,
	}
	s.pcaPoolAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding POOL", Short: "Add an IAM binding to a CA pool",
		Args: cobra.ExactArgs(1), RunE: s.runPCAPoolAddIam,
	}
	s.pcaPoolRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding POOL", Short: "Remove an IAM binding from a CA pool",
		Args: cobra.ExactArgs(1), RunE: s.runPCAPoolRemoveIam,
	}
	s.privatecaCertificatesCmd = &cobra.Command{
		Use:   "certificates",
		Short: "Manage Private CA certificates",
	}
	s.pcaCertCreateCmd = &cobra.Command{
		Use: "create CERTIFICATE", Short: "Create a certificate from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runPCACertCreate,
	}
	s.pcaCertDescribeCmd = &cobra.Command{
		Use: "describe CERTIFICATE", Short: "Describe a certificate",
		Args: cobra.ExactArgs(1), RunE: s.runPCACertDescribe,
	}
	s.pcaCertListCmd = &cobra.Command{
		Use: "list", Short: "List certificates in a CA pool",
		Args: cobra.NoArgs, RunE: s.runPCACertList,
	}
	s.pcaCertUpdateCmd = &cobra.Command{
		Use: "update CERTIFICATE", Short: "Update a certificate from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runPCACertUpdate,
	}
	s.pcaCertRevokeCmd = &cobra.Command{
		Use: "revoke CERTIFICATE", Short: "Revoke a certificate",
		Args: cobra.ExactArgs(1), RunE: s.runPCACertRevoke,
	}
	s.pcaCertExportCmd = &cobra.Command{
		Use: "export CERTIFICATE", Short: "Print the PEM-encoded certificate to stdout",
		Args: cobra.ExactArgs(1), RunE: s.runPCACertExport,
	}
	s.privatecaRootsCmd = &cobra.Command{
		Use:   "roots",
		Short: "Manage Private CA root certificate authorities",
	}
	s.privatecaSubordinatesCmd = &cobra.Command{
		Use:   "subordinates",
		Short: "Manage Private CA subordinate certificate authorities",
	}
	s.privatecaTemplatesCmd = &cobra.Command{
		Use:   "templates",
		Short: "Manage Private CA certificate templates",
	}
	s.pcaTplCreateCmd = &cobra.Command{
		Use: "create TEMPLATE", Short: "Create a certificate template from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runPCATplCreate,
	}
	s.pcaTplDeleteCmd = &cobra.Command{
		Use: "delete TEMPLATE", Short: "Delete a certificate template",
		Args: cobra.ExactArgs(1), RunE: s.runPCATplDelete,
	}
	s.pcaTplDescribeCmd = &cobra.Command{
		Use: "describe TEMPLATE", Short: "Describe a certificate template",
		Args: cobra.ExactArgs(1), RunE: s.runPCATplDescribe,
	}
	s.pcaTplListCmd = &cobra.Command{
		Use: "list", Short: "List certificate templates in a location",
		Args: cobra.NoArgs, RunE: s.runPCATplList,
	}
	s.pcaTplUpdateCmd = &cobra.Command{
		Use: "update TEMPLATE", Short: "Update a certificate template from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runPCATplUpdate,
	}
	s.pcaTplReplicateCmd = &cobra.Command{
		Use: "replicate TEMPLATE", Short: "Replicate a certificate template to another location",
		Args: cobra.ExactArgs(1), RunE: s.runPCATplReplicate,
	}
	s.pcaTplGetIamCmd = &cobra.Command{
		Use: "get-iam-policy TEMPLATE", Short: "Print the IAM policy for a certificate template",
		Args: cobra.ExactArgs(1), RunE: s.runPCATplGetIam,
	}
	s.pcaTplSetIamCmd = &cobra.Command{
		Use: "set-iam-policy TEMPLATE POLICY_FILE", Short: "Replace the IAM policy",
		Args: cobra.ExactArgs(2), RunE: s.runPCATplSetIam,
	}
	s.pcaTplAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding TEMPLATE", Short: "Add an IAM binding",
		Args: cobra.ExactArgs(1), RunE: s.runPCATplAddIam,
	}
	s.pcaTplRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding TEMPLATE", Short: "Remove an IAM binding",
		Args: cobra.ExactArgs(1), RunE: s.runPCATplRemoveIam,
	}
	s.initPrivateca()
	s.initPrivatecaAll()
	s.initPrivatecaAll2()
	s.initPrivatecaAll3()
	s.initPrivatecaAll4()
	s.initPrivatecaCertificateAuthorities()
	s.initPrivatecaTemplates()
	return s
}
