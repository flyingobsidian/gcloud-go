package cli

import (
	"fmt"
	"os"
	"path"
	"strings"

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

// --- schemas ---

// --- snapshots ---

func (g *pubsubCmds) initPubsubSchemasSnapshots() {
	// schemas
	for _, c := range []*cobra.Command{
		g.psSchemaCreateCmd, g.psSchemaDeleteCmd, g.psSchemaDeleteRevCmd, g.psSchemaDescribeCmd,
		g.psSchemaListCmd, g.psSchemaListRevsCmd, g.psSchemaCommitCmd, g.psSchemaRollbackCmd,
		g.psSchemaValidateSchemaCmd, g.psSchemaValidateMessageCmd,
	} {
		c.Flags().StringVar(&g.flagPSFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{g.psSchemaCreateCmd, g.psSchemaCommitCmd, g.psSchemaValidateSchemaCmd} {
		c.Flags().StringVar(&g.flagPSSchemaType, "type", "", "Schema type: AVRO or PROTOCOL_BUFFER (required)")
		c.Flags().StringVar(&g.flagPSSchemaDefinition, "definition", "", "Inline schema definition")
		c.Flags().StringVar(&g.flagPSSchemaDefFile, "definition-file", "",
			"Path to file with schema definition (mutually exclusive with --definition)")
		_ = c.MarkFlagRequired("type")
	}
	g.psSchemaDeleteRevCmd.Flags().StringVar(&g.flagPSSchemaRevision, "revision-id", "", "Revision ID to delete (required)")
	_ = g.psSchemaDeleteRevCmd.MarkFlagRequired("revision-id")
	g.psSchemaRollbackCmd.Flags().StringVar(&g.flagPSSchemaRevision, "revision-id", "", "Revision ID to roll back to (required)")
	_ = g.psSchemaRollbackCmd.MarkFlagRequired("revision-id")
	g.psSchemaValidateMessageCmd.Flags().StringVar(&g.flagPSSchemaMsgFile, "message-file", "",
		"Path to file with the message body (required)")
	_ = g.psSchemaValidateMessageCmd.MarkFlagRequired("message-file")
	g.psSchemaValidateMessageCmd.Flags().StringVar(&g.flagPSSchemaEncoding, "message-encoding", "JSON",
		"Message encoding: JSON or BINARY")
	g.pubsubSchemasCmd.AddCommand(
		g.psSchemaCreateCmd, g.psSchemaDeleteCmd, g.psSchemaDeleteRevCmd, g.psSchemaDescribeCmd,
		g.psSchemaListCmd, g.psSchemaListRevsCmd, g.psSchemaCommitCmd, g.psSchemaRollbackCmd,
		g.psSchemaValidateSchemaCmd, g.psSchemaValidateMessageCmd,
	)
	// pubsub.go still registers the schemas stub group; replace it with the real one.
	g.replacePubsubSubgroup("schemas", g.pubsubSchemasCmd)

	// snapshots
	g.psSnapCreateCmd.Flags().StringVar(&g.flagPSSnapshotSub, "subscription", "",
		"Subscription to snapshot from (required)")
	_ = g.psSnapCreateCmd.MarkFlagRequired("subscription")
	for _, c := range []*cobra.Command{g.psSnapDescribeCmd, g.psSnapListCmd} {
		c.Flags().StringVar(&g.flagPSFormat, "format", "", "Output format")
	}
	g.pubsubSnapshotsCmd.AddCommand(g.psSnapCreateCmd, g.psSnapDeleteCmd, g.psSnapDescribeCmd, g.psSnapListCmd)
	g.replacePubsubSubgroup("snapshots", g.pubsubSnapshotsCmd)
}

func (g *pubsubCmds) replacePubsubSubgroup(name string, real *cobra.Command) {
	for _, c := range g.pubsubCmd.Commands() {
		if c.Name() == name {
			g.pubsubCmd.RemoveCommand(c)
			break
		}
	}
	g.pubsubCmd.AddCommand(real)
}

// --- schemas impl ---

func (g *pubsubCmds) loadSchemaDefinition() (string, error) {
	if g.flagPSSchemaDefinition != "" && g.flagPSSchemaDefFile != "" {
		return "", fmt.Errorf("--definition and --definition-file are mutually exclusive")
	}
	if g.flagPSSchemaDefinition != "" {
		return g.flagPSSchemaDefinition, nil
	}
	if g.flagPSSchemaDefFile == "" {
		return "", fmt.Errorf("--definition or --definition-file is required")
	}
	data, err := os.ReadFile(g.flagPSSchemaDefFile)
	if err != nil {
		return "", fmt.Errorf("reading definition file: %w", err)
	}
	return string(data), nil
}

func (g *pubsubCmds) runPSSchemaCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	def, err := g.loadSchemaDefinition()
	if err != nil {
		return err
	}
	schema := &pubsub.Schema{Type: g.flagPSSchemaType, Definition: def}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Schemas.Create(fmt.Sprintf("projects/%s", project), schema).
		SchemaId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating schema: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *pubsubCmds) runPSSchemaDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Schemas.Delete(psSchemaName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting schema: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted schema [%s].\n", args[0])
	return nil
}

func (g *pubsubCmds) runPSSchemaDeleteRev(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s@%s", psSchemaName(project, args[0]), g.flagPSSchemaRevision)
	got, err := svc.Projects.Schemas.DeleteRevision(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting schema revision: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *pubsubCmds) runPSSchemaDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Schemas.Get(psSchemaName(project, args[0])).View("FULL").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing schema: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPSFormat)
}

func (g *pubsubCmds) runPSSchemaList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
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
	if g.flagPSFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagPSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "TYPE")
	for _, s := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(s.Name), s.Type)
	}
	return nil
}

func (g *pubsubCmds) runPSSchemaListRevs(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Schemas.ListRevisions(psSchemaName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing schema revisions: %w", err)
	}
	if g.flagPSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Schemas, g.flagPSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Schemas), "%-40s %s\n", "REVISION_ID", "CREATED")
	for _, s := range resp.Schemas {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", s.RevisionId, s.RevisionCreateTime)
	}
	return nil
}

func (g *pubsubCmds) runPSSchemaCommit(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	def, err := g.loadSchemaDefinition()
	if err != nil {
		return err
	}
	req := &pubsub.CommitSchemaRequest{Schema: &pubsub.Schema{Type: g.flagPSSchemaType, Definition: def}}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Schemas.Commit(psSchemaName(project, args[0]), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("committing schema revision: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *pubsubCmds) runPSSchemaRollback(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Schemas.Rollback(psSchemaName(project, args[0]),
		&pubsub.RollbackSchemaRequest{RevisionId: g.flagPSSchemaRevision}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rolling back schema: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *pubsubCmds) runPSSchemaValidateSchema(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	def, err := g.loadSchemaDefinition()
	if err != nil {
		return err
	}
	req := &pubsub.ValidateSchemaRequest{Schema: &pubsub.Schema{Type: g.flagPSSchemaType, Definition: def}}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Schemas.Validate(fmt.Sprintf("projects/%s", project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("validating schema: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *pubsubCmds) runPSSchemaValidateMessage(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	msg, err := os.ReadFile(g.flagPSSchemaMsgFile)
	if err != nil {
		return fmt.Errorf("reading message file: %w", err)
	}
	req := &pubsub.ValidateMessageRequest{
		Name:     psSchemaName(project, args[0]),
		Message:  string(msg),
		Encoding: g.flagPSSchemaEncoding,
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Schemas.ValidateMessage(fmt.Sprintf("projects/%s", project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("validating message: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

// --- snapshots impl ---

func (g *pubsubCmds) runPSSnapCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Snapshots.Create(psSnapshotName(project, args[0]),
		&pubsub.CreateSnapshotRequest{Subscription: psSubName(project, g.flagPSSnapshotSub)}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating snapshot: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *pubsubCmds) runPSSnapDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Snapshots.Delete(psSnapshotName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting snapshot: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted snapshot [%s].\n", args[0])
	return nil
}

func (g *pubsubCmds) runPSSnapDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Snapshots.Get(psSnapshotName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing snapshot: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPSFormat)
}

func (g *pubsubCmds) runPSSnapList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
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
	if g.flagPSFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagPSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "TOPIC")
	for _, s := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(s.Name), s.Topic)
	}
	return nil
}
