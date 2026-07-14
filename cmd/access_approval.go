package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	accessapproval "google.golang.org/api/accessapproval/v1"
)

// --- gcloud access-approval (#287, #566) ---

var accessApprovalCmd = &cobra.Command{Use: "access-approval", Short: "Manage Access Approval"}

var (
	flagAAFolder       string
	flagAAOrganization string
	flagAAConfigFile   string
	flagAAUpdateMask   string
	flagAAFormat       string
	flagAAApproveExp   string
	flagAADismissReason string
	flagAAFilter       string
)

// --- requests ---

var accessApprovalRequestsCmd = &cobra.Command{Use: "requests", Short: "Manage access approval requests"}

var (
	aaReqApproveCmd = &cobra.Command{
		Use: "approve REQUEST_ID", Short: "Approve an access approval request",
		Args: cobra.ExactArgs(1), RunE: runAAReqApprove,
	}
	aaReqDismissCmd = &cobra.Command{
		Use: "dismiss REQUEST_ID", Short: "Dismiss an access approval request",
		Args: cobra.ExactArgs(1), RunE: runAAReqDismiss,
	}
	aaReqGetCmd = &cobra.Command{
		Use: "get REQUEST_ID", Short: "Get an access approval request",
		Args: cobra.ExactArgs(1), RunE: runAAReqGet,
	}
	aaReqInvalidateCmd = &cobra.Command{
		Use: "invalidate REQUEST_ID", Short: "Invalidate an approved access approval request",
		Args: cobra.ExactArgs(1), RunE: runAAReqInvalidate,
	}
	aaReqListCmd = &cobra.Command{
		Use: "list", Short: "List access approval requests",
		Args: cobra.NoArgs, RunE: runAAReqList,
	}
)

// --- service-account ---

var accessApprovalServiceAccountCmd = &cobra.Command{Use: "service-account", Short: "Get the Access Approval service account"}

var aaSAGetCmd = &cobra.Command{
	Use: "get", Short: "Get the Access Approval service account",
	Args: cobra.NoArgs, RunE: runAASAGet,
}

// --- settings ---

var accessApprovalSettingsCmd = &cobra.Command{Use: "settings", Short: "Manage Access Approval settings"}

var (
	aaSetDeleteCmd = &cobra.Command{
		Use: "delete", Short: "Delete Access Approval settings",
		Args: cobra.NoArgs, RunE: runAASetDelete,
	}
	aaSetDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe Access Approval settings",
		Args: cobra.NoArgs, RunE: runAASetDescribe,
	}
	aaSetUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update Access Approval settings from a --config-file",
		Args: cobra.NoArgs, RunE: runAASetUpdate,
	}
)

func init() {
	// requests
	reqAll := []*cobra.Command{aaReqApproveCmd, aaReqDismissCmd, aaReqGetCmd, aaReqInvalidateCmd, aaReqListCmd}
	for _, c := range reqAll {
		addAAScopeFlags(c)
	}
	aaReqApproveCmd.Flags().StringVar(&flagAAApproveExp, "expire-time", "",
		"Expire time for the approval (RFC 3339)")
	aaReqDismissCmd.Flags().StringVar(&flagAADismissReason, "reason", "",
		"Reason for dismissal")
	aaReqListCmd.Flags().StringVar(&flagAAFilter, "state", "",
		"Filter by request state (PENDING, APPROVED, DISMISSED, HISTORY)")
	for _, c := range []*cobra.Command{aaReqGetCmd, aaReqListCmd} {
		c.Flags().StringVar(&flagAAFormat, "format", "", "Output format")
	}
	accessApprovalRequestsCmd.AddCommand(reqAll...)
	accessApprovalCmd.AddCommand(accessApprovalRequestsCmd)

	// service-account
	addAAScopeFlags(aaSAGetCmd)
	aaSAGetCmd.Flags().StringVar(&flagAAFormat, "format", "", "Output format")
	accessApprovalServiceAccountCmd.AddCommand(aaSAGetCmd)
	accessApprovalCmd.AddCommand(accessApprovalServiceAccountCmd)

	// settings
	setAll := []*cobra.Command{aaSetDeleteCmd, aaSetDescribeCmd, aaSetUpdateCmd}
	for _, c := range setAll {
		addAAScopeFlags(c)
	}
	aaSetUpdateCmd.Flags().StringVar(&flagAAConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the AccessApprovalSettings body (required)")
	_ = aaSetUpdateCmd.MarkFlagRequired("config-file")
	aaSetUpdateCmd.Flags().StringVar(&flagAAUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	aaSetDescribeCmd.Flags().StringVar(&flagAAFormat, "format", "", "Output format")
	accessApprovalSettingsCmd.AddCommand(setAll...)
	accessApprovalCmd.AddCommand(accessApprovalSettingsCmd)

	rootCmd.AddCommand(accessApprovalCmd)
}

func addAAScopeFlags(c *cobra.Command) {
	c.Flags().StringVar(&flagAAFolder, "folder", "", "Folder scope (alternative to --project/--organization)")
	c.Flags().StringVar(&flagAAOrganization, "organization", "", "Organization scope (alternative to --project/--folder)")
}

// aaScopeParent returns "projects/PROJECT", "folders/FOLDER", or "organizations/ORG"
// based on which flag is set. Falls back to the resolved project.
func aaScopeParent() (string, error) {
	set := 0
	if flagProject != "" {
		set++
	}
	if flagAAFolder != "" {
		set++
	}
	if flagAAOrganization != "" {
		set++
	}
	if set > 1 {
		return "", fmt.Errorf("only one of --project, --folder, --organization may be set")
	}
	if flagAAFolder != "" {
		return "folders/" + flagAAFolder, nil
	}
	if flagAAOrganization != "" {
		return "organizations/" + flagAAOrganization, nil
	}
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

func aaResourceName(scope, resource string) string {
	return fmt.Sprintf("%s/%s", scope, resource)
}

// --- requests impl ---

func runAAReqApprove(cmd *cobra.Command, args []string) error {
	scope, err := aaScopeParent()
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/approvalRequests/%s", scope, args[0])
	req := &accessapproval.ApproveApprovalRequestMessage{ExpireTime: flagAAApproveExp}
	ctx := context.Background()
	svc, err := gcp.AccessApprovalService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var got *accessapproval.ApprovalRequest
	switch {
	case flagAAFolder != "":
		got, err = svc.Folders.ApprovalRequests.Approve(name, req).Context(ctx).Do()
	case flagAAOrganization != "":
		got, err = svc.Organizations.ApprovalRequests.Approve(name, req).Context(ctx).Do()
	default:
		got, err = svc.Projects.ApprovalRequests.Approve(name, req).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("approving request: %w", err)
	}
	return emitFormatted(got, "")
}

func runAAReqDismiss(cmd *cobra.Command, args []string) error {
	scope, err := aaScopeParent()
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/approvalRequests/%s", scope, args[0])
	req := &accessapproval.DismissApprovalRequestMessage{}
	ctx := context.Background()
	svc, err := gcp.AccessApprovalService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var got *accessapproval.ApprovalRequest
	switch {
	case flagAAFolder != "":
		got, err = svc.Folders.ApprovalRequests.Dismiss(name, req).Context(ctx).Do()
	case flagAAOrganization != "":
		got, err = svc.Organizations.ApprovalRequests.Dismiss(name, req).Context(ctx).Do()
	default:
		got, err = svc.Projects.ApprovalRequests.Dismiss(name, req).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("dismissing request: %w", err)
	}
	return emitFormatted(got, "")
}

func runAAReqGet(cmd *cobra.Command, args []string) error {
	scope, err := aaScopeParent()
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/approvalRequests/%s", scope, args[0])
	ctx := context.Background()
	svc, err := gcp.AccessApprovalService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var got *accessapproval.ApprovalRequest
	switch {
	case flagAAFolder != "":
		got, err = svc.Folders.ApprovalRequests.Get(name).Context(ctx).Do()
	case flagAAOrganization != "":
		got, err = svc.Organizations.ApprovalRequests.Get(name).Context(ctx).Do()
	default:
		got, err = svc.Projects.ApprovalRequests.Get(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("getting request: %w", err)
	}
	return emitFormatted(got, flagAAFormat)
}

func runAAReqInvalidate(cmd *cobra.Command, args []string) error {
	scope, err := aaScopeParent()
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/approvalRequests/%s", scope, args[0])
	req := &accessapproval.InvalidateApprovalRequestMessage{}
	ctx := context.Background()
	svc, err := gcp.AccessApprovalService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var got *accessapproval.ApprovalRequest
	switch {
	case flagAAFolder != "":
		got, err = svc.Folders.ApprovalRequests.Invalidate(name, req).Context(ctx).Do()
	case flagAAOrganization != "":
		got, err = svc.Organizations.ApprovalRequests.Invalidate(name, req).Context(ctx).Do()
	default:
		got, err = svc.Projects.ApprovalRequests.Invalidate(name, req).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("invalidating request: %w", err)
	}
	return emitFormatted(got, "")
}

func runAAReqList(cmd *cobra.Command, args []string) error {
	scope, err := aaScopeParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AccessApprovalService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*accessapproval.ApprovalRequest
	pageToken := ""
	for {
		var (
			list []*accessapproval.ApprovalRequest
			next string
			err  error
		)
		switch {
		case flagAAFolder != "":
			call := svc.Folders.ApprovalRequests.List(scope).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			if flagAAFilter != "" {
				call = call.Filter(flagAAFilter)
			}
			resp, e := call.Do()
			list, next, err = resp.ApprovalRequests, resp.NextPageToken, e
		case flagAAOrganization != "":
			call := svc.Organizations.ApprovalRequests.List(scope).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			if flagAAFilter != "" {
				call = call.Filter(flagAAFilter)
			}
			resp, e := call.Do()
			list, next, err = resp.ApprovalRequests, resp.NextPageToken, e
		default:
			call := svc.Projects.ApprovalRequests.List(scope).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			if flagAAFilter != "" {
				call = call.Filter(flagAAFilter)
			}
			resp, e := call.Do()
			list, next, err = resp.ApprovalRequests, resp.NextPageToken, e
		}
		if err != nil {
			return fmt.Errorf("listing requests: %w", err)
		}
		all = append(all, list...)
		if next == "" {
			break
		}
		pageToken = next
	}
	if flagAAFormat != "" {
		return emitFormatted(all, flagAAFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "REQUEST_TIME")
	for _, r := range all {
		fmt.Printf("%-40s %s\n", path.Base(r.Name), r.RequestTime)
	}
	return nil
}

// --- service-account impl ---

func runAASAGet(cmd *cobra.Command, args []string) error {
	scope, err := aaScopeParent()
	if err != nil {
		return err
	}
	name := aaResourceName(scope, "serviceAccount")
	ctx := context.Background()
	svc, err := gcp.AccessApprovalService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var got *accessapproval.AccessApprovalServiceAccount
	switch {
	case flagAAFolder != "":
		got, err = svc.Folders.GetServiceAccount(name).Context(ctx).Do()
	case flagAAOrganization != "":
		got, err = svc.Organizations.GetServiceAccount(name).Context(ctx).Do()
	default:
		got, err = svc.Projects.GetServiceAccount(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("getting service account: %w", err)
	}
	return emitFormatted(got, flagAAFormat)
}

// --- settings impl ---

func aaSettingsName(scope string) string {
	return aaResourceName(scope, "accessApprovalSettings")
}

func runAASetDelete(cmd *cobra.Command, args []string) error {
	scope, err := aaScopeParent()
	if err != nil {
		return err
	}
	name := aaSettingsName(scope)
	ctx := context.Background()
	svc, err := gcp.AccessApprovalService(ctx, flagAccount)
	if err != nil {
		return err
	}
	switch {
	case flagAAFolder != "":
		_, err = svc.Folders.DeleteAccessApprovalSettings(name).Context(ctx).Do()
	case flagAAOrganization != "":
		_, err = svc.Organizations.DeleteAccessApprovalSettings(name).Context(ctx).Do()
	default:
		_, err = svc.Projects.DeleteAccessApprovalSettings(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("deleting settings: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Deleted Access Approval settings for %s.\n", scope)
	return nil
}

func runAASetDescribe(cmd *cobra.Command, args []string) error {
	scope, err := aaScopeParent()
	if err != nil {
		return err
	}
	name := aaSettingsName(scope)
	ctx := context.Background()
	svc, err := gcp.AccessApprovalService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var got *accessapproval.AccessApprovalSettings
	switch {
	case flagAAFolder != "":
		got, err = svc.Folders.GetAccessApprovalSettings(name).Context(ctx).Do()
	case flagAAOrganization != "":
		got, err = svc.Organizations.GetAccessApprovalSettings(name).Context(ctx).Do()
	default:
		got, err = svc.Projects.GetAccessApprovalSettings(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("describing settings: %w", err)
	}
	return emitFormatted(got, flagAAFormat)
}

func runAASetUpdate(cmd *cobra.Command, args []string) error {
	scope, err := aaScopeParent()
	if err != nil {
		return err
	}
	body := &accessapproval.AccessApprovalSettings{}
	if err := loadYAMLOrJSONInto(flagAAConfigFile, body); err != nil {
		return err
	}
	mask := flagAAUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := aaSettingsName(scope)
	ctx := context.Background()
	svc, err := gcp.AccessApprovalService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var got *accessapproval.AccessApprovalSettings
	switch {
	case flagAAFolder != "":
		got, err = svc.Folders.UpdateAccessApprovalSettings(name, body).UpdateMask(mask).Context(ctx).Do()
	case flagAAOrganization != "":
		got, err = svc.Organizations.UpdateAccessApprovalSettings(name, body).UpdateMask(mask).Context(ctx).Do()
	default:
		got, err = svc.Projects.UpdateAccessApprovalSettings(name, body).UpdateMask(mask).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("updating settings: %w", err)
	}
	return emitFormatted(got, "")
}
