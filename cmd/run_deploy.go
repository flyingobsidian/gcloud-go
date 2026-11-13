package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	runv2 "google.golang.org/api/run/v2"
)

// --- gcloud run deploy (#1056) ---
//
// Top-level `gcloud run deploy SERVICE` — the primary developer-facing
// command for pushing a container image to Cloud Run. Behaviour mirrors the
// Python `gcloud run deploy` for the managed platform.

func (s *runCmds) initRunDeploy() {
	s.runDeployCmd.Flags().StringVar(&s.flagRunDeployRegion, "region", "", "Cloud Run region (required)")
	s.runDeployCmd.Flags().StringVar(&s.flagRunDeployImage, "image", "", "Container image (required)")
	_ = s.runDeployCmd.MarkFlagRequired("region")
	_ = s.runDeployCmd.MarkFlagRequired("image")

	s.runDeployCmd.Flags().StringVar(&s.flagRunDeployFormat, "format", "", "Output format")
	s.runDeployCmd.Flags().BoolVar(&s.flagRunDeployAllowUnauth, "allow-unauthenticated", false,
		"Allow unauthenticated invocations (grants roles/run.invoker to allUsers)")
	s.runDeployCmd.Flags().BoolVar(&s.flagRunDeployNoAllowUnauth, "no-allow-unauthenticated", false,
		"Explicitly disallow unauthenticated invocations")
	s.runDeployCmd.Flags().StringVar(&s.flagRunDeployServiceAcct, "service-account", "", "Service account for the revision")
	s.runDeployCmd.Flags().StringVar(&s.flagRunDeployMemory, "memory", "", "Memory limit (e.g. 512Mi)")
	s.runDeployCmd.Flags().StringVar(&s.flagRunDeployCPU, "cpu", "", "CPU limit (e.g. 1, 2)")
	s.runDeployCmd.Flags().Int64Var(&s.flagRunDeployMaxInstances, "max-instances", 0, "Maximum number of instances")
	s.runDeployCmd.Flags().Int64Var(&s.flagRunDeployMinInstances, "min-instances", 0, "Minimum number of instances")
	s.runDeployCmd.Flags().Int64Var(&s.flagRunDeployConcurrency, "concurrency", 0, "Max concurrent requests per instance")
	s.runDeployCmd.Flags().StringVar(&s.flagRunDeployTimeout, "timeout", "", "Request timeout (e.g. 300s)")
	s.runDeployCmd.Flags().Int64Var(&s.flagRunDeployPort, "port", 0, "Container port")
	s.runDeployCmd.Flags().StringToStringVar(&s.flagRunDeployEnvVars, "env-vars", nil, "Container environment variables (KEY=VALUE)")
	s.runDeployCmd.Flags().StringToStringVar(&s.flagRunDeploySetEnvVars, "set-env-vars", nil, "Alias for --env-vars")
	s.runDeployCmd.Flags().StringSliceVar(&s.flagRunDeployCommand, "command", nil, "Container entrypoint override")
	s.runDeployCmd.Flags().StringSliceVar(&s.flagRunDeployArgs, "args", nil, "Container arguments override")
	s.runDeployCmd.Flags().StringVar(&s.flagRunDeployVpcConnector, "vpc-connector", "", "VPC connector to use")
	s.runDeployCmd.Flags().StringVar(&s.flagRunDeployIngress, "ingress", "",
		"Ingress setting: all, internal, internal-and-cloud-load-balancing")
	s.runDeployCmd.Flags().StringToStringVar(&s.flagRunDeployLabels, "labels", nil, "Service labels (KEY=VALUE)")
	s.runDeployCmd.Flags().StringVar(&s.flagRunDeployPlatform, "platform", "managed",
		"Deployment platform (only 'managed' is supported)")

	s.runCmd.AddCommand(s.runDeployCmd)
}

func (s *runCmds) runDeployRun(cmd *cobra.Command, args []string) error {
	if s.flagRunDeployPlatform != "" && s.flagRunDeployPlatform != "managed" {
		return fmt.Errorf("only --platform=managed is supported (got %q); this CLI does not support the deprecated gke/kubernetes platforms",
			s.flagRunDeployPlatform)
	}
	if s.flagRunDeployAllowUnauth && s.flagRunDeployNoAllowUnauth {
		return fmt.Errorf("--allow-unauthenticated and --no-allow-unauthenticated are mutually exclusive")
	}

	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunDeployRegion)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/services/%s",
		project, s.flagRunDeployRegion, args[0])
	existing, err := svc.Projects.Locations.Services.Get(name).Context(ctx).Do()
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("checking existing service: %w", err)
	}
	body := &runv2.GoogleCloudRunV2Service{}
	if existing != nil {
		body = existing
		body.Name = name
	}
	if body.Template == nil {
		body.Template = &runv2.GoogleCloudRunV2RevisionTemplate{}
	}
	var container *runv2.GoogleCloudRunV2Container
	if len(body.Template.Containers) > 0 {
		container = body.Template.Containers[0]
	} else {
		container = &runv2.GoogleCloudRunV2Container{}
		body.Template.Containers = []*runv2.GoogleCloudRunV2Container{container}
	}
	container.Image = s.flagRunDeployImage
	if s.flagRunDeployCommand != nil {
		container.Command = s.flagRunDeployCommand
	}
	if s.flagRunDeployArgs != nil {
		container.Args = s.flagRunDeployArgs
	}
	env := mergeEnvMaps(s.flagRunDeployEnvVars, s.flagRunDeploySetEnvVars)
	if len(env) > 0 {
		container.Env = envVarsFromMap(env)
	}
	if s.flagRunDeployPort > 0 {
		container.Ports = []*runv2.GoogleCloudRunV2ContainerPort{{ContainerPort: s.flagRunDeployPort}}
	}
	applyResourceLimits(&container.Resources, s.flagRunDeployMemory, s.flagRunDeployCPU)

	if s.flagRunDeployServiceAcct != "" {
		body.Template.ServiceAccount = s.flagRunDeployServiceAcct
	}
	if s.flagRunDeployTimeout != "" {
		body.Template.Timeout = s.flagRunDeployTimeout
	}
	if s.flagRunDeployConcurrency > 0 {
		body.Template.MaxInstanceRequestConcurrency = s.flagRunDeployConcurrency
	}
	if s.flagRunDeployMinInstances > 0 || s.flagRunDeployMaxInstances > 0 {
		if body.Template.Scaling == nil {
			body.Template.Scaling = &runv2.GoogleCloudRunV2RevisionScaling{}
		}
		if s.flagRunDeployMinInstances > 0 {
			body.Template.Scaling.MinInstanceCount = s.flagRunDeployMinInstances
		}
		if s.flagRunDeployMaxInstances > 0 {
			body.Template.Scaling.MaxInstanceCount = s.flagRunDeployMaxInstances
		}
	}
	if err := applyRunScalingTargets(body.Template, s.flagRunDeployScalingCPUTarget, s.flagRunDeployScalingConcurrencyTarget); err != nil {
		return err
	}
	if s.flagRunDeployVpcConnector != "" {
		if body.Template.VpcAccess == nil {
			body.Template.VpcAccess = &runv2.GoogleCloudRunV2VpcAccess{}
		}
		body.Template.VpcAccess.Connector = s.flagRunDeployVpcConnector
	}
	if s.flagRunDeployIngress != "" {
		mapped, err := mapDeployIngress(s.flagRunDeployIngress)
		if err != nil {
			return err
		}
		body.Ingress = mapped
	}
	if len(s.flagRunDeployLabels) > 0 {
		if body.Labels == nil {
			body.Labels = map[string]string{}
		}
		for k, v := range s.flagRunDeployLabels {
			body.Labels[k] = v
		}
	}

	var op *runv2.GoogleLongrunningOperation
	if existing == nil {
		op, err = svc.Projects.Locations.Services.Create(
			fmt.Sprintf("projects/%s/locations/%s", project, s.flagRunDeployRegion), body).
			ServiceId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating service: %w", err)
		}
	} else {
		op, err = svc.Projects.Locations.Services.Patch(name, body).
			ForceNewRevision(true).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("updating service: %w", err)
		}
	}

	if s.flagRunDeployAllowUnauth {
		if err := runDeploySetAllowUnauth(ctx, svc, name); err != nil {
			return fmt.Errorf("granting run.invoker to allUsers: %w", err)
		}
	}

	// Print the resulting service URL, refetching to pick up any URI
	// that only becomes populated after the reconcile completes.
	got, ferr := svc.Projects.Locations.Services.Get(name).Context(ctx).Do()
	if ferr == nil && got.Uri != "" {
		fmt.Fprintf(s.inv.stdout, "Service [%s] URL: %s\n", args[0], got.Uri)
	} else {
		fmt.Fprintf(s.inv.stdout, "Deploy request issued for service [%s] (operation: %s).\n", args[0], op.Name)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagRunDeployFormat)
}

// runDeploySetAllowUnauth adds the roles/run.invoker binding for allUsers on
// the just-deployed service, giving it public HTTP access.
func runDeploySetAllowUnauth(ctx context.Context, svc *runv2.Service, resource string) error {
	policy, err := svc.Projects.Locations.Services.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	runIamAddBinding(policy, "roles/run.invoker", "allUsers", nil)
	policy.Version = 3
	_, err = svc.Projects.Locations.Services.SetIamPolicy(resource,
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	return err
}

// mapDeployIngress converts the human-friendly --ingress flag value to the
// Cloud Run v2 enum string. Anything unrecognised returns an error rather
// than silently succeeding.
func mapDeployIngress(v string) (string, error) {
	switch v {
	case "all":
		return "INGRESS_TRAFFIC_ALL", nil
	case "internal":
		return "INGRESS_TRAFFIC_INTERNAL_ONLY", nil
	case "internal-and-cloud-load-balancing":
		return "INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER", nil
	default:
		return "", fmt.Errorf("invalid --ingress value %q (expected one of: all, internal, internal-and-cloud-load-balancing)", v)
	}
}

// mergeEnvMaps returns the union of a and b; entries in b overwrite a.
func mergeEnvMaps(a, b map[string]string) map[string]string {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	out := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}
