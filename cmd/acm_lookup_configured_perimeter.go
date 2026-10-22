package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager lookup-configured-perimeter (#1918) ---

var acmLookupConfiguredPerimeterCmd = &cobra.Command{
	Use:   "lookup-configured-perimeter",
	Short: "Look up the configured VPC Service Controls service perimeter for a resource",
	Long: "Looks up the configured VPC Service Controls service perimeter (both " +
		"enforced and dry-run) that applies to the specified project or folder " +
		"based on resource hierarchy ancestry.",
	Args: cobra.NoArgs,
	RunE: runACMLookupConfiguredPerimeter,
}

var (
	flagACMLCPResource string
	flagACMLCPFormat   string
)

func init() {
	acmLookupConfiguredPerimeterCmd.Flags().StringVar(&flagACMLCPResource, "resource", "",
		"The resource to look up effective service perimeters for. "+
			"Format: `projects/{project_number}` or `folders/{folder_number}`.")
	_ = acmLookupConfiguredPerimeterCmd.MarkFlagRequired("resource")
	acmLookupConfiguredPerimeterCmd.Flags().StringVar(&flagACMLCPFormat, "format", "", "Output format")
	accessContextManagerCmd.AddCommand(acmLookupConfiguredPerimeterCmd)
}

// validateACMLookupResource mirrors gcloud-python's check that --resource is
// a project or folder.
func validateACMLookupResource(resource string) error {
	if strings.HasPrefix(resource, "projects/") || strings.HasPrefix(resource, "folders/") {
		return nil
	}
	return fmt.Errorf("Invalid value for [--resource]: Resource must start with \"projects/\" or \"folders/\". Got: %s", resource)
}

// printACMConfiguredPerimeter renders gcloud-python's vertical key-value
// display, used when --format is not given.
func printACMConfiguredPerimeter(w io.Writer, r *accesscontextmanager.LookupConfiguredServicePerimeterResponse) {
	fmt.Fprintf(w, "SERVICE_PERIMETER:           %s\n", r.ServicePerimeter)
	fmt.Fprintf(w, "RESTRICTED_RESOURCE:         %s\n", r.RestrictedResource)
	fmt.Fprintf(w, "SERVICE_PERIMETER_DRY_RUN:   %s\n", r.ServicePerimeterDryRun)
	fmt.Fprintf(w, "RESTRICTED_RESOURCE_DRY_RUN: %s\n", r.RestrictedResourceDryRun)
}

func runACMLookupConfiguredPerimeter(cmd *cobra.Command, args []string) error {
	if err := validateACMLookupResource(flagACMLCPResource); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var resp *accesscontextmanager.LookupConfiguredServicePerimeterResponse
	if strings.HasPrefix(flagACMLCPResource, "projects/") {
		resp, err = svc.Projects.LookupConfiguredServicePerimeter(flagACMLCPResource).Context(ctx).Do()
	} else {
		resp, err = svc.Folders.LookupConfiguredServicePerimeter(flagACMLCPResource).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("looking up configured service perimeter: %w", err)
	}
	if cmd.Flags().Changed("format") {
		return emitFormatted(resp, flagACMLCPFormat)
	}
	printACMConfiguredPerimeter(os.Stdout, resp)
	return nil
}
