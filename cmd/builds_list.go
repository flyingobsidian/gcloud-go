package cmd

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudbuild "google.golang.org/api/cloudbuild/v1"
)

// --- gcloud builds list (#2006) ---

// buildsDefaultRegion is gcloud's cloudbuild_util.DEFAULT_REGION.
const buildsDefaultRegion = "global"

// buildsOngoingFilter is the server-side filter gcloud sends for --ongoing.
const buildsOngoingFilter = `status="WORKING" OR status="QUEUED"`

var (
	flagBuildsListRegion   string
	flagBuildsListOngoing  bool
	flagBuildsListLimit    int64
	flagBuildsListPageSize int64
)

func init() {
	list := &cobra.Command{
		Use:   "list",
		Short: "List builds",
		Args:  cobra.NoArgs,
		RunE:  runBuildsList,
	}
	list.Flags().StringVar(&flagBuildsListRegion, "region", "",
		"The region of the Cloud Build Service to use; default the builds/region property, else global")
	list.Flags().BoolVar(&flagBuildsListOngoing, "ongoing", false, "Only list builds that are currently QUEUED or WORKING")
	list.Flags().Int64Var(&flagBuildsListLimit, "limit", 50, "Maximum number of builds to list (0 = no limit)")
	list.Flags().Int64Var(&flagBuildsListPageSize, "page-size", 20, "Number of builds fetched per request")
	buildsCmd.AddCommand(list)
}

// buildsListRegion returns --region, else the builds/region property, else
// gcloud's default region.
func buildsListRegion() (string, error) {
	if flagBuildsListRegion != "" {
		return flagBuildsListRegion, nil
	}
	region, err := config.Property("builds", "region")
	if err != nil {
		return "", fmt.Errorf("reading builds/region: %w", err)
	}
	if region != "" {
		return region, nil
	}
	return buildsDefaultRegion, nil
}

// buildsPager fetches one page of builds for the given page token.
type buildsPager func(ctx context.Context, pageToken string) (*cloudbuild.ListBuildsResponse, error)

// listBuilds pages through builds until they run out or limit builds have
// been collected (limit <= 0 means no limit). It never returns a nil slice.
func listBuilds(ctx context.Context, page buildsPager, limit int64) ([]*cloudbuild.Build, error) {
	all := []*cloudbuild.Build{}
	pageToken := ""
	for {
		resp, err := page(ctx, pageToken)
		if err != nil {
			return nil, fmt.Errorf("listing builds: %w", err)
		}
		all = append(all, resp.Builds...)
		if limit > 0 && int64(len(all)) >= limit {
			return all[:limit], nil
		}
		if resp.NextPageToken == "" {
			return all, nil
		}
		pageToken = resp.NextPageToken
	}
}

func runBuildsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	region, err := buildsListRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := buildsLocationParent(project, region)
	page := func(ctx context.Context, pageToken string) (*cloudbuild.ListBuildsResponse, error) {
		call := svc.Projects.Locations.Builds.List(parent).Context(ctx)
		if flagBuildsListPageSize > 0 {
			call = call.PageSize(flagBuildsListPageSize)
		}
		if flagBuildsListOngoing {
			call = call.Filter(buildsOngoingFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		return call.Do()
	}
	builds, err := listBuilds(ctx, page, flagBuildsListLimit)
	if err != nil {
		return err
	}
	return emitBuildsList(os.Stdout, os.Stderr, builds, flagFormat, time.Now(), time.Local)
}

// buildsListFormat is gcloud's default `builds list` table, over the rows
// built by buildsListTableRow.
const buildsListFormat = "table(id,createTime,duration,source,images,status)"

// emitBuildsList prints builds with format, or gcloud's default table
// ("Listed 0 items." on errW when empty). now and loc stand in for the
// clock and local time zone gcloud uses for running builds and dates.
func emitBuildsList(w, errW io.Writer, builds []*cloudbuild.Build, format string, now time.Time, loc *time.Location) error {
	if format != "" {
		return emitFormattedTo(w, builds, format)
	}
	if len(builds) == 0 {
		_, err := fmt.Fprintln(errW, "Listed 0 items.")
		return err
	}
	rows := make([]map[string]any, 0, len(builds))
	for _, b := range builds {
		rows = append(rows, buildsListTableRow(b, now, loc))
	}
	return emitFormattedTo(w, rows, buildsListFormat)
}

// buildsListTableRow renders b as gcloud's default table columns.
func buildsListTableRow(b *cloudbuild.Build, now time.Time, loc *time.Location) map[string]any {
	createTime := "-"
	if t, err := time.Parse(time.RFC3339Nano, b.CreateTime); err == nil {
		createTime = t.In(loc).Format("2006-01-02T15:04:05-07:00")
	}
	return map[string]any{
		"id":         b.Id,
		"createTime": createTime,
		"duration":   buildsDuration(b.StartTime, b.FinishTime, now),
		"source":     buildsSource(b.Source),
		"images":     buildsImages(b.Results),
		"status":     b.Status,
	}
}

// buildsDuration is gcloud's
// duration(start=startTime,end=finishTime,precision=0,calendar=false,undefined="  -").slice(2:):
// the ISO 8601 duration from start to end (or now) without its "PT" prefix,
// or "-" when start is unset.
func buildsDuration(start, end string, now time.Time) string {
	startTime, err := time.Parse(time.RFC3339Nano, start)
	if err != nil {
		return "-"
	}
	endTime := now
	if end != "" {
		if endTime, err = time.Parse(time.RFC3339Nano, end); err != nil {
			return "-"
		}
	}
	d := endTime.Sub(startTime)
	var b strings.Builder
	if d < 0 {
		b.WriteString("-")
		d = -d
	}
	b.WriteString("PT")
	hours := int64(d / time.Hour)
	minutes := int64(d % time.Hour / time.Minute)
	seconds := math.RoundToEven(float64(d%time.Minute) / float64(time.Second))
	if hours > 0 {
		b.WriteString(strconv.FormatInt(hours, 10) + "H")
	}
	if minutes > 0 {
		b.WriteString(strconv.FormatInt(minutes, 10) + "M")
	}
	if seconds > 0 {
		b.WriteString(strconv.FormatFloat(seconds, 'f', 0, 64) + "S")
	}
	s := strings.TrimSuffix(b.String(), "T")
	if strings.HasSuffix(s, "P") {
		s += "0"
	}
	if len(s) <= 2 {
		return ""
	}
	return s[2:]
}

// buildsSource is gcloud's build_source(undefined="-") transform.
func buildsSource(src *cloudbuild.Source) string {
	if src == nil {
		return "-"
	}
	if s := src.StorageSource; s != nil {
		if s.Bucket == "" || s.Object == "" {
			return "-"
		}
		return fmt.Sprintf("gs://%s/%s", s.Bucket, s.Object)
	}
	if r := src.RepoSource; r != nil {
		name := r.RepoName
		if name == "" {
			name = "default"
		}
		for _, ref := range []string{r.BranchName, r.TagName, r.CommitSha} {
			if ref != "" {
				return name + "@" + ref
			}
		}
	}
	if g := src.GitSource; g != nil && g.Url != "" {
		return g.Url + "@" + g.Revision
	}
	if m := src.StorageSourceManifest; m != nil {
		if m.Bucket == "" || m.Object == "" {
			return "-"
		}
		return fmt.Sprintf("gs://%s/%s", m.Bucket, m.Object)
	}
	if c := src.ConnectedRepository; c != nil && c.Repository != "" {
		return c.Repository + "@" + c.Revision
	}
	if d := src.DeveloperConnectConfig; d != nil && d.GitRepositoryLink != "" {
		return d.GitRepositoryLink + "@" + d.Revision
	}
	return "-"
}

// buildsImages is gcloud's build_images(undefined="-") transform: the first
// built image, with a count of any others.
func buildsImages(results *cloudbuild.Results) string {
	if results == nil || len(results.Images) == 0 {
		return "-"
	}
	first := results.Images[0].Name
	if first == "" {
		first = "-"
	}
	if n := len(results.Images); n > 1 {
		return fmt.Sprintf("%s (+%d more)", first, n-1)
	}
	return first
}
