package cmd

import "github.com/spf13/cobra"

// --- gcloud recommender (#379) ---

var recommenderCmd = &cobra.Command{Use: "recommender", Short: "Manage Cloud Recommender"}

func init() {
	registerStubGroup(recommenderCmd, "insight-type-config", "Manage insight type configuration", "describe", "update")
	registerStubGroup(recommenderCmd, "insights", "Manage insights", "describe", "list", "mark-accepted", "mark-active", "mark-dismissed")
	registerStubGroup(recommenderCmd, "recommendations", "Manage recommendations", "describe", "list", "mark-active", "mark-claimed", "mark-dismissed", "mark-failed", "mark-succeeded")
	registerStubGroup(recommenderCmd, "recommender-config", "Manage recommender configuration", "describe", "update")
	rootCmd.AddCommand(recommenderCmd)
}
