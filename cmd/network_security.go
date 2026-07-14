package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networksecurity "google.golang.org/api/networksecurity/v1"
)

// --- gcloud network-security (#363, #819-#846) ---
//
// Every subgroup here is backed by the google.golang.org/api/networksecurity
// service (v1 primarily, v1beta1 for secure-access-connect which only exists
// in beta). The layout mirrors cmd/agent_registry.go and cmd/metastore.go:
//
//   - a per-group struct of cobra commands
//   - shared --location / --organization / --config-file / --async / --format
//     flags provided by a small builder
//   - LRO create/patch/delete calls block by default and wait via nsWaitOp
//   - update flows accept a JSON or YAML body via --config-file and derive
//     the update mask automatically when --update-mask is not supplied.

var networkSecurityCmd = &cobra.Command{Use: "network-security", Short: "Manage Network Security"}

// shared network-security flags
var (
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
	flagNSPageSize int64
)

// nsProjectParent returns projects/PROJECT/locations/LOCATION using the current
// project and --location.
func nsProjectParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if flagNSLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagNSLocation), nil
}

// nsOrgParent returns organizations/ORG/locations/LOCATION using --organization
// and --location.
func nsOrgParent() (string, error) {
	if flagNSOrganization == "" {
		return "", fmt.Errorf("--organization is required")
	}
	if flagNSLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("organizations/%s/locations/%s", flagNSOrganization, flagNSLocation), nil
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

func nsFinishOp(ctx context.Context, svc *networksecurity.Service, op *networksecurity.Operation, verb, name string) error {
	if flagNSAsync {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := nsWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

// addNSLocationFlag adds --location to each command.
func addNSLocationFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagNSLocation, "location", "", "Location (required)")
	}
}

// addNSOrgFlags adds --organization + --location to each command.
func addNSOrgFlags(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagNSOrganization, "organization", "", "Organization ID (required)")
		c.Flags().StringVar(&flagNSLocation, "location", "", "Location (required)")
	}
}

// addNSFormatFlag adds --format to each command.
func addNSFormatFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagNSFormat, "format", "", "Output format")
	}
}

// addNSFilterFlag adds --filter to each command.
func addNSFilterFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagNSFilter, "filter", "", "Server-side list filter")
	}
}

// addNSAsyncFlag adds --async to each command.
func addNSAsyncFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().BoolVar(&flagNSAsync, "async", false, "Do not wait for the operation to complete")
	}
}

// addNSRequestIDFlag adds --request-id to each command.
func addNSRequestIDFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagNSRequestID, "request-id", "", "Optional idempotency request ID")
	}
}

// addNSCreateConfigFlag adds a required --config-file flag.
func addNSCreateConfigFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagNSConfigFile, "config-file", "", "Path to a JSON/YAML file with the resource body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
}

// addNSUpdateConfigFlag adds a required --config-file plus --update-mask.
func addNSUpdateConfigFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagNSConfigFile, "config-file", "", "Path to a JSON/YAML file with the resource body (required)")
		_ = c.MarkFlagRequired("config-file")
		c.Flags().StringVar(&flagNSUpdateMask, "update-mask", "", "Comma-separated list of fields to update (defaults to every populated field)")
	}
}

// nsBasename returns the leaf of a resource name for listing display.
func nsBasename(name string) string { return path.Base(name) }

// nsResolveMask picks --update-mask if given, otherwise every top-level field
// present in body.
func nsResolveMask(body any) string {
	if flagNSUpdateMask != "" {
		return flagNSUpdateMask
	}
	return joinMask(nonEmptyJSONFields(body))
}

// --- operations (issue 833) ---

var networkSecurityOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Network Security operations"}

var (
	nsOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: runNSOpCancel,
	}
	nsOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an operation",
		Args: cobra.ExactArgs(1), RunE: runNSOpDelete,
	}
	nsOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: runNSOpDescribe,
	}
	nsOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a location",
		Args: cobra.NoArgs, RunE: runNSOpList,
	}
)

func runNSOpCancel(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name, err := nsResolveOperationName(args[0])
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

func runNSOpDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name, err := nsResolveOperationName(args[0])
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

func runNSOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name, err := nsResolveOperationName(args[0])
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
	return emitFormatted(op, flagNSFormat)
}

func runNSOpList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var parent string
	if flagNSOrganization != "" {
		p, err := nsOrgParent()
		if err != nil {
			return err
		}
		parent = p
	} else {
		p, err := nsProjectParent()
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
			if flagNSFilter != "" {
				call = call.Filter(flagNSFilter)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err = call.Do()
		} else {
			call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
			if flagNSFilter != "" {
				call = call.Filter(flagNSFilter)
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
	if flagNSFormat != "" {
		return emitFormatted(all, flagNSFormat)
	}
	fmt.Printf("%-60s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Printf("%-60s %v\n", nsBasename(o.Name), o.Done)
	}
	return nil
}

// nsResolveOperationName accepts either a fully qualified operation name or a
// short ID relative to the current --organization / --project + --location.
func nsResolveOperationName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") || strings.HasPrefix(id, "organizations/") {
		return id, nil
	}
	if flagNSOrganization != "" {
		parent, err := nsOrgParent()
		if err != nil {
			return "", err
		}
		return nsChild(parent, "operations", id), nil
	}
	parent, err := nsProjectParent()
	if err != nil {
		return "", err
	}
	return nsChild(parent, "operations", id), nil
}

func init() {
	// operations
	for _, c := range []*cobra.Command{nsOpCancelCmd, nsOpDeleteCmd, nsOpDescribeCmd, nsOpListCmd} {
		c.Flags().StringVar(&flagNSLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&flagNSOrganization, "organization", "", "Organization ID (for org-scoped operations)")
	}
	addNSFormatFlag(nsOpDescribeCmd, nsOpListCmd)
	addNSFilterFlag(nsOpListCmd)
	networkSecurityOperationsCmd.AddCommand(nsOpCancelCmd, nsOpDeleteCmd, nsOpDescribeCmd, nsOpListCmd)
	networkSecurityCmd.AddCommand(networkSecurityOperationsCmd)

	registerNSAddressGroups(networkSecurityCmd)
	registerNSPolicies(networkSecurityCmd)
	registerNSFirewall(networkSecurityCmd)
	registerNSInterceptMirroring(networkSecurityCmd)
	registerNSProfiles(networkSecurityCmd)
	registerNSSecureAccessConnect(networkSecurityCmd)

	rootCmd.AddCommand(networkSecurityCmd)
}
