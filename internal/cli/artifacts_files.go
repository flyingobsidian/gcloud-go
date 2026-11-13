package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts files (#1071) ---

func (s *artifactsCmds) initArtifactsFiles() {
	for _, c := range []*cobra.Command{
		s.artifactsFilesDeleteCmd, s.artifactsFilesDescribeCmd, s.artifactsFilesDownloadCmd,
		s.artifactsFilesListCmd, s.artifactsFilesUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagArtFilesLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&s.flagArtFilesRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&s.flagArtFilesFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	s.artifactsFilesDeleteCmd.Flags().BoolVar(&s.flagArtFilesAsync, "async", false, "Return without waiting for operation")
	s.artifactsFilesListCmd.Flags().StringVar(&s.flagArtFilesFilter, "filter", "", "Filter expression")
	s.artifactsFilesListCmd.Flags().Int64Var(&s.flagArtFilesPageSize, "page-size", 0, "Page size")
	s.artifactsFilesDownloadCmd.Flags().StringVar(&s.flagArtFilesDestination, "destination", "", "Local file to write the download to (required)")
	_ = s.artifactsFilesDownloadCmd.MarkFlagRequired("destination")
	s.artifactsFilesUpdateCmd.Flags().StringVar(&s.flagArtFilesConfigFile, "config-file", "", "YAML/JSON body for the File update")
	s.artifactsFilesUpdateCmd.Flags().StringVar(&s.flagArtFilesMask, "update-mask", "", "Fields to update")
	_ = s.artifactsFilesUpdateCmd.MarkFlagRequired("config-file")

	s.artifactsFilesCmd.AddCommand(
		s.artifactsFilesDeleteCmd, s.artifactsFilesDescribeCmd, s.artifactsFilesDownloadCmd,
		s.artifactsFilesListCmd, s.artifactsFilesUpdateCmd,
	)
	s.artifactsCmd.AddCommand(s.artifactsFilesCmd)
}

func artFileName(project, location, repo, raw string) string {
	return artFullName(artRepoParent(project, location, repo)+"/files", raw)
}

func (s *artifactsCmds) runArtifactsFilesDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artFileName(project, s.flagArtFilesLocation, s.flagArtFilesRepository, args[0])
	op, err := svc.Projects.Locations.Repositories.Files.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting file: %w", err)
	}
	if s.flagArtFilesAsync {
		fmt.Fprintf(s.inv.stderr, "Delete started. Operation: %s\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagArtFilesFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, final, s.flagArtFilesFormat)
}

func (s *artifactsCmds) runArtifactsFilesDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artFileName(project, s.flagArtFilesLocation, s.flagArtFilesRepository, args[0])
	f, err := svc.Projects.Locations.Repositories.Files.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing file: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, f, s.flagArtFilesFormat)
}

func (s *artifactsCmds) runArtifactsFilesDownload(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artFileName(project, s.flagArtFilesLocation, s.flagArtFilesRepository, args[0])
	resp, err := svc.Projects.Locations.Repositories.Files.Download(name).Context(ctx).Download()
	if err != nil {
		return fmt.Errorf("downloading file: %w", err)
	}
	defer resp.Body.Close()
	out, err := os.Create(s.flagArtFilesDestination)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Downloaded [%s] to [%s].\n", name, s.flagArtFilesDestination)
	return nil
}

func (s *artifactsCmds) runArtifactsFilesList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, s.flagArtFilesLocation, s.flagArtFilesRepository)
	var all []*artifactregistry.GoogleDevtoolsArtifactregistryV1File
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.Files.List(parent).Context(ctx)
		if s.flagArtFilesFilter != "" {
			call = call.Filter(s.flagArtFilesFilter)
		}
		if s.flagArtFilesPageSize > 0 {
			call = call.PageSize(s.flagArtFilesPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing files: %w", err)
		}
		all = append(all, resp.Files...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagArtFilesFormat)
}

func (s *artifactsCmds) runArtifactsFilesUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	body := &artifactregistry.GoogleDevtoolsArtifactregistryV1File{}
	if err := loadYAMLOrJSONInto(s.flagArtFilesConfigFile, body); err != nil {
		return err
	}
	mask := s.flagArtFilesMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := artFileName(project, s.flagArtFilesLocation, s.flagArtFilesRepository, args[0])
	call := svc.Projects.Locations.Repositories.Files.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating file: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtFilesFormat)
}
