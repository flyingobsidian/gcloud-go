package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp kms-configs (#1201) ---

var netappKMSCmd = &cobra.Command{Use: "kms-configs", Short: "Manage NetApp KMS configs"}

var (
	flagNetAppKMSLocation   string
	flagNetAppKMSConfigFile string
	flagNetAppKMSUpdateMask string
	flagNetAppKMSFormat     string
	flagNetAppKMSFilter     string
	flagNetAppKMSPageSize   int64
)

var (
	netappKMSCreateCmd = &cobra.Command{
		Use: "create KMS_CONFIG", Short: "Create a KMS config",
		Args: cobra.ExactArgs(1), RunE: runNetAppKMSCreate,
	}
	netappKMSDeleteCmd = &cobra.Command{
		Use: "delete KMS_CONFIG", Short: "Delete a KMS config",
		Args: cobra.ExactArgs(1), RunE: runNetAppKMSDelete,
	}
	netappKMSDescribeCmd = &cobra.Command{
		Use: "describe KMS_CONFIG", Short: "Describe a KMS config",
		Args: cobra.ExactArgs(1), RunE: runNetAppKMSDescribe,
	}
	netappKMSListCmd = &cobra.Command{
		Use: "list", Short: "List KMS configs",
		Args: cobra.NoArgs, RunE: runNetAppKMSList,
	}
	netappKMSUpdateCmd = &cobra.Command{
		Use: "update KMS_CONFIG", Short: "Update a KMS config",
		Args: cobra.ExactArgs(1), RunE: runNetAppKMSUpdate,
	}
	netappKMSEncryptCmd = &cobra.Command{
		Use: "encrypt KMS_CONFIG", Short: "Encrypt all existing volumes and storage pools with the KMS config",
		Args: cobra.ExactArgs(1), RunE: runNetAppKMSEncrypt,
	}
	netappKMSVerifyCmd = &cobra.Command{
		Use: "verify KMS_CONFIG", Short: "Verify that the KMS config is reachable",
		Args: cobra.ExactArgs(1), RunE: runNetAppKMSVerify,
	}
)

func init() {
	all := []*cobra.Command{
		netappKMSCreateCmd, netappKMSDeleteCmd, netappKMSDescribeCmd,
		netappKMSListCmd, netappKMSUpdateCmd, netappKMSEncryptCmd, netappKMSVerifyCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNetAppKMSLocation, "location", "", "Location for the KMS config (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNetAppKMSFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{netappKMSCreateCmd, netappKMSUpdateCmd} {
		c.Flags().StringVar(&flagNetAppKMSConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the KmsConfig body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	netappKMSUpdateCmd.Flags().StringVar(&flagNetAppKMSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	netappKMSListCmd.Flags().StringVar(&flagNetAppKMSFilter, "filter", "", "Server-side filter expression")
	netappKMSListCmd.Flags().Int64Var(&flagNetAppKMSPageSize, "page-size", 0, "Maximum number of results per page")

	netappKMSCmd.AddCommand(all...)
	netappCmd.AddCommand(netappKMSCmd)
}

func netappKMSParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, flagNetAppKMSLocation), nil
}

func netappKMSName(id string) (string, error) {
	parent, err := netappKMSParent()
	if err != nil {
		return "", err
	}
	return netappChild("kmsConfigs", id, parent), nil
}

func runNetAppKMSCreate(cmd *cobra.Command, args []string) error {
	parent, err := netappKMSParent()
	if err != nil {
		return err
	}
	body := &netapp.KmsConfig{}
	if err := loadYAMLOrJSONInto(flagNetAppKMSConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.KmsConfigs.Create(parent, body).KmsConfigId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating KMS config: %w", err)
	}
	fmt.Printf("Create request issued for KMS config [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppKMSFormat)
}

func runNetAppKMSDelete(cmd *cobra.Command, args []string) error {
	name, err := netappKMSName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.KmsConfigs.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting KMS config: %w", err)
	}
	fmt.Printf("Delete request issued for KMS config [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppKMSFormat)
}

func runNetAppKMSDescribe(cmd *cobra.Command, args []string) error {
	name, err := netappKMSName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.KmsConfigs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing KMS config: %w", err)
	}
	return emitFormatted(got, flagNetAppKMSFormat)
}

func runNetAppKMSList(cmd *cobra.Command, args []string) error {
	parent, err := netappKMSParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*netapp.KmsConfig
	pageToken := ""
	for {
		call := svc.Projects.Locations.KmsConfigs.List(parent).Context(ctx)
		if flagNetAppKMSFilter != "" {
			call = call.Filter(flagNetAppKMSFilter)
		}
		if flagNetAppKMSPageSize > 0 {
			call = call.PageSize(flagNetAppKMSPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing KMS configs: %w", err)
		}
		all = append(all, resp.KmsConfigs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNetAppKMSFormat)
}

func runNetAppKMSUpdate(cmd *cobra.Command, args []string) error {
	name, err := netappKMSName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.KmsConfig{}
	if err := loadYAMLOrJSONInto(flagNetAppKMSConfigFile, body); err != nil {
		return err
	}
	mask := flagNetAppKMSUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.KmsConfigs.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating KMS config: %w", err)
	}
	fmt.Printf("Update request issued for KMS config [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppKMSFormat)
}

func runNetAppKMSEncrypt(cmd *cobra.Command, args []string) error {
	name, err := netappKMSName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.KmsConfigs.Encrypt(name, &netapp.EncryptVolumesRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("encrypting with KMS config: %w", err)
	}
	fmt.Printf("Encrypt request issued for KMS config [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppKMSFormat)
}

func runNetAppKMSVerify(cmd *cobra.Command, args []string) error {
	name, err := netappKMSName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.KmsConfigs.Verify(name, &netapp.VerifyKmsConfigRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("verifying KMS config: %w", err)
	}
	return emitFormatted(resp, flagNetAppKMSFormat)
}
