package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	su "google.golang.org/api/serviceusage/v1"
)

// --- gcloud services (#284) ---

func (g *servicesCmds) initServices() {
	g.servicesDisableCmd.Flags().BoolVar(&g.flagServicesForce, "force", false, "Also disable enabled dependent services")

	g.servicesListCmd.Flags().BoolVar(&g.flagServicesEnabled, "enabled", false, "Return services which are enabled (default)")
	g.servicesListCmd.Flags().BoolVar(&g.flagServicesAvail, "available", false, "Return services which are available to be enabled on the project")
	g.servicesListCmd.Flags().StringVar(&g.flagServicesFormat, "format", "", "Output format (json, yaml, or table)")
	g.servicesListCmd.Flags().Int64Var(&g.flagServicesPage, "page-size", 0, "Page size for API pagination")
	g.servicesListCmd.Flags().Int64Var(&g.flagServicesLimit, "limit", 0, "Maximum number of services to list (0 = no limit)")

	g.servicesCmd.AddCommand(g.servicesEnableCmd, g.servicesDisableCmd, g.servicesListCmd)

}

// serviceResourceName builds the fully qualified name for a service enable/
// disable call: `projects/{project}/services/{serviceName}`. Accepts either
// the bare API service name (e.g. `compute.googleapis.com`) or a fully
// qualified resource name.
func serviceResourceName(project, service string) string {
	if strings.HasPrefix(service, "projects/") {
		return service
	}
	return fmt.Sprintf("projects/%s/services/%s", project, service)
}

func (g *servicesCmds) runServicesEnable(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceUsage(ctx)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		op, err := svc.Services.Enable(serviceResourceName(project, args[0]), &su.EnableServiceRequest{}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("enabling service: %w", err)
		}
		fmt.Fprintf(g.inv.stdout, "Enable service in progress (operation: %s).\n", op.Name)
		return g.inv.yaml(op)
	}
	op, err := svc.Services.BatchEnable("projects/"+project, &su.BatchEnableServicesRequest{
		ServiceIds: args,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling services: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Batch enable in progress (operation: %s).\n", op.Name)
	return g.inv.yaml(op)
}

func (g *servicesCmds) runServicesDisable(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceUsage(ctx)
	if err != nil {
		return err
	}

	req := &su.DisableServiceRequest{DisableDependentServices: g.flagServicesForce}
	// Disable does not have a batch endpoint; iterate serially so a partial
	// failure surfaces on the offending service.
	for _, s := range args {
		op, err := svc.Services.Disable(serviceResourceName(project, s), req).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("disabling service %q: %w", s, err)
		}
		fmt.Fprintf(g.inv.stdout, "Disable %s in progress (operation: %s).\n", s, op.Name)
	}
	return nil
}

// servicesListFilter maps the --enabled / --available toggles to the API's
// state filter. --enabled is the default when neither is specified.
func (g *servicesCmds) servicesListFilter() string {
	if g.flagServicesAvail && !g.flagServicesEnabled {
		// gcloud Python treats --available as "not enabled" and returns the
		// full catalogue; the Service Usage API achieves this with no filter.
		return ""
	}
	return "state:ENABLED"
}

func (g *servicesCmds) runServicesList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceUsage(ctx)
	if err != nil {
		return err
	}

	filter := g.servicesListFilter()
	var all []*su.GoogleApiServiceusageV1Service
	pageToken := ""
	for {
		call := svc.Services.List("projects/" + project).Context(ctx)
		if filter != "" {
			call = call.Filter(filter)
		}
		if g.flagServicesPage > 0 {
			call = call.PageSize(g.flagServicesPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing services: %w", err)
		}
		all = append(all, resp.Services...)
		if g.flagServicesLimit > 0 && int64(len(all)) >= g.flagServicesLimit {
			all = all[:g.flagServicesLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return g.inv.printListResults(all, g.flagServicesFormat, func() {
		printTableHeadingTo(g.inv.stdout, len(all), "%-60s %-40s %s\n", "NAME", "TITLE", "STATE")
		for _, s := range all {
			title := ""
			if s.Config != nil {
				title = s.Config.Title
			}
			fmt.Fprintf(g.inv.stdout, "%-60s %-40s %s\n", s.Name, title, s.State)
		}
	})
}

// servicesCmds holds the commands and flag values of one invocation's "services"
// command group, so invocations don't share state (#2070).
type servicesCmds struct {
	inv *invocation

	servicesCmd                      *cobra.Command
	servicesEnableCmd                *cobra.Command
	servicesDisableCmd               *cobra.Command
	servicesListCmd                  *cobra.Command
	flagServicesEnabled              bool
	flagServicesAvail                bool
	flagServicesForce                bool
	flagServicesFormat               string
	flagServicesPage                 int64
	flagServicesLimit                int64
	servicesAPIKeysCmd               *cobra.Command
	apiKeyCreateCmd                  *cobra.Command
	apiKeyDeleteCmd                  *cobra.Command
	apiKeyDescribeCmd                *cobra.Command
	apiKeyListCmd                    *cobra.Command
	apiKeyLookupCmd                  *cobra.Command
	apiKeyGetKeyStringCmd            *cobra.Command
	apiKeyUpdateCmd                  *cobra.Command
	apiKeyUndeleteCmd                *cobra.Command
	flagAPIKeyDisplayName            string
	flagAPIKeyAnnotations            map[string]string
	flagAPIKeyAllowedIPs             []string
	flagAPIKeyAllowedReferrers       []string
	flagAPIKeyAllowedBundleIDs       []string
	flagAPIKeyAllowedAndroidPackages []string
	flagAPIKeyAllowedAndroidSHAs     []string
	flagAPIKeyAPITargets             []string
	flagAPIKeyAppendRestrictions     bool
	flagAPIKeyServiceAccount         string
	flagAPIKeyListShowDeleted        bool
	flagAPIKeyListFormat             string
	flagAPIKeyListPageSize           int64
	flagAPIKeyListLimit              int64
	flagAPIKeyCheckUsage             bool
	flagAPIKeyNoCheckUsage           bool
	servicesOperationsCmd            *cobra.Command
	servicesOperationDescribeCmd     *cobra.Command
	servicesOperationWaitCmd         *cobra.Command
	flagServicesOpWaitInterval       time.Duration
	servicesPeeredDNSCmd             *cobra.Command
	peeredDNSCreateCmd               *cobra.Command
	peeredDNSDeleteCmd               *cobra.Command
	peeredDNSListCmd                 *cobra.Command
	flagPeeredDNSService             string
	flagPeeredDNSNetwork             string
	flagPeeredDNSDnsSuffix           string
	flagPeeredDNSListFormat          string
	servicesVpcPeeringsCmd           *cobra.Command
	vpcPeeringConnectCmd             *cobra.Command
	vpcPeeringListCmd                *cobra.Command
	vpcPeeringUpdateCmd              *cobra.Command
	vpcPeeringDeleteCmd              *cobra.Command
	vpcPeeringEnableVpcSCCmd         *cobra.Command
	vpcPeeringDisableVpcSCCmd        *cobra.Command
	vpcPeeringGetVpcSCCmd            *cobra.Command
	flagVpcPeeringService            string
	flagVpcPeeringNetwork            string
	flagVpcPeeringRanges             []string
	flagVpcPeeringForce              bool
	flagVpcPeeringListFormat         string
}

// newServicesCmd returns the "services" command group for inv.
func newServicesCmd(inv *invocation) *cobra.Command {
	return newServicesCmds(inv).servicesCmd
}

// newServicesCmds builds the "services" command group for inv and returns its state.
func newServicesCmds(inv *invocation) *servicesCmds {
	g := &servicesCmds{inv: inv}
	g.servicesCmd = &cobra.Command{
		Use:   "services",
		Short: "Manage Cloud services",
	}
	g.servicesEnableCmd = &cobra.Command{
		Use:   "enable SERVICE_NAME [SERVICE_NAME ...]",
		Short: "Enable one or more services for a project",
		Args:  cobra.MinimumNArgs(1),
		RunE:  g.runServicesEnable,
	}
	g.servicesDisableCmd = &cobra.Command{
		Use:   "disable SERVICE_NAME [SERVICE_NAME ...]",
		Short: "Disable one or more services for a project",
		Args:  cobra.MinimumNArgs(1),
		RunE:  g.runServicesDisable,
	}
	g.servicesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List services for a project",
		Args:  cobra.NoArgs,
		RunE:  g.runServicesList,
	}
	g.servicesAPIKeysCmd = &cobra.Command{
		Use:   "api-keys",
		Short: "Manage API keys",
	}
	g.apiKeyCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an API key",
		Args:  cobra.NoArgs,
		RunE:  g.runAPIKeyCreate,
	}
	g.apiKeyDeleteCmd = &cobra.Command{
		Use:   "delete KEY_ID_OR_NAME",
		Short: "Delete an API key",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runAPIKeyDelete,
	}
	g.apiKeyDescribeCmd = &cobra.Command{
		Use:   "describe KEY_ID_OR_NAME",
		Short: "Describe an API key",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runAPIKeyDescribe,
	}
	g.apiKeyListCmd = &cobra.Command{
		Use:   "list",
		Short: "List API keys",
		Args:  cobra.NoArgs,
		RunE:  g.runAPIKeyList,
	}
	g.apiKeyLookupCmd = &cobra.Command{
		Use:   "lookup KEY_STRING",
		Short: "Look up the resource name of an API key from its key string",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runAPIKeyLookup,
	}
	g.apiKeyGetKeyStringCmd = &cobra.Command{
		Use:   "get-key-string KEY_ID_OR_NAME",
		Short: "Get the string of an API key",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runAPIKeyGetKeyString,
	}
	g.apiKeyUpdateCmd = &cobra.Command{
		Use:   "update KEY_ID_OR_NAME",
		Short: "Update an API key",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runAPIKeyUpdate,
	}
	g.apiKeyUndeleteCmd = &cobra.Command{
		Use:   "undelete KEY_ID_OR_NAME",
		Short: "Undelete an API key deleted within the last 30 days",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runAPIKeyUndelete,
	}
	g.servicesOperationsCmd = &cobra.Command{
		Use:   "operations",
		Short: "Manage Service Usage operations",
	}
	g.servicesOperationDescribeCmd = &cobra.Command{
		Use:   "describe OPERATION_NAME",
		Short: "Describe a Service Usage operation",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runServicesOperationDescribe,
	}
	g.servicesOperationWaitCmd = &cobra.Command{
		Use:   "wait OPERATION_NAME",
		Short: "Wait for a Service Usage operation to complete",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runServicesOperationWait,
	}
	g.servicesPeeredDNSCmd = &cobra.Command{
		Use:   "peered-dns-domains",
		Short: "Manage peered DNS domains for private service connections",
	}
	g.peeredDNSCreateCmd = &cobra.Command{
		Use:   "create NAME",
		Short: "Create a peered DNS domain",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runPeeredDNSCreate,
	}
	g.peeredDNSDeleteCmd = &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a peered DNS domain",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runPeeredDNSDelete,
	}
	g.peeredDNSListCmd = &cobra.Command{
		Use:   "list",
		Short: "List peered DNS domains",
		Args:  cobra.NoArgs,
		RunE:  g.runPeeredDNSList,
	}
	g.servicesVpcPeeringsCmd = &cobra.Command{
		Use:   "vpc-peerings",
		Short: "Manage VPC peerings to services",
	}
	g.vpcPeeringConnectCmd = &cobra.Command{
		Use:   "connect",
		Short: "Create a peering connection between a consumer network and a service",
		Args:  cobra.NoArgs,
		RunE:  g.runVpcPeeringConnect,
	}
	g.vpcPeeringListCmd = &cobra.Command{
		Use:   "list",
		Short: "List VPC peering connections for a service",
		Args:  cobra.NoArgs,
		RunE:  g.runVpcPeeringList,
	}
	g.vpcPeeringUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update reserved ranges on a VPC peering connection",
		Args:  cobra.NoArgs,
		RunE:  g.runVpcPeeringUpdate,
	}
	g.vpcPeeringDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a VPC peering connection",
		Args:  cobra.NoArgs,
		RunE:  g.runVpcPeeringDelete,
	}
	g.vpcPeeringEnableVpcSCCmd = &cobra.Command{
		Use:   "enable-vpc-service-controls",
		Short: "Enable VPC Service Controls on a peering connection",
		Args:  cobra.NoArgs,
		RunE:  g.makeVpcSCToggle(true),
	}
	g.vpcPeeringDisableVpcSCCmd = &cobra.Command{
		Use:   "disable-vpc-service-controls",
		Short: "Disable VPC Service Controls on a peering connection",
		Args:  cobra.NoArgs,
		RunE:  g.makeVpcSCToggle(false),
	}
	g.vpcPeeringGetVpcSCCmd = &cobra.Command{
		Use:   "get-vpc-service-controls",
		Short: "Get VPC Service Controls status on a peering connection",
		Args:  cobra.NoArgs,
		RunE:  g.runVpcPeeringGetVpcSC,
	}
	g.initServices()
	g.initServicesApiKeys()
	g.initServicesOperations()
	g.initServicesPeeredDnsDomains()
	g.initServicesVpcPeerings()
	return g
}
