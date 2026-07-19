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

// --- gcloud artifacts versions (#1084) ---

var artifactsVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Manage Artifact Registry package versions",
}

var (
	flagArtVerLocation    string
	flagArtVerRepository  string
	flagArtVerPackage     string
	flagArtVerFormat      string
	flagArtVerFilter      string
	flagArtVerConfigFile  string
	flagArtVerMask        string
	flagArtVerDestination string
	flagArtVerAsync       bool
)

var artifactsVersionsDeleteCmd = &cobra.Command{
	Use:   "delete VERSION",
	Short: "Delete a version",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsVersionsDelete,
}

var artifactsVersionsDescribeCmd = &cobra.Command{
	Use:   "describe VERSION",
	Short: "Describe a version",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsVersionsDescribe,
}

var artifactsVersionsExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export versions for a package to a JSON file",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsVersionsExport,
}

var artifactsVersionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List versions of a package",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsVersionsList,
}

var artifactsVersionsUpdateCmd = &cobra.Command{
	Use:   "update VERSION",
	Short: "Update a version",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsVersionsUpdate,
}

func init() {
	for _, c := range []*cobra.Command{
		artifactsVersionsDeleteCmd, artifactsVersionsDescribeCmd, artifactsVersionsExportCmd,
		artifactsVersionsListCmd, artifactsVersionsUpdateCmd,
	} {
		c.Flags().StringVar(&flagArtVerLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&flagArtVerRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&flagArtVerPackage, "package", "", "Package (required)")
		c.Flags().StringVar(&flagArtVerFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
		_ = c.MarkFlagRequired("package")
	}
	artifactsVersionsDeleteCmd.Flags().BoolVar(&flagArtVerAsync, "async", false, "Return without waiting for operation")
	artifactsVersionsListCmd.Flags().StringVar(&flagArtVerFilter, "filter", "", "Filter expression")
	artifactsVersionsUpdateCmd.Flags().StringVar(&flagArtVerConfigFile, "config-file", "", "YAML/JSON body for the Version update (required)")
	artifactsVersionsUpdateCmd.Flags().StringVar(&flagArtVerMask, "update-mask", "", "Fields to update")
	_ = artifactsVersionsUpdateCmd.MarkFlagRequired("config-file")
	artifactsVersionsExportCmd.Flags().StringVar(&flagArtVerDestination, "destination", "", "Local JSON file to write versions to (required)")
	_ = artifactsVersionsExportCmd.MarkFlagRequired("destination")

	artifactsVersionsCmd.AddCommand(
		artifactsVersionsDeleteCmd, artifactsVersionsDescribeCmd, artifactsVersionsExportCmd,
		artifactsVersionsListCmd, artifactsVersionsUpdateCmd,
	)
	artifactsCmd.AddCommand(artifactsVersionsCmd)
}

func artVerPackageParent(project string) string {
	return artFullName(artRepoParent(project, flagArtVerLocation, flagArtVerRepository)+"/packages", flagArtVerPackage)
}

func artVerName(project, id string) string {
	return artFullName(artVerPackageParent(project)+"/versions", id)
}

func runArtifactsVersionsDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artVerName(project, args[0])
	op, err := svc.Projects.Locations.Repositories.Packages.Versions.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting version: %w", err)
	}
	if flagArtVerAsync {
		fmt.Fprintf(os.Stderr, "Delete started. Operation: %s\n", op.Name)
		return emitFormatted(op, flagArtVerFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagArtVerFormat)
}

func runArtifactsVersionsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artVerName(project, args[0])
	v, err := svc.Projects.Locations.Repositories.Packages.Versions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing version: %w", err)
	}
	return emitFormatted(v, flagArtVerFormat)
}

func runArtifactsVersionsExport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	all, err := listArtifactsVersions(project)
	if err != nil {
		return err
	}
	f, err := os.Create(flagArtVerDestination)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(all); err != nil {
		return fmt.Errorf("writing versions: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Exported %d versions to [%s].\n", len(all), flagArtVerDestination)
	return nil
}

func runArtifactsVersionsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	all, err := listArtifactsVersions(project)
	if err != nil {
		return err
	}
	return emitFormatted(all, flagArtVerFormat)
}

func listArtifactsVersions(project string) ([]*artifactregistry.Version, error) {
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return nil, err
	}
	pkg := artVerPackageParent(project)
	var all []*artifactregistry.Version
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.Packages.Versions.List(pkg).Context(ctx)
		if flagArtVerFilter != "" {
			call = call.Filter(flagArtVerFilter)
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

func runArtifactsVersionsUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &artifactregistry.Version{}
	if err := loadYAMLOrJSONInto(flagArtVerConfigFile, body); err != nil {
		return err
	}
	mask := flagArtVerMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := artVerName(project, args[0])
	call := svc.Projects.Locations.Repositories.Packages.Versions.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating version: %w", err)
	}
	return emitFormatted(out, flagArtVerFormat)
}
