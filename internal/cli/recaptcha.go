package cli

import "github.com/spf13/cobra"

// --- gcloud recaptcha (#378) ---

func (s *recaptchaCmds) initRecaptcha() {

}

// recaptchaCmds holds the commands and flag values of one invocation's "recaptcha"
// command group, so invocations don't share state (#2070).
type recaptchaCmds struct {
	inv *invocation

	recaptchaCmd                     *cobra.Command
	recaptchaFirewallPoliciesCmd     *cobra.Command
	flagRcpFPDescription             string
	flagRcpFPPath                    string
	flagRcpFPCondition               string
	flagRcpFPActions                 string
	flagRcpFPNames                   []string
	flagRcpFPFormat                  string
	flagRcpFPPageSize                int64
	flagRcpFPUpdateMask              string
	recaptchaFPCreateCmd             *cobra.Command
	recaptchaFPDeleteCmd             *cobra.Command
	recaptchaFPDescribeCmd           *cobra.Command
	recaptchaFPListCmd               *cobra.Command
	recaptchaFPReorderCmd            *cobra.Command
	recaptchaFPUpdateCmd             *cobra.Command
	recaptchaKeysCmd                 *cobra.Command
	flagRcpKeyConfigFile             string
	flagRcpKeyDisplayName            string
	flagRcpKeyLabels                 []string
	flagRcpKeyTestingScore           float64
	flagRcpKeyIntegration            string
	flagRcpKeyWeb                    bool
	flagRcpKeyAndroid                bool
	flagRcpKeyIOS                    bool
	flagRcpKeyExpress                bool
	flagRcpKeyUniversal              bool
	flagRcpKeyAllowAllDoms           bool
	flagRcpKeyDomains                []string
	flagRcpKeyAllowAllBundle         bool
	flagRcpKeyBundleIDs              []string
	flagRcpKeyAllowAllPkgs           bool
	flagRcpKeyPackageNames           []string
	flagRcpKeyIP                     string
	flagRcpKeyOverride               string
	flagRcpKeySkipBilling            bool
	flagRcpKeyUpdateMask             string
	flagRcpKeyFormat                 string
	flagRcpKeyPageSize               int64
	recaptchaKeysAddIpOverrideCmd    *cobra.Command
	recaptchaKeysCreateCmd           *cobra.Command
	recaptchaKeysDeleteCmd           *cobra.Command
	recaptchaKeysDescribeCmd         *cobra.Command
	recaptchaKeysListCmd             *cobra.Command
	recaptchaKeysListIpOverridesCmd  *cobra.Command
	recaptchaKeysMigrateCmd          *cobra.Command
	recaptchaKeysRemoveIpOverrideCmd *cobra.Command
	recaptchaKeysUpdateCmd           *cobra.Command
}

// newRecaptchaCmd returns the "recaptcha" command group for inv.
func newRecaptchaCmd(inv *invocation) *cobra.Command {
	return newRecaptchaCmds(inv).recaptchaCmd
}

// newRecaptchaCmds builds the "recaptcha" command group for inv and returns its state.
func newRecaptchaCmds(inv *invocation) *recaptchaCmds {
	s := &recaptchaCmds{inv: inv}
	s.recaptchaCmd = &cobra.Command{Use: "recaptcha", Short: "Manage reCAPTCHA"}
	s.recaptchaFirewallPoliciesCmd = &cobra.Command{
		Use:   "firewall-policies",
		Short: "Manage reCAPTCHA firewall policies",
	}
	s.recaptchaFPCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a reCAPTCHA firewall policy",
		Args: cobra.NoArgs, RunE: s.runRcpFPCreate,
	}
	s.recaptchaFPDeleteCmd = &cobra.Command{
		Use: "delete POLICY_ID", Short: "Delete a reCAPTCHA firewall policy",
		Args: cobra.ExactArgs(1), RunE: s.runRcpFPDelete,
	}
	s.recaptchaFPDescribeCmd = &cobra.Command{
		Use: "describe POLICY_ID", Short: "Describe a reCAPTCHA firewall policy",
		Args: cobra.ExactArgs(1), RunE: s.runRcpFPDescribe,
	}
	s.recaptchaFPListCmd = &cobra.Command{
		Use: "list", Short: "List reCAPTCHA firewall policies",
		Args: cobra.NoArgs, RunE: s.runRcpFPList,
	}
	s.recaptchaFPReorderCmd = &cobra.Command{
		Use: "reorder", Short: "Reorder reCAPTCHA firewall policies",
		Args: cobra.NoArgs, RunE: s.runRcpFPReorder,
	}
	s.recaptchaFPUpdateCmd = &cobra.Command{
		Use: "update POLICY_ID", Short: "Update a reCAPTCHA firewall policy",
		Args: cobra.ExactArgs(1), RunE: s.runRcpFPUpdate,
	}
	s.recaptchaKeysCmd = &cobra.Command{Use: "keys", Short: "Manage reCAPTCHA keys"}
	s.recaptchaKeysAddIpOverrideCmd = &cobra.Command{
		Use: "add-ip-override KEY", Short: "Add an IP override to a reCAPTCHA key",
		Args: cobra.ExactArgs(1), RunE: s.runRcpKeyAddIpOverride,
	}
	s.recaptchaKeysCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a reCAPTCHA key",
		Args: cobra.NoArgs, RunE: s.runRcpKeyCreate,
	}
	s.recaptchaKeysDeleteCmd = &cobra.Command{
		Use: "delete KEY", Short: "Delete a reCAPTCHA key",
		Args: cobra.ExactArgs(1), RunE: s.runRcpKeyDelete,
	}
	s.recaptchaKeysDescribeCmd = &cobra.Command{
		Use: "describe KEY", Short: "Describe a reCAPTCHA key",
		Args: cobra.ExactArgs(1), RunE: s.runRcpKeyDescribe,
	}
	s.recaptchaKeysListCmd = &cobra.Command{
		Use: "list", Short: "List reCAPTCHA keys",
		Args: cobra.NoArgs, RunE: s.runRcpKeyList,
	}
	s.recaptchaKeysListIpOverridesCmd = &cobra.Command{
		Use: "list-ip-overrides KEY", Short: "List IP overrides for a reCAPTCHA key",
		Args: cobra.ExactArgs(1), RunE: s.runRcpKeyListIpOverrides,
	}
	s.recaptchaKeysMigrateCmd = &cobra.Command{
		Use: "migrate KEY", Short: "Migrate a legacy reCAPTCHA key to reCAPTCHA Enterprise",
		Args: cobra.ExactArgs(1), RunE: s.runRcpKeyMigrate,
	}
	s.recaptchaKeysRemoveIpOverrideCmd = &cobra.Command{
		Use: "remove-ip-override KEY", Short: "Remove an IP override from a reCAPTCHA key",
		Args: cobra.ExactArgs(1), RunE: s.runRcpKeyRemoveIpOverride,
	}
	s.recaptchaKeysUpdateCmd = &cobra.Command{
		Use: "update KEY", Short: "Update a reCAPTCHA key",
		Args: cobra.ExactArgs(1), RunE: s.runRcpKeyUpdate,
	}
	s.initRecaptcha()
	s.initRecaptchaFirewallPolicies()
	s.initRecaptchaKeys()
	return s
}
