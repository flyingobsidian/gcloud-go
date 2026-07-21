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

// --- groups CRUD (#1717) ---

var (
	flagIDGroupCustomer    string
	flagIDGroupDisplayName string
	flagIDGroupDescription string
	flagIDGroupLabels      []string
	flagIDGroupInitConfig  string
	flagIDGroupUpdateMask  string
	flagIDGroupPageSize    int64
	flagIDGroupSearchView  string
)

var (
	idGroupsCreateCmd = &cobra.Command{
		Use: "create GROUP_EMAIL", Short: "Create a Cloud Identity group",
		Args: cobra.ExactArgs(1), RunE: runIDGroupsCreate,
	}
	idGroupsDeleteCmd = &cobra.Command{
		Use: "delete GROUP_EMAIL", Short: "Delete a Cloud Identity group",
		Args: cobra.ExactArgs(1), RunE: runIDGroupsDelete,
	}
	idGroupsDescribeCmd = &cobra.Command{
		Use: "describe GROUP_EMAIL", Short: "Describe a Cloud Identity group",
		Args: cobra.ExactArgs(1), RunE: runIDGroupsDescribe,
	}
	idGroupsSearchCmd = &cobra.Command{
		Use: "search", Short: "Search for Cloud Identity groups matching a query",
		Args: cobra.NoArgs, RunE: runIDGroupsSearch,
	}
	idGroupsUpdateCmd = &cobra.Command{
		Use: "update GROUP_EMAIL", Short: "Update a Cloud Identity group",
		Args: cobra.ExactArgs(1), RunE: runIDGroupsUpdate,
	}
	idGroupsPreviewCmd = &cobra.Command{
		Use: "preview", Short: "Preview users in a customer account using a CEL query",
		Args: cobra.NoArgs, RunE: runIDGroupsPreview,
	}
)

// --- groups config subgroup (#1719) ---

var (
	flagIDGroupConfigPath   string
	flagIDGroupConfigFormat string
)

var (
	idGroupsConfigCmd = &cobra.Command{Use: "config", Short: "Manage Cloud Identity group configurations"}

	idGroupsConfigExportCmd = &cobra.Command{
		Use: "export GROUP_EMAIL", Short: "Export a Cloud Identity group configuration",
		Args: cobra.ExactArgs(1), RunE: runIDGroupsConfigExport,
	}
)

func idInitGroupsCommands(groups *cobra.Command) {
	addFmt := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagIDFormat, "format", "", "Output format")
		}
	}

	// create
	addFmt(idGroupsCreateCmd)
	idGroupsCreateCmd.Flags().StringVar(&flagIDGroupCustomer, "customer", "",
		"Customer ID (e.g. Cxxxxxxx or 'customers/Cxxxxxxx') that owns the group (required)")
	_ = idGroupsCreateCmd.MarkFlagRequired("customer")
	idGroupsCreateCmd.Flags().StringVar(&flagIDGroupNamespace, "group-namespace", "",
		"Optional namespace for the group key (for external identities)")
	idGroupsCreateCmd.Flags().StringVar(&flagIDGroupDisplayName, "display-name", "", "Display name of the group")
	idGroupsCreateCmd.Flags().StringVar(&flagIDGroupDescription, "description", "", "Description of the group")
	idGroupsCreateCmd.Flags().StringSliceVar(&flagIDGroupLabels, "labels", nil,
		"Comma-separated KEY=VALUE labels; the default label 'cloudidentity.googleapis.com/groups.discussion_forum=' is applied when empty")
	idGroupsCreateCmd.Flags().StringVar(&flagIDGroupInitConfig, "initial-group-config", "",
		"Initial group configuration (empty, with-initial-owner)")

	// delete
	addFmt(idGroupsDeleteCmd)
	idGroupsDeleteCmd.Flags().StringVar(&flagIDGroupNamespace, "group-namespace", "",
		"Optional namespace for the group key")

	// describe
	addFmt(idGroupsDescribeCmd)
	idGroupsDescribeCmd.Flags().StringVar(&flagIDGroupNamespace, "group-namespace", "",
		"Optional namespace for the group key")

	// search
	addFmt(idGroupsSearchCmd)
	idGroupsSearchCmd.Flags().StringVar(&flagIDQuery, "query", "",
		"CEL query (required); e.g. \"parent == 'customers/Cxxxxxxx' && 'cloudidentity.googleapis.com/groups.discussion_forum' in labels\"")
	_ = idGroupsSearchCmd.MarkFlagRequired("query")
	idGroupsSearchCmd.Flags().StringVar(&flagIDGroupSearchView, "view", "",
		"Search view (BASIC or FULL)")
	idGroupsSearchCmd.Flags().Int64Var(&flagIDGroupPageSize, "page-size", 0, "Server page size hint")

	// update
	addFmt(idGroupsUpdateCmd)
	idGroupsUpdateCmd.Flags().StringVar(&flagIDGroupNamespace, "group-namespace", "",
		"Optional namespace for the group key")
	idGroupsUpdateCmd.Flags().StringVar(&flagIDGroupDisplayName, "display-name", "", "Update display name")
	idGroupsUpdateCmd.Flags().StringVar(&flagIDGroupDescription, "description", "", "Update description")
	idGroupsUpdateCmd.Flags().StringSliceVar(&flagIDGroupLabels, "labels", nil, "Replace labels (KEY=VALUE list)")
	idGroupsUpdateCmd.Flags().StringVar(&flagIDGroupUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (displayName, description, labels)")

	// preview (leaf; not a group)
	addFmt(idGroupsPreviewCmd)
	idGroupsPreviewCmd.Flags().StringVar(&flagIDGroupCustomer, "customer", "",
		"Customer ID (Cxxxxxxx) whose users should be previewed (required)")
	_ = idGroupsPreviewCmd.MarkFlagRequired("customer")
	idGroupsPreviewCmd.Flags().StringVar(&flagIDQuery, "query", "",
		"CEL query filtering the returned users (e.g. \"user.locations.exists(loc, loc.desk_code == 'abc')\")")
	idGroupsPreviewCmd.Flags().StringVar(&flagIDView, "view", "",
		"Search view (BASIC or FULL)")
	idGroupsPreviewCmd.Flags().Int64Var(&flagIDGroupPageSize, "page-size", 100, "Server page size hint")

	// config export (subgroup)
	addFmt(idGroupsConfigExportCmd)
	idGroupsConfigExportCmd.Flags().StringVar(&flagIDGroupNamespace, "group-namespace", "",
		"Optional namespace for the group key")
	idGroupsConfigExportCmd.Flags().StringVar(&flagIDGroupConfigPath, "path", "-",
		"Path of the directory or file to output configuration(s). Use '-' for stdout.")
	idGroupsConfigExportCmd.Flags().StringVar(&flagIDGroupConfigFormat, "resource-format", "krm",
		"Configuration format (krm, terraform)")

	groups.AddCommand(
		idGroupsCreateCmd,
		idGroupsDeleteCmd,
		idGroupsDescribeCmd,
		idGroupsPreviewCmd,
		idGroupsSearchCmd,
		idGroupsUpdateCmd,
	)
	idGroupsConfigCmd.AddCommand(idGroupsConfigExportCmd)
	groups.AddCommand(idGroupsConfigCmd)
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
	idInitGroupsCommands(groups)

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

// --- groups CRUD impl (#1717, #1718, #1719) ---

func idNormalizeCustomer(c string) string {
	if c == "" {
		return ""
	}
	if strings.HasPrefix(c, "customers/") {
		return c
	}
	return "customers/" + c
}

func idParseKVLabels(specs []string) (map[string]string, error) {
	if len(specs) == 0 {
		return nil, nil
	}
	out := map[string]string{}
	for _, kv := range specs {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			out[strings.TrimSpace(kv)] = ""
			continue
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return out, nil
}

func runIDGroupsCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	labels, err := idParseKVLabels(flagIDGroupLabels)
	if err != nil {
		return err
	}
	if labels == nil {
		labels = map[string]string{"cloudidentity.googleapis.com/groups.discussion_forum": ""}
	}
	body := &cloudidentity.Group{
		Parent:      idNormalizeCustomer(flagIDGroupCustomer),
		GroupKey:    &cloudidentity.EntityKey{Id: args[0], Namespace: flagIDGroupNamespace},
		DisplayName: flagIDGroupDisplayName,
		Description: flagIDGroupDescription,
		Labels:      labels,
	}
	call := svc.Groups.Create(body).Context(ctx)
	switch strings.ToLower(flagIDGroupInitConfig) {
	case "":
	case "empty":
		call = call.InitialGroupConfig("EMPTY")
	case "with-initial-owner":
		call = call.InitialGroupConfig("WITH_INITIAL_OWNER")
	default:
		return fmt.Errorf("invalid --initial-group-config %q (want empty or with-initial-owner)", flagIDGroupInitConfig)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating group: %w", err)
	}
	return emitFormatted(op, flagIDFormat)
}

func runIDGroupsDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name, err := idResolveGroupName(ctx, svc, args[0], flagIDGroupNamespace)
	if err != nil {
		return err
	}
	op, err := svc.Groups.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting group: %w", err)
	}
	return emitFormatted(op, flagIDFormat)
}

func runIDGroupsDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name, err := idResolveGroupName(ctx, svc, args[0], flagIDGroupNamespace)
	if err != nil {
		return err
	}
	got, err := svc.Groups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing group: %w", err)
	}
	return emitFormatted(got, flagIDFormat)
}

func runIDGroupsSearch(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*cloudidentity.Group
	pageToken := ""
	for {
		call := svc.Groups.Search().Query(flagIDQuery).Context(ctx)
		if flagIDGroupSearchView != "" {
			call = call.View(strings.ToUpper(flagIDGroupSearchView))
		}
		if flagIDGroupPageSize > 0 {
			call = call.PageSize(flagIDGroupPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("searching groups: %w", err)
		}
		all = append(all, resp.Groups...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagIDFormat != "" {
		return emitFormatted(all, flagIDFormat)
	}
	fmt.Printf("%-40s %-40s %s\n", "NAME", "GROUP_KEY", "DISPLAY_NAME")
	for _, g := range all {
		key := ""
		if g.GroupKey != nil {
			key = g.GroupKey.Id
		}
		fmt.Printf("%-40s %-40s %s\n", path.Base(g.Name), key, g.DisplayName)
	}
	return nil
}

func runIDGroupsUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name, err := idResolveGroupName(ctx, svc, args[0], flagIDGroupNamespace)
	if err != nil {
		return err
	}
	body := &cloudidentity.Group{}
	mask := flagIDGroupUpdateMask
	if flagIDGroupDisplayName != "" || strings.Contains(mask, "displayName") {
		body.DisplayName = flagIDGroupDisplayName
	}
	if flagIDGroupDescription != "" || strings.Contains(mask, "description") {
		body.Description = flagIDGroupDescription
	}
	if len(flagIDGroupLabels) > 0 || strings.Contains(mask, "labels") {
		labels, err := idParseKVLabels(flagIDGroupLabels)
		if err != nil {
			return err
		}
		body.Labels = labels
	}
	if mask == "" {
		var parts []string
		if body.DisplayName != "" {
			parts = append(parts, "displayName")
		}
		if body.Description != "" {
			parts = append(parts, "description")
		}
		if body.Labels != nil {
			parts = append(parts, "labels")
		}
		if len(parts) == 0 {
			return fmt.Errorf("at least one of --display-name, --description, or --labels is required")
		}
		mask = strings.Join(parts, ",")
	}
	op, err := svc.Groups.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating group: %w", err)
	}
	return emitFormatted(op, flagIDFormat)
}

func runIDGroupsPreview(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	customer := idNormalizeCustomer(flagIDGroupCustomer)
	q := flagIDQuery
	if q == "" {
		q = fmt.Sprintf("parent == '%s'", customer)
	}
	var all []*cloudidentity.Group
	pageToken := ""
	for {
		call := svc.Groups.Search().Query(q).Context(ctx)
		if flagIDView != "" {
			call = call.View(strings.ToUpper(flagIDView))
		}
		if flagIDGroupPageSize > 0 {
			call = call.PageSize(flagIDGroupPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("previewing groups: %w", err)
		}
		all = append(all, resp.Groups...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagIDFormat)
}

func runIDGroupsConfigExport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name, err := idResolveGroupName(ctx, svc, args[0], flagIDGroupNamespace)
	if err != nil {
		return err
	}
	g, err := svc.Groups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching group for export: %w", err)
	}
	format := strings.ToLower(flagIDGroupConfigFormat)
	if format == "" {
		format = "krm"
	}
	if format != "krm" && format != "terraform" {
		return fmt.Errorf("--resource-format must be one of: krm, terraform")
	}
	if format == "terraform" {
		return fmt.Errorf("terraform export is not yet implemented; use --resource-format=krm")
	}
	out := map[string]any{
		"apiVersion": "resourcemanager.cnrm.cloud.google.com/v1beta1",
		"kind":       "Group",
		"metadata": map[string]any{
			"name":        path.Base(g.Name),
			"annotations": map[string]any{"cnrm.cloud.google.com/project-id": strings.TrimPrefix(g.Parent, "customers/")},
		},
		"spec": map[string]any{
			"displayName": g.DisplayName,
			"description": g.Description,
			"groupKey":    g.GroupKey,
			"labels":      g.Labels,
			"parent":      g.Parent,
		},
	}
	return emitFormatted(out, "yaml")
}
