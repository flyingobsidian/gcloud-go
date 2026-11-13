package cmd

import (
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

func (s *networkManagementCmds) initNetworkManagementNetworkMonitoringProviders() {
	all := []*cobra.Command{
		s.netmgmtMonCreateCmd, s.netmgmtMonDeleteCmd, s.netmgmtMonDescribeCmd,
		s.netmgmtMonGenMPConfigCmd, s.netmgmtMonGenAccessTokenCmd, s.netmgmtMonListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNMMonLocation, "location", "",
			"Network monitoring provider location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNMMonFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		s.netmgmtMonCreateCmd, s.netmgmtMonGenMPConfigCmd, s.netmgmtMonGenAccessTokenCmd,
	} {
		c.Flags().StringVar(&s.flagNMMonConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.netmgmtMonListCmd.Flags().Int64Var(&s.flagNMMonPageSize, "page-size", 0, "Maximum results per page")

	s.netmgmtMonCmd.AddCommand(all...)
	s.networkManagementCmd.AddCommand(s.netmgmtMonCmd)
}

func (s *networkManagementCmds) nmMonParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagNMMonLocation), nil
}

func (s *networkManagementCmds) nmMonName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.nmMonParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/networkMonitoringProviders/%s", parent, id), nil
}

func (s *networkManagementCmds) runNMMonCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.nmMonParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagNMMonConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("networkMonitoringProviderId", args[0])
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(netmgmtMonRest).do(ctx, http.MethodPost, "/"+parent+"/networkMonitoringProviders", q, body, &op); err != nil {
		return fmt.Errorf("creating network monitoring provider: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for network monitoring provider [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagNMMonFormat)
}

func (s *networkManagementCmds) runNMMonDelete(cmd *cobra.Command, args []string) error {
	name, err := s.nmMonName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(netmgmtMonRest).do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting network monitoring provider: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for network monitoring provider [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagNMMonFormat)
}

func (s *networkManagementCmds) runNMMonDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.nmMonName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(netmgmtMonRest).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing network monitoring provider: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNMMonFormat)
}

func (s *networkManagementCmds) runNMMonGenMPConfig(cmd *cobra.Command, args []string) error {
	name, err := s.nmMonName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagNMMonConfigFile, &body); err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(netmgmtMonRest).do(ctx, http.MethodPost, "/"+name+":generateMonitoringPointConfig", nil, body, &got); err != nil {
		return fmt.Errorf("generating monitoring point config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNMMonFormat)
}

func (s *networkManagementCmds) runNMMonGenAccessToken(cmd *cobra.Command, args []string) error {
	name, err := s.nmMonName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagNMMonConfigFile, &body); err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(netmgmtMonRest).do(ctx, http.MethodPost, "/"+name+":generateProviderAccessToken", nil, body, &got); err != nil {
		return fmt.Errorf("generating provider access token: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNMMonFormat)
}

func (s *networkManagementCmds) runNMMonList(cmd *cobra.Command, args []string) error {
	parent, err := s.nmMonParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(netmgmtMonRest).paginate(ctx, "/"+parent+"/networkMonitoringProviders", nil, "networkMonitoringProviders", s.flagNMMonPageSize)
	if err != nil {
		return fmt.Errorf("listing network monitoring providers: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagNMMonFormat)
}
