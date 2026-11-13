package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	deploymentmanager "google.golang.org/api/deploymentmanager/v2"
	deploymentmanagerbeta "google.golang.org/api/deploymentmanager/v2beta"
)

// --- gcloud deployment-manager (#328, #855-#859) ---

const dpmDefaultPollSeconds = 5

func (s *deploymentManagerCmds) dpmResolveProject() (string, error) {
	return s.inv.project()
}

func (s *deploymentManagerCmds) dpmWaitOp(ctx context.Context, svc *deploymentmanager.Service, project string, op *deploymentmanager.Operation) (*deploymentmanager.Operation, error) {
	interval := time.Duration(s.flagDPMPollIntervalS) * time.Second
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

func (s *deploymentManagerCmds) dpmFinishOp(ctx context.Context, svc *deploymentmanager.Service, project string, op *deploymentmanager.Operation, verb, name string) error {
	if s.flagDPMAsync {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagDPMFormat)
	}
	final, err := s.dpmWaitOp(ctx, svc, project, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, name)
	return emitFormattedTo(s.inv.stdout, final, s.flagDPMFormat)
}

// --- deployments ---

// --- manifests ---

// --- operations ---

// --- resources ---

// --- types ---

func (s *deploymentManagerCmds) initDeploymentManager() {
	addDPMFormat := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagDPMFormat, "format", "", "Output format")
		}
	}
	addDPMFilter := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagDPMFilter, "filter", "", "Server-side list filter")
		}
	}
	addDPMAsync := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().BoolVar(&s.flagDPMAsync, "async", false, "Do not wait for the operation to complete")
			c.Flags().IntVar(&s.flagDPMPollIntervalS, "poll-interval", dpmDefaultPollSeconds, "Poll interval in seconds while waiting for operations")
		}
	}
	addDPMDeployment := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagDPMDeployment, "deployment", "", "Parent deployment (required)")
			_ = c.MarkFlagRequired("deployment")
		}
	}

	// deployments
	s.dpmDepCreateCmd.Flags().StringVar(&s.flagDPMConfigFile, "config-file", "", "Path to a JSON/YAML file with the Deployment body (required)")
	_ = s.dpmDepCreateCmd.MarkFlagRequired("config-file")
	s.dpmDepCreateCmd.Flags().StringVar(&s.flagDPMDescription, "description", "", "Optional deployment description (overrides body)")
	s.dpmDepCreateCmd.Flags().BoolVar(&s.flagDPMPreview, "preview", false, "Create the deployment in preview mode")
	s.dpmDepCreateCmd.Flags().StringVar(&s.flagDPMCreatePolicy, "create-policy", "", "Create policy (CREATE_OR_ACQUIRE|ACQUIRE)")

	s.dpmDepUpdateCmd.Flags().StringVar(&s.flagDPMConfigFile, "config-file", "", "Path to a JSON/YAML file with the Deployment body (required)")
	_ = s.dpmDepUpdateCmd.MarkFlagRequired("config-file")
	s.dpmDepUpdateCmd.Flags().StringVar(&s.flagDPMDescription, "description", "", "Optional deployment description (overrides body)")
	s.dpmDepUpdateCmd.Flags().BoolVar(&s.flagDPMPreview, "preview", false, "Update the deployment in preview mode")
	s.dpmDepUpdateCmd.Flags().StringVar(&s.flagDPMCreatePolicy, "create-policy", "", "Create policy (CREATE_OR_ACQUIRE|ACQUIRE)")
	s.dpmDepUpdateCmd.Flags().StringVar(&s.flagDPMDeletePolicy, "delete-policy", "", "Delete policy (DELETE|ABANDON)")

	s.dpmDepDeleteCmd.Flags().StringVar(&s.flagDPMDeletePolicy, "delete-policy", "", "Delete policy (DELETE|ABANDON)")

	s.dpmDepStopCmd.Flags().StringVar(&s.flagDPMFingerprint, "fingerprint", "", "Optional fingerprint of the deployment (fetched if empty)")
	s.dpmDepCancelPreviewCmd.Flags().StringVar(&s.flagDPMFingerprint, "fingerprint", "", "Optional fingerprint of the deployment (fetched if empty)")

	addDPMFormat(s.dpmDepCreateCmd, s.dpmDepDeleteCmd, s.dpmDepDescribeCmd, s.dpmDepListCmd, s.dpmDepUpdateCmd, s.dpmDepStopCmd, s.dpmDepCancelPreviewCmd)
	addDPMFilter(s.dpmDepListCmd)
	addDPMAsync(s.dpmDepCreateCmd, s.dpmDepDeleteCmd, s.dpmDepUpdateCmd, s.dpmDepStopCmd, s.dpmDepCancelPreviewCmd)
	s.deploymentManagerDeploymentsCmd.AddCommand(
		s.dpmDepCreateCmd, s.dpmDepDeleteCmd, s.dpmDepDescribeCmd, s.dpmDepListCmd,
		s.dpmDepUpdateCmd, s.dpmDepStopCmd, s.dpmDepCancelPreviewCmd)
	s.deploymentManagerCmd.AddCommand(s.deploymentManagerDeploymentsCmd)

	// manifests
	addDPMFormat(s.dpmManifestDescribeCmd, s.dpmManifestListCmd)
	addDPMFilter(s.dpmManifestListCmd)
	addDPMDeployment(s.dpmManifestDescribeCmd, s.dpmManifestListCmd)
	s.deploymentManagerManifestsCmd.AddCommand(s.dpmManifestDescribeCmd, s.dpmManifestListCmd)
	s.deploymentManagerCmd.AddCommand(s.deploymentManagerManifestsCmd)

	// operations
	addDPMFormat(s.dpmOpDescribeCmd, s.dpmOpListCmd, s.dpmOpWaitCmd)
	addDPMFilter(s.dpmOpListCmd)
	s.dpmOpWaitCmd.Flags().IntVar(&s.flagDPMPollIntervalS, "poll-interval", dpmDefaultPollSeconds, "Poll interval in seconds while waiting for the operation")
	s.deploymentManagerOperationsCmd.AddCommand(s.dpmOpDescribeCmd, s.dpmOpListCmd, s.dpmOpWaitCmd)
	s.deploymentManagerCmd.AddCommand(s.deploymentManagerOperationsCmd)

	// resources
	addDPMFormat(s.dpmResDescribeCmd, s.dpmResListCmd)
	addDPMFilter(s.dpmResListCmd)
	addDPMDeployment(s.dpmResDescribeCmd, s.dpmResListCmd)
	s.deploymentManagerResourcesCmd.AddCommand(s.dpmResDescribeCmd, s.dpmResListCmd)
	s.deploymentManagerCmd.AddCommand(s.deploymentManagerResourcesCmd)

	// types
	addDPMFormat(s.dpmTypeListCmd, s.dpmTypeProvidersCmd)
	addDPMFilter(s.dpmTypeListCmd, s.dpmTypeProvidersCmd)
	s.deploymentManagerTypesCmd.AddCommand(s.dpmTypeListCmd, s.dpmTypeProvidersCmd)
	s.deploymentManagerCmd.AddCommand(s.deploymentManagerTypesCmd)

}

// --- deployments impl ---

func (s *deploymentManagerCmds) runDPMDepCreate(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	body := &deploymentmanager.Deployment{}
	if err := loadYAMLOrJSONInto(s.flagDPMConfigFile, body); err != nil {
		return err
	}
	body.Name = args[0]
	if s.flagDPMDescription != "" {
		body.Description = s.flagDPMDescription
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
	if err != nil {
		return err
	}
	call := svc.Deployments.Insert(project, body).Context(ctx)
	if s.flagDPMPreview {
		call = call.Preview(true)
	}
	if s.flagDPMCreatePolicy != "" {
		call = call.CreatePolicy(s.flagDPMCreatePolicy)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}
	return s.dpmFinishOp(ctx, svc, project, op, "Create deployment", args[0])
}

func (s *deploymentManagerCmds) runDPMDepDelete(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
	if err != nil {
		return err
	}
	call := svc.Deployments.Delete(project, args[0]).Context(ctx)
	if s.flagDPMDeletePolicy != "" {
		call = call.DeletePolicy(s.flagDPMDeletePolicy)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting deployment: %w", err)
	}
	return s.dpmFinishOp(ctx, svc, project, op, "Delete deployment", args[0])
}

func (s *deploymentManagerCmds) runDPMDepDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Deployments.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing deployment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDPMFormat)
}

func (s *deploymentManagerCmds) runDPMDepList(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
	if err != nil {
		return err
	}
	var all []*deploymentmanager.Deployment
	pageToken := ""
	for {
		call := svc.Deployments.List(project).Context(ctx)
		if s.flagDPMFilter != "" {
			call = call.Filter(s.flagDPMFilter)
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
	if s.flagDPMFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagDPMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "DESCRIPTION")
	for _, d := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", d.Name, d.Description)
	}
	return nil
}

func (s *deploymentManagerCmds) runDPMDepUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	body := &deploymentmanager.Deployment{}
	if err := loadYAMLOrJSONInto(s.flagDPMConfigFile, body); err != nil {
		return err
	}
	body.Name = args[0]
	if s.flagDPMDescription != "" {
		body.Description = s.flagDPMDescription
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
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
	if s.flagDPMPreview {
		call = call.Preview(true)
	}
	if s.flagDPMCreatePolicy != "" {
		call = call.CreatePolicy(s.flagDPMCreatePolicy)
	}
	if s.flagDPMDeletePolicy != "" {
		call = call.DeletePolicy(s.flagDPMDeletePolicy)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating deployment: %w", err)
	}
	return s.dpmFinishOp(ctx, svc, project, op, "Update deployment", args[0])
}

func (s *deploymentManagerCmds) runDPMDepStop(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
	if err != nil {
		return err
	}
	fingerprint := s.flagDPMFingerprint
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
	return s.dpmFinishOp(ctx, svc, project, op, "Stop deployment", args[0])
}

func (s *deploymentManagerCmds) runDPMDepCancelPreview(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
	if err != nil {
		return err
	}
	fingerprint := s.flagDPMFingerprint
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
	return s.dpmFinishOp(ctx, svc, project, op, "Cancel preview", args[0])
}

// --- manifests impl ---

func (s *deploymentManagerCmds) runDPMManifestDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Manifests.Get(project, s.flagDPMDeployment, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing manifest: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDPMFormat)
}

func (s *deploymentManagerCmds) runDPMManifestList(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
	if err != nil {
		return err
	}
	var all []*deploymentmanager.Manifest
	pageToken := ""
	for {
		call := svc.Manifests.List(project, s.flagDPMDeployment).Context(ctx)
		if s.flagDPMFilter != "" {
			call = call.Filter(s.flagDPMFilter)
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
	if s.flagDPMFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagDPMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "SELF_LINK")
	for _, m := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", m.Name, m.SelfLink)
	}
	return nil
}

// --- operations impl ---

func (s *deploymentManagerCmds) runDPMOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Operations.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDPMFormat)
}

func (s *deploymentManagerCmds) runDPMOpList(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
	if err != nil {
		return err
	}
	var all []*deploymentmanager.Operation
	pageToken := ""
	for {
		call := svc.Operations.List(project).Context(ctx)
		if s.flagDPMFilter != "" {
			call = call.Filter(s.flagDPMFilter)
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
	if s.flagDPMFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagDPMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %-15s %s\n", "NAME", "TYPE", "STATUS")
	for _, o := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %-15s %s\n", o.Name, o.OperationType, o.Status)
	}
	return nil
}

func (s *deploymentManagerCmds) runDPMOpWait(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Operations.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting operation: %w", err)
	}
	final, err := s.dpmWaitOp(ctx, svc, project, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, final, s.flagDPMFormat)
}

// --- resources impl ---

func (s *deploymentManagerCmds) runDPMResourceDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Resources.Get(project, s.flagDPMDeployment, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing resource: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDPMFormat)
}

func (s *deploymentManagerCmds) runDPMResourceList(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
	if err != nil {
		return err
	}
	var all []*deploymentmanager.Resource
	pageToken := ""
	for {
		call := svc.Resources.List(project, s.flagDPMDeployment).Context(ctx)
		if s.flagDPMFilter != "" {
			call = call.Filter(s.flagDPMFilter)
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
	if s.flagDPMFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagDPMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %-30s %s\n", "NAME", "TYPE", "STATE")
	for _, r := range all {
		state := ""
		if r.Update != nil {
			state = r.Update.State
		}
		fmt.Fprintf(s.inv.stdout, "%-40s %-30s %s\n", r.Name, r.Type, state)
	}
	return nil
}

// --- types impl ---

func (s *deploymentManagerCmds) runDPMTypeList(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManager(ctx)
	if err != nil {
		return err
	}
	var all []*deploymentmanager.Type
	pageToken := ""
	for {
		call := svc.Types.List(project).Context(ctx)
		if s.flagDPMFilter != "" {
			call = call.Filter(s.flagDPMFilter)
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
	if s.flagDPMFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagDPMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "SELF_LINK")
	for _, t := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", t.Name, t.SelfLink)
	}
	return nil
}

func (s *deploymentManagerCmds) runDPMTypeProviders(cmd *cobra.Command, args []string) error {
	project, err := s.dpmResolveProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeploymentManagerBeta(ctx)
	if err != nil {
		return err
	}
	var all []*deploymentmanagerbeta.TypeProvider
	pageToken := ""
	for {
		call := svc.TypeProviders.List(project).Context(ctx)
		if s.flagDPMFilter != "" {
			call = call.Filter(s.flagDPMFilter)
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
	if s.flagDPMFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagDPMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "DESCRIPTION")
	for _, tp := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", tp.Name, tp.Description)
	}
	return nil
}

// deploymentManagerCmds holds the commands and flag values of one invocation's "deploymentManager"
// command group, so invocations don't share state (#2070).
type deploymentManagerCmds struct {
	inv *invocation

	deploymentManagerCmd            *cobra.Command
	flagDPMFormat                   string
	flagDPMFilter                   string
	flagDPMConfigFile               string
	flagDPMDeployment               string
	flagDPMDescription              string
	flagDPMPreview                  bool
	flagDPMFingerprint              string
	flagDPMCreatePolicy             string
	flagDPMDeletePolicy             string
	flagDPMAsync                    bool
	flagDPMPollIntervalS            int
	deploymentManagerDeploymentsCmd *cobra.Command
	dpmDepCreateCmd                 *cobra.Command
	dpmDepDeleteCmd                 *cobra.Command
	dpmDepDescribeCmd               *cobra.Command
	dpmDepListCmd                   *cobra.Command
	dpmDepUpdateCmd                 *cobra.Command
	dpmDepStopCmd                   *cobra.Command
	dpmDepCancelPreviewCmd          *cobra.Command
	deploymentManagerManifestsCmd   *cobra.Command
	dpmManifestDescribeCmd          *cobra.Command
	dpmManifestListCmd              *cobra.Command
	deploymentManagerOperationsCmd  *cobra.Command
	dpmOpDescribeCmd                *cobra.Command
	dpmOpListCmd                    *cobra.Command
	dpmOpWaitCmd                    *cobra.Command
	deploymentManagerResourcesCmd   *cobra.Command
	dpmResDescribeCmd               *cobra.Command
	dpmResListCmd                   *cobra.Command
	deploymentManagerTypesCmd       *cobra.Command
	dpmTypeListCmd                  *cobra.Command
	dpmTypeProvidersCmd             *cobra.Command
}

// newDeploymentManagerCmd returns the "deploymentManager" command group for inv.
func newDeploymentManagerCmd(inv *invocation) *cobra.Command {
	return newDeploymentManagerCmds(inv).deploymentManagerCmd
}

// newDeploymentManagerCmds builds the "deploymentManager" command group for inv and returns its state.
func newDeploymentManagerCmds(inv *invocation) *deploymentManagerCmds {
	s := &deploymentManagerCmds{inv: inv}
	s.deploymentManagerCmd = &cobra.Command{Use: "deployment-manager", Short: "Manage Deployment Manager"}
	s.deploymentManagerDeploymentsCmd = &cobra.Command{Use: "deployments", Short: "Manage deployments"}
	s.dpmDepCreateCmd = &cobra.Command{
		Use: "create DEPLOYMENT", Short: "Create a deployment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runDPMDepCreate,
	}
	s.dpmDepDeleteCmd = &cobra.Command{
		Use: "delete DEPLOYMENT", Short: "Delete a deployment",
		Args: cobra.ExactArgs(1), RunE: s.runDPMDepDelete,
	}
	s.dpmDepDescribeCmd = &cobra.Command{
		Use: "describe DEPLOYMENT", Short: "Describe a deployment",
		Args: cobra.ExactArgs(1), RunE: s.runDPMDepDescribe,
	}
	s.dpmDepListCmd = &cobra.Command{
		Use: "list", Short: "List deployments in the current project",
		Args: cobra.NoArgs, RunE: s.runDPMDepList,
	}
	s.dpmDepUpdateCmd = &cobra.Command{
		Use: "update DEPLOYMENT", Short: "Update a deployment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runDPMDepUpdate,
	}
	s.dpmDepStopCmd = &cobra.Command{
		Use: "stop DEPLOYMENT", Short: "Stop a deployment update",
		Args: cobra.ExactArgs(1), RunE: s.runDPMDepStop,
	}
	s.dpmDepCancelPreviewCmd = &cobra.Command{
		Use: "cancel-preview DEPLOYMENT", Short: "Cancel a preview of a deployment update",
		Args: cobra.ExactArgs(1), RunE: s.runDPMDepCancelPreview,
	}
	s.deploymentManagerManifestsCmd = &cobra.Command{Use: "manifests", Short: "Manage manifests"}
	s.dpmManifestDescribeCmd = &cobra.Command{
		Use: "describe MANIFEST", Short: "Describe a manifest for a deployment",
		Args: cobra.ExactArgs(1), RunE: s.runDPMManifestDescribe,
	}
	s.dpmManifestListCmd = &cobra.Command{
		Use: "list", Short: "List manifests for a deployment",
		Args: cobra.NoArgs, RunE: s.runDPMManifestList,
	}
	s.deploymentManagerOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage operations"}
	s.dpmOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: s.runDPMOpDescribe,
	}
	s.dpmOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in the current project",
		Args: cobra.NoArgs, RunE: s.runDPMOpList,
	}
	s.dpmOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for an operation to complete",
		Args: cobra.ExactArgs(1), RunE: s.runDPMOpWait,
	}
	s.deploymentManagerResourcesCmd = &cobra.Command{Use: "resources", Short: "Manage resources"}
	s.dpmResDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe a resource for a deployment",
		Args: cobra.ExactArgs(1), RunE: s.runDPMResourceDescribe,
	}
	s.dpmResListCmd = &cobra.Command{
		Use: "list", Short: "List resources for a deployment",
		Args: cobra.NoArgs, RunE: s.runDPMResourceList,
	}
	s.deploymentManagerTypesCmd = &cobra.Command{Use: "types", Short: "Manage types"}
	s.dpmTypeListCmd = &cobra.Command{
		Use: "list", Short: "List types in the current project",
		Args: cobra.NoArgs, RunE: s.runDPMTypeList,
	}
	s.dpmTypeProvidersCmd = &cobra.Command{
		Use: "providers", Short: "List type providers in the current project (v2beta)",
		Args: cobra.NoArgs, RunE: s.runDPMTypeProviders,
	}
	s.initDeploymentManager()
	return s
}
