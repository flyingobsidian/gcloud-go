package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable materialized-views (#1487) ---

func (s *bigtableCmds) initBigtableMaterializedViews() {
	all := []*cobra.Command{s.bigtableMVCreateCmd, s.bigtableMVDeleteCmd, s.bigtableMVDescribeCmd, s.bigtableMVListCmd, s.bigtableMVUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBTMVInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&s.flagBTMVFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.bigtableMVCreateCmd, s.bigtableMVUpdateCmd} {
		c.Flags().StringVar(&s.flagBTMVConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the MaterializedView body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.bigtableMVUpdateCmd.Flags().StringVar(&s.flagBTMVUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.bigtableMVCreateCmd.Flags().BoolVar(&s.flagBTMVIgnoreWarn, "ignore-warnings", false,
		"If true, ignore safety checks when creating the materialized view.")
	s.bigtableMVListCmd.Flags().Int64Var(&s.flagBTMVPageSize, "page-size", 0, "Maximum results per page")

	s.bigtableMVCmd.AddCommand(all...)
	s.bigtableCmd.AddCommand(s.bigtableMVCmd)
}

func (s *bigtableCmds) btMVParent() (string, error) {
	return s.bigtableInstanceParent(s.flagBTMVInstance)
}

func (s *bigtableCmds) btMVName(id string) (string, error) {
	parent, err := s.btMVParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/materializedViews/%s", parent, id), nil
}

func (s *bigtableCmds) runBTMVCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.btMVParent()
	if err != nil {
		return err
	}
	body := &bigtableadmin.MaterializedView{}
	if err := loadYAMLOrJSONInto(s.flagBTMVConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.MaterializedViews.Create(parent, body).MaterializedViewId(args[0]).Context(ctx)
	if s.flagBTMVIgnoreWarn {
		call = call.IgnoreWarnings(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating materialized view: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for materialized view [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBTMVFormat)
}

func (s *bigtableCmds) runBTMVDelete(cmd *cobra.Command, args []string) error {
	name, err := s.btMVName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.MaterializedViews.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting materialized view: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted materialized view [%s].\n", args[0])
	return nil
}

func (s *bigtableCmds) runBTMVDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.btMVName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.MaterializedViews.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing materialized view: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBTMVFormat)
}

func (s *bigtableCmds) runBTMVList(cmd *cobra.Command, args []string) error {
	parent, err := s.btMVParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.MaterializedView
	pageToken := ""
	for {
		call := svc.Projects.Instances.MaterializedViews.List(parent).Context(ctx)
		if s.flagBTMVPageSize > 0 {
			call = call.PageSize(s.flagBTMVPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing materialized views: %w", err)
		}
		all = append(all, resp.MaterializedViews...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagBTMVFormat)
}

func (s *bigtableCmds) runBTMVUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.btMVName(args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.MaterializedView{}
	if err := loadYAMLOrJSONInto(s.flagBTMVConfigFile, body); err != nil {
		return err
	}
	mask := s.flagBTMVUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.MaterializedViews.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating materialized view: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for materialized view [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBTMVFormat)
}
