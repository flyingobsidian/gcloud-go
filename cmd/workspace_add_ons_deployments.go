package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud workspace-add-ons deployments (#853) ---

func (s *workspaceAddOnsCmds) initWorkspaceAddOnsDeployments() {
	all := []*cobra.Command{
		s.waoDepCreateCmd, s.waoDepDeleteCmd, s.waoDepDescribeCmd, s.waoDepListCmd,
		s.waoDepReplaceCmd, s.waoDepInstallCmd, s.waoDepUninstallCmd, s.waoDepInstallStatusCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagWAODepFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.waoDepCreateCmd, s.waoDepReplaceCmd} {
		c.Flags().StringVar(&s.flagWAODepConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the deployment body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.waoDepListCmd.Flags().Int64Var(&s.flagWAODepPageSize, "page-size", 0, "Maximum results per page")

	s.waoDepCmd.AddCommand(all...)
	s.workspaceAddOnsCmd.AddCommand(s.waoDepCmd)
}

func (s *workspaceAddOnsCmds) waoDepParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s", project), nil
}

func (s *workspaceAddOnsCmds) waoDepName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.waoDepParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/deployments/%s", parent, id), nil
}

func (s *workspaceAddOnsCmds) runWAODepCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.waoDepParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagWAODepConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("deploymentId", args[0])
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(workspaceAddOnsRest).do(ctx, http.MethodPost, "/"+parent+"/deployments", q, body, &got); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created deployment [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagWAODepFormat)
}

func (s *workspaceAddOnsCmds) runWAODepDelete(cmd *cobra.Command, args []string) error {
	name, err := s.waoDepName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if err := s.inv.rest(workspaceAddOnsRest).do(ctx, http.MethodDelete, "/"+name, nil, nil, nil); err != nil {
		return fmt.Errorf("deleting deployment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted deployment [%s].\n", args[0])
	return nil
}

func (s *workspaceAddOnsCmds) runWAODepDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.waoDepName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(workspaceAddOnsRest).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing deployment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagWAODepFormat)
}

func (s *workspaceAddOnsCmds) runWAODepList(cmd *cobra.Command, args []string) error {
	parent, err := s.waoDepParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(workspaceAddOnsRest).paginate(ctx, "/"+parent+"/deployments", nil, "deployments", s.flagWAODepPageSize)
	if err != nil {
		return fmt.Errorf("listing deployments: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagWAODepFormat)
}

func (s *workspaceAddOnsCmds) runWAODepReplace(cmd *cobra.Command, args []string) error {
	name, err := s.waoDepName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagWAODepConfigFile, &body); err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(workspaceAddOnsRest).do(ctx, http.MethodPut, "/"+name, nil, body, &got); err != nil {
		return fmt.Errorf("replacing deployment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Replaced deployment [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagWAODepFormat)
}

func (s *workspaceAddOnsCmds) runWAODepInstall(cmd *cobra.Command, args []string) error {
	name, err := s.waoDepName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if err := s.inv.rest(workspaceAddOnsRest).do(ctx, http.MethodPost, "/"+name+":install", nil, map[string]any{}, nil); err != nil {
		return fmt.Errorf("installing deployment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Installed deployment [%s].\n", args[0])
	return nil
}

func (s *workspaceAddOnsCmds) runWAODepUninstall(cmd *cobra.Command, args []string) error {
	name, err := s.waoDepName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if err := s.inv.rest(workspaceAddOnsRest).do(ctx, http.MethodPost, "/"+name+":uninstall", nil, map[string]any{}, nil); err != nil {
		return fmt.Errorf("uninstalling deployment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Uninstalled deployment [%s].\n", args[0])
	return nil
}

func (s *workspaceAddOnsCmds) runWAODepInstallStatus(cmd *cobra.Command, args []string) error {
	name, err := s.waoDepName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(workspaceAddOnsRest).do(ctx, http.MethodGet, "/"+name+"/installStatus", nil, nil, &got); err != nil {
		return fmt.Errorf("getting deployment install status: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagWAODepFormat)
}
