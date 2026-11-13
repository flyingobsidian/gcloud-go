package cli

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	servicedirectory "google.golang.org/api/servicedirectory/v1"
)

// --- gcloud service-directory (#382) ---

func sdLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func sdChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

// --- locations ---

// --- namespaces ---

// --- services ---

// --- endpoints ---

func (g *serviceDirectoryCmds) initServiceDirectory() {
	// locations
	g.sdLocDescribeCmd.Flags().StringVar(&g.flagSDFormat, "format", "", "Output format")
	g.sdLocListCmd.Flags().StringVar(&g.flagSDFormat, "format", "", "Output format")
	g.sdLocationsCmd.AddCommand(g.sdLocDescribeCmd, g.sdLocListCmd)
	g.serviceDirectoryCmd.AddCommand(g.sdLocationsCmd)

	// namespaces
	nsAll := []*cobra.Command{g.sdNSCreateCmd, g.sdNSDeleteCmd, g.sdNSDescribeCmd, g.sdNSListCmd, g.sdNSUpdateCmd,
		g.sdNSGetIamCmd, g.sdNSSetIamCmd, g.sdNSAddIamCmd, g.sdNSRemoveIamCmd}
	for _, c := range nsAll {
		c.Flags().StringVar(&g.flagSDLocation, "location", "", "Location containing the namespace (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{g.sdNSCreateCmd, g.sdNSUpdateCmd} {
		c.Flags().StringVar(&g.flagSDConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Namespace body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.sdNSUpdateCmd.Flags().StringVar(&g.flagSDUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{g.sdNSDescribeCmd, g.sdNSListCmd, g.sdNSGetIamCmd} {
		c.Flags().StringVar(&g.flagSDFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{g.sdNSAddIamCmd, g.sdNSRemoveIamCmd} {
		c.Flags().StringVar(&g.flagSDIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&g.flagSDIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	g.sdNamespacesCmd.AddCommand(nsAll...)
	g.serviceDirectoryCmd.AddCommand(g.sdNamespacesCmd)

	// services
	svcAll := []*cobra.Command{g.sdSvcCreateCmd, g.sdSvcDeleteCmd, g.sdSvcDescribeCmd, g.sdSvcListCmd, g.sdSvcUpdateCmd, g.sdSvcResolveCmd}
	for _, c := range svcAll {
		c.Flags().StringVar(&g.flagSDLocation, "location", "", "Location containing the service (required)")
		c.Flags().StringVar(&g.flagSDNamespace, "namespace", "", "Namespace containing the service (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("namespace")
	}
	for _, c := range []*cobra.Command{g.sdSvcCreateCmd, g.sdSvcUpdateCmd} {
		c.Flags().StringVar(&g.flagSDConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Service body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.sdSvcUpdateCmd.Flags().StringVar(&g.flagSDUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{g.sdSvcDescribeCmd, g.sdSvcListCmd, g.sdSvcResolveCmd} {
		c.Flags().StringVar(&g.flagSDFormat, "format", "", "Output format")
	}
	g.sdServicesCmd.AddCommand(svcAll...)
	g.serviceDirectoryCmd.AddCommand(g.sdServicesCmd)

	// endpoints
	epAll := []*cobra.Command{g.sdEPCreateCmd, g.sdEPDeleteCmd, g.sdEPDescribeCmd, g.sdEPListCmd, g.sdEPUpdateCmd}
	for _, c := range epAll {
		c.Flags().StringVar(&g.flagSDLocation, "location", "", "Location containing the endpoint (required)")
		c.Flags().StringVar(&g.flagSDNamespace, "namespace", "", "Namespace containing the endpoint (required)")
		c.Flags().StringVar(&g.flagSDService, "service", "", "Service containing the endpoint (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("namespace")
		_ = c.MarkFlagRequired("service")
	}
	for _, c := range []*cobra.Command{g.sdEPCreateCmd, g.sdEPUpdateCmd} {
		c.Flags().StringVar(&g.flagSDConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Endpoint body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.sdEPUpdateCmd.Flags().StringVar(&g.flagSDUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{g.sdEPDescribeCmd, g.sdEPListCmd} {
		c.Flags().StringVar(&g.flagSDFormat, "format", "", "Output format")
	}
	g.sdEndpointsCmd.AddCommand(epAll...)
	g.serviceDirectoryCmd.AddCommand(g.sdEndpointsCmd)

}

// --- locations impl ---

func (g *serviceDirectoryCmds) runSDLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(sdLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, loc, g.flagSDFormat)
}

func (g *serviceDirectoryCmds) runSDLocList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	if g.flagSDFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Locations, g.flagSDFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Locations), "%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Fprintf(g.inv.stdout, "%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

// --- namespaces impl ---

func sdNSName(id, project, location string) string {
	return sdChild("namespaces", id, sdLocationParent(project, location))
}

func (g *serviceDirectoryCmds) runSDNSCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ns := &servicedirectory.Namespace{}
	if err := loadYAMLOrJSONInto(g.flagSDConfigFile, ns); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Create(sdLocationParent(project, g.flagSDLocation), ns).
		NamespaceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *serviceDirectoryCmds) runSDNSDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Namespaces.Delete(sdNSName(args[0], project, g.flagSDLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting namespace: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted namespace [%s].\n", args[0])
	return nil
}

func (g *serviceDirectoryCmds) runSDNSDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Get(sdNSName(args[0], project, g.flagSDLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing namespace: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagSDFormat)
}

func (g *serviceDirectoryCmds) runSDNSList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Namespaces.List(sdLocationParent(project, g.flagSDLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing namespaces: %w", err)
	}
	if g.flagSDFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Namespaces, g.flagSDFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Namespaces), "%-40s\n", "NAME")
	for _, n := range resp.Namespaces {
		fmt.Fprintln(g.inv.stdout, path.Base(n.Name))
	}
	return nil
}

func (g *serviceDirectoryCmds) runSDNSUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ns := &servicedirectory.Namespace{}
	if err := loadYAMLOrJSONInto(g.flagSDConfigFile, ns); err != nil {
		return err
	}
	mask := g.flagSDUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(ns))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Patch(sdNSName(args[0], project, g.flagSDLocation), ns).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating namespace: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *serviceDirectoryCmds) runSDNSGetIam(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Namespaces.GetIamPolicy(sdNSName(args[0], project, g.flagSDLocation), &servicedirectory.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, policy, g.flagSDFormat)
}

func (g *serviceDirectoryCmds) runSDNSSetIam(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	policy := &servicedirectory.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.SetIamPolicy(sdNSName(args[0], project, g.flagSDLocation), &servicedirectory.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *serviceDirectoryCmds) runSDNSAddIam(cmd *cobra.Command, args []string) error {
	return g.sdNSModifyIam(args[0], func(p *servicedirectory.Policy) {
		for _, b := range p.Bindings {
			if b.Role == g.flagSDIamRole {
				for _, m := range b.Members {
					if m == g.flagSDIamMember {
						return
					}
				}
				b.Members = append(b.Members, g.flagSDIamMember)
				return
			}
		}
		p.Bindings = append(p.Bindings, &servicedirectory.Binding{Role: g.flagSDIamRole, Members: []string{g.flagSDIamMember}})
	})
}

func (g *serviceDirectoryCmds) runSDNSRemoveIam(cmd *cobra.Command, args []string) error {
	return g.sdNSModifyIam(args[0], func(p *servicedirectory.Policy) {
		for _, b := range p.Bindings {
			if b.Role != g.flagSDIamRole {
				continue
			}
			out := b.Members[:0]
			for _, m := range b.Members {
				if m != g.flagSDIamMember {
					out = append(out, m)
				}
			}
			b.Members = out
		}
	})
}

func (g *serviceDirectoryCmds) sdNSModifyIam(name string, mutate func(*servicedirectory.Policy)) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := g.inv.context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	resource := sdNSName(name, project, g.flagSDLocation)
	policy, err := svc.Projects.Locations.Namespaces.GetIamPolicy(resource, &servicedirectory.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.Namespaces.SetIamPolicy(resource, &servicedirectory.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

// --- services impl ---

func sdSvcParent(project, location, ns string) string {
	return fmt.Sprintf("%s/namespaces/%s", sdLocationParent(project, location), ns)
}

func sdSvcName(id, project, location, ns string) string {
	return sdChild("services", id, sdSvcParent(project, location, ns))
}

func (g *serviceDirectoryCmds) runSDSvcCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	s := &servicedirectory.Service{}
	if err := loadYAMLOrJSONInto(g.flagSDConfigFile, s); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Services.Create(sdSvcParent(project, g.flagSDLocation, g.flagSDNamespace), s).
		ServiceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *serviceDirectoryCmds) runSDSvcDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Namespaces.Services.Delete(sdSvcName(args[0], project, g.flagSDLocation, g.flagSDNamespace)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted service [%s].\n", args[0])
	return nil
}

func (g *serviceDirectoryCmds) runSDSvcDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Services.Get(sdSvcName(args[0], project, g.flagSDLocation, g.flagSDNamespace)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagSDFormat)
}

func (g *serviceDirectoryCmds) runSDSvcList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Namespaces.Services.List(sdSvcParent(project, g.flagSDLocation, g.flagSDNamespace)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing services: %w", err)
	}
	if g.flagSDFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Services, g.flagSDFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Services), "%-40s\n", "NAME")
	for _, s := range resp.Services {
		fmt.Fprintln(g.inv.stdout, path.Base(s.Name))
	}
	return nil
}

func (g *serviceDirectoryCmds) runSDSvcUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	s := &servicedirectory.Service{}
	if err := loadYAMLOrJSONInto(g.flagSDConfigFile, s); err != nil {
		return err
	}
	mask := g.flagSDUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(s))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Services.Patch(sdSvcName(args[0], project, g.flagSDLocation, g.flagSDNamespace), s).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating service: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *serviceDirectoryCmds) runSDSvcResolve(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Services.Resolve(sdSvcName(args[0], project, g.flagSDLocation, g.flagSDNamespace), &servicedirectory.ResolveServiceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resolving service: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagSDFormat)
}

// --- endpoints impl ---

func sdEPParent(project, location, ns, service string) string {
	return fmt.Sprintf("%s/services/%s", sdSvcParent(project, location, ns), service)
}

func sdEPName(id, project, location, ns, service string) string {
	return sdChild("endpoints", id, sdEPParent(project, location, ns, service))
}

func (g *serviceDirectoryCmds) runSDEPCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	e := &servicedirectory.Endpoint{}
	if err := loadYAMLOrJSONInto(g.flagSDConfigFile, e); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Services.Endpoints.Create(sdEPParent(project, g.flagSDLocation, g.flagSDNamespace, g.flagSDService), e).
		EndpointId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating endpoint: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *serviceDirectoryCmds) runSDEPDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Namespaces.Services.Endpoints.Delete(sdEPName(args[0], project, g.flagSDLocation, g.flagSDNamespace, g.flagSDService)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting endpoint: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted endpoint [%s].\n", args[0])
	return nil
}

func (g *serviceDirectoryCmds) runSDEPDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Services.Endpoints.Get(sdEPName(args[0], project, g.flagSDLocation, g.flagSDNamespace, g.flagSDService)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing endpoint: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagSDFormat)
}

func (g *serviceDirectoryCmds) runSDEPList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Namespaces.Services.Endpoints.List(sdEPParent(project, g.flagSDLocation, g.flagSDNamespace, g.flagSDService)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing endpoints: %w", err)
	}
	if g.flagSDFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Endpoints, g.flagSDFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Endpoints), "%-40s %s\n", "NAME", "ADDRESS")
	for _, e := range resp.Endpoints {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(e.Name), e.Address)
	}
	return nil
}

func (g *serviceDirectoryCmds) runSDEPUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	e := &servicedirectory.Endpoint{}
	if err := loadYAMLOrJSONInto(g.flagSDConfigFile, e); err != nil {
		return err
	}
	mask := g.flagSDUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(e))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceDirectory(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Services.Endpoints.Patch(sdEPName(args[0], project, g.flagSDLocation, g.flagSDNamespace, g.flagSDService), e).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating endpoint: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

// serviceDirectoryCmds holds the commands and flag values of one invocation's "serviceDirectory"
// command group, so invocations don't share state (#2070).
type serviceDirectoryCmds struct {
	inv *invocation

	serviceDirectoryCmd *cobra.Command
	flagSDLocation      string
	flagSDNamespace     string
	flagSDService       string
	flagSDConfigFile    string
	flagSDUpdateMask    string
	flagSDFormat        string
	flagSDIamMember     string
	flagSDIamRole       string
	sdLocationsCmd      *cobra.Command
	sdLocDescribeCmd    *cobra.Command
	sdLocListCmd        *cobra.Command
	sdNamespacesCmd     *cobra.Command
	sdNSCreateCmd       *cobra.Command
	sdNSDeleteCmd       *cobra.Command
	sdNSDescribeCmd     *cobra.Command
	sdNSListCmd         *cobra.Command
	sdNSUpdateCmd       *cobra.Command
	sdNSGetIamCmd       *cobra.Command
	sdNSSetIamCmd       *cobra.Command
	sdNSAddIamCmd       *cobra.Command
	sdNSRemoveIamCmd    *cobra.Command
	sdServicesCmd       *cobra.Command
	sdSvcCreateCmd      *cobra.Command
	sdSvcDeleteCmd      *cobra.Command
	sdSvcDescribeCmd    *cobra.Command
	sdSvcListCmd        *cobra.Command
	sdSvcUpdateCmd      *cobra.Command
	sdSvcResolveCmd     *cobra.Command
	sdEndpointsCmd      *cobra.Command
	sdEPCreateCmd       *cobra.Command
	sdEPDeleteCmd       *cobra.Command
	sdEPDescribeCmd     *cobra.Command
	sdEPListCmd         *cobra.Command
	sdEPUpdateCmd       *cobra.Command
}

// newServiceDirectoryCmd returns the "serviceDirectory" command group for inv.
func newServiceDirectoryCmd(inv *invocation) *cobra.Command {
	return newServiceDirectoryCmds(inv).serviceDirectoryCmd
}

// newServiceDirectoryCmds builds the "serviceDirectory" command group for inv and returns its state.
func newServiceDirectoryCmds(inv *invocation) *serviceDirectoryCmds {
	g := &serviceDirectoryCmds{inv: inv}
	g.serviceDirectoryCmd = &cobra.Command{Use: "service-directory", Short: "Manage Service Directory"}
	g.sdLocationsCmd = &cobra.Command{Use: "locations", Short: "Explore Service Directory locations"}
	g.sdLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a location",
		Args: cobra.ExactArgs(1), RunE: g.runSDLocDescribe,
	}
	g.sdLocListCmd = &cobra.Command{
		Use: "list", Short: "List locations",
		Args: cobra.NoArgs, RunE: g.runSDLocList,
	}
	g.sdNamespacesCmd = &cobra.Command{Use: "namespaces", Short: "Manage Service Directory namespaces"}
	g.sdNSCreateCmd = &cobra.Command{
		Use: "create NAMESPACE", Short: "Create a namespace from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSDNSCreate,
	}
	g.sdNSDeleteCmd = &cobra.Command{
		Use: "delete NAMESPACE", Short: "Delete a namespace",
		Args: cobra.ExactArgs(1), RunE: g.runSDNSDelete,
	}
	g.sdNSDescribeCmd = &cobra.Command{
		Use: "describe NAMESPACE", Short: "Describe a namespace",
		Args: cobra.ExactArgs(1), RunE: g.runSDNSDescribe,
	}
	g.sdNSListCmd = &cobra.Command{
		Use: "list", Short: "List namespaces in a location",
		Args: cobra.NoArgs, RunE: g.runSDNSList,
	}
	g.sdNSUpdateCmd = &cobra.Command{
		Use: "update NAMESPACE", Short: "Update a namespace from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSDNSUpdate,
	}
	g.sdNSGetIamCmd = &cobra.Command{
		Use: "get-iam-policy NAMESPACE", Short: "Get the IAM policy for a namespace",
		Args: cobra.ExactArgs(1), RunE: g.runSDNSGetIam,
	}
	g.sdNSSetIamCmd = &cobra.Command{
		Use: "set-iam-policy NAMESPACE POLICY_FILE", Short: "Replace the IAM policy",
		Args: cobra.ExactArgs(2), RunE: g.runSDNSSetIam,
	}
	g.sdNSAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding NAMESPACE", Short: "Add an IAM binding",
		Args: cobra.ExactArgs(1), RunE: g.runSDNSAddIam,
	}
	g.sdNSRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding NAMESPACE", Short: "Remove an IAM binding",
		Args: cobra.ExactArgs(1), RunE: g.runSDNSRemoveIam,
	}
	g.sdServicesCmd = &cobra.Command{Use: "services", Short: "Manage Service Directory services"}
	g.sdSvcCreateCmd = &cobra.Command{
		Use: "create SERVICE", Short: "Create a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSDSvcCreate,
	}
	g.sdSvcDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete a service",
		Args: cobra.ExactArgs(1), RunE: g.runSDSvcDelete,
	}
	g.sdSvcDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a service",
		Args: cobra.ExactArgs(1), RunE: g.runSDSvcDescribe,
	}
	g.sdSvcListCmd = &cobra.Command{
		Use: "list", Short: "List services in a namespace",
		Args: cobra.NoArgs, RunE: g.runSDSvcList,
	}
	g.sdSvcUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSDSvcUpdate,
	}
	g.sdSvcResolveCmd = &cobra.Command{
		Use: "resolve SERVICE", Short: "Resolve a service (return endpoints)",
		Args: cobra.ExactArgs(1), RunE: g.runSDSvcResolve,
	}
	g.sdEndpointsCmd = &cobra.Command{Use: "endpoints", Short: "Manage Service Directory endpoints"}
	g.sdEPCreateCmd = &cobra.Command{
		Use: "create ENDPOINT", Short: "Create an endpoint from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSDEPCreate,
	}
	g.sdEPDeleteCmd = &cobra.Command{
		Use: "delete ENDPOINT", Short: "Delete an endpoint",
		Args: cobra.ExactArgs(1), RunE: g.runSDEPDelete,
	}
	g.sdEPDescribeCmd = &cobra.Command{
		Use: "describe ENDPOINT", Short: "Describe an endpoint",
		Args: cobra.ExactArgs(1), RunE: g.runSDEPDescribe,
	}
	g.sdEPListCmd = &cobra.Command{
		Use: "list", Short: "List endpoints in a service",
		Args: cobra.NoArgs, RunE: g.runSDEPList,
	}
	g.sdEPUpdateCmd = &cobra.Command{
		Use: "update ENDPOINT", Short: "Update an endpoint from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSDEPUpdate,
	}
	g.initServiceDirectory()
	return g
}
