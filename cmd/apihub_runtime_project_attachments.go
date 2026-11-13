package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub runtime-project-attachments (#1165) ---

func (g *apihubCmds) initApihubRuntimeProjectAttachments() {
	all := []*cobra.Command{
		g.apihubRPACreateCmd, g.apihubRPADeleteCmd, g.apihubRPADescribeCmd,
		g.apihubRPAListCmd, g.apihubRPALookupCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&g.flagAPRPALocation, "location", "",
			"Location that owns the runtime project attachments (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagAPRPAFormat, "format", "", "Output format")
	}
	g.apihubRPACreateCmd.Flags().StringVar(&g.flagAPRPAConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the RuntimeProjectAttachment body (required)")
	_ = g.apihubRPACreateCmd.MarkFlagRequired("config-file")
	g.apihubRPAListCmd.Flags().StringVar(&g.flagAPRPAFilter, "filter", "", "Server-side filter expression")
	g.apihubRPAListCmd.Flags().Int64Var(&g.flagAPRPAPageSize, "page-size", 0, "Maximum number of results per page")
	g.apihubRPALookupCmd.Flags().StringVar(&g.flagAPRPARuntimeProject, "runtime-project", "",
		"Runtime project id or number to look up the attachment for (required)")
	_ = g.apihubRPALookupCmd.MarkFlagRequired("runtime-project")

	g.apihubRPACmd.AddCommand(all...)
	g.apihubCmd.AddCommand(g.apihubRPACmd)
}

func (g *apihubCmds) apihubRPAParent() (string, error) {
	project, err := g.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, g.flagAPRPALocation), nil
}

func (g *apihubCmds) apihubRPAName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := g.apihubRPAParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/runtimeProjectAttachments/%s", parent, id), nil
}

func (g *apihubCmds) runAPRPACreate(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubRPAParent()
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1RuntimeProjectAttachment{}
	if err := loadYAMLOrJSONInto(g.flagAPRPAConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.RuntimeProjectAttachments.Create(parent, body).
		RuntimeProjectAttachmentId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating runtime project attachment: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAPRPAFormat)
}

func (g *apihubCmds) runAPRPADelete(cmd *cobra.Command, args []string) error {
	name, err := g.apihubRPAName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.RuntimeProjectAttachments.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting runtime project attachment: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted runtime project attachment [%s].\n", args[0])
	return nil
}

func (g *apihubCmds) runAPRPADescribe(cmd *cobra.Command, args []string) error {
	name, err := g.apihubRPAName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.RuntimeProjectAttachments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing runtime project attachment: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAPRPAFormat)
}

func (g *apihubCmds) runAPRPAList(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubRPAParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1RuntimeProjectAttachment
	pageToken := ""
	for {
		call := svc.Projects.Locations.RuntimeProjectAttachments.List(parent).Context(ctx)
		if g.flagAPRPAFilter != "" {
			call = call.Filter(g.flagAPRPAFilter)
		}
		if g.flagAPRPAPageSize > 0 {
			call = call.PageSize(g.flagAPRPAPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing runtime project attachments: %w", err)
		}
		all = append(all, resp.RuntimeProjectAttachments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagAPRPAFormat)
}

func (g *apihubCmds) runAPRPALookup(cmd *cobra.Command, args []string) error {
	// The Lookup RPC takes name = projects/{runtime-project}/locations/{location}
	// and searches across regions to find the attachment for the runtime project.
	name := fmt.Sprintf("projects/%s/locations/%s", g.flagAPRPARuntimeProject, g.flagAPRPALocation)
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.LookupRuntimeProjectAttachment(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up runtime project attachment: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAPRPAFormat)
}
