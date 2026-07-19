package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee environments (#1380) ---

var apigeeEnvironmentsCmd = &cobra.Command{Use: "environments", Short: "Manage Apigee environments"}

var (
	flagApigeeEnvOrganization string
	flagApigeeEnvFormat       string
	flagApigeeEnvConfigFile   string
)

var (
	apigeeEnvCreateCmd = &cobra.Command{
		Use: "create RESOURCE", Short: "Create an Apigee environment",
		Args: cobra.ExactArgs(1), RunE: runApigeeEnvCreate,
	}
	apigeeEnvDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an Apigee environment",
		Args: cobra.ExactArgs(1), RunE: runApigeeEnvDelete,
	}
	apigeeEnvDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an Apigee environment",
		Args: cobra.ExactArgs(1), RunE: runApigeeEnvDescribe,
	}
	apigeeEnvUpdateCmd = &cobra.Command{
		Use: "update RESOURCE", Short: "Update an Apigee environment",
		Args: cobra.ExactArgs(1), RunE: runApigeeEnvUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		apigeeEnvCreateCmd, apigeeEnvDeleteCmd, apigeeEnvDescribeCmd, apigeeEnvUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagApigeeEnvOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&flagApigeeEnvFormat, "format", "", "Output format")
	}
	apigeeEnvCreateCmd.Flags().StringVar(&flagApigeeEnvConfigFile, "config-file", "", "Path to YAML/JSON environment definition (required)")
	_ = apigeeEnvCreateCmd.MarkFlagRequired("config-file")
	apigeeEnvUpdateCmd.Flags().StringVar(&flagApigeeEnvConfigFile, "config-file", "", "Path to YAML/JSON environment definition (required)")
	_ = apigeeEnvUpdateCmd.MarkFlagRequired("config-file")

	apigeeEnvironmentsCmd.AddCommand(all...)
	apigeeCmd.AddCommand(apigeeEnvironmentsCmd)
}

func apigeeEnvName(id string) (string, error) {
	return apigeeResource(flagApigeeEnvOrganization, "environments", id)
}

func runApigeeEnvCreate(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(flagApigeeEnvOrganization)
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1Environment{}
	if err := loadYAMLOrJSONInto(flagApigeeEnvConfigFile, body); err != nil {
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
	op, err := svc.Organizations.Environments.Create(parent, body).Name(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating environment: %w", err)
	}
	return emitFormatted(op, flagApigeeEnvFormat)
}

func runApigeeEnvDelete(cmd *cobra.Command, args []string) error {
	name, err := apigeeEnvName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Environments.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting environment: %w", err)
	}
	if op != nil && op.Name != "" {
		fmt.Printf("Delete environment [%s] initiated (operation: %s).\n", args[0], op.Name)
	} else {
		fmt.Printf("Deleted environment [%s].\n", args[0])
	}
	return nil
}

func runApigeeEnvDescribe(cmd *cobra.Command, args []string) error {
	name, err := apigeeEnvName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Environments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing environment: %w", err)
	}
	return emitFormatted(got, flagApigeeEnvFormat)
}

func runApigeeEnvUpdate(cmd *cobra.Command, args []string) error {
	name, err := apigeeEnvName(args[0])
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1Environment{}
	if err := loadYAMLOrJSONInto(flagApigeeEnvConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Environments.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating environment: %w", err)
	}
	return emitFormatted(got, flagApigeeEnvFormat)
}
