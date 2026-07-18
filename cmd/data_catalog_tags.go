package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	datacatalog "google.golang.org/api/datacatalog/v1"
)

// --- gcloud data-catalog tags (#1507) ---

var dcTagsCmd = &cobra.Command{Use: "tags", Short: "Manage Data Catalog tags"}

var (
	flagDCTagLocation   string
	flagDCTagEntryGroup string
	flagDCTagEntry      string
	flagDCTagFormat     string
	flagDCTagConfigFile string
	flagDCTagUpdateMask string
	flagDCTagPageSize   int64
)

var (
	dcTagCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a tag",
		Args: cobra.NoArgs, RunE: runDCTagCreate,
	}
	dcTagDeleteCmd = &cobra.Command{
		Use: "delete TAG", Short: "Delete a tag",
		Args: cobra.ExactArgs(1), RunE: runDCTagDelete,
	}
	dcTagListCmd = &cobra.Command{
		Use: "list", Short: "List tags on an entry or entry group",
		Args: cobra.NoArgs, RunE: runDCTagList,
	}
	dcTagUpdateCmd = &cobra.Command{
		Use: "update TAG", Short: "Update a tag",
		Args: cobra.ExactArgs(1), RunE: runDCTagUpdate,
	}
)

func init() {
	all := []*cobra.Command{dcTagCreateCmd, dcTagDeleteCmd, dcTagListCmd, dcTagUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagDCTagLocation, "location", "", "Location that owns the entry group (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagDCTagEntryGroup, "entry-group", "", "Entry group ID (required)")
		_ = c.MarkFlagRequired("entry-group")
		c.Flags().StringVar(&flagDCTagEntry, "entry", "",
			"Entry ID (omit to attach the tag to the entry group itself)")
		c.Flags().StringVar(&flagDCTagFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dcTagCreateCmd, dcTagUpdateCmd} {
		c.Flags().StringVar(&flagDCTagConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Tag body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dcTagUpdateCmd.Flags().StringVar(&flagDCTagUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	dcTagListCmd.Flags().Int64Var(&flagDCTagPageSize, "page-size", 0, "Maximum results per page")

	dcTagsCmd.AddCommand(all...)
	dataCatalogCmd.AddCommand(dcTagsCmd)
}

func dcTagParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	group := dcChild("entryGroups", flagDCTagEntryGroup, dcLocationParent(project, flagDCTagLocation))
	if flagDCTagEntry != "" {
		return dcChild("entries", flagDCTagEntry, group), nil
	}
	return group, nil
}

func dcTagName(id string) (string, error) {
	parent, err := dcTagParent()
	if err != nil {
		return "", err
	}
	return dcChild("tags", id, parent), nil
}

func runDCTagCreate(cmd *cobra.Command, args []string) error {
	parent, err := dcTagParent()
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1Tag{}
	if err := loadYAMLOrJSONInto(flagDCTagConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var (
		got *datacatalog.GoogleCloudDatacatalogV1Tag
	)
	if flagDCTagEntry != "" {
		got, err = svc.Projects.Locations.EntryGroups.Entries.Tags.Create(parent, body).Context(ctx).Do()
	} else {
		got, err = svc.Projects.Locations.EntryGroups.Tags.Create(parent, body).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("creating tag: %w", err)
	}
	fmt.Printf("Created tag [%s].\n", got.Name)
	return emitFormatted(got, flagDCTagFormat)
}

func runDCTagDelete(cmd *cobra.Command, args []string) error {
	name, err := dcTagName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if flagDCTagEntry != "" {
		_, err = svc.Projects.Locations.EntryGroups.Entries.Tags.Delete(name).Context(ctx).Do()
	} else {
		_, err = svc.Projects.Locations.EntryGroups.Tags.Delete(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}
	fmt.Printf("Deleted tag [%s].\n", args[0])
	return nil
}

func runDCTagList(cmd *cobra.Command, args []string) error {
	parent, err := dcTagParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*datacatalog.GoogleCloudDatacatalogV1Tag
	pageToken := ""
	for {
		var resp *datacatalog.GoogleCloudDatacatalogV1ListTagsResponse
		if flagDCTagEntry != "" {
			call := svc.Projects.Locations.EntryGroups.Entries.Tags.List(parent).Context(ctx)
			if flagDCTagPageSize > 0 {
				call = call.PageSize(flagDCTagPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err = call.Do()
		} else {
			call := svc.Projects.Locations.EntryGroups.Tags.List(parent).Context(ctx)
			if flagDCTagPageSize > 0 {
				call = call.PageSize(flagDCTagPageSize)
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
	return emitFormatted(all, flagDCTagFormat)
}

func runDCTagUpdate(cmd *cobra.Command, args []string) error {
	name, err := dcTagName(args[0])
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1Tag{}
	if err := loadYAMLOrJSONInto(flagDCTagConfigFile, body); err != nil {
		return err
	}
	mask := flagDCTagUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var got *datacatalog.GoogleCloudDatacatalogV1Tag
	if flagDCTagEntry != "" {
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
	fmt.Printf("Updated tag [%s].\n", args[0])
	return emitFormatted(got, flagDCTagFormat)
}
