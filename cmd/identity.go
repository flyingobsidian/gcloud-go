package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudidentity "google.golang.org/api/cloudidentity/v1"
)

// --- gcloud identity (#346) ---

var identityCmd = &cobra.Command{Use: "identity", Short: "Manage Cloud Identity"}

var (
	flagIDGroupEmail        string
	flagIDMemberEmail       string
	flagIDGroupNamespace    string
	flagIDMemberNamespace   string
	flagIDRoles             string
	flagIDAddRoles          string
	flagIDRemoveRoles       string
	flagIDMemberType        string
	flagIDView              string
	flagIDQuery             string
	flagIDPageSize          int64
	flagIDFormat            string
	flagIDDeliverySetting   string
)

func idResolveGroupName(ctx context.Context, svc *cloudidentity.Service, emailOrName, namespace string) (string, error) {
	if emailOrName == "" {
		return "", fmt.Errorf("--group-email is required")
	}
	if strings.HasPrefix(emailOrName, "groups/") {
		return emailOrName, nil
	}
	call := svc.Groups.Lookup().GroupKeyId(emailOrName).Context(ctx)
	if namespace != "" {
		call = call.GroupKeyNamespace(namespace)
	}
	resp, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("looking up group %q: %w", emailOrName, err)
	}
	return resp.Name, nil
}

func idResolveMembershipName(ctx context.Context, svc *cloudidentity.Service, groupName, memberEmail, namespace string) (string, error) {
	if memberEmail == "" {
		return "", fmt.Errorf("--member-email is required")
	}
	if strings.HasPrefix(memberEmail, "groups/") && strings.Contains(memberEmail, "/memberships/") {
		return memberEmail, nil
	}
	call := svc.Groups.Memberships.Lookup(groupName).MemberKeyId(memberEmail).Context(ctx)
	if namespace != "" {
		call = call.MemberKeyNamespace(namespace)
	}
	resp, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("looking up membership for %q in %q: %w", memberEmail, groupName, err)
	}
	return resp.Name, nil
}

func idParseRoles(spec string) []*cloudidentity.MembershipRole {
	if spec == "" {
		return []*cloudidentity.MembershipRole{{Name: "MEMBER"}}
	}
	var roles []*cloudidentity.MembershipRole
	for _, r := range strings.Split(spec, ",") {
		r = strings.ToUpper(strings.TrimSpace(r))
		if r == "" {
			continue
		}
		roles = append(roles, &cloudidentity.MembershipRole{Name: r})
	}
	if len(roles) == 0 {
		roles = []*cloudidentity.MembershipRole{{Name: "MEMBER"}}
	}
	return roles
}

// --- groups CRUD (out of scope; stubs remain) ---

func idInitGroupsStubs(groups *cobra.Command) {
	for _, n := range []string{"create", "delete", "describe", "list", "update", "search", "get-iam-policy", "set-iam-policy"} {
		registerStubCommand(groups, n, "Not yet implemented")
	}
	registerStubGroup(groups, "preview", "Preview commands", "list")
}

// --- memberships (#849) ---

var identityMembershipsCmd = &cobra.Command{Use: "memberships", Short: "Manage Cloud Identity group memberships"}

var (
	idMembAddCmd = &cobra.Command{
		Use: "add", Short: "Add a member to a Cloud Identity group",
		Args: cobra.NoArgs, RunE: runIDMembAdd,
	}
	idMembDeleteCmd = &cobra.Command{
		Use: "delete", Short: "Delete a member from a Cloud Identity group",
		Args: cobra.NoArgs, RunE: runIDMembDelete,
	}
	idMembDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe a Cloud Identity group membership",
		Args: cobra.NoArgs, RunE: runIDMembDescribe,
	}
	idMembListCmd = &cobra.Command{
		Use: "list", Short: "List memberships of a Cloud Identity group",
		Args: cobra.NoArgs, RunE: runIDMembList,
	}
	idMembCheckTransitiveCmd = &cobra.Command{
		Use: "check-transitive-membership", Short: "Check whether a member has a transitive membership in a group",
		Args: cobra.NoArgs, RunE: runIDMembCheckTransitive,
	}
	idMembGetGraphCmd = &cobra.Command{
		Use: "get-membership-graph", Short: "Get the membership graph of a group",
		Args: cobra.NoArgs, RunE: runIDMembGetGraph,
	}
	idMembModifyRolesCmd = &cobra.Command{
		Use: "modify-membership-roles", Short: "Modify roles of a Cloud Identity group membership",
		Args: cobra.NoArgs, RunE: runIDMembModifyRoles,
	}
	idMembSearchTransitiveGroupsCmd = &cobra.Command{
		Use: "search-transitive-groups", Short: "Search transitive groups a member belongs to",
		Args: cobra.NoArgs, RunE: runIDMembSearchTransitiveGroups,
	}
	idMembSearchTransitiveMembershipsCmd = &cobra.Command{
		Use: "search-transitive-memberships", Short: "Search transitive memberships of a group",
		Args: cobra.NoArgs, RunE: runIDMembSearchTransitiveMemberships,
	}
)

func init() {
	groups := &cobra.Command{Use: "groups", Short: "Manage Cloud Identity Groups"}
	idInitGroupsStubs(groups)

	addGroupFlag := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagIDGroupEmail, "group-email", "",
				"Email or resource name (groups/{id}) of the Cloud Identity group")
			c.Flags().StringVar(&flagIDGroupNamespace, "group-namespace", "",
				"Optional namespace for the group key (for external identities)")
		}
	}
	addMemberFlag := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagIDMemberEmail, "member-email", "",
				"Email or full membership resource name of the member")
			c.Flags().StringVar(&flagIDMemberNamespace, "member-namespace", "",
				"Optional namespace for the member key (for external identities)")
		}
	}
	addFormatFlag := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagIDFormat, "format", "", "Output format")
		}
	}

	// add
	addGroupFlag(idMembAddCmd)
	addMemberFlag(idMembAddCmd)
	addFormatFlag(idMembAddCmd)
	idMembAddCmd.Flags().StringVar(&flagIDRoles, "roles", "MEMBER",
		"Comma-separated list of roles (OWNER, MANAGER, MEMBER)")
	idMembAddCmd.Flags().StringVar(&flagIDMemberType, "member-type", "",
		"Optional membership type (USER, SERVICE_ACCOUNT, GROUP, SHARED_DRIVE, OTHER)")
	idMembAddCmd.Flags().StringVar(&flagIDDeliverySetting, "delivery-setting", "",
		"Optional delivery setting (ALL_MAIL, DIGEST, DAILY, NONE, DISABLED)")
	_ = idMembAddCmd.MarkFlagRequired("group-email")
	_ = idMembAddCmd.MarkFlagRequired("member-email")

	// delete
	addGroupFlag(idMembDeleteCmd)
	addMemberFlag(idMembDeleteCmd)
	addFormatFlag(idMembDeleteCmd)
	_ = idMembDeleteCmd.MarkFlagRequired("group-email")
	_ = idMembDeleteCmd.MarkFlagRequired("member-email")

	// describe
	addGroupFlag(idMembDescribeCmd)
	addMemberFlag(idMembDescribeCmd)
	addFormatFlag(idMembDescribeCmd)
	_ = idMembDescribeCmd.MarkFlagRequired("group-email")
	_ = idMembDescribeCmd.MarkFlagRequired("member-email")

	// list
	addGroupFlag(idMembListCmd)
	addFormatFlag(idMembListCmd)
	idMembListCmd.Flags().StringVar(&flagIDView, "view", "",
		"Optional view (BASIC or FULL)")
	idMembListCmd.Flags().Int64Var(&flagIDPageSize, "page-size", 0, "Server page size hint")
	_ = idMembListCmd.MarkFlagRequired("group-email")

	// check-transitive-membership
	addGroupFlag(idMembCheckTransitiveCmd)
	addMemberFlag(idMembCheckTransitiveCmd)
	addFormatFlag(idMembCheckTransitiveCmd)
	_ = idMembCheckTransitiveCmd.MarkFlagRequired("group-email")
	_ = idMembCheckTransitiveCmd.MarkFlagRequired("member-email")

	// get-membership-graph
	addGroupFlag(idMembGetGraphCmd)
	addFormatFlag(idMembGetGraphCmd)
	idMembGetGraphCmd.Flags().StringVar(&flagIDQuery, "query", "",
		"CEL filter over Membership metadata (see Cloud Identity Groups API docs)")
	_ = idMembGetGraphCmd.MarkFlagRequired("group-email")

	// modify-membership-roles
	addGroupFlag(idMembModifyRolesCmd)
	addMemberFlag(idMembModifyRolesCmd)
	addFormatFlag(idMembModifyRolesCmd)
	idMembModifyRolesCmd.Flags().StringVar(&flagIDAddRoles, "add-roles", "",
		"Comma-separated list of roles to add (OWNER, MANAGER, MEMBER)")
	idMembModifyRolesCmd.Flags().StringVar(&flagIDRemoveRoles, "remove-roles", "",
		"Comma-separated list of roles to remove (OWNER, MANAGER)")
	_ = idMembModifyRolesCmd.MarkFlagRequired("group-email")
	_ = idMembModifyRolesCmd.MarkFlagRequired("member-email")

	// search-transitive-groups
	addMemberFlag(idMembSearchTransitiveGroupsCmd)
	addFormatFlag(idMembSearchTransitiveGroupsCmd)
	idMembSearchTransitiveGroupsCmd.Flags().StringVar(&flagIDQuery, "query", "",
		"CEL filter over the searched groups (see Cloud Identity Groups API docs)")
	idMembSearchTransitiveGroupsCmd.Flags().Int64Var(&flagIDPageSize, "page-size", 0, "Server page size hint")

	// search-transitive-memberships
	addGroupFlag(idMembSearchTransitiveMembershipsCmd)
	addFormatFlag(idMembSearchTransitiveMembershipsCmd)
	idMembSearchTransitiveMembershipsCmd.Flags().Int64Var(&flagIDPageSize, "page-size", 0, "Server page size hint")
	_ = idMembSearchTransitiveMembershipsCmd.MarkFlagRequired("group-email")

	identityMembershipsCmd.AddCommand(
		idMembAddCmd,
		idMembCheckTransitiveCmd,
		idMembDeleteCmd,
		idMembDescribeCmd,
		idMembGetGraphCmd,
		idMembListCmd,
		idMembModifyRolesCmd,
		idMembSearchTransitiveGroupsCmd,
		idMembSearchTransitiveMembershipsCmd,
	)
	groups.AddCommand(identityMembershipsCmd)

	identityCmd.AddCommand(groups)
	rootCmd.AddCommand(identityCmd)
}

// --- memberships impl ---

func runIDMembAdd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, flagIDGroupEmail, flagIDGroupNamespace)
	if err != nil {
		return err
	}
	ns := flagIDMemberNamespace
	if ns == "" {
		ns = flagIDGroupNamespace
	}
	memb := &cloudidentity.Membership{
		PreferredMemberKey: &cloudidentity.EntityKey{Id: flagIDMemberEmail, Namespace: ns},
		Roles:              idParseRoles(flagIDRoles),
		Type:               flagIDMemberType,
		DeliverySetting:    flagIDDeliverySetting,
	}
	op, err := svc.Groups.Memberships.Create(groupName, memb).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding membership: %w", err)
	}
	return emitFormatted(op, flagIDFormat)
}

func runIDMembDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, flagIDGroupEmail, flagIDGroupNamespace)
	if err != nil {
		return err
	}
	name, err := idResolveMembershipName(ctx, svc, groupName, flagIDMemberEmail, flagIDMemberNamespace)
	if err != nil {
		return err
	}
	op, err := svc.Groups.Memberships.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting membership: %w", err)
	}
	return emitFormatted(op, flagIDFormat)
}

func runIDMembDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, flagIDGroupEmail, flagIDGroupNamespace)
	if err != nil {
		return err
	}
	name, err := idResolveMembershipName(ctx, svc, groupName, flagIDMemberEmail, flagIDMemberNamespace)
	if err != nil {
		return err
	}
	got, err := svc.Groups.Memberships.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing membership: %w", err)
	}
	return emitFormatted(got, flagIDFormat)
}

func runIDMembList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, flagIDGroupEmail, flagIDGroupNamespace)
	if err != nil {
		return err
	}
	var all []*cloudidentity.Membership
	pageToken := ""
	for {
		call := svc.Groups.Memberships.List(groupName).Context(ctx)
		if flagIDView != "" {
			call = call.View(strings.ToUpper(flagIDView))
		}
		if flagIDPageSize > 0 {
			call = call.PageSize(flagIDPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing memberships: %w", err)
		}
		all = append(all, resp.Memberships...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagIDFormat != "" {
		return emitFormatted(all, flagIDFormat)
	}
	fmt.Printf("%-40s %-40s %s\n", "MEMBERSHIP", "MEMBER", "ROLES")
	for _, m := range all {
		var roles []string
		for _, r := range m.Roles {
			roles = append(roles, r.Name)
		}
		id := ""
		if m.PreferredMemberKey != nil {
			id = m.PreferredMemberKey.Id
		}
		fmt.Printf("%-40s %-40s %s\n", path.Base(m.Name), id, strings.Join(roles, ","))
	}
	return nil
}

func runIDMembCheckTransitive(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, flagIDGroupEmail, flagIDGroupNamespace)
	if err != nil {
		return err
	}
	q := fmt.Sprintf("member_key_id == '%s'", flagIDMemberEmail)
	ns := flagIDMemberNamespace
	if ns == "" {
		ns = flagIDGroupNamespace
	}
	if ns != "" {
		q += fmt.Sprintf(" && member_key_namespace == '%s'", ns)
	}
	resp, err := svc.Groups.Memberships.CheckTransitiveMembership(groupName).Query(q).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("checking transitive membership: %w", err)
	}
	return emitFormatted(resp, flagIDFormat)
}

func runIDMembGetGraph(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, flagIDGroupEmail, flagIDGroupNamespace)
	if err != nil {
		return err
	}
	call := svc.Groups.Memberships.GetMembershipGraph(groupName).Context(ctx)
	if flagIDQuery != "" {
		call = call.Query(flagIDQuery)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("getting membership graph: %w", err)
	}
	return emitFormatted(op, flagIDFormat)
}

func runIDMembModifyRoles(cmd *cobra.Command, args []string) error {
	if flagIDAddRoles == "" && flagIDRemoveRoles == "" {
		return fmt.Errorf("at least one of --add-roles or --remove-roles is required")
	}
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, flagIDGroupEmail, flagIDGroupNamespace)
	if err != nil {
		return err
	}
	name, err := idResolveMembershipName(ctx, svc, groupName, flagIDMemberEmail, flagIDMemberNamespace)
	if err != nil {
		return err
	}
	req := &cloudidentity.ModifyMembershipRolesRequest{}
	if flagIDAddRoles != "" {
		req.AddRoles = idParseRoles(flagIDAddRoles)
	}
	if flagIDRemoveRoles != "" {
		for _, r := range strings.Split(flagIDRemoveRoles, ",") {
			r = strings.ToUpper(strings.TrimSpace(r))
			if r != "" {
				req.RemoveRoles = append(req.RemoveRoles, r)
			}
		}
	}
	resp, err := svc.Groups.Memberships.ModifyMembershipRoles(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("modifying membership roles: %w", err)
	}
	return emitFormatted(resp, flagIDFormat)
}

func runIDMembSearchTransitiveGroups(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	q := flagIDQuery
	if q == "" {
		if flagIDMemberEmail == "" {
			return fmt.Errorf("either --query or --member-email is required")
		}
		q = fmt.Sprintf("member_key_id == '%s'", flagIDMemberEmail)
		ns := flagIDMemberNamespace
		if ns != "" {
			q += fmt.Sprintf(" && member_key_namespace == '%s'", ns)
		}
		q += " && 'cloudidentity.googleapis.com/groups.discussion_forum' in labels"
	}
	var all []*cloudidentity.GroupRelation
	pageToken := ""
	for {
		call := svc.Groups.Memberships.SearchTransitiveGroups("groups/-").Query(q).Context(ctx)
		if flagIDPageSize > 0 {
			call = call.PageSize(flagIDPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("searching transitive groups: %w", err)
		}
		all = append(all, resp.Memberships...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagIDFormat)
}

func runIDMembSearchTransitiveMemberships(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, flagIDGroupEmail, flagIDGroupNamespace)
	if err != nil {
		return err
	}
	var all []*cloudidentity.MemberRelation
	pageToken := ""
	for {
		call := svc.Groups.Memberships.SearchTransitiveMemberships(groupName).Context(ctx)
		if flagIDPageSize > 0 {
			call = call.PageSize(flagIDPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("searching transitive memberships: %w", err)
		}
		all = append(all, resp.Memberships...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagIDFormat)
}
