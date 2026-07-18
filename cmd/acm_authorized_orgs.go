package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager authorized-orgs (#1441) ---

var acmAOCmd = &cobra.Command{Use: "authorized-orgs", Short: "Manage authorized org descriptions"}

var (
	flagACMAOFormat     string
	flagACMAOPolicy     string
	flagACMAOConfigFile string
	flagACMAOPageSize   int64
)

var (
	acmAOCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Create an authorized orgs description",
		Args: cobra.ExactArgs(1), RunE: runACMAOCreate,
	}
	acmAODeleteCmd = &cobra.Command{
		Use: "delete NAME", Short: "Delete an authorized orgs description",
		Args: cobra.ExactArgs(1), RunE: runACMAODelete,
	}
	acmAODescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Describe an authorized orgs description",
		Args: cobra.ExactArgs(1), RunE: runACMAODescribe,
	}
	acmAOListCmd = &cobra.Command{
		Use: "list", Short: "List authorized orgs descriptions",
		Args: cobra.NoArgs, RunE: runACMAOList,
	}
	acmAOUpdateCmd = &cobra.Command{
		Use: "update NAME", Short: "Update an authorized orgs description",
		Args: cobra.ExactArgs(1), RunE: runACMAOUpdate,
	}
)

func init() {
	all := []*cobra.Command{acmAOCreateCmd, acmAODeleteCmd, acmAODescribeCmd, acmAOListCmd, acmAOUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagACMAOFormat, "format", "", "Output format")
		c.Flags().StringVar(&flagACMAOPolicy, "policy", "", "Access policy ID (required)")
		_ = c.MarkFlagRequired("policy")
	}
	for _, c := range []*cobra.Command{acmAOCreateCmd, acmAOUpdateCmd} {
		c.Flags().StringVar(&flagACMAOConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the AuthorizedOrgsDesc body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	acmAOListCmd.Flags().Int64Var(&flagACMAOPageSize, "page-size", 0, "Maximum results per page")

	acmAOCmd.AddCommand(all...)
	accessContextManagerCmd.AddCommand(acmAOCmd)
}

func acmAOResource(policy, name string) string {
	return fmt.Sprintf("%s/authorizedOrgsDescs/%s", acmPolicyResource(policy), name)
}

func runACMAOCreate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.AuthorizedOrgsDesc{}
	if err := loadYAMLOrJSONInto(flagACMAOConfigFile, body); err != nil {
		return err
	}
	body.Name = acmAOResource(flagACMAOPolicy, args[0])
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.AuthorizedOrgsDescs.Create(acmPolicyResource(flagACMAOPolicy), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating authorized orgs desc: %w", err)
	}
	fmt.Printf("Create request issued for authorized orgs desc [%s].\n", args[0])
	return emitFormatted(op, flagACMAOFormat)
}

func runACMAODelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.AuthorizedOrgsDescs.Delete(acmAOResource(flagACMAOPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting authorized orgs desc: %w", err)
	}
	fmt.Printf("Delete request issued for authorized orgs desc [%s].\n", args[0])
	return emitFormatted(op, flagACMAOFormat)
}

func runACMAODescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.AccessPolicies.AuthorizedOrgsDescs.Get(acmAOResource(flagACMAOPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing authorized orgs desc: %w", err)
	}
	return emitFormatted(got, flagACMAOFormat)
}

func runACMAOList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*accesscontextmanager.AuthorizedOrgsDesc
	pageToken := ""
	for {
		call := svc.AccessPolicies.AuthorizedOrgsDescs.List(acmPolicyResource(flagACMAOPolicy)).Context(ctx)
		if flagACMAOPageSize > 0 {
			call = call.PageSize(flagACMAOPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing authorized orgs descs: %w", err)
		}
		all = append(all, resp.AuthorizedOrgsDescs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagACMAOFormat)
}

func runACMAOUpdate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.AuthorizedOrgsDesc{}
	if err := loadYAMLOrJSONInto(flagACMAOConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.AccessPolicies.AuthorizedOrgsDescs.Patch(acmAOResource(flagACMAOPolicy, args[0]), body).Context(ctx)
	if mask := joinMask(nonEmptyJSONFields(body)); mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating authorized orgs desc: %w", err)
	}
	fmt.Printf("Update request issued for authorized orgs desc [%s].\n", args[0])
	return emitFormatted(op, flagACMAOFormat)
}
