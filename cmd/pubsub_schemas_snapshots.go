package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	pubsub "google.golang.org/api/pubsub/v1"
)

// --- gcloud pubsub schemas + snapshots (#1176, #1177) ---

func psSchemaName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/schemas/%s", project, id)
}

func psSnapshotName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/snapshots/%s", project, id)
}

var (
	flagPSSchemaType       string
	flagPSSchemaDefinition string
	flagPSSchemaDefFile    string
	flagPSSchemaMsgFile    string
	flagPSSchemaEncoding   string
	flagPSSchemaRevision   string

	flagPSSnapshotSub string
)

// --- schemas ---

var pubsubSchemasCmd = &cobra.Command{Use: "schemas", Short: "Manage Pub/Sub schemas"}

var (
	psSchemaCreateCmd = &cobra.Command{
		Use: "create SCHEMA", Short: "Create a Pub/Sub schema",
		Args: cobra.ExactArgs(1), RunE: runPSSchemaCreate,
	}
	psSchemaDeleteCmd = &cobra.Command{
		Use: "delete SCHEMA", Short: "Delete a Pub/Sub schema",
		Args: cobra.ExactArgs(1), RunE: runPSSchemaDelete,
	}
	psSchemaDeleteRevCmd = &cobra.Command{
		Use: "delete-revision SCHEMA", Short: "Delete a schema revision (--revision required)",
		Args: cobra.ExactArgs(1), RunE: runPSSchemaDeleteRev,
	}
	psSchemaDescribeCmd = &cobra.Command{
		Use: "describe SCHEMA", Short: "Describe a Pub/Sub schema",
		Args: cobra.ExactArgs(1), RunE: runPSSchemaDescribe,
	}
	psSchemaListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub schemas",
		Args: cobra.NoArgs, RunE: runPSSchemaList,
	}
	psSchemaListRevsCmd = &cobra.Command{
		Use: "list-revisions SCHEMA", Short: "List revisions of a Pub/Sub schema",
		Args: cobra.ExactArgs(1), RunE: runPSSchemaListRevs,
	}
	psSchemaCommitCmd = &cobra.Command{
		Use: "commit SCHEMA", Short: "Commit a new revision of a schema",
		Args: cobra.ExactArgs(1), RunE: runPSSchemaCommit,
	}
	psSchemaRollbackCmd = &cobra.Command{
		Use: "rollback SCHEMA", Short: "Roll back a schema to a given revision",
		Args: cobra.ExactArgs(1), RunE: runPSSchemaRollback,
	}
	psSchemaValidateSchemaCmd = &cobra.Command{
		Use: "validate-schema", Short: "Validate a schema definition without persisting it",
		Args: cobra.NoArgs, RunE: runPSSchemaValidateSchema,
	}
	psSchemaValidateMessageCmd = &cobra.Command{
		Use: "validate-message SCHEMA", Short: "Validate a message against a schema",
		Args: cobra.ExactArgs(1), RunE: runPSSchemaValidateMessage,
	}
)

// --- snapshots ---

var pubsubSnapshotsCmd = &cobra.Command{Use: "snapshots", Short: "Manage Pub/Sub snapshots"}

var (
	psSnapCreateCmd = &cobra.Command{
		Use: "create SNAPSHOT", Short: "Create a Pub/Sub snapshot",
		Args: cobra.ExactArgs(1), RunE: runPSSnapCreate,
	}
	psSnapDeleteCmd = &cobra.Command{
		Use: "delete SNAPSHOT", Short: "Delete a Pub/Sub snapshot",
		Args: cobra.ExactArgs(1), RunE: runPSSnapDelete,
	}
	psSnapDescribeCmd = &cobra.Command{
		Use: "describe SNAPSHOT", Short: "Describe a Pub/Sub snapshot",
		Args: cobra.ExactArgs(1), RunE: runPSSnapDescribe,
	}
	psSnapListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub snapshots",
		Args: cobra.NoArgs, RunE: runPSSnapList,
	}
)

func init() {
	// schemas
	for _, c := range []*cobra.Command{
		psSchemaCreateCmd, psSchemaDeleteCmd, psSchemaDeleteRevCmd, psSchemaDescribeCmd,
		psSchemaListCmd, psSchemaListRevsCmd, psSchemaCommitCmd, psSchemaRollbackCmd,
		psSchemaValidateSchemaCmd, psSchemaValidateMessageCmd,
	} {
		c.Flags().StringVar(&flagPSFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{psSchemaCreateCmd, psSchemaCommitCmd, psSchemaValidateSchemaCmd} {
		c.Flags().StringVar(&flagPSSchemaType, "type", "", "Schema type: AVRO or PROTOCOL_BUFFER (required)")
		c.Flags().StringVar(&flagPSSchemaDefinition, "definition", "", "Inline schema definition")
		c.Flags().StringVar(&flagPSSchemaDefFile, "definition-file", "",
			"Path to file with schema definition (mutually exclusive with --definition)")
		_ = c.MarkFlagRequired("type")
	}
	psSchemaDeleteRevCmd.Flags().StringVar(&flagPSSchemaRevision, "revision-id", "", "Revision ID to delete (required)")
	_ = psSchemaDeleteRevCmd.MarkFlagRequired("revision-id")
	psSchemaRollbackCmd.Flags().StringVar(&flagPSSchemaRevision, "revision-id", "", "Revision ID to roll back to (required)")
	_ = psSchemaRollbackCmd.MarkFlagRequired("revision-id")
	psSchemaValidateMessageCmd.Flags().StringVar(&flagPSSchemaMsgFile, "message-file", "",
		"Path to file with the message body (required)")
	_ = psSchemaValidateMessageCmd.MarkFlagRequired("message-file")
	psSchemaValidateMessageCmd.Flags().StringVar(&flagPSSchemaEncoding, "message-encoding", "JSON",
		"Message encoding: JSON or BINARY")
	pubsubSchemasCmd.AddCommand(
		psSchemaCreateCmd, psSchemaDeleteCmd, psSchemaDeleteRevCmd, psSchemaDescribeCmd,
		psSchemaListCmd, psSchemaListRevsCmd, psSchemaCommitCmd, psSchemaRollbackCmd,
		psSchemaValidateSchemaCmd, psSchemaValidateMessageCmd,
	)
	// pubsub.go still registers the schemas stub group; replace it with the real one.
	replacePubsubSubgroup("schemas", pubsubSchemasCmd)

	// snapshots
	psSnapCreateCmd.Flags().StringVar(&flagPSSnapshotSub, "subscription", "",
		"Subscription to snapshot from (required)")
	_ = psSnapCreateCmd.MarkFlagRequired("subscription")
	for _, c := range []*cobra.Command{psSnapDescribeCmd, psSnapListCmd} {
		c.Flags().StringVar(&flagPSFormat, "format", "", "Output format")
	}
	pubsubSnapshotsCmd.AddCommand(psSnapCreateCmd, psSnapDeleteCmd, psSnapDescribeCmd, psSnapListCmd)
	replacePubsubSubgroup("snapshots", pubsubSnapshotsCmd)
}

func replacePubsubSubgroup(name string, real *cobra.Command) {
	for _, c := range pubsubCmd.Commands() {
		if c.Name() == name {
			pubsubCmd.RemoveCommand(c)
			break
		}
	}
	pubsubCmd.AddCommand(real)
}

// --- schemas impl ---

func loadSchemaDefinition() (string, error) {
	if flagPSSchemaDefinition != "" && flagPSSchemaDefFile != "" {
		return "", fmt.Errorf("--definition and --definition-file are mutually exclusive")
	}
	if flagPSSchemaDefinition != "" {
		return flagPSSchemaDefinition, nil
	}
	if flagPSSchemaDefFile == "" {
		return "", fmt.Errorf("--definition or --definition-file is required")
	}
	data, err := os.ReadFile(flagPSSchemaDefFile)
	if err != nil {
		return "", fmt.Errorf("reading definition file: %w", err)
	}
	return string(data), nil
}

func runPSSchemaCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	def, err := loadSchemaDefinition()
	if err != nil {
		return err
	}
	schema := &pubsub.Schema{Type: flagPSSchemaType, Definition: def}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Schemas.Create(fmt.Sprintf("projects/%s", project), schema).
		SchemaId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating schema: %w", err)
	}
	return emitFormatted(got, "")
}

func runPSSchemaDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Schemas.Delete(psSchemaName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting schema: %w", err)
	}
	fmt.Printf("Deleted schema [%s].\n", args[0])
	return nil
}

func runPSSchemaDeleteRev(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s@%s", psSchemaName(project, args[0]), flagPSSchemaRevision)
	got, err := svc.Projects.Schemas.DeleteRevision(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting schema revision: %w", err)
	}
	return emitFormatted(got, "")
}

func runPSSchemaDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Schemas.Get(psSchemaName(project, args[0])).View("FULL").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing schema: %w", err)
	}
	return emitFormatted(got, flagPSFormat)
}

func runPSSchemaList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*pubsub.Schema
	pageToken := ""
	for {
		call := svc.Projects.Schemas.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing schemas: %w", err)
		}
		all = append(all, resp.Schemas...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagPSFormat != "" {
		return emitFormatted(all, flagPSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "TYPE")
	for _, s := range all {
		fmt.Printf("%-40s %s\n", path.Base(s.Name), s.Type)
	}
	return nil
}

func runPSSchemaListRevs(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Schemas.ListRevisions(psSchemaName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing schema revisions: %w", err)
	}
	if flagPSFormat != "" {
		return emitFormatted(resp.Schemas, flagPSFormat)
	}
	fmt.Printf("%-40s %s\n", "REVISION_ID", "CREATED")
	for _, s := range resp.Schemas {
		fmt.Printf("%-40s %s\n", s.RevisionId, s.RevisionCreateTime)
	}
	return nil
}

func runPSSchemaCommit(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	def, err := loadSchemaDefinition()
	if err != nil {
		return err
	}
	req := &pubsub.CommitSchemaRequest{Schema: &pubsub.Schema{Type: flagPSSchemaType, Definition: def}}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Schemas.Commit(psSchemaName(project, args[0]), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("committing schema revision: %w", err)
	}
	return emitFormatted(got, "")
}

func runPSSchemaRollback(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Schemas.Rollback(psSchemaName(project, args[0]),
		&pubsub.RollbackSchemaRequest{RevisionId: flagPSSchemaRevision}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rolling back schema: %w", err)
	}
	return emitFormatted(got, "")
}

func runPSSchemaValidateSchema(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	def, err := loadSchemaDefinition()
	if err != nil {
		return err
	}
	req := &pubsub.ValidateSchemaRequest{Schema: &pubsub.Schema{Type: flagPSSchemaType, Definition: def}}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Schemas.Validate(fmt.Sprintf("projects/%s", project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("validating schema: %w", err)
	}
	return emitFormatted(got, "")
}

func runPSSchemaValidateMessage(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	msg, err := os.ReadFile(flagPSSchemaMsgFile)
	if err != nil {
		return fmt.Errorf("reading message file: %w", err)
	}
	req := &pubsub.ValidateMessageRequest{
		Name:     psSchemaName(project, args[0]),
		Message:  string(msg),
		Encoding: flagPSSchemaEncoding,
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Schemas.ValidateMessage(fmt.Sprintf("projects/%s", project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("validating message: %w", err)
	}
	return emitFormatted(got, "")
}

// --- snapshots impl ---

func runPSSnapCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Snapshots.Create(psSnapshotName(project, args[0]),
		&pubsub.CreateSnapshotRequest{Subscription: psSubName(project, flagPSSnapshotSub)}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating snapshot: %w", err)
	}
	return emitFormatted(got, "")
}

func runPSSnapDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Snapshots.Delete(psSnapshotName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting snapshot: %w", err)
	}
	fmt.Printf("Deleted snapshot [%s].\n", args[0])
	return nil
}

func runPSSnapDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Snapshots.Get(psSnapshotName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing snapshot: %w", err)
	}
	return emitFormatted(got, flagPSFormat)
}

func runPSSnapList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*pubsub.Snapshot
	pageToken := ""
	for {
		call := svc.Projects.Snapshots.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing snapshots: %w", err)
		}
		all = append(all, resp.Snapshots...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagPSFormat != "" {
		return emitFormatted(all, flagPSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "TOPIC")
	for _, s := range all {
		fmt.Printf("%-40s %s\n", path.Base(s.Name), s.Topic)
	}
	return nil
}
