package cli

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	accessapproval "google.golang.org/api/accessapproval/v1"
)

// --- gcloud access-approval (#287, #566) ---

// --- requests ---

// --- service-account ---

// --- settings ---

func (s *accessApprovalCmds) initAccessApproval() {
	// requests
	reqAll := []*cobra.Command{s.aaReqApproveCmd, s.aaReqDismissCmd, s.aaReqGetCmd, s.aaReqInvalidateCmd, s.aaReqListCmd}
	for _, c := range reqAll {
		s.addAAScopeFlags(c)
	}
	s.aaReqApproveCmd.Flags().StringVar(&s.flagAAApproveExp, "expire-time", "",
		"Expire time for the approval (RFC 3339)")
	s.aaReqDismissCmd.Flags().StringVar(&s.flagAADismissReason, "reason", "",
		"Reason for dismissal")
	s.aaReqListCmd.Flags().StringVar(&s.flagAAFilter, "state", "",
		"Filter by request state (PENDING, APPROVED, DISMISSED, HISTORY)")
	for _, c := range []*cobra.Command{s.aaReqGetCmd, s.aaReqListCmd} {
		c.Flags().StringVar(&s.flagAAFormat, "format", "", "Output format")
	}
	s.accessApprovalRequestsCmd.AddCommand(reqAll...)
	s.accessApprovalCmd.AddCommand(s.accessApprovalRequestsCmd)

	// service-account
	s.addAAScopeFlags(s.aaSAGetCmd)
	s.aaSAGetCmd.Flags().StringVar(&s.flagAAFormat, "format", "", "Output format")
	s.accessApprovalServiceAccountCmd.AddCommand(s.aaSAGetCmd)
	s.accessApprovalCmd.AddCommand(s.accessApprovalServiceAccountCmd)

	// settings
	setAll := []*cobra.Command{s.aaSetDeleteCmd, s.aaSetDescribeCmd, s.aaSetUpdateCmd}
	for _, c := range setAll {
		s.addAAScopeFlags(c)
	}
	s.aaSetUpdateCmd.Flags().StringVar(&s.flagAAConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the AccessApprovalSettings body (required)")
	_ = s.aaSetUpdateCmd.MarkFlagRequired("config-file")
	s.aaSetUpdateCmd.Flags().StringVar(&s.flagAAUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.aaSetDescribeCmd.Flags().StringVar(&s.flagAAFormat, "format", "", "Output format")
	s.accessApprovalSettingsCmd.AddCommand(setAll...)
	s.accessApprovalCmd.AddCommand(s.accessApprovalSettingsCmd)

}

func (s *accessApprovalCmds) addAAScopeFlags(c *cobra.Command) {
	c.Flags().StringVar(&s.flagAAFolder, "folder", "", "Folder scope (alternative to --project/--organization)")
	c.Flags().StringVar(&s.flagAAOrganization, "organization", "", "Organization scope (alternative to --project/--folder)")
}

// aaScopeParent returns "projects/PROJECT", "folders/FOLDER", or "organizations/ORG"
// based on which flag is set. Falls back to the resolved project.
func (s *accessApprovalCmds) aaScopeParent() (string, error) {
	set := 0
	if s.inv.flags.project != "" {
		set++
	}
	if s.flagAAFolder != "" {
		set++
	}
	if s.flagAAOrganization != "" {
		set++
	}
	if set > 1 {
		return "", fmt.Errorf("only one of --project, --folder, --organization may be set")
	}
	if s.flagAAFolder != "" {
		return "folders/" + s.flagAAFolder, nil
	}
	if s.flagAAOrganization != "" {
		return "organizations/" + s.flagAAOrganization, nil
	}
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

func aaResourceName(scope, resource string) string {
	return fmt.Sprintf("%s/%s", scope, resource)
}

// --- requests impl ---

func (s *accessApprovalCmds) runAAReqApprove(cmd *cobra.Command, args []string) error {
	scope, err := s.aaScopeParent()
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/approvalRequests/%s", scope, args[0])
	req := &accessapproval.ApproveApprovalRequestMessage{ExpireTime: s.flagAAApproveExp}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessApproval(ctx)
	if err != nil {
		return err
	}
	var got *accessapproval.ApprovalRequest
	switch {
	case s.flagAAFolder != "":
		got, err = svc.Folders.ApprovalRequests.Approve(name, req).Context(ctx).Do()
	case s.flagAAOrganization != "":
		got, err = svc.Organizations.ApprovalRequests.Approve(name, req).Context(ctx).Do()
	default:
		got, err = svc.Projects.ApprovalRequests.Approve(name, req).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("approving request: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *accessApprovalCmds) runAAReqDismiss(cmd *cobra.Command, args []string) error {
	scope, err := s.aaScopeParent()
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/approvalRequests/%s", scope, args[0])
	req := &accessapproval.DismissApprovalRequestMessage{}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessApproval(ctx)
	if err != nil {
		return err
	}
	var got *accessapproval.ApprovalRequest
	switch {
	case s.flagAAFolder != "":
		got, err = svc.Folders.ApprovalRequests.Dismiss(name, req).Context(ctx).Do()
	case s.flagAAOrganization != "":
		got, err = svc.Organizations.ApprovalRequests.Dismiss(name, req).Context(ctx).Do()
	default:
		got, err = svc.Projects.ApprovalRequests.Dismiss(name, req).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("dismissing request: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *accessApprovalCmds) runAAReqGet(cmd *cobra.Command, args []string) error {
	scope, err := s.aaScopeParent()
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/approvalRequests/%s", scope, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessApproval(ctx)
	if err != nil {
		return err
	}
	var got *accessapproval.ApprovalRequest
	switch {
	case s.flagAAFolder != "":
		got, err = svc.Folders.ApprovalRequests.Get(name).Context(ctx).Do()
	case s.flagAAOrganization != "":
		got, err = svc.Organizations.ApprovalRequests.Get(name).Context(ctx).Do()
	default:
		got, err = svc.Projects.ApprovalRequests.Get(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("getting request: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAAFormat)
}

func (s *accessApprovalCmds) runAAReqInvalidate(cmd *cobra.Command, args []string) error {
	scope, err := s.aaScopeParent()
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/approvalRequests/%s", scope, args[0])
	req := &accessapproval.InvalidateApprovalRequestMessage{}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessApproval(ctx)
	if err != nil {
		return err
	}
	var got *accessapproval.ApprovalRequest
	switch {
	case s.flagAAFolder != "":
		got, err = svc.Folders.ApprovalRequests.Invalidate(name, req).Context(ctx).Do()
	case s.flagAAOrganization != "":
		got, err = svc.Organizations.ApprovalRequests.Invalidate(name, req).Context(ctx).Do()
	default:
		got, err = svc.Projects.ApprovalRequests.Invalidate(name, req).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("invalidating request: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *accessApprovalCmds) runAAReqList(cmd *cobra.Command, args []string) error {
	scope, err := s.aaScopeParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessApproval(ctx)
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
		case s.flagAAFolder != "":
			call := svc.Folders.ApprovalRequests.List(scope).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			if s.flagAAFilter != "" {
				call = call.Filter(s.flagAAFilter)
			}
			resp, e := call.Do()
			list, next, err = resp.ApprovalRequests, resp.NextPageToken, e
		case s.flagAAOrganization != "":
			call := svc.Organizations.ApprovalRequests.List(scope).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			if s.flagAAFilter != "" {
				call = call.Filter(s.flagAAFilter)
			}
			resp, e := call.Do()
			list, next, err = resp.ApprovalRequests, resp.NextPageToken, e
		default:
			call := svc.Projects.ApprovalRequests.List(scope).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			if s.flagAAFilter != "" {
				call = call.Filter(s.flagAAFilter)
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
	if s.flagAAFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagAAFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "REQUEST_TIME")
	for _, r := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(r.Name), r.RequestTime)
	}
	return nil
}

// --- service-account impl ---

func (s *accessApprovalCmds) runAASAGet(cmd *cobra.Command, args []string) error {
	scope, err := s.aaScopeParent()
	if err != nil {
		return err
	}
	name := aaResourceName(scope, "serviceAccount")
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessApproval(ctx)
	if err != nil {
		return err
	}
	var got *accessapproval.AccessApprovalServiceAccount
	switch {
	case s.flagAAFolder != "":
		got, err = svc.Folders.GetServiceAccount(name).Context(ctx).Do()
	case s.flagAAOrganization != "":
		got, err = svc.Organizations.GetServiceAccount(name).Context(ctx).Do()
	default:
		got, err = svc.Projects.GetServiceAccount(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("getting service account: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAAFormat)
}

// --- settings impl ---

func aaSettingsName(scope string) string {
	return aaResourceName(scope, "accessApprovalSettings")
}

func (s *accessApprovalCmds) runAASetDelete(cmd *cobra.Command, args []string) error {
	scope, err := s.aaScopeParent()
	if err != nil {
		return err
	}
	name := aaSettingsName(scope)
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessApproval(ctx)
	if err != nil {
		return err
	}
	switch {
	case s.flagAAFolder != "":
		_, err = svc.Folders.DeleteAccessApprovalSettings(name).Context(ctx).Do()
	case s.flagAAOrganization != "":
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

func (s *accessApprovalCmds) runAASetDescribe(cmd *cobra.Command, args []string) error {
	scope, err := s.aaScopeParent()
	if err != nil {
		return err
	}
	name := aaSettingsName(scope)
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessApproval(ctx)
	if err != nil {
		return err
	}
	var got *accessapproval.AccessApprovalSettings
	switch {
	case s.flagAAFolder != "":
		got, err = svc.Folders.GetAccessApprovalSettings(name).Context(ctx).Do()
	case s.flagAAOrganization != "":
		got, err = svc.Organizations.GetAccessApprovalSettings(name).Context(ctx).Do()
	default:
		got, err = svc.Projects.GetAccessApprovalSettings(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("describing settings: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAAFormat)
}

func (s *accessApprovalCmds) runAASetUpdate(cmd *cobra.Command, args []string) error {
	scope, err := s.aaScopeParent()
	if err != nil {
		return err
	}
	body := &accessapproval.AccessApprovalSettings{}
	if err := loadYAMLOrJSONInto(s.flagAAConfigFile, body); err != nil {
		return err
	}
	mask := s.flagAAUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := aaSettingsName(scope)
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessApproval(ctx)
	if err != nil {
		return err
	}
	var got *accessapproval.AccessApprovalSettings
	switch {
	case s.flagAAFolder != "":
		got, err = svc.Folders.UpdateAccessApprovalSettings(name, body).UpdateMask(mask).Context(ctx).Do()
	case s.flagAAOrganization != "":
		got, err = svc.Organizations.UpdateAccessApprovalSettings(name, body).UpdateMask(mask).Context(ctx).Do()
	default:
		got, err = svc.Projects.UpdateAccessApprovalSettings(name, body).UpdateMask(mask).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("updating settings: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// accessApprovalCmds holds the commands and flag values of one invocation's "accessApproval"
// command group, so invocations don't share state (#2070).
type accessApprovalCmds struct {
	inv *invocation

	accessApprovalCmd               *cobra.Command
	flagAAFolder                    string
	flagAAOrganization              string
	flagAAConfigFile                string
	flagAAUpdateMask                string
	flagAAFormat                    string
	flagAAApproveExp                string
	flagAADismissReason             string
	flagAAFilter                    string
	accessApprovalRequestsCmd       *cobra.Command
	aaReqApproveCmd                 *cobra.Command
	aaReqDismissCmd                 *cobra.Command
	aaReqGetCmd                     *cobra.Command
	aaReqInvalidateCmd              *cobra.Command
	aaReqListCmd                    *cobra.Command
	accessApprovalServiceAccountCmd *cobra.Command
	aaSAGetCmd                      *cobra.Command
	accessApprovalSettingsCmd       *cobra.Command
	aaSetDeleteCmd                  *cobra.Command
	aaSetDescribeCmd                *cobra.Command
	aaSetUpdateCmd                  *cobra.Command
}

// newAccessApprovalCmd returns the "accessApproval" command group for inv.
func newAccessApprovalCmd(inv *invocation) *cobra.Command {
	return newAccessApprovalCmds(inv).accessApprovalCmd
}

// newAccessApprovalCmds builds the "accessApproval" command group for inv and returns its state.
func newAccessApprovalCmds(inv *invocation) *accessApprovalCmds {
	s := &accessApprovalCmds{inv: inv}
	s.accessApprovalCmd = &cobra.Command{Use: "access-approval", Short: "Manage Access Approval"}
	s.accessApprovalRequestsCmd = &cobra.Command{Use: "requests", Short: "Manage access approval requests"}
	s.aaReqApproveCmd = &cobra.Command{
		Use: "approve REQUEST_ID", Short: "Approve an access approval request",
		Args: cobra.ExactArgs(1), RunE: s.runAAReqApprove,
	}
	s.aaReqDismissCmd = &cobra.Command{
		Use: "dismiss REQUEST_ID", Short: "Dismiss an access approval request",
		Args: cobra.ExactArgs(1), RunE: s.runAAReqDismiss,
	}
	s.aaReqGetCmd = &cobra.Command{
		Use: "get REQUEST_ID", Short: "Get an access approval request",
		Args: cobra.ExactArgs(1), RunE: s.runAAReqGet,
	}
	s.aaReqInvalidateCmd = &cobra.Command{
		Use: "invalidate REQUEST_ID", Short: "Invalidate an approved access approval request",
		Args: cobra.ExactArgs(1), RunE: s.runAAReqInvalidate,
	}
	s.aaReqListCmd = &cobra.Command{
		Use: "list", Short: "List access approval requests",
		Args: cobra.NoArgs, RunE: s.runAAReqList,
	}
	s.accessApprovalServiceAccountCmd = &cobra.Command{Use: "service-account", Short: "Get the Access Approval service account"}
	s.aaSAGetCmd = &cobra.Command{
		Use: "get", Short: "Get the Access Approval service account",
		Args: cobra.NoArgs, RunE: s.runAASAGet,
	}
	s.accessApprovalSettingsCmd = &cobra.Command{Use: "settings", Short: "Manage Access Approval settings"}
	s.aaSetDeleteCmd = &cobra.Command{
		Use: "delete", Short: "Delete Access Approval settings",
		Args: cobra.NoArgs, RunE: s.runAASetDelete,
	}
	s.aaSetDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe Access Approval settings",
		Args: cobra.NoArgs, RunE: s.runAASetDescribe,
	}
	s.aaSetUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update Access Approval settings from a --config-file",
		Args: cobra.NoArgs, RunE: s.runAASetUpdate,
	}
	s.initAccessApproval()
	return s
}
