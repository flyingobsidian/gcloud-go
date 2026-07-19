package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub host-project-registrations (#1162) ---

var apihubHostProjectRegistrationsCmd = &cobra.Command{Use: "host-project-registrations", Short: "Manage API Hub host project registrations"}

var (
	flagApihubHprLocation   string
	flagApihubHprFormat     string
	flagApihubHprConfigFile string
	flagApihubHprPageSize   int64
)

var (
	apihubHprCreateCmd = &cobra.Command{
		Use: "create REGISTRATION", Short: "Create a host project registration",
		Args: cobra.ExactArgs(1), RunE: runApihubHprCreate,
	}
	apihubHprDescribeCmd = &cobra.Command{
		Use: "describe REGISTRATION", Short: "Describe a host project registration",
		Args: cobra.ExactArgs(1), RunE: runApihubHprDescribe,
	}
	apihubHprListCmd = &cobra.Command{
		Use: "list", Short: "List host project registrations in a location",
		Args: cobra.NoArgs, RunE: runApihubHprList,
	}
)

func init() {
	all := []*cobra.Command{apihubHprCreateCmd, apihubHprDescribeCmd, apihubHprListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagApihubHprLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagApihubHprFormat, "format", "", "Output format")
	}
	apihubHprCreateCmd.Flags().StringVar(&flagApihubHprConfigFile, "config-file", "", "YAML/JSON file with the HostProjectRegistration body (required)")
	_ = apihubHprCreateCmd.MarkFlagRequired("config-file")
	apihubHprListCmd.Flags().Int64Var(&flagApihubHprPageSize, "page-size", 0, "Maximum results per page")

	apihubHostProjectRegistrationsCmd.AddCommand(all...)
	apihubCmd.AddCommand(apihubHostProjectRegistrationsCmd)
}

func apihubHprName(id string) (string, error) {
	return apihubResource(flagApihubHprLocation, "hostProjectRegistrations", id)
}

func runApihubHprCreate(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubHprLocation)
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1HostProjectRegistration{}
	if err := loadYAMLOrJSONInto(flagApihubHprConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.HostProjectRegistrations.Create(parent, body).HostProjectRegistrationId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating host project registration: %w", err)
	}
	fmt.Printf("Created host project registration [%s].\n", args[0])
	return emitFormatted(got, flagApihubHprFormat)
}

func runApihubHprDescribe(cmd *cobra.Command, args []string) error {
	name, err := apihubHprName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.HostProjectRegistrations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing host project registration: %w", err)
	}
	return emitFormatted(got, flagApihubHprFormat)
}

func runApihubHprList(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubHprLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1HostProjectRegistration
	pageToken := ""
	for {
		call := svc.Projects.Locations.HostProjectRegistrations.List(parent).Context(ctx)
		if flagApihubHprPageSize > 0 {
			call = call.PageSize(flagApihubHprPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing host project registrations: %w", err)
		}
		all = append(all, resp.HostProjectRegistrations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagApihubHprFormat)
}
