package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable app-profiles (#1301) ---

func (s *bigtableCmds) initBigtableAppProfiles() {
	all := []*cobra.Command{
		s.bigtableAppCreateCmd, s.bigtableAppDeleteCmd, s.bigtableAppDescribeCmd,
		s.bigtableAppListCmd, s.bigtableAppUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBtAppInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&s.flagBtAppFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.bigtableAppCreateCmd, s.bigtableAppUpdateCmd} {
		c.Flags().StringVar(&s.flagBtAppConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the AppProfile body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.bigtableAppUpdateCmd.Flags().StringVar(&s.flagBtAppUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.bigtableAppListCmd.Flags().Int64Var(&s.flagBtAppPageSize, "page-size", 0, "Maximum results per page")

	s.bigtableAppProfilesCmd.AddCommand(all...)
	s.bigtableCmd.AddCommand(s.bigtableAppProfilesCmd)
}

func (s *bigtableCmds) btAppName(id string) (string, error) {
	parent, err := s.btInstanceName(s.flagBtAppInstance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/appProfiles/%s", parent, id), nil
}

func (s *bigtableCmds) runBtAppCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.btInstanceName(s.flagBtAppInstance)
	if err != nil {
		return err
	}
	body := &bigtableadmin.AppProfile{}
	if err := loadYAMLOrJSONInto(s.flagBtAppConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.AppProfiles.Create(parent, body).AppProfileId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating app profile: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created app profile [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagBtAppFormat)
}

func (s *bigtableCmds) runBtAppDelete(cmd *cobra.Command, args []string) error {
	name, err := s.btAppName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.AppProfiles.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting app profile: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted app profile [%s].\n", args[0])
	return nil
}

func (s *bigtableCmds) runBtAppDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.btAppName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.AppProfiles.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing app profile: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBtAppFormat)
}

func (s *bigtableCmds) runBtAppList(cmd *cobra.Command, args []string) error {
	parent, err := s.btInstanceName(s.flagBtAppInstance)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.AppProfile
	pageToken := ""
	for {
		call := svc.Projects.Instances.AppProfiles.List(parent).Context(ctx)
		if s.flagBtAppPageSize > 0 {
			call = call.PageSize(s.flagBtAppPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing app profiles: %w", err)
		}
		all = append(all, resp.AppProfiles...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagBtAppFormat)
}

func (s *bigtableCmds) runBtAppUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.btAppName(args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.AppProfile{}
	if err := loadYAMLOrJSONInto(s.flagBtAppConfigFile, body); err != nil {
		return err
	}
	mask := s.flagBtAppUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.AppProfiles.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating app profile: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for app profile [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBtAppFormat)
}
