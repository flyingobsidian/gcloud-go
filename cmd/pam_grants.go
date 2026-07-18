package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud pam grants (#963) ---

var pamGrantsCmd = &cobra.Command{Use: "grants", Short: "Manage Privileged Access Manager grants"}

var (
	flagPAMGrantLocation    string
	flagPAMGrantEntitlement string
	flagPAMGrantFormat      string
	flagPAMGrantConfigFile  string
	flagPAMGrantReason      string
	flagPAMGrantCallerRel   string
	flagPAMGrantPageSize    int64
)

var (
	pamGrantApproveCmd = &cobra.Command{
		Use: "approve GRANT", Short: "Approve a Privileged Access Manager grant",
		Args: cobra.ExactArgs(1), RunE: runPAMGrantApprove,
	}
	pamGrantCreateCmd = &cobra.Command{
		Use: "create GRANT", Short: "Create a Privileged Access Manager grant",
		Args: cobra.ExactArgs(1), RunE: runPAMGrantCreate,
	}
	pamGrantDenyCmd = &cobra.Command{
		Use: "deny GRANT", Short: "Deny a Privileged Access Manager grant",
		Args: cobra.ExactArgs(1), RunE: runPAMGrantDeny,
	}
	pamGrantDescribeCmd = &cobra.Command{
		Use: "describe GRANT", Short: "Describe a Privileged Access Manager grant",
		Args: cobra.ExactArgs(1), RunE: runPAMGrantDescribe,
	}
	pamGrantListCmd = &cobra.Command{
		Use: "list", Short: "List Privileged Access Manager grants",
		Args: cobra.NoArgs, RunE: runPAMGrantList,
	}
	pamGrantRevokeCmd = &cobra.Command{
		Use: "revoke GRANT", Short: "Revoke a Privileged Access Manager grant",
		Args: cobra.ExactArgs(1), RunE: runPAMGrantRevoke,
	}
	pamGrantSearchCmd = &cobra.Command{
		Use: "search", Short: "Search Privileged Access Manager grants for the caller",
		Args: cobra.NoArgs, RunE: runPAMGrantSearch,
	}
)

func init() {
	all := []*cobra.Command{
		pamGrantApproveCmd, pamGrantCreateCmd, pamGrantDenyCmd, pamGrantDescribeCmd,
		pamGrantListCmd, pamGrantRevokeCmd, pamGrantSearchCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagPAMGrantLocation, "location", "",
			"Location of the parent entitlement (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagPAMGrantFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{pamGrantCreateCmd, pamGrantListCmd, pamGrantSearchCmd} {
		c.Flags().StringVar(&flagPAMGrantEntitlement, "entitlement", "",
			"Parent entitlement ID (required)")
		_ = c.MarkFlagRequired("entitlement")
	}
	// GRANT-taking commands also accept --entitlement so a bare grant ID can be
	// combined with the location + entitlement to form the full resource path.
	for _, c := range []*cobra.Command{
		pamGrantApproveCmd, pamGrantDenyCmd, pamGrantDescribeCmd, pamGrantRevokeCmd,
	} {
		c.Flags().StringVar(&flagPAMGrantEntitlement, "entitlement", "",
			"Parent entitlement ID (required unless a full grant resource name is passed)")
	}
	pamGrantCreateCmd.Flags().StringVar(&flagPAMGrantConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the grant body (required)")
	_ = pamGrantCreateCmd.MarkFlagRequired("config-file")

	for _, c := range []*cobra.Command{pamGrantApproveCmd, pamGrantDenyCmd, pamGrantRevokeCmd} {
		c.Flags().StringVar(&flagPAMGrantReason, "reason", "",
			"Optional reason to record on the request")
	}
	pamGrantSearchCmd.Flags().StringVar(&flagPAMGrantCallerRel, "caller-relationship", "HAD_CREATED",
		"Caller relationship: HAD_CREATED, CAN_APPROVE, or HAD_APPROVED")
	pamGrantListCmd.Flags().Int64Var(&flagPAMGrantPageSize, "page-size", 0, "Maximum results per page")
	pamGrantSearchCmd.Flags().Int64Var(&flagPAMGrantPageSize, "page-size", 0, "Maximum results per page")

	pamGrantsCmd.AddCommand(all...)
	pamCmd.AddCommand(pamGrantsCmd)
}

func pamGrantEntitlementParent() (string, error) {
	if strings.HasPrefix(flagPAMGrantEntitlement, "projects/") ||
		strings.HasPrefix(flagPAMGrantEntitlement, "folders/") ||
		strings.HasPrefix(flagPAMGrantEntitlement, "organizations/") {
		return flagPAMGrantEntitlement, nil
	}
	if flagPAMGrantEntitlement == "" {
		return "", fmt.Errorf("--entitlement is required")
	}
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s/entitlements/%s", project, flagPAMGrantLocation, flagPAMGrantEntitlement), nil
}

func pamGrantName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") ||
		strings.HasPrefix(id, "folders/") ||
		strings.HasPrefix(id, "organizations/") {
		return id, nil
	}
	parent, err := pamGrantEntitlementParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/grants/%s", parent, id), nil
}

func pamGrantReasonBody() map[string]any {
	body := map[string]any{}
	if flagPAMGrantReason != "" {
		body["reason"] = flagPAMGrantReason
	}
	return body
}

func runPAMGrantApprove(cmd *cobra.Command, args []string) error {
	name, err := pamGrantName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := pamRest.do(ctx, http.MethodPost, "/"+name+":approve", nil, pamGrantReasonBody(), &got); err != nil {
		return fmt.Errorf("approving grant: %w", err)
	}
	fmt.Printf("Approved grant [%s].\n", args[0])
	return emitFormatted(got, flagPAMGrantFormat)
}

func runPAMGrantCreate(cmd *cobra.Command, args []string) error {
	parent, err := pamGrantEntitlementParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagPAMGrantConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("grantId", args[0])
	ctx := context.Background()
	var got map[string]any
	if err := pamRest.do(ctx, http.MethodPost, "/"+parent+"/grants", q, body, &got); err != nil {
		return fmt.Errorf("creating grant: %w", err)
	}
	fmt.Printf("Created grant [%s].\n", args[0])
	return emitFormatted(got, flagPAMGrantFormat)
}

func runPAMGrantDeny(cmd *cobra.Command, args []string) error {
	name, err := pamGrantName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := pamRest.do(ctx, http.MethodPost, "/"+name+":deny", nil, pamGrantReasonBody(), &got); err != nil {
		return fmt.Errorf("denying grant: %w", err)
	}
	fmt.Printf("Denied grant [%s].\n", args[0])
	return emitFormatted(got, flagPAMGrantFormat)
}

func runPAMGrantDescribe(cmd *cobra.Command, args []string) error {
	name, err := pamGrantName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := pamRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing grant: %w", err)
	}
	return emitFormatted(got, flagPAMGrantFormat)
}

func runPAMGrantList(cmd *cobra.Command, args []string) error {
	parent, err := pamGrantEntitlementParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := pamRest.paginate(ctx, "/"+parent+"/grants", nil, "grants", flagPAMGrantPageSize)
	if err != nil {
		return fmt.Errorf("listing grants: %w", err)
	}
	return emitFormatted(items, flagPAMGrantFormat)
}

func runPAMGrantRevoke(cmd *cobra.Command, args []string) error {
	name, err := pamGrantName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := pamRest.do(ctx, http.MethodPost, "/"+name+":revoke", nil, pamGrantReasonBody(), &got); err != nil {
		return fmt.Errorf("revoking grant: %w", err)
	}
	fmt.Printf("Revoked grant [%s].\n", args[0])
	return emitFormatted(got, flagPAMGrantFormat)
}

func runPAMGrantSearch(cmd *cobra.Command, args []string) error {
	parent, err := pamGrantEntitlementParent()
	if err != nil {
		return err
	}
	rel := flagPAMGrantCallerRel
	if rel == "" {
		rel = "HAD_CREATED"
	}
	ctx := context.Background()
	base := url.Values{}
	base.Set("callerRelationship", rel)
	items, err := pamRest.paginate(ctx, "/"+parent+":searchGrants", base, "grants", flagPAMGrantPageSize)
	if err != nil {
		return fmt.Errorf("searching grants: %w", err)
	}
	return emitFormatted(items, flagPAMGrantFormat)
}
