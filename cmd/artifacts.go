package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
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
	flagArtifactsScanRemote   bool
	flagArtifactsScanAsync    bool
)

func init() {
	artifactsScanCmd.Flags().StringVar(&flagArtifactsScanLocation, "location", "", "Location for the scan (e.g. us, europe)")
	artifactsScanCmd.Flags().StringVar(&flagArtifactsScanFormat, "format", "", "Output format (e.g. json)")
	artifactsScanCmd.Flags().BoolVar(&flagArtifactsScanRemote, "remote", false, "Scan the image remotely in the registry (skip local extraction)")
	artifactsScanCmd.Flags().BoolVar(&flagArtifactsScanAsync, "async", false, "Return immediately without waiting for scan to complete")
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

	if !flagArtifactsScanRemote {
		// Local scan: extract packages from the image using Docker, then send
		// the package list to the API for vulnerability analysis.
		fmt.Fprintf(os.Stderr, "Locally extracting packages from %s...\n", image)
		pkgs, err := extractLocalPackages(ctx, image)
		if err != nil {
			return fmt.Errorf("local extraction failed: %w\nConsider using --remote to scan the image in the registry", err)
		}
		req.Packages = pkgs
		fmt.Fprintf(os.Stderr, "Extracted %d packages.\n", len(pkgs))
	}

	op, err := svc.Projects.Locations.Scans.AnalyzePackages(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting scan: %w", err)
	}

	if flagArtifactsScanAsync {
		if flagArtifactsScanFormat == "json" {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(op)
		}
		fmt.Printf("Scan started. Operation: %s\n", op.Name)
		return nil
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

	var allOccurrences []*ondemandscanning.Occurrence
	pageToken := ""
	for {
		call := svc.Projects.Locations.Scans.Vulnerabilities.List(scanResource).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing vulnerabilities: %w", err)
		}
		allOccurrences = append(allOccurrences, resp.Occurrences...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagArtifactsVulnFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allOccurrences)
	}

	if len(allOccurrences) == 0 {
		fmt.Println("No vulnerabilities found.")
		return nil
	}

	fmt.Printf("Found %d vulnerabilities:\n", len(allOccurrences))
	for _, occ := range allOccurrences {
		if occ.Vulnerability != nil {
			fmt.Printf("  Severity: %-10s Package: %s\n",
				occ.Vulnerability.Severity,
				occ.Vulnerability.ShortDescription)
		}
	}
	return nil
}

// extractLocalPackages runs Docker to extract installed OS packages from
// the image. It tries dpkg (Debian/Ubuntu), rpm (RHEL/CentOS), and apk
// (Alpine) in sequence, matching the local extraction that gcloud's
// local-extract binary performs.
func extractLocalPackages(ctx context.Context, image string) ([]*ondemandscanning.PackageData, error) {
	dockerBin, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("docker not found in PATH: %w", err)
	}

	// Detect OS and extract packages. Try each package manager in order.
	type extractor struct {
		name    string
		cmd     string
		cpeBase string
		parse   func(output string) []*ondemandscanning.PackageData
	}

	extractors := []extractor{
		{
			name:    "dpkg",
			cmd:     `dpkg-query -W -f '${Package}\t${Version}\t${Architecture}\n'`,
			cpeBase: "cpe:/o:debian:debian_linux",
			parse:   parseDpkgOutput,
		},
		{
			name:    "rpm",
			cmd:     `rpm -qa --queryformat '%{NAME}\t%{VERSION}-%{RELEASE}\t%{ARCH}\n'`,
			cpeBase: "cpe:/o:redhat:enterprise_linux",
			parse:   parseRpmOutput,
		},
		{
			name:    "apk",
			cmd:     `apk info -v 2>/dev/null | sort`,
			cpeBase: "cpe:/o:alpine:alpine_linux",
			parse:   parseApkOutput,
		},
	}

	for _, ext := range extractors {
		cmd := exec.CommandContext(ctx, dockerBin, "run", "--rm", "--entrypoint", "sh", image, "-c", ext.cmd)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			continue
		}
		output := stdout.String()
		if strings.TrimSpace(output) == "" {
			continue
		}
		pkgs := ext.parse(output)
		// Set cpeUri on all packages.
		for _, pkg := range pkgs {
			if pkg.CpeUri == "" {
				pkg.CpeUri = ext.cpeBase
			}
		}
		return pkgs, nil
	}

	return nil, fmt.Errorf("could not extract packages from image %s (no supported package manager found)", image)
}

func parseDpkgOutput(output string) []*ondemandscanning.PackageData {
	var pkgs []*ondemandscanning.PackageData
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 || parts[0] == "" {
			continue
		}
		pkg := &ondemandscanning.PackageData{
			Package:     parts[0],
			Version:     parts[1],
			PackageType: "DEBIAN",
		}
		if len(parts) >= 3 {
			pkg.Architecture = parts[2]
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

func parseRpmOutput(output string) []*ondemandscanning.PackageData {
	var pkgs []*ondemandscanning.PackageData
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 || parts[0] == "" {
			continue
		}
		pkg := &ondemandscanning.PackageData{
			Package:     parts[0],
			Version:     parts[1],
			PackageType: "RPM",
		}
		if len(parts) >= 3 {
			pkg.Architecture = parts[2]
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

func parseApkOutput(output string) []*ondemandscanning.PackageData {
	var pkgs []*ondemandscanning.PackageData
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// apk info -v output: "package-name-1.2.3-r0"
		// Split on last hyphen that precedes a digit to separate name from version.
		name, version := splitApkPackage(line)
		if name == "" {
			continue
		}
		pkgs = append(pkgs, &ondemandscanning.PackageData{
			Package:     name,
			Version:     version,
			PackageType: "APK",
		})
	}
	return pkgs
}

// splitApkPackage splits "package-name-1.2.3-r0" into ("package-name", "1.2.3-r0").
func splitApkPackage(s string) (string, string) {
	// Find the last hyphen followed by a digit.
	for i := len(s) - 1; i > 0; i-- {
		if s[i] == '-' && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9' {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}
