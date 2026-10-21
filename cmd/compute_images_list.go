package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	compute "google.golang.org/api/compute/v1"
)

// --- compute images list (#2000) ---

// computePublicImageProjects is gcloud's constants.PUBLIC_IMAGE_PROJECTS: the
// projects whose images --standard-images (the default) adds to the list.
var computePublicImageProjects = []string{
	"cos-cloud", "debian-cloud", "rocky-linux-cloud", "ubuntu-os-cloud",
	"almalinux-cloud", "centos-cloud", "fedora-coreos-cloud", "opensuse-cloud",
	"oracle-linux-cloud", "rhel-cloud", "rhel-sap-cloud", "rocky-linux-accelerator-cloud",
	"suse-cloud", "suse-sap-cloud", "ubuntu-os-accelerator-images", "ubuntu-os-pro-cloud",
	"windows-cloud", "windows-sql-cloud",
}

// computePreviewImageProjects is gcloud's constants.PREVIEW_IMAGE_PROJECTS,
// added by --preview-images. It is currently empty in gcloud too.
var computePreviewImageProjects []string

var (
	flagImagesListStandard      bool
	flagImagesListNoStandard    bool
	flagImagesListShowDeprecate bool
	flagImagesListPreview       bool
	flagImagesListURI           bool
)

func init() {
	list := &cobra.Command{
		Use:   "list",
		Short: "List Compute Engine images",
		Args:  cobra.NoArgs,
		RunE:  runComputeImagesList,
	}
	list.Flags().BoolVar(&flagImagesListStandard, "standard-images", true,
		"List images from public image projects (default)")
	list.Flags().BoolVar(&flagImagesListNoStandard, "no-standard-images", false,
		"List only the project's own images, not those of public image projects")
	list.Flags().BoolVar(&flagImagesListShowDeprecate, "show-deprecated", false, "If provided, deprecated images are shown")
	list.Flags().BoolVar(&flagImagesListPreview, "preview-images", false, "Show images that are in limited preview")
	list.Flags().BoolVar(&flagImagesListPreview, "show-preview-images", false, "Show images that are in limited preview")
	_ = list.Flags().MarkHidden("show-preview-images")
	list.Flags().BoolVar(&flagImagesListURI, "uri", false, "Print a list of resource URIs instead of the default output")

	group := &cobra.Command{Use: "images", Short: "Manage disk images"}
	group.AddCommand(list)
	registerStubCommand(group, "describe", "Not yet implemented")
	computeCmd.AddCommand(group)
}

// computeImageProjects returns the projects `images list` reads: the
// current project, then the public (and preview) image projects in sorted
// order.
func computeImageProjects(project string, standard, preview bool) []string {
	var extra []string
	if standard {
		extra = append(extra, computePublicImageProjects...)
	}
	if preview {
		extra = append(extra, computePreviewImageProjects...)
	}
	sort.Strings(extra)
	projects := []string{project}
	for _, p := range extra {
		if p != project {
			projects = append(projects, p)
		}
	}
	return projects
}

// computeImageDeprecated reports whether gcloud hides img without
// --show-deprecated: its deprecation state is set and not ACTIVE.
func computeImageDeprecated(img *compute.Image) bool {
	return img.Deprecated != nil && img.Deprecated.State != "" && img.Deprecated.State != "ACTIVE"
}

// imagesLister supplies the default table and --uri for `images list`; images
// are fetched by runComputeImagesList across several projects.
var imagesLister = computeLister[*compute.Image]{
	noun:        "images",
	selfLink:    func(img *compute.Image) string { return img.SelfLink },
	tableFormat: "table(name,project,family,deprecated,status)",
	tableRow: func(img *compute.Image) map[string]any {
		deprecated := ""
		if img.Deprecated != nil {
			deprecated = img.Deprecated.State
		}
		return map[string]any{
			"name":       img.Name,
			"project":    computeImageProject(img.SelfLink),
			"family":     img.Family,
			"deprecated": deprecated,
			"status":     img.Status,
		}
	},
}

// computeImageProject is gcloud's selfLink.scope(projects).segment(0): the
// project id in an image's self link.
func computeImageProject(selfLink string) string {
	_, rest, ok := strings.Cut(selfLink, "/projects/")
	if !ok {
		return ""
	}
	project, _, _ := strings.Cut(rest, "/")
	return project
}

// listComputeImages lists the images of each project via list, dropping
// deprecated images unless showDeprecated. It never returns a nil slice.
func listComputeImages(ctx context.Context, projects []string, showDeprecated bool,
	list func(ctx context.Context, project string) ([]*compute.Image, error)) ([]*compute.Image, error) {
	all := []*compute.Image{}
	for _, p := range projects {
		images, err := list(ctx, p)
		if err != nil {
			return nil, fmt.Errorf("listing images in project %s: %w", p, err)
		}
		for _, img := range images {
			if showDeprecated || !computeImageDeprecated(img) {
				all = append(all, img)
			}
		}
	}
	return all, nil
}

func runComputeImagesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}
	list := func(ctx context.Context, project string) ([]*compute.Image, error) {
		var out []*compute.Image
		err := svc.Images.List(project).Pages(ctx, func(r *compute.ImageList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	}
	standard := flagImagesListStandard && !flagImagesListNoStandard
	projects := computeImageProjects(project, standard, flagImagesListPreview)
	images, err := listComputeImages(ctx, projects, flagImagesListShowDeprecate, list)
	if err != nil {
		return err
	}
	return imagesLister.emit(os.Stdout, os.Stderr, images, flagFormat, flagImagesListURI)
}
