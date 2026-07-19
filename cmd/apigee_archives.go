package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee archives (#1377) ---

var apigeeArchivesCmd = &cobra.Command{Use: "archives", Short: "Manage Apigee archive deployments"}

var (
	flagApigeeArcOrganization string
	flagApigeeArcEnvironment  string
	flagApigeeArcFormat       string
	flagApigeeArcConfigFile   string
	flagApigeeArcPageSize     int64
)

var (
	apigeeArcCreateCmd = &cobra.Command{
		Use: "create RESOURCE", Short: "Create an archive deployment",
		Args: cobra.ExactArgs(1), RunE: runApigeeArcCreate,
	}
	apigeeArcDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an archive deployment",
		Args: cobra.ExactArgs(1), RunE: runApigeeArcDelete,
	}
	apigeeArcDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an archive deployment",
		Args: cobra.ExactArgs(1), RunE: runApigeeArcDescribe,
	}
	apigeeArcListCmd = &cobra.Command{
		Use: "list", Short: "List archive deployments in an environment",
		Args: cobra.NoArgs, RunE: runApigeeArcList,
	}
	apigeeArcUpdateCmd = &cobra.Command{
		Use: "update RESOURCE", Short: "Update an archive deployment",
		Args: cobra.ExactArgs(1), RunE: runApigeeArcUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		apigeeArcCreateCmd, apigeeArcDeleteCmd, apigeeArcDescribeCmd,
		apigeeArcListCmd, apigeeArcUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagApigeeArcOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&flagApigeeArcEnvironment, "environment", "", "Apigee environment (required)")
		_ = c.MarkFlagRequired("environment")
		c.Flags().StringVar(&flagApigeeArcFormat, "format", "", "Output format")
	}
	apigeeArcCreateCmd.Flags().StringVar(&flagApigeeArcConfigFile, "config-file", "", "Path to YAML/JSON archive deployment definition (required)")
	_ = apigeeArcCreateCmd.MarkFlagRequired("config-file")
	apigeeArcUpdateCmd.Flags().StringVar(&flagApigeeArcConfigFile, "config-file", "", "Path to YAML/JSON archive deployment definition (required)")
	_ = apigeeArcUpdateCmd.MarkFlagRequired("config-file")
	apigeeArcListCmd.Flags().Int64Var(&flagApigeeArcPageSize, "page-size", 0, "Maximum results per page")

	apigeeArchivesCmd.AddCommand(all...)
	apigeeCmd.AddCommand(apigeeArchivesCmd)
}

func apigeeArcParent() (string, error) {
	if flagApigeeArcEnvironment == "" {
		return "", fmt.Errorf("--environment is required")
	}
	org, err := apigeeOrgName(flagApigeeArcOrganization)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/environments/%s", org, flagApigeeArcEnvironment), nil
}

func apigeeArcName(id string) (string, error) {
	parent, err := apigeeArcParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/archiveDeployments/%s", parent, id), nil
}

func runApigeeArcCreate(cmd *cobra.Command, args []string) error {
	parent, err := apigeeArcParent()
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1ArchiveDeployment{}
	if err := loadYAMLOrJSONInto(flagApigeeArcConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = args[0]
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Environments.ArchiveDeployments.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating archive deployment: %w", err)
	}
	return emitFormatted(op, flagApigeeArcFormat)
}

func runApigeeArcDelete(cmd *cobra.Command, args []string) error {
	name, err := apigeeArcName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Organizations.Environments.ArchiveDeployments.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting archive deployment: %w", err)
	}
	fmt.Printf("Deleted archive deployment [%s].\n", args[0])
	return nil
}

func runApigeeArcDescribe(cmd *cobra.Command, args []string) error {
	name, err := apigeeArcName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Environments.ArchiveDeployments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing archive deployment: %w", err)
	}
	return emitFormatted(got, flagApigeeArcFormat)
}

func runApigeeArcList(cmd *cobra.Command, args []string) error {
	parent, err := apigeeArcParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apigee.GoogleCloudApigeeV1ArchiveDeployment
	pageToken := ""
	for {
		call := svc.Organizations.Environments.ArchiveDeployments.List(parent).Context(ctx)
		if flagApigeeArcPageSize > 0 {
			call = call.PageSize(flagApigeeArcPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing archive deployments: %w", err)
		}
		all = append(all, resp.ArchiveDeployments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagApigeeArcFormat)
}

func runApigeeArcUpdate(cmd *cobra.Command, args []string) error {
	name, err := apigeeArcName(args[0])
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1ArchiveDeployment{}
	if err := loadYAMLOrJSONInto(flagApigeeArcConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Environments.ArchiveDeployments.Patch(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating archive deployment: %w", err)
	}
	return emitFormatted(got, flagApigeeArcFormat)
}
