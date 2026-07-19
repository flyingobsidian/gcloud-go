package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts files (#1071) ---

var artifactsFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Manage individual files in an Artifact Registry repository",
}

var (
	flagArtFilesLocation    string
	flagArtFilesRepository  string
	flagArtFilesFormat      string
	flagArtFilesFilter      string
	flagArtFilesPageSize    int64
	flagArtFilesDestination string
	flagArtFilesConfigFile  string
	flagArtFilesMask        string
	flagArtFilesAsync       bool
)

var artifactsFilesDeleteCmd = &cobra.Command{
	Use:   "delete FILE",
	Short: "Delete a file from a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsFilesDelete,
}

var artifactsFilesDescribeCmd = &cobra.Command{
	Use:   "describe FILE",
	Short: "Describe a file in a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsFilesDescribe,
}

var artifactsFilesDownloadCmd = &cobra.Command{
	Use:   "download FILE",
	Short: "Download a file from a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsFilesDownload,
}

var artifactsFilesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List files in a repository",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsFilesList,
}

var artifactsFilesUpdateCmd = &cobra.Command{
	Use:   "update FILE",
	Short: "Update a file in a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsFilesUpdate,
}

func init() {
	for _, c := range []*cobra.Command{
		artifactsFilesDeleteCmd, artifactsFilesDescribeCmd, artifactsFilesDownloadCmd,
		artifactsFilesListCmd, artifactsFilesUpdateCmd,
	} {
		c.Flags().StringVar(&flagArtFilesLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&flagArtFilesRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&flagArtFilesFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	artifactsFilesDeleteCmd.Flags().BoolVar(&flagArtFilesAsync, "async", false, "Return without waiting for operation")
	artifactsFilesListCmd.Flags().StringVar(&flagArtFilesFilter, "filter", "", "Filter expression")
	artifactsFilesListCmd.Flags().Int64Var(&flagArtFilesPageSize, "page-size", 0, "Page size")
	artifactsFilesDownloadCmd.Flags().StringVar(&flagArtFilesDestination, "destination", "", "Local file to write the download to (required)")
	_ = artifactsFilesDownloadCmd.MarkFlagRequired("destination")
	artifactsFilesUpdateCmd.Flags().StringVar(&flagArtFilesConfigFile, "config-file", "", "YAML/JSON body for the File update")
	artifactsFilesUpdateCmd.Flags().StringVar(&flagArtFilesMask, "update-mask", "", "Fields to update")
	_ = artifactsFilesUpdateCmd.MarkFlagRequired("config-file")

	artifactsFilesCmd.AddCommand(
		artifactsFilesDeleteCmd, artifactsFilesDescribeCmd, artifactsFilesDownloadCmd,
		artifactsFilesListCmd, artifactsFilesUpdateCmd,
	)
	artifactsCmd.AddCommand(artifactsFilesCmd)
}

func artFileName(project, location, repo, raw string) string {
	return artFullName(artRepoParent(project, location, repo)+"/files", raw)
}

func runArtifactsFilesDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artFileName(project, flagArtFilesLocation, flagArtFilesRepository, args[0])
	op, err := svc.Projects.Locations.Repositories.Files.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting file: %w", err)
	}
	if flagArtFilesAsync {
		fmt.Fprintf(os.Stderr, "Delete started. Operation: %s\n", op.Name)
		return emitFormatted(op, flagArtFilesFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagArtFilesFormat)
}

func runArtifactsFilesDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artFileName(project, flagArtFilesLocation, flagArtFilesRepository, args[0])
	f, err := svc.Projects.Locations.Repositories.Files.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing file: %w", err)
	}
	return emitFormatted(f, flagArtFilesFormat)
}

func runArtifactsFilesDownload(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artFileName(project, flagArtFilesLocation, flagArtFilesRepository, args[0])
	resp, err := svc.Projects.Locations.Repositories.Files.Download(name).Context(ctx).Download()
	if err != nil {
		return fmt.Errorf("downloading file: %w", err)
	}
	defer resp.Body.Close()
	out, err := os.Create(flagArtFilesDestination)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Downloaded [%s] to [%s].\n", name, flagArtFilesDestination)
	return nil
}

func runArtifactsFilesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, flagArtFilesLocation, flagArtFilesRepository)
	var all []*artifactregistry.GoogleDevtoolsArtifactregistryV1File
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.Files.List(parent).Context(ctx)
		if flagArtFilesFilter != "" {
			call = call.Filter(flagArtFilesFilter)
		}
		if flagArtFilesPageSize > 0 {
			call = call.PageSize(flagArtFilesPageSize)
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
	return emitFormatted(all, flagArtFilesFormat)
}

func runArtifactsFilesUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &artifactregistry.GoogleDevtoolsArtifactregistryV1File{}
	if err := loadYAMLOrJSONInto(flagArtFilesConfigFile, body); err != nil {
		return err
	}
	mask := flagArtFilesMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := artFileName(project, flagArtFilesLocation, flagArtFilesRepository, args[0])
	call := svc.Projects.Locations.Repositories.Files.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating file: %w", err)
	}
	return emitFormatted(out, flagArtFilesFormat)
}
