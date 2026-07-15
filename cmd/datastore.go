package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	datastore "google.golang.org/api/datastore/v1"
)

// --- gcloud datastore (#325, #969, #970) ---

var datastoreCmd = &cobra.Command{Use: "datastore", Short: "Manage Cloud Datastore"}

var (
	flagDatastoreFile     string
	flagDatastoreKind     string
	flagDatastoreAncestor string
	flagDatastoreProps    []string
	flagDatastoreFilter   string
	flagDatastorePageSize int64
	flagDatastoreFormat   string
)

// --- indexes ---

var datastoreIndexesCmd = &cobra.Command{Use: "indexes", Short: "Manage Datastore indexes"}

var (
	datastoreIndexesCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a Datastore index",
		Args:  cobra.NoArgs,
		RunE:  runDSIndexesCreate,
	}
	datastoreIndexesDescribeCmd = &cobra.Command{
		Use:   "describe INDEX_ID",
		Short: "Describe a Datastore index",
		Args:  cobra.ExactArgs(1),
		RunE:  runDSIndexesDescribe,
	}
	datastoreIndexesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Datastore indexes",
		Args:  cobra.NoArgs,
		RunE:  runDSIndexesList,
	}
	datastoreIndexesCleanupCmd = &cobra.Command{
		Use:   "cleanup",
		Short: "Delete unused Datastore indexes (indexes present in the project but not in the provided --file)",
		Args:  cobra.NoArgs,
		RunE:  runDSIndexesCleanup,
	}
)

// --- operations ---

var datastoreOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Datastore long-running operations"}

var (
	datastoreOperationsCancelCmd = &cobra.Command{
		Use:   "cancel OPERATION",
		Short: "Cancel a Datastore operation",
		Args:  cobra.ExactArgs(1),
		RunE:  runDSOperationsCancel,
	}
	datastoreOperationsDescribeCmd = &cobra.Command{
		Use:   "describe OPERATION",
		Short: "Describe a Datastore operation",
		Args:  cobra.ExactArgs(1),
		RunE:  runDSOperationsDescribe,
	}
	datastoreOperationsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Datastore operations",
		Args:  cobra.NoArgs,
		RunE:  runDSOperationsList,
	}
)

func init() {
	// indexes flags
	datastoreIndexesCreateCmd.Flags().StringVar(&flagDatastoreKind, "kind", "", "Entity kind (required unless --file provides indexes)")
	datastoreIndexesCreateCmd.Flags().StringVar(&flagDatastoreAncestor, "ancestor", "NONE", "Ancestor mode: NONE or ALL_ANCESTORS")
	datastoreIndexesCreateCmd.Flags().StringSliceVar(&flagDatastoreProps, "property", nil, "Indexed property in NAME:DIRECTION form (may be repeated). DIRECTION is ASCENDING or DESCENDING")
	datastoreIndexesCreateCmd.Flags().StringVar(&flagDatastoreFile, "file", "", "Path to a YAML/JSON file containing one index or a list of indexes to create")
	datastoreIndexesCleanupCmd.Flags().StringVar(&flagDatastoreFile, "file", "", "Path to a YAML/JSON file describing the indexes that should be kept (required)")
	_ = datastoreIndexesCleanupCmd.MarkFlagRequired("file")
	datastoreIndexesListCmd.Flags().StringVar(&flagDatastoreFilter, "filter", "", "Server-side filter expression")
	datastoreIndexesListCmd.Flags().Int64Var(&flagDatastorePageSize, "page-size", 0, "Maximum number of results per page")
	for _, c := range []*cobra.Command{datastoreIndexesDescribeCmd, datastoreIndexesListCmd, datastoreIndexesCreateCmd} {
		c.Flags().StringVar(&flagDatastoreFormat, "format", "", "Output format")
	}
	datastoreIndexesCmd.AddCommand(datastoreIndexesCreateCmd, datastoreIndexesDescribeCmd, datastoreIndexesListCmd, datastoreIndexesCleanupCmd)

	// operations flags
	datastoreOperationsListCmd.Flags().StringVar(&flagDatastoreFilter, "filter", "", "Server-side filter expression")
	datastoreOperationsListCmd.Flags().Int64Var(&flagDatastorePageSize, "page-size", 0, "Maximum number of results per page")
	for _, c := range []*cobra.Command{datastoreOperationsDescribeCmd, datastoreOperationsListCmd} {
		c.Flags().StringVar(&flagDatastoreFormat, "format", "", "Output format")
	}
	datastoreOperationsCmd.AddCommand(datastoreOperationsCancelCmd, datastoreOperationsDescribeCmd, datastoreOperationsListCmd)

	datastoreCmd.AddCommand(datastoreIndexesCmd, datastoreOperationsCmd)
	for _, name := range []string{"export", "import"} {
		registerStubCommand(datastoreCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(datastoreCmd)
}

// --- helpers ---

// dsIndexFile mirrors the shape of the YAML/JSON documents accepted by the
// gcloud CLI: either a single index or a `{ "indexes": [ ... ] }` list.
type dsIndexFile struct {
	Indexes []dsIndexSpec `json:"indexes,omitempty" yaml:"indexes,omitempty"`
	dsIndexSpec
}

type dsIndexSpec struct {
	Kind       string             `json:"kind,omitempty" yaml:"kind,omitempty"`
	Ancestor   string             `json:"ancestor,omitempty" yaml:"ancestor,omitempty"`
	Properties []dsIndexPropSpec  `json:"properties,omitempty" yaml:"properties,omitempty"`
}

type dsIndexPropSpec struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Direction string `json:"direction,omitempty" yaml:"direction,omitempty"`
}

func dsIndexesFromFile(path string) ([]*datastore.GoogleDatastoreAdminV1Index, error) {
	var doc dsIndexFile
	if err := loadYAMLOrJSONInto(path, &doc); err != nil {
		return nil, err
	}
	specs := doc.Indexes
	if len(specs) == 0 && doc.Kind != "" {
		specs = []dsIndexSpec{doc.dsIndexSpec}
	}
	out := make([]*datastore.GoogleDatastoreAdminV1Index, 0, len(specs))
	for _, s := range specs {
		idx := &datastore.GoogleDatastoreAdminV1Index{Kind: s.Kind, Ancestor: strings.ToUpper(s.Ancestor)}
		if idx.Ancestor == "" {
			idx.Ancestor = "NONE"
		}
		for _, p := range s.Properties {
			idx.Properties = append(idx.Properties, &datastore.GoogleDatastoreAdminV1IndexedProperty{
				Name:      p.Name,
				Direction: strings.ToUpper(p.Direction),
			})
		}
		out = append(out, idx)
	}
	return out, nil
}

func dsIndexFromFlags() (*datastore.GoogleDatastoreAdminV1Index, error) {
	if flagDatastoreKind == "" {
		return nil, fmt.Errorf("--kind is required when --file is not provided")
	}
	idx := &datastore.GoogleDatastoreAdminV1Index{
		Kind:     flagDatastoreKind,
		Ancestor: strings.ToUpper(flagDatastoreAncestor),
	}
	for _, spec := range flagDatastoreProps {
		name, dir, ok := strings.Cut(spec, ":")
		if !ok {
			return nil, fmt.Errorf("invalid --property %q (expected NAME:DIRECTION)", spec)
		}
		idx.Properties = append(idx.Properties, &datastore.GoogleDatastoreAdminV1IndexedProperty{
			Name:      name,
			Direction: strings.ToUpper(dir),
		})
	}
	return idx, nil
}

// dsIndexSignature returns a stable key for comparing indexes across create/list.
func dsIndexSignature(idx *datastore.GoogleDatastoreAdminV1Index) string {
	parts := []string{idx.Kind, idx.Ancestor}
	for _, p := range idx.Properties {
		parts = append(parts, p.Name+":"+p.Direction)
	}
	return strings.Join(parts, "|")
}

// dsOperationName qualifies a bare id under projects/{project}/operations/{id}
// but leaves already-qualified names untouched.
func dsOperationName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/operations/%s", project, id)
}

// --- indexes impl ---

func runDSIndexesCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	var indexes []*datastore.GoogleDatastoreAdminV1Index
	if flagDatastoreFile != "" {
		indexes, err = dsIndexesFromFile(flagDatastoreFile)
		if err != nil {
			return err
		}
	} else {
		one, err := dsIndexFromFlags()
		if err != nil {
			return err
		}
		indexes = []*datastore.GoogleDatastoreAdminV1Index{one}
	}
	if len(indexes) == 0 {
		return fmt.Errorf("no indexes to create")
	}
	ctx := context.Background()
	svc, err := gcp.DatastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var ops []*datastore.GoogleLongrunningOperation
	for _, idx := range indexes {
		op, err := svc.Projects.Indexes.Create(project, idx).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating index (kind=%q): %w", idx.Kind, err)
		}
		fmt.Printf("Create request issued for index (kind=%q, operation: %s).\n", idx.Kind, op.Name)
		ops = append(ops, op)
	}
	return emitFormatted(ops, flagDatastoreFormat)
}

func runDSIndexesDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	idx, err := svc.Projects.Indexes.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing index: %w", err)
	}
	return emitFormatted(idx, flagDatastoreFormat)
}

func runDSIndexesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*datastore.GoogleDatastoreAdminV1Index
	pageToken := ""
	for {
		call := svc.Projects.Indexes.List(project).Context(ctx)
		if flagDatastoreFilter != "" {
			call = call.Filter(flagDatastoreFilter)
		}
		if flagDatastorePageSize > 0 {
			call = call.PageSize(flagDatastorePageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing indexes: %w", err)
		}
		all = append(all, resp.Indexes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDatastoreFormat)
}

func runDSIndexesCleanup(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	wanted, err := dsIndexesFromFile(flagDatastoreFile)
	if err != nil {
		return err
	}
	keep := map[string]struct{}{}
	for _, idx := range wanted {
		keep[dsIndexSignature(idx)] = struct{}{}
	}
	ctx := context.Background()
	svc, err := gcp.DatastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var existing []*datastore.GoogleDatastoreAdminV1Index
	pageToken := ""
	for {
		call := svc.Projects.Indexes.List(project).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing indexes: %w", err)
		}
		existing = append(existing, resp.Indexes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	var deleted int
	for _, idx := range existing {
		if _, ok := keep[dsIndexSignature(idx)]; ok {
			continue
		}
		op, err := svc.Projects.Indexes.Delete(project, idx.IndexId).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("deleting index %s: %w", idx.IndexId, err)
		}
		fmt.Printf("Delete request issued for index [%s] (operation: %s).\n", idx.IndexId, op.Name)
		deleted++
	}
	fmt.Printf("Cleanup complete: %d unused index(es) deleted.\n", deleted)
	return nil
}

// --- operations impl ---

func runDSOperationsCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Operations.Cancel(dsOperationName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runDSOperationsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Operations.Get(dsOperationName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagDatastoreFormat)
}

func runDSOperationsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("projects/%s", project)
	var all []*datastore.GoogleLongrunningOperation
	pageToken := ""
	for {
		call := svc.Projects.Operations.List(parent).Context(ctx)
		if flagDatastoreFilter != "" {
			call = call.Filter(flagDatastoreFilter)
		}
		if flagDatastorePageSize > 0 {
			call = call.PageSize(flagDatastorePageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDatastoreFormat)
}
