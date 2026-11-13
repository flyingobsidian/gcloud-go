package cli

import (
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud ai (#291) ---
//
// Vertex AI is served by the aiplatform API surface. Every subgroup uses a
// regional aiplatform client (endpoint of the form
// https://<region>-aiplatform.googleapis.com/) and takes --region on every
// leaf command.

func (s *aiCmds) initAi() {

}

// aiCmds holds the commands and flag values of one invocation's "ai"
// command group, so invocations don't share state (#2070).
type aiCmds struct {
	inv *invocation

	aiCmd                         *cobra.Command
	aiCustomJobsCmd               *cobra.Command
	flagAICJRegion                string
	flagAICJFormat                string
	flagAICJConfigFile            string
	flagAICJFilter                string
	flagAICJOrderBy               string
	flagAICJPageSize              int64
	flagAICJReadMask              string
	flagAICJTimeout               time.Duration
	aiCJCancelCmd                 *cobra.Command
	aiCJCreateCmd                 *cobra.Command
	aiCJDescribeCmd               *cobra.Command
	aiCJListCmd                   *cobra.Command
	aiCJLocalRunCmd               *cobra.Command
	aiCJStreamLogsCmd             *cobra.Command
	aiEndpointsCmd                *cobra.Command
	flagAIEPRegion                string
	flagAIEPFormat                string
	flagAIEPConfigFile            string
	flagAIEPUpdateMask            string
	flagAIEPEndpointID            string
	flagAIEPFilter                string
	flagAIEPOrderBy               string
	flagAIEPPageSize              int64
	flagAIEPReadMask              string
	aiEPCreateCmd                 *cobra.Command
	aiEPDeleteCmd                 *cobra.Command
	aiEPDeployModelCmd            *cobra.Command
	aiEPUndeployModelCmd          *cobra.Command
	aiEPDescribeCmd               *cobra.Command
	aiEPListCmd                   *cobra.Command
	aiEPUpdateCmd                 *cobra.Command
	aiEPPredictCmd                *cobra.Command
	aiEPRawPredictCmd             *cobra.Command
	aiEPExplainCmd                *cobra.Command
	aiEPDirectPredictCmd          *cobra.Command
	aiEPDirectRawPredictCmd       *cobra.Command
	aiEPStreamDirectPredictCmd    *cobra.Command
	aiEPStreamDirectRawPredictCmd *cobra.Command
	aiEPStreamRawPredictCmd       *cobra.Command
	aiHPTCmd                      *cobra.Command
	flagAIHPTRegion               string
	flagAIHPTFormat               string
	flagAIHPTConfigFile           string
	flagAIHPTFilter               string
	flagAIHPTPageSize             int64
	flagAIHPTReadMask             string
	flagAIHPTTimeout              time.Duration
	aiHPTCancelCmd                *cobra.Command
	aiHPTCreateCmd                *cobra.Command
	aiHPTDescribeCmd              *cobra.Command
	aiHPTListCmd                  *cobra.Command
	aiHPTStreamLogsCmd            *cobra.Command
	aiIECmd                       *cobra.Command
	flagAIIERegion                string
	flagAIIEFormat                string
	flagAIIEConfigFile            string
	flagAIIEUpdateMask            string
	flagAIIEFilter                string
	flagAIIEPageSize              int64
	flagAIIEReadMask              string
	aiIECreateCmd                 *cobra.Command
	aiIEDeleteCmd                 *cobra.Command
	aiIEDeployIndexCmd            *cobra.Command
	aiIEDescribeCmd               *cobra.Command
	aiIEListCmd                   *cobra.Command
	aiIEMutateDeployedIndexCmd    *cobra.Command
	aiIEUndeployIndexCmd          *cobra.Command
	aiIEUpdateCmd                 *cobra.Command
	aiIdxCmd                      *cobra.Command
	flagAIIdxRegion               string
	flagAIIdxFormat               string
	flagAIIdxConfigFile           string
	flagAIIdxUpdateMask           string
	flagAIIdxFilter               string
	flagAIIdxPageSize             int64
	flagAIIdxReadMask             string
	aiIdxCreateCmd                *cobra.Command
	aiIdxDeleteCmd                *cobra.Command
	aiIdxDescribeCmd              *cobra.Command
	aiIdxListCmd                  *cobra.Command
	aiIdxUpdateCmd                *cobra.Command
	aiIdxUpsertCmd                *cobra.Command
	aiIdxRemoveCmd                *cobra.Command
	aiMGCmd                       *cobra.Command
	aiMGModelsCmd                 *cobra.Command
	flagAIMGRegion                string
	flagAIMGFormat                string
	flagAIMGPublisher             string
	flagAIMGFilter                string
	flagAIMGOrderBy               string
	flagAIMGPageSize              int64
	flagAIMGListAll               bool
	flagAIMGView                  string
	flagAIMGLanguage              string
	aiMGListCmd                   *cobra.Command
	aiMGDescribeCmd               *cobra.Command
	aiMMJCmd                      *cobra.Command
	flagAIMMJRegion               string
	flagAIMMJFormat               string
	flagAIMMJConfigFile           string
	flagAIMMJUpdateMask           string
	flagAIMMJFilter               string
	flagAIMMJPageSize             int64
	flagAIMMJReadMask             string
	aiMMJCreateCmd                *cobra.Command
	aiMMJDeleteCmd                *cobra.Command
	aiMMJDescribeCmd              *cobra.Command
	aiMMJListCmd                  *cobra.Command
	aiMMJUpdateCmd                *cobra.Command
	aiMMJPauseCmd                 *cobra.Command
	aiMMJResumeCmd                *cobra.Command
	aiModelsCmd                   *cobra.Command
	flagAIMRegion                 string
	flagAIMFormat                 string
	flagAIMConfigFile             string
	flagAIMVersion                string
	flagAIMFilter                 string
	flagAIMOrderBy                string
	flagAIMPageSize               int64
	flagAIMReadMask               string
	aiMDeleteCmd                  *cobra.Command
	aiMDescribeCmd                *cobra.Command
	aiMListCmd                    *cobra.Command
	aiMCopyCmd                    *cobra.Command
	aiMUploadCmd                  *cobra.Command
	aiMListVersionsCmd            *cobra.Command
	aiMDeleteVersionCmd           *cobra.Command
	aiOpsCmd                      *cobra.Command
	flagAIOpsRegion               string
	flagAIOpsFormat               string
	aiOpsDescribeCmd              *cobra.Command
	aiPRCmd                       *cobra.Command
	flagAIPRRegion                string
	flagAIPRFormat                string
	flagAIPRConfigFile            string
	flagAIPRPRId                  string
	flagAIPRPageSize              int64
	aiPRCreateCmd                 *cobra.Command
	aiPRDeleteCmd                 *cobra.Command
	aiPRDescribeCmd               *cobra.Command
	aiPRListCmd                   *cobra.Command
	aiPRRebootCmd                 *cobra.Command
	aiSGPCmd                      *cobra.Command
	aiSGPECmd                     *cobra.Command
	flagAISGPRegion               string
	flagAISGPFormat               string
	flagAISGPConfigFile           string
	flagAISGPUpdateMask           string
	flagAISGPPolicyID             string
	flagAISGPPageSize             int64
	flagAISGPDenialMessage        string
	aiSGPCreateCmd                *cobra.Command
	aiSGPDeleteCmd                *cobra.Command
	aiSGPDescribeCmd              *cobra.Command
	aiSGPListCmd                  *cobra.Command
	aiSGPUpdateCmd                *cobra.Command
	aiSGPEDeprovisionCmd          *cobra.Command
	aiTBCmd                       *cobra.Command
	flagAITBRegion                string
	flagAITBFormat                string
	flagAITBConfigFile            string
	flagAITBUpdateMask            string
	flagAITBFilter                string
	flagAITBOrderBy               string
	flagAITBPageSize              int64
	flagAITBReadMask              string
	aiTBCreateCmd                 *cobra.Command
	aiTBDeleteCmd                 *cobra.Command
	aiTBDescribeCmd               *cobra.Command
	aiTBListCmd                   *cobra.Command
	aiTBUpdateCmd                 *cobra.Command
	aiTJCmd                       *cobra.Command
	flagAITJRegion                string
	flagAITJFormat                string
	flagAITJConfigFile            string
	flagAITJFilter                string
	flagAITJPageSize              int64
	aiTJCancelCmd                 *cobra.Command
	aiTJCreateCmd                 *cobra.Command
	aiTJDescribeCmd               *cobra.Command
	aiTJListCmd                   *cobra.Command
}

// newAiCmd returns the "ai" command group for inv.
func newAiCmd(inv *invocation) *cobra.Command {
	return newAiCmds(inv).aiCmd
}

// newAiCmds builds the "ai" command group for inv and returns its state.
func newAiCmds(inv *invocation) *aiCmds {
	s := &aiCmds{inv: inv}
	s.aiCmd = &cobra.Command{Use: "ai", Short: "Manage Vertex AI"}
	s.aiCustomJobsCmd = &cobra.Command{Use: "custom-jobs", Short: "Manage Vertex AI custom training jobs"}
	s.aiCJCancelCmd = &cobra.Command{
		Use: "cancel CUSTOM_JOB", Short: "Cancel a custom training job",
		Args: cobra.ExactArgs(1), RunE: s.runAICJCancel,
	}
	s.aiCJCreateCmd = &cobra.Command{
		Use: "create", Short: "Submit a custom training job",
		Args: cobra.NoArgs, RunE: s.runAICJCreate,
	}
	s.aiCJDescribeCmd = &cobra.Command{
		Use: "describe CUSTOM_JOB", Short: "Describe a custom training job",
		Args: cobra.ExactArgs(1), RunE: s.runAICJDescribe,
	}
	s.aiCJListCmd = &cobra.Command{
		Use: "list", Short: "List custom training jobs",
		Args: cobra.NoArgs, RunE: s.runAICJList,
	}
	s.aiCJLocalRunCmd = &cobra.Command{
		Use:   "local-run",
		Short: "Not supported: run a training job locally (requires Docker)",
		Long: "local-run requires Docker and is not supported by gcloud-go; " +
			"use `gcloud ai custom-jobs create` to submit the job to Vertex AI instead.",
		Args: cobra.ArbitraryArgs, RunE: runAICJLocalRun,
	}
	s.aiCJStreamLogsCmd = &cobra.Command{
		Use: "stream-logs CUSTOM_JOB",
		Short: "Poll a custom job until it reaches a terminal state, " +
			"printing its state each interval",
		Args: cobra.ExactArgs(1), RunE: s.runAICJStreamLogs,
	}
	s.aiEndpointsCmd = &cobra.Command{Use: "endpoints", Short: "Manage Vertex AI endpoints"}
	s.aiEPCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an endpoint",
		Args: cobra.NoArgs, RunE: s.runAIEPCreate,
	}
	s.aiEPDeleteCmd = &cobra.Command{
		Use: "delete ENDPOINT", Short: "Delete an endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIEPDelete,
	}
	s.aiEPDeployModelCmd = &cobra.Command{
		Use: "deploy-model ENDPOINT", Short: "Deploy a model to an endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIEPDeployModel,
	}
	s.aiEPUndeployModelCmd = &cobra.Command{
		Use: "undeploy-model ENDPOINT", Short: "Undeploy a model from an endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIEPUndeployModel,
	}
	s.aiEPDescribeCmd = &cobra.Command{
		Use: "describe ENDPOINT", Short: "Describe an endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIEPDescribe,
	}
	s.aiEPListCmd = &cobra.Command{
		Use: "list", Short: "List endpoints",
		Args: cobra.NoArgs, RunE: s.runAIEPList,
	}
	s.aiEPUpdateCmd = &cobra.Command{
		Use: "update ENDPOINT", Short: "Update an endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIEPUpdate,
	}
	s.aiEPPredictCmd = &cobra.Command{
		Use: "predict ENDPOINT", Short: "Send a prediction request to an endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIEPPredict,
	}
	s.aiEPRawPredictCmd = &cobra.Command{
		Use: "raw-predict ENDPOINT", Short: "Send a raw prediction request to an endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIEPRawPredict,
	}
	s.aiEPExplainCmd = &cobra.Command{
		Use: "explain ENDPOINT", Short: "Send an explanation request to an endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIEPExplain,
	}
	s.aiEPDirectPredictCmd = &cobra.Command{
		Use: "direct-predict ENDPOINT", Short: "Send a direct prediction request to a dedicated endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIEPDirectPredict,
	}
	s.aiEPDirectRawPredictCmd = &cobra.Command{
		Use: "direct-raw-predict ENDPOINT", Short: "Send a direct raw prediction request to a dedicated endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIEPDirectRawPredict,
	}
	s.aiEPStreamDirectPredictCmd = &cobra.Command{
		Use: "stream-direct-predict ENDPOINT",
		Short: "Send a streaming direct prediction request " +
			"(REST fallback: returns the single response emitted by DirectPredict)",
		Long: "Vertex AI's stream-direct-predict is gRPC-only. Over the REST " +
			"transport used by gcloud-go this issues a single DirectPredict " +
			"call and prints the one response returned by the server.",
		Args: cobra.ExactArgs(1), RunE: s.runAIEPDirectPredict,
	}
	s.aiEPStreamDirectRawPredictCmd = &cobra.Command{
		Use: "stream-direct-raw-predict ENDPOINT",
		Short: "Send a streaming direct raw prediction request " +
			"(REST fallback: returns the single response emitted by DirectRawPredict)",
		Long: "Vertex AI's stream-direct-raw-predict is gRPC-only. Over the REST " +
			"transport used by gcloud-go this issues a single DirectRawPredict " +
			"call and prints the one response returned by the server.",
		Args: cobra.ExactArgs(1), RunE: s.runAIEPDirectRawPredict,
	}
	s.aiEPStreamRawPredictCmd = &cobra.Command{
		Use: "stream-raw-predict ENDPOINT",
		Short: "Send a streaming raw prediction request " +
			"(returns the full aggregated response body over REST)",
		Args: cobra.ExactArgs(1), RunE: s.runAIEPStreamRawPredict,
	}
	s.aiHPTCmd = &cobra.Command{Use: "hp-tuning-jobs", Short: "Manage Vertex AI hyperparameter tuning jobs"}
	s.aiHPTCancelCmd = &cobra.Command{
		Use: "cancel HP_TUNING_JOB", Short: "Cancel a hyperparameter tuning job",
		Args: cobra.ExactArgs(1), RunE: s.runAIHPTCancel,
	}
	s.aiHPTCreateCmd = &cobra.Command{
		Use: "create", Short: "Submit a hyperparameter tuning job",
		Args: cobra.NoArgs, RunE: s.runAIHPTCreate,
	}
	s.aiHPTDescribeCmd = &cobra.Command{
		Use: "describe HP_TUNING_JOB", Short: "Describe a hyperparameter tuning job",
		Args: cobra.ExactArgs(1), RunE: s.runAIHPTDescribe,
	}
	s.aiHPTListCmd = &cobra.Command{
		Use: "list", Short: "List hyperparameter tuning jobs",
		Args: cobra.NoArgs, RunE: s.runAIHPTList,
	}
	s.aiHPTStreamLogsCmd = &cobra.Command{
		Use: "stream-logs HP_TUNING_JOB",
		Short: "Poll a hyperparameter tuning job until it reaches a terminal state, " +
			"printing its state each interval",
		Args: cobra.ExactArgs(1), RunE: s.runAIHPTStreamLogs,
	}
	s.aiIECmd = &cobra.Command{Use: "index-endpoints", Short: "Manage Vertex AI index endpoints"}
	s.aiIECreateCmd = &cobra.Command{
		Use: "create", Short: "Create an index endpoint",
		Args: cobra.NoArgs, RunE: s.runAIIECreate,
	}
	s.aiIEDeleteCmd = &cobra.Command{
		Use: "delete INDEX_ENDPOINT", Short: "Delete an index endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIIEDelete,
	}
	s.aiIEDeployIndexCmd = &cobra.Command{
		Use: "deploy-index INDEX_ENDPOINT", Short: "Deploy an index to an index endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIIEDeployIndex,
	}
	s.aiIEDescribeCmd = &cobra.Command{
		Use: "describe INDEX_ENDPOINT", Short: "Describe an index endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIIEDescribe,
	}
	s.aiIEListCmd = &cobra.Command{
		Use: "list", Short: "List index endpoints",
		Args: cobra.NoArgs, RunE: s.runAIIEList,
	}
	s.aiIEMutateDeployedIndexCmd = &cobra.Command{
		Use:   "mutate-deployed-index INDEX_ENDPOINT",
		Short: "Update a deployed index on an index endpoint",
		Args:  cobra.ExactArgs(1), RunE: s.runAIIEMutateDeployedIndex,
	}
	s.aiIEUndeployIndexCmd = &cobra.Command{
		Use: "undeploy-index INDEX_ENDPOINT", Short: "Undeploy an index from an index endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIIEUndeployIndex,
	}
	s.aiIEUpdateCmd = &cobra.Command{
		Use: "update INDEX_ENDPOINT", Short: "Update an index endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runAIIEUpdate,
	}
	s.aiIdxCmd = &cobra.Command{Use: "indexes", Short: "Manage Vertex AI indexes"}
	s.aiIdxCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an index",
		Args: cobra.NoArgs, RunE: s.runAIIdxCreate,
	}
	s.aiIdxDeleteCmd = &cobra.Command{
		Use: "delete INDEX", Short: "Delete an index",
		Args: cobra.ExactArgs(1), RunE: s.runAIIdxDelete,
	}
	s.aiIdxDescribeCmd = &cobra.Command{
		Use: "describe INDEX", Short: "Describe an index",
		Args: cobra.ExactArgs(1), RunE: s.runAIIdxDescribe,
	}
	s.aiIdxListCmd = &cobra.Command{
		Use: "list", Short: "List indexes",
		Args: cobra.NoArgs, RunE: s.runAIIdxList,
	}
	s.aiIdxUpdateCmd = &cobra.Command{
		Use: "update INDEX", Short: "Update an index",
		Args: cobra.ExactArgs(1), RunE: s.runAIIdxUpdate,
	}
	s.aiIdxUpsertCmd = &cobra.Command{
		Use: "upsert-datapoints INDEX", Short: "Upsert datapoints in an index",
		Args: cobra.ExactArgs(1), RunE: s.runAIIdxUpsert,
	}
	s.aiIdxRemoveCmd = &cobra.Command{
		Use: "remove-datapoints INDEX", Short: "Remove datapoints from an index",
		Args: cobra.ExactArgs(1), RunE: s.runAIIdxRemove,
	}
	s.aiMGCmd = &cobra.Command{Use: "model-garden", Short: "Interact with Vertex AI Model Garden"}
	s.aiMGModelsCmd = &cobra.Command{Use: "models", Short: "Browse Model Garden publisher models"}
	s.aiMGListCmd = &cobra.Command{
		Use: "list", Short: "List publisher models in Model Garden",
		Args: cobra.NoArgs, RunE: s.runAIMGList,
	}
	s.aiMGDescribeCmd = &cobra.Command{
		Use: "describe MODEL", Short: "Describe a publisher model in Model Garden",
		Args: cobra.ExactArgs(1), RunE: s.runAIMGDescribe,
	}
	s.aiMMJCmd = &cobra.Command{Use: "model-monitoring-jobs", Short: "Manage Vertex AI model deployment monitoring jobs"}
	s.aiMMJCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a model deployment monitoring job",
		Args: cobra.NoArgs, RunE: s.runAIMMJCreate,
	}
	s.aiMMJDeleteCmd = &cobra.Command{
		Use: "delete MONITORING_JOB", Short: "Delete a model deployment monitoring job",
		Args: cobra.ExactArgs(1), RunE: s.runAIMMJDelete,
	}
	s.aiMMJDescribeCmd = &cobra.Command{
		Use: "describe MONITORING_JOB", Short: "Describe a model deployment monitoring job",
		Args: cobra.ExactArgs(1), RunE: s.runAIMMJDescribe,
	}
	s.aiMMJListCmd = &cobra.Command{
		Use: "list", Short: "List model deployment monitoring jobs",
		Args: cobra.NoArgs, RunE: s.runAIMMJList,
	}
	s.aiMMJUpdateCmd = &cobra.Command{
		Use: "update MONITORING_JOB", Short: "Update a model deployment monitoring job",
		Args: cobra.ExactArgs(1), RunE: s.runAIMMJUpdate,
	}
	s.aiMMJPauseCmd = &cobra.Command{
		Use: "pause MONITORING_JOB", Short: "Pause a model deployment monitoring job",
		Args: cobra.ExactArgs(1), RunE: s.runAIMMJPause,
	}
	s.aiMMJResumeCmd = &cobra.Command{
		Use: "resume MONITORING_JOB", Short: "Resume a paused model deployment monitoring job",
		Args: cobra.ExactArgs(1), RunE: s.runAIMMJResume,
	}
	s.aiModelsCmd = &cobra.Command{Use: "models", Short: "Manage Vertex AI models"}
	s.aiMDeleteCmd = &cobra.Command{
		Use: "delete MODEL", Short: "Delete a model",
		Args: cobra.ExactArgs(1), RunE: s.runAIMDelete,
	}
	s.aiMDescribeCmd = &cobra.Command{
		Use: "describe MODEL", Short: "Describe a model",
		Args: cobra.ExactArgs(1), RunE: s.runAIMDescribe,
	}
	s.aiMListCmd = &cobra.Command{
		Use: "list", Short: "List models",
		Args: cobra.NoArgs, RunE: s.runAIMList,
	}
	s.aiMCopyCmd = &cobra.Command{
		Use: "copy", Short: "Copy a model",
		Args: cobra.NoArgs, RunE: s.runAIMCopy,
	}
	s.aiMUploadCmd = &cobra.Command{
		Use: "upload", Short: "Upload a model",
		Args: cobra.NoArgs, RunE: s.runAIMUpload,
	}
	s.aiMListVersionsCmd = &cobra.Command{
		Use: "list-version MODEL", Short: "List versions of a model",
		Args: cobra.ExactArgs(1), RunE: s.runAIMListVersions,
	}
	s.aiMDeleteVersionCmd = &cobra.Command{
		Use: "delete-version MODEL", Short: "Delete a specific version of a model",
		Long: "MODEL may be given as \"MODEL@VERSION\" or as MODEL together with " +
			"--version=VERSION. In either form the resource name sent to the " +
			"server is MODEL@VERSION.",
		Args: cobra.ExactArgs(1), RunE: s.runAIMDeleteVersion,
	}
	s.aiOpsCmd = &cobra.Command{Use: "operations", Short: "Manage Vertex AI long-running operations"}
	s.aiOpsDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Vertex AI long-running operation",
		Args: cobra.ExactArgs(1), RunE: s.runAIOpsDescribe,
	}
	s.aiPRCmd = &cobra.Command{Use: "persistent-resources", Short: "Manage Vertex AI persistent resources"}
	s.aiPRCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a persistent resource",
		Args: cobra.NoArgs, RunE: s.runAIPRCreate,
	}
	s.aiPRDeleteCmd = &cobra.Command{
		Use: "delete PERSISTENT_RESOURCE", Short: "Delete a persistent resource",
		Args: cobra.ExactArgs(1), RunE: s.runAIPRDelete,
	}
	s.aiPRDescribeCmd = &cobra.Command{
		Use: "describe PERSISTENT_RESOURCE", Short: "Describe a persistent resource",
		Args: cobra.ExactArgs(1), RunE: s.runAIPRDescribe,
	}
	s.aiPRListCmd = &cobra.Command{
		Use: "list", Short: "List persistent resources",
		Args: cobra.NoArgs, RunE: s.runAIPRList,
	}
	s.aiPRRebootCmd = &cobra.Command{
		Use: "reboot PERSISTENT_RESOURCE", Short: "Reboot a persistent resource",
		Args: cobra.ExactArgs(1), RunE: s.runAIPRReboot,
	}
	s.aiSGPCmd = &cobra.Command{
		Use:   "semantic-governance-policies",
		Short: "Manage Vertex AI Semantic Governance policies",
	}
	s.aiSGPECmd = &cobra.Command{
		Use:   "semantic-governance-policy-engine",
		Short: "Manage the Vertex AI Semantic Governance policy engine",
	}
	s.aiSGPCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a semantic governance policy",
		Args: cobra.NoArgs, RunE: s.runAISGPCreate,
	}
	s.aiSGPDeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete a semantic governance policy",
		Args: cobra.ExactArgs(1), RunE: s.runAISGPDelete,
	}
	s.aiSGPDescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe a semantic governance policy",
		Args: cobra.ExactArgs(1), RunE: s.runAISGPDescribe,
	}
	s.aiSGPListCmd = &cobra.Command{
		Use: "list", Short: "List semantic governance policies",
		Args: cobra.NoArgs, RunE: s.runAISGPList,
	}
	s.aiSGPUpdateCmd = &cobra.Command{
		Use: "update POLICY", Short: "Update a semantic governance policy",
		Args: cobra.ExactArgs(1), RunE: s.runAISGPUpdate,
	}
	s.aiSGPEDeprovisionCmd = &cobra.Command{
		Use: "deprovision", Short: "Tear down the semantic governance policy engine (tenant project, GKE cluster, PSC service attachments)",
		Args: cobra.NoArgs, RunE: s.runAISGPEDeprovision,
	}
	s.aiTBCmd = &cobra.Command{Use: "tensorboards", Short: "Manage Vertex AI Tensorboards"}
	s.aiTBCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a Tensorboard",
		Args: cobra.NoArgs, RunE: s.runAITBCreate,
	}
	s.aiTBDeleteCmd = &cobra.Command{
		Use: "delete TENSORBOARD", Short: "Delete a Tensorboard",
		Args: cobra.ExactArgs(1), RunE: s.runAITBDelete,
	}
	s.aiTBDescribeCmd = &cobra.Command{
		Use: "describe TENSORBOARD", Short: "Describe a Tensorboard",
		Args: cobra.ExactArgs(1), RunE: s.runAITBDescribe,
	}
	s.aiTBListCmd = &cobra.Command{
		Use: "list", Short: "List Tensorboards",
		Args: cobra.NoArgs, RunE: s.runAITBList,
	}
	s.aiTBUpdateCmd = &cobra.Command{
		Use: "update TENSORBOARD", Short: "Update a Tensorboard",
		Args: cobra.ExactArgs(1), RunE: s.runAITBUpdate,
	}
	s.aiTJCmd = &cobra.Command{Use: "tuning-jobs", Short: "Manage Vertex AI GenAI tuning jobs"}
	s.aiTJCancelCmd = &cobra.Command{
		Use: "cancel TUNING_JOB", Short: "Cancel a tuning job",
		Args: cobra.ExactArgs(1), RunE: s.runAITJCancel,
	}
	s.aiTJCreateCmd = &cobra.Command{
		Use: "create", Short: "Submit a tuning job",
		Args: cobra.NoArgs, RunE: s.runAITJCreate,
	}
	s.aiTJDescribeCmd = &cobra.Command{
		Use: "describe TUNING_JOB", Short: "Describe a tuning job",
		Args: cobra.ExactArgs(1), RunE: s.runAITJDescribe,
	}
	s.aiTJListCmd = &cobra.Command{
		Use: "list", Short: "List tuning jobs",
		Args: cobra.NoArgs, RunE: s.runAITJList,
	}
	s.initAi()
	s.initAiCustomJobs()
	s.initAiEndpoints()
	s.initAiHpTuningJobs()
	s.initAiIndexEndpoints()
	s.initAiIndexes()
	s.initAiModelGarden()
	s.initAiModelMonitoringJobs()
	s.initAiModels()
	s.initAiOperations()
	s.initAiPersistentResources()
	s.initAiSemanticGovernance()
	s.initAiTensorboards()
	s.initAiTuningJobs()
	return s
}
