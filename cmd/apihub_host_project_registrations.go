package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub host-project-registrations (#1162) ---

func (g *apihubCmds) initApihubHostProjectRegistrations() {
	all := []*cobra.Command{g.apihubHprCreateCmd, g.apihubHprDescribeCmd, g.apihubHprListCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagApihubHprLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagApihubHprFormat, "format", "", "Output format")
	}
	g.apihubHprCreateCmd.Flags().StringVar(&g.flagApihubHprConfigFile, "config-file", "", "YAML/JSON file with the HostProjectRegistration body (required)")
	_ = g.apihubHprCreateCmd.MarkFlagRequired("config-file")
	g.apihubHprListCmd.Flags().Int64Var(&g.flagApihubHprPageSize, "page-size", 0, "Maximum results per page")

	g.apihubHostProjectRegistrationsCmd.AddCommand(all...)
	g.apihubCmd.AddCommand(g.apihubHostProjectRegistrationsCmd)
}

func (g *apihubCmds) apihubHprName(id string) (string, error) {
	return g.apihubResource(g.flagApihubHprLocation, "hostProjectRegistrations", id)
}

func (g *apihubCmds) runApihubHprCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubHprLocation)
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1HostProjectRegistration{}
	if err := loadYAMLOrJSONInto(g.flagApihubHprConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.HostProjectRegistrations.Create(parent, body).HostProjectRegistrationId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating host project registration: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Created host project registration [%s].\n", args[0])
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubHprFormat)
}

func (g *apihubCmds) runApihubHprDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.apihubHprName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.HostProjectRegistrations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing host project registration: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubHprFormat)
}

func (g *apihubCmds) runApihubHprList(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubHprLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1HostProjectRegistration
	pageToken := ""
	for {
		call := svc.Projects.Locations.HostProjectRegistrations.List(parent).Context(ctx)
		if g.flagApihubHprPageSize > 0 {
			call = call.PageSize(g.flagApihubHprPageSize)
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
	return emitFormattedTo(g.inv.stdout, all, g.flagApihubHprFormat)
}
