package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable logical-views (#1307) ---

func (s *bigtableCmds) initBigtableLogicalViews() {
	all := []*cobra.Command{
		s.bigtableLvCreateCmd, s.bigtableLvDeleteCmd, s.bigtableLvDescribeCmd,
		s.bigtableLvListCmd, s.bigtableLvUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBtLvInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&s.flagBtLvFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.bigtableLvCreateCmd, s.bigtableLvUpdateCmd} {
		c.Flags().StringVar(&s.flagBtLvConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the LogicalView body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.bigtableLvUpdateCmd.Flags().StringVar(&s.flagBtLvUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.bigtableLvListCmd.Flags().Int64Var(&s.flagBtLvPageSize, "page-size", 0, "Maximum results per page")

	s.bigtableLogicalViewsCmd.AddCommand(all...)
	s.bigtableCmd.AddCommand(s.bigtableLogicalViewsCmd)
}

func (s *bigtableCmds) btLvName(id string) (string, error) {
	parent, err := s.btInstanceName(s.flagBtLvInstance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/logicalViews/%s", parent, id), nil
}

func (s *bigtableCmds) runBtLvCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.btInstanceName(s.flagBtLvInstance)
	if err != nil {
		return err
	}
	body := &bigtableadmin.LogicalView{}
	if err := loadYAMLOrJSONInto(s.flagBtLvConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.LogicalViews.Create(parent, body).LogicalViewId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating logical view: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for logical view [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBtLvFormat)
}

func (s *bigtableCmds) runBtLvDelete(cmd *cobra.Command, args []string) error {
	name, err := s.btLvName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.LogicalViews.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting logical view: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted logical view [%s].\n", args[0])
	return nil
}

func (s *bigtableCmds) runBtLvDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.btLvName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.LogicalViews.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing logical view: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBtLvFormat)
}

func (s *bigtableCmds) runBtLvList(cmd *cobra.Command, args []string) error {
	parent, err := s.btInstanceName(s.flagBtLvInstance)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.LogicalView
	pageToken := ""
	for {
		call := svc.Projects.Instances.LogicalViews.List(parent).Context(ctx)
		if s.flagBtLvPageSize > 0 {
			call = call.PageSize(s.flagBtLvPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing logical views: %w", err)
		}
		all = append(all, resp.LogicalViews...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagBtLvFormat)
}

func (s *bigtableCmds) runBtLvUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.btLvName(args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.LogicalView{}
	if err := loadYAMLOrJSONInto(s.flagBtLvConfigFile, body); err != nil {
		return err
	}
	mask := s.flagBtLvUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.LogicalViews.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating logical view: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for logical view [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBtLvFormat)
}
