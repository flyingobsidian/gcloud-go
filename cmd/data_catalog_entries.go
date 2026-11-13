package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	datacatalog "google.golang.org/api/datacatalog/v1"
)

// --- gcloud data-catalog entries (#1504) ---

func (s *dataCatalogCmds) initDataCatalogEntries() {
	scoped := []*cobra.Command{
		s.dcEntriesCreateCmd, s.dcEntriesDeleteCmd, s.dcEntriesDescribeCmd,
		s.dcEntriesListCmd, s.dcEntriesUpdateCmd, s.dcEntriesStarCmd, s.dcEntriesUnstarCmd,
	}
	for _, c := range scoped {
		c.Flags().StringVar(&s.flagDCEntryLocation, "location", "", "Location that owns the entry group (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDCEntryGroup, "entry-group", "", "Entry group ID (required)")
		_ = c.MarkFlagRequired("entry-group")
		c.Flags().StringVar(&s.flagDCEntryFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dcEntriesCreateCmd, s.dcEntriesUpdateCmd} {
		c.Flags().StringVar(&s.flagDCEntryConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Entry body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dcEntriesUpdateCmd.Flags().StringVar(&s.flagDCEntryUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.dcEntriesListCmd.Flags().Int64Var(&s.flagDCEntryPageSize, "page-size", 0, "Maximum results per page")

	s.dcEntriesLookupCmd.Flags().StringVar(&s.flagDCEntryLookupSql, "sql-resource", "", "SQL resource path")
	s.dcEntriesLookupCmd.Flags().StringVar(&s.flagDCEntryLookupLinked, "linked-resource", "", "Linked resource name")
	s.dcEntriesLookupCmd.Flags().StringVar(&s.flagDCEntryLookupFQN, "fully-qualified-name", "", "Fully-qualified name")
	s.dcEntriesLookupCmd.Flags().StringVar(&s.flagDCEntryLookupProject, "project-id", "",
		"Project ID for the resource (for fully-qualified-name lookups)")
	s.dcEntriesLookupCmd.Flags().StringVar(&s.flagDCEntryLookupLoc, "location-id", "",
		"Location ID for the resource (for fully-qualified-name lookups)")
	s.dcEntriesLookupCmd.Flags().StringVar(&s.flagDCEntryFormat, "format", "", "Output format")

	s.dcEntriesCmd.AddCommand(append(scoped, s.dcEntriesLookupCmd)...)
	s.dataCatalogCmd.AddCommand(s.dcEntriesCmd)
}

func (s *dataCatalogCmds) dcEntryGroupParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return dcChild("entryGroups", s.flagDCEntryGroup, dcLocationParent(project, s.flagDCEntryLocation)), nil
}

func (s *dataCatalogCmds) dcEntryName(id string) (string, error) {
	group, err := s.dcEntryGroupParent()
	if err != nil {
		return "", err
	}
	return dcChild("entries", id, group), nil
}

func (s *dataCatalogCmds) runDCEntryCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.dcEntryGroupParent()
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1Entry{}
	if err := loadYAMLOrJSONInto(s.flagDCEntryConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.EntryGroups.Entries.Create(parent, body).EntryId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating entry: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created entry [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDCEntryFormat)
}

func (s *dataCatalogCmds) runDCEntryDelete(cmd *cobra.Command, args []string) error {
	name, err := s.dcEntryName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.EntryGroups.Entries.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting entry: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted entry [%s].\n", args[0])
	return nil
}

func (s *dataCatalogCmds) runDCEntryDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.dcEntryName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.EntryGroups.Entries.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing entry: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDCEntryFormat)
}

func (s *dataCatalogCmds) runDCEntryList(cmd *cobra.Command, args []string) error {
	parent, err := s.dcEntryGroupParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	var all []*datacatalog.GoogleCloudDatacatalogV1Entry
	pageToken := ""
	for {
		call := svc.Projects.Locations.EntryGroups.Entries.List(parent).Context(ctx)
		if s.flagDCEntryPageSize > 0 {
			call = call.PageSize(s.flagDCEntryPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagDCEntryFormat)
}

func (s *dataCatalogCmds) runDCEntryUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.dcEntryName(args[0])
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1Entry{}
	if err := loadYAMLOrJSONInto(s.flagDCEntryConfigFile, body); err != nil {
		return err
	}
	mask := s.flagDCEntryUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Updated entry [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDCEntryFormat)
}

func (s *dataCatalogCmds) runDCEntryLookup(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	call := svc.Entries.Lookup().Context(ctx)
	if s.flagDCEntryLookupSql != "" {
		call = call.SqlResource(s.flagDCEntryLookupSql)
	}
	if s.flagDCEntryLookupLinked != "" {
		call = call.LinkedResource(s.flagDCEntryLookupLinked)
	}
	if s.flagDCEntryLookupFQN != "" {
		call = call.FullyQualifiedName(s.flagDCEntryLookupFQN)
	}
	if s.flagDCEntryLookupProject != "" {
		call = call.Project(s.flagDCEntryLookupProject)
	}
	if s.flagDCEntryLookupLoc != "" {
		call = call.Location(s.flagDCEntryLookupLoc)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("looking up entry: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDCEntryFormat)
}

func (s *dataCatalogCmds) runDCEntryStar(cmd *cobra.Command, args []string) error {
	name, err := s.dcEntryName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.EntryGroups.Entries.Star(name, &datacatalog.GoogleCloudDatacatalogV1StarEntryRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("starring entry: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Starred entry [%s].\n", args[0])
	return nil
}

func (s *dataCatalogCmds) runDCEntryUnstar(cmd *cobra.Command, args []string) error {
	name, err := s.dcEntryName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.EntryGroups.Entries.Unstar(name, &datacatalog.GoogleCloudDatacatalogV1UnstarEntryRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unstarring entry: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Unstarred entry [%s].\n", args[0])
	return nil
}
