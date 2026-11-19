package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	compute "google.golang.org/api/compute/v1"
)

// --- compute image-views and project-views (#1948) ---
//
// gcloud-python promoted `image-views describe` (585.0.0), `image-views list`
// (585.0.0, with --standard-images, --preview-images and --show-deprecated
// from 587.0.0) and `project-views describe` (584.0.0) to GA.

func (s *computeCmds) initComputeImageViews() {
	list := &cobra.Command{
		Use: "list", Short: "List Compute Engine image-view resources",
		Args: cobra.NoArgs, RunE: s.runImageViewsList,
	}
	list.Flags().StringVar(&s.flagImageViewsRegion, "region", "", "The region of the image-view resources to list")
	list.Flags().BoolVar(&s.flagImageViewsStandard, "standard-images", true,
		"List image views from public image projects (default; --no-standard-images lists only the project's own)")
	list.Flags().BoolVar(&s.flagImageViewsNoStandard, "no-standard-images", false,
		"List only image views from the current project")
	list.Flags().BoolVar(&s.flagImageViewsPreview, "preview-images", false, "Show image views that are in limited preview")
	list.Flags().BoolVar(&s.flagImageViewsPreview, "show-preview-images", false, "Show image views that are in limited preview")
	_ = list.Flags().MarkHidden("show-preview-images")
	list.Flags().BoolVar(&s.flagImageViewsShowDeprecated, "show-deprecated", false, "If provided, deprecated images are shown")
	list.Flags().BoolVar(&s.flagImageViewsURI, "uri", false, "Print a list of resource URIs instead of the default output")

	describe := &cobra.Command{
		Use: "describe IMAGE_VIEW", Short: "Describe a Compute Engine image-view resource",
		Args: cobra.ExactArgs(1), RunE: s.runImageViewsDescribe,
	}
	describe.Flags().StringVar(&s.flagImageViewsRegion, "region", "", "The region of the image-view resource")

	imageViews := &cobra.Command{Use: "image-views", Short: "Read Compute Engine image-view resources"}
	imageViews.AddCommand(list, describe)

	pvDescribe := &cobra.Command{
		Use: "describe", Short: "Describe the project view of a project in a region",
		Args: cobra.NoArgs, RunE: s.runProjectViewsDescribe,
	}
	pvDescribe.Flags().StringVar(&s.flagImageViewsRegion, "region", "", "Region of the project view")
	projectViews := &cobra.Command{Use: "project-views", Short: "Read Compute Engine project views"}
	projectViews.AddCommand(pvDescribe)

	s.computeCmd.AddCommand(imageViews, projectViews)
}

// imageViewsRegion returns --region, else compute/region, failing as
// gcloud-python's GetOrFail does when neither is set.
func (s *computeCmds) imageViewsRegion() (string, string, error) {
	project, region, err := s.inv.region()
	if err != nil {
		return "", "", err
	}
	if s.flagImageViewsRegion != "" {
		region = s.flagImageViewsRegion
	}
	if region == "" {
		return "", "", fmt.Errorf("The required property [region] is not currently set.\n" +
			"You may set it for your current workspace by running:\n\n" +
			"  $ gcloud config set compute/region VALUE\n\n" +
			"or it can be set temporarily by the environment variable [CLOUDSDK_COMPUTE_REGION]")
	}
	return project, region, nil
}

// imageViewDeprecated reports whether the image view's image is deprecated
// (a state other than ACTIVE), which list hides without --show-deprecated.
func imageViewDeprecated(v *compute.ImageView) bool {
	return v.Image != nil && computeImageDeprecated(v.Image)
}

// imageViewsLister supplies gcloud-python's image-views list table and --uri.
func imageViewsLister() computeLister[*compute.ImageView] {
	return computeLister[*compute.ImageView]{
		noun: "image views",
		// An image view has no self link of its own; --uri prints its image's.
		selfLink: func(v *compute.ImageView) string {
			if v.Image == nil {
				return ""
			}
			return v.Image.SelfLink
		},
		tableFormat: "table(name,project,family,deprecated,status)",
		tableRow: func(v *compute.ImageView) map[string]any {
			row := map[string]any{"name": "", "project": "", "family": "", "deprecated": "", "status": ""}
			if img := v.Image; img != nil {
				row["name"] = img.Name
				row["project"] = computeImageProject(img.SelfLink)
				row["family"] = img.Family
				row["status"] = img.Status
				if img.Deprecated != nil {
					row["deprecated"] = img.Deprecated.State
				}
			}
			return row
		},
	}
}

// listImageViews lists the image views of each project in region via list,
// dropping deprecated images unless showDeprecated.
func listImageViews(ctx context.Context, projects []string, showDeprecated bool,
	list func(ctx context.Context, project string) ([]*compute.ImageView, error)) ([]*compute.ImageView, error) {
	all := []*compute.ImageView{}
	for _, p := range projects {
		views, err := list(ctx, p)
		if err != nil {
			return nil, fmt.Errorf("listing image views in project %s: %w", p, err)
		}
		for _, v := range views {
			if showDeprecated || !imageViewDeprecated(v) {
				all = append(all, v)
			}
		}
	}
	return all, nil
}

func (s *computeCmds) runImageViewsList(cmd *cobra.Command, args []string) error {
	project, region, err := s.imageViewsRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}
	list := func(ctx context.Context, project string) ([]*compute.ImageView, error) {
		var out []*compute.ImageView
		err := svc.ImageViews.List(project, region).Pages(ctx, func(r *compute.ImageViewsListResponse) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	}
	standard := s.flagImageViewsStandard && !s.flagImageViewsNoStandard
	projects := computeImageProjects(project, standard, s.flagImageViewsPreview)
	views, err := listImageViews(ctx, projects, s.flagImageViewsShowDeprecated, list)
	if err != nil {
		return err
	}
	return imageViewsLister().emit(s.inv.stdout, s.inv.stderr, views, s.inv.flags.format, s.flagImageViewsURI)
}

func (s *computeCmds) runImageViewsDescribe(cmd *cobra.Command, args []string) error {
	project, region, err := s.imageViewsRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ImageViews.Get(project, region, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing image view: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.inv.flags.format)
}

func (s *computeCmds) runProjectViewsDescribe(cmd *cobra.Command, args []string) error {
	project, region, err := s.imageViewsRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ProjectViews.Get(project, region).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing project view: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.inv.flags.format)
}
