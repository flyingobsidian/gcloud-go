package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts tags (#1083) ---

func (s *artifactsCmds) initArtifactsTags() {
	for _, c := range []*cobra.Command{
		s.artifactsTagsCreateCmd, s.artifactsTagsDeleteCmd, s.artifactsTagsExportCmd,
		s.artifactsTagsListCmd, s.artifactsTagsUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagArtTagLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&s.flagArtTagRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&s.flagArtTagPackage, "package", "", "Package (required)")
		c.Flags().StringVar(&s.flagArtTagFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
		_ = c.MarkFlagRequired("package")
	}
	s.artifactsTagsCreateCmd.Flags().StringVar(&s.flagArtTagVersion, "version", "", "Target version (required)")
	_ = s.artifactsTagsCreateCmd.MarkFlagRequired("version")
	s.artifactsTagsUpdateCmd.Flags().StringVar(&s.flagArtTagVersion, "version", "", "New target version (required)")
	_ = s.artifactsTagsUpdateCmd.MarkFlagRequired("version")
	s.artifactsTagsListCmd.Flags().StringVar(&s.flagArtTagFilter, "filter", "", "Filter expression")
	s.artifactsTagsExportCmd.Flags().StringVar(&s.flagArtTagDestination, "destination", "", "Local JSON file to write tags to (required)")
	_ = s.artifactsTagsExportCmd.MarkFlagRequired("destination")

	s.artifactsTagsCmd.AddCommand(
		s.artifactsTagsCreateCmd, s.artifactsTagsDeleteCmd, s.artifactsTagsExportCmd,
		s.artifactsTagsListCmd, s.artifactsTagsUpdateCmd,
	)
	s.artifactsCmd.AddCommand(s.artifactsTagsCmd)
}

func (s *artifactsCmds) artTagPackageParent(project string) string {
	return artFullName(artRepoParent(project, s.flagArtTagLocation, s.flagArtTagRepository)+"/packages", s.flagArtTagPackage)
}

func (s *artifactsCmds) artTagVersionName(project string) string {
	pkg := s.artTagPackageParent(project)
	return artFullName(pkg+"/versions", s.flagArtTagVersion)
}

func (s *artifactsCmds) artTagName(project, tagID string) string {
	pkg := s.artTagPackageParent(project)
	return artFullName(pkg+"/tags", tagID)
}

func (s *artifactsCmds) runArtifactsTagsCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	tag := &artifactregistry.Tag{Version: s.artTagVersionName(project)}
	pkg := s.artTagPackageParent(project)
	out, err := svc.Projects.Locations.Repositories.Packages.Tags.Create(pkg, tag).
		TagId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating tag: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtTagFormat)
}

func (s *artifactsCmds) runArtifactsTagsDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := s.artTagName(project, args[0])
	if _, err := svc.Projects.Locations.Repositories.Packages.Tags.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Deleted tag [%s].\n", args[0])
	return nil
}

func (s *artifactsCmds) runArtifactsTagsExport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	pkg := s.artTagPackageParent(project)
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
	f, err := os.Create(s.flagArtTagDestination)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(all); err != nil {
		return fmt.Errorf("writing tags: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Exported %d tags to [%s].\n", len(all), s.flagArtTagDestination)
	return nil
}

func (s *artifactsCmds) runArtifactsTagsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	pkg := s.artTagPackageParent(project)
	var all []*artifactregistry.Tag
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.Packages.Tags.List(pkg).Context(ctx)
		if s.flagArtTagFilter != "" {
			call = call.Filter(s.flagArtTagFilter)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagArtTagFormat)
}

func (s *artifactsCmds) runArtifactsTagsUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := s.artTagName(project, args[0])
	tag := &artifactregistry.Tag{Version: s.artTagVersionName(project)}
	out, err := svc.Projects.Locations.Repositories.Packages.Tags.Patch(name, tag).
		UpdateMask("version").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating tag: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtTagFormat)
}
