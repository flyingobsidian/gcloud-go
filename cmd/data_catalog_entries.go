package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	datacatalog "google.golang.org/api/datacatalog/v1"
)

// --- gcloud data-catalog entries (#1504) ---

var dcEntriesCmd = &cobra.Command{Use: "entries", Short: "Manage Data Catalog entries"}

var (
	flagDCEntryLocation      string
	flagDCEntryGroup         string
	flagDCEntryFormat        string
	flagDCEntryConfigFile    string
	flagDCEntryUpdateMask    string
	flagDCEntryPageSize      int64
	flagDCEntryLookupSql     string
	flagDCEntryLookupLinked  string
	flagDCEntryLookupFQN     string
	flagDCEntryLookupProject string
	flagDCEntryLookupLoc     string
)

var (
	dcEntriesCreateCmd = &cobra.Command{
		Use: "create ENTRY", Short: "Create a Data Catalog entry",
		Args: cobra.ExactArgs(1), RunE: runDCEntryCreate,
	}
	dcEntriesDeleteCmd = &cobra.Command{
		Use: "delete ENTRY", Short: "Delete a Data Catalog entry",
		Args: cobra.ExactArgs(1), RunE: runDCEntryDelete,
	}
	dcEntriesDescribeCmd = &cobra.Command{
		Use: "describe ENTRY", Short: "Describe a Data Catalog entry",
		Args: cobra.ExactArgs(1), RunE: runDCEntryDescribe,
	}
	dcEntriesListCmd = &cobra.Command{
		Use: "list", Short: "List Data Catalog entries in an entry-group",
		Args: cobra.NoArgs, RunE: runDCEntryList,
	}
	dcEntriesUpdateCmd = &cobra.Command{
		Use: "update ENTRY", Short: "Update a Data Catalog entry",
		Args: cobra.ExactArgs(1), RunE: runDCEntryUpdate,
	}
	dcEntriesLookupCmd = &cobra.Command{
		Use: "lookup", Short: "Look up a Data Catalog entry by resource / SQL / fully-qualified name",
		Args: cobra.NoArgs, RunE: runDCEntryLookup,
	}
	dcEntriesStarCmd = &cobra.Command{
		Use: "star ENTRY", Short: "Star (bookmark) an entry",
		Args: cobra.ExactArgs(1), RunE: runDCEntryStar,
	}
	dcEntriesUnstarCmd = &cobra.Command{
		Use: "unstar ENTRY", Short: "Unstar an entry",
		Args: cobra.ExactArgs(1), RunE: runDCEntryUnstar,
	}
)

func init() {
	scoped := []*cobra.Command{
		dcEntriesCreateCmd, dcEntriesDeleteCmd, dcEntriesDescribeCmd,
		dcEntriesListCmd, dcEntriesUpdateCmd, dcEntriesStarCmd, dcEntriesUnstarCmd,
	}
	for _, c := range scoped {
		c.Flags().StringVar(&flagDCEntryLocation, "location", "", "Location that owns the entry group (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagDCEntryGroup, "entry-group", "", "Entry group ID (required)")
		_ = c.MarkFlagRequired("entry-group")
		c.Flags().StringVar(&flagDCEntryFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dcEntriesCreateCmd, dcEntriesUpdateCmd} {
		c.Flags().StringVar(&flagDCEntryConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Entry body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dcEntriesUpdateCmd.Flags().StringVar(&flagDCEntryUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	dcEntriesListCmd.Flags().Int64Var(&flagDCEntryPageSize, "page-size", 0, "Maximum results per page")

	dcEntriesLookupCmd.Flags().StringVar(&flagDCEntryLookupSql, "sql-resource", "", "SQL resource path")
	dcEntriesLookupCmd.Flags().StringVar(&flagDCEntryLookupLinked, "linked-resource", "", "Linked resource name")
	dcEntriesLookupCmd.Flags().StringVar(&flagDCEntryLookupFQN, "fully-qualified-name", "", "Fully-qualified name")
	dcEntriesLookupCmd.Flags().StringVar(&flagDCEntryLookupProject, "project-id", "",
		"Project ID for the resource (for fully-qualified-name lookups)")
	dcEntriesLookupCmd.Flags().StringVar(&flagDCEntryLookupLoc, "location-id", "",
		"Location ID for the resource (for fully-qualified-name lookups)")
	dcEntriesLookupCmd.Flags().StringVar(&flagDCEntryFormat, "format", "", "Output format")

	dcEntriesCmd.AddCommand(append(scoped, dcEntriesLookupCmd)...)
	dataCatalogCmd.AddCommand(dcEntriesCmd)
}

func dcEntryGroupParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return dcChild("entryGroups", flagDCEntryGroup, dcLocationParent(project, flagDCEntryLocation)), nil
}

func dcEntryName(id string) (string, error) {
	group, err := dcEntryGroupParent()
	if err != nil {
		return "", err
	}
	return dcChild("entries", id, group), nil
}

func runDCEntryCreate(cmd *cobra.Command, args []string) error {
	parent, err := dcEntryGroupParent()
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1Entry{}
	if err := loadYAMLOrJSONInto(flagDCEntryConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.EntryGroups.Entries.Create(parent, body).EntryId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating entry: %w", err)
	}
	fmt.Printf("Created entry [%s].\n", args[0])
	return emitFormatted(got, flagDCEntryFormat)
}

func runDCEntryDelete(cmd *cobra.Command, args []string) error {
	name, err := dcEntryName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.EntryGroups.Entries.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting entry: %w", err)
	}
	fmt.Printf("Deleted entry [%s].\n", args[0])
	return nil
}

func runDCEntryDescribe(cmd *cobra.Command, args []string) error {
	name, err := dcEntryName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.EntryGroups.Entries.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing entry: %w", err)
	}
	return emitFormatted(got, flagDCEntryFormat)
}

func runDCEntryList(cmd *cobra.Command, args []string) error {
	parent, err := dcEntryGroupParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*datacatalog.GoogleCloudDatacatalogV1Entry
	pageToken := ""
	for {
		call := svc.Projects.Locations.EntryGroups.Entries.List(parent).Context(ctx)
		if flagDCEntryPageSize > 0 {
			call = call.PageSize(flagDCEntryPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing entries: %w", err)
		}
		all = append(all, resp.Entries...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDCEntryFormat)
}

func runDCEntryUpdate(cmd *cobra.Command, args []string) error {
	name, err := dcEntryName(args[0])
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1Entry{}
	if err := loadYAMLOrJSONInto(flagDCEntryConfigFile, body); err != nil {
		return err
	}
	mask := flagDCEntryUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.EntryGroups.Entries.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating entry: %w", err)
	}
	fmt.Printf("Updated entry [%s].\n", args[0])
	return emitFormatted(got, flagDCEntryFormat)
}

func runDCEntryLookup(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Entries.Lookup().Context(ctx)
	if flagDCEntryLookupSql != "" {
		call = call.SqlResource(flagDCEntryLookupSql)
	}
	if flagDCEntryLookupLinked != "" {
		call = call.LinkedResource(flagDCEntryLookupLinked)
	}
	if flagDCEntryLookupFQN != "" {
		call = call.FullyQualifiedName(flagDCEntryLookupFQN)
	}
	if flagDCEntryLookupProject != "" {
		call = call.Project(flagDCEntryLookupProject)
	}
	if flagDCEntryLookupLoc != "" {
		call = call.Location(flagDCEntryLookupLoc)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("looking up entry: %w", err)
	}
	return emitFormatted(got, flagDCEntryFormat)
}

func runDCEntryStar(cmd *cobra.Command, args []string) error {
	name, err := dcEntryName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.EntryGroups.Entries.Star(name, &datacatalog.GoogleCloudDatacatalogV1StarEntryRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("starring entry: %w", err)
	}
	fmt.Printf("Starred entry [%s].\n", args[0])
	return nil
}

func runDCEntryUnstar(cmd *cobra.Command, args []string) error {
	name, err := dcEntryName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.EntryGroups.Entries.Unstar(name, &datacatalog.GoogleCloudDatacatalogV1UnstarEntryRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unstarring entry: %w", err)
	}
	fmt.Printf("Unstarred entry [%s].\n", args[0])
	return nil
}
