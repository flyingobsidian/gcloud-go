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
	artifactregistry "google.golang.org/api/artifactregistry/v1"
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

// --- artifacts docker images list (#186) ---

var artifactsDockerImagesListCmd = &cobra.Command{
	Use:   "list REPOSITORY",
	Short: "List Docker images in a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsDockerImagesList,
}

var (
	flagArtImgListFormat      string
	flagArtImgListIncludeTags bool
	flagArtImgListURI         bool
)

// --- artifacts docker images describe (#187) ---

var artifactsDockerImagesDescribeCmd = &cobra.Command{
	Use:   "describe IMAGE",
	Short: "Describe a Docker image",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsDockerImagesDescribe,
}

// --- artifacts docker images delete (#188) ---

var artifactsDockerImagesDeleteCmd = &cobra.Command{
	Use:   "delete IMAGE",
	Short: "Delete a Docker image",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsDockerImagesDelete,
}


// --- scan extra flags (#189, #190) ---

var (
	flagArtifactsScanLocation     string
	flagArtifactsScanFormat       string
	flagArtifactsVulnFormat       string
	flagArtifactsScanRemote       bool
	flagArtifactsScanAsync        bool
	flagArtifactsScanExtraTypes   []string
)

func init() {
	artifactsScanCmd.Flags().StringVar(&flagArtifactsScanLocation, "location", "", "Location for the scan (e.g. us, europe)")
	artifactsScanCmd.Flags().StringVar(&flagArtifactsScanFormat, "format", "", "Output format (e.g. json)")
	artifactsScanCmd.Flags().BoolVar(&flagArtifactsScanRemote, "remote", false, "Scan the image remotely in the registry (skip local extraction)")
	artifactsScanCmd.Flags().BoolVar(&flagArtifactsScanAsync, "async", false, "Return immediately without waiting for scan to complete")
	artifactsScanCmd.Flags().StringSliceVar(&flagArtifactsScanExtraTypes, "additional-package-types", nil, "Additional package types to scan (GO, MAVEN, PIP, NPM)")
	artifactsListVulnerabilitiesCmd.Flags().StringVar(&flagArtifactsVulnFormat, "format", "", "Output format (e.g. json)")

	artifactsDockerImagesListCmd.Flags().StringVar(&flagArtImgListFormat, "format", "", "Output format (e.g. json)")
	artifactsDockerImagesListCmd.Flags().BoolVar(&flagArtImgListIncludeTags, "include-tags", false, "Include image tags")
	artifactsDockerImagesListCmd.Flags().BoolVar(&flagArtImgListURI, "uri", false, "Print resource names")
	artifactsDockerImagesDescribeCmd.Flags().StringVar(&flagArtifactsScanFormat, "format", "", "Output format (e.g. json)")
	artifactsDockerImagesCmd.AddCommand(artifactsScanCmd)
	artifactsDockerImagesCmd.AddCommand(artifactsListVulnerabilitiesCmd)
	artifactsDockerImagesCmd.AddCommand(artifactsDockerImagesListCmd)
	artifactsDockerImagesCmd.AddCommand(artifactsDockerImagesDescribeCmd)
	artifactsDockerImagesCmd.AddCommand(artifactsDockerImagesDeleteCmd)
	artifactsDockerCmd.AddCommand(artifactsDockerImagesCmd)
	artifactsCmd.AddCommand(artifactsDockerCmd)

	// Stub registrations for artifacts subgroups present in gcloud-python but
	// not yet implemented in gcloud-go (#537). Registered so `--help` lists
	// them and later PRs can fill in real behavior per subgroup.
	registerStubGroup(artifactsCmd, "apt", "Manage Apt package operations", "list", "delete")
	registerStubGroup(artifactsCmd, "attachments", "Manage artifact attachments", "create", "delete", "describe", "list")
	registerStubGroup(artifactsCmd, "files", "Manage individual files in a repository", "delete", "describe", "list")
	registerStubGroup(artifactsCmd, "generic", "Manage generic package format", "download", "upload")
	registerStubGroup(artifactsCmd, "go", "Manage Go modules", "upload")
	registerStubGroup(artifactsCmd, "image-streaming-cache", "Manage image streaming caches", "create", "delete", "describe", "list")
	registerStubGroup(artifactsCmd, "locations", "List regional metadata", "list", "describe")
	registerStubGroup(artifactsCmd, "operations", "Manage long-running operations", "describe", "list", "wait")
	registerStubGroup(artifactsCmd, "packages", "Manage packages", "delete", "describe", "list")
	registerStubGroup(artifactsCmd, "projects", "Manage per-project Artifact Registry settings", "describe")
	registerStubGroup(artifactsCmd, "repositories", "Manage Artifact Registry repositories",
		"add-iam-policy-binding", "create", "delete", "describe", "get-iam-policy",
		"list", "remove-iam-policy-binding", "set-iam-policy", "update", "upgrade-to-remote")
	registerStubGroup(artifactsCmd, "rules", "Manage cleanup rules", "create", "delete", "describe", "list", "update")
	registerStubGroup(artifactsCmd, "sbom", "Manage SBOMs (Software Bill of Materials)", "export", "load")
	registerStubGroup(artifactsCmd, "settings", "Manage Artifact Registry settings", "describe", "enable-upgrade-redirection", "disable-upgrade-redirection")
	registerStubGroup(artifactsCmd, "tags", "Manage Docker/other tags", "create", "delete", "describe", "list", "update")
	registerStubGroup(artifactsCmd, "versions", "Manage package versions", "delete", "describe", "list", "tag")
	registerStubGroup(artifactsCmd, "vpcsc-config", "Manage VPC-SC configuration", "allow", "deny", "describe")
	registerStubGroup(artifactsCmd, "vulnerabilities", "Manage vulnerability reports", "list", "load-vex")

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

// --- artifacts docker images list (#186) ---

func runArtifactsDockerImagesList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}

	// Repository arg is like: LOCATION-docker.pkg.dev/PROJECT/REPO
	parent := fmt.Sprintf("projects/-/locations/-/repositories/-/packages/-")
	// Try to parse the repository path into the API format.
	parts := strings.Split(strings.TrimSuffix(args[0], "/"), "/")
	if len(parts) >= 3 {
		// e.g., us-docker.pkg.dev/my-project/my-repo
		host := parts[0]
		location := strings.TrimSuffix(host, "-docker.pkg.dev")
		project := parts[1]
		repo := parts[2]
		parent = fmt.Sprintf("projects/%s/locations/%s/repositories/%s/packages/-", project, location, repo)
	}

	var allImages []*artifactregistry.DockerImage
	pageToken := ""
	for {
		call := svc.Projects.Locations.Repositories.DockerImages.List(strings.TrimSuffix(parent, "/packages/-")).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing docker images: %w", err)
		}
		allImages = append(allImages, resp.DockerImages...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagArtImgListURI {
		for _, img := range allImages {
			fmt.Println(img.Name)
		}
		return nil
	}

	if flagArtImgListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allImages)
	}

	fmt.Printf("%-60s %-20s %s\n", "IMAGE", "DIGEST", "TAGS")
	for _, img := range allImages {
		tags := ""
		if flagArtImgListIncludeTags && len(img.Tags) > 0 {
			tags = strings.Join(img.Tags, ",")
		}
		fmt.Printf("%-60s %-20s %s\n", img.Uri, truncateDigest(img.Uri), tags)
	}
	return nil
}

func truncateDigest(uri string) string {
	if idx := strings.Index(uri, "@sha256:"); idx >= 0 {
		digest := uri[idx+8:]
		if len(digest) > 12 {
			return "sha256:" + digest[:12]
		}
		return "sha256:" + digest
	}
	return ""
}

// --- artifacts docker images describe (#187) ---

func runArtifactsDockerImagesDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}

	// Parse the image path. Format: LOCATION-docker.pkg.dev/PROJECT/REPO/IMAGE[@sha256:DIGEST]
	image := args[0]
	name, err := parseArtifactImageName(image)
	if err != nil {
		return err
	}

	img, err := svc.Projects.Locations.Repositories.DockerImages.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing docker image: %w", err)
	}

	return formatOutput(img, "")
}

func parseArtifactImageName(image string) (string, error) {
	// Input: us-docker.pkg.dev/project/repo/image@sha256:abc123
	// Output: projects/project/locations/us/repositories/repo/dockerImages/image@sha256:abc123
	parts := strings.SplitN(image, "/", 4)
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid image path: expected LOCATION-docker.pkg.dev/PROJECT/REPO/IMAGE")
	}
	host := parts[0]
	location := strings.TrimSuffix(host, "-docker.pkg.dev")
	project := parts[1]
	repo := parts[2]
	imagePart := parts[3]
	return fmt.Sprintf("projects/%s/locations/%s/repositories/%s/dockerImages/%s", project, location, repo, imagePart), nil
}

// --- artifacts docker images delete (#188) ---

func runArtifactsDockerImagesDelete(cmd *cobra.Command, args []string) error {
	if !flagQuiet {
		fmt.Printf("You are about to delete image [%s].\n", args[0])
		fmt.Print("Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}

	// Parse image path to get the package name.
	parts := strings.SplitN(args[0], "/", 4)
	if len(parts) < 4 {
		return fmt.Errorf("invalid image path: expected LOCATION-docker.pkg.dev/PROJECT/REPO/IMAGE")
	}
	host := parts[0]
	location := strings.TrimSuffix(host, "-docker.pkg.dev")
	project := parts[1]
	repo := parts[2]
	imagePart := parts[3]

	// Strip tag/digest for package name.
	pkgName := imagePart
	if idx := strings.Index(pkgName, "@"); idx >= 0 {
		pkgName = pkgName[:idx]
	}
	if idx := strings.Index(pkgName, ":"); idx >= 0 {
		pkgName = pkgName[:idx]
	}

	name := fmt.Sprintf("projects/%s/locations/%s/repositories/%s/packages/%s", project, location, repo, pkgName)
	if _, err := svc.Projects.Locations.Repositories.Packages.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}

	fmt.Printf("Deleted image [%s].\n", args[0])
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

	// Application-level extractors (for --additional-package-types or expanded extraction).
	appExtractors := []extractor{
		{
			name:    "go",
			cmd:     `find / -name go.sum -exec cat {} \; 2>/dev/null`,
			cpeBase: "cpe:/a:golang:go",
			parse:   parseGoSumOutput,
		},
		{
			name:    "pip",
			cmd:     `pip list --format=freeze 2>/dev/null || pip3 list --format=freeze 2>/dev/null`,
			cpeBase: "cpe:/a:python:python",
			parse:   parsePipOutput,
		},
		{
			name:    "npm",
			cmd:     `find / -name package.json -not -path '*/node_modules/.package-lock.json' -exec sh -c 'cat "$1" 2>/dev/null' _ {} \; 2>/dev/null | head -5000`,
			cpeBase: "cpe:/a:npmjs:npm",
			parse:   parseNpmOutput,
		},
		{
			name:    "maven",
			cmd:     `find / -name pom.xml -exec grep -h '<artifactId>\|<version>' {} \; 2>/dev/null | head -5000`,
			cpeBase: "cpe:/a:apache:maven",
			parse:   parseMavenOutput,
		},
	}

	wantExtra := make(map[string]bool)
	for _, t := range flagArtifactsScanExtraTypes {
		wantExtra[strings.ToUpper(t)] = true
	}

	var allPkgs []*ondemandscanning.PackageData
	// Run OS extractors first.
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
		for _, pkg := range pkgs {
			if pkg.CpeUri == "" {
				pkg.CpeUri = ext.cpeBase
			}
		}
		allPkgs = append(allPkgs, pkgs...)
		break // Only use the first successful OS extractor
	}

	// Run application extractors if requested.
	if len(wantExtra) > 0 {
		for _, ext := range appExtractors {
			if !wantExtra[strings.ToUpper(ext.name)] {
				continue
			}
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
			for _, pkg := range pkgs {
				if pkg.CpeUri == "" {
					pkg.CpeUri = ext.cpeBase
				}
			}
			allPkgs = append(allPkgs, pkgs...)
		}
	}

	if len(allPkgs) == 0 {
		return nil, fmt.Errorf("could not extract packages from image %s (no supported package manager found)", image)
	}
	return allPkgs, nil
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

// parseGoSumOutput parses go.sum lines: "module v1.2.3 h1:hash="
func parseGoSumOutput(output string) []*ondemandscanning.PackageData {
	seen := make(map[string]bool)
	var pkgs []*ondemandscanning.PackageData
	for _, line := range strings.Split(output, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		mod := parts[0]
		ver := strings.TrimSuffix(parts[1], "/go.mod")
		ver = strings.TrimPrefix(ver, "v")
		key := mod + "@" + ver
		if seen[key] {
			continue
		}
		seen[key] = true
		pkgs = append(pkgs, &ondemandscanning.PackageData{
			Package:     mod,
			Version:     ver,
			PackageType: "GO",
		})
	}
	return pkgs
}

// parsePipOutput parses "pip list --format=freeze" output: "package==version"
func parsePipOutput(output string) []*ondemandscanning.PackageData {
	var pkgs []*ondemandscanning.PackageData
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "==", 2)
		if len(parts) != 2 {
			continue
		}
		pkgs = append(pkgs, &ondemandscanning.PackageData{
			Package:     parts[0],
			Version:     parts[1],
			PackageType: "PIP",
		})
	}
	return pkgs
}

// parseNpmOutput is a basic parser for package.json name/version.
func parseNpmOutput(output string) []*ondemandscanning.PackageData {
	var pkgs []*ondemandscanning.PackageData
	type pkgJSON struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	// Try to parse each JSON object in the output.
	decoder := json.NewDecoder(strings.NewReader(output))
	for decoder.More() {
		var p pkgJSON
		if err := decoder.Decode(&p); err != nil {
			break
		}
		if p.Name != "" && p.Version != "" {
			pkgs = append(pkgs, &ondemandscanning.PackageData{
				Package:     p.Name,
				Version:     p.Version,
				PackageType: "NPM",
			})
		}
	}
	return pkgs
}

// parseMavenOutput parses artifactId and version from Maven pom.xml grep output.
func parseMavenOutput(output string) []*ondemandscanning.PackageData {
	var pkgs []*ondemandscanning.PackageData
	lines := strings.Split(output, "\n")
	for i := 0; i < len(lines)-1; i++ {
		aid := extractXMLValue(lines[i], "artifactId")
		if aid == "" {
			continue
		}
		ver := extractXMLValue(lines[i+1], "version")
		if ver == "" {
			continue
		}
		pkgs = append(pkgs, &ondemandscanning.PackageData{
			Package:     aid,
			Version:     ver,
			PackageType: "MAVEN",
		})
		i++
	}
	return pkgs
}

func extractXMLValue(line, tag string) string {
	line = strings.TrimSpace(line)
	prefix := "<" + tag + ">"
	suffix := "</" + tag + ">"
	if strings.HasPrefix(line, prefix) && strings.HasSuffix(line, suffix) {
		return line[len(prefix) : len(line)-len(suffix)]
	}
	return ""
}
