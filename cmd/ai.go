package cmd

import "github.com/spf13/cobra"

// --- gcloud ai (#291) ---

var aiCmd = &cobra.Command{Use: "ai", Short: "Manage Vertex AI (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(aiCmd, "custom-jobs", "Manage custom training jobs", "cancel", "create", "describe", "list", "stream-logs")
	registerStubGroup(aiCmd, "endpoints", "Manage endpoints", append(crud, "deploy-model", "undeploy-model", "predict", "explain", "raw-predict")...)
	registerStubGroup(aiCmd, "hp-tuning-jobs", "Manage hyperparameter tuning jobs", "cancel", "create", "describe", "list", "stream-logs")
	registerStubGroup(aiCmd, "index-endpoints", "Manage index endpoints", append(crud, "deploy-index", "undeploy-index", "mutate-deployed-index")...)
	registerStubGroup(aiCmd, "indexes", "Manage indexes", append(crud, "update-metadata", "remove-datapoints", "upsert-datapoints", "list-datapoints")...)
	registerStubGroup(aiCmd, "model-garden", "Vertex Model Garden", "models")
	registerStubGroup(aiCmd, "model-monitoring-jobs", "Manage monitoring jobs", "create", "delete", "describe", "list", "pause", "resume")
	registerStubGroup(aiCmd, "models", "Manage models", append(crud, "upload", "copy")...)
	registerStubGroup(aiCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list", "wait")
	registerStubGroup(aiCmd, "persistent-resources", "Manage persistent resources", "create", "delete", "describe", "list", "reboot")
	registerStubGroup(aiCmd, "tensorboards", "Manage Tensorboards", crud...)
	registerStubGroup(aiCmd, "tuning-jobs", "Manage tuning jobs", "cancel", "create", "describe", "list")
	rootCmd.AddCommand(aiCmd)
}
