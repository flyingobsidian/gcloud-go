package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware operations (#1122) ---

func (s *vmwareCmds) initVmwareOperations() {
	all := []*cobra.Command{s.vmwareOpDeleteCmd, s.vmwareOpDescribeCmd, s.vmwareOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagVmwareOpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagVmwareOpFormat, "format", "", "Output format")
	}
	s.vmwareOpListCmd.Flags().StringVar(&s.flagVmwareOpFilter, "filter", "", "Server-side filter expression")
	s.vmwareOpListCmd.Flags().Int64Var(&s.flagVmwareOpPageSize, "page-size", 0, "Maximum results per page")

	s.vmwareOperationsCmd.AddCommand(all...)
	s.vmwareCmd.AddCommand(s.vmwareOperationsCmd)
}

func (s *vmwareCmds) vmwareOpName(id string) (string, error) {
	return s.vmwareResource(s.flagVmwareOpLocation, "operations", id)
}

func (s *vmwareCmds) runVmwareOpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (s *vmwareCmds) runVmwareOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagVmwareOpFormat)
}

func (s *vmwareCmds) runVmwareOpList(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwareOpLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	var all []*vmwareengine.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if s.flagVmwareOpFilter != "" {
			call = call.Filter(s.flagVmwareOpFilter)
		}
		if s.flagVmwareOpPageSize > 0 {
			call = call.PageSize(s.flagVmwareOpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagVmwareOpFormat)
}
