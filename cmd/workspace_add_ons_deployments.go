package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud workspace-add-ons deployments (#853) ---

var waoDepCmd = &cobra.Command{Use: "deployments", Short: "Manage Google Workspace Add-ons deployments"}

var (
	flagWAODepFormat     string
	flagWAODepConfigFile string
	flagWAODepPageSize   int64
)

var (
	waoDepCreateCmd = &cobra.Command{
		Use: "create DEPLOYMENT", Short: "Create a Workspace Add-ons deployment",
		Args: cobra.ExactArgs(1), RunE: runWAODepCreate,
	}
	waoDepDeleteCmd = &cobra.Command{
		Use: "delete DEPLOYMENT", Short: "Delete a Workspace Add-ons deployment",
		Args: cobra.ExactArgs(1), RunE: runWAODepDelete,
	}
	waoDepDescribeCmd = &cobra.Command{
		Use: "describe DEPLOYMENT", Short: "Describe a Workspace Add-ons deployment",
		Args: cobra.ExactArgs(1), RunE: runWAODepDescribe,
	}
	waoDepListCmd = &cobra.Command{
		Use: "list", Short: "List Workspace Add-ons deployments",
		Args: cobra.NoArgs, RunE: runWAODepList,
	}
	waoDepReplaceCmd = &cobra.Command{
		Use: "replace DEPLOYMENT", Short: "Replace a Workspace Add-ons deployment",
		Args: cobra.ExactArgs(1), RunE: runWAODepReplace,
	}
	waoDepInstallCmd = &cobra.Command{
		Use: "install DEPLOYMENT", Short: "Install a Workspace Add-ons deployment in developer mode",
		Args: cobra.ExactArgs(1), RunE: runWAODepInstall,
	}
	waoDepUninstallCmd = &cobra.Command{
		Use: "uninstall DEPLOYMENT", Short: "Uninstall a Workspace Add-ons deployment from developer mode",
		Args: cobra.ExactArgs(1), RunE: runWAODepUninstall,
	}
	waoDepInstallStatusCmd = &cobra.Command{
		Use: "install-status DEPLOYMENT", Short: "Get the install status of a Workspace Add-ons deployment",
		Args: cobra.ExactArgs(1), RunE: runWAODepInstallStatus,
	}
)

func init() {
	all := []*cobra.Command{
		waoDepCreateCmd, waoDepDeleteCmd, waoDepDescribeCmd, waoDepListCmd,
		waoDepReplaceCmd, waoDepInstallCmd, waoDepUninstallCmd, waoDepInstallStatusCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagWAODepFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{waoDepCreateCmd, waoDepReplaceCmd} {
		c.Flags().StringVar(&flagWAODepConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the deployment body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	waoDepListCmd.Flags().Int64Var(&flagWAODepPageSize, "page-size", 0, "Maximum results per page")

	waoDepCmd.AddCommand(all...)
	workspaceAddOnsCmd.AddCommand(waoDepCmd)
}

func waoDepParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s", project), nil
}

func waoDepName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := waoDepParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/deployments/%s", parent, id), nil
}

func runWAODepCreate(cmd *cobra.Command, args []string) error {
	parent, err := waoDepParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagWAODepConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("deploymentId", args[0])
	ctx := context.Background()
	var got map[string]any
	if err := workspaceAddOnsRest.do(ctx, http.MethodPost, "/"+parent+"/deployments", q, body, &got); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}
	fmt.Printf("Created deployment [%s].\n", args[0])
	return emitFormatted(got, flagWAODepFormat)
}

func runWAODepDelete(cmd *cobra.Command, args []string) error {
	name, err := waoDepName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := workspaceAddOnsRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, nil); err != nil {
		return fmt.Errorf("deleting deployment: %w", err)
	}
	fmt.Printf("Deleted deployment [%s].\n", args[0])
	return nil
}

func runWAODepDescribe(cmd *cobra.Command, args []string) error {
	name, err := waoDepName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := workspaceAddOnsRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing deployment: %w", err)
	}
	return emitFormatted(got, flagWAODepFormat)
}

func runWAODepList(cmd *cobra.Command, args []string) error {
	parent, err := waoDepParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := workspaceAddOnsRest.paginate(ctx, "/"+parent+"/deployments", nil, "deployments", flagWAODepPageSize)
	if err != nil {
		return fmt.Errorf("listing deployments: %w", err)
	}
	return emitFormatted(items, flagWAODepFormat)
}

func runWAODepReplace(cmd *cobra.Command, args []string) error {
	name, err := waoDepName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagWAODepConfigFile, &body); err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := workspaceAddOnsRest.do(ctx, http.MethodPut, "/"+name, nil, body, &got); err != nil {
		return fmt.Errorf("replacing deployment: %w", err)
	}
	fmt.Printf("Replaced deployment [%s].\n", args[0])
	return emitFormatted(got, flagWAODepFormat)
}

func runWAODepInstall(cmd *cobra.Command, args []string) error {
	name, err := waoDepName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := workspaceAddOnsRest.do(ctx, http.MethodPost, "/"+name+":install", nil, map[string]any{}, nil); err != nil {
		return fmt.Errorf("installing deployment: %w", err)
	}
	fmt.Printf("Installed deployment [%s].\n", args[0])
	return nil
}

func runWAODepUninstall(cmd *cobra.Command, args []string) error {
	name, err := waoDepName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := workspaceAddOnsRest.do(ctx, http.MethodPost, "/"+name+":uninstall", nil, map[string]any{}, nil); err != nil {
		return fmt.Errorf("uninstalling deployment: %w", err)
	}
	fmt.Printf("Uninstalled deployment [%s].\n", args[0])
	return nil
}

func runWAODepInstallStatus(cmd *cobra.Command, args []string) error {
	name, err := waoDepName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := workspaceAddOnsRest.do(ctx, http.MethodGet, "/"+name+"/installStatus", nil, nil, &got); err != nil {
		return fmt.Errorf("getting deployment install status: %w", err)
	}
	return emitFormatted(got, flagWAODepFormat)
}
