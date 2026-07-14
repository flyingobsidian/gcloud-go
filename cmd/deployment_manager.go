package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	deploymentmanager "google.golang.org/api/deploymentmanager/v2"
	deploymentmanagerbeta "google.golang.org/api/deploymentmanager/v2beta"
)

// --- gcloud deployment-manager (#328, #855-#859) ---

var deploymentManagerCmd = &cobra.Command{Use: "deployment-manager", Short: "Manage Deployment Manager"}

var (
	flagDPMFormat        string
	flagDPMFilter        string
	flagDPMConfigFile    string
	flagDPMDeployment    string
	flagDPMDescription   string
	flagDPMPreview       bool
	flagDPMFingerprint   string
	flagDPMCreatePolicy  string
	flagDPMDeletePolicy  string
	flagDPMAsync         bool
	flagDPMPollIntervalS int
)

const dpmDefaultPollSeconds = 5

func dpmResolveProject() (string, error) {
	return resolveProject()
}

func dpmWaitOp(ctx context.Context, svc *deploymentmanager.Service, project string, op *deploymentmanager.Operation) (*deploymentmanager.Operation, error) {
	interval := time.Duration(flagDPMPollIntervalS) * time.Second
	if interval <= 0 {
		interval = dpmDefaultPollSeconds * time.Second
	}
	for op.Status != "DONE" {
		time.Sleep(interval)
		got, err := svc.Operations.Get(project, op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil && len(op.Error.Errors) > 0 {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Errors[0].Message)
	}
	return op, nil
}

func dpmFinishOp(ctx context.Context, svc *deploymentmanager.Service, project string, op *deploymentmanager.Operation, verb, name string) error {
	if flagDPMAsync {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, flagDPMFormat)
	}
	final, err := dpmWaitOp(ctx, svc, project, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	return emitFormatted(final, flagDPMFormat)
}

// --- deployments ---

var deploymentManagerDeploymentsCmd = &cobra.Command{Use: "deployments", Short: "Manage deployments"}

var (
	dpmDepCreateCmd = &cobra.Command{
		Use: "create DEPLOYMENT", Short: "Create a deployment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDPMDepCreate,
	}
	dpmDepDeleteCmd = &cobra.Command{
		Use: "delete DEPLOYMENT", Short: "Delete a deployment",
		Args: cobra.ExactArgs(1), RunE: runDPMDepDelete,
	}
	dpmDepDescribeCmd = &cobra.Command{
		Use: "describe DEPLOYMENT", Short: "Describe a deployment",
		Args: cobra.ExactArgs(1), RunE: runDPMDepDescribe,
	}
	dpmDepListCmd = &cobra.Command{
		Use: "list", Short: "List deployments in the current project",
		Args: cobra.NoArgs, RunE: runDPMDepList,
	}
	dpmDepUpdateCmd = &cobra.Command{
		Use: "update DEPLOYMENT", Short: "Update a deployment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDPMDepUpdate,
	}
	dpmDepStopCmd = &cobra.Command{
		Use: "stop DEPLOYMENT", Short: "Stop a deployment update",
		Args: cobra.ExactArgs(1), RunE: runDPMDepStop,
	}
	dpmDepCancelPreviewCmd = &cobra.Command{
		Use: "cancel-preview DEPLOYMENT", Short: "Cancel a preview of a deployment update",
		Args: cobra.ExactArgs(1), RunE: runDPMDepCancelPreview,
	}
)

// --- manifests ---

var deploymentManagerManifestsCmd = &cobra.Command{Use: "manifests", Short: "Manage manifests"}

var (
	dpmManifestDescribeCmd = &cobra.Command{
		Use: "describe MANIFEST", Short: "Describe a manifest for a deployment",
		Args: cobra.ExactArgs(1), RunE: runDPMManifestDescribe,
	}
	dpmManifestListCmd = &cobra.Command{
		Use: "list", Short: "List manifests for a deployment",
		Args: cobra.NoArgs, RunE: runDPMManifestList,
	}
)

// --- operations ---

var deploymentManagerOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage operations"}

var (
	dpmOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: runDPMOpDescribe,
	}
	dpmOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in the current project",
		Args: cobra.NoArgs, RunE: runDPMOpList,
	}
	dpmOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for an operation to complete",
		Args: cobra.ExactArgs(1), RunE: runDPMOpWait,
	}
)

// --- resources ---

var deploymentManagerResourcesCmd = &cobra.Command{Use: "resources", Short: "Manage resources"}

var (
	dpmResDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe a resource for a deployment",
		Args: cobra.ExactArgs(1), RunE: runDPMResourceDescribe,
	}
	dpmResListCmd = &cobra.Command{
		Use: "list", Short: "List resources for a deployment",
		Args: cobra.NoArgs, RunE: runDPMResourceList,
	}
)

// --- types ---

var deploymentManagerTypesCmd = &cobra.Command{Use: "types", Short: "Manage types"}

var (
	dpmTypeListCmd = &cobra.Command{
		Use: "list", Short: "List types in the current project",
		Args: cobra.NoArgs, RunE: runDPMTypeList,
	}
	dpmTypeProvidersCmd = &cobra.Command{
		Use: "providers", Short: "List type providers in the current project (v2beta)",
		Args: cobra.NoArgs, RunE: runDPMTypeProviders,
	}
)

func init() {
	addDPMFormat := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagDPMFormat, "format", "", "Output format")
		}
	}
	addDPMFilter := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagDPMFilter, "filter", "", "Server-side list filter")
		}
	}
	addDPMAsync := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().BoolVar(&flagDPMAsync, "async", false, "Do not wait for the operation to complete")
			c.Flags().IntVar(&flagDPMPollIntervalS, "poll-interval", dpmDefaultPollSeconds, "Poll interval in seconds while waiting for operations")
		}
	}
	addDPMDeployment := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagDPMDeployment, "deployment", "", "Parent deployment (required)")
			_ = c.MarkFlagRequired("deployment")
		}
	}

	// deployments
	dpmDepCreateCmd.Flags().StringVar(&flagDPMConfigFile, "config-file", "", "Path to a JSON/YAML file with the Deployment body (required)")
	_ = dpmDepCreateCmd.MarkFlagRequired("config-file")
	dpmDepCreateCmd.Flags().StringVar(&flagDPMDescription, "description", "", "Optional deployment description (overrides body)")
	dpmDepCreateCmd.Flags().BoolVar(&flagDPMPreview, "preview", false, "Create the deployment in preview mode")
	dpmDepCreateCmd.Flags().StringVar(&flagDPMCreatePolicy, "create-policy", "", "Create policy (CREATE_OR_ACQUIRE|ACQUIRE)")

	dpmDepUpdateCmd.Flags().StringVar(&flagDPMConfigFile, "config-file", "", "Path to a JSON/YAML file with the Deployment body (required)")
	_ = dpmDepUpdateCmd.MarkFlagRequired("config-file")
	dpmDepUpdateCmd.Flags().StringVar(&flagDPMDescription, "description", "", "Optional deployment description (overrides body)")
	dpmDepUpdateCmd.Flags().BoolVar(&flagDPMPreview, "preview", false, "Update the deployment in preview mode")
	dpmDepUpdateCmd.Flags().StringVar(&flagDPMCreatePolicy, "create-policy", "", "Create policy (CREATE_OR_ACQUIRE|ACQUIRE)")
	dpmDepUpdateCmd.Flags().StringVar(&flagDPMDeletePolicy, "delete-policy", "", "Delete policy (DELETE|ABANDON)")

	dpmDepDeleteCmd.Flags().StringVar(&flagDPMDeletePolicy, "delete-policy", "", "Delete policy (DELETE|ABANDON)")

	dpmDepStopCmd.Flags().StringVar(&flagDPMFingerprint, "fingerprint", "", "Optional fingerprint of the deployment (fetched if empty)")
	dpmDepCancelPreviewCmd.Flags().StringVar(&flagDPMFingerprint, "fingerprint", "", "Optional fingerprint of the deployment (fetched if empty)")

	addDPMFormat(dpmDepCreateCmd, dpmDepDeleteCmd, dpmDepDescribeCmd, dpmDepListCmd, dpmDepUpdateCmd, dpmDepStopCmd, dpmDepCancelPreviewCmd)
	addDPMFilter(dpmDepListCmd)
	addDPMAsync(dpmDepCreateCmd, dpmDepDeleteCmd, dpmDepUpdateCmd, dpmDepStopCmd, dpmDepCancelPreviewCmd)
	deploymentManagerDeploymentsCmd.AddCommand(
		dpmDepCreateCmd, dpmDepDeleteCmd, dpmDepDescribeCmd, dpmDepListCmd,
		dpmDepUpdateCmd, dpmDepStopCmd, dpmDepCancelPreviewCmd)
	deploymentManagerCmd.AddCommand(deploymentManagerDeploymentsCmd)

	// manifests
	addDPMFormat(dpmManifestDescribeCmd, dpmManifestListCmd)
	addDPMFilter(dpmManifestListCmd)
	addDPMDeployment(dpmManifestDescribeCmd, dpmManifestListCmd)
	deploymentManagerManifestsCmd.AddCommand(dpmManifestDescribeCmd, dpmManifestListCmd)
	deploymentManagerCmd.AddCommand(deploymentManagerManifestsCmd)

	// operations
	addDPMFormat(dpmOpDescribeCmd, dpmOpListCmd, dpmOpWaitCmd)
	addDPMFilter(dpmOpListCmd)
	dpmOpWaitCmd.Flags().IntVar(&flagDPMPollIntervalS, "poll-interval", dpmDefaultPollSeconds, "Poll interval in seconds while waiting for the operation")
	deploymentManagerOperationsCmd.AddCommand(dpmOpDescribeCmd, dpmOpListCmd, dpmOpWaitCmd)
	deploymentManagerCmd.AddCommand(deploymentManagerOperationsCmd)

	// resources
	addDPMFormat(dpmResDescribeCmd, dpmResListCmd)
	addDPMFilter(dpmResListCmd)
	addDPMDeployment(dpmResDescribeCmd, dpmResListCmd)
	deploymentManagerResourcesCmd.AddCommand(dpmResDescribeCmd, dpmResListCmd)
	deploymentManagerCmd.AddCommand(deploymentManagerResourcesCmd)

	// types
	addDPMFormat(dpmTypeListCmd, dpmTypeProvidersCmd)
	addDPMFilter(dpmTypeListCmd, dpmTypeProvidersCmd)
	deploymentManagerTypesCmd.AddCommand(dpmTypeListCmd, dpmTypeProvidersCmd)
	deploymentManagerCmd.AddCommand(deploymentManagerTypesCmd)

	rootCmd.AddCommand(deploymentManagerCmd)
}

// --- deployments impl ---

func runDPMDepCreate(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	body := &deploymentmanager.Deployment{}
	if err := loadYAMLOrJSONInto(flagDPMConfigFile, body); err != nil {
		return err
	}
	body.Name = args[0]
	if flagDPMDescription != "" {
		body.Description = flagDPMDescription
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Deployments.Insert(project, body).Context(ctx)
	if flagDPMPreview {
		call = call.Preview(true)
	}
	if flagDPMCreatePolicy != "" {
		call = call.CreatePolicy(flagDPMCreatePolicy)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}
	return dpmFinishOp(ctx, svc, project, op, "Create deployment", args[0])
}

func runDPMDepDelete(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Deployments.Delete(project, args[0]).Context(ctx)
	if flagDPMDeletePolicy != "" {
		call = call.DeletePolicy(flagDPMDeletePolicy)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting deployment: %w", err)
	}
	return dpmFinishOp(ctx, svc, project, op, "Delete deployment", args[0])
}

func runDPMDepDescribe(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Deployments.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing deployment: %w", err)
	}
	return emitFormatted(got, flagDPMFormat)
}

func runDPMDepList(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*deploymentmanager.Deployment
	pageToken := ""
	for {
		call := svc.Deployments.List(project).Context(ctx)
		if flagDPMFilter != "" {
			call = call.Filter(flagDPMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing deployments: %w", err)
		}
		all = append(all, resp.Deployments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagDPMFormat != "" {
		return emitFormatted(all, flagDPMFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DESCRIPTION")
	for _, d := range all {
		fmt.Printf("%-40s %s\n", d.Name, d.Description)
	}
	return nil
}

func runDPMDepUpdate(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	body := &deploymentmanager.Deployment{}
	if err := loadYAMLOrJSONInto(flagDPMConfigFile, body); err != nil {
		return err
	}
	body.Name = args[0]
	if flagDPMDescription != "" {
		body.Description = flagDPMDescription
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if body.Fingerprint == "" {
		existing, err := svc.Deployments.Get(project, args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("fetching deployment fingerprint: %w", err)
		}
		body.Fingerprint = existing.Fingerprint
	}
	call := svc.Deployments.Update(project, args[0], body).Context(ctx)
	if flagDPMPreview {
		call = call.Preview(true)
	}
	if flagDPMCreatePolicy != "" {
		call = call.CreatePolicy(flagDPMCreatePolicy)
	}
	if flagDPMDeletePolicy != "" {
		call = call.DeletePolicy(flagDPMDeletePolicy)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating deployment: %w", err)
	}
	return dpmFinishOp(ctx, svc, project, op, "Update deployment", args[0])
}

func runDPMDepStop(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	fingerprint := flagDPMFingerprint
	if fingerprint == "" {
		existing, err := svc.Deployments.Get(project, args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("fetching deployment fingerprint: %w", err)
		}
		fingerprint = existing.Fingerprint
	}
	op, err := svc.Deployments.Stop(project, args[0], &deploymentmanager.DeploymentsStopRequest{Fingerprint: fingerprint}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping deployment: %w", err)
	}
	return dpmFinishOp(ctx, svc, project, op, "Stop deployment", args[0])
}

func runDPMDepCancelPreview(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	fingerprint := flagDPMFingerprint
	if fingerprint == "" {
		existing, err := svc.Deployments.Get(project, args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("fetching deployment fingerprint: %w", err)
		}
		fingerprint = existing.Fingerprint
	}
	op, err := svc.Deployments.CancelPreview(project, args[0], &deploymentmanager.DeploymentsCancelPreviewRequest{Fingerprint: fingerprint}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling deployment preview: %w", err)
	}
	return dpmFinishOp(ctx, svc, project, op, "Cancel preview", args[0])
}

// --- manifests impl ---

func runDPMManifestDescribe(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Manifests.Get(project, flagDPMDeployment, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing manifest: %w", err)
	}
	return emitFormatted(got, flagDPMFormat)
}

func runDPMManifestList(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*deploymentmanager.Manifest
	pageToken := ""
	for {
		call := svc.Manifests.List(project, flagDPMDeployment).Context(ctx)
		if flagDPMFilter != "" {
			call = call.Filter(flagDPMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing manifests: %w", err)
		}
		all = append(all, resp.Manifests...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagDPMFormat != "" {
		return emitFormatted(all, flagDPMFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "SELF_LINK")
	for _, m := range all {
		fmt.Printf("%-40s %s\n", m.Name, m.SelfLink)
	}
	return nil
}

// --- operations impl ---

func runDPMOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Operations.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagDPMFormat)
}

func runDPMOpList(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*deploymentmanager.Operation
	pageToken := ""
	for {
		call := svc.Operations.List(project).Context(ctx)
		if flagDPMFilter != "" {
			call = call.Filter(flagDPMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagDPMFormat != "" {
		return emitFormatted(all, flagDPMFormat)
	}
	fmt.Printf("%-40s %-15s %s\n", "NAME", "TYPE", "STATUS")
	for _, o := range all {
		fmt.Printf("%-40s %-15s %s\n", o.Name, o.OperationType, o.Status)
	}
	return nil
}

func runDPMOpWait(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Operations.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting operation: %w", err)
	}
	final, err := dpmWaitOp(ctx, svc, project, op)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagDPMFormat)
}

// --- resources impl ---

func runDPMResourceDescribe(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Resources.Get(project, flagDPMDeployment, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing resource: %w", err)
	}
	return emitFormatted(got, flagDPMFormat)
}

func runDPMResourceList(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*deploymentmanager.Resource
	pageToken := ""
	for {
		call := svc.Resources.List(project, flagDPMDeployment).Context(ctx)
		if flagDPMFilter != "" {
			call = call.Filter(flagDPMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing resources: %w", err)
		}
		all = append(all, resp.Resources...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagDPMFormat != "" {
		return emitFormatted(all, flagDPMFormat)
	}
	fmt.Printf("%-40s %-30s %s\n", "NAME", "TYPE", "STATE")
	for _, r := range all {
		state := ""
		if r.Update != nil {
			state = r.Update.State
		}
		fmt.Printf("%-40s %-30s %s\n", r.Name, r.Type, state)
	}
	return nil
}

// --- types impl ---

func runDPMTypeList(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*deploymentmanager.Type
	pageToken := ""
	for {
		call := svc.Types.List(project).Context(ctx)
		if flagDPMFilter != "" {
			call = call.Filter(flagDPMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing types: %w", err)
		}
		all = append(all, resp.Types...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagDPMFormat != "" {
		return emitFormatted(all, flagDPMFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "SELF_LINK")
	for _, t := range all {
		fmt.Printf("%-40s %s\n", t.Name, t.SelfLink)
	}
	return nil
}

func runDPMTypeProviders(cmd *cobra.Command, args []string) error {
	project, err := dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeploymentManagerBetaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*deploymentmanagerbeta.TypeProvider
	pageToken := ""
	for {
		call := svc.TypeProviders.List(project).Context(ctx)
		if flagDPMFilter != "" {
			call = call.Filter(flagDPMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing type providers: %w", err)
		}
		all = append(all, resp.TypeProviders...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagDPMFormat != "" {
		return emitFormatted(all, flagDPMFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DESCRIPTION")
	for _, tp := range all {
		fmt.Printf("%-40s %s\n", tp.Name, tp.Description)
	}
	return nil
}
