package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud network-management network-monitoring-providers (#954) ---
//
// The networkMonitoringProviders resource lives on the v1beta1 surface of
// networkmanagement.googleapis.com and is not modeled by the v1
// google.golang.org/api client, so this command surface uses the shared
// restClient (see netmgmtMonRest in network_management.go).

var netmgmtMonCmd = &cobra.Command{
	Use:   "network-monitoring-providers",
	Short: "Manage Network Monitoring providers",
}

var (
	flagNMMonLocation   string
	flagNMMonFormat     string
	flagNMMonConfigFile string
	flagNMMonPageSize   int64
)

var (
	netmgmtMonCreateCmd = &cobra.Command{
		Use: "create PROVIDER", Short: "Create a network monitoring provider",
		Args: cobra.ExactArgs(1), RunE: runNMMonCreate,
	}
	netmgmtMonDeleteCmd = &cobra.Command{
		Use: "delete PROVIDER", Short: "Delete a network monitoring provider",
		Args: cobra.ExactArgs(1), RunE: runNMMonDelete,
	}
	netmgmtMonDescribeCmd = &cobra.Command{
		Use: "describe PROVIDER", Short: "Describe a network monitoring provider",
		Args: cobra.ExactArgs(1), RunE: runNMMonDescribe,
	}
	netmgmtMonGenMPConfigCmd = &cobra.Command{
		Use:   "generate-monitoring-point-config PROVIDER",
		Short: "Generate a monitoring point config for a network monitoring provider",
		Args:  cobra.ExactArgs(1), RunE: runNMMonGenMPConfig,
	}
	netmgmtMonGenAccessTokenCmd = &cobra.Command{
		Use:   "generate-provider-access-token PROVIDER",
		Short: "Generate a provider access token for a network monitoring provider",
		Args:  cobra.ExactArgs(1), RunE: runNMMonGenAccessToken,
	}
	netmgmtMonListCmd = &cobra.Command{
		Use: "list", Short: "List network monitoring providers",
		Args: cobra.NoArgs, RunE: runNMMonList,
	}
)

func init() {
	all := []*cobra.Command{
		netmgmtMonCreateCmd, netmgmtMonDeleteCmd, netmgmtMonDescribeCmd,
		netmgmtMonGenMPConfigCmd, netmgmtMonGenAccessTokenCmd, netmgmtMonListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNMMonLocation, "location", "",
			"Network monitoring provider location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNMMonFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		netmgmtMonCreateCmd, netmgmtMonGenMPConfigCmd, netmgmtMonGenAccessTokenCmd,
	} {
		c.Flags().StringVar(&flagNMMonConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	netmgmtMonListCmd.Flags().Int64Var(&flagNMMonPageSize, "page-size", 0, "Maximum results per page")

	netmgmtMonCmd.AddCommand(all...)
	networkManagementCmd.AddCommand(netmgmtMonCmd)
}

func nmMonParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagNMMonLocation), nil
}

func nmMonName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := nmMonParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/networkMonitoringProviders/%s", parent, id), nil
}

func runNMMonCreate(cmd *cobra.Command, args []string) error {
	parent, err := nmMonParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagNMMonConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("networkMonitoringProviderId", args[0])
	ctx := context.Background()
	var op map[string]any
	if err := netmgmtMonRest.do(ctx, http.MethodPost, "/"+parent+"/networkMonitoringProviders", q, body, &op); err != nil {
		return fmt.Errorf("creating network monitoring provider: %w", err)
	}
	fmt.Printf("Create request issued for network monitoring provider [%s].\n", args[0])
	return emitFormatted(op, flagNMMonFormat)
}

func runNMMonDelete(cmd *cobra.Command, args []string) error {
	name, err := nmMonName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := netmgmtMonRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting network monitoring provider: %w", err)
	}
	fmt.Printf("Delete request issued for network monitoring provider [%s].\n", args[0])
	return emitFormatted(op, flagNMMonFormat)
}

func runNMMonDescribe(cmd *cobra.Command, args []string) error {
	name, err := nmMonName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := netmgmtMonRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing network monitoring provider: %w", err)
	}
	return emitFormatted(got, flagNMMonFormat)
}

func runNMMonGenMPConfig(cmd *cobra.Command, args []string) error {
	name, err := nmMonName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagNMMonConfigFile, &body); err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := netmgmtMonRest.do(ctx, http.MethodPost, "/"+name+":generateMonitoringPointConfig", nil, body, &got); err != nil {
		return fmt.Errorf("generating monitoring point config: %w", err)
	}
	return emitFormatted(got, flagNMMonFormat)
}

func runNMMonGenAccessToken(cmd *cobra.Command, args []string) error {
	name, err := nmMonName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagNMMonConfigFile, &body); err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := netmgmtMonRest.do(ctx, http.MethodPost, "/"+name+":generateProviderAccessToken", nil, body, &got); err != nil {
		return fmt.Errorf("generating provider access token: %w", err)
	}
	return emitFormatted(got, flagNMMonFormat)
}

func runNMMonList(cmd *cobra.Command, args []string) error {
	parent, err := nmMonParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := netmgmtMonRest.paginate(ctx, "/"+parent+"/networkMonitoringProviders", nil, "networkMonitoringProviders", flagNMMonPageSize)
	if err != nil {
		return fmt.Errorf("listing network monitoring providers: %w", err)
	}
	return emitFormatted(items, flagNMMonFormat)
}
