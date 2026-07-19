package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee developers (#1379) ---

var apigeeDevelopersCmd = &cobra.Command{Use: "developers", Short: "Manage Apigee developers"}

var (
	flagApigeeDevOrganization string
	flagApigeeDevFormat       string
	flagApigeeDevConfigFile   string
)

var (
	apigeeDevCreateCmd = &cobra.Command{
		Use: "create RESOURCE", Short: "Create an Apigee developer",
		Args: cobra.ExactArgs(1), RunE: runApigeeDevCreate,
	}
	apigeeDevDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an Apigee developer",
		Args: cobra.ExactArgs(1), RunE: runApigeeDevDelete,
	}
	apigeeDevDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an Apigee developer",
		Args: cobra.ExactArgs(1), RunE: runApigeeDevDescribe,
	}
	apigeeDevListCmd = &cobra.Command{
		Use: "list", Short: "List Apigee developers in an organization",
		Args: cobra.NoArgs, RunE: runApigeeDevList,
	}
	apigeeDevUpdateCmd = &cobra.Command{
		Use: "update RESOURCE", Short: "Update an Apigee developer",
		Args: cobra.ExactArgs(1), RunE: runApigeeDevUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		apigeeDevCreateCmd, apigeeDevDeleteCmd, apigeeDevDescribeCmd,
		apigeeDevListCmd, apigeeDevUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagApigeeDevOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&flagApigeeDevFormat, "format", "", "Output format")
	}
	apigeeDevCreateCmd.Flags().StringVar(&flagApigeeDevConfigFile, "config-file", "", "Path to YAML/JSON developer definition (required)")
	_ = apigeeDevCreateCmd.MarkFlagRequired("config-file")
	apigeeDevUpdateCmd.Flags().StringVar(&flagApigeeDevConfigFile, "config-file", "", "Path to YAML/JSON developer definition (required)")
	_ = apigeeDevUpdateCmd.MarkFlagRequired("config-file")

	apigeeDevelopersCmd.AddCommand(all...)
	apigeeCmd.AddCommand(apigeeDevelopersCmd)
}

func apigeeDevName(id string) (string, error) {
	return apigeeResource(flagApigeeDevOrganization, "developers", id)
}

func runApigeeDevCreate(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(flagApigeeDevOrganization)
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1Developer{}
	if err := loadYAMLOrJSONInto(flagApigeeDevConfigFile, body); err != nil {
		return err
	}
	if body.Email == "" {
		body.Email = args[0]
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Developers.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating developer: %w", err)
	}
	return emitFormatted(got, flagApigeeDevFormat)
}

func runApigeeDevDelete(cmd *cobra.Command, args []string) error {
	name, err := apigeeDevName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Organizations.Developers.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting developer: %w", err)
	}
	fmt.Printf("Deleted developer [%s].\n", args[0])
	return nil
}

func runApigeeDevDescribe(cmd *cobra.Command, args []string) error {
	name, err := apigeeDevName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Developers.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing developer: %w", err)
	}
	return emitFormatted(got, flagApigeeDevFormat)
}

func runApigeeDevList(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(flagApigeeDevOrganization)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Organizations.Developers.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing developers: %w", err)
	}
	var all []*apigee.GoogleCloudApigeeV1Developer
	all = append(all, resp.Developer...)
	return emitFormatted(all, flagApigeeDevFormat)
}

func runApigeeDevUpdate(cmd *cobra.Command, args []string) error {
	name, err := apigeeDevName(args[0])
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1Developer{}
	if err := loadYAMLOrJSONInto(flagApigeeDevConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Developers.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating developer: %w", err)
	}
	return emitFormatted(got, flagApigeeDevFormat)
}
