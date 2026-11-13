package cli

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	cloudidentity "google.golang.org/api/cloudidentity/v1"
)

// --- gcloud identity (#346) ---

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

// --- groups config subgroup (#1719) ---

func (s *identityCmds) idInitGroupsCommands(groups *cobra.Command) {
	addFmt := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagIDFormat, "format", "", "Output format")
		}
	}

	// create
	addFmt(s.idGroupsCreateCmd)
	s.idGroupsCreateCmd.Flags().StringVar(&s.flagIDGroupCustomer, "customer", "",
		"Customer ID (e.g. Cxxxxxxx or 'customers/Cxxxxxxx') that owns the group (required)")
	_ = s.idGroupsCreateCmd.MarkFlagRequired("customer")
	s.idGroupsCreateCmd.Flags().StringVar(&s.flagIDGroupNamespace, "group-namespace", "",
		"Optional namespace for the group key (for external identities)")
	s.idGroupsCreateCmd.Flags().StringVar(&s.flagIDGroupDisplayName, "display-name", "", "Display name of the group")
	s.idGroupsCreateCmd.Flags().StringVar(&s.flagIDGroupDescription, "description", "", "Description of the group")
	s.idGroupsCreateCmd.Flags().StringSliceVar(&s.flagIDGroupLabels, "labels", nil,
		"Comma-separated KEY=VALUE labels; the default label 'cloudidentity.googleapis.com/groups.discussion_forum=' is applied when empty")
	s.idGroupsCreateCmd.Flags().StringVar(&s.flagIDGroupInitConfig, "initial-group-config", "",
		"Initial group configuration (empty, with-initial-owner)")

	// delete
	addFmt(s.idGroupsDeleteCmd)
	s.idGroupsDeleteCmd.Flags().StringVar(&s.flagIDGroupNamespace, "group-namespace", "",
		"Optional namespace for the group key")

	// describe
	addFmt(s.idGroupsDescribeCmd)
	s.idGroupsDescribeCmd.Flags().StringVar(&s.flagIDGroupNamespace, "group-namespace", "",
		"Optional namespace for the group key")

	// search
	addFmt(s.idGroupsSearchCmd)
	s.idGroupsSearchCmd.Flags().StringVar(&s.flagIDQuery, "query", "",
		"CEL query (required); e.g. \"parent == 'customers/Cxxxxxxx' && 'cloudidentity.googleapis.com/groups.discussion_forum' in labels\"")
	_ = s.idGroupsSearchCmd.MarkFlagRequired("query")
	s.idGroupsSearchCmd.Flags().StringVar(&s.flagIDGroupSearchView, "view", "",
		"Search view (BASIC or FULL)")
	s.idGroupsSearchCmd.Flags().Int64Var(&s.flagIDGroupPageSize, "page-size", 0, "Server page size hint")

	// update
	addFmt(s.idGroupsUpdateCmd)
	s.idGroupsUpdateCmd.Flags().StringVar(&s.flagIDGroupNamespace, "group-namespace", "",
		"Optional namespace for the group key")
	s.idGroupsUpdateCmd.Flags().StringVar(&s.flagIDGroupDisplayName, "display-name", "", "Update display name")
	s.idGroupsUpdateCmd.Flags().StringVar(&s.flagIDGroupDescription, "description", "", "Update description")
	s.idGroupsUpdateCmd.Flags().StringSliceVar(&s.flagIDGroupLabels, "labels", nil, "Replace labels (KEY=VALUE list)")
	s.idGroupsUpdateCmd.Flags().StringVar(&s.flagIDGroupUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (displayName, description, labels)")

	// preview (leaf; not a group)
	addFmt(s.idGroupsPreviewCmd)
	s.idGroupsPreviewCmd.Flags().StringVar(&s.flagIDGroupCustomer, "customer", "",
		"Customer ID (Cxxxxxxx) whose users should be previewed (required)")
	_ = s.idGroupsPreviewCmd.MarkFlagRequired("customer")
	s.idGroupsPreviewCmd.Flags().StringVar(&s.flagIDQuery, "query", "",
		"CEL query filtering the returned users (e.g. \"user.locations.exists(loc, loc.desk_code == 'abc')\")")
	s.idGroupsPreviewCmd.Flags().StringVar(&s.flagIDView, "view", "",
		"Search view (BASIC or FULL)")
	s.idGroupsPreviewCmd.Flags().Int64Var(&s.flagIDGroupPageSize, "page-size", 100, "Server page size hint")

	// config export (subgroup)
	addFmt(s.idGroupsConfigExportCmd)
	s.idGroupsConfigExportCmd.Flags().StringVar(&s.flagIDGroupNamespace, "group-namespace", "",
		"Optional namespace for the group key")
	s.idGroupsConfigExportCmd.Flags().StringVar(&s.flagIDGroupConfigPath, "path", "-",
		"Path of the directory or file to output configuration(s). Use '-' for stdout.")
	s.idGroupsConfigExportCmd.Flags().StringVar(&s.flagIDGroupConfigFormat, "resource-format", "krm",
		"Configuration format (krm, terraform)")

	groups.AddCommand(
		s.idGroupsCreateCmd,
		s.idGroupsDeleteCmd,
		s.idGroupsDescribeCmd,
		s.idGroupsPreviewCmd,
		s.idGroupsSearchCmd,
		s.idGroupsUpdateCmd,
	)
	s.idGroupsConfigCmd.AddCommand(s.idGroupsConfigExportCmd)
	groups.AddCommand(s.idGroupsConfigCmd)
}

// --- memberships (#849) ---

func (s *identityCmds) initIdentity() {
	groups := &cobra.Command{Use: "groups", Short: "Manage Cloud Identity Groups"}
	s.idInitGroupsCommands(groups)

	addGroupFlag := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagIDGroupEmail, "group-email", "",
				"Email or resource name (groups/{id}) of the Cloud Identity group")
			c.Flags().StringVar(&s.flagIDGroupNamespace, "group-namespace", "",
				"Optional namespace for the group key (for external identities)")
		}
	}
	addMemberFlag := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagIDMemberEmail, "member-email", "",
				"Email or full membership resource name of the member")
			c.Flags().StringVar(&s.flagIDMemberNamespace, "member-namespace", "",
				"Optional namespace for the member key (for external identities)")
		}
	}
	addFormatFlag := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagIDFormat, "format", "", "Output format")
		}
	}

	// add
	addGroupFlag(s.idMembAddCmd)
	addMemberFlag(s.idMembAddCmd)
	addFormatFlag(s.idMembAddCmd)
	s.idMembAddCmd.Flags().StringVar(&s.flagIDRoles, "roles", "MEMBER",
		"Comma-separated list of roles (OWNER, MANAGER, MEMBER)")
	s.idMembAddCmd.Flags().StringVar(&s.flagIDMemberType, "member-type", "",
		"Optional membership type (USER, SERVICE_ACCOUNT, GROUP, SHARED_DRIVE, OTHER)")
	s.idMembAddCmd.Flags().StringVar(&s.flagIDDeliverySetting, "delivery-setting", "",
		"Optional delivery setting (ALL_MAIL, DIGEST, DAILY, NONE, DISABLED)")
	_ = s.idMembAddCmd.MarkFlagRequired("group-email")
	_ = s.idMembAddCmd.MarkFlagRequired("member-email")

	// delete
	addGroupFlag(s.idMembDeleteCmd)
	addMemberFlag(s.idMembDeleteCmd)
	addFormatFlag(s.idMembDeleteCmd)
	_ = s.idMembDeleteCmd.MarkFlagRequired("group-email")
	_ = s.idMembDeleteCmd.MarkFlagRequired("member-email")

	// describe
	addGroupFlag(s.idMembDescribeCmd)
	addMemberFlag(s.idMembDescribeCmd)
	addFormatFlag(s.idMembDescribeCmd)
	_ = s.idMembDescribeCmd.MarkFlagRequired("group-email")
	_ = s.idMembDescribeCmd.MarkFlagRequired("member-email")

	// list
	addGroupFlag(s.idMembListCmd)
	addFormatFlag(s.idMembListCmd)
	s.idMembListCmd.Flags().StringVar(&s.flagIDView, "view", "",
		"Optional view (BASIC or FULL)")
	s.idMembListCmd.Flags().Int64Var(&s.flagIDPageSize, "page-size", 0, "Server page size hint")
	_ = s.idMembListCmd.MarkFlagRequired("group-email")

	// check-transitive-membership
	addGroupFlag(s.idMembCheckTransitiveCmd)
	addMemberFlag(s.idMembCheckTransitiveCmd)
	addFormatFlag(s.idMembCheckTransitiveCmd)
	_ = s.idMembCheckTransitiveCmd.MarkFlagRequired("group-email")
	_ = s.idMembCheckTransitiveCmd.MarkFlagRequired("member-email")

	// get-membership-graph
	addGroupFlag(s.idMembGetGraphCmd)
	addFormatFlag(s.idMembGetGraphCmd)
	s.idMembGetGraphCmd.Flags().StringVar(&s.flagIDQuery, "query", "",
		"CEL filter over Membership metadata (see Cloud Identity Groups API docs)")
	_ = s.idMembGetGraphCmd.MarkFlagRequired("group-email")

	// modify-membership-roles
	addGroupFlag(s.idMembModifyRolesCmd)
	addMemberFlag(s.idMembModifyRolesCmd)
	addFormatFlag(s.idMembModifyRolesCmd)
	s.idMembModifyRolesCmd.Flags().StringVar(&s.flagIDAddRoles, "add-roles", "",
		"Comma-separated list of roles to add (OWNER, MANAGER, MEMBER)")
	s.idMembModifyRolesCmd.Flags().StringVar(&s.flagIDRemoveRoles, "remove-roles", "",
		"Comma-separated list of roles to remove (OWNER, MANAGER)")
	_ = s.idMembModifyRolesCmd.MarkFlagRequired("group-email")
	_ = s.idMembModifyRolesCmd.MarkFlagRequired("member-email")

	// search-transitive-groups
	addMemberFlag(s.idMembSearchTransitiveGroupsCmd)
	addFormatFlag(s.idMembSearchTransitiveGroupsCmd)
	s.idMembSearchTransitiveGroupsCmd.Flags().StringVar(&s.flagIDQuery, "query", "",
		"CEL filter over the searched groups (see Cloud Identity Groups API docs)")
	s.idMembSearchTransitiveGroupsCmd.Flags().Int64Var(&s.flagIDPageSize, "page-size", 0, "Server page size hint")

	// search-transitive-memberships
	addGroupFlag(s.idMembSearchTransitiveMembershipsCmd)
	addFormatFlag(s.idMembSearchTransitiveMembershipsCmd)
	s.idMembSearchTransitiveMembershipsCmd.Flags().Int64Var(&s.flagIDPageSize, "page-size", 0, "Server page size hint")
	_ = s.idMembSearchTransitiveMembershipsCmd.MarkFlagRequired("group-email")

	s.identityMembershipsCmd.AddCommand(
		s.idMembAddCmd,
		s.idMembCheckTransitiveCmd,
		s.idMembDeleteCmd,
		s.idMembDescribeCmd,
		s.idMembGetGraphCmd,
		s.idMembListCmd,
		s.idMembModifyRolesCmd,
		s.idMembSearchTransitiveGroupsCmd,
		s.idMembSearchTransitiveMembershipsCmd,
	)
	groups.AddCommand(s.identityMembershipsCmd)

	s.identityCmd.AddCommand(groups)

}

// --- memberships impl ---

func (s *identityCmds) runIDMembAdd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, s.flagIDGroupEmail, s.flagIDGroupNamespace)
	if err != nil {
		return err
	}
	ns := s.flagIDMemberNamespace
	if ns == "" {
		ns = s.flagIDGroupNamespace
	}
	memb := &cloudidentity.Membership{
		PreferredMemberKey: &cloudidentity.EntityKey{Id: s.flagIDMemberEmail, Namespace: ns},
		Roles:              idParseRoles(s.flagIDRoles),
		Type:               s.flagIDMemberType,
		DeliverySetting:    s.flagIDDeliverySetting,
	}
	op, err := svc.Groups.Memberships.Create(groupName, memb).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding membership: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagIDFormat)
}

func (s *identityCmds) runIDMembDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, s.flagIDGroupEmail, s.flagIDGroupNamespace)
	if err != nil {
		return err
	}
	name, err := idResolveMembershipName(ctx, svc, groupName, s.flagIDMemberEmail, s.flagIDMemberNamespace)
	if err != nil {
		return err
	}
	op, err := svc.Groups.Memberships.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting membership: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagIDFormat)
}

func (s *identityCmds) runIDMembDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, s.flagIDGroupEmail, s.flagIDGroupNamespace)
	if err != nil {
		return err
	}
	name, err := idResolveMembershipName(ctx, svc, groupName, s.flagIDMemberEmail, s.flagIDMemberNamespace)
	if err != nil {
		return err
	}
	got, err := svc.Groups.Memberships.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing membership: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIDFormat)
}

func (s *identityCmds) runIDMembList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, s.flagIDGroupEmail, s.flagIDGroupNamespace)
	if err != nil {
		return err
	}
	var all []*cloudidentity.Membership
	pageToken := ""
	for {
		call := svc.Groups.Memberships.List(groupName).Context(ctx)
		if s.flagIDView != "" {
			call = call.View(strings.ToUpper(s.flagIDView))
		}
		if s.flagIDPageSize > 0 {
			call = call.PageSize(s.flagIDPageSize)
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
	if s.flagIDFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagIDFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %-40s %s\n", "MEMBERSHIP", "MEMBER", "ROLES")
	for _, m := range all {
		var roles []string
		for _, r := range m.Roles {
			roles = append(roles, r.Name)
		}
		id := ""
		if m.PreferredMemberKey != nil {
			id = m.PreferredMemberKey.Id
		}
		fmt.Fprintf(s.inv.stdout, "%-40s %-40s %s\n", path.Base(m.Name), id, strings.Join(roles, ","))
	}
	return nil
}

func (s *identityCmds) runIDMembCheckTransitive(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, s.flagIDGroupEmail, s.flagIDGroupNamespace)
	if err != nil {
		return err
	}
	q := fmt.Sprintf("member_key_id == '%s'", s.flagIDMemberEmail)
	ns := s.flagIDMemberNamespace
	if ns == "" {
		ns = s.flagIDGroupNamespace
	}
	if ns != "" {
		q += fmt.Sprintf(" && member_key_namespace == '%s'", ns)
	}
	resp, err := svc.Groups.Memberships.CheckTransitiveMembership(groupName).Query(q).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("checking transitive membership: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagIDFormat)
}

func (s *identityCmds) runIDMembGetGraph(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, s.flagIDGroupEmail, s.flagIDGroupNamespace)
	if err != nil {
		return err
	}
	call := svc.Groups.Memberships.GetMembershipGraph(groupName).Context(ctx)
	if s.flagIDQuery != "" {
		call = call.Query(s.flagIDQuery)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("getting membership graph: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagIDFormat)
}

func (s *identityCmds) runIDMembModifyRoles(cmd *cobra.Command, args []string) error {
	if s.flagIDAddRoles == "" && s.flagIDRemoveRoles == "" {
		return fmt.Errorf("at least one of --add-roles or --remove-roles is required")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, s.flagIDGroupEmail, s.flagIDGroupNamespace)
	if err != nil {
		return err
	}
	name, err := idResolveMembershipName(ctx, svc, groupName, s.flagIDMemberEmail, s.flagIDMemberNamespace)
	if err != nil {
		return err
	}
	req := &cloudidentity.ModifyMembershipRolesRequest{}
	if s.flagIDAddRoles != "" {
		req.AddRoles = idParseRoles(s.flagIDAddRoles)
	}
	if s.flagIDRemoveRoles != "" {
		for _, r := range strings.Split(s.flagIDRemoveRoles, ",") {
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
	return emitFormattedTo(s.inv.stdout, resp, s.flagIDFormat)
}

func (s *identityCmds) runIDMembSearchTransitiveGroups(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	q := s.flagIDQuery
	if q == "" {
		if s.flagIDMemberEmail == "" {
			return fmt.Errorf("either --query or --member-email is required")
		}
		q = fmt.Sprintf("member_key_id == '%s'", s.flagIDMemberEmail)
		ns := s.flagIDMemberNamespace
		if ns != "" {
			q += fmt.Sprintf(" && member_key_namespace == '%s'", ns)
		}
		q += " && 'cloudidentity.googleapis.com/groups.discussion_forum' in labels"
	}
	var all []*cloudidentity.GroupRelation
	pageToken := ""
	for {
		call := svc.Groups.Memberships.SearchTransitiveGroups("groups/-").Query(q).Context(ctx)
		if s.flagIDPageSize > 0 {
			call = call.PageSize(s.flagIDPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagIDFormat)
}

func (s *identityCmds) runIDMembSearchTransitiveMemberships(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	groupName, err := idResolveGroupName(ctx, svc, s.flagIDGroupEmail, s.flagIDGroupNamespace)
	if err != nil {
		return err
	}
	var all []*cloudidentity.MemberRelation
	pageToken := ""
	for {
		call := svc.Groups.Memberships.SearchTransitiveMemberships(groupName).Context(ctx)
		if s.flagIDPageSize > 0 {
			call = call.PageSize(s.flagIDPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagIDFormat)
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

func (s *identityCmds) runIDGroupsCreate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	labels, err := idParseKVLabels(s.flagIDGroupLabels)
	if err != nil {
		return err
	}
	if labels == nil {
		labels = map[string]string{"cloudidentity.googleapis.com/groups.discussion_forum": ""}
	}
	body := &cloudidentity.Group{
		Parent:      idNormalizeCustomer(s.flagIDGroupCustomer),
		GroupKey:    &cloudidentity.EntityKey{Id: args[0], Namespace: s.flagIDGroupNamespace},
		DisplayName: s.flagIDGroupDisplayName,
		Description: s.flagIDGroupDescription,
		Labels:      labels,
	}
	call := svc.Groups.Create(body).Context(ctx)
	switch strings.ToLower(s.flagIDGroupInitConfig) {
	case "":
	case "empty":
		call = call.InitialGroupConfig("EMPTY")
	case "with-initial-owner":
		call = call.InitialGroupConfig("WITH_INITIAL_OWNER")
	default:
		return fmt.Errorf("invalid --initial-group-config %q (want empty or with-initial-owner)", s.flagIDGroupInitConfig)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating group: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagIDFormat)
}

func (s *identityCmds) runIDGroupsDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	name, err := idResolveGroupName(ctx, svc, args[0], s.flagIDGroupNamespace)
	if err != nil {
		return err
	}
	op, err := svc.Groups.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting group: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagIDFormat)
}

func (s *identityCmds) runIDGroupsDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	name, err := idResolveGroupName(ctx, svc, args[0], s.flagIDGroupNamespace)
	if err != nil {
		return err
	}
	got, err := svc.Groups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing group: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIDFormat)
}

func (s *identityCmds) runIDGroupsSearch(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	var all []*cloudidentity.Group
	pageToken := ""
	for {
		call := svc.Groups.Search().Query(s.flagIDQuery).Context(ctx)
		if s.flagIDGroupSearchView != "" {
			call = call.View(strings.ToUpper(s.flagIDGroupSearchView))
		}
		if s.flagIDGroupPageSize > 0 {
			call = call.PageSize(s.flagIDGroupPageSize)
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
	if s.flagIDFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagIDFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %-40s %s\n", "NAME", "GROUP_KEY", "DISPLAY_NAME")
	for _, g := range all {
		key := ""
		if g.GroupKey != nil {
			key = g.GroupKey.Id
		}
		fmt.Fprintf(s.inv.stdout, "%-40s %-40s %s\n", path.Base(g.Name), key, g.DisplayName)
	}
	return nil
}

func (s *identityCmds) runIDGroupsUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	name, err := idResolveGroupName(ctx, svc, args[0], s.flagIDGroupNamespace)
	if err != nil {
		return err
	}
	body := &cloudidentity.Group{}
	mask := s.flagIDGroupUpdateMask
	if s.flagIDGroupDisplayName != "" || strings.Contains(mask, "displayName") {
		body.DisplayName = s.flagIDGroupDisplayName
	}
	if s.flagIDGroupDescription != "" || strings.Contains(mask, "description") {
		body.Description = s.flagIDGroupDescription
	}
	if len(s.flagIDGroupLabels) > 0 || strings.Contains(mask, "labels") {
		labels, err := idParseKVLabels(s.flagIDGroupLabels)
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
	return emitFormattedTo(s.inv.stdout, op, s.flagIDFormat)
}

func (s *identityCmds) runIDGroupsPreview(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	customer := idNormalizeCustomer(s.flagIDGroupCustomer)
	q := s.flagIDQuery
	if q == "" {
		q = fmt.Sprintf("parent == '%s'", customer)
	}
	var all []*cloudidentity.Group
	pageToken := ""
	for {
		call := svc.Groups.Search().Query(q).Context(ctx)
		if s.flagIDView != "" {
			call = call.View(strings.ToUpper(s.flagIDView))
		}
		if s.flagIDGroupPageSize > 0 {
			call = call.PageSize(s.flagIDGroupPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagIDFormat)
}

func (s *identityCmds) runIDGroupsConfigExport(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudIdentity(ctx)
	if err != nil {
		return err
	}
	name, err := idResolveGroupName(ctx, svc, args[0], s.flagIDGroupNamespace)
	if err != nil {
		return err
	}
	g, err := svc.Groups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching group for export: %w", err)
	}
	format := strings.ToLower(s.flagIDGroupConfigFormat)
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
	return emitFormattedTo(s.inv.stdout, out, "yaml")
}

// identityCmds holds the commands and flag values of one invocation's "identity"
// command group, so invocations don't share state (#2070).
type identityCmds struct {
	inv *invocation

	identityCmd                          *cobra.Command
	flagIDGroupEmail                     string
	flagIDMemberEmail                    string
	flagIDGroupNamespace                 string
	flagIDMemberNamespace                string
	flagIDRoles                          string
	flagIDAddRoles                       string
	flagIDRemoveRoles                    string
	flagIDMemberType                     string
	flagIDView                           string
	flagIDQuery                          string
	flagIDPageSize                       int64
	flagIDFormat                         string
	flagIDDeliverySetting                string
	flagIDGroupCustomer                  string
	flagIDGroupDisplayName               string
	flagIDGroupDescription               string
	flagIDGroupLabels                    []string
	flagIDGroupInitConfig                string
	flagIDGroupUpdateMask                string
	flagIDGroupPageSize                  int64
	flagIDGroupSearchView                string
	idGroupsCreateCmd                    *cobra.Command
	idGroupsDeleteCmd                    *cobra.Command
	idGroupsDescribeCmd                  *cobra.Command
	idGroupsSearchCmd                    *cobra.Command
	idGroupsUpdateCmd                    *cobra.Command
	idGroupsPreviewCmd                   *cobra.Command
	flagIDGroupConfigPath                string
	flagIDGroupConfigFormat              string
	idGroupsConfigCmd                    *cobra.Command
	idGroupsConfigExportCmd              *cobra.Command
	identityMembershipsCmd               *cobra.Command
	idMembAddCmd                         *cobra.Command
	idMembDeleteCmd                      *cobra.Command
	idMembDescribeCmd                    *cobra.Command
	idMembListCmd                        *cobra.Command
	idMembCheckTransitiveCmd             *cobra.Command
	idMembGetGraphCmd                    *cobra.Command
	idMembModifyRolesCmd                 *cobra.Command
	idMembSearchTransitiveGroupsCmd      *cobra.Command
	idMembSearchTransitiveMembershipsCmd *cobra.Command
}

// newIdentityCmd returns the "identity" command group for inv.
func newIdentityCmd(inv *invocation) *cobra.Command {
	return newIdentityCmds(inv).identityCmd
}

// newIdentityCmds builds the "identity" command group for inv and returns its state.
func newIdentityCmds(inv *invocation) *identityCmds {
	s := &identityCmds{inv: inv}
	s.identityCmd = &cobra.Command{Use: "identity", Short: "Manage Cloud Identity"}
	s.idGroupsCreateCmd = &cobra.Command{
		Use: "create GROUP_EMAIL", Short: "Create a Cloud Identity group",
		Args: cobra.ExactArgs(1), RunE: s.runIDGroupsCreate,
	}
	s.idGroupsDeleteCmd = &cobra.Command{
		Use: "delete GROUP_EMAIL", Short: "Delete a Cloud Identity group",
		Args: cobra.ExactArgs(1), RunE: s.runIDGroupsDelete,
	}
	s.idGroupsDescribeCmd = &cobra.Command{
		Use: "describe GROUP_EMAIL", Short: "Describe a Cloud Identity group",
		Args: cobra.ExactArgs(1), RunE: s.runIDGroupsDescribe,
	}
	s.idGroupsSearchCmd = &cobra.Command{
		Use: "search", Short: "Search for Cloud Identity groups matching a query",
		Args: cobra.NoArgs, RunE: s.runIDGroupsSearch,
	}
	s.idGroupsUpdateCmd = &cobra.Command{
		Use: "update GROUP_EMAIL", Short: "Update a Cloud Identity group",
		Args: cobra.ExactArgs(1), RunE: s.runIDGroupsUpdate,
	}
	s.idGroupsPreviewCmd = &cobra.Command{
		Use: "preview", Short: "Preview users in a customer account using a CEL query",
		Args: cobra.NoArgs, RunE: s.runIDGroupsPreview,
	}
	s.idGroupsConfigCmd = &cobra.Command{Use: "config", Short: "Manage Cloud Identity group configurations"}
	s.idGroupsConfigExportCmd = &cobra.Command{
		Use: "export GROUP_EMAIL", Short: "Export a Cloud Identity group configuration",
		Args: cobra.ExactArgs(1), RunE: s.runIDGroupsConfigExport,
	}
	s.identityMembershipsCmd = &cobra.Command{Use: "memberships", Short: "Manage Cloud Identity group memberships"}
	s.idMembAddCmd = &cobra.Command{
		Use: "add", Short: "Add a member to a Cloud Identity group",
		Args: cobra.NoArgs, RunE: s.runIDMembAdd,
	}
	s.idMembDeleteCmd = &cobra.Command{
		Use: "delete", Short: "Delete a member from a Cloud Identity group",
		Args: cobra.NoArgs, RunE: s.runIDMembDelete,
	}
	s.idMembDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe a Cloud Identity group membership",
		Args: cobra.NoArgs, RunE: s.runIDMembDescribe,
	}
	s.idMembListCmd = &cobra.Command{
		Use: "list", Short: "List memberships of a Cloud Identity group",
		Args: cobra.NoArgs, RunE: s.runIDMembList,
	}
	s.idMembCheckTransitiveCmd = &cobra.Command{
		Use: "check-transitive-membership", Short: "Check whether a member has a transitive membership in a group",
		Args: cobra.NoArgs, RunE: s.runIDMembCheckTransitive,
	}
	s.idMembGetGraphCmd = &cobra.Command{
		Use: "get-membership-graph", Short: "Get the membership graph of a group",
		Args: cobra.NoArgs, RunE: s.runIDMembGetGraph,
	}
	s.idMembModifyRolesCmd = &cobra.Command{
		Use: "modify-membership-roles", Short: "Modify roles of a Cloud Identity group membership",
		Args: cobra.NoArgs, RunE: s.runIDMembModifyRoles,
	}
	s.idMembSearchTransitiveGroupsCmd = &cobra.Command{
		Use: "search-transitive-groups", Short: "Search transitive groups a member belongs to",
		Args: cobra.NoArgs, RunE: s.runIDMembSearchTransitiveGroups,
	}
	s.idMembSearchTransitiveMembershipsCmd = &cobra.Command{
		Use: "search-transitive-memberships", Short: "Search transitive memberships of a group",
		Args: cobra.NoArgs, RunE: s.runIDMembSearchTransitiveMemberships,
	}
	s.initIdentity()
	return s
}
