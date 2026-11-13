package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud telco-automation orchestration-cluster (#848) ---

func (s *telcoAutomationCmds) initTelcoAutomationOrchestrationCluster() {
	all := []*cobra.Command{s.taOCCreateCmd, s.taOCDeleteCmd, s.taOCDescribeCmd, s.taOCListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagTAOCLocation, "location", "", "Telco Automation location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagTAOCFormat, "format", "", "Output format")
	}
	s.taOCCreateCmd.Flags().StringVar(&s.flagTAOCConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the orchestration cluster body (required)")
	_ = s.taOCCreateCmd.MarkFlagRequired("config-file")
	s.taOCListCmd.Flags().Int64Var(&s.flagTAOCPageSize, "page-size", 0, "Maximum results per page")

	s.taOCCmd.AddCommand(all...)
	s.telcoAutomationCmd.AddCommand(s.taOCCmd)
}

func (s *telcoAutomationCmds) taOCParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagTAOCLocation), nil
}

func (s *telcoAutomationCmds) taOCName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.taOCParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/orchestrationClusters/%s", parent, id), nil
}

func (s *telcoAutomationCmds) runTAOCCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.taOCParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagTAOCConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("orchestrationClusterId", args[0])
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(telcoAutomationRest).do(ctx, http.MethodPost, "/"+parent+"/orchestrationClusters", q, body, &op); err != nil {
		return fmt.Errorf("creating orchestration cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for orchestration cluster [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagTAOCFormat)
}

func (s *telcoAutomationCmds) runTAOCDelete(cmd *cobra.Command, args []string) error {
	name, err := s.taOCName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(telcoAutomationRest).do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting orchestration cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for orchestration cluster [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagTAOCFormat)
}

func (s *telcoAutomationCmds) runTAOCDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.taOCName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(telcoAutomationRest).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing orchestration cluster: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagTAOCFormat)
}

func (s *telcoAutomationCmds) runTAOCList(cmd *cobra.Command, args []string) error {
	parent, err := s.taOCParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(telcoAutomationRest).paginate(ctx, "/"+parent+"/orchestrationClusters", nil, "orchestrationClusters", s.flagTAOCPageSize)
	if err != nil {
		return fmt.Errorf("listing orchestration clusters: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagTAOCFormat)
}
