package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware node-types (#1121) ---

func (s *vmwareCmds) initVmwareNodeTypes() {
	all := []*cobra.Command{s.vmwareNtDescribeCmd, s.vmwareNtListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagVmwareNtLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagVmwareNtFormat, "format", "", "Output format")
	}
	s.vmwareNtListCmd.Flags().Int64Var(&s.flagVmwareNtPageSize, "page-size", 0, "Maximum results per page")

	s.vmwareNodeTypesCmd.AddCommand(all...)
	s.vmwareCmd.AddCommand(s.vmwareNodeTypesCmd)
}

func (s *vmwareCmds) vmwareNtName(id string) (string, error) {
	return s.vmwareResource(s.flagVmwareNtLocation, "nodeTypes", id)
}

func (s *vmwareCmds) runVmwareNtDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareNtName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NodeTypes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing node type: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagVmwareNtFormat)
}

func (s *vmwareCmds) runVmwareNtList(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwareNtLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	var all []*vmwareengine.NodeType
	pageToken := ""
	for {
		call := svc.Projects.Locations.NodeTypes.List(parent).Context(ctx)
		if s.flagVmwareNtPageSize > 0 {
			call = call.PageSize(s.flagVmwareNtPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing node types: %w", err)
		}
		all = append(all, resp.NodeTypes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagVmwareNtFormat)
}
