package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms (#349) ---

func (s *kmsCmds) initKms() {

}

// kmsLocationParent returns "projects/PROJECT/locations/LOCATION".
func kmsLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

// kmsKeyringParent returns
// "projects/PROJECT/locations/LOCATION/keyRings/KEYRING".
func kmsKeyringParent(project, location, keyring string) string {
	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", project, location, keyring)
}

// kmsKeyName returns
// "projects/PROJECT/locations/LOCATION/keyRings/KEYRING/cryptoKeys/KEY".
func kmsKeyName(project, location, keyring, key string) string {
	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		project, location, keyring, key)
}

// kmsFullName returns raw if it already looks like a fully-qualified resource
// name; otherwise it joins parent and raw.
func kmsFullName(parent, raw string) string {
	if strings.HasPrefix(raw, "projects/") || strings.HasPrefix(raw, "folders/") ||
		strings.HasPrefix(raw, "organizations/") {
		return raw
	}
	return parent + "/" + raw
}

// --- shared IAM helpers ---

// kmsIamBuildCondition returns a *cloudkms.Expr or nil.
func kmsIamBuildCondition(expr, title, desc string) *cloudkms.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &cloudkms.Expr{Expression: expr, Title: title, Description: desc}
}

// kmsIamCondsEqual reports whether two IAM condition exprs are equivalent.
func kmsIamCondsEqual(a, b *cloudkms.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

// kmsIamAddBinding adds (role, member[, cond]) to policy in-place.
func kmsIamAddBinding(policy *cloudkms.Policy, role, member string, cond *cloudkms.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !kmsIamCondsEqual(b.Condition, cond) {
			continue
		}
		for _, m := range b.Members {
			if m == member {
				return
			}
		}
		b.Members = append(b.Members, member)
		return
	}
	policy.Bindings = append(policy.Bindings, &cloudkms.Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

// kmsIamRemoveBinding removes (role, member[, cond]) from policy in-place. If
// allConds is true, matching bindings across all conditions are cleared.
func kmsIamRemoveBinding(policy *cloudkms.Policy, role, member string, cond *cloudkms.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || kmsIamCondsEqual(b.Condition, cond))
		if !match {
			kept = append(kept, b)
			continue
		}
		newMembers := b.Members[:0]
		for _, m := range b.Members {
			if m == member {
				continue
			}
			newMembers = append(newMembers, m)
		}
		if len(newMembers) != len(b.Members) {
			changed = true
		}
		b.Members = newMembers
		if len(b.Members) > 0 {
			kept = append(kept, b)
		} else {
			changed = true
		}
	}
	policy.Bindings = kept
	return changed
}

// kmsUpdatedIam prints the standard "Updated IAM policy for ..." message.
func (s *kmsCmds) kmsUpdatedIam(who string) {
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for %s.\n", who)
}

// kmsIamMemberFlags binds the common --member/--role/--condition-* flags for a
// (add|remove)-iam-policy-binding subcommand.
func kmsIamMemberFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

// kmsCmds holds the commands and flag values of one invocation's "kms"
// command group, so invocations don't share state (#2070).
type kmsCmds struct {
	inv *invocation

	kmsCmd                     *cobra.Command
	kmsAutokeyConfigCmd        *cobra.Command
	flagKmsAutokeyFolder       string
	flagKmsAutokeyProject      string
	flagKmsAutokeyFormat       string
	flagKmsAutokeyConfigFile   string
	flagKmsAutokeyMask         string
	flagKmsAutokeyKeyProject   string
	kmsAutokeyDescribeCmd      *cobra.Command
	kmsAutokeyShowEffectiveCmd *cobra.Command
	kmsAutokeyUpdateCmd        *cobra.Command
	kmsEkmConfigCmd            *cobra.Command
	flagKmsEkmCfgLocation      string
	flagKmsEkmCfgFormat        string
	flagKmsEkmCfgConfigFile    string
	flagKmsEkmCfgMask          string
	flagKmsEkmCfgAllConds      bool
	flagKmsEkmCfgIamMember     string
	flagKmsEkmCfgIamRole       string
	flagKmsEkmCfgIamCondExpr   string
	flagKmsEkmCfgIamCondTitle  string
	flagKmsEkmCfgIamCondDesc   string
	kmsEkmCfgDescribeCmd       *cobra.Command
	kmsEkmCfgUpdateCmd         *cobra.Command
	kmsEkmCfgGetIamCmd         *cobra.Command
	kmsEkmCfgSetIamCmd         *cobra.Command
	kmsEkmCfgAddIamCmd         *cobra.Command
	kmsEkmCfgRemoveIamCmd      *cobra.Command
	kmsEkmConnectionsCmd       *cobra.Command
	flagKmsEkmConnLocation     string
	flagKmsEkmConnFormat       string
	flagKmsEkmConnFilter       string
	flagKmsEkmConnPageSize     int64
	flagKmsEkmConnConfigFile   string
	flagKmsEkmConnMask         string
	flagKmsEkmConnAllConds     bool
	flagKmsEkmConnIamMember    string
	flagKmsEkmConnIamRole      string
	flagKmsEkmConnIamCondExpr  string
	flagKmsEkmConnIamCondTitle string
	flagKmsEkmConnIamCondDesc  string
	kmsEkmConnCreateCmd        *cobra.Command
	kmsEkmConnDescribeCmd      *cobra.Command
	kmsEkmConnListCmd          *cobra.Command
	kmsEkmConnUpdateCmd        *cobra.Command
	kmsEkmConnGetIamCmd        *cobra.Command
	kmsEkmConnSetIamCmd        *cobra.Command
	kmsEkmConnAddIamCmd        *cobra.Command
	kmsEkmConnRemoveIamCmd     *cobra.Command
	kmsImportJobsCmd           *cobra.Command
	flagKmsImpLocation         string
	flagKmsImpKeyring          string
	flagKmsImpFormat           string
	flagKmsImpFilter           string
	flagKmsImpPageSize         int64
	flagKmsImpConfigFile       string
	flagKmsImpAllConds         bool
	flagKmsImpIamMember        string
	flagKmsImpIamRole          string
	flagKmsImpIamCondExpr      string
	flagKmsImpIamCondTitle     string
	flagKmsImpIamCondDesc      string
	kmsImpCreateCmd            *cobra.Command
	kmsImpDescribeCmd          *cobra.Command
	kmsImpListCmd              *cobra.Command
	kmsImpGetIamCmd            *cobra.Command
	kmsImpSetIamCmd            *cobra.Command
	kmsImpAddIamCmd            *cobra.Command
	kmsImpRemoveIamCmd         *cobra.Command
	kmsInventoryCmd            *cobra.Command
	flagKmsInvFormat           string
	flagKmsInvPageSize         int64
	flagKmsInvScope            string
	flagKmsInvCryptoKey        string
	flagKmsInvResourceTypes    []string
	flagKmsInvKey              string
	flagKmsInvFallbackScope    string
	kmsInvSummaryCmd           *cobra.Command
	kmsInvListKeysCmd          *cobra.Command
	kmsInvSearchCmd            *cobra.Command
	kmsKeyHandlesCmd           *cobra.Command
	flagKmsKHLocation          string
	flagKmsKHFormat            string
	flagKmsKHFilter            string
	flagKmsKHPageSize          int64
	flagKmsKHConfigFile        string
	flagKmsKHHandleID          string
	kmsKHCreateCmd             *cobra.Command
	kmsKHDescribeCmd           *cobra.Command
	kmsKHListCmd               *cobra.Command
	kmsKeyringsCmd             *cobra.Command
	flagKmsKRLocation          string
	flagKmsKRFormat            string
	flagKmsKRFilter            string
	flagKmsKRPageSize          int64
	flagKmsKRConfigFile        string
	flagKmsKRAllConds          bool
	flagKmsKRIamMember         string
	flagKmsKRIamRole           string
	flagKmsKRIamCondExpr       string
	flagKmsKRIamCondTitle      string
	flagKmsKRIamCondDesc       string
	kmsKRCreateCmd             *cobra.Command
	kmsKRDeleteCmd             *cobra.Command
	kmsKRDescribeCmd           *cobra.Command
	kmsKRListCmd               *cobra.Command
	kmsKRGetIamCmd             *cobra.Command
	kmsKRSetIamCmd             *cobra.Command
	kmsKRAddIamCmd             *cobra.Command
	kmsKRRemoveIamCmd          *cobra.Command
	kmsKeysCmd                 *cobra.Command
	flagKmsKLocation           string
	flagKmsKKeyring            string
	flagKmsKFormat             string
	flagKmsKFilter             string
	flagKmsKPageSize           int64
	flagKmsKConfigFile         string
	flagKmsKMask               string
	flagKmsKAllConds           bool
	flagKmsKIamMember          string
	flagKmsKIamRole            string
	flagKmsKIamCondExpr        string
	flagKmsKIamCondTitle       string
	flagKmsKIamCondDesc        string
	flagKmsKPurpose            string
	flagKmsKProtectionLevel    string
	flagKmsKAlgorithm          string
	flagKmsKRotationPeriod     string
	flagKmsKNextRotation       string
	flagKmsKHsmTrustedWrapping bool
	flagKmsKBackend            string
	flagKmsKPrimaryVersion     string
	kmsKCreateCmd              *cobra.Command
	kmsKDeleteCmd              *cobra.Command
	kmsKDescribeCmd            *cobra.Command
	kmsKListCmd                *cobra.Command
	kmsKUpdateCmd              *cobra.Command
	kmsKSetPrimaryCmd          *cobra.Command
	kmsKSetRotationCmd         *cobra.Command
	kmsKRemoveRotationCmd      *cobra.Command
	kmsKGetIamCmd              *cobra.Command
	kmsKSetIamCmd              *cobra.Command
	kmsKAddIamCmd              *cobra.Command
	kmsKRemoveIamCmd           *cobra.Command
	kmsVersionsCmd             *cobra.Command
	flagKmsVKey                string
	flagKmsVFormat             string
	flagKmsVFilter             string
	flagKmsVPageSize           int64
	flagKmsVConfigFile         string
	flagKmsVAlgorithm          string
	flagKmsVProtection         string
	flagKmsVBackend            string
	flagKmsVState              string
	flagKmsVMask               string
	kmsVCreateCmd              *cobra.Command
	kmsVDescribeCmd            *cobra.Command
	kmsVDestroyCmd             *cobra.Command
	kmsVDisableCmd             *cobra.Command
	kmsVEnableCmd              *cobra.Command
	kmsVListCmd                *cobra.Command
	kmsVRestoreCmd             *cobra.Command
	// kmsVUpdateCmd was added in gcloud-python 582.0.0. It supports patching
	// individual CryptoKeyVersion fields via named flags in addition to
	// --config-file / --update-mask.
	kmsVUpdateCmd                *cobra.Command
	kmsLocationsCmd              *cobra.Command
	flagKmsLocFormat             string
	flagKmsLocFilter             string
	flagKmsLocPageSize           int64
	kmsLocListCmd                *cobra.Command
	kmsOperationsCmd             *cobra.Command
	flagKmsOpLocation            string
	flagKmsOpFormat              string
	kmsOpDescribeCmd             *cobra.Command
	kmsRetiredCmd                *cobra.Command
	flagKmsRRLocation            string
	flagKmsRRFormat              string
	flagKmsRRPageSize            int64
	kmsRRDescribeCmd             *cobra.Command
	kmsRRListCmd                 *cobra.Command
	kmsSTHsmCmd                  *cobra.Command
	flagKmsHsmLocation           string
	flagKmsHsmFormat             string
	flagKmsHsmFilter             string
	flagKmsHsmPageSize           int64
	flagKmsHsmConfigFile         string
	flagKmsHsmInstance           string
	flagKmsHsmOperationType      string
	flagKmsHsmCryptoKeyVersion   string
	flagKmsHsmTwoFactorPublicKey string
	kmsHsmCreateCmd              *cobra.Command
	kmsHsmDescribeCmd            *cobra.Command
	kmsHsmListCmd                *cobra.Command
	kmsHsmProposalCmd            *cobra.Command
	kmsHsmProposalCreateCmd      *cobra.Command
	kmsHsmProposalDescribeCmd    *cobra.Command
	kmsHsmProposalListCmd        *cobra.Command
}

// newKmsCmd returns the "kms" command group for inv.
func newKmsCmd(inv *invocation) *cobra.Command {
	return newKmsCmds(inv).kmsCmd
}

// newKmsCmds builds the "kms" command group for inv and returns its state.
func newKmsCmds(inv *invocation) *kmsCmds {
	s := &kmsCmds{inv: inv}
	s.kmsCmd = &cobra.Command{Use: "kms", Short: "Manage Cloud KMS"}
	s.kmsAutokeyConfigCmd = &cobra.Command{
		Use:   "autokey-config",
		Short: "Manage Cloud KMS Autokey configurations",
	}
	s.kmsAutokeyDescribeCmd = &cobra.Command{
		Use:   "describe",
		Short: "Describe the AutokeyConfig for a folder or project",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsAutokeyDescribe,
	}
	s.kmsAutokeyShowEffectiveCmd = &cobra.Command{
		Use:   "show-effective-config",
		Short: "Show the effective AutokeyConfig for a folder or project",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsAutokeyShowEffective,
	}
	s.kmsAutokeyUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the AutokeyConfig for a folder or project",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsAutokeyUpdate,
	}
	s.kmsEkmConfigCmd = &cobra.Command{
		Use:   "ekm-config",
		Short: "Manage per-location EkmConfig resources",
	}
	s.kmsEkmCfgDescribeCmd = &cobra.Command{
		Use:   "describe",
		Short: "Describe the EkmConfig for a location",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsEkmCfgDescribe,
	}
	s.kmsEkmCfgUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the EkmConfig for a location",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsEkmCfgUpdate,
	}
	s.kmsEkmCfgGetIamCmd = &cobra.Command{
		Use:   "get-iam-policy",
		Short: "Get the IAM policy for the EkmConfig",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsEkmCfgGetIam,
	}
	s.kmsEkmCfgSetIamCmd = &cobra.Command{
		Use:   "set-iam-policy POLICY_FILE",
		Short: "Set the IAM policy for the EkmConfig",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsEkmCfgSetIam,
	}
	s.kmsEkmCfgAddIamCmd = &cobra.Command{
		Use:   "add-iam-policy-binding",
		Short: "Add an IAM policy binding on the EkmConfig",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsEkmCfgAddIam,
	}
	s.kmsEkmCfgRemoveIamCmd = &cobra.Command{
		Use:   "remove-iam-policy-binding",
		Short: "Remove an IAM policy binding from the EkmConfig",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsEkmCfgRemoveIam,
	}
	s.kmsEkmConnectionsCmd = &cobra.Command{
		Use:   "ekm-connections",
		Short: "Manage Cloud KMS EKM connections",
	}
	s.kmsEkmConnCreateCmd = &cobra.Command{
		Use:   "create CONNECTION",
		Short: "Create an EKM connection",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsEkmConnCreate,
	}
	s.kmsEkmConnDescribeCmd = &cobra.Command{
		Use:   "describe CONNECTION",
		Short: "Describe an EKM connection",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsEkmConnDescribe,
	}
	s.kmsEkmConnListCmd = &cobra.Command{
		Use:   "list",
		Short: "List EKM connections in a location",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsEkmConnList,
	}
	s.kmsEkmConnUpdateCmd = &cobra.Command{
		Use:   "update CONNECTION",
		Short: "Update an EKM connection",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsEkmConnUpdate,
	}
	s.kmsEkmConnGetIamCmd = &cobra.Command{
		Use:   "get-iam-policy CONNECTION",
		Short: "Get the IAM policy for an EKM connection",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsEkmConnGetIam,
	}
	s.kmsEkmConnSetIamCmd = &cobra.Command{
		Use:   "set-iam-policy CONNECTION POLICY_FILE",
		Short: "Set the IAM policy for an EKM connection",
		Args:  cobra.ExactArgs(2),
		RunE:  s.runKmsEkmConnSetIam,
	}
	s.kmsEkmConnAddIamCmd = &cobra.Command{
		Use:   "add-iam-policy-binding CONNECTION",
		Short: "Add an IAM policy binding on an EKM connection",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsEkmConnAddIam,
	}
	s.kmsEkmConnRemoveIamCmd = &cobra.Command{
		Use:   "remove-iam-policy-binding CONNECTION",
		Short: "Remove an IAM policy binding from an EKM connection",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsEkmConnRemoveIam,
	}
	s.kmsImportJobsCmd = &cobra.Command{
		Use:   "import-jobs",
		Short: "Manage Cloud KMS import jobs",
	}
	s.kmsImpCreateCmd = &cobra.Command{
		Use:   "create JOB",
		Short: "Create an import job",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsImpCreate,
	}
	s.kmsImpDescribeCmd = &cobra.Command{
		Use:   "describe JOB",
		Short: "Describe an import job",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsImpDescribe,
	}
	s.kmsImpListCmd = &cobra.Command{
		Use:   "list",
		Short: "List import jobs in a keyring",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsImpList,
	}
	s.kmsImpGetIamCmd = &cobra.Command{
		Use:   "get-iam-policy JOB",
		Short: "Get the IAM policy for an import job",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsImpGetIam,
	}
	s.kmsImpSetIamCmd = &cobra.Command{
		Use:   "set-iam-policy JOB POLICY_FILE",
		Short: "Set the IAM policy for an import job",
		Args:  cobra.ExactArgs(2),
		RunE:  s.runKmsImpSetIam,
	}
	s.kmsImpAddIamCmd = &cobra.Command{
		Use:   "add-iam-policy-binding JOB",
		Short: "Add an IAM policy binding on an import job",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsImpAddIam,
	}
	s.kmsImpRemoveIamCmd = &cobra.Command{
		Use:   "remove-iam-policy-binding JOB",
		Short: "Remove an IAM policy binding from an import job",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsImpRemoveIam,
	}
	s.kmsInventoryCmd = &cobra.Command{
		Use:   "inventory",
		Short: "KMS inventory and key-tracking",
	}
	s.kmsInvSummaryCmd = &cobra.Command{
		Use:   "get-protected-resources-summary",
		Short: "Get an aggregate summary of resources protected by a CryptoKey",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsInvSummary,
	}
	s.kmsInvListKeysCmd = &cobra.Command{
		Use:   "list-keys",
		Short: "List Cloud KMS keys in a project (from inventory snapshots)",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsInvListKeys,
	}
	s.kmsInvSearchCmd = &cobra.Command{
		Use:   "search-protected-resources",
		Short: "Search resources protected by a Cloud KMS CryptoKey",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsInvSearch,
	}
	s.kmsKeyHandlesCmd = &cobra.Command{
		Use:   "key-handles",
		Short: "Manage Cloud KMS KeyHandle resources",
	}
	s.kmsKHCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a KeyHandle (returns a long-running Operation)",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsKHCreate,
	}
	s.kmsKHDescribeCmd = &cobra.Command{
		Use:   "describe HANDLE",
		Short: "Describe a KeyHandle",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKHDescribe,
	}
	s.kmsKHListCmd = &cobra.Command{
		Use:   "list",
		Short: "List KeyHandles in a location",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsKHList,
	}
	s.kmsKeyringsCmd = &cobra.Command{
		Use:   "keyrings",
		Short: "Manage Cloud KMS key rings",
	}
	s.kmsKRCreateCmd = &cobra.Command{
		Use:   "create KEYRING",
		Short: "Create a key ring",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKRCreate,
	}
	s.kmsKRDeleteCmd = &cobra.Command{
		Use:   "delete KEYRING",
		Short: "Delete a key ring",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKRDelete,
	}
	s.kmsKRDescribeCmd = &cobra.Command{
		Use:   "describe KEYRING",
		Short: "Describe a key ring",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKRDescribe,
	}
	s.kmsKRListCmd = &cobra.Command{
		Use:   "list",
		Short: "List key rings in a location",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsKRList,
	}
	s.kmsKRGetIamCmd = &cobra.Command{
		Use:   "get-iam-policy KEYRING",
		Short: "Get the IAM policy for a key ring",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKRGetIam,
	}
	s.kmsKRSetIamCmd = &cobra.Command{
		Use:   "set-iam-policy KEYRING POLICY_FILE",
		Short: "Set the IAM policy for a key ring",
		Args:  cobra.ExactArgs(2),
		RunE:  s.runKmsKRSetIam,
	}
	s.kmsKRAddIamCmd = &cobra.Command{
		Use:   "add-iam-policy-binding KEYRING",
		Short: "Add an IAM policy binding on a key ring",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKRAddIam,
	}
	s.kmsKRRemoveIamCmd = &cobra.Command{
		Use:   "remove-iam-policy-binding KEYRING",
		Short: "Remove an IAM policy binding from a key ring",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKRRemoveIam,
	}
	s.kmsKeysCmd = &cobra.Command{
		Use:   "keys",
		Short: "Manage Cloud KMS crypto keys",
	}
	s.kmsKCreateCmd = &cobra.Command{
		Use:   "create KEY",
		Short: "Create a crypto key",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKCreate,
	}
	s.kmsKDeleteCmd = &cobra.Command{
		Use:   "delete KEY",
		Short: "Destroy all versions of a crypto key (Cloud KMS keys cannot be deleted directly)",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKDelete,
	}
	s.kmsKDescribeCmd = &cobra.Command{
		Use:   "describe KEY",
		Short: "Describe a crypto key",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKDescribe,
	}
	s.kmsKListCmd = &cobra.Command{
		Use:   "list",
		Short: "List crypto keys in a keyring",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsKList,
	}
	s.kmsKUpdateCmd = &cobra.Command{
		Use:   "update KEY",
		Short: "Update a crypto key",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKUpdate,
	}
	s.kmsKSetPrimaryCmd = &cobra.Command{
		Use:   "set-primary-version KEY",
		Short: "Set the primary version of a crypto key",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKSetPrimary,
	}
	s.kmsKSetRotationCmd = &cobra.Command{
		Use:   "set-rotation-schedule KEY",
		Short: "Set the rotation schedule on a crypto key",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKSetRotation,
	}
	s.kmsKRemoveRotationCmd = &cobra.Command{
		Use:   "remove-rotation-schedule KEY",
		Short: "Clear the rotation schedule on a crypto key",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKRemoveRotation,
	}
	s.kmsKGetIamCmd = &cobra.Command{
		Use:   "get-iam-policy KEY",
		Short: "Get the IAM policy for a crypto key",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKGetIam,
	}
	s.kmsKSetIamCmd = &cobra.Command{
		Use:   "set-iam-policy KEY POLICY_FILE",
		Short: "Set the IAM policy for a crypto key",
		Args:  cobra.ExactArgs(2),
		RunE:  s.runKmsKSetIam,
	}
	s.kmsKAddIamCmd = &cobra.Command{
		Use:   "add-iam-policy-binding KEY",
		Short: "Add an IAM policy binding on a crypto key",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKAddIam,
	}
	s.kmsKRemoveIamCmd = &cobra.Command{
		Use:   "remove-iam-policy-binding KEY",
		Short: "Remove an IAM policy binding from a crypto key",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsKRemoveIam,
	}
	s.kmsVersionsCmd = &cobra.Command{
		Use:   "versions",
		Short: "Manage crypto key versions",
	}
	s.kmsVCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a new CryptoKeyVersion",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsVCreate,
	}
	s.kmsVDescribeCmd = &cobra.Command{
		Use:   "describe VERSION",
		Short: "Describe a CryptoKeyVersion",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsVDescribe,
	}
	s.kmsVDestroyCmd = &cobra.Command{
		Use:   "destroy VERSION",
		Short: "Schedule a CryptoKeyVersion for destruction",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsVDestroy,
	}
	s.kmsVDisableCmd = &cobra.Command{
		Use:   "disable VERSION",
		Short: "Disable a CryptoKeyVersion",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsVDisable,
	}
	s.kmsVEnableCmd = &cobra.Command{
		Use:   "enable VERSION",
		Short: "Enable a CryptoKeyVersion",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsVEnable,
	}
	s.kmsVListCmd = &cobra.Command{
		Use:   "list",
		Short: "List CryptoKeyVersions for a CryptoKey",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsVList,
	}
	s.kmsVRestoreCmd = &cobra.Command{
		Use:   "restore VERSION",
		Short: "Restore a CryptoKeyVersion pending destruction",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsVRestore,
	}
	s.kmsVUpdateCmd = &cobra.Command{
		Use:   "update VERSION",
		Short: "Update a CryptoKeyVersion",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsVUpdate,
	}
	s.kmsLocationsCmd = &cobra.Command{
		Use:   "locations",
		Short: "View Cloud KMS locations",
	}
	s.kmsLocListCmd = &cobra.Command{
		Use:   "list",
		Short: "List available Cloud KMS locations",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsLocList,
	}
	s.kmsOperationsCmd = &cobra.Command{
		Use:   "operations",
		Short: "Manage Cloud KMS long-running operations",
	}
	s.kmsOpDescribeCmd = &cobra.Command{
		Use:   "describe OPERATION",
		Short: "Describe a Cloud KMS long-running operation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsOpDescribe,
	}
	s.kmsRetiredCmd = &cobra.Command{
		Use:   "retired-resources",
		Short: "Manage Cloud KMS retired resources",
	}
	s.kmsRRDescribeCmd = &cobra.Command{
		Use:   "describe RESOURCE",
		Short: "Describe a retired resource",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsRRDescribe,
	}
	s.kmsRRListCmd = &cobra.Command{
		Use:   "list",
		Short: "List retired resources in a location",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsRRList,
	}
	s.kmsSTHsmCmd = &cobra.Command{
		Use:   "single-tenant-hsm",
		Short: "Manage Cloud KMS single-tenant HSM instances and proposals",
	}
	s.kmsHsmCreateCmd = &cobra.Command{
		Use:   "create INSTANCE",
		Short: "Create a single-tenant HSM instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsHsmCreate,
	}
	s.kmsHsmDescribeCmd = &cobra.Command{
		Use:   "describe INSTANCE",
		Short: "Describe a single-tenant HSM instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsHsmDescribe,
	}
	s.kmsHsmListCmd = &cobra.Command{
		Use:   "list",
		Short: "List single-tenant HSM instances in a location",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsHsmList,
	}
	s.kmsHsmProposalCmd = &cobra.Command{
		Use:   "proposal",
		Short: "Manage single-tenant HSM instance proposals",
	}
	s.kmsHsmProposalCreateCmd = &cobra.Command{
		Use:   "create PROPOSAL",
		Short: "Create a proposal for a single-tenant HSM instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsHsmProposalCreate,
	}
	s.kmsHsmProposalDescribeCmd = &cobra.Command{
		Use:   "describe PROPOSAL",
		Short: "Describe a single-tenant HSM instance proposal",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runKmsHsmProposalDescribe,
	}
	s.kmsHsmProposalListCmd = &cobra.Command{
		Use:   "list",
		Short: "List proposals for a single-tenant HSM instance",
		Args:  cobra.NoArgs,
		RunE:  s.runKmsHsmProposalList,
	}
	s.initKms()
	s.initKmsAutokeyConfig()
	s.initKmsEkmConfig()
	s.initKmsEkmConnections()
	s.initKmsImportJobs()
	s.initKmsInventory()
	s.initKmsKeyHandles()
	s.initKmsKeyrings()
	s.initKmsKeys()
	s.initKmsKeys9()
	s.initKmsLocations()
	s.initKmsOperations()
	s.initKmsRetiredResources()
	s.initKmsSingleTenantHsm()
	s.initKmsSingleTenantHsm14()
	return s
}
