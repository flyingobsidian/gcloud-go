package cli

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

// --- firestore fields (has a single 'ttls' subgroup in gcloud-python) ---

func (g *firestoreCmds) initFirestoreFieldsIndexes() {
	for _, c := range []*cobra.Command{g.fsFieldsTTLListCmd, g.fsFieldsTTLUpdateCmd} {
		firestoreAddDatabaseFlag(c, &g.flagFSFieldsDatabase, true)
		c.Flags().StringVar(&g.flagFSFieldsCollection, "collection-group", "",
			"Collection group ID (defaults to '__default__' meaning all fields)")
	}
	g.fsFieldsTTLUpdateCmd.Flags().StringVar(&g.flagFSFieldsConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Field message body (required)")
	_ = g.fsFieldsTTLUpdateCmd.MarkFlagRequired("config-file")
	g.fsFieldsTTLUpdateCmd.Flags().StringVar(&g.flagFSFieldsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to 'ttlConfig')")
	g.fsFieldsTTLUpdateCmd.Flags().BoolVar(&g.flagFSFieldsAsync, "async", false, "Return the long-running operation without waiting")
	g.fsFieldsTTLListCmd.Flags().StringVar(&g.flagFSFieldsFormat, "format", "", "Output format")

	g.firestoreFieldsTTLsCmd.AddCommand(g.fsFieldsTTLListCmd, g.fsFieldsTTLUpdateCmd)
	g.firestoreFieldsCmd.AddCommand(g.firestoreFieldsTTLsCmd)
	g.firestoreCmd.AddCommand(g.firestoreFieldsCmd)
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

func (g *firestoreCmds) runFSFieldsTTLList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	group := g.flagFSFieldsCollection
	if group == "" {
		group = "__default__"
	}
	parent := fmt.Sprintf("%s/collectionGroups/%s", firestoreDatabaseName(project, g.flagFSFieldsDatabase), group)
	call := svc.Projects.Databases.CollectionGroups.Fields.List(parent).Context(ctx).
		Filter(`indexConfig.usesAncestorConfig:false AND ttlConfig:*`)
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing TTL fields: %w", err)
	}
	if g.flagFSFieldsFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Fields, g.flagFSFieldsFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Fields), "%-40s %s\n", "NAME", "TTL_STATE")
	for _, f := range resp.Fields {
		state := ""
		if f.TtlConfig != nil {
			state = f.TtlConfig.State
		}
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(f.Name), state)
	}
	return nil
}

func (g *firestoreCmds) runFSFieldsTTLUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	field := &firestore.GoogleFirestoreAdminV1Field{}
	if err := loadYAMLOrJSONInto(g.flagFSFieldsConfigFile, field); err != nil {
		return err
	}
	mask := g.flagFSFieldsUpdateMask
	if mask == "" {
		mask = "ttlConfig"
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.CollectionGroups.Fields.Patch(firestoreFieldName(args[0], project, g.flagFSFieldsDatabase, g.flagFSFieldsCollection), field).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating field TTL: %w", err)
	}
	return g.firestoreFinishOp(ctx, svc, op, "Update field TTL", args[0], g.flagFSFieldsAsync)
}

// --- firestore indexes (composite + fields subgroups) ---

func (g *firestoreCmds) initFirestoreFieldsIndexes5() {
	compositeCmds := []*cobra.Command{g.fsIdxCompCreateCmd, g.fsIdxCompDeleteCmd, g.fsIdxCompDescribeCmd, g.fsIdxCompListCmd}
	fieldsCmds := []*cobra.Command{g.fsIdxFieldsDescribeCmd, g.fsIdxFieldsListCmd, g.fsIdxFieldsUpdateCmd}
	for _, c := range append(compositeCmds, fieldsCmds...) {
		firestoreAddDatabaseFlag(c, &g.flagFSIdxDatabase, true)
		c.Flags().StringVar(&g.flagFSIdxCollection, "collection-group", "",
			"Collection group ID (defaults to '__default__')")
	}
	for _, c := range []*cobra.Command{g.fsIdxCompCreateCmd, g.fsIdxFieldsUpdateCmd} {
		c.Flags().StringVar(&g.flagFSIdxConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Index/Field message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.fsIdxFieldsUpdateCmd.Flags().StringVar(&g.flagFSIdxUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{g.fsIdxCompCreateCmd, g.fsIdxCompDeleteCmd, g.fsIdxFieldsUpdateCmd} {
		c.Flags().BoolVar(&g.flagFSIdxAsync, "async", false, "Return the long-running operation without waiting")
	}
	g.fsIdxCompDescribeCmd.Flags().StringVar(&g.flagFSIdxFormat, "format", "", "Output format")
	g.fsIdxCompListCmd.Flags().StringVar(&g.flagFSIdxFormat, "format", "", "Output format")
	g.fsIdxFieldsDescribeCmd.Flags().StringVar(&g.flagFSIdxFormat, "format", "", "Output format")
	g.fsIdxFieldsListCmd.Flags().StringVar(&g.flagFSIdxFormat, "format", "", "Output format")

	g.firestoreIndexesCompositeCmd.AddCommand(compositeCmds...)
	g.firestoreIndexesFieldsCmd.AddCommand(fieldsCmds...)
	g.firestoreIndexesCmd.AddCommand(g.firestoreIndexesCompositeCmd, g.firestoreIndexesFieldsCmd)
	g.firestoreCmd.AddCommand(g.firestoreIndexesCmd)
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

func (g *firestoreCmds) runFSIdxCompCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	idx := &firestore.GoogleFirestoreAdminV1Index{}
	if err := loadYAMLOrJSONInto(g.flagFSIdxConfigFile, idx); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.CollectionGroups.Indexes.Create(firestoreIndexParent(project, g.flagFSIdxDatabase, g.flagFSIdxCollection), idx).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating composite index: %w", err)
	}
	return g.firestoreFinishOp(ctx, svc, op, "Create composite index", "", g.flagFSIdxAsync)
}

func (g *firestoreCmds) runFSIdxCompDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Databases.CollectionGroups.Indexes.Delete(firestoreCompositeIndexName(args[0], project, g.flagFSIdxDatabase, g.flagFSIdxCollection)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting composite index: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted composite index [%s].\n", args[0])
	return nil
}

func (g *firestoreCmds) runFSIdxCompDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.CollectionGroups.Indexes.Get(firestoreCompositeIndexName(args[0], project, g.flagFSIdxDatabase, g.flagFSIdxCollection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing composite index: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagFSIdxFormat)
}

func (g *firestoreCmds) runFSIdxCompList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Databases.CollectionGroups.Indexes.List(firestoreIndexParent(project, g.flagFSIdxDatabase, g.flagFSIdxCollection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing composite indexes: %w", err)
	}
	if g.flagFSIdxFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Indexes, g.flagFSIdxFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Indexes), "%-40s %-15s %s\n", "NAME", "STATE", "QUERY_SCOPE")
	for _, i := range resp.Indexes {
		fmt.Fprintf(g.inv.stdout, "%-40s %-15s %s\n", path.Base(i.Name), i.State, i.QueryScope)
	}
	return nil
}

func (g *firestoreCmds) runFSIdxFieldsDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.CollectionGroups.Fields.Get(firestoreFieldName(args[0], project, g.flagFSIdxDatabase, g.flagFSIdxCollection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing field index: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagFSIdxFormat)
}

func (g *firestoreCmds) runFSIdxFieldsList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Databases.CollectionGroups.Fields.List(firestoreIndexParent(project, g.flagFSIdxDatabase, g.flagFSIdxCollection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing field indexes: %w", err)
	}
	if g.flagFSIdxFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Fields, g.flagFSIdxFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Fields), "%-40s\n", "NAME")
	for _, f := range resp.Fields {
		fmt.Fprintln(g.inv.stdout, path.Base(f.Name))
	}
	return nil
}

func (g *firestoreCmds) runFSIdxFieldsUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	field := &firestore.GoogleFirestoreAdminV1Field{}
	if err := loadYAMLOrJSONInto(g.flagFSIdxConfigFile, field); err != nil {
		return err
	}
	mask := g.flagFSIdxUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(field))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.CollectionGroups.Fields.Patch(firestoreFieldName(args[0], project, g.flagFSIdxDatabase, g.flagFSIdxCollection), field).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating field index: %w", err)
	}
	return g.firestoreFinishOp(ctx, svc, op, "Update field index", args[0], g.flagFSIdxAsync)
}
