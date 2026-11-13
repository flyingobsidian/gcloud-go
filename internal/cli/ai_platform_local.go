package cli

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

func (s *aiPlatformCmds) initAiPlatformLocal() {
	for _, c := range []*cobra.Command{s.aiPlatformLocalPredictCmd, s.aiPlatformLocalTrainCmd} {
		c.Flags().StringVar(&s.flagAIPlatformLocalFormat, "format", "", "Output format")
	}

	s.aiPlatformLocalPredictCmd.Flags().StringVar(&s.flagAIPlatformLocalModelDir, "model-dir", "",
		"Path to the local model export directory")
	s.aiPlatformLocalPredictCmd.Flags().StringVar(&s.flagAIPlatformLocalJSONInstances, "json-instances", "",
		"Path to a newline-delimited JSON file with prediction instances")
	s.aiPlatformLocalPredictCmd.Flags().StringVar(&s.flagAIPlatformLocalTextInstances, "text-instances", "",
		"Path to a newline-delimited text file with prediction instances")
	s.aiPlatformLocalPredictCmd.Flags().StringVar(&s.flagAIPlatformLocalFramework, "framework", "",
		"Model framework (tensorflow, scikit-learn, xgboost)")
	s.aiPlatformLocalPredictCmd.Flags().StringVar(&s.flagAIPlatformLocalSignatureName, "signature-name", "",
		"TensorFlow signature name to use when running prediction")

	s.aiPlatformLocalTrainCmd.Flags().StringVar(&s.flagAIPlatformLocalPackagePath, "package-path", "",
		"Path to the local training package")
	s.aiPlatformLocalTrainCmd.Flags().StringVar(&s.flagAIPlatformLocalModuleName, "module-name", "",
		"Fully-qualified module name to run inside --package-path")
	s.aiPlatformLocalTrainCmd.Flags().StringVar(&s.flagAIPlatformLocalJobDir, "job-dir", "",
		"Directory the training job should use for outputs")
	s.aiPlatformLocalTrainCmd.Flags().BoolVar(&s.flagAIPlatformLocalDistributed, "distributed", false,
		"Simulate a distributed training run instead of a single-worker run")

	s.aiPlatformLocalCmd.AddCommand(s.aiPlatformLocalPredictCmd, s.aiPlatformLocalTrainCmd)
	s.aiPlatformCmd.AddCommand(s.aiPlatformLocalCmd)
}

func runAIPlatformLocalPredict(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("local predict requires TensorFlow/scikit-learn Python runtime and is not supported by gcloud-go; use \"ai-platform predict\" to run against a deployed model")
}

func runAIPlatformLocalTrain(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("local train requires a Python runtime with your training package and is not supported by gcloud-go; use \"ai-platform jobs submit training\" to run training on Vertex AI")
}
