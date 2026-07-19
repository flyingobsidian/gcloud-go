package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms single-tenant-hsm (#1111) ---

var kmsSTHsmCmd = &cobra.Command{
	Use:   "single-tenant-hsm",
	Short: "Manage Cloud KMS single-tenant HSM instances and proposals",
}

var (
	flagKmsHsmLocation   string
	flagKmsHsmFormat     string
	flagKmsHsmFilter     string
	flagKmsHsmPageSize   int64
	flagKmsHsmConfigFile string
	flagKmsHsmInstance   string
)

var kmsHsmCreateCmd = &cobra.Command{
	Use:   "create INSTANCE",
	Short: "Create a single-tenant HSM instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsHsmCreate,
}

var kmsHsmDescribeCmd = &cobra.Command{
	Use:   "describe INSTANCE",
	Short: "Describe a single-tenant HSM instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsHsmDescribe,
}

var kmsHsmListCmd = &cobra.Command{
	Use:   "list",
	Short: "List single-tenant HSM instances in a location",
	Args:  cobra.NoArgs,
	RunE:  runKmsHsmList,
}

func init() {
	for _, c := range []*cobra.Command{kmsHsmCreateCmd, kmsHsmDescribeCmd, kmsHsmListCmd} {
		c.Flags().StringVar(&flagKmsHsmLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&flagKmsHsmFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	kmsHsmCreateCmd.Flags().StringVar(&flagKmsHsmConfigFile, "config-file", "", "YAML/JSON body for the SingleTenantHsmInstance (required)")
	_ = kmsHsmCreateCmd.MarkFlagRequired("config-file")

	kmsHsmListCmd.Flags().StringVar(&flagKmsHsmFilter, "filter", "", "Filter expression")
	kmsHsmListCmd.Flags().Int64Var(&flagKmsHsmPageSize, "page-size", 0, "Page size")

	kmsSTHsmCmd.AddCommand(kmsHsmCreateCmd, kmsHsmDescribeCmd, kmsHsmListCmd)
	kmsSTHsmCmd.AddCommand(kmsHsmProposalCmd)
	kmsCmd.AddCommand(kmsSTHsmCmd)
}

func kmsHsmInstanceName(project, location, raw string) string {
	parent := kmsLocationParent(project, location) + "/singleTenantHsmInstances"
	return kmsFullName(parent, raw)
}

func runKmsHsmCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &cloudkms.SingleTenantHsmInstance{}
	if err := loadYAMLOrJSONInto(flagKmsHsmConfigFile, body); err != nil {
		return err
	}
	parent := kmsLocationParent(project, flagKmsHsmLocation)
	out, err := svc.Projects.Locations.SingleTenantHsmInstances.Create(parent, body).
		SingleTenantHsmInstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating single-tenant HSM instance: %w", err)
	}
	return emitFormatted(out, flagKmsHsmFormat)
}

func runKmsHsmDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsHsmInstanceName(project, flagKmsHsmLocation, args[0])
	out, err := svc.Projects.Locations.SingleTenantHsmInstances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing single-tenant HSM instance: %w", err)
	}
	return emitFormatted(out, flagKmsHsmFormat)
}

func runKmsHsmList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := kmsLocationParent(project, flagKmsHsmLocation)
	var all []*cloudkms.SingleTenantHsmInstance
	token := ""
	for {
		call := svc.Projects.Locations.SingleTenantHsmInstances.List(parent).Context(ctx)
		if flagKmsHsmFilter != "" {
			call = call.Filter(flagKmsHsmFilter)
		}
		if flagKmsHsmPageSize > 0 {
			call = call.PageSize(flagKmsHsmPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing single-tenant HSM instances: %w", err)
		}
		all = append(all, resp.SingleTenantHsmInstances...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagKmsHsmFormat)
}

// --- proposal subgroup ---

var kmsHsmProposalCmd = &cobra.Command{
	Use:   "proposal",
	Short: "Manage single-tenant HSM instance proposals",
}

var kmsHsmProposalCreateCmd = &cobra.Command{
	Use:   "create PROPOSAL",
	Short: "Create a proposal for a single-tenant HSM instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsHsmProposalCreate,
}

var kmsHsmProposalDescribeCmd = &cobra.Command{
	Use:   "describe PROPOSAL",
	Short: "Describe a single-tenant HSM instance proposal",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsHsmProposalDescribe,
}

var kmsHsmProposalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List proposals for a single-tenant HSM instance",
	Args:  cobra.NoArgs,
	RunE:  runKmsHsmProposalList,
}

func init() {
	for _, c := range []*cobra.Command{
		kmsHsmProposalCreateCmd, kmsHsmProposalDescribeCmd, kmsHsmProposalListCmd,
	} {
		c.Flags().StringVar(&flagKmsHsmLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&flagKmsHsmInstance, "hsm-instance", "", "Parent single-tenant HSM instance id (required)")
		c.Flags().StringVar(&flagKmsHsmFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("hsm-instance")
	}
	kmsHsmProposalCreateCmd.Flags().StringVar(&flagKmsHsmConfigFile, "config-file", "", "YAML/JSON body for the SingleTenantHsmInstanceProposal (required)")
	_ = kmsHsmProposalCreateCmd.MarkFlagRequired("config-file")

	kmsHsmProposalListCmd.Flags().StringVar(&flagKmsHsmFilter, "filter", "", "Filter expression")
	kmsHsmProposalListCmd.Flags().Int64Var(&flagKmsHsmPageSize, "page-size", 0, "Page size")

	kmsHsmProposalCmd.AddCommand(kmsHsmProposalCreateCmd, kmsHsmProposalDescribeCmd, kmsHsmProposalListCmd)
}

func kmsHsmProposalName(project, location, instance, raw string) string {
	parent := kmsHsmInstanceName(project, location, instance) + "/proposals"
	return kmsFullName(parent, raw)
}

func runKmsHsmProposalCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &cloudkms.SingleTenantHsmInstanceProposal{}
	if err := loadYAMLOrJSONInto(flagKmsHsmConfigFile, body); err != nil {
		return err
	}
	parent := kmsHsmInstanceName(project, flagKmsHsmLocation, flagKmsHsmInstance)
	out, err := svc.Projects.Locations.SingleTenantHsmInstances.Proposals.Create(parent, body).
		SingleTenantHsmInstanceProposalId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating HSM proposal: %w", err)
	}
	return emitFormatted(out, flagKmsHsmFormat)
}

func runKmsHsmProposalDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsHsmProposalName(project, flagKmsHsmLocation, flagKmsHsmInstance, args[0])
	out, err := svc.Projects.Locations.SingleTenantHsmInstances.Proposals.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing HSM proposal: %w", err)
	}
	return emitFormatted(out, flagKmsHsmFormat)
}

func runKmsHsmProposalList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := kmsHsmInstanceName(project, flagKmsHsmLocation, flagKmsHsmInstance)
	var all []*cloudkms.SingleTenantHsmInstanceProposal
	token := ""
	for {
		call := svc.Projects.Locations.SingleTenantHsmInstances.Proposals.List(parent).Context(ctx)
		if flagKmsHsmFilter != "" {
			call = call.Filter(flagKmsHsmFilter)
		}
		if flagKmsHsmPageSize > 0 {
			call = call.PageSize(flagKmsHsmPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing HSM proposals: %w", err)
		}
		all = append(all, resp.SingleTenantHsmInstanceProposals...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagKmsHsmFormat)
}
