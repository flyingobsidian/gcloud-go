package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	notebooksv1 "google.golang.org/api/notebooks/v1"
)

// --- gcloud notebooks environments (#1061) ---

var notebooksEnvCmd = &cobra.Command{Use: "environments", Short: "Manage notebook environments"}

var (
	flagNotebooksEnvLocation   string
	flagNotebooksEnvFormat     string
	flagNotebooksEnvConfigFile string
	flagNotebooksEnvPageSize   int64
)

var (
	notebooksEnvCreateCmd = &cobra.Command{
		Use: "create ENVIRONMENT", Short: "Create a notebook environment",
		Args: cobra.ExactArgs(1), RunE: runNotebooksEnvCreate,
	}
	notebooksEnvDeleteCmd = &cobra.Command{
		Use: "delete ENVIRONMENT", Short: "Delete a notebook environment",
		Args: cobra.ExactArgs(1), RunE: runNotebooksEnvDelete,
	}
	notebooksEnvDescribeCmd = &cobra.Command{
		Use: "describe ENVIRONMENT", Short: "Describe a notebook environment",
		Args: cobra.ExactArgs(1), RunE: runNotebooksEnvDescribe,
	}
	notebooksEnvListCmd = &cobra.Command{
		Use: "list", Short: "List notebook environments",
		Args: cobra.NoArgs, RunE: runNotebooksEnvList,
	}
)

func init() {
	all := []*cobra.Command{
		notebooksEnvCreateCmd, notebooksEnvDeleteCmd, notebooksEnvDescribeCmd, notebooksEnvListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNotebooksEnvLocation, "location", "", "Notebook location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNotebooksEnvFormat, "format", "", "Output format")
	}
	notebooksEnvCreateCmd.Flags().StringVar(&flagNotebooksEnvConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Environment body (required)")
	_ = notebooksEnvCreateCmd.MarkFlagRequired("config-file")
	notebooksEnvListCmd.Flags().Int64Var(&flagNotebooksEnvPageSize, "page-size", 0, "Maximum results per page")

	notebooksEnvCmd.AddCommand(all...)
	notebooksCmd.AddCommand(notebooksEnvCmd)
}

func notebooksEnvParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagNotebooksEnvLocation), nil
}

func notebooksEnvName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := notebooksEnvParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/environments/%s", parent, id), nil
}

func runNotebooksEnvCreate(cmd *cobra.Command, args []string) error {
	parent, err := notebooksEnvParent()
	if err != nil {
		return err
	}
	body := &notebooksv1.Environment{}
	if err := loadYAMLOrJSONInto(flagNotebooksEnvConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.Create(parent, body).EnvironmentId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating notebook environment: %w", err)
	}
	fmt.Printf("Create request issued for notebook environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksEnvFormat)
}

func runNotebooksEnvDelete(cmd *cobra.Command, args []string) error {
	name, err := notebooksEnvName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting notebook environment: %w", err)
	}
	fmt.Printf("Delete request issued for notebook environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksEnvFormat)
}

func runNotebooksEnvDescribe(cmd *cobra.Command, args []string) error {
	name, err := notebooksEnvName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing notebook environment: %w", err)
	}
	return emitFormatted(got, flagNotebooksEnvFormat)
}

func runNotebooksEnvList(cmd *cobra.Command, args []string) error {
	parent, err := notebooksEnvParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*notebooksv1.Environment
	pageToken := ""
	for {
		call := svc.Projects.Locations.Environments.List(parent).Context(ctx)
		if flagNotebooksEnvPageSize > 0 {
			call = call.PageSize(flagNotebooksEnvPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing notebook environments: %w", err)
		}
		all = append(all, resp.Environments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNotebooksEnvFormat)
}
