package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

// --- firestore fields (has a single 'ttls' subgroup in gcloud-python) ---

var firestoreFieldsCmd = &cobra.Command{
	Use:   "fields",
	Short: "Manage single-field indexes for Cloud Firestore",
}

var firestoreFieldsTTLsCmd = &cobra.Command{
	Use:   "ttls",
	Short: "Manage Time-to-Live metadata for Cloud Firestore fields",
}

var (
	fsFieldsTTLListCmd = &cobra.Command{
		Use: "list", Short: "List TTL-configured fields for a collection group",
		Args: cobra.NoArgs, RunE: runFSFieldsTTLList,
	}
	fsFieldsTTLUpdateCmd = &cobra.Command{
		Use: "update FIELD", Short: "Update a field's TTL configuration from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runFSFieldsTTLUpdate,
	}
)

var (
	flagFSFieldsDatabase   string
	flagFSFieldsCollection string
	flagFSFieldsConfigFile string
	flagFSFieldsUpdateMask string
	flagFSFieldsFormat     string
	flagFSFieldsAsync      bool
)

func init() {
	for _, c := range []*cobra.Command{fsFieldsTTLListCmd, fsFieldsTTLUpdateCmd} {
		firestoreAddDatabaseFlag(c, &flagFSFieldsDatabase, true)
		c.Flags().StringVar(&flagFSFieldsCollection, "collection-group", "",
			"Collection group ID (defaults to '__default__' meaning all fields)")
	}
	fsFieldsTTLUpdateCmd.Flags().StringVar(&flagFSFieldsConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Field message body (required)")
	_ = fsFieldsTTLUpdateCmd.MarkFlagRequired("config-file")
	fsFieldsTTLUpdateCmd.Flags().StringVar(&flagFSFieldsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to 'ttlConfig')")
	fsFieldsTTLUpdateCmd.Flags().BoolVar(&flagFSFieldsAsync, "async", false, "Return the long-running operation without waiting")
	fsFieldsTTLListCmd.Flags().StringVar(&flagFSFieldsFormat, "format", "", "Output format")

	firestoreFieldsTTLsCmd.AddCommand(fsFieldsTTLListCmd, fsFieldsTTLUpdateCmd)
	firestoreFieldsCmd.AddCommand(firestoreFieldsTTLsCmd)
	firestoreCmd.AddCommand(firestoreFieldsCmd)
}

func firestoreFieldName(id, project, db, collection string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	group := collection
	if group == "" {
		group = "__default__"
	}
	return fmt.Sprintf("%s/collectionGroups/%s/fields/%s", firestoreDatabaseName(project, db), group, id)
}

func runFSFieldsTTLList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	group := flagFSFieldsCollection
	if group == "" {
		group = "__default__"
	}
	parent := fmt.Sprintf("%s/collectionGroups/%s", firestoreDatabaseName(project, flagFSFieldsDatabase), group)
	call := svc.Projects.Databases.CollectionGroups.Fields.List(parent).Context(ctx).
		Filter(`indexConfig.usesAncestorConfig:false AND ttlConfig:*`)
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing TTL fields: %w", err)
	}
	if flagFSFieldsFormat != "" {
		return emitFormatted(resp.Fields, flagFSFieldsFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "TTL_STATE")
	for _, f := range resp.Fields {
		state := ""
		if f.TtlConfig != nil {
			state = f.TtlConfig.State
		}
		fmt.Printf("%-40s %s\n", path.Base(f.Name), state)
	}
	return nil
}

func runFSFieldsTTLUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	field := &firestore.GoogleFirestoreAdminV1Field{}
	if err := loadYAMLOrJSONInto(flagFSFieldsConfigFile, field); err != nil {
		return err
	}
	mask := flagFSFieldsUpdateMask
	if mask == "" {
		mask = "ttlConfig"
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.CollectionGroups.Fields.Patch(firestoreFieldName(args[0], project, flagFSFieldsDatabase, flagFSFieldsCollection), field).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating field TTL: %w", err)
	}
	return firestoreFinishOp(ctx, svc, op, "Update field TTL", args[0], flagFSFieldsAsync)
}

// --- firestore indexes (composite + fields subgroups) ---

var firestoreIndexesCmd = &cobra.Command{
	Use:   "indexes",
	Short: "Manage Cloud Firestore indexes",
}

var firestoreIndexesCompositeCmd = &cobra.Command{
	Use:   "composite",
	Short: "Manage composite indexes for Cloud Firestore",
}

var firestoreIndexesFieldsCmd = &cobra.Command{
	Use:   "fields",
	Short: "Manage single-field indexes for Cloud Firestore",
}

var (
	fsIdxCompCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a composite index from a --config-file",
		Args: cobra.NoArgs, RunE: runFSIdxCompCreate,
	}
	fsIdxCompDeleteCmd = &cobra.Command{
		Use: "delete INDEX", Short: "Delete a composite index",
		Args: cobra.ExactArgs(1), RunE: runFSIdxCompDelete,
	}
	fsIdxCompDescribeCmd = &cobra.Command{
		Use: "describe INDEX", Short: "Describe a composite index",
		Args: cobra.ExactArgs(1), RunE: runFSIdxCompDescribe,
	}
	fsIdxCompListCmd = &cobra.Command{
		Use: "list", Short: "List composite indexes for a collection group",
		Args: cobra.NoArgs, RunE: runFSIdxCompList,
	}
	fsIdxFieldsDescribeCmd = &cobra.Command{
		Use: "describe FIELD", Short: "Describe a single-field index",
		Args: cobra.ExactArgs(1), RunE: runFSIdxFieldsDescribe,
	}
	fsIdxFieldsListCmd = &cobra.Command{
		Use: "list", Short: "List single-field indexes for a collection group",
		Args: cobra.NoArgs, RunE: runFSIdxFieldsList,
	}
	fsIdxFieldsUpdateCmd = &cobra.Command{
		Use: "update FIELD", Short: "Update a single-field index from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runFSIdxFieldsUpdate,
	}
)

var (
	flagFSIdxDatabase   string
	flagFSIdxCollection string
	flagFSIdxConfigFile string
	flagFSIdxUpdateMask string
	flagFSIdxFormat     string
	flagFSIdxAsync      bool
)

func init() {
	compositeCmds := []*cobra.Command{fsIdxCompCreateCmd, fsIdxCompDeleteCmd, fsIdxCompDescribeCmd, fsIdxCompListCmd}
	fieldsCmds := []*cobra.Command{fsIdxFieldsDescribeCmd, fsIdxFieldsListCmd, fsIdxFieldsUpdateCmd}
	for _, c := range append(compositeCmds, fieldsCmds...) {
		firestoreAddDatabaseFlag(c, &flagFSIdxDatabase, true)
		c.Flags().StringVar(&flagFSIdxCollection, "collection-group", "",
			"Collection group ID (defaults to '__default__')")
	}
	for _, c := range []*cobra.Command{fsIdxCompCreateCmd, fsIdxFieldsUpdateCmd} {
		c.Flags().StringVar(&flagFSIdxConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Index/Field message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	fsIdxFieldsUpdateCmd.Flags().StringVar(&flagFSIdxUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{fsIdxCompCreateCmd, fsIdxCompDeleteCmd, fsIdxFieldsUpdateCmd} {
		c.Flags().BoolVar(&flagFSIdxAsync, "async", false, "Return the long-running operation without waiting")
	}
	fsIdxCompDescribeCmd.Flags().StringVar(&flagFSIdxFormat, "format", "", "Output format")
	fsIdxCompListCmd.Flags().StringVar(&flagFSIdxFormat, "format", "", "Output format")
	fsIdxFieldsDescribeCmd.Flags().StringVar(&flagFSIdxFormat, "format", "", "Output format")
	fsIdxFieldsListCmd.Flags().StringVar(&flagFSIdxFormat, "format", "", "Output format")

	firestoreIndexesCompositeCmd.AddCommand(compositeCmds...)
	firestoreIndexesFieldsCmd.AddCommand(fieldsCmds...)
	firestoreIndexesCmd.AddCommand(firestoreIndexesCompositeCmd, firestoreIndexesFieldsCmd)
	firestoreCmd.AddCommand(firestoreIndexesCmd)
}

func firestoreIndexParent(project, db, group string) string {
	if group == "" {
		group = "__default__"
	}
	return fmt.Sprintf("%s/collectionGroups/%s", firestoreDatabaseName(project, db), group)
}

func firestoreCompositeIndexName(id, project, db, group string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/indexes/%s", firestoreIndexParent(project, db, group), id)
}

func runFSIdxCompCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	idx := &firestore.GoogleFirestoreAdminV1Index{}
	if err := loadYAMLOrJSONInto(flagFSIdxConfigFile, idx); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.CollectionGroups.Indexes.Create(firestoreIndexParent(project, flagFSIdxDatabase, flagFSIdxCollection), idx).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating composite index: %w", err)
	}
	return firestoreFinishOp(ctx, svc, op, "Create composite index", "", flagFSIdxAsync)
}

func runFSIdxCompDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Databases.CollectionGroups.Indexes.Delete(firestoreCompositeIndexName(args[0], project, flagFSIdxDatabase, flagFSIdxCollection)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting composite index: %w", err)
	}
	fmt.Printf("Deleted composite index [%s].\n", args[0])
	return nil
}

func runFSIdxCompDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.CollectionGroups.Indexes.Get(firestoreCompositeIndexName(args[0], project, flagFSIdxDatabase, flagFSIdxCollection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing composite index: %w", err)
	}
	return emitFormatted(got, flagFSIdxFormat)
}

func runFSIdxCompList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Databases.CollectionGroups.Indexes.List(firestoreIndexParent(project, flagFSIdxDatabase, flagFSIdxCollection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing composite indexes: %w", err)
	}
	if flagFSIdxFormat != "" {
		return emitFormatted(resp.Indexes, flagFSIdxFormat)
	}
	fmt.Printf("%-40s %-15s %s\n", "NAME", "STATE", "QUERY_SCOPE")
	for _, i := range resp.Indexes {
		fmt.Printf("%-40s %-15s %s\n", path.Base(i.Name), i.State, i.QueryScope)
	}
	return nil
}

func runFSIdxFieldsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.CollectionGroups.Fields.Get(firestoreFieldName(args[0], project, flagFSIdxDatabase, flagFSIdxCollection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing field index: %w", err)
	}
	return emitFormatted(got, flagFSIdxFormat)
}

func runFSIdxFieldsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Databases.CollectionGroups.Fields.List(firestoreIndexParent(project, flagFSIdxDatabase, flagFSIdxCollection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing field indexes: %w", err)
	}
	if flagFSIdxFormat != "" {
		return emitFormatted(resp.Fields, flagFSIdxFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, f := range resp.Fields {
		fmt.Println(path.Base(f.Name))
	}
	return nil
}

func runFSIdxFieldsUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	field := &firestore.GoogleFirestoreAdminV1Field{}
	if err := loadYAMLOrJSONInto(flagFSIdxConfigFile, field); err != nil {
		return err
	}
	mask := flagFSIdxUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(field))
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.CollectionGroups.Fields.Patch(firestoreFieldName(args[0], project, flagFSIdxDatabase, flagFSIdxCollection), field).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating field index: %w", err)
	}
	return firestoreFinishOp(ctx, svc, op, "Update field index", args[0], flagFSIdxAsync)
}
