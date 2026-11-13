package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	dataflow "google.golang.org/api/dataflow/v1b3"
	storage "google.golang.org/api/storage/v1"
)

// --- gcloud dataflow flex-template / snapshots / yaml (#936, #937, #938) ---

var (
	// flex-template flags
	flagDFFTJobName string

	// flex-template build flags
	flagDFFTBuildTemplateFileGCS string
)

func (s *dataflowCmds) registerDataflowExtras(parent *cobra.Command) {
	// --- flex-template subgroup ---
	ftCmd := &cobra.Command{Use: "flex-template", Short: "Manage Dataflow Flex Templates"}

	ftRun := &cobra.Command{
		Use:   "run JOB_NAME",
		Short: "Run a Dataflow Flex Template",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDFFTLaunch,
	}
	ftRun.Flags().StringVar(&s.flagDataflowRegion, "region", "", "Region for the job")
	ftRun.Flags().StringVar(&s.flagDFFTTemplateGCS, "template-file-gcs-location", "", "GCS path to the flex template spec (required)")
	ftRun.MarkFlagRequired("template-file-gcs-location")
	ftRun.Flags().StringVar(&s.flagDFFTStagingLocation, "staging-location", "", "GCS staging location")
	ftRun.Flags().StringVar(&s.flagDFFTTempLocation, "temp-location", "", "GCS temp location")
	ftRun.Flags().StringVar(&s.flagDFFTServiceAccount, "service-account-email", "", "Worker service account email")
	ftRun.Flags().Int64Var(&s.flagDFFTMaxWorkers, "max-workers", 0, "Maximum number of workers")
	ftRun.Flags().Int64Var(&s.flagDFFTNumWorkers, "num-workers", 0, "Initial number of workers")
	ftRun.Flags().StringVar(&s.flagDFFTWorkerMachine, "worker-machine-type", "", "Worker machine type")
	ftRun.Flags().StringVar(&s.flagDFFTLauncherMachine, "launcher-machine-type", "", "Launcher VM machine type")
	ftRun.Flags().Int64Var(&s.flagDFFTLauncherTimeout, "launcher-vm-timeout-secs", 0, "Launcher VM timeout in seconds")
	ftRun.Flags().StringVar(&s.flagDFFTSubnetwork, "subnetwork", "", "Compute Engine subnetwork")
	ftRun.Flags().StringVar(&s.flagDFFTNetwork, "network", "", "Compute Engine network")
	ftRun.Flags().StringVar(&s.flagDFFTKmsKey, "dataflow-kms-key", "", "Cloud KMS key")
	ftRun.Flags().StringVar(&s.flagDFFTWorkerRegion, "worker-region", "", "Region to run workers in")
	ftRun.Flags().StringVar(&s.flagDFFTWorkerZone, "worker-zone", "", "Zone to run workers in")
	ftRun.Flags().BoolVar(&s.flagDFFTEnableStreaming, "enable-streaming-engine", false, "Enable Streaming Engine")
	ftRun.Flags().BoolVar(&s.flagDFFTDisablePublicIPs, "disable-public-ips", false, "Disable public IPs on workers")
	ftRun.Flags().StringSliceVar(&s.flagDFFTExperiments, "additional-experiments", nil, "Additional Dataflow experiments")
	ftRun.Flags().StringSliceVar(&s.flagDFFTPipelineOptions, "additional-pipeline-options", nil, "Additional pipeline options")
	ftRun.Flags().StringToStringVar(&s.flagDFFTUserLabels, "additional-user-labels", nil, "Additional user labels (key=value)")
	ftRun.Flags().StringToStringVar(&s.flagDFFTParameters, "parameters", nil, "Template parameters (key=value)")
	ftRun.Flags().StringToStringVar(&s.flagDFFTTransformName, "transform-name-mappings", nil, "Streaming update transform name mappings")
	ftRun.Flags().BoolVar(&s.flagDFFTUpdate, "update", false, "Update an existing streaming job")
	ftRun.Flags().StringVar(&s.flagDFFTFlexRSGoal, "flexrs-goal", "", "FlexRS goal (COST_OPTIMIZED, SPEED_OPTIMIZED)")
	ftRun.Flags().BoolVar(&s.flagDFFTTurnkeyAlerts, "enable-turnkey-alerts", false, "Enable turnkey alerts for the launched job")

	ftBuild := &cobra.Command{
		Use:   "build TEMPLATE_FILE_GCS_PATH",
		Short: "Build a Dataflow flex template spec and upload it to Cloud Storage",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDFFTBuild,
	}
	ftBuild.Flags().StringVar(&s.flagDFFTBuildImage, "image", "", "Container image URI for the flex template")
	ftBuild.Flags().StringVar(&s.flagDFFTBuildImageGCRPath, "image-gcr-path", "", "Alternate image GCR path (metadata only)")
	ftBuild.Flags().StringVar(&s.flagDFFTBuildSDK, "sdk-language", "", "SDK language (JAVA, PYTHON, GO)")
	ftBuild.Flags().StringVar(&s.flagDFFTBuildMetadataFile, "metadata-file", "", "JSON file with template metadata")
	ftBuild.Flags().StringToStringVar(&s.flagDFFTBuildEnv, "env", nil, "Container spec env vars (key=value)")
	ftBuild.Flags().StringToStringVar(&s.flagDFFTBuildFlexOptions, "flex-template-base-image", nil, "Extra container spec options (key=value)")

	ftCmd.AddCommand(ftRun, ftBuild)

	// --- snapshots subgroup ---
	snapCmd := &cobra.Command{Use: "snapshots", Short: "Manage Dataflow job snapshots"}

	snapCreate := &cobra.Command{
		Use:   "create",
		Short: "Create a snapshot of a Dataflow job",
		Args:  cobra.NoArgs,
		RunE:  s.runDFSnapCreate,
	}
	snapCreate.Flags().StringVar(&s.flagDFSnapRegion, "region", "", "Region containing the job")
	snapCreate.Flags().StringVar(&s.flagDFSnapJobID, "job-id", "", "Job ID to snapshot (required)")
	snapCreate.Flags().StringVar(&s.flagDFSnapTTL, "snapshot-ttl", "", "Snapshot TTL (e.g. 604800s)")
	snapCreate.Flags().StringVar(&s.flagDFSnapDescription, "description", "", "Snapshot description")
	snapCreate.Flags().BoolVar(&s.flagDFSnapSources, "snapshot-sources", false, "Also snapshot sources that support it")
	snapCreate.MarkFlagRequired("job-id")

	snapDelete := &cobra.Command{
		Use:   "delete SNAPSHOT_ID",
		Short: "Delete a Dataflow snapshot",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDFSnapDelete,
	}
	snapDelete.Flags().StringVar(&s.flagDFSnapRegion, "region", "", "Region containing the snapshot")

	snapDescribe := &cobra.Command{
		Use:   "describe SNAPSHOT_ID",
		Short: "Describe a Dataflow snapshot",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDFSnapDescribe,
	}
	snapDescribe.Flags().StringVar(&s.flagDFSnapRegion, "region", "", "Region containing the snapshot")

	snapList := &cobra.Command{
		Use:   "list",
		Short: "List Dataflow snapshots",
		Args:  cobra.NoArgs,
		RunE:  s.runDFSnapList,
	}
	snapList.Flags().StringVar(&s.flagDFSnapRegion, "region", "", "Region to list snapshots in")
	snapList.Flags().StringVar(&s.flagDFSnapJobID, "job-id", "", "Only list snapshots belonging to this job")

	snapCmd.AddCommand(snapCreate, snapDelete, snapDescribe, snapList)

	// --- yaml subgroup ---
	yamlCmd := &cobra.Command{Use: "yaml", Short: "Run Beam YAML pipelines"}
	yamlRun := &cobra.Command{
		Use:   "run JOB_NAME",
		Short: "Run a Dataflow job from a YAML pipeline description",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDFYamlRun,
	}
	yamlRun.Flags().StringVar(&s.flagDFYamlRegion, "region", "", "Region for the job (defaults to us-central1)")
	yamlRun.Flags().StringVar(&s.flagDFYamlPipelineFile, "yaml-pipeline-file", "", "Path to a local YAML pipeline file or gs:// URL")
	yamlRun.Flags().StringVar(&s.flagDFYamlPipeline, "yaml-pipeline", "", "Inline YAML pipeline definition")
	yamlRun.Flags().StringToStringVar(&s.flagDFYamlOptions, "pipeline-options", nil, "Additional pipeline options (key=value)")
	yamlRun.Flags().StringToStringVar(&s.flagDFYamlJinjaVars, "jinja-variables", nil, "Jinja variables for template substitution (key=value)")
	yamlRun.Flags().StringVar(&s.flagDFYamlGCSLocation, "template-file-gcs-location", "", "Override GCS location of the Beam YAML flex template")

	yamlCmd.AddCommand(yamlRun)

	parent.AddCommand(ftCmd, snapCmd, yamlCmd)
}

// --- flex-template run ---

func (s *dataflowCmds) runDFFTLaunch(cmd *cobra.Command, args []string) error {
	project, region, err := s.resolveDataflowRegion()
	if err != nil {
		return err
	}
	env := &dataflow.FlexTemplateRuntimeEnvironment{}
	if s.flagDFFTMaxWorkers > 0 {
		env.MaxWorkers = s.flagDFFTMaxWorkers
	}
	if s.flagDFFTNumWorkers > 0 {
		env.NumWorkers = s.flagDFFTNumWorkers
	}
	if s.flagDFFTNetwork != "" {
		env.Network = s.flagDFFTNetwork
	}
	if s.flagDFFTSubnetwork != "" {
		env.Subnetwork = s.flagDFFTSubnetwork
	}
	if s.flagDFFTWorkerMachine != "" {
		env.MachineType = s.flagDFFTWorkerMachine
	}
	if s.flagDFFTLauncherMachine != "" {
		env.LauncherMachineType = s.flagDFFTLauncherMachine
	}
	// --launcher-vm-timeout-secs is accepted but ignored — the v1b3 Go client
	// does not expose LauncherVmTimeoutSeconds in FlexTemplateRuntimeEnvironment.
	_ = s.flagDFFTLauncherTimeout
	if s.flagDFFTServiceAccount != "" {
		env.ServiceAccountEmail = s.flagDFFTServiceAccount
	}
	if s.flagDFFTStagingLocation != "" {
		env.StagingLocation = s.flagDFFTStagingLocation
	}
	if s.flagDFFTTempLocation != "" {
		env.TempLocation = s.flagDFFTTempLocation
	}
	if s.flagDFFTKmsKey != "" {
		env.KmsKeyName = s.flagDFFTKmsKey
	}
	if s.flagDFFTWorkerRegion != "" {
		env.WorkerRegion = s.flagDFFTWorkerRegion
	}
	if s.flagDFFTWorkerZone != "" {
		env.WorkerZone = s.flagDFFTWorkerZone
	}
	if s.flagDFFTEnableStreaming {
		env.EnableStreamingEngine = true
	}
	if s.flagDFFTDisablePublicIPs {
		env.IpConfiguration = "WORKER_IP_PRIVATE"
	}
	if len(s.flagDFFTExperiments) > 0 {
		env.AdditionalExperiments = s.flagDFFTExperiments
	}
	if s.flagDFFTTurnkeyAlerts {
		env.AdditionalExperiments = appendExperiment(env.AdditionalExperiments, turnkeyAlertsExperiment)
	}
	if len(s.flagDFFTPipelineOptions) > 0 {
		env.AdditionalPipelineOptions = s.flagDFFTPipelineOptions
	}
	if len(s.flagDFFTUserLabels) > 0 {
		env.AdditionalUserLabels = s.flagDFFTUserLabels
	}
	if s.flagDFFTFlexRSGoal != "" {
		env.FlexrsGoal = s.flagDFFTFlexRSGoal
	}

	param := &dataflow.LaunchFlexTemplateParameter{
		JobName:              args[0],
		ContainerSpecGcsPath: s.flagDFFTTemplateGCS,
		Environment:          env,
	}
	if len(s.flagDFFTParameters) > 0 {
		param.Parameters = s.flagDFFTParameters
	}
	if len(s.flagDFFTTransformName) > 0 {
		param.TransformNameMappings = s.flagDFFTTransformName
	}
	if s.flagDFFTUpdate {
		param.Update = true
	}

	req := &dataflow.LaunchFlexTemplateRequest{LaunchParameter: param}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataflow(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.FlexTemplates.Launch(project, region, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("launching flex template: %w", err)
	}
	if resp.Job != nil {
		fmt.Fprintf(s.inv.stdout, "Launched flex template job [%s] (ID: %s).\n", resp.Job.Name, resp.Job.Id)
	}
	return emitFormattedTo(s.inv.stdout, resp, "")
}

// --- flex-template build ---
//
// Writes a container-spec JSON file (metadata for the flex template) to the
// provided GCS location. Container image building is a client-side concern
// that requires Docker or Cloud Build; we surface the option via --image and
// embed it in the spec, but do not shell out to docker here.

func (s *dataflowCmds) runDFFTBuild(cmd *cobra.Command, args []string) error {
	if s.flagDFFTBuildImage == "" {
		return fmt.Errorf("--image is required (container URI to embed in the flex template spec)")
	}
	spec := map[string]any{
		"image": s.flagDFFTBuildImage,
	}
	if s.flagDFFTBuildSDK != "" {
		spec["sdkInfo"] = map[string]string{"language": s.flagDFFTBuildSDK}
	}
	if len(s.flagDFFTBuildEnv) > 0 {
		spec["defaultEnvironment"] = s.flagDFFTBuildEnv
	}
	if s.flagDFFTBuildMetadataFile != "" {
		meta, err := os.ReadFile(s.flagDFFTBuildMetadataFile)
		if err != nil {
			return fmt.Errorf("reading %s: %w", s.flagDFFTBuildMetadataFile, err)
		}
		var m any
		if err := json.Unmarshal(meta, &m); err != nil {
			return fmt.Errorf("parsing metadata file: %w", err)
		}
		spec["metadata"] = m
	}
	if s.flagDFFTBuildImageGCRPath != "" {
		spec["imageGcrPath"] = s.flagDFFTBuildImageGCRPath
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
	ctx := cmd.Context()
	stg, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	obj := &storage.Object{Name: object, ContentType: "application/json"}
	if _, err := stg.Objects.Insert(bucket, obj).Media(strings.NewReader(string(body))).Context(ctx).Do(); err != nil {
		return fmt.Errorf("writing spec to %s: %w", dest, err)
	}
	fmt.Fprintf(s.inv.stdout, "Wrote flex template spec to %s\n", dest)
	return nil
}

// --- snapshots ---

func (s *dataflowCmds) runDFSnapCreate(cmd *cobra.Command, args []string) error {
	project, region, err := s.dfSnapRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataflow(ctx)
	if err != nil {
		return err
	}
	req := &dataflow.SnapshotJobRequest{
		Location:        region,
		Description:     s.flagDFSnapDescription,
		Ttl:             s.flagDFSnapTTL,
		SnapshotSources: s.flagDFSnapSources,
	}
	snap, err := svc.Projects.Locations.Jobs.Snapshot(project, region, s.flagDFSnapJobID, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("snapshotting job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created snapshot [%s] for job [%s].\n", snap.Id, s.flagDFSnapJobID)
	return emitFormattedTo(s.inv.stdout, snap, "")
}

func (s *dataflowCmds) runDFSnapDelete(cmd *cobra.Command, args []string) error {
	project, region, err := s.dfSnapRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataflow(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Snapshots.Delete(project, region, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting snapshot: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted snapshot [%s].\n", args[0])
	return nil
}

func (s *dataflowCmds) runDFSnapDescribe(cmd *cobra.Command, args []string) error {
	project, region, err := s.dfSnapRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataflow(ctx)
	if err != nil {
		return err
	}
	snap, err := svc.Projects.Locations.Snapshots.Get(project, region, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing snapshot: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, snap, "")
}

func (s *dataflowCmds) runDFSnapList(cmd *cobra.Command, args []string) error {
	project, region, err := s.dfSnapRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataflow(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Snapshots.List(project, region).Context(ctx)
	if s.flagDFSnapJobID != "" {
		call = call.JobId(s.flagDFSnapJobID)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing snapshots: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp.Snapshots, "")
}

func (s *dataflowCmds) dfSnapRegion() (string, string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", "", err
	}
	if s.flagDFSnapRegion == "" {
		if _, r, err := s.inv.region(); err == nil && r != "" {
			return project, r, nil
		}
		return "", "", fmt.Errorf("--region is required")
	}
	return project, s.flagDFSnapRegion, nil
}

// --- yaml run ---
//
// Launches the built-in Beam YAML flex template at
// gs://dataflow-templates-<region>/latest/flex/Yaml_Template, passing the
// user's YAML pipeline (inline or by-reference) as template parameters.

func (s *dataflowCmds) runDFYamlRun(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	region := s.flagDFYamlRegion
	if region == "" {
		if _, r, err := s.inv.region(); err == nil && r != "" {
			region = r
		}
	}
	if region == "" {
		region = "us-central1"
	}

	if s.flagDFYamlPipelineFile == "" && s.flagDFYamlPipeline == "" {
		return fmt.Errorf("one of --yaml-pipeline-file or --yaml-pipeline is required")
	}

	params := map[string]string{}
	for k, v := range s.flagDFYamlOptions {
		params[k] = v
	}

	if s.flagDFYamlPipeline != "" {
		params["yaml_pipeline"] = s.flagDFYamlPipeline
	} else if strings.HasPrefix(s.flagDFYamlPipelineFile, "gs://") {
		params["yaml_pipeline_file"] = s.flagDFYamlPipelineFile
	} else {
		f, err := os.Open(s.flagDFYamlPipelineFile)
		if err != nil {
			return fmt.Errorf("opening %s: %w", s.flagDFYamlPipelineFile, err)
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}
		params["yaml_pipeline"] = string(b)
	}

	if len(s.flagDFYamlJinjaVars) > 0 {
		jv, err := json.Marshal(s.flagDFYamlJinjaVars)
		if err != nil {
			return err
		}
		params["jinja_variables"] = string(jv)
	}

	gcsLocation := s.flagDFYamlGCSLocation
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

	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataflow(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.FlexTemplates.Launch(project, region, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("launching yaml pipeline: %w", err)
	}
	if resp.Job != nil {
		fmt.Fprintf(s.inv.stdout, "Launched yaml pipeline job [%s] (ID: %s).\n", resp.Job.Name, resp.Job.Id)
	}
	return emitFormattedTo(s.inv.stdout, resp, "")
}
