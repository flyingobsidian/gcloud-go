package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager cloud-bindings (#1442) ---

var acmCBCmd = &cobra.Command{Use: "cloud-bindings", Short: "Manage cloud access bindings"}

var (
	flagACMCBFormat       string
	flagACMCBOrganization string
	flagACMCBConfigFile   string
	flagACMCBPageSize     int64
)

var (
	acmCBCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a cloud access binding",
		Args: cobra.NoArgs, RunE: runACMCBCreate,
	}
	acmCBDeleteCmd = &cobra.Command{
		Use: "delete BINDING_ID", Short: "Delete a cloud access binding",
		Args: cobra.ExactArgs(1), RunE: runACMCBDelete,
	}
	acmCBDescribeCmd = &cobra.Command{
		Use: "describe BINDING_ID", Short: "Describe a cloud access binding",
		Args: cobra.ExactArgs(1), RunE: runACMCBDescribe,
	}
	acmCBListCmd = &cobra.Command{
		Use: "list", Short: "List cloud access bindings",
		Args: cobra.NoArgs, RunE: runACMCBList,
	}
	acmCBUpdateCmd = &cobra.Command{
		Use: "update BINDING_ID", Short: "Update a cloud access binding",
		Args: cobra.ExactArgs(1), RunE: runACMCBUpdate,
	}
)

func init() {
	all := []*cobra.Command{acmCBCreateCmd, acmCBDeleteCmd, acmCBDescribeCmd, acmCBListCmd, acmCBUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagACMCBFormat, "format", "", "Output format")
		c.Flags().StringVar(&flagACMCBOrganization, "organization", "", "Organization ID (required)")
		_ = c.MarkFlagRequired("organization")
	}
	for _, c := range []*cobra.Command{acmCBCreateCmd, acmCBUpdateCmd} {
		c.Flags().StringVar(&flagACMCBConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the GcpUserAccessBinding body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	acmCBListCmd.Flags().Int64Var(&flagACMCBPageSize, "page-size", 0, "Maximum results per page")

	acmCBCmd.AddCommand(all...)
	accessContextManagerCmd.AddCommand(acmCBCmd)
}

func acmCBParent(org string) string {
	return "organizations/" + org
}

func acmCBResource(org, binding string) string {
	return fmt.Sprintf("%s/gcpUserAccessBindings/%s", acmCBParent(org), binding)
}

func runACMCBCreate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.GcpUserAccessBinding{}
	if err := loadYAMLOrJSONInto(flagACMCBConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.GcpUserAccessBindings.Create(acmCBParent(flagACMCBOrganization), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cloud access binding: %w", err)
	}
	fmt.Println("Create request issued for cloud access binding.")
	return emitFormatted(op, flagACMCBFormat)
}

func runACMCBDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.GcpUserAccessBindings.Delete(acmCBResource(flagACMCBOrganization, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cloud access binding: %w", err)
	}
	fmt.Printf("Delete request issued for cloud access binding [%s].\n", args[0])
	return emitFormatted(op, flagACMCBFormat)
}

func runACMCBDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.GcpUserAccessBindings.Get(acmCBResource(flagACMCBOrganization, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cloud access binding: %w", err)
	}
	return emitFormatted(got, flagACMCBFormat)
}

func runACMCBList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*accesscontextmanager.GcpUserAccessBinding
	pageToken := ""
	for {
		call := svc.Organizations.GcpUserAccessBindings.List(acmCBParent(flagACMCBOrganization)).Context(ctx)
		if flagACMCBPageSize > 0 {
			call = call.PageSize(flagACMCBPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing cloud access bindings: %w", err)
		}
		all = append(all, resp.GcpUserAccessBindings...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagACMCBFormat)
}

func runACMCBUpdate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.GcpUserAccessBinding{}
	if err := loadYAMLOrJSONInto(flagACMCBConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Organizations.GcpUserAccessBindings.Patch(acmCBResource(flagACMCBOrganization, args[0]), body).Context(ctx)
	if mask := joinMask(nonEmptyJSONFields(body)); mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating cloud access binding: %w", err)
	}
	fmt.Printf("Update request issued for cloud access binding [%s].\n", args[0])
	return emitFormatted(op, flagACMCBFormat)
}
