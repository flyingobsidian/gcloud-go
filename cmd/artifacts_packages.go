package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts packages (#1077) ---

func (s *artifactsCmds) initArtifactsPackages() {
	for _, c := range []*cobra.Command{
		s.artifactsPackagesDeleteCmd, s.artifactsPackagesDescribeCmd,
		s.artifactsPackagesListCmd, s.artifactsPackagesUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagArtPkgLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&s.flagArtPkgRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&s.flagArtPkgFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	s.artifactsPackagesDeleteCmd.Flags().BoolVar(&s.flagArtPkgAsync, "async", false, "Return without waiting")
	s.artifactsPackagesListCmd.Flags().StringVar(&s.flagArtPkgFilter, "filter", "", "Filter expression")
	s.artifactsPackagesListCmd.Flags().Int64Var(&s.flagArtPkgPageSize, "page-size", 0, "Page size")
	s.artifactsPackagesUpdateCmd.Flags().StringVar(&s.flagArtPkgConfigFile, "config-file", "", "YAML/JSON body for the Package update (required)")
	s.artifactsPackagesUpdateCmd.Flags().StringVar(&s.flagArtPkgMask, "update-mask", "", "Fields to update")
	_ = s.artifactsPackagesUpdateCmd.MarkFlagRequired("config-file")

	s.artifactsPackagesCmd.AddCommand(
		s.artifactsPackagesDeleteCmd, s.artifactsPackagesDescribeCmd,
		s.artifactsPackagesListCmd, s.artifactsPackagesUpdateCmd,
	)
	s.artifactsCmd.AddCommand(s.artifactsPackagesCmd)
}

func artPackageName(project, location, repo, raw string) string {
	return artFullName(artRepoParent(project, location, repo)+"/packages", raw)
}

func (s *artifactsCmds) runArtifactsPackagesDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artPackageName(project, s.flagArtPkgLocation, s.flagArtPkgRepository, args[0])
	op, err := svc.Projects.Locations.Repositories.Packages.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting package: %w", err)
	}
	if s.flagArtPkgAsync {
		fmt.Fprintf(s.inv.stderr, "Delete started. Operation: %s\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagArtPkgFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, final, s.flagArtPkgFormat)
}

func (s *artifactsCmds) runArtifactsPackagesDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artPackageName(project, s.flagArtPkgLocation, s.flagArtPkgRepository, args[0])
	p, err := svc.Projects.Locations.Repositories.Packages.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing package: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, p, s.flagArtPkgFormat)
}

func (s *artifactsCmds) runArtifactsPackagesList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, s.flagArtPkgLocation, s.flagArtPkgRepository)
	var all []*artifactregistry.Package
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.Packages.List(parent).Context(ctx)
		if s.flagArtPkgFilter != "" {
			call = call.Filter(s.flagArtPkgFilter)
		}
		if s.flagArtPkgPageSize > 0 {
			call = call.PageSize(s.flagArtPkgPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing packages: %w", err)
		}
		all = append(all, resp.Packages...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagArtPkgFormat)
}

func (s *artifactsCmds) runArtifactsPackagesUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	body := &artifactregistry.Package{}
	if err := loadYAMLOrJSONInto(s.flagArtPkgConfigFile, body); err != nil {
		return err
	}
	mask := s.flagArtPkgMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := artPackageName(project, s.flagArtPkgLocation, s.flagArtPkgRepository, args[0])
	call := svc.Projects.Locations.Repositories.Packages.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating package: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtPkgFormat)
}
