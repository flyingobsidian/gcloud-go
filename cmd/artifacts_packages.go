package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts packages (#1077) ---

var artifactsPackagesCmd = &cobra.Command{
	Use:   "packages",
	Short: "Manage Artifact Registry packages",
}

var (
	flagArtPkgLocation   string
	flagArtPkgRepository string
	flagArtPkgFormat     string
	flagArtPkgFilter     string
	flagArtPkgPageSize   int64
	flagArtPkgConfigFile string
	flagArtPkgMask       string
	flagArtPkgAsync      bool
)

var artifactsPackagesDeleteCmd = &cobra.Command{
	Use:   "delete PACKAGE",
	Short: "Delete a package",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsPackagesDelete,
}

var artifactsPackagesDescribeCmd = &cobra.Command{
	Use:   "describe PACKAGE",
	Short: "Describe a package",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsPackagesDescribe,
}

var artifactsPackagesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List packages",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsPackagesList,
}

var artifactsPackagesUpdateCmd = &cobra.Command{
	Use:   "update PACKAGE",
	Short: "Update a package",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsPackagesUpdate,
}

func init() {
	for _, c := range []*cobra.Command{
		artifactsPackagesDeleteCmd, artifactsPackagesDescribeCmd,
		artifactsPackagesListCmd, artifactsPackagesUpdateCmd,
	} {
		c.Flags().StringVar(&flagArtPkgLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&flagArtPkgRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&flagArtPkgFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	artifactsPackagesDeleteCmd.Flags().BoolVar(&flagArtPkgAsync, "async", false, "Return without waiting")
	artifactsPackagesListCmd.Flags().StringVar(&flagArtPkgFilter, "filter", "", "Filter expression")
	artifactsPackagesListCmd.Flags().Int64Var(&flagArtPkgPageSize, "page-size", 0, "Page size")
	artifactsPackagesUpdateCmd.Flags().StringVar(&flagArtPkgConfigFile, "config-file", "", "YAML/JSON body for the Package update (required)")
	artifactsPackagesUpdateCmd.Flags().StringVar(&flagArtPkgMask, "update-mask", "", "Fields to update")
	_ = artifactsPackagesUpdateCmd.MarkFlagRequired("config-file")

	artifactsPackagesCmd.AddCommand(
		artifactsPackagesDeleteCmd, artifactsPackagesDescribeCmd,
		artifactsPackagesListCmd, artifactsPackagesUpdateCmd,
	)
	artifactsCmd.AddCommand(artifactsPackagesCmd)
}

func artPackageName(project, location, repo, raw string) string {
	return artFullName(artRepoParent(project, location, repo)+"/packages", raw)
}

func runArtifactsPackagesDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artPackageName(project, flagArtPkgLocation, flagArtPkgRepository, args[0])
	op, err := svc.Projects.Locations.Repositories.Packages.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting package: %w", err)
	}
	if flagArtPkgAsync {
		fmt.Fprintf(os.Stderr, "Delete started. Operation: %s\n", op.Name)
		return emitFormatted(op, flagArtPkgFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagArtPkgFormat)
}

func runArtifactsPackagesDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artPackageName(project, flagArtPkgLocation, flagArtPkgRepository, args[0])
	p, err := svc.Projects.Locations.Repositories.Packages.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing package: %w", err)
	}
	return emitFormatted(p, flagArtPkgFormat)
}

func runArtifactsPackagesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, flagArtPkgLocation, flagArtPkgRepository)
	var all []*artifactregistry.Package
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.Packages.List(parent).Context(ctx)
		if flagArtPkgFilter != "" {
			call = call.Filter(flagArtPkgFilter)
		}
		if flagArtPkgPageSize > 0 {
			call = call.PageSize(flagArtPkgPageSize)
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
	return emitFormatted(all, flagArtPkgFormat)
}

func runArtifactsPackagesUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &artifactregistry.Package{}
	if err := loadYAMLOrJSONInto(flagArtPkgConfigFile, body); err != nil {
		return err
	}
	mask := flagArtPkgMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := artPackageName(project, flagArtPkgLocation, flagArtPkgRepository, args[0])
	call := svc.Projects.Locations.Repositories.Packages.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating package: %w", err)
	}
	return emitFormatted(out, flagArtPkgFormat)
}
