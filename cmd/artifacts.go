package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
	ondemandscanning "google.golang.org/api/ondemandscanning/v1"
)

// --- gcloud artifacts (root) ---

// artifactsRest is used for the handful of Artifact Registry surfaces that are
// not (yet) exposed by the generated Go client, e.g. imageStreamingCaches,
// per-location settings, SBOM export/load, and vulnerabilities.
func artifactsRest() *restClient {
	return newRESTClient("https://artifactregistry.googleapis.com/v1")
}

// --- shared parent resolution helpers ---

// artProjectParent returns "projects/PROJECT".
func artProjectParent(project string) string {
	return "projects/" + project
}

// artLocationParent returns "projects/PROJECT/locations/LOCATION".
func artLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

// artRepoParent returns "projects/PROJECT/locations/LOCATION/repositories/REPO".
func artRepoParent(project, location, repo string) string {
	return fmt.Sprintf("projects/%s/locations/%s/repositories/%s", project, location, repo)
}

// artFullName returns raw if it already looks like a fully-qualified
// "projects/..." resource name; otherwise it joins parent and raw.
func artFullName(parent, raw string) string {
	if strings.HasPrefix(raw, "projects/") {
		return raw
	}
	return parent + "/" + raw
}

// --- shared IAM helpers ---

// artIamBuildCondition returns a *artifactregistry.Expr or nil.
func artIamBuildCondition(expr, title, desc string) *artifactregistry.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &artifactregistry.Expr{Expression: expr, Title: title, Description: desc}
}

// artIamCondsEqual reports whether two IAM condition exprs are equivalent.
func artIamCondsEqual(a, b *artifactregistry.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

// artIamAddBinding adds (role, member[, cond]) to policy in-place.
func artIamAddBinding(policy *artifactregistry.Policy, role, member string, cond *artifactregistry.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !artIamCondsEqual(b.Condition, cond) {
			continue
		}
		for _, m := range b.Members {
			if m == member {
				return
			}
		}
		b.Members = append(b.Members, member)
		return
	}
	policy.Bindings = append(policy.Bindings, &artifactregistry.Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

// artIamRemoveBinding removes (role, member[, cond]) from policy in-place.
// If allConds is true, matching bindings across all conditions are cleared.
func artIamRemoveBinding(policy *artifactregistry.Policy, role, member string, cond *artifactregistry.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || artIamCondsEqual(b.Condition, cond))
		if !match {
			kept = append(kept, b)
			continue
		}
		newMembers := b.Members[:0]
		for _, m := range b.Members {
			if m == member {
				continue
			}
			newMembers = append(newMembers, m)
		}
		if len(newMembers) != len(b.Members) {
			changed = true
		}
		b.Members = newMembers
		if len(b.Members) > 0 {
			kept = append(kept, b)
		} else {
			changed = true
		}
	}
	policy.Bindings = kept
	return changed
}

// artUpdatedIam prints the standard "Updated IAM policy for ..." message.
func (s *artifactsCmds) artUpdatedIam(who string) {
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for %s.\n", who)
}

// artIamMemberFlags binds the common --member/--role/--condition-* flags for a
// (add|remove)-iam-policy-binding subcommand.
func artIamMemberFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

// --- existing scan/list-vulnerabilities surface (docker images scan/etc.) ---

func (s *artifactsCmds) initArtifacts() {
	s.artifactsScanCmd.Flags().StringVar(&s.flagArtifactsScanLocation, "location", "", "Location for the scan (e.g. us, europe)")
	s.artifactsScanCmd.Flags().StringVar(&s.flagArtifactsScanFormat, "format", "", "Output format (e.g. json)")
	s.artifactsScanCmd.Flags().BoolVar(&s.flagArtifactsScanRemote, "remote", false, "Scan the image remotely in the registry (skip local extraction)")
	s.artifactsScanCmd.Flags().BoolVar(&s.flagArtifactsScanAsync, "async", false, "Return immediately without waiting for scan to complete")
	s.artifactsScanCmd.Flags().StringSliceVar(&s.flagArtifactsScanExtraTypes, "additional-package-types", nil, "Additional package types to scan (GO, MAVEN, PIP, NPM)")
	s.artifactsListVulnerabilitiesCmd.Flags().StringVar(&s.flagArtifactsVulnFormat, "format", "", "Output format (e.g. json)")

	s.artifactsDockerImagesListCmd.Flags().StringVar(&s.flagArtImgListFormat, "format", "", "Output format (e.g. json)")
	s.artifactsDockerImagesListCmd.Flags().BoolVar(&s.flagArtImgListIncludeTags, "include-tags", false, "Include image tags")
	s.artifactsDockerImagesListCmd.Flags().BoolVar(&s.flagArtImgListURI, "uri", false, "Print resource names")
	s.artifactsDockerImagesListCmd.Flags().Int64Var(&s.flagArtImgListLimit, "limit", 0,
		"Maximum number of images to list. Bounding the fetch avoids the slow full-repo scan"+
			" when a repository holds a large number of image versions (0 = no limit).")
	s.artifactsDockerImagesListCmd.Flags().Int64Var(&s.flagArtImgListPageSize, "page-size", 0,
		"Page size for API pagination (0 = server default)")
	s.artifactsDockerImagesDescribeCmd.Flags().StringVar(&s.flagArtifactsScanFormat, "format", "", "Output format (e.g. json)")

	s.artifactsDockerImagesCmd.AddCommand(s.artifactsScanCmd)
	s.artifactsDockerImagesCmd.AddCommand(s.artifactsListVulnerabilitiesCmd)
	s.artifactsDockerImagesCmd.AddCommand(s.artifactsDockerImagesListCmd)
	s.artifactsDockerImagesCmd.AddCommand(s.artifactsDockerImagesDescribeCmd)
	s.artifactsDockerImagesCmd.AddCommand(s.artifactsDockerImagesDeleteCmd)
	s.artifactsDockerCmd.AddCommand(s.artifactsDockerImagesCmd)
	s.artifactsCmd.AddCommand(s.artifactsDockerCmd)

}

func (s *artifactsCmds) runArtifactsScan(cmd *cobra.Command, args []string) error {
	image := args[0]
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	location := s.flagArtifactsScanLocation
	if location == "" {
		return fmt.Errorf("--location is required")
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().OnDemandScanning(ctx)
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	req := &ondemandscanning.AnalyzePackagesRequestV1{
		ResourceUri: image,
	}

	if !s.flagArtifactsScanRemote {
		fmt.Fprintf(s.inv.stderr, "Locally extracting packages from %s...\n", image)
		pkgs, err := s.extractLocalPackages(ctx, image)
		if err != nil {
			return fmt.Errorf("local extraction failed: %w\nConsider using --remote to scan the image in the registry", err)
		}
		req.Packages = pkgs
		fmt.Fprintf(s.inv.stderr, "Extracted %d packages.\n", len(pkgs))
	}

	op, err := svc.Projects.Locations.Scans.AnalyzePackages(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting scan: %w", err)
	}

	if s.flagArtifactsScanAsync {
		if s.flagArtifactsScanFormat == "json" {
			return s.inv.printJSON(op)
		}
		fmt.Fprintf(s.inv.stdout, "Scan started. Operation: %s\n", op.Name)
		return nil
	}

	fmt.Fprintf(s.inv.stderr, "Scanning %s...\n", image)
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

	if s.flagArtifactsScanFormat == "json" {
		return s.inv.printJSON(op.Response)
	}

	fmt.Fprintf(s.inv.stdout, "Scan completed. Operation: %s\n", op.Name)
	fmt.Fprintln(s.inv.stdout, "Use 'gcloud artifacts docker images list-vulnerabilities' with the scan resource to view results.")
	return nil
}

func (s *artifactsCmds) runArtifactsListVulnerabilities(cmd *cobra.Command, args []string) error {
	scanResource := args[0]

	ctx := cmd.Context()
	svc, err := s.inv.clients().OnDemandScanning(ctx)
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

	if s.flagArtifactsVulnFormat == "json" {
		return s.inv.printJSON(allOccurrences)
	}

	if len(allOccurrences) == 0 {
		fmt.Fprintln(s.inv.stdout, "No vulnerabilities found.")
		return nil
	}

	fmt.Fprintf(s.inv.stdout, "Found %d vulnerabilities:\n", len(allOccurrences))
	for _, occ := range allOccurrences {
		if occ.Vulnerability != nil {
			fmt.Fprintf(s.inv.stdout, "  Severity: %-10s Package: %s\n",
				occ.Vulnerability.Severity,
				occ.Vulnerability.ShortDescription)
		}
	}
	return nil
}

func (s *artifactsCmds) runArtifactsDockerImagesList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/-/locations/-/repositories/-/packages/-")
	if location, project, repo, _, ok := parseDockerPath(args[0]); ok {
		parent = fmt.Sprintf("projects/%s/locations/%s/repositories/%s/packages/-", project, location, repo)
	}

	var allImages []*artifactregistry.DockerImage
	pageToken := ""
	for {
		call := svc.Projects.Locations.Repositories.DockerImages.List(strings.TrimSuffix(parent, "/packages/-")).Context(ctx)
		if s.flagArtImgListPageSize > 0 {
			call = call.PageSize(s.flagArtImgListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing docker images: %w", err)
		}
		allImages = append(allImages, resp.DockerImages...)
		if s.flagArtImgListLimit > 0 && int64(len(allImages)) >= s.flagArtImgListLimit {
			allImages = allImages[:s.flagArtImgListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if s.flagArtImgListURI {
		for _, img := range allImages {
			fmt.Fprintln(s.inv.stdout, img.Name)
		}
		return nil
	}

	if s.flagArtImgListFormat == "json" {
		return s.inv.printJSON(allImages)
	}

	printTableHeadingTo(s.inv.stdout, len(allImages), "%-60s %-20s %s\n", "IMAGE", "DIGEST", "TAGS")
	for _, img := range allImages {
		tags := ""
		if s.flagArtImgListIncludeTags && len(img.Tags) > 0 {
			tags = strings.Join(img.Tags, ",")
		}
		fmt.Fprintf(s.inv.stdout, "%-60s %-20s %s\n", img.Uri, truncateDigest(img.Uri), tags)
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

func (s *artifactsCmds) runArtifactsDockerImagesDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}

	image := args[0]
	name, err := parseArtifactImageName(image)
	if err != nil {
		return err
	}

	img, err := svc.Projects.Locations.Repositories.DockerImages.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing docker image: %w", err)
	}

	return s.inv.formatOutput(img, "")
}

func parseArtifactImageName(image string) (string, error) {
	location, project, repo, imagePart, ok := parseDockerPath(image)
	if !ok || imagePart == "" {
		return "", fmt.Errorf("invalid image path: expected LOCATION-docker.pkg.dev/PROJECT/REPO/IMAGE")
	}
	return fmt.Sprintf("projects/%s/locations/%s/repositories/%s/dockerImages/%s", project, location, repo, imagePart), nil
}

// parseDockerPath splits an Artifact Registry Docker path
// ([https://]LOCATION-docker.pkg.dev/PROJECT/REPO[/IMAGE...]) into its parts.
// A domain-scoped project ("example.com:proj") appears in Docker paths as two
// segments ("example.com/proj"); a first project segment containing a dot is
// read that way and rejoined with ":", matching gcloud-python 586.0.0's
// docker_util regexes. image is everything after REPO ("" if absent).
func parseDockerPath(path string) (location, project, repo, image string, ok bool) {
	path = strings.TrimPrefix(path, "https://")
	var parts []string
	for _, p := range strings.Split(path, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) < 3 {
		return "", "", "", "", false
	}
	host := parts[0]
	if i := strings.Index(host, "docker."); i > 0 && (host[i-1] == '-' || host[i-1] == '.') {
		location = host[:i-1]
	} else {
		location = strings.TrimSuffix(host, "-docker.pkg.dev")
	}
	repoIndex := 2
	if strings.Contains(parts[1], ".") && len(parts) >= 4 {
		project = parts[1] + ":" + parts[2]
		repoIndex = 3
	} else {
		project = parts[1]
	}
	repo = parts[repoIndex]
	image = strings.Join(parts[repoIndex+1:], "/")
	return location, project, repo, image, true
}

func (s *artifactsCmds) runArtifactsDockerImagesDelete(cmd *cobra.Command, args []string) error {
	if !s.inv.flags.quiet {
		fmt.Fprintf(s.inv.stdout, "You are about to delete image [%s].\n", args[0])
		fmt.Fprint(s.inv.stdout, "Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Fprintln(s.inv.stdout, "Aborted.")
			return nil
		}
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}

	location, project, repo, imagePart, ok := parseDockerPath(args[0])
	if !ok || imagePart == "" {
		return fmt.Errorf("invalid image path: expected LOCATION-docker.pkg.dev/PROJECT/REPO/IMAGE")
	}

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

	fmt.Fprintf(s.inv.stdout, "Deleted image [%s].\n", args[0])
	return nil
}

// --- package extraction helpers (unchanged from earlier revisions) ---

func (s *artifactsCmds) extractLocalPackages(ctx context.Context, image string) ([]*ondemandscanning.PackageData, error) {
	dockerBin, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("docker not found in PATH: %w", err)
	}

	type extractor struct {
		name    string
		cmd     string
		cpeBase string
		parse   func(output string) []*ondemandscanning.PackageData
	}

	extractors := []extractor{
		{name: "dpkg", cmd: `dpkg-query -W -f '${Package}\t${Version}\t${Architecture}\n'`, cpeBase: "cpe:/o:debian:debian_linux", parse: parseDpkgOutput},
		{name: "rpm", cmd: `rpm -qa --queryformat '%{NAME}\t%{VERSION}-%{RELEASE}\t%{ARCH}\n'`, cpeBase: "cpe:/o:redhat:enterprise_linux", parse: parseRpmOutput},
		{name: "apk", cmd: `apk info -v 2>/dev/null | sort`, cpeBase: "cpe:/o:alpine:alpine_linux", parse: parseApkOutput},
	}

	appExtractors := []extractor{
		{name: "go", cmd: `find / -name go.sum -exec cat {} \; 2>/dev/null`, cpeBase: "cpe:/a:golang:go", parse: parseGoSumOutput},
		{name: "pip", cmd: `pip list --format=freeze 2>/dev/null || pip3 list --format=freeze 2>/dev/null`, cpeBase: "cpe:/a:python:python", parse: parsePipOutput},
		{name: "npm", cmd: `find / -name package.json -not -path '*/node_modules/.package-lock.json' -exec sh -c 'cat "$1" 2>/dev/null' _ {} \; 2>/dev/null | head -5000`, cpeBase: "cpe:/a:npmjs:npm", parse: parseNpmOutput},
		{name: "maven", cmd: `find / -name pom.xml -exec grep -h '<artifactId>\|<version>' {} \; 2>/dev/null | head -5000`, cpeBase: "cpe:/a:apache:maven", parse: parseMavenOutput},
	}

	wantExtra := make(map[string]bool)
	for _, t := range s.flagArtifactsScanExtraTypes {
		wantExtra[strings.ToUpper(t)] = true
	}

	var allPkgs []*ondemandscanning.PackageData
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
		break
	}

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

func splitApkPackage(s string) (string, string) {
	for i := len(s) - 1; i > 0; i-- {
		if s[i] == '-' && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9' {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}

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

func parseNpmOutput(output string) []*ondemandscanning.PackageData {
	var pkgs []*ondemandscanning.PackageData
	type pkgJSON struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
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

// artifactsCmds holds the commands and flag values of one invocation's "artifacts"
// command group, so invocations don't share state (#2070).
type artifactsCmds struct {
	inv *invocation

	artifactsCmd *cobra.Command
	// artifactsDockerCmd and artifactsDockerImagesCmd are shared parents used by
	// the vulnerability-scan surface below.
	artifactsDockerCmd                      *cobra.Command
	artifactsDockerImagesCmd                *cobra.Command
	artifactsScanCmd                        *cobra.Command
	artifactsListVulnerabilitiesCmd         *cobra.Command
	artifactsDockerImagesListCmd            *cobra.Command
	flagArtImgListFormat                    string
	flagArtImgListIncludeTags               bool
	flagArtImgListURI                       bool
	flagArtImgListLimit                     int64
	flagArtImgListPageSize                  int64
	artifactsDockerImagesDescribeCmd        *cobra.Command
	artifactsDockerImagesDeleteCmd          *cobra.Command
	flagArtifactsScanLocation               string
	flagArtifactsScanFormat                 string
	flagArtifactsVulnFormat                 string
	flagArtifactsScanRemote                 bool
	flagArtifactsScanAsync                  bool
	flagArtifactsScanExtraTypes             []string
	artifactsAptCmd                         *cobra.Command
	flagArtAptLocation                      string
	flagArtAptRepository                    string
	flagArtAptFormat                        string
	flagArtAptConfigFile                    string
	flagArtAptSource                        string
	flagArtAptAsync                         bool
	artifactsAptImportCmd                   *cobra.Command
	artifactsAptUploadCmd                   *cobra.Command
	artifactsAttachmentsCmd                 *cobra.Command
	flagArtAttLocation                      string
	flagArtAttRepository                    string
	flagArtAttFormat                        string
	flagArtAttFilter                        string
	flagArtAttConfigFile                    string
	flagArtAttTarget                        string
	flagArtAttType                          string
	flagArtAttDestination                   string
	flagArtAttAsync                         bool
	artifactsAttachmentsCreateCmd           *cobra.Command
	artifactsAttachmentsDeleteCmd           *cobra.Command
	artifactsAttachmentsDescribeCmd         *cobra.Command
	artifactsAttachmentsDownloadCmd         *cobra.Command
	artifactsAttachmentsListCmd             *cobra.Command
	artifactsFilesCmd                       *cobra.Command
	flagArtFilesLocation                    string
	flagArtFilesRepository                  string
	flagArtFilesFormat                      string
	flagArtFilesFilter                      string
	flagArtFilesPageSize                    int64
	flagArtFilesDestination                 string
	flagArtFilesConfigFile                  string
	flagArtFilesMask                        string
	flagArtFilesAsync                       bool
	artifactsFilesDeleteCmd                 *cobra.Command
	artifactsFilesDescribeCmd               *cobra.Command
	artifactsFilesDownloadCmd               *cobra.Command
	artifactsFilesListCmd                   *cobra.Command
	artifactsFilesUpdateCmd                 *cobra.Command
	flagArtFilesUpFile                      string
	flagArtFilesUpFileType                  string
	flagArtFilesUpSource                    string
	flagArtFilesUpSourceDir                 string
	flagArtFilesUpSkipExisting              bool
	flagArtFilesUpAsync                     bool
	artifactsFilesUploadCmd                 *cobra.Command
	artifactsGenericCmd                     *cobra.Command
	flagArtGenLocation                      string
	flagArtGenRepository                    string
	flagArtGenFormat                        string
	flagArtGenSource                        string
	flagArtGenPackage                       string
	flagArtGenVersion                       string
	flagArtGenDestination                   string
	flagArtGenAsync                         bool
	artifactsGenericDownloadCmd             *cobra.Command
	artifactsGenericUploadCmd               *cobra.Command
	artifactsGoCmd                          *cobra.Command
	flagArtGoLocation                       string
	flagArtGoRepository                     string
	flagArtGoSource                         string
	flagArtGoModule                         string
	flagArtGoVersion                        string
	flagArtGoFormat                         string
	flagArtGoAsync                          bool
	artifactsGoAuthCmd                      *cobra.Command
	artifactsGoUploadCmd                    *cobra.Command
	artifactsImageStreamingCacheCmd         *cobra.Command
	flagArtISCLocation                      string
	flagArtISCFormat                        string
	flagArtISCConfigFile                    string
	artifactsImageStreamingCacheCreateCmd   *cobra.Command
	artifactsImageStreamingCacheDeleteCmd   *cobra.Command
	artifactsImageStreamingCacheDescribeCmd *cobra.Command
	artifactsImageStreamingCacheListCmd     *cobra.Command
	artifactsLocationsCmd                   *cobra.Command
	flagArtLocationsFormat                  string
	flagArtLocationsFilter                  string
	flagArtLocationsPageSize                int64
	artifactsLocationsListCmd               *cobra.Command
	artifactsOperationsCmd                  *cobra.Command
	flagArtOpsLocation                      string
	flagArtOpsFormat                        string
	artifactsOperationsDescribeCmd          *cobra.Command
	artifactsPackagesCmd                    *cobra.Command
	flagArtPkgLocation                      string
	flagArtPkgRepository                    string
	flagArtPkgFormat                        string
	flagArtPkgFilter                        string
	flagArtPkgPageSize                      int64
	flagArtPkgConfigFile                    string
	flagArtPkgMask                          string
	flagArtPkgAsync                         bool
	artifactsPackagesDeleteCmd              *cobra.Command
	artifactsPackagesDescribeCmd            *cobra.Command
	artifactsPackagesListCmd                *cobra.Command
	artifactsPackagesUpdateCmd              *cobra.Command
	flagArtPrintLocation                    string
	flagArtPrintRepository                  string
	flagArtPrintJSONKey                     string
	flagArtPrintNpmScope                    string
	artifactsPrintSettingsCmd               *cobra.Command
	artifactsPrintSettingsGradleCmd         *cobra.Command
	artifactsPrintSettingsMvnCmd            *cobra.Command
	artifactsPrintSettingsNpmCmd            *cobra.Command
	artifactsPrintSettingsPythonCmd         *cobra.Command
	artifactsProjectsCmd                    *cobra.Command
	flagArtProjectsFormat                   string
	flagArtProjectsConfigFile               string
	flagArtProjectsMask                     string
	artifactsProjectsDescribeCmd            *cobra.Command
	artifactsProjectsUpdateCmd              *cobra.Command
	artifactsRepositoriesCmd                *cobra.Command
	flagArtRepoLocation                     string
	flagArtRepoFormat                       string
	flagArtRepoFilter                       string
	flagArtRepoPageSize                     int64
	flagArtRepoConfigFile                   string
	flagArtRepoMask                         string
	flagArtRepoAsync                        bool
	flagArtRepoAllConds                     bool
	flagArtRepoPolicyNames                  []string
	flagArtRepoIamMember                    string
	flagArtRepoIamRole                      string
	flagArtRepoIamCondExpr                  string
	flagArtRepoIamCondTitle                 string
	flagArtRepoIamCondDesc                  string
	// --mode passthrough, mirroring gcloud-python's CLI vocabulary.
	// NONE / STANDARD-REPOSITORY / VIRTUAL-REPOSITORY / REMOTE-REPOSITORY /
	// CONNECTOR-REPOSITORY (the last one is REMOTE_REPOSITORY on the wire).
	flagArtRepoMode                               string
	artifactsRepositoriesCreateCmd                *cobra.Command
	artifactsRepositoriesDeleteCmd                *cobra.Command
	artifactsRepositoriesDescribeCmd              *cobra.Command
	artifactsRepositoriesListCmd                  *cobra.Command
	artifactsRepositoriesUpdateCmd                *cobra.Command
	artifactsRepositoriesGetIamCmd                *cobra.Command
	artifactsRepositoriesSetIamCmd                *cobra.Command
	artifactsRepositoriesAddIamCmd                *cobra.Command
	artifactsRepositoriesRemoveIamCmd             *cobra.Command
	artifactsRepositoriesListCleanupPoliciesCmd   *cobra.Command
	artifactsRepositoriesSetCleanupPoliciesCmd    *cobra.Command
	artifactsRepositoriesDeleteCleanupPoliciesCmd *cobra.Command
	artifactsRulesCmd                             *cobra.Command
	flagArtRuleLocation                           string
	flagArtRuleRepository                         string
	flagArtRuleFormat                             string
	flagArtRuleFilter                             string
	flagArtRuleConfigFile                         string
	flagArtRuleMask                               string
	artifactsRulesCreateCmd                       *cobra.Command
	artifactsRulesDeleteCmd                       *cobra.Command
	artifactsRulesDescribeCmd                     *cobra.Command
	artifactsRulesListCmd                         *cobra.Command
	artifactsRulesUpdateCmd                       *cobra.Command
	artifactsSbomCmd                              *cobra.Command
	flagArtSbomLocation                           string
	flagArtSbomRepository                         string
	flagArtSbomFormat                             string
	flagArtSbomDestination                        string
	flagArtSbomConfigFile                         string
	artifactsSbomExportCmd                        *cobra.Command
	artifactsSbomListCmd                          *cobra.Command
	artifactsSbomLoadCmd                          *cobra.Command
	artifactsSettingsCmd                          *cobra.Command
	flagArtSettingsLocation                       string
	flagArtSettingsFormat                         string
	artifactsSettingsDescribeCmd                  *cobra.Command
	artifactsSettingsEnableRedirCmd               *cobra.Command
	artifactsSettingsDisableRedirCmd              *cobra.Command
	artifactsTagsCmd                              *cobra.Command
	flagArtTagLocation                            string
	flagArtTagRepository                          string
	flagArtTagPackage                             string
	flagArtTagVersion                             string
	flagArtTagFormat                              string
	flagArtTagFilter                              string
	flagArtTagDestination                         string
	artifactsTagsCreateCmd                        *cobra.Command
	artifactsTagsDeleteCmd                        *cobra.Command
	artifactsTagsExportCmd                        *cobra.Command
	artifactsTagsListCmd                          *cobra.Command
	artifactsTagsUpdateCmd                        *cobra.Command
	artifactsVersionsCmd                          *cobra.Command
	flagArtVerLocation                            string
	flagArtVerRepository                          string
	flagArtVerPackage                             string
	flagArtVerFormat                              string
	flagArtVerFilter                              string
	flagArtVerConfigFile                          string
	flagArtVerMask                                string
	flagArtVerDestination                         string
	flagArtVerAsync                               bool
	artifactsVersionsDeleteCmd                    *cobra.Command
	artifactsVersionsDescribeCmd                  *cobra.Command
	artifactsVersionsExportCmd                    *cobra.Command
	artifactsVersionsListCmd                      *cobra.Command
	artifactsVersionsUpdateCmd                    *cobra.Command
	artifactsVpcscConfigCmd                       *cobra.Command
	flagArtVpcLocation                            string
	flagArtVpcFormat                              string
	artifactsVpcscConfigDescribeCmd               *cobra.Command
	artifactsVpcscConfigAllowCmd                  *cobra.Command
	artifactsVpcscConfigDenyCmd                   *cobra.Command
	artifactsVulnerabilitiesCmd                   *cobra.Command
	flagArtVulnLocation                           string
	flagArtVulnRepository                         string
	flagArtVulnImage                              string
	flagArtVulnFormat                             string
	flagArtVulnConfigFile                         string
	artifactsVulnerabilitiesListCmd               *cobra.Command
	artifactsVulnerabilitiesLoadVexCmd            *cobra.Command
	artifactsYumCmd                               *cobra.Command
	artifactsYumImportCmd                         *cobra.Command
	artifactsYumUploadCmd                         *cobra.Command
	flagArtYumLocation                            string
	flagArtYumGcsSrc                              []string
	flagArtYumSource                              string
	flagArtYumAsync                               bool
}

// newArtifactsCmd returns the "artifacts" command group for inv.
func newArtifactsCmd(inv *invocation) *cobra.Command {
	return newArtifactsCmds(inv).artifactsCmd
}

// newArtifactsCmds builds the "artifacts" command group for inv and returns its state.
func newArtifactsCmds(inv *invocation) *artifactsCmds {
	s := &artifactsCmds{inv: inv}
	s.artifactsCmd = &cobra.Command{
		Use:   "artifacts",
		Short: "Manage Artifact Registry",
	}
	s.artifactsDockerCmd = &cobra.Command{
		Use:   "docker",
		Short: "Manage Docker resources",
	}
	s.artifactsDockerImagesCmd = &cobra.Command{
		Use:   "images",
		Short: "Manage Docker images",
	}
	s.artifactsScanCmd = &cobra.Command{
		Use:   "scan IMAGE",
		Short: "Scan a Docker image for vulnerabilities",
		Long: `Scan a container image for known vulnerabilities using On-Demand Scanning.
Example:
  gcloud artifacts docker images scan us-docker.pkg.dev/my-project/repo/image:tag --location=us`,
		Args: cobra.ExactArgs(1),
		RunE: s.runArtifactsScan,
	}
	s.artifactsListVulnerabilitiesCmd = &cobra.Command{
		Use:   "list-vulnerabilities SCAN_RESOURCE",
		Short: "List vulnerabilities from a scan",
		Long: `List vulnerabilities found by an on-demand scan.
Example:
  gcloud artifacts docker images list-vulnerabilities projects/my-project/locations/us/scans/SCAN_ID`,
		Args: cobra.ExactArgs(1),
		RunE: s.runArtifactsListVulnerabilities,
	}
	s.artifactsDockerImagesListCmd = &cobra.Command{
		Use:   "list REPOSITORY",
		Short: "List Docker images in a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsDockerImagesList,
	}
	s.artifactsDockerImagesDescribeCmd = &cobra.Command{
		Use:   "describe IMAGE",
		Short: "Describe a Docker image",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsDockerImagesDescribe,
	}
	s.artifactsDockerImagesDeleteCmd = &cobra.Command{
		Use:   "delete IMAGE",
		Short: "Delete a Docker image",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsDockerImagesDelete,
	}
	s.artifactsAptCmd = &cobra.Command{
		Use:   "apt",
		Short: "Manage Artifact Registry Debian (APT) packages",
	}
	s.artifactsAptImportCmd = &cobra.Command{
		Use:   "import",
		Short: "Import one or more Debian (APT) packages into a repository",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsAptImport,
	}
	s.artifactsAptUploadCmd = &cobra.Command{
		Use:   "upload",
		Short: "Upload a Debian (APT) package to a repository",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsAptUpload,
	}
	s.artifactsAttachmentsCmd = &cobra.Command{
		Use:   "attachments",
		Short: "Manage Artifact Registry attachments",
	}
	s.artifactsAttachmentsCreateCmd = &cobra.Command{
		Use:   "create ATTACHMENT",
		Short: "Create an attachment",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsAttachmentsCreate,
	}
	s.artifactsAttachmentsDeleteCmd = &cobra.Command{
		Use:   "delete ATTACHMENT",
		Short: "Delete an attachment",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsAttachmentsDelete,
	}
	s.artifactsAttachmentsDescribeCmd = &cobra.Command{
		Use:   "describe ATTACHMENT",
		Short: "Describe an attachment",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsAttachmentsDescribe,
	}
	s.artifactsAttachmentsDownloadCmd = &cobra.Command{
		Use:   "download ATTACHMENT",
		Short: "Download all files referenced by an attachment",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsAttachmentsDownload,
	}
	s.artifactsAttachmentsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List attachments in a repository",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsAttachmentsList,
	}
	s.artifactsFilesCmd = &cobra.Command{
		Use:   "files",
		Short: "Manage individual files in an Artifact Registry repository",
	}
	s.artifactsFilesDeleteCmd = &cobra.Command{
		Use:   "delete FILE",
		Short: "Delete a file from a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsFilesDelete,
	}
	s.artifactsFilesDescribeCmd = &cobra.Command{
		Use:   "describe FILE",
		Short: "Describe a file in a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsFilesDescribe,
	}
	s.artifactsFilesDownloadCmd = &cobra.Command{
		Use:   "download FILE",
		Short: "Download a file from a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsFilesDownload,
	}
	s.artifactsFilesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List files in a repository",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsFilesList,
	}
	s.artifactsFilesUpdateCmd = &cobra.Command{
		Use:   "update FILE",
		Short: "Update a file in a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsFilesUpdate,
	}
	s.artifactsFilesUploadCmd = &cobra.Command{
		Use:   "upload",
		Short: "Upload files to Artifact Registry",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsFilesUpload,
	}
	s.artifactsGenericCmd = &cobra.Command{
		Use:   "generic",
		Short: "Manage generic Artifact Registry artifacts",
	}
	s.artifactsGenericDownloadCmd = &cobra.Command{
		Use:   "download PACKAGE VERSION FILE_NAME",
		Short: "Download a file from a generic artifact",
		Args:  cobra.ExactArgs(3),
		RunE:  s.runArtifactsGenericDownload,
	}
	s.artifactsGenericUploadCmd = &cobra.Command{
		Use:   "upload",
		Short: "Upload a file to a generic artifact",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsGenericUpload,
	}
	s.artifactsGoCmd = &cobra.Command{
		Use:   "go",
		Short: "Manage Artifact Registry Go modules",
	}
	s.artifactsGoAuthCmd = &cobra.Command{
		Use:   "auth",
		Short: "Print credentials and environment for using a Go repository",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsGoAuth,
	}
	s.artifactsGoUploadCmd = &cobra.Command{
		Use:   "upload",
		Short: "Upload a Go module .zip to a repository",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsGoUpload,
	}
	s.artifactsImageStreamingCacheCmd = &cobra.Command{
		Use:   "image-streaming-cache",
		Short: "Manage Artifact Registry image streaming caches",
	}
	s.artifactsImageStreamingCacheCreateCmd = &cobra.Command{
		Use:   "create CACHE",
		Short: "Create an image streaming cache",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsISCCreate,
	}
	s.artifactsImageStreamingCacheDeleteCmd = &cobra.Command{
		Use:   "delete CACHE",
		Short: "Delete an image streaming cache",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsISCDelete,
	}
	s.artifactsImageStreamingCacheDescribeCmd = &cobra.Command{
		Use:   "describe CACHE",
		Short: "Describe an image streaming cache",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsISCDescribe,
	}
	s.artifactsImageStreamingCacheListCmd = &cobra.Command{
		Use:   "list",
		Short: "List image streaming caches",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsISCList,
	}
	s.artifactsLocationsCmd = &cobra.Command{
		Use:   "locations",
		Short: "List regional metadata for Artifact Registry",
	}
	s.artifactsLocationsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Artifact Registry locations for a project",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsLocationsList,
	}
	s.artifactsOperationsCmd = &cobra.Command{
		Use:   "operations",
		Short: "Manage Artifact Registry long-running operations",
	}
	s.artifactsOperationsDescribeCmd = &cobra.Command{
		Use:   "describe OPERATION",
		Short: "Describe an Artifact Registry operation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsOperationsDescribe,
	}
	s.artifactsPackagesCmd = &cobra.Command{
		Use:   "packages",
		Short: "Manage Artifact Registry packages",
	}
	s.artifactsPackagesDeleteCmd = &cobra.Command{
		Use:   "delete PACKAGE",
		Short: "Delete a package",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsPackagesDelete,
	}
	s.artifactsPackagesDescribeCmd = &cobra.Command{
		Use:   "describe PACKAGE",
		Short: "Describe a package",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsPackagesDescribe,
	}
	s.artifactsPackagesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List packages",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsPackagesList,
	}
	s.artifactsPackagesUpdateCmd = &cobra.Command{
		Use:   "update PACKAGE",
		Short: "Update a package",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsPackagesUpdate,
	}
	s.artifactsPrintSettingsCmd = &cobra.Command{
		Use:   "print-settings",
		Short: "Print snippets to add to native tools settings files",
		Long: `The snippets provide a credentials placeholder and configurations to allow
native tools to interact with Artifact Registry repositories.`,
	}
	s.artifactsPrintSettingsGradleCmd = &cobra.Command{
		Use:   "gradle",
		Short: "Print a snippet to add a repository to the Gradle build.gradle file",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsPrintSettingsGradle,
	}
	s.artifactsPrintSettingsMvnCmd = &cobra.Command{
		Use:   "mvn",
		Short: "Print a snippet to add a Maven repository to the pom.xml file",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsPrintSettingsMvn,
	}
	s.artifactsPrintSettingsNpmCmd = &cobra.Command{
		Use:   "npm",
		Short: "Print credential settings to add to the .npmrc file",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsPrintSettingsNpm,
	}
	s.artifactsPrintSettingsPythonCmd = &cobra.Command{
		Use:   "python",
		Short: "Print credential settings to add to the .pypirc and pip.conf files",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsPrintSettingsPython,
	}
	s.artifactsProjectsCmd = &cobra.Command{
		Use:   "projects",
		Short: "Manage per-project Artifact Registry settings",
	}
	s.artifactsProjectsDescribeCmd = &cobra.Command{
		Use:   "describe",
		Short: "Describe the Artifact Registry project settings",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsProjectsDescribe,
	}
	s.artifactsProjectsUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the Artifact Registry project settings",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsProjectsUpdate,
	}
	s.artifactsRepositoriesCmd = &cobra.Command{
		Use:   "repositories",
		Short: "Manage Artifact Registry repositories",
	}
	s.artifactsRepositoriesCreateCmd = &cobra.Command{
		Use:   "create REPOSITORY",
		Short: "Create a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsRepositoriesCreate,
	}
	s.artifactsRepositoriesDeleteCmd = &cobra.Command{
		Use:   "delete REPOSITORY",
		Short: "Delete a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsRepositoriesDelete,
	}
	s.artifactsRepositoriesDescribeCmd = &cobra.Command{
		Use:   "describe REPOSITORY",
		Short: "Describe a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsRepositoriesDescribe,
	}
	s.artifactsRepositoriesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List repositories in a location",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsRepositoriesList,
	}
	s.artifactsRepositoriesUpdateCmd = &cobra.Command{
		Use:   "update REPOSITORY",
		Short: "Update a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsRepositoriesUpdate,
	}
	s.artifactsRepositoriesGetIamCmd = &cobra.Command{
		Use:   "get-iam-policy REPOSITORY",
		Short: "Get the IAM policy for a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsRepositoriesGetIam,
	}
	s.artifactsRepositoriesSetIamCmd = &cobra.Command{
		Use:   "set-iam-policy REPOSITORY POLICY_FILE",
		Short: "Set the IAM policy for a repository",
		Args:  cobra.ExactArgs(2),
		RunE:  s.runArtifactsRepositoriesSetIam,
	}
	s.artifactsRepositoriesAddIamCmd = &cobra.Command{
		Use:   "add-iam-policy-binding REPOSITORY",
		Short: "Add an IAM policy binding on a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsRepositoriesAddIam,
	}
	s.artifactsRepositoriesRemoveIamCmd = &cobra.Command{
		Use:   "remove-iam-policy-binding REPOSITORY",
		Short: "Remove an IAM policy binding from a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsRepositoriesRemoveIam,
	}
	s.artifactsRepositoriesListCleanupPoliciesCmd = &cobra.Command{
		Use:   "list-cleanup-policies REPOSITORY",
		Short: "List cleanup policies on a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsRepositoriesListCleanup,
	}
	s.artifactsRepositoriesSetCleanupPoliciesCmd = &cobra.Command{
		Use:   "set-cleanup-policies REPOSITORY",
		Short: "Replace cleanup policies on a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsRepositoriesSetCleanup,
	}
	s.artifactsRepositoriesDeleteCleanupPoliciesCmd = &cobra.Command{
		Use:   "delete-cleanup-policies REPOSITORY",
		Short: "Delete cleanup policies from a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsRepositoriesDeleteCleanup,
	}
	s.artifactsRulesCmd = &cobra.Command{
		Use:   "rules",
		Short: "Manage Artifact Registry cleanup rules",
	}
	s.artifactsRulesCreateCmd = &cobra.Command{
		Use:   "create RULE",
		Short: "Create a cleanup rule",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsRulesCreate,
	}
	s.artifactsRulesDeleteCmd = &cobra.Command{
		Use:   "delete RULE",
		Short: "Delete a cleanup rule",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsRulesDelete,
	}
	s.artifactsRulesDescribeCmd = &cobra.Command{
		Use:   "describe RULE",
		Short: "Describe a cleanup rule",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsRulesDescribe,
	}
	s.artifactsRulesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List cleanup rules",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsRulesList,
	}
	s.artifactsRulesUpdateCmd = &cobra.Command{
		Use:   "update RULE",
		Short: "Update a cleanup rule",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsRulesUpdate,
	}
	s.artifactsSbomCmd = &cobra.Command{
		Use:   "sbom",
		Short: "Manage Software Bills of Materials (SBOMs)",
	}
	s.artifactsSbomExportCmd = &cobra.Command{
		Use:   "export IMAGE_URI",
		Short: "Export the SBOM for a Docker image",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsSbomExport,
	}
	s.artifactsSbomListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Docker images in a repository, noting SBOM availability",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsSbomList,
	}
	s.artifactsSbomLoadCmd = &cobra.Command{
		Use:   "load IMAGE_URI",
		Short: "Attach an SBOM statement to a Docker image",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsSbomLoad,
	}
	s.artifactsSettingsCmd = &cobra.Command{
		Use:   "settings",
		Short: "Manage Artifact Registry per-location settings",
	}
	s.artifactsSettingsDescribeCmd = &cobra.Command{
		Use:   "describe",
		Short: "Describe the settings for a location",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsSettingsDescribe,
	}
	s.artifactsSettingsEnableRedirCmd = &cobra.Command{
		Use:   "enable-upgrade-redirection",
		Short: "Enable redirection of legacy container registry traffic to Artifact Registry",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsSettingsEnableRedir,
	}
	s.artifactsSettingsDisableRedirCmd = &cobra.Command{
		Use:   "disable-upgrade-redirection",
		Short: "Disable redirection of legacy container registry traffic",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsSettingsDisableRedir,
	}
	s.artifactsTagsCmd = &cobra.Command{
		Use:   "tags",
		Short: "Manage Artifact Registry package tags",
	}
	s.artifactsTagsCreateCmd = &cobra.Command{
		Use:   "create TAG",
		Short: "Create a tag pointing to a version",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsTagsCreate,
	}
	s.artifactsTagsDeleteCmd = &cobra.Command{
		Use:   "delete TAG",
		Short: "Delete a tag",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsTagsDelete,
	}
	s.artifactsTagsExportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export tags for a package to a JSON file",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsTagsExport,
	}
	s.artifactsTagsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List tags for a package",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsTagsList,
	}
	s.artifactsTagsUpdateCmd = &cobra.Command{
		Use:   "update TAG",
		Short: "Move a tag to a different version",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsTagsUpdate,
	}
	s.artifactsVersionsCmd = &cobra.Command{
		Use:   "versions",
		Short: "Manage Artifact Registry package versions",
	}
	s.artifactsVersionsDeleteCmd = &cobra.Command{
		Use:   "delete VERSION",
		Short: "Delete a version",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsVersionsDelete,
	}
	s.artifactsVersionsDescribeCmd = &cobra.Command{
		Use:   "describe VERSION",
		Short: "Describe a version",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsVersionsDescribe,
	}
	s.artifactsVersionsExportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export versions for a package to a JSON file",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsVersionsExport,
	}
	s.artifactsVersionsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List versions of a package",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsVersionsList,
	}
	s.artifactsVersionsUpdateCmd = &cobra.Command{
		Use:   "update VERSION",
		Short: "Update a version",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runArtifactsVersionsUpdate,
	}
	s.artifactsVpcscConfigCmd = &cobra.Command{
		Use:   "vpcsc-config",
		Short: "Manage the VPC Service Controls config for Artifact Registry",
	}
	s.artifactsVpcscConfigDescribeCmd = &cobra.Command{
		Use:   "describe",
		Short: "Describe the VPC-SC config for a location",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsVpcscConfigDescribe,
	}
	s.artifactsVpcscConfigAllowCmd = &cobra.Command{
		Use:   "allow",
		Short: "Allow Artifact Registry to serve requests inside a VPC-SC perimeter",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, args []string) error { return s.artSetVpcscPolicy("ALLOW") },
	}
	s.artifactsVpcscConfigDenyCmd = &cobra.Command{
		Use:   "deny",
		Short: "Deny Artifact Registry from serving requests inside a VPC-SC perimeter",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, args []string) error { return s.artSetVpcscPolicy("DENY") },
	}
	s.artifactsVulnerabilitiesCmd = &cobra.Command{
		Use:   "vulnerabilities",
		Short: "Manage Artifact Registry vulnerability reports",
	}
	s.artifactsVulnerabilitiesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List vulnerabilities for a Docker image",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsVulnerabilitiesList,
	}
	s.artifactsVulnerabilitiesLoadVexCmd = &cobra.Command{
		Use:   "load-vex",
		Short: "Load a VEX statement into Artifact Registry",
		Args:  cobra.NoArgs,
		RunE:  s.runArtifactsVulnerabilitiesLoadVex,
	}
	s.artifactsYumCmd = &cobra.Command{
		Use:   "yum",
		Short: "Manage Artifact Registry RPM packages",
	}
	s.artifactsYumImportCmd = &cobra.Command{
		Use:   "import [REPOSITORY]",
		Short: "Import one or more RPM packages into an artifact repository",
		Args:  cobra.MaximumNArgs(1),
		RunE:  s.runArtifactsYumImport,
	}
	s.artifactsYumUploadCmd = &cobra.Command{
		Use:   "upload [REPOSITORY]",
		Short: "Upload an RPM package to an artifact repository",
		Args:  cobra.MaximumNArgs(1),
		RunE:  s.runArtifactsYumUpload,
	}
	s.initArtifacts()
	s.initArtifactsApt()
	s.initArtifactsAttachments()
	s.initArtifactsFiles()
	s.initArtifactsFilesUpload()
	s.initArtifactsGeneric()
	s.initArtifactsGo()
	s.initArtifactsImageStreamingCache()
	s.initArtifactsLocations()
	s.initArtifactsOperations()
	s.initArtifactsPackages()
	s.initArtifactsPrintSettings()
	s.initArtifactsProjects()
	s.initArtifactsRepositories()
	s.initArtifactsRules()
	s.initArtifactsSbom()
	s.initArtifactsSettings()
	s.initArtifactsTags()
	s.initArtifactsVersions()
	s.initArtifactsVpcscConfig()
	s.initArtifactsVulnerabilities()
	s.initArtifactsYum()
	return s
}
