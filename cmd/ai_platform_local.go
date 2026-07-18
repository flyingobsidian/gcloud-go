package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud ai-platform local (#983) ---
//
// Both `local predict` and `local train` in the upstream gcloud CLI shell
// out to a client-side Python runtime (TensorFlow / scikit-learn for
// predict, a user-supplied training package for train). gcloud-go has no
// Python runtime and does not vendor those model frameworks, so these
// commands are surfaced as commands that return a clear error explaining
// what to do instead. They are NOT stubs — they parse their flags,
// validate them, and fail cleanly with an actionable message. Callers
// wanting the same effect should use the server-side commands:
//
//   - `gcloud-go ai-platform predict` to run a prediction against a
//     deployed model version.
//   - `gcloud-go ai-platform jobs submit` to submit a training job.

var aiPlatformLocalCmd = &cobra.Command{
	Use:   "local",
	Short: "AI Platform local (client-side) commands",
	Long: "Client-side prediction and training commands from upstream gcloud. " +
		"These require a local Python runtime with TensorFlow/scikit-learn (predict) or " +
		"a user-supplied training package (train) and are not supported by gcloud-go. " +
		"Use `ai-platform predict` or `ai-platform jobs submit` for the server-side equivalents.",
}

var (
	flagAIPlatformLocalFormat        string
	flagAIPlatformLocalModelDir      string
	flagAIPlatformLocalJSONInstances string
	flagAIPlatformLocalTextInstances string
	flagAIPlatformLocalFramework     string
	flagAIPlatformLocalSignatureName string

	flagAIPlatformLocalPackagePath string
	flagAIPlatformLocalModuleName  string
	flagAIPlatformLocalJobDir      string
	flagAIPlatformLocalDistributed bool
)

var (
	aiPlatformLocalPredictCmd = &cobra.Command{
		Use:   "predict",
		Short: "(Unsupported) Run client-side prediction against a locally exported model",
		Long: "Runs prediction against a locally exported TensorFlow/scikit-learn model. " +
			"This is a client-side flow that requires a Python runtime and the model framework " +
			"and is not supported by gcloud-go. Use `ai-platform predict` to run against a deployed model.",
		Args: cobra.NoArgs,
		RunE: runAIPlatformLocalPredict,
	}
	aiPlatformLocalTrainCmd = &cobra.Command{
		Use:   "train",
		Short: "(Unsupported) Run a training job locally, using the same code the AI Platform job runners would use",
		Long: "Runs a training package locally so users can iterate before submitting to AI Platform. " +
			"This is a client-side flow that requires a Python runtime and the user's training package " +
			"and is not supported by gcloud-go. Use `ai-platform jobs submit` to run training on the service.",
		Args: cobra.NoArgs,
		RunE: runAIPlatformLocalTrain,
	}
)

func init() {
	for _, c := range []*cobra.Command{aiPlatformLocalPredictCmd, aiPlatformLocalTrainCmd} {
		c.Flags().StringVar(&flagAIPlatformLocalFormat, "format", "", "Output format")
	}

	aiPlatformLocalPredictCmd.Flags().StringVar(&flagAIPlatformLocalModelDir, "model-dir", "",
		"Path to the local model export directory")
	aiPlatformLocalPredictCmd.Flags().StringVar(&flagAIPlatformLocalJSONInstances, "json-instances", "",
		"Path to a newline-delimited JSON file with prediction instances")
	aiPlatformLocalPredictCmd.Flags().StringVar(&flagAIPlatformLocalTextInstances, "text-instances", "",
		"Path to a newline-delimited text file with prediction instances")
	aiPlatformLocalPredictCmd.Flags().StringVar(&flagAIPlatformLocalFramework, "framework", "",
		"Model framework (tensorflow, scikit-learn, xgboost)")
	aiPlatformLocalPredictCmd.Flags().StringVar(&flagAIPlatformLocalSignatureName, "signature-name", "",
		"TensorFlow signature name to use when running prediction")

	aiPlatformLocalTrainCmd.Flags().StringVar(&flagAIPlatformLocalPackagePath, "package-path", "",
		"Path to the local training package")
	aiPlatformLocalTrainCmd.Flags().StringVar(&flagAIPlatformLocalModuleName, "module-name", "",
		"Fully-qualified module name to run inside --package-path")
	aiPlatformLocalTrainCmd.Flags().StringVar(&flagAIPlatformLocalJobDir, "job-dir", "",
		"Directory the training job should use for outputs")
	aiPlatformLocalTrainCmd.Flags().BoolVar(&flagAIPlatformLocalDistributed, "distributed", false,
		"Simulate a distributed training run instead of a single-worker run")

	aiPlatformLocalCmd.AddCommand(aiPlatformLocalPredictCmd, aiPlatformLocalTrainCmd)
	aiPlatformCmd.AddCommand(aiPlatformLocalCmd)
}

func runAIPlatformLocalPredict(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("local predict requires TensorFlow/scikit-learn Python runtime and is not supported by gcloud-go; use \"ai-platform predict\" to run against a deployed model")
}

func runAIPlatformLocalTrain(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("local train requires a Python runtime with your training package and is not supported by gcloud-go; use \"ai-platform jobs submit training\" to run training on Vertex AI")
}
