package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dataflow "google.golang.org/api/dataflow/v1b3"
	storage "google.golang.org/api/storage/v1"
)

// --- gcloud dataflow flex-template / snapshots / yaml (#936, #937, #938) ---

var (
	// flex-template flags
	flagDFFTJobName          string
	flagDFFTTemplateGCS      string
	flagDFFTStagingLocation  string
	flagDFFTTempLocation     string
	flagDFFTServiceAccount   string
	flagDFFTMaxWorkers       int64
	flagDFFTNumWorkers       int64
	flagDFFTWorkerMachine    string
	flagDFFTLauncherMachine  string
	flagDFFTLauncherTimeout  int64
	flagDFFTSubnetwork       string
	flagDFFTNetwork          string
	flagDFFTKmsKey           string
	flagDFFTWorkerRegion     string
	flagDFFTWorkerZone       string
	flagDFFTEnableStreaming  bool
	flagDFFTDisablePublicIPs bool
	flagDFFTExperiments      []string
	flagDFFTPipelineOptions  []string
	flagDFFTUserLabels       map[string]string
	flagDFFTParameters       map[string]string
	flagDFFTTransformName    map[string]string
	flagDFFTUpdate           bool
	flagDFFTFlexRSGoal       string

	// flex-template build flags
	flagDFFTBuildTemplateFileGCS string
	flagDFFTBuildImage           string
	flagDFFTBuildImageGCRPath    string
	flagDFFTBuildSDK             string
	flagDFFTBuildMetadataFile    string
	flagDFFTBuildEnv             map[string]string
	flagDFFTBuildFlexOptions     map[string]string

	// snapshots flags
	flagDFSnapRegion      string
	flagDFSnapJobID       string
	flagDFSnapTTL         string
	flagDFSnapDescription string
	flagDFSnapSources     bool

	// yaml run flags
	flagDFYamlPipelineFile string
	flagDFYamlPipeline     string
	flagDFYamlOptions      map[string]string
	flagDFYamlJinjaVars    map[string]string
	flagDFYamlRegion       string
	flagDFYamlGCSLocation  string
)

func registerDataflowExtras(parent *cobra.Command) {
	// --- flex-template subgroup ---
	ftCmd := &cobra.Command{Use: "flex-template", Short: "Manage Dataflow Flex Templates"}

	ftRun := &cobra.Command{
		Use:   "run JOB_NAME",
		Short: "Run a Dataflow Flex Template",
		Args:  cobra.ExactArgs(1),
		RunE:  runDFFTLaunch,
	}
	ftRun.Flags().StringVar(&flagDataflowRegion, "region", "", "Region for the job")
	ftRun.Flags().StringVar(&flagDFFTTemplateGCS, "template-file-gcs-location", "", "GCS path to the flex template spec (required)")
	ftRun.MarkFlagRequired("template-file-gcs-location")
	ftRun.Flags().StringVar(&flagDFFTStagingLocation, "staging-location", "", "GCS staging location")
	ftRun.Flags().StringVar(&flagDFFTTempLocation, "temp-location", "", "GCS temp location")
	ftRun.Flags().StringVar(&flagDFFTServiceAccount, "service-account-email", "", "Worker service account email")
	ftRun.Flags().Int64Var(&flagDFFTMaxWorkers, "max-workers", 0, "Maximum number of workers")
	ftRun.Flags().Int64Var(&flagDFFTNumWorkers, "num-workers", 0, "Initial number of workers")
	ftRun.Flags().StringVar(&flagDFFTWorkerMachine, "worker-machine-type", "", "Worker machine type")
	ftRun.Flags().StringVar(&flagDFFTLauncherMachine, "launcher-machine-type", "", "Launcher VM machine type")
	ftRun.Flags().Int64Var(&flagDFFTLauncherTimeout, "launcher-vm-timeout-secs", 0, "Launcher VM timeout in seconds")
	ftRun.Flags().StringVar(&flagDFFTSubnetwork, "subnetwork", "", "Compute Engine subnetwork")
	ftRun.Flags().StringVar(&flagDFFTNetwork, "network", "", "Compute Engine network")
	ftRun.Flags().StringVar(&flagDFFTKmsKey, "dataflow-kms-key", "", "Cloud KMS key")
	ftRun.Flags().StringVar(&flagDFFTWorkerRegion, "worker-region", "", "Region to run workers in")
	ftRun.Flags().StringVar(&flagDFFTWorkerZone, "worker-zone", "", "Zone to run workers in")
	ftRun.Flags().BoolVar(&flagDFFTEnableStreaming, "enable-streaming-engine", false, "Enable Streaming Engine")
	ftRun.Flags().BoolVar(&flagDFFTDisablePublicIPs, "disable-public-ips", false, "Disable public IPs on workers")
	ftRun.Flags().StringSliceVar(&flagDFFTExperiments, "additional-experiments", nil, "Additional Dataflow experiments")
	ftRun.Flags().StringSliceVar(&flagDFFTPipelineOptions, "additional-pipeline-options", nil, "Additional pipeline options")
	ftRun.Flags().StringToStringVar(&flagDFFTUserLabels, "additional-user-labels", nil, "Additional user labels (key=value)")
	ftRun.Flags().StringToStringVar(&flagDFFTParameters, "parameters", nil, "Template parameters (key=value)")
	ftRun.Flags().StringToStringVar(&flagDFFTTransformName, "transform-name-mappings", nil, "Streaming update transform name mappings")
	ftRun.Flags().BoolVar(&flagDFFTUpdate, "update", false, "Update an existing streaming job")
	ftRun.Flags().StringVar(&flagDFFTFlexRSGoal, "flexrs-goal", "", "FlexRS goal (COST_OPTIMIZED, SPEED_OPTIMIZED)")

	ftBuild := &cobra.Command{
		Use:   "build TEMPLATE_FILE_GCS_PATH",
		Short: "Build a Dataflow flex template spec and upload it to Cloud Storage",
		Args:  cobra.ExactArgs(1),
		RunE:  runDFFTBuild,
	}
	ftBuild.Flags().StringVar(&flagDFFTBuildImage, "image", "", "Container image URI for the flex template")
	ftBuild.Flags().StringVar(&flagDFFTBuildImageGCRPath, "image-gcr-path", "", "Alternate image GCR path (metadata only)")
	ftBuild.Flags().StringVar(&flagDFFTBuildSDK, "sdk-language", "", "SDK language (JAVA, PYTHON, GO)")
	ftBuild.Flags().StringVar(&flagDFFTBuildMetadataFile, "metadata-file", "", "JSON file with template metadata")
	ftBuild.Flags().StringToStringVar(&flagDFFTBuildEnv, "env", nil, "Container spec env vars (key=value)")
	ftBuild.Flags().StringToStringVar(&flagDFFTBuildFlexOptions, "flex-template-base-image", nil, "Extra container spec options (key=value)")

	ftCmd.AddCommand(ftRun, ftBuild)

	// --- snapshots subgroup ---
	snapCmd := &cobra.Command{Use: "snapshots", Short: "Manage Dataflow job snapshots"}

	snapCreate := &cobra.Command{
		Use:   "create",
		Short: "Create a snapshot of a Dataflow job",
		Args:  cobra.NoArgs,
		RunE:  runDFSnapCreate,
	}
	snapCreate.Flags().StringVar(&flagDFSnapRegion, "region", "", "Region containing the job")
	snapCreate.Flags().StringVar(&flagDFSnapJobID, "job-id", "", "Job ID to snapshot (required)")
	snapCreate.Flags().StringVar(&flagDFSnapTTL, "snapshot-ttl", "", "Snapshot TTL (e.g. 604800s)")
	snapCreate.Flags().StringVar(&flagDFSnapDescription, "description", "", "Snapshot description")
	snapCreate.Flags().BoolVar(&flagDFSnapSources, "snapshot-sources", false, "Also snapshot sources that support it")
	snapCreate.MarkFlagRequired("job-id")

	snapDelete := &cobra.Command{
		Use:   "delete SNAPSHOT_ID",
		Short: "Delete a Dataflow snapshot",
		Args:  cobra.ExactArgs(1),
		RunE:  runDFSnapDelete,
	}
	snapDelete.Flags().StringVar(&flagDFSnapRegion, "region", "", "Region containing the snapshot")

	snapDescribe := &cobra.Command{
		Use:   "describe SNAPSHOT_ID",
		Short: "Describe a Dataflow snapshot",
		Args:  cobra.ExactArgs(1),
		RunE:  runDFSnapDescribe,
	}
	snapDescribe.Flags().StringVar(&flagDFSnapRegion, "region", "", "Region containing the snapshot")

	snapList := &cobra.Command{
		Use:   "list",
		Short: "List Dataflow snapshots",
		Args:  cobra.NoArgs,
		RunE:  runDFSnapList,
	}
	snapList.Flags().StringVar(&flagDFSnapRegion, "region", "", "Region to list snapshots in")
	snapList.Flags().StringVar(&flagDFSnapJobID, "job-id", "", "Only list snapshots belonging to this job")

	snapCmd.AddCommand(snapCreate, snapDelete, snapDescribe, snapList)

	// --- yaml subgroup ---
	yamlCmd := &cobra.Command{Use: "yaml", Short: "Run Beam YAML pipelines"}
	yamlRun := &cobra.Command{
		Use:   "run JOB_NAME",
		Short: "Run a Dataflow job from a YAML pipeline description",
		Args:  cobra.ExactArgs(1),
		RunE:  runDFYamlRun,
	}
	yamlRun.Flags().StringVar(&flagDFYamlRegion, "region", "", "Region for the job (defaults to us-central1)")
	yamlRun.Flags().StringVar(&flagDFYamlPipelineFile, "yaml-pipeline-file", "", "Path to a local YAML pipeline file or gs:// URL")
	yamlRun.Flags().StringVar(&flagDFYamlPipeline, "yaml-pipeline", "", "Inline YAML pipeline definition")
	yamlRun.Flags().StringToStringVar(&flagDFYamlOptions, "pipeline-options", nil, "Additional pipeline options (key=value)")
	yamlRun.Flags().StringToStringVar(&flagDFYamlJinjaVars, "jinja-variables", nil, "Jinja variables for template substitution (key=value)")
	yamlRun.Flags().StringVar(&flagDFYamlGCSLocation, "template-file-gcs-location", "", "Override GCS location of the Beam YAML flex template")

	yamlCmd.AddCommand(yamlRun)

	parent.AddCommand(ftCmd, snapCmd, yamlCmd)
}

// --- flex-template run ---

func runDFFTLaunch(cmd *cobra.Command, args []string) error {
	project, region, err := resolveDataflowRegion()
	if err != nil {
		return err
	}
	env := &dataflow.FlexTemplateRuntimeEnvironment{}
	if flagDFFTMaxWorkers > 0 {
		env.MaxWorkers = flagDFFTMaxWorkers
	}
	if flagDFFTNumWorkers > 0 {
		env.NumWorkers = flagDFFTNumWorkers
	}
	if flagDFFTNetwork != "" {
		env.Network = flagDFFTNetwork
	}
	if flagDFFTSubnetwork != "" {
		env.Subnetwork = flagDFFTSubnetwork
	}
	if flagDFFTWorkerMachine != "" {
		env.MachineType = flagDFFTWorkerMachine
	}
	if flagDFFTLauncherMachine != "" {
		env.LauncherMachineType = flagDFFTLauncherMachine
	}
	// --launcher-vm-timeout-secs is accepted but ignored — the v1b3 Go client
	// does not expose LauncherVmTimeoutSeconds in FlexTemplateRuntimeEnvironment.
	_ = flagDFFTLauncherTimeout
	if flagDFFTServiceAccount != "" {
		env.ServiceAccountEmail = flagDFFTServiceAccount
	}
	if flagDFFTStagingLocation != "" {
		env.StagingLocation = flagDFFTStagingLocation
	}
	if flagDFFTTempLocation != "" {
		env.TempLocation = flagDFFTTempLocation
	}
	if flagDFFTKmsKey != "" {
		env.KmsKeyName = flagDFFTKmsKey
	}
	if flagDFFTWorkerRegion != "" {
		env.WorkerRegion = flagDFFTWorkerRegion
	}
	if flagDFFTWorkerZone != "" {
		env.WorkerZone = flagDFFTWorkerZone
	}
	if flagDFFTEnableStreaming {
		env.EnableStreamingEngine = true
	}
	if flagDFFTDisablePublicIPs {
		env.IpConfiguration = "WORKER_IP_PRIVATE"
	}
	if len(flagDFFTExperiments) > 0 {
		env.AdditionalExperiments = flagDFFTExperiments
	}
	if len(flagDFFTPipelineOptions) > 0 {
		env.AdditionalPipelineOptions = flagDFFTPipelineOptions
	}
	if len(flagDFFTUserLabels) > 0 {
		env.AdditionalUserLabels = flagDFFTUserLabels
	}
	if flagDFFTFlexRSGoal != "" {
		env.FlexrsGoal = flagDFFTFlexRSGoal
	}

	param := &dataflow.LaunchFlexTemplateParameter{
		JobName:              args[0],
		ContainerSpecGcsPath: flagDFFTTemplateGCS,
		Environment:          env,
	}
	if len(flagDFFTParameters) > 0 {
		param.Parameters = flagDFFTParameters
	}
	if len(flagDFFTTransformName) > 0 {
		param.TransformNameMappings = flagDFFTTransformName
	}
	if flagDFFTUpdate {
		param.Update = true
	}

	req := &dataflow.LaunchFlexTemplateRequest{LaunchParameter: param}

	ctx := context.Background()
	svc, err := gcp.DataflowService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.FlexTemplates.Launch(project, region, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("launching flex template: %w", err)
	}
	if resp.Job != nil {
		fmt.Printf("Launched flex template job [%s] (ID: %s).\n", resp.Job.Name, resp.Job.Id)
	}
	return emitFormatted(resp, "")
}

// --- flex-template build ---
//
// Writes a container-spec JSON file (metadata for the flex template) to the
// provided GCS location. Container image building is a client-side concern
// that requires Docker or Cloud Build; we surface the option via --image and
// embed it in the spec, but do not shell out to docker here.

func runDFFTBuild(cmd *cobra.Command, args []string) error {
	if flagDFFTBuildImage == "" {
		return fmt.Errorf("--image is required (container URI to embed in the flex template spec)")
	}
	spec := map[string]any{
		"image": flagDFFTBuildImage,
	}
	if flagDFFTBuildSDK != "" {
		spec["sdkInfo"] = map[string]string{"language": flagDFFTBuildSDK}
	}
	if len(flagDFFTBuildEnv) > 0 {
		spec["defaultEnvironment"] = flagDFFTBuildEnv
	}
	if flagDFFTBuildMetadataFile != "" {
		meta, err := os.ReadFile(flagDFFTBuildMetadataFile)
		if err != nil {
			return fmt.Errorf("reading %s: %w", flagDFFTBuildMetadataFile, err)
		}
		var m any
		if err := json.Unmarshal(meta, &m); err != nil {
			return fmt.Errorf("parsing metadata file: %w", err)
		}
		spec["metadata"] = m
	}
	if flagDFFTBuildImageGCRPath != "" {
		spec["imageGcrPath"] = flagDFFTBuildImageGCRPath
	}
	body, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}
	dest := args[0]
	if !strings.HasPrefix(dest, "gs://") {
		return fmt.Errorf("TEMPLATE_FILE_GCS_PATH must begin with gs://")
	}
	trimmed := strings.TrimPrefix(dest, "gs://")
	slash := strings.IndexByte(trimmed, '/')
	if slash < 0 {
		return fmt.Errorf("TEMPLATE_FILE_GCS_PATH must include an object name")
	}
	bucket := trimmed[:slash]
	object := trimmed[slash+1:]
	ctx := context.Background()
	stg, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	obj := &storage.Object{Name: object, ContentType: "application/json"}
	if _, err := stg.Objects.Insert(bucket, obj).Media(strings.NewReader(string(body))).Context(ctx).Do(); err != nil {
		return fmt.Errorf("writing spec to %s: %w", dest, err)
	}
	fmt.Printf("Wrote flex template spec to %s\n", dest)
	return nil
}

// --- snapshots ---

func runDFSnapCreate(cmd *cobra.Command, args []string) error {
	project, region, err := dfSnapRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataflowService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &dataflow.SnapshotJobRequest{
		Location:        region,
		Description:     flagDFSnapDescription,
		Ttl:             flagDFSnapTTL,
		SnapshotSources: flagDFSnapSources,
	}
	snap, err := svc.Projects.Locations.Jobs.Snapshot(project, region, flagDFSnapJobID, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("snapshotting job: %w", err)
	}
	fmt.Printf("Created snapshot [%s] for job [%s].\n", snap.Id, flagDFSnapJobID)
	return emitFormatted(snap, "")
}

func runDFSnapDelete(cmd *cobra.Command, args []string) error {
	project, region, err := dfSnapRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataflowService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Snapshots.Delete(project, region, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting snapshot: %w", err)
	}
	fmt.Printf("Deleted snapshot [%s].\n", args[0])
	return nil
}

func runDFSnapDescribe(cmd *cobra.Command, args []string) error {
	project, region, err := dfSnapRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataflowService(ctx, flagAccount)
	if err != nil {
		return err
	}
	snap, err := svc.Projects.Locations.Snapshots.Get(project, region, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing snapshot: %w", err)
	}
	return emitFormatted(snap, "")
}

func runDFSnapList(cmd *cobra.Command, args []string) error {
	project, region, err := dfSnapRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataflowService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Snapshots.List(project, region).Context(ctx)
	if flagDFSnapJobID != "" {
		call = call.JobId(flagDFSnapJobID)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing snapshots: %w", err)
	}
	return emitFormatted(resp.Snapshots, "")
}

func dfSnapRegion() (string, string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", "", err
	}
	if flagDFSnapRegion == "" {
		if _, r, err := resolveRegion(); err == nil && r != "" {
			return project, r, nil
		}
		return "", "", fmt.Errorf("--region is required")
	}
	return project, flagDFSnapRegion, nil
}

// --- yaml run ---
//
// Launches the built-in Beam YAML flex template at
// gs://dataflow-templates-<region>/latest/flex/Yaml_Template, passing the
// user's YAML pipeline (inline or by-reference) as template parameters.

func runDFYamlRun(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	region := flagDFYamlRegion
	if region == "" {
		if _, r, err := resolveRegion(); err == nil && r != "" {
			region = r
		}
	}
	if region == "" {
		region = "us-central1"
	}

	if flagDFYamlPipelineFile == "" && flagDFYamlPipeline == "" {
		return fmt.Errorf("one of --yaml-pipeline-file or --yaml-pipeline is required")
	}

	params := map[string]string{}
	for k, v := range flagDFYamlOptions {
		params[k] = v
	}

	if flagDFYamlPipeline != "" {
		params["yaml_pipeline"] = flagDFYamlPipeline
	} else if strings.HasPrefix(flagDFYamlPipelineFile, "gs://") {
		params["yaml_pipeline_file"] = flagDFYamlPipelineFile
	} else {
		f, err := os.Open(flagDFYamlPipelineFile)
		if err != nil {
			return fmt.Errorf("opening %s: %w", flagDFYamlPipelineFile, err)
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}
		params["yaml_pipeline"] = string(b)
	}

	if len(flagDFYamlJinjaVars) > 0 {
		jv, err := json.Marshal(flagDFYamlJinjaVars)
		if err != nil {
			return err
		}
		params["jinja_variables"] = string(jv)
	}

	gcsLocation := flagDFYamlGCSLocation
	if gcsLocation == "" {
		gcsLocation = fmt.Sprintf("gs://dataflow-templates-%s/latest/flex/Yaml_Template", region)
	}

	req := &dataflow.LaunchFlexTemplateRequest{
		LaunchParameter: &dataflow.LaunchFlexTemplateParameter{
			JobName:              args[0],
			ContainerSpecGcsPath: gcsLocation,
			Parameters:           params,
		},
	}

	ctx := context.Background()
	svc, err := gcp.DataflowService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.FlexTemplates.Launch(project, region, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("launching yaml pipeline: %w", err)
	}
	if resp.Job != nil {
		fmt.Printf("Launched yaml pipeline job [%s] (ID: %s).\n", resp.Job.Name, resp.Job.Id)
	}
	return emitFormatted(resp, "")
}
