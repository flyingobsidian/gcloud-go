package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	cloudlocationfinder "google.golang.org/api/cloudlocationfinder/v1"
)

// --- gcloud cloudlocationfinder cloud-locations (#968) ---

func (s *cloudLocationFinderCmds) initCloudlocationfinder() {
	all := []*cobra.Command{s.cloudLocationsDescribeCmd, s.cloudLocationsListCmd, s.cloudLocationsSearchCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagCLFLocation, "location", "global", "Location that owns the cloud locations collection (e.g. \"global\")")
		c.Flags().StringVar(&s.flagCLFFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.cloudLocationsListCmd, s.cloudLocationsSearchCmd} {
		c.Flags().Int64Var(&s.flagCLFPageSize, "page-size", 0, "Maximum number of results per page")
	}
	s.cloudLocationsListCmd.Flags().StringVar(&s.flagCLFFilter, "filter", "", "Filter expression (e.g. cloud_location_type=CLOUD_LOCATION_TYPE_REGION)")
	s.cloudLocationsSearchCmd.Flags().StringVar(&s.flagCLFQuery, "query", "", "Search query string")
	s.cloudLocationsSearchCmd.Flags().StringVar(&s.flagCLFSourceCloudLocation, "source-cloud-location", "", "Source cloud location to search from (required)")
	_ = s.cloudLocationsSearchCmd.MarkFlagRequired("source-cloud-location")

	s.cloudLocationsCmd.AddCommand(all...)
	s.cloudLocationFinderCmd.AddCommand(s.cloudLocationsCmd)

}

func (s *cloudLocationFinderCmds) clfParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	loc := s.flagCLFLocation
	if loc == "" {
		loc = "global"
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, loc), nil
}

func (s *cloudLocationFinderCmds) runCLFDescribe(cmd *cobra.Command, args []string) error {
	parent, err := s.clfParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudLocationFinder(ctx)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/cloudLocations/%s", parent, args[0])
	loc, err := svc.Projects.Locations.CloudLocations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cloud location: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, loc, s.flagCLFFormat)
}

func (s *cloudLocationFinderCmds) runCLFList(cmd *cobra.Command, args []string) error {
	parent, err := s.clfParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudLocationFinder(ctx)
	if err != nil {
		return err
	}
	var all []*cloudlocationfinder.CloudLocation
	pageToken := ""
	for {
		call := svc.Projects.Locations.CloudLocations.List(parent).Context(ctx)
		if s.flagCLFFilter != "" {
			call = call.Filter(s.flagCLFFilter)
		}
		if s.flagCLFPageSize > 0 {
			call = call.PageSize(s.flagCLFPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing cloud locations: %w", err)
		}
		all = append(all, resp.CloudLocations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagCLFFormat)
}

func (s *cloudLocationFinderCmds) runCLFSearch(cmd *cobra.Command, args []string) error {
	parent, err := s.clfParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudLocationFinder(ctx)
	if err != nil {
		return err
	}
	var all []*cloudlocationfinder.CloudLocation
	pageToken := ""
	for {
		call := svc.Projects.Locations.CloudLocations.Search(parent).
			SourceCloudLocation(s.flagCLFSourceCloudLocation).Context(ctx)
		if s.flagCLFQuery != "" {
			call = call.Query(s.flagCLFQuery)
		}
		if s.flagCLFPageSize > 0 {
			call = call.PageSize(s.flagCLFPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("searching cloud locations: %w", err)
		}
		all = append(all, resp.CloudLocations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagCLFFormat)
}

// cloudLocationFinderCmds holds the commands and flag values of one invocation's "cloudLocationFinder"
// command group, so invocations don't share state (#2070).
type cloudLocationFinderCmds struct {
	inv *invocation

	cloudLocationFinderCmd     *cobra.Command
	flagCLFLocation            string
	flagCLFFilter              string
	flagCLFQuery               string
	flagCLFSourceCloudLocation string
	flagCLFPageSize            int64
	flagCLFFormat              string
	cloudLocationsCmd          *cobra.Command
	cloudLocationsDescribeCmd  *cobra.Command
	cloudLocationsListCmd      *cobra.Command
	cloudLocationsSearchCmd    *cobra.Command
}

// newCloudLocationFinderCmd returns the "cloudLocationFinder" command group for inv.
func newCloudLocationFinderCmd(inv *invocation) *cobra.Command {
	return newCloudLocationFinderCmds(inv).cloudLocationFinderCmd
}

// newCloudLocationFinderCmds builds the "cloudLocationFinder" command group for inv and returns its state.
func newCloudLocationFinderCmds(inv *invocation) *cloudLocationFinderCmds {
	s := &cloudLocationFinderCmds{inv: inv}
	s.cloudLocationFinderCmd = &cobra.Command{
		Use:   "cloudlocationfinder",
		Short: "Manage Cloud Location Finder",
	}
	s.cloudLocationsCmd = &cobra.Command{
		Use:   "cloud-locations",
		Short: "Manage cloud locations",
	}
	s.cloudLocationsDescribeCmd = &cobra.Command{
		Use:   "describe CLOUD_LOCATION",
		Short: "Describe a cloud location",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runCLFDescribe,
	}
	s.cloudLocationsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List cloud locations",
		Args:  cobra.NoArgs,
		RunE:  s.runCLFList,
	}
	s.cloudLocationsSearchCmd = &cobra.Command{
		Use:   "search",
		Short: "Search for cloud locations from a source cloud location",
		Args:  cobra.NoArgs,
		RunE:  s.runCLFSearch,
	}
	s.initCloudlocationfinder()
	return s
}
