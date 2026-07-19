package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts attachments (#1070) ---

var artifactsAttachmentsCmd = &cobra.Command{
	Use:   "attachments",
	Short: "Manage Artifact Registry attachments",
}

var (
	flagArtAttLocation    string
	flagArtAttRepository  string
	flagArtAttFormat      string
	flagArtAttFilter      string
	flagArtAttConfigFile  string
	flagArtAttTarget      string
	flagArtAttType        string
	flagArtAttDestination string
	flagArtAttAsync       bool
)

var artifactsAttachmentsCreateCmd = &cobra.Command{
	Use:   "create ATTACHMENT",
	Short: "Create an attachment",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsAttachmentsCreate,
}

var artifactsAttachmentsDeleteCmd = &cobra.Command{
	Use:   "delete ATTACHMENT",
	Short: "Delete an attachment",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsAttachmentsDelete,
}

var artifactsAttachmentsDescribeCmd = &cobra.Command{
	Use:   "describe ATTACHMENT",
	Short: "Describe an attachment",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsAttachmentsDescribe,
}

var artifactsAttachmentsDownloadCmd = &cobra.Command{
	Use:   "download ATTACHMENT",
	Short: "Download all files referenced by an attachment",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsAttachmentsDownload,
}

var artifactsAttachmentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List attachments in a repository",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsAttachmentsList,
}

func init() {
	for _, c := range []*cobra.Command{
		artifactsAttachmentsCreateCmd, artifactsAttachmentsDeleteCmd,
		artifactsAttachmentsDescribeCmd, artifactsAttachmentsDownloadCmd,
		artifactsAttachmentsListCmd,
	} {
		c.Flags().StringVar(&flagArtAttLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&flagArtAttRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&flagArtAttFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	artifactsAttachmentsCreateCmd.Flags().StringVar(&flagArtAttConfigFile, "config-file", "", "YAML/JSON body for the Attachment (optional)")
	artifactsAttachmentsCreateCmd.Flags().StringVar(&flagArtAttTarget, "target", "", "Attachment target resource name")
	artifactsAttachmentsCreateCmd.Flags().StringVar(&flagArtAttType, "type", "", "MIME type of the attachment (e.g. application/vnd.spdx+json)")
	artifactsAttachmentsCreateCmd.Flags().BoolVar(&flagArtAttAsync, "async", false, "Return without waiting for operation")
	artifactsAttachmentsDeleteCmd.Flags().BoolVar(&flagArtAttAsync, "async", false, "Return without waiting for operation")
	artifactsAttachmentsListCmd.Flags().StringVar(&flagArtAttFilter, "filter", "", "Filter expression")
	artifactsAttachmentsDownloadCmd.Flags().StringVar(&flagArtAttDestination, "destination", "", "Local directory to write files to (required)")
	_ = artifactsAttachmentsDownloadCmd.MarkFlagRequired("destination")

	artifactsAttachmentsCmd.AddCommand(
		artifactsAttachmentsCreateCmd, artifactsAttachmentsDeleteCmd,
		artifactsAttachmentsDescribeCmd, artifactsAttachmentsDownloadCmd,
		artifactsAttachmentsListCmd,
	)
	artifactsCmd.AddCommand(artifactsAttachmentsCmd)
}

func artAttachmentName(project, location, repo, raw string) string {
	return artFullName(artRepoParent(project, location, repo)+"/attachments", raw)
}

func runArtifactsAttachmentsCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &artifactregistry.Attachment{}
	if flagArtAttConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagArtAttConfigFile, body); err != nil {
			return err
		}
	}
	if flagArtAttTarget != "" {
		body.Target = flagArtAttTarget
	}
	if flagArtAttType != "" {
		body.Type = flagArtAttType
	}
	parent := artRepoParent(project, flagArtAttLocation, flagArtAttRepository)
	op, err := svc.Projects.Locations.Repositories.Attachments.Create(parent, body).
		AttachmentId(args[0]).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating attachment: %w", err)
	}
	if flagArtAttAsync {
		fmt.Fprintf(os.Stderr, "Create started. Operation: %s\n", op.Name)
		return emitFormatted(op, flagArtAttFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagArtAttFormat)
}

func runArtifactsAttachmentsDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artAttachmentName(project, flagArtAttLocation, flagArtAttRepository, args[0])
	op, err := svc.Projects.Locations.Repositories.Attachments.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting attachment: %w", err)
	}
	if flagArtAttAsync {
		fmt.Fprintf(os.Stderr, "Delete started. Operation: %s\n", op.Name)
		return emitFormatted(op, flagArtAttFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagArtAttFormat)
}

func runArtifactsAttachmentsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artAttachmentName(project, flagArtAttLocation, flagArtAttRepository, args[0])
	a, err := svc.Projects.Locations.Repositories.Attachments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing attachment: %w", err)
	}
	return emitFormatted(a, flagArtAttFormat)
}

func runArtifactsAttachmentsDownload(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artAttachmentName(project, flagArtAttLocation, flagArtAttRepository, args[0])
	a, err := svc.Projects.Locations.Repositories.Attachments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing attachment: %w", err)
	}
	if len(a.Files) == 0 {
		return fmt.Errorf("attachment %s references no files; nothing to download", args[0])
	}
	if err := os.MkdirAll(flagArtAttDestination, 0o755); err != nil {
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
		out, err := os.Create(filepath.Join(flagArtAttDestination, base))
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
		fmt.Fprintf(os.Stderr, "Downloaded [%s] to [%s].\n", fileName, filepath.Join(flagArtAttDestination, base))
	}
	return nil
}

func runArtifactsAttachmentsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, flagArtAttLocation, flagArtAttRepository)
	var all []*artifactregistry.Attachment
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.Attachments.List(parent).Context(ctx)
		if flagArtAttFilter != "" {
			call = call.Filter(flagArtAttFilter)
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
	return emitFormatted(all, flagArtAttFormat)
}
