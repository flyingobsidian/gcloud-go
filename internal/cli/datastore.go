package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	datastore "google.golang.org/api/datastore/v1"
)

// --- gcloud datastore (#325, #969, #970) ---

// --- indexes ---

// --- operations ---

func (s *datastoreCmds) initDatastore() {
	// indexes flags
	s.datastoreIndexesCreateCmd.Flags().StringVar(&s.flagDatastoreKind, "kind", "", "Entity kind (required unless --file provides indexes)")
	s.datastoreIndexesCreateCmd.Flags().StringVar(&s.flagDatastoreAncestor, "ancestor", "NONE", "Ancestor mode: NONE or ALL_ANCESTORS")
	s.datastoreIndexesCreateCmd.Flags().StringSliceVar(&s.flagDatastoreProps, "property", nil, "Indexed property in NAME:DIRECTION form (may be repeated). DIRECTION is ASCENDING or DESCENDING")
	s.datastoreIndexesCreateCmd.Flags().StringVar(&s.flagDatastoreFile, "file", "", "Path to a YAML/JSON file containing one index or a list of indexes to create")
	s.datastoreIndexesCleanupCmd.Flags().StringVar(&s.flagDatastoreFile, "file", "", "Path to a YAML/JSON file describing the indexes that should be kept (required)")
	_ = s.datastoreIndexesCleanupCmd.MarkFlagRequired("file")
	s.datastoreIndexesListCmd.Flags().StringVar(&s.flagDatastoreFilter, "filter", "", "Server-side filter expression")
	s.datastoreIndexesListCmd.Flags().Int64Var(&s.flagDatastorePageSize, "page-size", 0, "Maximum number of results per page")
	for _, c := range []*cobra.Command{s.datastoreIndexesDescribeCmd, s.datastoreIndexesListCmd, s.datastoreIndexesCreateCmd} {
		c.Flags().StringVar(&s.flagDatastoreFormat, "format", "", "Output format")
	}
	s.datastoreIndexesCmd.AddCommand(s.datastoreIndexesCreateCmd, s.datastoreIndexesDescribeCmd, s.datastoreIndexesListCmd, s.datastoreIndexesCleanupCmd)

	// operations flags
	s.datastoreOperationsListCmd.Flags().StringVar(&s.flagDatastoreFilter, "filter", "", "Server-side filter expression")
	s.datastoreOperationsListCmd.Flags().Int64Var(&s.flagDatastorePageSize, "page-size", 0, "Maximum number of results per page")
	for _, c := range []*cobra.Command{s.datastoreOperationsDescribeCmd, s.datastoreOperationsListCmd} {
		c.Flags().StringVar(&s.flagDatastoreFormat, "format", "", "Output format")
	}
	s.datastoreOperationsCmd.AddCommand(s.datastoreOperationsCancelCmd, s.datastoreOperationsDescribeCmd, s.datastoreOperationsListCmd)

	s.datastoreCmd.AddCommand(s.datastoreIndexesCmd, s.datastoreOperationsCmd)
	for _, name := range []string{"export", "import"} {
		registerStubCommand(s.datastoreCmd, name, "Not yet implemented")
	}

}

// --- helpers ---

// dsIndexFile mirrors the shape of the YAML/JSON documents accepted by the
// gcloud CLI: either a single index or a `{ "indexes": [ ... ] }` list.
type dsIndexFile struct {
	Indexes []dsIndexSpec `json:"indexes,omitempty" yaml:"indexes,omitempty"`
	dsIndexSpec
}

type dsIndexSpec struct {
	Kind       string            `json:"kind,omitempty" yaml:"kind,omitempty"`
	Ancestor   string            `json:"ancestor,omitempty" yaml:"ancestor,omitempty"`
	Properties []dsIndexPropSpec `json:"properties,omitempty" yaml:"properties,omitempty"`
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

func (s *datastoreCmds) dsIndexFromFlags() (*datastore.GoogleDatastoreAdminV1Index, error) {
	if s.flagDatastoreKind == "" {
		return nil, fmt.Errorf("--kind is required when --file is not provided")
	}
	idx := &datastore.GoogleDatastoreAdminV1Index{
		Kind:     s.flagDatastoreKind,
		Ancestor: strings.ToUpper(s.flagDatastoreAncestor),
	}
	for _, spec := range s.flagDatastoreProps {
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

func (s *datastoreCmds) runDSIndexesCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	var indexes []*datastore.GoogleDatastoreAdminV1Index
	if s.flagDatastoreFile != "" {
		indexes, err = dsIndexesFromFile(s.flagDatastoreFile)
		if err != nil {
			return err
		}
	} else {
		one, err := s.dsIndexFromFlags()
		if err != nil {
			return err
		}
		indexes = []*datastore.GoogleDatastoreAdminV1Index{one}
	}
	if len(indexes) == 0 {
		return fmt.Errorf("no indexes to create")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Datastore(ctx)
	if err != nil {
		return err
	}
	var ops []*datastore.GoogleLongrunningOperation
	for _, idx := range indexes {
		op, err := svc.Projects.Indexes.Create(project, idx).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating index (kind=%q): %w", idx.Kind, err)
		}
		fmt.Fprintf(s.inv.stdout, "Create request issued for index (kind=%q, operation: %s).\n", idx.Kind, op.Name)
		ops = append(ops, op)
	}
	return emitFormattedTo(s.inv.stdout, ops, s.flagDatastoreFormat)
}

func (s *datastoreCmds) runDSIndexesDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Datastore(ctx)
	if err != nil {
		return err
	}
	idx, err := svc.Projects.Indexes.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing index: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, idx, s.flagDatastoreFormat)
}

func (s *datastoreCmds) runDSIndexesList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Datastore(ctx)
	if err != nil {
		return err
	}
	var all []*datastore.GoogleDatastoreAdminV1Index
	pageToken := ""
	for {
		call := svc.Projects.Indexes.List(project).Context(ctx)
		if s.flagDatastoreFilter != "" {
			call = call.Filter(s.flagDatastoreFilter)
		}
		if s.flagDatastorePageSize > 0 {
			call = call.PageSize(s.flagDatastorePageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagDatastoreFormat)
}

func (s *datastoreCmds) runDSIndexesCleanup(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	wanted, err := dsIndexesFromFile(s.flagDatastoreFile)
	if err != nil {
		return err
	}
	keep := map[string]struct{}{}
	for _, idx := range wanted {
		keep[dsIndexSignature(idx)] = struct{}{}
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Datastore(ctx)
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
		fmt.Fprintf(s.inv.stdout, "Delete request issued for index [%s] (operation: %s).\n", idx.IndexId, op.Name)
		deleted++
	}
	fmt.Fprintf(s.inv.stdout, "Cleanup complete: %d unused index(es) deleted.\n", deleted)
	return nil
}

// --- operations impl ---

func (s *datastoreCmds) runDSOperationsCancel(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Datastore(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Operations.Cancel(dsOperationName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *datastoreCmds) runDSOperationsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Datastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Operations.Get(dsOperationName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagDatastoreFormat)
}

func (s *datastoreCmds) runDSOperationsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Datastore(ctx)
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("projects/%s", project)
	var all []*datastore.GoogleLongrunningOperation
	pageToken := ""
	for {
		call := svc.Projects.Operations.List(parent).Context(ctx)
		if s.flagDatastoreFilter != "" {
			call = call.Filter(s.flagDatastoreFilter)
		}
		if s.flagDatastorePageSize > 0 {
			call = call.PageSize(s.flagDatastorePageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagDatastoreFormat)
}

// datastoreCmds holds the commands and flag values of one invocation's "datastore"
// command group, so invocations don't share state (#2070).
type datastoreCmds struct {
	inv *invocation

	datastoreCmd                   *cobra.Command
	flagDatastoreFile              string
	flagDatastoreKind              string
	flagDatastoreAncestor          string
	flagDatastoreProps             []string
	flagDatastoreFilter            string
	flagDatastorePageSize          int64
	flagDatastoreFormat            string
	datastoreIndexesCmd            *cobra.Command
	datastoreIndexesCreateCmd      *cobra.Command
	datastoreIndexesDescribeCmd    *cobra.Command
	datastoreIndexesListCmd        *cobra.Command
	datastoreIndexesCleanupCmd     *cobra.Command
	datastoreOperationsCmd         *cobra.Command
	datastoreOperationsCancelCmd   *cobra.Command
	datastoreOperationsDescribeCmd *cobra.Command
	datastoreOperationsListCmd     *cobra.Command
}

// newDatastoreCmd returns the "datastore" command group for inv.
func newDatastoreCmd(inv *invocation) *cobra.Command {
	return newDatastoreCmds(inv).datastoreCmd
}

// newDatastoreCmds builds the "datastore" command group for inv and returns its state.
func newDatastoreCmds(inv *invocation) *datastoreCmds {
	s := &datastoreCmds{inv: inv}
	s.datastoreCmd = &cobra.Command{Use: "datastore", Short: "Manage Cloud Datastore"}
	s.datastoreIndexesCmd = &cobra.Command{Use: "indexes", Short: "Manage Datastore indexes"}
	s.datastoreIndexesCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a Datastore index",
		Args:  cobra.NoArgs,
		RunE:  s.runDSIndexesCreate,
	}
	s.datastoreIndexesDescribeCmd = &cobra.Command{
		Use:   "describe INDEX_ID",
		Short: "Describe a Datastore index",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDSIndexesDescribe,
	}
	s.datastoreIndexesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Datastore indexes",
		Args:  cobra.NoArgs,
		RunE:  s.runDSIndexesList,
	}
	s.datastoreIndexesCleanupCmd = &cobra.Command{
		Use:   "cleanup",
		Short: "Delete unused Datastore indexes (indexes present in the project but not in the provided --file)",
		Args:  cobra.NoArgs,
		RunE:  s.runDSIndexesCleanup,
	}
	s.datastoreOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Datastore long-running operations"}
	s.datastoreOperationsCancelCmd = &cobra.Command{
		Use:   "cancel OPERATION",
		Short: "Cancel a Datastore operation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDSOperationsCancel,
	}
	s.datastoreOperationsDescribeCmd = &cobra.Command{
		Use:   "describe OPERATION",
		Short: "Describe a Datastore operation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDSOperationsDescribe,
	}
	s.datastoreOperationsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Datastore operations",
		Args:  cobra.NoArgs,
		RunE:  s.runDSOperationsList,
	}
	s.initDatastore()
	return s
}
