package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudiam "google.golang.org/api/iam/v1"
)

// --- gcloud iam workforce-pools (#1014) ---
//
// Workforce pools live under organizations (not projects) at
// locations/LOCATION. This surface adds the top-level pool CRUD plus the
// providers subgroup. subjects/operations sub-subgroups remain stubs for
// follow-up issues since the surface is large.

var iamWorkforcePoolsCmd = &cobra.Command{Use: "workforce-pools", Short: "Manage Workforce Identity Federation pools"}
var iamWorkforcePoolsProvidersCmd = &cobra.Command{Use: "providers", Short: "Manage workforce-pool providers"}

var (
	flagIamWpLocation     string
	flagIamWpOrganization string
	flagIamWpFormat       string
	flagIamWpConfigFile   string
	flagIamWpUpdateMask   string
	flagIamWpPageSize     int64
)

var (
	iamWpCreateCmd = &cobra.Command{
		Use: "create POOL", Short: "Create a workforce pool",
		Args: cobra.ExactArgs(1), RunE: runIamWpCreate,
	}
	iamWpDeleteCmd = &cobra.Command{
		Use: "delete POOL", Short: "Delete a workforce pool",
		Args: cobra.ExactArgs(1), RunE: runIamWpDelete,
	}
	iamWpDescribeCmd = &cobra.Command{
		Use: "describe POOL", Short: "Describe a workforce pool",
		Args: cobra.ExactArgs(1), RunE: runIamWpDescribe,
	}
	iamWpListCmd = &cobra.Command{
		Use: "list", Short: "List workforce pools in an organization's location",
		Args: cobra.NoArgs, RunE: runIamWpList,
	}
	iamWpUpdateCmd = &cobra.Command{
		Use: "update POOL", Short: "Update a workforce pool",
		Args: cobra.ExactArgs(1), RunE: runIamWpUpdate,
	}
	iamWpUndeleteCmd = &cobra.Command{
		Use: "undelete POOL", Short: "Undelete a workforce pool",
		Args: cobra.ExactArgs(1), RunE: runIamWpUndelete,
	}

	iamWpProvCreateCmd = &cobra.Command{
		Use: "create PROVIDER", Short: "Create a workforce-pool provider",
		Args: cobra.ExactArgs(1), RunE: runIamWpProvCreate,
	}
	iamWpProvDeleteCmd = &cobra.Command{
		Use: "delete PROVIDER", Short: "Delete a workforce-pool provider",
		Args: cobra.ExactArgs(1), RunE: runIamWpProvDelete,
	}
	iamWpProvDescribeCmd = &cobra.Command{
		Use: "describe PROVIDER", Short: "Describe a workforce-pool provider",
		Args: cobra.ExactArgs(1), RunE: runIamWpProvDescribe,
	}
	iamWpProvListCmd = &cobra.Command{
		Use: "list", Short: "List workforce-pool providers",
		Args: cobra.NoArgs, RunE: runIamWpProvList,
	}
	iamWpProvUpdateCmd = &cobra.Command{
		Use: "update PROVIDER", Short: "Update a workforce-pool provider",
		Args: cobra.ExactArgs(1), RunE: runIamWpProvUpdate,
	}
	iamWpProvUndeleteCmd = &cobra.Command{
		Use: "undelete PROVIDER", Short: "Undelete a workforce-pool provider",
		Args: cobra.ExactArgs(1), RunE: runIamWpProvUndelete,
	}
)

var flagIamWpPool string

func init() {
	pools := []*cobra.Command{
		iamWpCreateCmd, iamWpDeleteCmd, iamWpDescribeCmd,
		iamWpListCmd, iamWpUpdateCmd, iamWpUndeleteCmd,
	}
	providers := []*cobra.Command{
		iamWpProvCreateCmd, iamWpProvDeleteCmd, iamWpProvDescribeCmd,
		iamWpProvListCmd, iamWpProvUpdateCmd, iamWpProvUndeleteCmd,
	}
	all := append(append([]*cobra.Command{}, pools...), providers...)
	for _, c := range all {
		c.Flags().StringVar(&flagIamWpOrganization, "organization", "", "Owning organization ID (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&flagIamWpLocation, "location", "global", "Location (defaults to global)")
		c.Flags().StringVar(&flagIamWpFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{iamWpCreateCmd, iamWpUpdateCmd, iamWpProvCreateCmd, iamWpProvUpdateCmd} {
		c.Flags().StringVar(&flagIamWpConfigFile, "config-file", "", "YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	for _, c := range []*cobra.Command{iamWpUpdateCmd, iamWpProvUpdateCmd} {
		c.Flags().StringVar(&flagIamWpUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	}
	for _, c := range []*cobra.Command{iamWpListCmd, iamWpProvListCmd} {
		c.Flags().Int64Var(&flagIamWpPageSize, "page-size", 0, "Maximum results per page")
	}
	for _, c := range providers {
		c.Flags().StringVar(&flagIamWpPool, "workforce-pool", "", "Owning workforce pool (required)")
		_ = c.MarkFlagRequired("workforce-pool")
	}

	iamWorkforcePoolsProvidersCmd.AddCommand(providers...)
	iamWorkforcePoolsCmd.AddCommand(pools...)
	iamWorkforcePoolsCmd.AddCommand(iamWorkforcePoolsProvidersCmd)
	iamCmd.AddCommand(iamWorkforcePoolsCmd)
}

func iamWpLocationParent() (string, error) {
	if flagIamWpOrganization == "" {
		return "", fmt.Errorf("--organization is required")
	}
	loc := flagIamWpLocation
	if loc == "" {
		loc = "global"
	}
	return fmt.Sprintf("locations/%s", loc), nil
}

func iamWpPoolName(id string) (string, error) {
	if strings.HasPrefix(id, "locations/") {
		return id, nil
	}
	parent, err := iamWpLocationParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/workforcePools/%s", parent, id), nil
}

func iamWpProviderName(id string) (string, error) {
	if strings.HasPrefix(id, "locations/") {
		return id, nil
	}
	if flagIamWpPool == "" {
		return "", fmt.Errorf("--workforce-pool is required")
	}
	parent, err := iamWpLocationParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/workforcePools/%s/providers/%s", parent, flagIamWpPool, id), nil
}

// pools

func runIamWpCreate(cmd *cobra.Command, args []string) error {
	parent, err := iamWpLocationParent()
	if err != nil {
		return err
	}
	body := &cloudiam.WorkforcePool{}
	if err := loadYAMLOrJSONInto(flagIamWpConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Create(parent, body).WorkforcePoolId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workforce pool: %w", err)
	}
	fmt.Printf("Create workforce pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagIamWpFormat)
}

func runIamWpDelete(cmd *cobra.Command, args []string) error {
	name, err := iamWpPoolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting workforce pool: %w", err)
	}
	fmt.Printf("Delete workforce pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagIamWpFormat)
}

func runIamWpDescribe(cmd *cobra.Command, args []string) error {
	name, err := iamWpPoolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Locations.WorkforcePools.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workforce pool: %w", err)
	}
	return emitFormatted(got, flagIamWpFormat)
}

func runIamWpList(cmd *cobra.Command, args []string) error {
	parent, err := iamWpLocationParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*cloudiam.WorkforcePool
	pageToken := ""
	for {
		call := svc.Locations.WorkforcePools.List(parent).Parent("organizations/" + flagIamWpOrganization).Context(ctx)
		if flagIamWpPageSize > 0 {
			call = call.PageSize(flagIamWpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing workforce pools: %w", err)
		}
		all = append(all, resp.WorkforcePools...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagIamWpFormat)
}

func runIamWpUpdate(cmd *cobra.Command, args []string) error {
	name, err := iamWpPoolName(args[0])
	if err != nil {
		return err
	}
	body := &cloudiam.WorkforcePool{}
	if err := loadYAMLOrJSONInto(flagIamWpConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagIamWpUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Locations.WorkforcePools.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating workforce pool: %w", err)
	}
	fmt.Printf("Update workforce pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagIamWpFormat)
}

func runIamWpUndelete(cmd *cobra.Command, args []string) error {
	name, err := iamWpPoolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Undelete(name, &cloudiam.UndeleteWorkforcePoolRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting workforce pool: %w", err)
	}
	fmt.Printf("Undelete workforce pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagIamWpFormat)
}

// providers

func runIamWpProvCreate(cmd *cobra.Command, args []string) error {
	parent, err := iamWpLocationParent()
	if err != nil {
		return err
	}
	poolParent := fmt.Sprintf("%s/workforcePools/%s", parent, flagIamWpPool)
	body := &cloudiam.WorkforcePoolProvider{}
	if err := loadYAMLOrJSONInto(flagIamWpConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Providers.Create(poolParent, body).WorkforcePoolProviderId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating provider: %w", err)
	}
	fmt.Printf("Create workforce-pool provider [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagIamWpFormat)
}

func runIamWpProvDelete(cmd *cobra.Command, args []string) error {
	name, err := iamWpProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Providers.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting provider: %w", err)
	}
	fmt.Printf("Delete workforce-pool provider [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagIamWpFormat)
}

func runIamWpProvDescribe(cmd *cobra.Command, args []string) error {
	name, err := iamWpProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Locations.WorkforcePools.Providers.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing provider: %w", err)
	}
	return emitFormatted(got, flagIamWpFormat)
}

func runIamWpProvList(cmd *cobra.Command, args []string) error {
	parent, err := iamWpLocationParent()
	if err != nil {
		return err
	}
	poolParent := fmt.Sprintf("%s/workforcePools/%s", parent, flagIamWpPool)
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*cloudiam.WorkforcePoolProvider
	pageToken := ""
	for {
		call := svc.Locations.WorkforcePools.Providers.List(poolParent).Context(ctx)
		if flagIamWpPageSize > 0 {
			call = call.PageSize(flagIamWpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing providers: %w", err)
		}
		all = append(all, resp.WorkforcePoolProviders...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagIamWpFormat)
}

func runIamWpProvUpdate(cmd *cobra.Command, args []string) error {
	name, err := iamWpProviderName(args[0])
	if err != nil {
		return err
	}
	body := &cloudiam.WorkforcePoolProvider{}
	if err := loadYAMLOrJSONInto(flagIamWpConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagIamWpUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Locations.WorkforcePools.Providers.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating provider: %w", err)
	}
	fmt.Printf("Update workforce-pool provider [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagIamWpFormat)
}

func runIamWpProvUndelete(cmd *cobra.Command, args []string) error {
	name, err := iamWpProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Providers.Undelete(name, &cloudiam.UndeleteWorkforcePoolProviderRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting provider: %w", err)
	}
	fmt.Printf("Undelete workforce-pool provider [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagIamWpFormat)
}
