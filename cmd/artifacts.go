package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	ondemandscanning "google.golang.org/api/ondemandscanning/v1"
)

var artifactsCmd = &cobra.Command{
	Use:   "artifacts",
	Short: "Manage Artifact Registry",
}

var artifactsDockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Manage Docker resources",
}

var artifactsDockerImagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Manage Docker images",
}

var artifactsScanCmd = &cobra.Command{
	Use:   "scan IMAGE",
	Short: "Scan a Docker image for vulnerabilities",
	Long: `Scan a container image for known vulnerabilities using On-Demand Scanning.
Example:
  gcloud artifacts docker images scan us-docker.pkg.dev/my-project/repo/image:tag --location=us`,
	Args: cobra.ExactArgs(1),
	RunE: runArtifactsScan,
}

var artifactsListVulnerabilitiesCmd = &cobra.Command{
	Use:   "list-vulnerabilities SCAN_RESOURCE",
	Short: "List vulnerabilities from a scan",
	Long: `List vulnerabilities found by an on-demand scan.
Example:
  gcloud artifacts docker images list-vulnerabilities projects/my-project/locations/us/scans/SCAN_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runArtifactsListVulnerabilities,
}

var (
	flagArtifactsScanLocation string
	flagArtifactsScanFormat   string
	flagArtifactsVulnFormat   string
)

func init() {
	artifactsScanCmd.Flags().StringVar(&flagArtifactsScanLocation, "location", "", "Location for the scan (e.g. us, europe)")
	artifactsScanCmd.Flags().StringVar(&flagArtifactsScanFormat, "format", "", "Output format (e.g. json)")
	artifactsListVulnerabilitiesCmd.Flags().StringVar(&flagArtifactsVulnFormat, "format", "", "Output format (e.g. json)")

	artifactsDockerImagesCmd.AddCommand(artifactsScanCmd)
	artifactsDockerImagesCmd.AddCommand(artifactsListVulnerabilitiesCmd)
	artifactsDockerCmd.AddCommand(artifactsDockerImagesCmd)
	artifactsCmd.AddCommand(artifactsDockerCmd)
	rootCmd.AddCommand(artifactsCmd)
}

func runArtifactsScan(cmd *cobra.Command, args []string) error {
	image := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	location := flagArtifactsScanLocation
	if location == "" {
		return fmt.Errorf("--location is required")
	}

	ctx := context.Background()
	svc, err := gcp.OnDemandScanningService(ctx, flagAccount)
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	req := &ondemandscanning.AnalyzePackagesRequestV1{
		ResourceUri: image,
	}

	op, err := svc.Projects.Locations.Scans.AnalyzePackages(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting scan: %w", err)
	}

	// Poll the operation until done.
	fmt.Fprintf(os.Stderr, "Scanning %s...\n", image)
	for !op.Done {
		time.Sleep(5 * time.Second)
		op, err = svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling scan operation: %w", err)
		}
	}

	if op.Error != nil {
		return fmt.Errorf("scan failed: %s", op.Error.Message)
	}

	if flagArtifactsScanFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(op.Response)
	}

	// The scan resource name is in the response.
	fmt.Printf("Scan completed. Operation: %s\n", op.Name)
	fmt.Println("Use 'gcloud artifacts docker images list-vulnerabilities' with the scan resource to view results.")
	return nil
}

func runArtifactsListVulnerabilities(cmd *cobra.Command, args []string) error {
	scanResource := args[0]

	ctx := context.Background()
	svc, err := gcp.OnDemandScanningService(ctx, flagAccount)
	if err != nil {
		return err
	}

	resp, err := svc.Projects.Locations.Scans.Vulnerabilities.List(scanResource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing vulnerabilities: %w", err)
	}

	if flagArtifactsVulnFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp.Occurrences)
	}

	if len(resp.Occurrences) == 0 {
		fmt.Println("No vulnerabilities found.")
		return nil
	}

	fmt.Printf("Found %d vulnerabilities:\n", len(resp.Occurrences))
	for _, occ := range resp.Occurrences {
		if occ.Vulnerability != nil {
			fmt.Printf("  Severity: %-10s Package: %s\n",
				occ.Vulnerability.Severity,
				occ.Vulnerability.ShortDescription)
		}
	}
	return nil
}
