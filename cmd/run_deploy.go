package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	runv2 "google.golang.org/api/run/v2"
)

// --- gcloud run deploy (#1056) ---
//
// Top-level `gcloud run deploy SERVICE` — the primary developer-facing
// command for pushing a container image to Cloud Run. Behaviour mirrors the
// Python `gcloud run deploy` for the managed platform.

var (
	flagRunDeployRegion       string
	flagRunDeployImage        string
	flagRunDeployFormat       string
	flagRunDeployAllowUnauth  bool
	flagRunDeployNoAllowUnauth bool
	flagRunDeployServiceAcct  string
	flagRunDeployMemory       string
	flagRunDeployCPU          string
	flagRunDeployMaxInstances int64
	flagRunDeployMinInstances int64
	flagRunDeployConcurrency  int64
	flagRunDeployTimeout      string
	flagRunDeployPort         int64
	flagRunDeployEnvVars      map[string]string
	flagRunDeploySetEnvVars   map[string]string
	flagRunDeployCommand      []string
	flagRunDeployArgs         []string
	flagRunDeployVpcConnector string
	flagRunDeployIngress      string
	flagRunDeployLabels       map[string]string
	flagRunDeployPlatform     string
)

var runDeployCmd = &cobra.Command{
	Use:   "deploy SERVICE",
	Short: "Deploy a Cloud Run service",
	Args:  cobra.ExactArgs(1),
	RunE:  runDeployRun,
}

func init() {
	runDeployCmd.Flags().StringVar(&flagRunDeployRegion, "region", "", "Cloud Run region (required)")
	runDeployCmd.Flags().StringVar(&flagRunDeployImage, "image", "", "Container image (required)")
	_ = runDeployCmd.MarkFlagRequired("region")
	_ = runDeployCmd.MarkFlagRequired("image")

	runDeployCmd.Flags().StringVar(&flagRunDeployFormat, "format", "", "Output format")
	runDeployCmd.Flags().BoolVar(&flagRunDeployAllowUnauth, "allow-unauthenticated", false,
		"Allow unauthenticated invocations (grants roles/run.invoker to allUsers)")
	runDeployCmd.Flags().BoolVar(&flagRunDeployNoAllowUnauth, "no-allow-unauthenticated", false,
		"Explicitly disallow unauthenticated invocations")
	runDeployCmd.Flags().StringVar(&flagRunDeployServiceAcct, "service-account", "", "Service account for the revision")
	runDeployCmd.Flags().StringVar(&flagRunDeployMemory, "memory", "", "Memory limit (e.g. 512Mi)")
	runDeployCmd.Flags().StringVar(&flagRunDeployCPU, "cpu", "", "CPU limit (e.g. 1, 2)")
	runDeployCmd.Flags().Int64Var(&flagRunDeployMaxInstances, "max-instances", 0, "Maximum number of instances")
	runDeployCmd.Flags().Int64Var(&flagRunDeployMinInstances, "min-instances", 0, "Minimum number of instances")
	runDeployCmd.Flags().Int64Var(&flagRunDeployConcurrency, "concurrency", 0, "Max concurrent requests per instance")
	runDeployCmd.Flags().StringVar(&flagRunDeployTimeout, "timeout", "", "Request timeout (e.g. 300s)")
	runDeployCmd.Flags().Int64Var(&flagRunDeployPort, "port", 0, "Container port")
	runDeployCmd.Flags().StringToStringVar(&flagRunDeployEnvVars, "env-vars", nil, "Container environment variables (KEY=VALUE)")
	runDeployCmd.Flags().StringToStringVar(&flagRunDeploySetEnvVars, "set-env-vars", nil, "Alias for --env-vars")
	runDeployCmd.Flags().StringSliceVar(&flagRunDeployCommand, "command", nil, "Container entrypoint override")
	runDeployCmd.Flags().StringSliceVar(&flagRunDeployArgs, "args", nil, "Container arguments override")
	runDeployCmd.Flags().StringVar(&flagRunDeployVpcConnector, "vpc-connector", "", "VPC connector to use")
	runDeployCmd.Flags().StringVar(&flagRunDeployIngress, "ingress", "",
		"Ingress setting: all, internal, internal-and-cloud-load-balancing")
	runDeployCmd.Flags().StringToStringVar(&flagRunDeployLabels, "labels", nil, "Service labels (KEY=VALUE)")
	runDeployCmd.Flags().StringVar(&flagRunDeployPlatform, "platform", "managed",
		"Deployment platform (only 'managed' is supported)")

	runCmd.AddCommand(runDeployCmd)
}

func runDeployRun(cmd *cobra.Command, args []string) error {
	if flagRunDeployPlatform != "" && flagRunDeployPlatform != "managed" {
		return fmt.Errorf("only --platform=managed is supported (got %q); this CLI does not support the deprecated gke/kubernetes platforms",
			flagRunDeployPlatform)
	}
	if flagRunDeployAllowUnauth && flagRunDeployNoAllowUnauth {
		return fmt.Errorf("--allow-unauthenticated and --no-allow-unauthenticated are mutually exclusive")
	}

	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunDeployRegion)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/services/%s",
		project, flagRunDeployRegion, args[0])
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
	container.Image = flagRunDeployImage
	if flagRunDeployCommand != nil {
		container.Command = flagRunDeployCommand
	}
	if flagRunDeployArgs != nil {
		container.Args = flagRunDeployArgs
	}
	env := mergeEnvMaps(flagRunDeployEnvVars, flagRunDeploySetEnvVars)
	if len(env) > 0 {
		container.Env = envVarsFromMap(env)
	}
	if flagRunDeployPort > 0 {
		container.Ports = []*runv2.GoogleCloudRunV2ContainerPort{{ContainerPort: flagRunDeployPort}}
	}
	applyResourceLimits(&container.Resources, flagRunDeployMemory, flagRunDeployCPU)

	if flagRunDeployServiceAcct != "" {
		body.Template.ServiceAccount = flagRunDeployServiceAcct
	}
	if flagRunDeployTimeout != "" {
		body.Template.Timeout = flagRunDeployTimeout
	}
	if flagRunDeployConcurrency > 0 {
		body.Template.MaxInstanceRequestConcurrency = flagRunDeployConcurrency
	}
	if flagRunDeployMinInstances > 0 || flagRunDeployMaxInstances > 0 {
		if body.Template.Scaling == nil {
			body.Template.Scaling = &runv2.GoogleCloudRunV2RevisionScaling{}
		}
		if flagRunDeployMinInstances > 0 {
			body.Template.Scaling.MinInstanceCount = flagRunDeployMinInstances
		}
		if flagRunDeployMaxInstances > 0 {
			body.Template.Scaling.MaxInstanceCount = flagRunDeployMaxInstances
		}
	}
	if flagRunDeployVpcConnector != "" {
		if body.Template.VpcAccess == nil {
			body.Template.VpcAccess = &runv2.GoogleCloudRunV2VpcAccess{}
		}
		body.Template.VpcAccess.Connector = flagRunDeployVpcConnector
	}
	if flagRunDeployIngress != "" {
		mapped, err := mapDeployIngress(flagRunDeployIngress)
		if err != nil {
			return err
		}
		body.Ingress = mapped
	}
	if len(flagRunDeployLabels) > 0 {
		if body.Labels == nil {
			body.Labels = map[string]string{}
		}
		for k, v := range flagRunDeployLabels {
			body.Labels[k] = v
		}
	}

	var op *runv2.GoogleLongrunningOperation
	if existing == nil {
		op, err = svc.Projects.Locations.Services.Create(
			fmt.Sprintf("projects/%s/locations/%s", project, flagRunDeployRegion), body).
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

	if flagRunDeployAllowUnauth {
		if err := runDeploySetAllowUnauth(ctx, svc, name); err != nil {
			return fmt.Errorf("granting run.invoker to allUsers: %w", err)
		}
	}

	// Print the resulting service URL, refetching to pick up any URI
	// that only becomes populated after the reconcile completes.
	got, ferr := svc.Projects.Locations.Services.Get(name).Context(ctx).Do()
	if ferr == nil && got.Uri != "" {
		fmt.Printf("Service [%s] URL: %s\n", args[0], got.Uri)
	} else {
		fmt.Printf("Deploy request issued for service [%s] (operation: %s).\n", args[0], op.Name)
	}
	return emitFormatted(op, flagRunDeployFormat)
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
