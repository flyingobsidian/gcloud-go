package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable authorized-views (#1302) ---

func (s *bigtableCmds) initBigtableAuthorizedViews() {
	all := []*cobra.Command{
		s.bigtableAvCreateCmd, s.bigtableAvDeleteCmd, s.bigtableAvDescribeCmd,
		s.bigtableAvListCmd, s.bigtableAvUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBtAvInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&s.flagBtAvTable, "table", "", "Bigtable table ID (required)")
		_ = c.MarkFlagRequired("table")
		c.Flags().StringVar(&s.flagBtAvFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.bigtableAvCreateCmd, s.bigtableAvUpdateCmd} {
		c.Flags().StringVar(&s.flagBtAvConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the AuthorizedView body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.bigtableAvUpdateCmd.Flags().StringVar(&s.flagBtAvUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.bigtableAvListCmd.Flags().Int64Var(&s.flagBtAvPageSize, "page-size", 0, "Maximum results per page")

	s.bigtableAuthorizedViewsCmd.AddCommand(all...)
	s.bigtableCmd.AddCommand(s.bigtableAuthorizedViewsCmd)
}

func (s *bigtableCmds) btAvName(id string) (string, error) {
	parent, err := s.btTableName(s.flagBtAvInstance, s.flagBtAvTable)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/authorizedViews/%s", parent, id), nil
}

func (s *bigtableCmds) runBtAvCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.btTableName(s.flagBtAvInstance, s.flagBtAvTable)
	if err != nil {
		return err
	}
	body := &bigtableadmin.AuthorizedView{}
	if err := loadYAMLOrJSONInto(s.flagBtAvConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Tables.AuthorizedViews.Create(parent, body).AuthorizedViewId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating authorized view: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for authorized view [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBtAvFormat)
}

func (s *bigtableCmds) runBtAvDelete(cmd *cobra.Command, args []string) error {
	name, err := s.btAvName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Tables.AuthorizedViews.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting authorized view: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted authorized view [%s].\n", args[0])
	return nil
}

func (s *bigtableCmds) runBtAvDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.btAvName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Tables.AuthorizedViews.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing authorized view: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBtAvFormat)
}

func (s *bigtableCmds) runBtAvList(cmd *cobra.Command, args []string) error {
	parent, err := s.btTableName(s.flagBtAvInstance, s.flagBtAvTable)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.AuthorizedView
	pageToken := ""
	for {
		call := svc.Projects.Instances.Tables.AuthorizedViews.List(parent).Context(ctx)
		if s.flagBtAvPageSize > 0 {
			call = call.PageSize(s.flagBtAvPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing authorized views: %w", err)
		}
		all = append(all, resp.AuthorizedViews...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagBtAvFormat)
}

func (s *bigtableCmds) runBtAvUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.btAvName(args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.AuthorizedView{}
	if err := loadYAMLOrJSONInto(s.flagBtAvConfigFile, body); err != nil {
		return err
	}
	mask := s.flagBtAvUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.Tables.AuthorizedViews.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating authorized view: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for authorized view [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBtAvFormat)
}
