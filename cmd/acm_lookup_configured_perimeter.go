package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager lookup-configured-perimeter (#1918) ---

func (s *accessContextManagerCmds) initAcmLookupConfiguredPerimeter() {
	s.acmLookupConfiguredPerimeterCmd.Flags().StringVar(&s.flagACMLCPResource, "resource", "",
		"The resource to look up effective service perimeters for. "+
			"Format: `projects/{project_number}` or `folders/{folder_number}`.")
	_ = s.acmLookupConfiguredPerimeterCmd.MarkFlagRequired("resource")
	s.acmLookupConfiguredPerimeterCmd.Flags().StringVar(&s.flagACMLCPFormat, "format", "", "Output format")
	s.accessContextManagerCmd.AddCommand(s.acmLookupConfiguredPerimeterCmd)
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

func (s *accessContextManagerCmds) runACMLookupConfiguredPerimeter(cmd *cobra.Command, args []string) error {
	if err := validateACMLookupResource(s.flagACMLCPResource); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	var resp *accesscontextmanager.LookupConfiguredServicePerimeterResponse
	if strings.HasPrefix(s.flagACMLCPResource, "projects/") {
		resp, err = svc.Projects.LookupConfiguredServicePerimeter(s.flagACMLCPResource).Context(ctx).Do()
	} else {
		resp, err = svc.Folders.LookupConfiguredServicePerimeter(s.flagACMLCPResource).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("looking up configured service perimeter: %w", err)
	}
	if cmd.Flags().Changed("format") {
		return emitFormattedTo(s.inv.stdout, resp, s.flagACMLCPFormat)
	}
	printACMConfiguredPerimeter(s.inv.stdout, resp)
	return nil
}
