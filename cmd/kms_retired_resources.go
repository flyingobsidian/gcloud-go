package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms retired-resources (#1110) ---

var kmsRetiredCmd = &cobra.Command{
	Use:   "retired-resources",
	Short: "Manage Cloud KMS retired resources",
}

var (
	flagKmsRRLocation string
	flagKmsRRFormat   string
	flagKmsRRPageSize int64
)

var kmsRRDescribeCmd = &cobra.Command{
	Use:   "describe RESOURCE",
	Short: "Describe a retired resource",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsRRDescribe,
}

var kmsRRListCmd = &cobra.Command{
	Use:   "list",
	Short: "List retired resources in a location",
	Args:  cobra.NoArgs,
	RunE:  runKmsRRList,
}

func init() {
	for _, c := range []*cobra.Command{kmsRRDescribeCmd, kmsRRListCmd} {
		c.Flags().StringVar(&flagKmsRRLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&flagKmsRRFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	kmsRRListCmd.Flags().Int64Var(&flagKmsRRPageSize, "page-size", 0, "Page size")

	kmsRetiredCmd.AddCommand(kmsRRDescribeCmd, kmsRRListCmd)
	kmsCmd.AddCommand(kmsRetiredCmd)
}

func kmsRRName(project, location, raw string) string {
	parent := kmsLocationParent(project, location) + "/retiredResources"
	return kmsFullName(parent, raw)
}

func runKmsRRDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsRRName(project, flagKmsRRLocation, args[0])
	out, err := svc.Projects.Locations.RetiredResources.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing retired resource: %w", err)
	}
	return emitFormatted(out, flagKmsRRFormat)
}

func runKmsRRList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := kmsLocationParent(project, flagKmsRRLocation)
	var all []*cloudkms.RetiredResource
	token := ""
	for {
		call := svc.Projects.Locations.RetiredResources.List(parent).Context(ctx)
		if flagKmsRRPageSize > 0 {
			call = call.PageSize(flagKmsRRPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing retired resources: %w", err)
		}
		all = append(all, resp.RetiredResources...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagKmsRRFormat)
}
