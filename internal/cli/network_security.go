package cli

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	networksecurity "google.golang.org/api/networksecurity/v1"
)

// --- gcloud network-security (#363, #819-#846) ---
//
// Every subgroup here is backed by the google.golang.org/api/networksecurity
// service (v1 primarily, v1beta1 for secure-access-connect which only exists
// in beta). The layout mirrors internal/cli/agent_registry.go and internal/cli/metastore.go:
//
//   - a per-group struct of cobra commands
//   - shared --location / --organization / --config-file / --async / --format
//     flags provided by a small builder
//   - LRO create/patch/delete calls block by default and wait via nsWaitOp
//   - update flows accept a JSON or YAML body via --config-file and derive
//     the update mask automatically when --update-mask is not supplied.

// nsProjectParent returns projects/PROJECT/locations/LOCATION using the current
// project and --location.
func (s *networkSecurityCmds) nsProjectParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if s.flagNSLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagNSLocation), nil
}

// nsOrgParent returns organizations/ORG/locations/LOCATION using --organization
// and --location.
func (s *networkSecurityCmds) nsOrgParent() (string, error) {
	if s.flagNSOrganization == "" {
		return "", fmt.Errorf("--organization is required")
	}
	if s.flagNSLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("organizations/%s/locations/%s", s.flagNSOrganization, s.flagNSLocation), nil
}

// nsChild appends "collection/id" to parent unless id is already a fully
// qualified resource name.
func nsChild(parent, collection, id string) string {
	if strings.HasPrefix(id, "projects/") || strings.HasPrefix(id, "organizations/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func nsWaitOp(ctx context.Context, svc *networksecurity.Service, op *networksecurity.Operation) (*networksecurity.Operation, error) {
	for !op.Done {
		var (
			got *networksecurity.Operation
			err error
		)
		switch {
		case strings.HasPrefix(op.Name, "organizations/"):
			got, err = svc.Organizations.Locations.Operations.Get(op.Name).Context(ctx).Do()
		default:
			got, err = svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		}
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func (s *networkSecurityCmds) nsFinishOp(ctx context.Context, svc *networksecurity.Service, op *networksecurity.Operation, verb, name string) error {
	if s.flagNSAsync {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	final, err := nsWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(s.inv.stdout, final.Response, "")
	}
	return nil
}

// addNSLocationFlag adds --location to each command.
func (s *networkSecurityCmds) addNSLocationFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&s.flagNSLocation, "location", "", "Location (required)")
	}
}

// addNSOrgFlags adds --organization + --location to each command.
func (s *networkSecurityCmds) addNSOrgFlags(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&s.flagNSOrganization, "organization", "", "Organization ID (required)")
		c.Flags().StringVar(&s.flagNSLocation, "location", "", "Location (required)")
	}
}

// addNSFormatFlag adds --format to each command.
func (s *networkSecurityCmds) addNSFormatFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&s.flagNSFormat, "format", "", "Output format")
	}
}

// addNSFilterFlag adds --filter to each command.
func (s *networkSecurityCmds) addNSFilterFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&s.flagNSFilter, "filter", "", "Server-side list filter")
	}
}

// addNSAsyncFlag adds --async to each command.
func (s *networkSecurityCmds) addNSAsyncFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().BoolVar(&s.flagNSAsync, "async", false, "Do not wait for the operation to complete")
	}
}

// addNSRequestIDFlag adds --request-id to each command.
func (s *networkSecurityCmds) addNSRequestIDFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&s.flagNSRequestID, "request-id", "", "Optional idempotency request ID")
	}
}

// addNSCreateConfigFlag adds a required --config-file flag.
func (s *networkSecurityCmds) addNSCreateConfigFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&s.flagNSConfigFile, "config-file", "", "Path to a JSON/YAML file with the resource body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
}

// addNSUpdateConfigFlag adds a required --config-file plus --update-mask.
func (s *networkSecurityCmds) addNSUpdateConfigFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&s.flagNSConfigFile, "config-file", "", "Path to a JSON/YAML file with the resource body (required)")
		_ = c.MarkFlagRequired("config-file")
		c.Flags().StringVar(&s.flagNSUpdateMask, "update-mask", "", "Comma-separated list of fields to update (defaults to every populated field)")
	}
}

// nsBasename returns the leaf of a resource name for listing display.
func nsBasename(name string) string { return path.Base(name) }

// nsResolveMask picks --update-mask if given, otherwise every top-level field
// present in body.
func (s *networkSecurityCmds) nsResolveMask(body any) string {
	if s.flagNSUpdateMask != "" {
		return s.flagNSUpdateMask
	}
	return joinMask(nonEmptyJSONFields(body))
}

// --- operations (issue 833) ---

func (s *networkSecurityCmds) runNSOpCancel(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name, err := s.nsResolveOperationName(args[0])
	if err != nil {
		return err
	}
	if strings.HasPrefix(name, "organizations/") {
		if _, err := svc.Organizations.Locations.Operations.Cancel(name, &networksecurity.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
			return fmt.Errorf("cancelling operation: %w", err)
		}
	} else {
		if _, err := svc.Projects.Locations.Operations.Cancel(name, &networksecurity.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
			return fmt.Errorf("cancelling operation: %w", err)
		}
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Cancel request issued for operation %s.\n", args[0])
	return nil
}

func (s *networkSecurityCmds) runNSOpDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name, err := s.nsResolveOperationName(args[0])
	if err != nil {
		return err
	}
	if strings.HasPrefix(name, "organizations/") {
		if _, err := svc.Organizations.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
			return fmt.Errorf("deleting operation: %w", err)
		}
	} else {
		if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
			return fmt.Errorf("deleting operation: %w", err)
		}
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Deleted operation %s.\n", args[0])
	return nil
}

func (s *networkSecurityCmds) runNSOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name, err := s.nsResolveOperationName(args[0])
	if err != nil {
		return err
	}
	var op *networksecurity.Operation
	if strings.HasPrefix(name, "organizations/") {
		op, err = svc.Organizations.Locations.Operations.Get(name).Context(ctx).Do()
	} else {
		op, err = svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagNSFormat)
}

func (s *networkSecurityCmds) runNSOpList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	var parent string
	if s.flagNSOrganization != "" {
		p, err := s.nsOrgParent()
		if err != nil {
			return err
		}
		parent = p
	} else {
		p, err := s.nsProjectParent()
		if err != nil {
			return err
		}
		parent = p
	}
	var all []*networksecurity.Operation
	pageToken := ""
	for {
		var (
			resp *networksecurity.ListOperationsResponse
			err  error
		)
		if strings.HasPrefix(parent, "organizations/") {
			call := svc.Organizations.Locations.Operations.List(parent).Context(ctx)
			if s.flagNSFilter != "" {
				call = call.Filter(s.flagNSFilter)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err = call.Do()
		} else {
			call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
			if s.flagNSFilter != "" {
				call = call.Filter(s.flagNSFilter)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err = call.Do()
		}
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagNSFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagNSFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-60s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Fprintf(s.inv.stdout, "%-60s %v\n", nsBasename(o.Name), o.Done)
	}
	return nil
}

// nsResolveOperationName accepts either a fully qualified operation name or a
// short ID relative to the current --organization / --project + --location.
func (s *networkSecurityCmds) nsResolveOperationName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") || strings.HasPrefix(id, "organizations/") {
		return id, nil
	}
	if s.flagNSOrganization != "" {
		parent, err := s.nsOrgParent()
		if err != nil {
			return "", err
		}
		return nsChild(parent, "operations", id), nil
	}
	parent, err := s.nsProjectParent()
	if err != nil {
		return "", err
	}
	return nsChild(parent, "operations", id), nil
}

func (s *networkSecurityCmds) initNetworkSecurity() {
	// operations
	for _, c := range []*cobra.Command{s.nsOpCancelCmd, s.nsOpDeleteCmd, s.nsOpDescribeCmd, s.nsOpListCmd} {
		c.Flags().StringVar(&s.flagNSLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&s.flagNSOrganization, "organization", "", "Organization ID (for org-scoped operations)")
	}
	s.addNSFormatFlag(s.nsOpDescribeCmd, s.nsOpListCmd)
	s.addNSFilterFlag(s.nsOpListCmd)
	s.networkSecurityOperationsCmd.AddCommand(s.nsOpCancelCmd, s.nsOpDeleteCmd, s.nsOpDescribeCmd, s.nsOpListCmd)
	s.networkSecurityCmd.AddCommand(s.networkSecurityOperationsCmd)

	s.registerNSAddressGroups(s.networkSecurityCmd)
	s.registerNSPolicies(s.networkSecurityCmd)
	s.registerNSFirewall(s.networkSecurityCmd)
	s.registerNSInterceptMirroring(s.networkSecurityCmd)
	s.registerNSProfiles(s.networkSecurityCmd)
	s.registerNSSecureAccessConnect(s.networkSecurityCmd)

}

// networkSecurityCmds holds the commands and flag values of one invocation's "networkSecurity"
// command group, so invocations don't share state (#2070).
type networkSecurityCmds struct {
	inv *invocation

	networkSecurityCmd *cobra.Command
	flagNSLocation     string
	flagNSOrganization string
	flagNSFormat       string
	flagNSFilter       string
	flagNSConfigFile   string
	flagNSUpdateMask   string
	flagNSAsync        bool
	flagNSRequestID    string
	// address-groups item flags
	flagNSItems              []string
	flagNSSourceAddressGroup string
	// address-groups list-references
	flagNSPageSize                     int64
	networkSecurityOperationsCmd       *cobra.Command
	nsOpCancelCmd                      *cobra.Command
	nsOpDeleteCmd                      *cobra.Command
	nsOpDescribeCmd                    *cobra.Command
	nsOpListCmd                        *cobra.Command
	networkSecurityAddressGroupsCmd    *cobra.Command
	networkSecurityOrgAddressGroupsCmd *cobra.Command
	nsAGCreateCmd                      *cobra.Command
	nsAGDeleteCmd                      *cobra.Command
	nsAGDescribeCmd                    *cobra.Command
	nsAGListCmd                        *cobra.Command
	nsAGUpdateCmd                      *cobra.Command
	nsAGAddItemsCmd                    *cobra.Command
	nsAGRemoveItemsCmd                 *cobra.Command
	nsAGCloneItemsCmd                  *cobra.Command
	nsAGListRefsCmd                    *cobra.Command
	nsOAGCreateCmd                     *cobra.Command
	nsOAGDeleteCmd                     *cobra.Command
	nsOAGDescribeCmd                   *cobra.Command
	nsOAGListCmd                       *cobra.Command
	nsOAGUpdateCmd                     *cobra.Command
	nsOAGAddItemsCmd                   *cobra.Command
	nsOAGRemoveItemsCmd                *cobra.Command
	nsOAGCloneItemsCmd                 *cobra.Command
	nsOAGListRefsCmd                   *cobra.Command
	// flagNSGatewayPolicy identifies the parent gateway security policy for rules
	// subcommands.
	flagNSGatewayPolicy string
}

// newNetworkSecurityCmd returns the "networkSecurity" command group for inv.
func newNetworkSecurityCmd(inv *invocation) *cobra.Command {
	return newNetworkSecurityCmds(inv).networkSecurityCmd
}

// newNetworkSecurityCmds builds the "networkSecurity" command group for inv and returns its state.
func newNetworkSecurityCmds(inv *invocation) *networkSecurityCmds {
	s := &networkSecurityCmds{inv: inv}
	s.networkSecurityCmd = &cobra.Command{Use: "network-security", Short: "Manage Network Security"}
	s.networkSecurityOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Network Security operations"}
	s.nsOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: s.runNSOpCancel,
	}
	s.nsOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an operation",
		Args: cobra.ExactArgs(1), RunE: s.runNSOpDelete,
	}
	s.nsOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: s.runNSOpDescribe,
	}
	s.nsOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a location",
		Args: cobra.NoArgs, RunE: s.runNSOpList,
	}
	s.networkSecurityAddressGroupsCmd = &cobra.Command{Use: "address-groups", Short: "Manage address groups"}
	s.networkSecurityOrgAddressGroupsCmd = &cobra.Command{Use: "org-address-groups", Short: "Manage organization address groups"}
	s.nsAGCreateCmd = &cobra.Command{
		Use: "create ADDRESS_GROUP", Short: "Create an address group from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runNSAGCreate,
	}
	s.nsAGDeleteCmd = &cobra.Command{
		Use: "delete ADDRESS_GROUP", Short: "Delete an address group",
		Args: cobra.ExactArgs(1), RunE: s.runNSAGDelete,
	}
	s.nsAGDescribeCmd = &cobra.Command{
		Use: "describe ADDRESS_GROUP", Short: "Describe an address group",
		Args: cobra.ExactArgs(1), RunE: s.runNSAGDescribe,
	}
	s.nsAGListCmd = &cobra.Command{
		Use: "list", Short: "List address groups in a location",
		Args: cobra.NoArgs, RunE: s.runNSAGList,
	}
	s.nsAGUpdateCmd = &cobra.Command{
		Use: "update ADDRESS_GROUP", Short: "Update an address group from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runNSAGUpdate,
	}
	s.nsAGAddItemsCmd = &cobra.Command{
		Use: "add-items ADDRESS_GROUP", Short: "Add items to an address group",
		Args: cobra.ExactArgs(1), RunE: s.runNSAGAddItems,
	}
	s.nsAGRemoveItemsCmd = &cobra.Command{
		Use: "remove-items ADDRESS_GROUP", Short: "Remove items from an address group",
		Args: cobra.ExactArgs(1), RunE: s.runNSAGRemoveItems,
	}
	s.nsAGCloneItemsCmd = &cobra.Command{
		Use: "clone-items ADDRESS_GROUP", Short: "Clone items into an address group",
		Args: cobra.ExactArgs(1), RunE: s.runNSAGCloneItems,
	}
	s.nsAGListRefsCmd = &cobra.Command{
		Use: "list-references ADDRESS_GROUP", Short: "List references to an address group",
		Args: cobra.ExactArgs(1), RunE: s.runNSAGListRefs,
	}
	s.nsOAGCreateCmd = &cobra.Command{
		Use: "create ADDRESS_GROUP", Short: "Create an organization address group from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runNSOAGCreate,
	}
	s.nsOAGDeleteCmd = &cobra.Command{
		Use: "delete ADDRESS_GROUP", Short: "Delete an organization address group",
		Args: cobra.ExactArgs(1), RunE: s.runNSOAGDelete,
	}
	s.nsOAGDescribeCmd = &cobra.Command{
		Use: "describe ADDRESS_GROUP", Short: "Describe an organization address group",
		Args: cobra.ExactArgs(1), RunE: s.runNSOAGDescribe,
	}
	s.nsOAGListCmd = &cobra.Command{
		Use: "list", Short: "List organization address groups",
		Args: cobra.NoArgs, RunE: s.runNSOAGList,
	}
	s.nsOAGUpdateCmd = &cobra.Command{
		Use: "update ADDRESS_GROUP", Short: "Update an organization address group",
		Args: cobra.ExactArgs(1), RunE: s.runNSOAGUpdate,
	}
	s.nsOAGAddItemsCmd = &cobra.Command{
		Use: "add-items ADDRESS_GROUP", Short: "Add items to an organization address group",
		Args: cobra.ExactArgs(1), RunE: s.runNSOAGAddItems,
	}
	s.nsOAGRemoveItemsCmd = &cobra.Command{
		Use: "remove-items ADDRESS_GROUP", Short: "Remove items from an organization address group",
		Args: cobra.ExactArgs(1), RunE: s.runNSOAGRemoveItems,
	}
	s.nsOAGCloneItemsCmd = &cobra.Command{
		Use: "clone-items ADDRESS_GROUP", Short: "Clone items into an organization address group",
		Args: cobra.ExactArgs(1), RunE: s.runNSOAGCloneItems,
	}
	s.nsOAGListRefsCmd = &cobra.Command{
		Use: "list-references ADDRESS_GROUP", Short: "List references to an organization address group",
		Args: cobra.ExactArgs(1), RunE: s.runNSOAGListRefs,
	}
	s.initNetworkSecurity()
	return s
}
