package cli

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp operations (#1203) ---

func (s *netappCmds) initNetappOperations() {
	all := []*cobra.Command{s.netappOpCancelCmd, s.netappOpDeleteCmd, s.netappOpDescribeCmd, s.netappOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNetAppOpLocation, "location", "", "Location for the operation (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNetAppOpFormat, "format", "", "Output format")
	}
	s.netappOpListCmd.Flags().StringVar(&s.flagNetAppOpFilter, "filter", "", "Server-side filter expression")
	s.netappOpListCmd.Flags().Int64Var(&s.flagNetAppOpPageSize, "page-size", 0, "Maximum number of results per page")

	s.netappOpCmd.AddCommand(all...)
	s.netappCmd.AddCommand(s.netappOpCmd)
}

func (s *netappCmds) netappOpParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, s.flagNetAppOpLocation), nil
}

func (s *netappCmds) netappOpName(id string) (string, error) {
	parent, err := s.netappOpParent()
	if err != nil {
		return "", err
	}
	return netappChild("operations", id, parent), nil
}

func (s *netappCmds) runNetAppOpCancel(cmd *cobra.Command, args []string) error {
	name, err := s.netappOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &netapp.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancel request issued for operation %s.\n", args[0])
	return nil
}

func (s *netappCmds) runNetAppOpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.netappOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation %s.\n", args[0])
	return nil
}

func (s *netappCmds) runNetAppOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.netappOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppOpFormat)
}

func (s *netappCmds) runNetAppOpList(cmd *cobra.Command, args []string) error {
	parent, err := s.netappOpParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	var all []*netapp.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if s.flagNetAppOpFilter != "" {
			call = call.Filter(s.flagNetAppOpFilter)
		}
		if s.flagNetAppOpPageSize > 0 {
			call = call.PageSize(s.flagNetAppOpPageSize)
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
	if s.flagNetAppOpFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagNetAppOpFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}
