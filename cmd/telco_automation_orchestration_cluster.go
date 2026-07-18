package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud telco-automation orchestration-cluster (#848) ---

var taOCCmd = &cobra.Command{Use: "orchestration-cluster", Short: "Manage Telco Automation orchestration clusters"}

var (
	flagTAOCLocation   string
	flagTAOCFormat     string
	flagTAOCConfigFile string
	flagTAOCPageSize   int64
)

var (
	taOCCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a Telco Automation orchestration cluster",
		Args: cobra.ExactArgs(1), RunE: runTAOCCreate,
	}
	taOCDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a Telco Automation orchestration cluster",
		Args: cobra.ExactArgs(1), RunE: runTAOCDelete,
	}
	taOCDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a Telco Automation orchestration cluster",
		Args: cobra.ExactArgs(1), RunE: runTAOCDescribe,
	}
	taOCListCmd = &cobra.Command{
		Use: "list", Short: "List Telco Automation orchestration clusters",
		Args: cobra.NoArgs, RunE: runTAOCList,
	}
)

func init() {
	all := []*cobra.Command{taOCCreateCmd, taOCDeleteCmd, taOCDescribeCmd, taOCListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagTAOCLocation, "location", "", "Telco Automation location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagTAOCFormat, "format", "", "Output format")
	}
	taOCCreateCmd.Flags().StringVar(&flagTAOCConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the orchestration cluster body (required)")
	_ = taOCCreateCmd.MarkFlagRequired("config-file")
	taOCListCmd.Flags().Int64Var(&flagTAOCPageSize, "page-size", 0, "Maximum results per page")

	taOCCmd.AddCommand(all...)
	telcoAutomationCmd.AddCommand(taOCCmd)
}

func taOCParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagTAOCLocation), nil
}

func taOCName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := taOCParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/orchestrationClusters/%s", parent, id), nil
}

func runTAOCCreate(cmd *cobra.Command, args []string) error {
	parent, err := taOCParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagTAOCConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("orchestrationClusterId", args[0])
	ctx := context.Background()
	var op map[string]any
	if err := telcoAutomationRest.do(ctx, http.MethodPost, "/"+parent+"/orchestrationClusters", q, body, &op); err != nil {
		return fmt.Errorf("creating orchestration cluster: %w", err)
	}
	fmt.Printf("Create request issued for orchestration cluster [%s].\n", args[0])
	return emitFormatted(op, flagTAOCFormat)
}

func runTAOCDelete(cmd *cobra.Command, args []string) error {
	name, err := taOCName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := telcoAutomationRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting orchestration cluster: %w", err)
	}
	fmt.Printf("Delete request issued for orchestration cluster [%s].\n", args[0])
	return emitFormatted(op, flagTAOCFormat)
}

func runTAOCDescribe(cmd *cobra.Command, args []string) error {
	name, err := taOCName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := telcoAutomationRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing orchestration cluster: %w", err)
	}
	return emitFormatted(got, flagTAOCFormat)
}

func runTAOCList(cmd *cobra.Command, args []string) error {
	parent, err := taOCParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := telcoAutomationRest.paginate(ctx, "/"+parent+"/orchestrationClusters", nil, "orchestrationClusters", flagTAOCPageSize)
	if err != nil {
		return fmt.Errorf("listing orchestration clusters: %w", err)
	}
	return emitFormatted(items, flagTAOCFormat)
}
