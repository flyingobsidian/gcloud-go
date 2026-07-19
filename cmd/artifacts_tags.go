package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts tags (#1083) ---

var artifactsTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage Artifact Registry package tags",
}

var (
	flagArtTagLocation    string
	flagArtTagRepository  string
	flagArtTagPackage     string
	flagArtTagVersion     string
	flagArtTagFormat      string
	flagArtTagFilter      string
	flagArtTagDestination string
)

var artifactsTagsCreateCmd = &cobra.Command{
	Use:   "create TAG",
	Short: "Create a tag pointing to a version",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsTagsCreate,
}

var artifactsTagsDeleteCmd = &cobra.Command{
	Use:   "delete TAG",
	Short: "Delete a tag",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsTagsDelete,
}

var artifactsTagsExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export tags for a package to a JSON file",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsTagsExport,
}

var artifactsTagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tags for a package",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsTagsList,
}

var artifactsTagsUpdateCmd = &cobra.Command{
	Use:   "update TAG",
	Short: "Move a tag to a different version",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsTagsUpdate,
}

func init() {
	for _, c := range []*cobra.Command{
		artifactsTagsCreateCmd, artifactsTagsDeleteCmd, artifactsTagsExportCmd,
		artifactsTagsListCmd, artifactsTagsUpdateCmd,
	} {
		c.Flags().StringVar(&flagArtTagLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&flagArtTagRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&flagArtTagPackage, "package", "", "Package (required)")
		c.Flags().StringVar(&flagArtTagFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
		_ = c.MarkFlagRequired("package")
	}
	artifactsTagsCreateCmd.Flags().StringVar(&flagArtTagVersion, "version", "", "Target version (required)")
	_ = artifactsTagsCreateCmd.MarkFlagRequired("version")
	artifactsTagsUpdateCmd.Flags().StringVar(&flagArtTagVersion, "version", "", "New target version (required)")
	_ = artifactsTagsUpdateCmd.MarkFlagRequired("version")
	artifactsTagsListCmd.Flags().StringVar(&flagArtTagFilter, "filter", "", "Filter expression")
	artifactsTagsExportCmd.Flags().StringVar(&flagArtTagDestination, "destination", "", "Local JSON file to write tags to (required)")
	_ = artifactsTagsExportCmd.MarkFlagRequired("destination")

	artifactsTagsCmd.AddCommand(
		artifactsTagsCreateCmd, artifactsTagsDeleteCmd, artifactsTagsExportCmd,
		artifactsTagsListCmd, artifactsTagsUpdateCmd,
	)
	artifactsCmd.AddCommand(artifactsTagsCmd)
}

func artTagPackageParent(project string) string {
	return artFullName(artRepoParent(project, flagArtTagLocation, flagArtTagRepository)+"/packages", flagArtTagPackage)
}

func artTagVersionName(project string) string {
	pkg := artTagPackageParent(project)
	return artFullName(pkg+"/versions", flagArtTagVersion)
}

func artTagName(project, tagID string) string {
	pkg := artTagPackageParent(project)
	return artFullName(pkg+"/tags", tagID)
}

func runArtifactsTagsCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	tag := &artifactregistry.Tag{Version: artTagVersionName(project)}
	pkg := artTagPackageParent(project)
	out, err := svc.Projects.Locations.Repositories.Packages.Tags.Create(pkg, tag).
		TagId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating tag: %w", err)
	}
	return emitFormatted(out, flagArtTagFormat)
}

func runArtifactsTagsDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artTagName(project, args[0])
	if _, err := svc.Projects.Locations.Repositories.Packages.Tags.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Deleted tag [%s].\n", args[0])
	return nil
}

func runArtifactsTagsExport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	pkg := artTagPackageParent(project)
	var all []*artifactregistry.Tag
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.Packages.Tags.List(pkg).Context(ctx)
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing tags: %w", err)
		}
		all = append(all, resp.Tags...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	f, err := os.Create(flagArtTagDestination)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(all); err != nil {
		return fmt.Errorf("writing tags: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Exported %d tags to [%s].\n", len(all), flagArtTagDestination)
	return nil
}

func runArtifactsTagsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	pkg := artTagPackageParent(project)
	var all []*artifactregistry.Tag
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.Packages.Tags.List(pkg).Context(ctx)
		if flagArtTagFilter != "" {
			call = call.Filter(flagArtTagFilter)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing tags: %w", err)
		}
		all = append(all, resp.Tags...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagArtTagFormat)
}

func runArtifactsTagsUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artTagName(project, args[0])
	tag := &artifactregistry.Tag{Version: artTagVersionName(project)}
	out, err := svc.Projects.Locations.Repositories.Packages.Tags.Patch(name, tag).
		UpdateMask("version").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating tag: %w", err)
	}
	return emitFormatted(out, flagArtTagFormat)
}
