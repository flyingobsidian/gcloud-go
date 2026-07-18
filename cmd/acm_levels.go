package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager levels (#1443) ---

var acmLVCmd = &cobra.Command{Use: "levels", Short: "Manage access levels"}

var (
	flagACMLVFormat     string
	flagACMLVPolicy     string
	flagACMLVConfigFile string
	flagACMLVSourceFile string
	flagACMLVPageSize   int64
	flagACMLVEtag       string
)

var (
	acmLVCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Create an access level",
		Args: cobra.ExactArgs(1), RunE: runACMLVCreate,
	}
	acmLVDeleteCmd = &cobra.Command{
		Use: "delete NAME", Short: "Delete an access level",
		Args: cobra.ExactArgs(1), RunE: runACMLVDelete,
	}
	acmLVDescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Describe an access level",
		Args: cobra.ExactArgs(1), RunE: runACMLVDescribe,
	}
	acmLVListCmd = &cobra.Command{
		Use: "list", Short: "List access levels",
		Args: cobra.NoArgs, RunE: runACMLVList,
	}
	acmLVUpdateCmd = &cobra.Command{
		Use: "update NAME", Short: "Update an access level",
		Args: cobra.ExactArgs(1), RunE: runACMLVUpdate,
	}
	acmLVReplaceAllCmd = &cobra.Command{
		Use: "replace-all", Short: "Replace all access levels in a policy",
		Args: cobra.NoArgs, RunE: runACMLVReplaceAll,
	}
)

var acmLVCondCmd = &cobra.Command{Use: "conditions", Short: "Manage conditions on an access level"}

var (
	acmLVCondCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Add a condition to an access level",
		Args: cobra.ExactArgs(1), RunE: runACMLVCondCreate,
	}
	acmLVCondDeleteCmd = &cobra.Command{
		Use: "delete NAME", Short: "Remove a condition from an access level",
		Args: cobra.ExactArgs(1), RunE: runACMLVCondDelete,
	}
	acmLVCondListCmd = &cobra.Command{
		Use: "list NAME", Short: "List conditions on an access level",
		Args: cobra.ExactArgs(1), RunE: runACMLVCondList,
	}
)

var (
	flagACMLVCondConfigFile string
	flagACMLVCondIndex      int
)

func init() {
	all := []*cobra.Command{
		acmLVCreateCmd, acmLVDeleteCmd, acmLVDescribeCmd, acmLVListCmd, acmLVUpdateCmd, acmLVReplaceAllCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagACMLVFormat, "format", "", "Output format")
		c.Flags().StringVar(&flagACMLVPolicy, "policy", "", "Access policy ID (required)")
		_ = c.MarkFlagRequired("policy")
	}
	for _, c := range []*cobra.Command{acmLVCreateCmd, acmLVUpdateCmd} {
		c.Flags().StringVar(&flagACMLVConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the AccessLevel body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	acmLVReplaceAllCmd.Flags().StringVar(&flagACMLVSourceFile, "source-file", "",
		"Path to a YAML/JSON file containing a list of AccessLevel bodies (required)")
	_ = acmLVReplaceAllCmd.MarkFlagRequired("source-file")
	acmLVReplaceAllCmd.Flags().StringVar(&flagACMLVEtag, "etag", "", "Expected etag of the parent AccessPolicy")
	acmLVListCmd.Flags().Int64Var(&flagACMLVPageSize, "page-size", 0, "Maximum results per page")

	condAll := []*cobra.Command{acmLVCondCreateCmd, acmLVCondDeleteCmd, acmLVCondListCmd}
	for _, c := range condAll {
		c.Flags().StringVar(&flagACMLVFormat, "format", "", "Output format")
		c.Flags().StringVar(&flagACMLVPolicy, "policy", "", "Access policy ID (required)")
		_ = c.MarkFlagRequired("policy")
	}
	acmLVCondCreateCmd.Flags().StringVar(&flagACMLVCondConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Condition body (required)")
	_ = acmLVCondCreateCmd.MarkFlagRequired("config-file")
	acmLVCondDeleteCmd.Flags().IntVar(&flagACMLVCondIndex, "index", -1,
		"Zero-based index of the condition to remove (required)")
	_ = acmLVCondDeleteCmd.MarkFlagRequired("index")

	acmLVCondCmd.AddCommand(condAll...)
	acmLVCmd.AddCommand(all...)
	acmLVCmd.AddCommand(acmLVCondCmd)
	accessContextManagerCmd.AddCommand(acmLVCmd)
}

func acmLVResource(policy, name string) string {
	return fmt.Sprintf("%s/accessLevels/%s", acmPolicyResource(policy), name)
}

func runACMLVCreate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.AccessLevel{}
	if err := loadYAMLOrJSONInto(flagACMLVConfigFile, body); err != nil {
		return err
	}
	body.Name = acmLVResource(flagACMLVPolicy, args[0])
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.AccessLevels.Create(acmPolicyResource(flagACMLVPolicy), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating access level: %w", err)
	}
	fmt.Printf("Create request issued for access level [%s].\n", args[0])
	return emitFormatted(op, flagACMLVFormat)
}

func runACMLVDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.AccessLevels.Delete(acmLVResource(flagACMLVPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting access level: %w", err)
	}
	fmt.Printf("Delete request issued for access level [%s].\n", args[0])
	return emitFormatted(op, flagACMLVFormat)
}

func runACMLVDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.AccessPolicies.AccessLevels.Get(acmLVResource(flagACMLVPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing access level: %w", err)
	}
	return emitFormatted(got, flagACMLVFormat)
}

func runACMLVList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*accesscontextmanager.AccessLevel
	pageToken := ""
	for {
		call := svc.AccessPolicies.AccessLevels.List(acmPolicyResource(flagACMLVPolicy)).Context(ctx)
		if flagACMLVPageSize > 0 {
			call = call.PageSize(flagACMLVPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing access levels: %w", err)
		}
		all = append(all, resp.AccessLevels...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagACMLVFormat)
}

func runACMLVUpdate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.AccessLevel{}
	if err := loadYAMLOrJSONInto(flagACMLVConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.AccessPolicies.AccessLevels.Patch(acmLVResource(flagACMLVPolicy, args[0]), body).Context(ctx)
	if mask := joinMask(nonEmptyJSONFields(body)); mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating access level: %w", err)
	}
	fmt.Printf("Update request issued for access level [%s].\n", args[0])
	return emitFormatted(op, flagACMLVFormat)
}

func runACMLVReplaceAll(cmd *cobra.Command, args []string) error {
	var levels []*accesscontextmanager.AccessLevel
	if err := loadYAMLOrJSONInto(flagACMLVSourceFile, &levels); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &accesscontextmanager.ReplaceAccessLevelsRequest{
		AccessLevels: levels,
		Etag:         flagACMLVEtag,
	}
	op, err := svc.AccessPolicies.AccessLevels.ReplaceAll(acmPolicyResource(flagACMLVPolicy), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("replacing access levels: %w", err)
	}
	fmt.Println("Replace-all request issued for access levels.")
	return emitFormatted(op, flagACMLVFormat)
}

func acmLVFetchForPatch(ctx context.Context, svc *accesscontextmanager.Service, policy, name string) (*accesscontextmanager.AccessLevel, error) {
	level, err := svc.AccessPolicies.AccessLevels.Get(acmLVResource(policy, name)).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("getting access level: %w", err)
	}
	if level.Basic == nil {
		level.Basic = &accesscontextmanager.BasicLevel{}
	}
	return level, nil
}

func runACMLVCondCreate(cmd *cobra.Command, args []string) error {
	cond := &accesscontextmanager.Condition{}
	if err := loadYAMLOrJSONInto(flagACMLVCondConfigFile, cond); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	level, err := acmLVFetchForPatch(ctx, svc, flagACMLVPolicy, args[0])
	if err != nil {
		return err
	}
	level.Basic.Conditions = append(level.Basic.Conditions, cond)
	patch := &accesscontextmanager.AccessLevel{Basic: level.Basic}
	op, err := svc.AccessPolicies.AccessLevels.Patch(acmLVResource(flagACMLVPolicy, args[0]), patch).
		UpdateMask("basic").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding condition: %w", err)
	}
	fmt.Printf("Added condition to access level [%s].\n", args[0])
	return emitFormatted(op, flagACMLVFormat)
}

func runACMLVCondDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	level, err := acmLVFetchForPatch(ctx, svc, flagACMLVPolicy, args[0])
	if err != nil {
		return err
	}
	conds := level.Basic.Conditions
	if flagACMLVCondIndex < 0 || flagACMLVCondIndex >= len(conds) {
		return fmt.Errorf("condition index %d out of range [0, %d)", flagACMLVCondIndex, len(conds))
	}
	level.Basic.Conditions = append(conds[:flagACMLVCondIndex], conds[flagACMLVCondIndex+1:]...)
	patch := &accesscontextmanager.AccessLevel{Basic: level.Basic}
	op, err := svc.AccessPolicies.AccessLevels.Patch(acmLVResource(flagACMLVPolicy, args[0]), patch).
		UpdateMask("basic").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing condition: %w", err)
	}
	fmt.Printf("Removed condition [%d] from access level [%s].\n", flagACMLVCondIndex, args[0])
	return emitFormatted(op, flagACMLVFormat)
}

func runACMLVCondList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	level, err := svc.AccessPolicies.AccessLevels.Get(acmLVResource(flagACMLVPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting access level: %w", err)
	}
	var conds []*accesscontextmanager.Condition
	if level.Basic != nil {
		conds = level.Basic.Conditions
	}
	return emitFormatted(conds, flagACMLVFormat)
}
