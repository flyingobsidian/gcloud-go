package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable app-profiles (#1301) ---

var bigtableAppProfilesCmd = &cobra.Command{Use: "app-profiles", Short: "Manage Bigtable app profiles"}

var (
	flagBtAppInstance   string
	flagBtAppFormat     string
	flagBtAppConfigFile string
	flagBtAppUpdateMask string
	flagBtAppPageSize   int64
)

var (
	bigtableAppCreateCmd = &cobra.Command{
		Use: "create APP_PROFILE", Short: "Create a Bigtable app profile",
		Args: cobra.ExactArgs(1), RunE: runBtAppCreate,
	}
	bigtableAppDeleteCmd = &cobra.Command{
		Use: "delete APP_PROFILE", Short: "Delete a Bigtable app profile",
		Args: cobra.ExactArgs(1), RunE: runBtAppDelete,
	}
	bigtableAppDescribeCmd = &cobra.Command{
		Use: "describe APP_PROFILE", Short: "Describe a Bigtable app profile",
		Args: cobra.ExactArgs(1), RunE: runBtAppDescribe,
	}
	bigtableAppListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable app profiles on an instance",
		Args: cobra.NoArgs, RunE: runBtAppList,
	}
	bigtableAppUpdateCmd = &cobra.Command{
		Use: "update APP_PROFILE", Short: "Update a Bigtable app profile",
		Args: cobra.ExactArgs(1), RunE: runBtAppUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		bigtableAppCreateCmd, bigtableAppDeleteCmd, bigtableAppDescribeCmd,
		bigtableAppListCmd, bigtableAppUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagBtAppInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&flagBtAppFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{bigtableAppCreateCmd, bigtableAppUpdateCmd} {
		c.Flags().StringVar(&flagBtAppConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the AppProfile body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	bigtableAppUpdateCmd.Flags().StringVar(&flagBtAppUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	bigtableAppListCmd.Flags().Int64Var(&flagBtAppPageSize, "page-size", 0, "Maximum results per page")

	bigtableAppProfilesCmd.AddCommand(all...)
	bigtableCmd.AddCommand(bigtableAppProfilesCmd)
}

func btAppName(id string) (string, error) {
	parent, err := btInstanceName(flagBtAppInstance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/appProfiles/%s", parent, id), nil
}

func runBtAppCreate(cmd *cobra.Command, args []string) error {
	parent, err := btInstanceName(flagBtAppInstance)
	if err != nil {
		return err
	}
	body := &bigtableadmin.AppProfile{}
	if err := loadYAMLOrJSONInto(flagBtAppConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.AppProfiles.Create(parent, body).AppProfileId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating app profile: %w", err)
	}
	fmt.Printf("Created app profile [%s].\n", args[0])
	return emitFormatted(got, flagBtAppFormat)
}

func runBtAppDelete(cmd *cobra.Command, args []string) error {
	name, err := btAppName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.AppProfiles.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting app profile: %w", err)
	}
	fmt.Printf("Deleted app profile [%s].\n", args[0])
	return nil
}

func runBtAppDescribe(cmd *cobra.Command, args []string) error {
	name, err := btAppName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.AppProfiles.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing app profile: %w", err)
	}
	return emitFormatted(got, flagBtAppFormat)
}

func runBtAppList(cmd *cobra.Command, args []string) error {
	parent, err := btInstanceName(flagBtAppInstance)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.AppProfile
	pageToken := ""
	for {
		call := svc.Projects.Instances.AppProfiles.List(parent).Context(ctx)
		if flagBtAppPageSize > 0 {
			call = call.PageSize(flagBtAppPageSize)
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
	return emitFormatted(all, flagBtAppFormat)
}

func runBtAppUpdate(cmd *cobra.Command, args []string) error {
	name, err := btAppName(args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.AppProfile{}
	if err := loadYAMLOrJSONInto(flagBtAppConfigFile, body); err != nil {
		return err
	}
	mask := flagBtAppUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
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
	fmt.Printf("Update request issued for app profile [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBtAppFormat)
}
