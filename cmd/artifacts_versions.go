package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts versions (#1084) ---

func (s *artifactsCmds) initArtifactsVersions() {
	for _, c := range []*cobra.Command{
		s.artifactsVersionsDeleteCmd, s.artifactsVersionsDescribeCmd, s.artifactsVersionsExportCmd,
		s.artifactsVersionsListCmd, s.artifactsVersionsUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagArtVerLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&s.flagArtVerRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&s.flagArtVerPackage, "package", "", "Package (required)")
		c.Flags().StringVar(&s.flagArtVerFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
		_ = c.MarkFlagRequired("package")
	}
	s.artifactsVersionsDeleteCmd.Flags().BoolVar(&s.flagArtVerAsync, "async", false, "Return without waiting for operation")
	s.artifactsVersionsListCmd.Flags().StringVar(&s.flagArtVerFilter, "filter", "", "Filter expression")
	s.artifactsVersionsUpdateCmd.Flags().StringVar(&s.flagArtVerConfigFile, "config-file", "", "YAML/JSON body for the Version update (required)")
	s.artifactsVersionsUpdateCmd.Flags().StringVar(&s.flagArtVerMask, "update-mask", "", "Fields to update")
	_ = s.artifactsVersionsUpdateCmd.MarkFlagRequired("config-file")
	s.artifactsVersionsExportCmd.Flags().StringVar(&s.flagArtVerDestination, "destination", "", "Local JSON file to write versions to (required)")
	_ = s.artifactsVersionsExportCmd.MarkFlagRequired("destination")

	s.artifactsVersionsCmd.AddCommand(
		s.artifactsVersionsDeleteCmd, s.artifactsVersionsDescribeCmd, s.artifactsVersionsExportCmd,
		s.artifactsVersionsListCmd, s.artifactsVersionsUpdateCmd,
	)
	s.artifactsCmd.AddCommand(s.artifactsVersionsCmd)
}

func (s *artifactsCmds) artVerPackageParent(project string) string {
	return artFullName(artRepoParent(project, s.flagArtVerLocation, s.flagArtVerRepository)+"/packages", s.flagArtVerPackage)
}

func (s *artifactsCmds) artVerName(project, id string) string {
	return artFullName(s.artVerPackageParent(project)+"/versions", id)
}

func (s *artifactsCmds) runArtifactsVersionsDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := s.artVerName(project, args[0])
	op, err := svc.Projects.Locations.Repositories.Packages.Versions.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting version: %w", err)
	}
	if s.flagArtVerAsync {
		fmt.Fprintf(s.inv.stderr, "Delete started. Operation: %s\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagArtVerFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, final, s.flagArtVerFormat)
}

func (s *artifactsCmds) runArtifactsVersionsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := s.artVerName(project, args[0])
	v, err := svc.Projects.Locations.Repositories.Packages.Versions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing version: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, v, s.flagArtVerFormat)
}

func (s *artifactsCmds) runArtifactsVersionsExport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	all, err := s.listArtifactsVersions(project)
	if err != nil {
		return err
	}
	f, err := os.Create(s.flagArtVerDestination)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(all); err != nil {
		return fmt.Errorf("writing versions: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Exported %d versions to [%s].\n", len(all), s.flagArtVerDestination)
	return nil
}

func (s *artifactsCmds) runArtifactsVersionsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	all, err := s.listArtifactsVersions(project)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagArtVerFormat)
}

func (s *artifactsCmds) listArtifactsVersions(project string) ([]*artifactregistry.Version, error) {
	ctx := s.inv.context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return nil, err
	}
	pkg := s.artVerPackageParent(project)
	var all []*artifactregistry.Version
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.Packages.Versions.List(pkg).Context(ctx)
		if s.flagArtVerFilter != "" {
			call = call.Filter(s.flagArtVerFilter)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("listing versions: %w", err)
		}
		all = append(all, resp.Versions...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return all, nil
}

func (s *artifactsCmds) runArtifactsVersionsUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	body := &artifactregistry.Version{}
	if err := loadYAMLOrJSONInto(s.flagArtVerConfigFile, body); err != nil {
		return err
	}
	mask := s.flagArtVerMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := s.artVerName(project, args[0])
	call := svc.Projects.Locations.Repositories.Packages.Versions.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating version: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtVerFormat)
}
