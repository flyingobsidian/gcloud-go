package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts attachments (#1070) ---

func (s *artifactsCmds) initArtifactsAttachments() {
	for _, c := range []*cobra.Command{
		s.artifactsAttachmentsCreateCmd, s.artifactsAttachmentsDeleteCmd,
		s.artifactsAttachmentsDescribeCmd, s.artifactsAttachmentsDownloadCmd,
		s.artifactsAttachmentsListCmd,
	} {
		c.Flags().StringVar(&s.flagArtAttLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&s.flagArtAttRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&s.flagArtAttFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	s.artifactsAttachmentsCreateCmd.Flags().StringVar(&s.flagArtAttConfigFile, "config-file", "", "YAML/JSON body for the Attachment (optional)")
	s.artifactsAttachmentsCreateCmd.Flags().StringVar(&s.flagArtAttTarget, "target", "", "Attachment target resource name")
	s.artifactsAttachmentsCreateCmd.Flags().StringVar(&s.flagArtAttType, "type", "", "MIME type of the attachment (e.g. application/vnd.spdx+json)")
	s.artifactsAttachmentsCreateCmd.Flags().BoolVar(&s.flagArtAttAsync, "async", false, "Return without waiting for operation")
	s.artifactsAttachmentsDeleteCmd.Flags().BoolVar(&s.flagArtAttAsync, "async", false, "Return without waiting for operation")
	s.artifactsAttachmentsListCmd.Flags().StringVar(&s.flagArtAttFilter, "filter", "", "Filter expression")
	s.artifactsAttachmentsDownloadCmd.Flags().StringVar(&s.flagArtAttDestination, "destination", "", "Local directory to write files to (required)")
	_ = s.artifactsAttachmentsDownloadCmd.MarkFlagRequired("destination")

	s.artifactsAttachmentsCmd.AddCommand(
		s.artifactsAttachmentsCreateCmd, s.artifactsAttachmentsDeleteCmd,
		s.artifactsAttachmentsDescribeCmd, s.artifactsAttachmentsDownloadCmd,
		s.artifactsAttachmentsListCmd,
	)
	s.artifactsCmd.AddCommand(s.artifactsAttachmentsCmd)
}

func artAttachmentName(project, location, repo, raw string) string {
	return artFullName(artRepoParent(project, location, repo)+"/attachments", raw)
}

func (s *artifactsCmds) runArtifactsAttachmentsCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	body := &artifactregistry.Attachment{}
	if s.flagArtAttConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagArtAttConfigFile, body); err != nil {
			return err
		}
	}
	if s.flagArtAttTarget != "" {
		body.Target = s.flagArtAttTarget
	}
	if s.flagArtAttType != "" {
		body.Type = s.flagArtAttType
	}
	parent := artRepoParent(project, s.flagArtAttLocation, s.flagArtAttRepository)
	op, err := svc.Projects.Locations.Repositories.Attachments.Create(parent, body).
		AttachmentId(args[0]).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating attachment: %w", err)
	}
	if s.flagArtAttAsync {
		fmt.Fprintf(s.inv.stderr, "Create started. Operation: %s\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagArtAttFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, final, s.flagArtAttFormat)
}

func (s *artifactsCmds) runArtifactsAttachmentsDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artAttachmentName(project, s.flagArtAttLocation, s.flagArtAttRepository, args[0])
	op, err := svc.Projects.Locations.Repositories.Attachments.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting attachment: %w", err)
	}
	if s.flagArtAttAsync {
		fmt.Fprintf(s.inv.stderr, "Delete started. Operation: %s\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagArtAttFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, final, s.flagArtAttFormat)
}

func (s *artifactsCmds) runArtifactsAttachmentsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artAttachmentName(project, s.flagArtAttLocation, s.flagArtAttRepository, args[0])
	a, err := svc.Projects.Locations.Repositories.Attachments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing attachment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, a, s.flagArtAttFormat)
}

func (s *artifactsCmds) runArtifactsAttachmentsDownload(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artAttachmentName(project, s.flagArtAttLocation, s.flagArtAttRepository, args[0])
	a, err := svc.Projects.Locations.Repositories.Attachments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing attachment: %w", err)
	}
	if len(a.Files) == 0 {
		return fmt.Errorf("attachment %s references no files; nothing to download", args[0])
	}
	if err := os.MkdirAll(s.flagArtAttDestination, 0o755); err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	for _, fileName := range a.Files {
		resp, err := svc.Projects.Locations.Repositories.Files.Download(fileName).Context(ctx).Download()
		if err != nil {
			return fmt.Errorf("downloading %s: %w", fileName, err)
		}
		base := filepath.Base(fileName)
		if base == "" || base == "." || base == "/" {
			base = strings.ReplaceAll(fileName, "/", "_")
		}
		out, err := os.Create(filepath.Join(s.flagArtAttDestination, base))
		if err != nil {
			resp.Body.Close()
			return fmt.Errorf("creating file: %w", err)
		}
		_, copyErr := io.Copy(out, resp.Body)
		out.Close()
		resp.Body.Close()
		if copyErr != nil {
			return fmt.Errorf("writing %s: %w", base, copyErr)
		}
		fmt.Fprintf(s.inv.stderr, "Downloaded [%s] to [%s].\n", fileName, filepath.Join(s.flagArtAttDestination, base))
	}
	return nil
}

func (s *artifactsCmds) runArtifactsAttachmentsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, s.flagArtAttLocation, s.flagArtAttRepository)
	var all []*artifactregistry.Attachment
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.Attachments.List(parent).Context(ctx)
		if s.flagArtAttFilter != "" {
			call = call.Filter(s.flagArtAttFilter)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing attachments: %w", err)
		}
		all = append(all, resp.Attachments...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagArtAttFormat)
}
