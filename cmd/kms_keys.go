package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms keys (#1107) ---

var kmsKeysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage Cloud KMS crypto keys",
}

var (
	flagKmsKLocation   string
	flagKmsKKeyring    string
	flagKmsKFormat     string
	flagKmsKFilter     string
	flagKmsKPageSize   int64
	flagKmsKConfigFile string
	flagKmsKMask       string
	flagKmsKAllConds   bool

	flagKmsKIamMember    string
	flagKmsKIamRole      string
	flagKmsKIamCondExpr  string
	flagKmsKIamCondTitle string
	flagKmsKIamCondDesc  string

	flagKmsKPurpose         string
	flagKmsKProtectionLevel string
	flagKmsKAlgorithm       string
	flagKmsKRotationPeriod  string
	flagKmsKNextRotation    string
	flagKmsKPrimaryVersion  string
)

var kmsKCreateCmd = &cobra.Command{
	Use:   "create KEY",
	Short: "Create a crypto key",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKCreate,
}

var kmsKDeleteCmd = &cobra.Command{
	Use:   "delete KEY",
	Short: "Destroy all versions of a crypto key (Cloud KMS keys cannot be deleted directly)",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKDelete,
}

var kmsKDescribeCmd = &cobra.Command{
	Use:   "describe KEY",
	Short: "Describe a crypto key",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKDescribe,
}

var kmsKListCmd = &cobra.Command{
	Use:   "list",
	Short: "List crypto keys in a keyring",
	Args:  cobra.NoArgs,
	RunE:  runKmsKList,
}

var kmsKUpdateCmd = &cobra.Command{
	Use:   "update KEY",
	Short: "Update a crypto key",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKUpdate,
}

var kmsKSetPrimaryCmd = &cobra.Command{
	Use:   "set-primary-version KEY",
	Short: "Set the primary version of a crypto key",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKSetPrimary,
}

var kmsKSetRotationCmd = &cobra.Command{
	Use:   "set-rotation-schedule KEY",
	Short: "Set the rotation schedule on a crypto key",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKSetRotation,
}

var kmsKRemoveRotationCmd = &cobra.Command{
	Use:   "remove-rotation-schedule KEY",
	Short: "Clear the rotation schedule on a crypto key",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKRemoveRotation,
}

var kmsKGetIamCmd = &cobra.Command{
	Use:   "get-iam-policy KEY",
	Short: "Get the IAM policy for a crypto key",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKGetIam,
}

var kmsKSetIamCmd = &cobra.Command{
	Use:   "set-iam-policy KEY POLICY_FILE",
	Short: "Set the IAM policy for a crypto key",
	Args:  cobra.ExactArgs(2),
	RunE:  runKmsKSetIam,
}

var kmsKAddIamCmd = &cobra.Command{
	Use:   "add-iam-policy-binding KEY",
	Short: "Add an IAM policy binding on a crypto key",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKAddIam,
}

var kmsKRemoveIamCmd = &cobra.Command{
	Use:   "remove-iam-policy-binding KEY",
	Short: "Remove an IAM policy binding from a crypto key",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKRemoveIam,
}

func init() {
	all := []*cobra.Command{
		kmsKCreateCmd, kmsKDeleteCmd, kmsKDescribeCmd, kmsKListCmd, kmsKUpdateCmd,
		kmsKSetPrimaryCmd, kmsKSetRotationCmd, kmsKRemoveRotationCmd,
		kmsKGetIamCmd, kmsKSetIamCmd, kmsKAddIamCmd, kmsKRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagKmsKLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&flagKmsKKeyring, "keyring", "", "Parent key ring (required)")
		c.Flags().StringVar(&flagKmsKFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("keyring")
	}

	kmsKCreateCmd.Flags().StringVar(&flagKmsKConfigFile, "config-file", "", "YAML/JSON body for the CryptoKey (optional)")
	kmsKCreateCmd.Flags().StringVar(&flagKmsKPurpose, "purpose", "ENCRYPT_DECRYPT", "CryptoKey purpose")
	kmsKCreateCmd.Flags().StringVar(&flagKmsKProtectionLevel, "protection-level", "", "ProtectionLevel for the default VersionTemplate")
	kmsKCreateCmd.Flags().StringVar(&flagKmsKAlgorithm, "default-algorithm", "", "Algorithm for the default VersionTemplate")
	kmsKCreateCmd.Flags().StringVar(&flagKmsKRotationPeriod, "rotation-period", "", "Rotation period (e.g. 30d, 7776000s)")
	kmsKCreateCmd.Flags().StringVar(&flagKmsKNextRotation, "next-rotation-time", "", "RFC 3339 timestamp for next rotation")

	kmsKListCmd.Flags().StringVar(&flagKmsKFilter, "filter", "", "Filter expression")
	kmsKListCmd.Flags().Int64Var(&flagKmsKPageSize, "page-size", 0, "Page size")

	kmsKUpdateCmd.Flags().StringVar(&flagKmsKConfigFile, "config-file", "", "YAML/JSON body for the CryptoKey update (required)")
	kmsKUpdateCmd.Flags().StringVar(&flagKmsKMask, "update-mask", "", "Fields to update; defaults to populated fields")
	_ = kmsKUpdateCmd.MarkFlagRequired("config-file")

	kmsKSetPrimaryCmd.Flags().StringVar(&flagKmsKPrimaryVersion, "version", "", "CryptoKeyVersion id to promote (required)")
	_ = kmsKSetPrimaryCmd.MarkFlagRequired("version")

	kmsKSetRotationCmd.Flags().StringVar(&flagKmsKRotationPeriod, "rotation-period", "", "Rotation period (e.g. 30d, 7776000s) (required)")
	kmsKSetRotationCmd.Flags().StringVar(&flagKmsKNextRotation, "next-rotation-time", "", "RFC 3339 timestamp for next rotation (required)")
	_ = kmsKSetRotationCmd.MarkFlagRequired("rotation-period")
	_ = kmsKSetRotationCmd.MarkFlagRequired("next-rotation-time")

	kmsIamMemberFlags(kmsKAddIamCmd, &flagKmsKIamMember, &flagKmsKIamRole,
		&flagKmsKIamCondExpr, &flagKmsKIamCondTitle, &flagKmsKIamCondDesc)
	kmsIamMemberFlags(kmsKRemoveIamCmd, &flagKmsKIamMember, &flagKmsKIamRole,
		&flagKmsKIamCondExpr, &flagKmsKIamCondTitle, &flagKmsKIamCondDesc)
	kmsKRemoveIamCmd.Flags().BoolVar(&flagKmsKAllConds, "all", false, "Match bindings for the role across all conditions")

	kmsKeysCmd.AddCommand(all...)
	kmsKeysCmd.AddCommand(kmsVersionsCmd)
	kmsCmd.AddCommand(kmsKeysCmd)
}

func kmsKeyResource(project, location, keyring, raw string) string {
	parent := kmsKeyringParent(project, location, keyring) + "/cryptoKeys"
	return kmsFullName(parent, raw)
}

func runKmsKCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &cloudkms.CryptoKey{}
	if flagKmsKConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagKmsKConfigFile, body); err != nil {
			return err
		}
	}
	if flagKmsKPurpose != "" && body.Purpose == "" {
		body.Purpose = flagKmsKPurpose
	}
	if flagKmsKProtectionLevel != "" || flagKmsKAlgorithm != "" {
		if body.VersionTemplate == nil {
			body.VersionTemplate = &cloudkms.CryptoKeyVersionTemplate{}
		}
		if flagKmsKProtectionLevel != "" {
			body.VersionTemplate.ProtectionLevel = flagKmsKProtectionLevel
		}
		if flagKmsKAlgorithm != "" {
			body.VersionTemplate.Algorithm = flagKmsKAlgorithm
		}
	}
	if flagKmsKRotationPeriod != "" {
		body.RotationPeriod = flagKmsKRotationPeriod
	}
	if flagKmsKNextRotation != "" {
		body.NextRotationTime = flagKmsKNextRotation
	}
	parent := kmsKeyringParent(project, flagKmsKLocation, flagKmsKKeyring)
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.Create(parent, body).
		CryptoKeyId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating crypto key: %w", err)
	}
	return emitFormatted(out, flagKmsKFormat)
}

// runKmsKDelete iterates through all versions of a crypto key and schedules
// each non-destroyed one for destruction. Cloud KMS does not permit deleting
// keys directly; this mirrors gcloud Python's behavior for `keys delete`.
func runKmsKDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	keyName := kmsKeyResource(project, flagKmsKLocation, flagKmsKKeyring, args[0])
	var destroyed []*cloudkms.CryptoKeyVersion
	token := ""
	for {
		call := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.List(keyName).Context(ctx)
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing versions: %w", err)
		}
		for _, v := range resp.CryptoKeyVersions {
			switch v.State {
			case "DESTROYED", "DESTROY_SCHEDULED", "PENDING_IMPORT", "PENDING_GENERATION":
				continue
			}
			out, err := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.
				Destroy(v.Name, &cloudkms.DestroyCryptoKeyVersionRequest{}).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("destroying %s: %w", v.Name, err)
			}
			destroyed = append(destroyed, out)
		}
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	fmt.Fprintf(os.Stderr, "Scheduled %d version(s) of %s for destruction.\n", len(destroyed), keyName)
	return emitFormatted(destroyed, flagKmsKFormat)
}

func runKmsKDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsKeyResource(project, flagKmsKLocation, flagKmsKKeyring, args[0])
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing crypto key: %w", err)
	}
	return emitFormatted(out, flagKmsKFormat)
}

func runKmsKList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := kmsKeyringParent(project, flagKmsKLocation, flagKmsKKeyring)
	var all []*cloudkms.CryptoKey
	token := ""
	for {
		call := svc.Projects.Locations.KeyRings.CryptoKeys.List(parent).Context(ctx)
		if flagKmsKFilter != "" {
			call = call.Filter(flagKmsKFilter)
		}
		if flagKmsKPageSize > 0 {
			call = call.PageSize(flagKmsKPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing crypto keys: %w", err)
		}
		all = append(all, resp.CryptoKeys...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagKmsKFormat)
}

func runKmsKUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &cloudkms.CryptoKey{}
	if err := loadYAMLOrJSONInto(flagKmsKConfigFile, body); err != nil {
		return err
	}
	mask := flagKmsKMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := kmsKeyResource(project, flagKmsKLocation, flagKmsKKeyring, args[0])
	call := svc.Projects.Locations.KeyRings.CryptoKeys.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating crypto key: %w", err)
	}
	return emitFormatted(out, flagKmsKFormat)
}

func runKmsKSetPrimary(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsKeyResource(project, flagKmsKLocation, flagKmsKKeyring, args[0])
	req := &cloudkms.UpdateCryptoKeyPrimaryVersionRequest{CryptoKeyVersionId: flagKmsKPrimaryVersion}
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.UpdatePrimaryVersion(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting primary version: %w", err)
	}
	return emitFormatted(out, flagKmsKFormat)
}

func runKmsKSetRotation(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &cloudkms.CryptoKey{
		RotationPeriod:   flagKmsKRotationPeriod,
		NextRotationTime: flagKmsKNextRotation,
	}
	name := kmsKeyResource(project, flagKmsKLocation, flagKmsKKeyring, args[0])
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.Patch(name, body).
		UpdateMask("rotationPeriod,nextRotationTime").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting rotation schedule: %w", err)
	}
	return emitFormatted(out, flagKmsKFormat)
}

func runKmsKRemoveRotation(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &cloudkms.CryptoKey{
		NullFields: []string{"RotationPeriod", "NextRotationTime"},
	}
	name := kmsKeyResource(project, flagKmsKLocation, flagKmsKKeyring, args[0])
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.Patch(name, body).
		UpdateMask("rotationPeriod,nextRotationTime").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing rotation schedule: %w", err)
	}
	return emitFormatted(out, flagKmsKFormat)
}

func runKmsKGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsKeyResource(project, flagKmsKLocation, flagKmsKKeyring, args[0])
	pol, err := svc.Projects.Locations.KeyRings.CryptoKeys.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(pol, flagKmsKFormat)
}

func runKmsKSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy := &cloudkms.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	name := kmsKeyResource(project, flagKmsKLocation, flagKmsKKeyring, args[0])
	req := &cloudkms.SetIamPolicyRequest{Policy: policy}
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsKFormat)
}

func runKmsKAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsKeyResource(project, flagKmsKLocation, flagKmsKKeyring, args[0])
	pol, err := svc.Projects.Locations.KeyRings.CryptoKeys.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(flagKmsKIamCondExpr, flagKmsKIamCondTitle, flagKmsKIamCondDesc)
	kmsIamAddBinding(pol, flagKmsKIamRole, flagKmsKIamMember, cond)
	if cond != nil && pol.Version < 3 {
		pol.Version = 3
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsKFormat)
}

func runKmsKRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsKeyResource(project, flagKmsKLocation, flagKmsKKeyring, args[0])
	pol, err := svc.Projects.Locations.KeyRings.CryptoKeys.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(flagKmsKIamCondExpr, flagKmsKIamCondTitle, flagKmsKIamCondDesc)
	if !kmsIamRemoveBinding(pol, flagKmsKIamRole, flagKmsKIamMember, cond, flagKmsKAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsKFormat)
}

// --- versions subgroup ---

var kmsVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Manage crypto key versions",
}

var (
	flagKmsVKey        string
	flagKmsVFormat     string
	flagKmsVFilter     string
	flagKmsVPageSize   int64
	flagKmsVConfigFile string
	flagKmsVAlgorithm  string
	flagKmsVProtection string
)

var kmsVCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new CryptoKeyVersion",
	Args:  cobra.NoArgs,
	RunE:  runKmsVCreate,
}

var kmsVDescribeCmd = &cobra.Command{
	Use:   "describe VERSION",
	Short: "Describe a CryptoKeyVersion",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsVDescribe,
}

var kmsVDestroyCmd = &cobra.Command{
	Use:   "destroy VERSION",
	Short: "Schedule a CryptoKeyVersion for destruction",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsVDestroy,
}

var kmsVDisableCmd = &cobra.Command{
	Use:   "disable VERSION",
	Short: "Disable a CryptoKeyVersion",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsVDisable,
}

var kmsVEnableCmd = &cobra.Command{
	Use:   "enable VERSION",
	Short: "Enable a CryptoKeyVersion",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsVEnable,
}

var kmsVListCmd = &cobra.Command{
	Use:   "list",
	Short: "List CryptoKeyVersions for a CryptoKey",
	Args:  cobra.NoArgs,
	RunE:  runKmsVList,
}

var kmsVRestoreCmd = &cobra.Command{
	Use:   "restore VERSION",
	Short: "Restore a CryptoKeyVersion pending destruction",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsVRestore,
}

func init() {
	all := []*cobra.Command{
		kmsVCreateCmd, kmsVDescribeCmd, kmsVDestroyCmd, kmsVDisableCmd,
		kmsVEnableCmd, kmsVListCmd, kmsVRestoreCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagKmsKLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&flagKmsKKeyring, "keyring", "", "Parent key ring (required)")
		c.Flags().StringVar(&flagKmsVKey, "key", "", "Parent CryptoKey id (required)")
		c.Flags().StringVar(&flagKmsVFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("keyring")
		_ = c.MarkFlagRequired("key")
	}
	kmsVCreateCmd.Flags().StringVar(&flagKmsVConfigFile, "config-file", "", "YAML/JSON body for the CryptoKeyVersion (optional)")
	kmsVCreateCmd.Flags().StringVar(&flagKmsVAlgorithm, "algorithm", "", "CryptoKeyVersion algorithm")
	kmsVCreateCmd.Flags().StringVar(&flagKmsVProtection, "protection-level", "", "CryptoKeyVersion protection level")

	kmsVListCmd.Flags().StringVar(&flagKmsVFilter, "filter", "", "Filter expression")
	kmsVListCmd.Flags().Int64Var(&flagKmsVPageSize, "page-size", 0, "Page size")

	kmsVersionsCmd.AddCommand(all...)
}

func kmsVersionName(project, location, keyring, key, raw string) string {
	parent := kmsKeyResource(project, location, keyring, key) + "/cryptoKeyVersions"
	return kmsFullName(parent, raw)
}

func runKmsVCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &cloudkms.CryptoKeyVersion{}
	if flagKmsVConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagKmsVConfigFile, body); err != nil {
			return err
		}
	}
	if flagKmsVAlgorithm != "" {
		body.Algorithm = flagKmsVAlgorithm
	}
	if flagKmsVProtection != "" {
		body.ProtectionLevel = flagKmsVProtection
	}
	parent := kmsKeyResource(project, flagKmsKLocation, flagKmsKKeyring, flagKmsVKey)
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating crypto key version: %w", err)
	}
	return emitFormatted(out, flagKmsVFormat)
}

func runKmsVDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsVersionName(project, flagKmsKLocation, flagKmsKKeyring, flagKmsVKey, args[0])
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing crypto key version: %w", err)
	}
	return emitFormatted(out, flagKmsVFormat)
}

func runKmsVDestroy(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsVersionName(project, flagKmsKLocation, flagKmsKKeyring, flagKmsVKey, args[0])
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.
		Destroy(name, &cloudkms.DestroyCryptoKeyVersionRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("destroying crypto key version: %w", err)
	}
	return emitFormatted(out, flagKmsVFormat)
}

func runKmsVDisable(cmd *cobra.Command, args []string) error {
	return kmsVSetState(args[0], "DISABLED")
}

func runKmsVEnable(cmd *cobra.Command, args []string) error {
	return kmsVSetState(args[0], "ENABLED")
}

func kmsVSetState(id, state string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsVersionName(project, flagKmsKLocation, flagKmsKKeyring, flagKmsVKey, id)
	body := &cloudkms.CryptoKeyVersion{State: state}
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.
		Patch(name, body).UpdateMask("state").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating version state: %w", err)
	}
	return emitFormatted(out, flagKmsVFormat)
}

func runKmsVList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := kmsKeyResource(project, flagKmsKLocation, flagKmsKKeyring, flagKmsVKey)
	var all []*cloudkms.CryptoKeyVersion
	token := ""
	for {
		call := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.List(parent).Context(ctx)
		if flagKmsVFilter != "" {
			call = call.Filter(flagKmsVFilter)
		}
		if flagKmsVPageSize > 0 {
			call = call.PageSize(flagKmsVPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing crypto key versions: %w", err)
		}
		all = append(all, resp.CryptoKeyVersions...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagKmsVFormat)
}

func runKmsVRestore(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsVersionName(project, flagKmsKLocation, flagKmsKKeyring, flagKmsVKey, args[0])
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.
		Restore(name, &cloudkms.RestoreCryptoKeyVersionRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring crypto key version: %w", err)
	}
	return emitFormatted(out, flagKmsVFormat)
}
