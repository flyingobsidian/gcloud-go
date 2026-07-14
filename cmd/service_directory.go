package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	servicedirectory "google.golang.org/api/servicedirectory/v1"
)

// --- gcloud service-directory (#382) ---

var serviceDirectoryCmd = &cobra.Command{Use: "service-directory", Short: "Manage Service Directory"}

func sdLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func sdChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

var (
	flagSDLocation   string
	flagSDNamespace  string
	flagSDService    string
	flagSDConfigFile string
	flagSDUpdateMask string
	flagSDFormat     string
	flagSDIamMember  string
	flagSDIamRole    string
)

// --- locations ---

var sdLocationsCmd = &cobra.Command{Use: "locations", Short: "Explore Service Directory locations"}

var (
	sdLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a location",
		Args: cobra.ExactArgs(1), RunE: runSDLocDescribe,
	}
	sdLocListCmd = &cobra.Command{
		Use: "list", Short: "List locations",
		Args: cobra.NoArgs, RunE: runSDLocList,
	}
)

// --- namespaces ---

var sdNamespacesCmd = &cobra.Command{Use: "namespaces", Short: "Manage Service Directory namespaces"}

var (
	sdNSCreateCmd = &cobra.Command{
		Use: "create NAMESPACE", Short: "Create a namespace from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSDNSCreate,
	}
	sdNSDeleteCmd = &cobra.Command{
		Use: "delete NAMESPACE", Short: "Delete a namespace",
		Args: cobra.ExactArgs(1), RunE: runSDNSDelete,
	}
	sdNSDescribeCmd = &cobra.Command{
		Use: "describe NAMESPACE", Short: "Describe a namespace",
		Args: cobra.ExactArgs(1), RunE: runSDNSDescribe,
	}
	sdNSListCmd = &cobra.Command{
		Use: "list", Short: "List namespaces in a location",
		Args: cobra.NoArgs, RunE: runSDNSList,
	}
	sdNSUpdateCmd = &cobra.Command{
		Use: "update NAMESPACE", Short: "Update a namespace from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSDNSUpdate,
	}
	sdNSGetIamCmd = &cobra.Command{
		Use: "get-iam-policy NAMESPACE", Short: "Get the IAM policy for a namespace",
		Args: cobra.ExactArgs(1), RunE: runSDNSGetIam,
	}
	sdNSSetIamCmd = &cobra.Command{
		Use: "set-iam-policy NAMESPACE POLICY_FILE", Short: "Replace the IAM policy",
		Args: cobra.ExactArgs(2), RunE: runSDNSSetIam,
	}
	sdNSAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding NAMESPACE", Short: "Add an IAM binding",
		Args: cobra.ExactArgs(1), RunE: runSDNSAddIam,
	}
	sdNSRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding NAMESPACE", Short: "Remove an IAM binding",
		Args: cobra.ExactArgs(1), RunE: runSDNSRemoveIam,
	}
)

// --- services ---

var sdServicesCmd = &cobra.Command{Use: "services", Short: "Manage Service Directory services"}

var (
	sdSvcCreateCmd = &cobra.Command{
		Use: "create SERVICE", Short: "Create a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSDSvcCreate,
	}
	sdSvcDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete a service",
		Args: cobra.ExactArgs(1), RunE: runSDSvcDelete,
	}
	sdSvcDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a service",
		Args: cobra.ExactArgs(1), RunE: runSDSvcDescribe,
	}
	sdSvcListCmd = &cobra.Command{
		Use: "list", Short: "List services in a namespace",
		Args: cobra.NoArgs, RunE: runSDSvcList,
	}
	sdSvcUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSDSvcUpdate,
	}
	sdSvcResolveCmd = &cobra.Command{
		Use: "resolve SERVICE", Short: "Resolve a service (return endpoints)",
		Args: cobra.ExactArgs(1), RunE: runSDSvcResolve,
	}
)

// --- endpoints ---

var sdEndpointsCmd = &cobra.Command{Use: "endpoints", Short: "Manage Service Directory endpoints"}

var (
	sdEPCreateCmd = &cobra.Command{
		Use: "create ENDPOINT", Short: "Create an endpoint from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSDEPCreate,
	}
	sdEPDeleteCmd = &cobra.Command{
		Use: "delete ENDPOINT", Short: "Delete an endpoint",
		Args: cobra.ExactArgs(1), RunE: runSDEPDelete,
	}
	sdEPDescribeCmd = &cobra.Command{
		Use: "describe ENDPOINT", Short: "Describe an endpoint",
		Args: cobra.ExactArgs(1), RunE: runSDEPDescribe,
	}
	sdEPListCmd = &cobra.Command{
		Use: "list", Short: "List endpoints in a service",
		Args: cobra.NoArgs, RunE: runSDEPList,
	}
	sdEPUpdateCmd = &cobra.Command{
		Use: "update ENDPOINT", Short: "Update an endpoint from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSDEPUpdate,
	}
)

func init() {
	// locations
	sdLocDescribeCmd.Flags().StringVar(&flagSDFormat, "format", "", "Output format")
	sdLocListCmd.Flags().StringVar(&flagSDFormat, "format", "", "Output format")
	sdLocationsCmd.AddCommand(sdLocDescribeCmd, sdLocListCmd)
	serviceDirectoryCmd.AddCommand(sdLocationsCmd)

	// namespaces
	nsAll := []*cobra.Command{sdNSCreateCmd, sdNSDeleteCmd, sdNSDescribeCmd, sdNSListCmd, sdNSUpdateCmd,
		sdNSGetIamCmd, sdNSSetIamCmd, sdNSAddIamCmd, sdNSRemoveIamCmd}
	for _, c := range nsAll {
		c.Flags().StringVar(&flagSDLocation, "location", "", "Location containing the namespace (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{sdNSCreateCmd, sdNSUpdateCmd} {
		c.Flags().StringVar(&flagSDConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Namespace body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	sdNSUpdateCmd.Flags().StringVar(&flagSDUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{sdNSDescribeCmd, sdNSListCmd, sdNSGetIamCmd} {
		c.Flags().StringVar(&flagSDFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{sdNSAddIamCmd, sdNSRemoveIamCmd} {
		c.Flags().StringVar(&flagSDIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&flagSDIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	sdNamespacesCmd.AddCommand(nsAll...)
	serviceDirectoryCmd.AddCommand(sdNamespacesCmd)

	// services
	svcAll := []*cobra.Command{sdSvcCreateCmd, sdSvcDeleteCmd, sdSvcDescribeCmd, sdSvcListCmd, sdSvcUpdateCmd, sdSvcResolveCmd}
	for _, c := range svcAll {
		c.Flags().StringVar(&flagSDLocation, "location", "", "Location containing the service (required)")
		c.Flags().StringVar(&flagSDNamespace, "namespace", "", "Namespace containing the service (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("namespace")
	}
	for _, c := range []*cobra.Command{sdSvcCreateCmd, sdSvcUpdateCmd} {
		c.Flags().StringVar(&flagSDConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Service body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	sdSvcUpdateCmd.Flags().StringVar(&flagSDUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{sdSvcDescribeCmd, sdSvcListCmd, sdSvcResolveCmd} {
		c.Flags().StringVar(&flagSDFormat, "format", "", "Output format")
	}
	sdServicesCmd.AddCommand(svcAll...)
	serviceDirectoryCmd.AddCommand(sdServicesCmd)

	// endpoints
	epAll := []*cobra.Command{sdEPCreateCmd, sdEPDeleteCmd, sdEPDescribeCmd, sdEPListCmd, sdEPUpdateCmd}
	for _, c := range epAll {
		c.Flags().StringVar(&flagSDLocation, "location", "", "Location containing the endpoint (required)")
		c.Flags().StringVar(&flagSDNamespace, "namespace", "", "Namespace containing the endpoint (required)")
		c.Flags().StringVar(&flagSDService, "service", "", "Service containing the endpoint (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("namespace")
		_ = c.MarkFlagRequired("service")
	}
	for _, c := range []*cobra.Command{sdEPCreateCmd, sdEPUpdateCmd} {
		c.Flags().StringVar(&flagSDConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Endpoint body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	sdEPUpdateCmd.Flags().StringVar(&flagSDUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{sdEPDescribeCmd, sdEPListCmd} {
		c.Flags().StringVar(&flagSDFormat, "format", "", "Output format")
	}
	sdEndpointsCmd.AddCommand(epAll...)
	serviceDirectoryCmd.AddCommand(sdEndpointsCmd)

	rootCmd.AddCommand(serviceDirectoryCmd)
}

// --- locations impl ---

func runSDLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(sdLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormatted(loc, flagSDFormat)
}

func runSDLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	if flagSDFormat != "" {
		return emitFormatted(resp.Locations, flagSDFormat)
	}
	fmt.Printf("%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Printf("%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

// --- namespaces impl ---

func sdNSName(id, project, location string) string {
	return sdChild("namespaces", id, sdLocationParent(project, location))
}

func runSDNSCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ns := &servicedirectory.Namespace{}
	if err := loadYAMLOrJSONInto(flagSDConfigFile, ns); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Create(sdLocationParent(project, flagSDLocation), ns).
		NamespaceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}
	return emitFormatted(got, "")
}

func runSDNSDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Namespaces.Delete(sdNSName(args[0], project, flagSDLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting namespace: %w", err)
	}
	fmt.Printf("Deleted namespace [%s].\n", args[0])
	return nil
}

func runSDNSDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Get(sdNSName(args[0], project, flagSDLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing namespace: %w", err)
	}
	return emitFormatted(got, flagSDFormat)
}

func runSDNSList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Namespaces.List(sdLocationParent(project, flagSDLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing namespaces: %w", err)
	}
	if flagSDFormat != "" {
		return emitFormatted(resp.Namespaces, flagSDFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, n := range resp.Namespaces {
		fmt.Println(path.Base(n.Name))
	}
	return nil
}

func runSDNSUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ns := &servicedirectory.Namespace{}
	if err := loadYAMLOrJSONInto(flagSDConfigFile, ns); err != nil {
		return err
	}
	mask := flagSDUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(ns))
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Patch(sdNSName(args[0], project, flagSDLocation), ns).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating namespace: %w", err)
	}
	return emitFormatted(got, "")
}

func runSDNSGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Namespaces.GetIamPolicy(sdNSName(args[0], project, flagSDLocation), &servicedirectory.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagSDFormat)
}

func runSDNSSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &servicedirectory.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.SetIamPolicy(sdNSName(args[0], project, flagSDLocation), &servicedirectory.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runSDNSAddIam(cmd *cobra.Command, args []string) error {
	return sdNSModifyIam(args[0], func(p *servicedirectory.Policy) {
		for _, b := range p.Bindings {
			if b.Role == flagSDIamRole {
				for _, m := range b.Members {
					if m == flagSDIamMember {
						return
					}
				}
				b.Members = append(b.Members, flagSDIamMember)
				return
			}
		}
		p.Bindings = append(p.Bindings, &servicedirectory.Binding{Role: flagSDIamRole, Members: []string{flagSDIamMember}})
	})
}

func runSDNSRemoveIam(cmd *cobra.Command, args []string) error {
	return sdNSModifyIam(args[0], func(p *servicedirectory.Policy) {
		for _, b := range p.Bindings {
			if b.Role != flagSDIamRole {
				continue
			}
			out := b.Members[:0]
			for _, m := range b.Members {
				if m != flagSDIamMember {
					out = append(out, m)
				}
			}
			b.Members = out
		}
	})
}

func sdNSModifyIam(name string, mutate func(*servicedirectory.Policy)) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := sdNSName(name, project, flagSDLocation)
	policy, err := svc.Projects.Locations.Namespaces.GetIamPolicy(resource, &servicedirectory.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.Namespaces.SetIamPolicy(resource, &servicedirectory.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

// --- services impl ---

func sdSvcParent(project, location, ns string) string {
	return fmt.Sprintf("%s/namespaces/%s", sdLocationParent(project, location), ns)
}

func sdSvcName(id, project, location, ns string) string {
	return sdChild("services", id, sdSvcParent(project, location, ns))
}

func runSDSvcCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	s := &servicedirectory.Service{}
	if err := loadYAMLOrJSONInto(flagSDConfigFile, s); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Services.Create(sdSvcParent(project, flagSDLocation, flagSDNamespace), s).
		ServiceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service: %w", err)
	}
	return emitFormatted(got, "")
}

func runSDSvcDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Namespaces.Services.Delete(sdSvcName(args[0], project, flagSDLocation, flagSDNamespace)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	fmt.Printf("Deleted service [%s].\n", args[0])
	return nil
}

func runSDSvcDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Services.Get(sdSvcName(args[0], project, flagSDLocation, flagSDNamespace)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	return emitFormatted(got, flagSDFormat)
}

func runSDSvcList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Namespaces.Services.List(sdSvcParent(project, flagSDLocation, flagSDNamespace)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing services: %w", err)
	}
	if flagSDFormat != "" {
		return emitFormatted(resp.Services, flagSDFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, s := range resp.Services {
		fmt.Println(path.Base(s.Name))
	}
	return nil
}

func runSDSvcUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	s := &servicedirectory.Service{}
	if err := loadYAMLOrJSONInto(flagSDConfigFile, s); err != nil {
		return err
	}
	mask := flagSDUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(s))
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Services.Patch(sdSvcName(args[0], project, flagSDLocation, flagSDNamespace), s).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating service: %w", err)
	}
	return emitFormatted(got, "")
}

func runSDSvcResolve(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Services.Resolve(sdSvcName(args[0], project, flagSDLocation, flagSDNamespace), &servicedirectory.ResolveServiceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resolving service: %w", err)
	}
	return emitFormatted(got, flagSDFormat)
}

// --- endpoints impl ---

func sdEPParent(project, location, ns, service string) string {
	return fmt.Sprintf("%s/services/%s", sdSvcParent(project, location, ns), service)
}

func sdEPName(id, project, location, ns, service string) string {
	return sdChild("endpoints", id, sdEPParent(project, location, ns, service))
}

func runSDEPCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	e := &servicedirectory.Endpoint{}
	if err := loadYAMLOrJSONInto(flagSDConfigFile, e); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Services.Endpoints.Create(sdEPParent(project, flagSDLocation, flagSDNamespace, flagSDService), e).
		EndpointId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating endpoint: %w", err)
	}
	return emitFormatted(got, "")
}

func runSDEPDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Namespaces.Services.Endpoints.Delete(sdEPName(args[0], project, flagSDLocation, flagSDNamespace, flagSDService)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting endpoint: %w", err)
	}
	fmt.Printf("Deleted endpoint [%s].\n", args[0])
	return nil
}

func runSDEPDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Services.Endpoints.Get(sdEPName(args[0], project, flagSDLocation, flagSDNamespace, flagSDService)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing endpoint: %w", err)
	}
	return emitFormatted(got, flagSDFormat)
}

func runSDEPList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Namespaces.Services.Endpoints.List(sdEPParent(project, flagSDLocation, flagSDNamespace, flagSDService)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing endpoints: %w", err)
	}
	if flagSDFormat != "" {
		return emitFormatted(resp.Endpoints, flagSDFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "ADDRESS")
	for _, e := range resp.Endpoints {
		fmt.Printf("%-40s %s\n", path.Base(e.Name), e.Address)
	}
	return nil
}

func runSDEPUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	e := &servicedirectory.Endpoint{}
	if err := loadYAMLOrJSONInto(flagSDConfigFile, e); err != nil {
		return err
	}
	mask := flagSDUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(e))
	}
	ctx := context.Background()
	svc, err := gcp.ServiceDirectoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Namespaces.Services.Endpoints.Patch(sdEPName(args[0], project, flagSDLocation, flagSDNamespace, flagSDService), e).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating endpoint: %w", err)
	}
	return emitFormatted(got, "")
}
