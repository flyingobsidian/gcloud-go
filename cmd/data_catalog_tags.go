package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	datacatalog "google.golang.org/api/datacatalog/v1"
)

// --- gcloud data-catalog tags (#1507) ---

func (s *dataCatalogCmds) initDataCatalogTags() {
	all := []*cobra.Command{s.dcTagCreateCmd, s.dcTagDeleteCmd, s.dcTagListCmd, s.dcTagUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDCTagLocation, "location", "", "Location that owns the entry group (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDCTagEntryGroup, "entry-group", "", "Entry group ID (required)")
		_ = c.MarkFlagRequired("entry-group")
		c.Flags().StringVar(&s.flagDCTagEntry, "entry", "",
			"Entry ID (omit to attach the tag to the entry group itself)")
		c.Flags().StringVar(&s.flagDCTagFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dcTagCreateCmd, s.dcTagUpdateCmd} {
		c.Flags().StringVar(&s.flagDCTagConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Tag body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dcTagUpdateCmd.Flags().StringVar(&s.flagDCTagUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.dcTagListCmd.Flags().Int64Var(&s.flagDCTagPageSize, "page-size", 0, "Maximum results per page")

	s.dcTagsCmd.AddCommand(all...)
	s.dataCatalogCmd.AddCommand(s.dcTagsCmd)
}

func (s *dataCatalogCmds) dcTagParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	group := dcChild("entryGroups", s.flagDCTagEntryGroup, dcLocationParent(project, s.flagDCTagLocation))
	if s.flagDCTagEntry != "" {
		return dcChild("entries", s.flagDCTagEntry, group), nil
	}
	return group, nil
}

func (s *dataCatalogCmds) dcTagName(id string) (string, error) {
	parent, err := s.dcTagParent()
	if err != nil {
		return "", err
	}
	return dcChild("tags", id, parent), nil
}

func (s *dataCatalogCmds) runDCTagCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.dcTagParent()
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1Tag{}
	if err := loadYAMLOrJSONInto(s.flagDCTagConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	var (
		got *datacatalog.GoogleCloudDatacatalogV1Tag
	)
	if s.flagDCTagEntry != "" {
		got, err = svc.Projects.Locations.EntryGroups.Entries.Tags.Create(parent, body).Context(ctx).Do()
	} else {
		got, err = svc.Projects.Locations.EntryGroups.Tags.Create(parent, body).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("creating tag: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created tag [%s].\n", got.Name)
	return emitFormattedTo(s.inv.stdout, got, s.flagDCTagFormat)
}

func (s *dataCatalogCmds) runDCTagDelete(cmd *cobra.Command, args []string) error {
	name, err := s.dcTagName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	if s.flagDCTagEntry != "" {
		_, err = svc.Projects.Locations.EntryGroups.Entries.Tags.Delete(name).Context(ctx).Do()
	} else {
		_, err = svc.Projects.Locations.EntryGroups.Tags.Delete(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted tag [%s].\n", args[0])
	return nil
}

func (s *dataCatalogCmds) runDCTagList(cmd *cobra.Command, args []string) error {
	parent, err := s.dcTagParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	var all []*datacatalog.GoogleCloudDatacatalogV1Tag
	pageToken := ""
	for {
		var resp *datacatalog.GoogleCloudDatacatalogV1ListTagsResponse
		if s.flagDCTagEntry != "" {
			call := svc.Projects.Locations.EntryGroups.Entries.Tags.List(parent).Context(ctx)
			if s.flagDCTagPageSize > 0 {
				call = call.PageSize(s.flagDCTagPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err = call.Do()
		} else {
			call := svc.Projects.Locations.EntryGroups.Tags.List(parent).Context(ctx)
			if s.flagDCTagPageSize > 0 {
				call = call.PageSize(s.flagDCTagPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err = call.Do()
		}
		if err != nil {
			return fmt.Errorf("listing tags: %w", err)
		}
		all = append(all, resp.Tags...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDCTagFormat)
}

func (s *dataCatalogCmds) runDCTagUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.dcTagName(args[0])
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1Tag{}
	if err := loadYAMLOrJSONInto(s.flagDCTagConfigFile, body); err != nil {
		return err
	}
	mask := s.flagDCTagUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	var got *datacatalog.GoogleCloudDatacatalogV1Tag
	if s.flagDCTagEntry != "" {
		call := svc.Projects.Locations.EntryGroups.Entries.Tags.Patch(name, body).Context(ctx)
		if mask != "" {
			call = call.UpdateMask(mask)
		}
		got, err = call.Do()
	} else {
		call := svc.Projects.Locations.EntryGroups.Tags.Patch(name, body).Context(ctx)
		if mask != "" {
			call = call.UpdateMask(mask)
		}
		got, err = call.Do()
	}
	if err != nil {
		return fmt.Errorf("updating tag: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated tag [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDCTagFormat)
}
