package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub api-hub-instances (#1167) ---

func (g *apihubCmds) initApihubApiHubInstances() {
	all := []*cobra.Command{g.apihubAHICreateCmd, g.apihubAHIDeleteCmd, g.apihubAHIDescribeCmd, g.apihubAHILookupCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagAHILocation, "location", "",
			"Location that owns the API Hub instance (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagAHIFormat, "format", "", "Output format")
	}
	g.apihubAHICreateCmd.Flags().StringVar(&g.flagAHIDescription, "description", "",
		"Description of the API Hub instance")
	g.apihubAHICreateCmd.Flags().StringSliceVar(&g.flagAHILabels, "labels", nil,
		"Labels as KEY=VALUE pairs")
	g.apihubAHICreateCmd.Flags().StringVar(&g.flagAHICmekKey, "config-cmek-key-name", "",
		"Fully qualified CMEK key (projects/*/locations/*/keyRings/*/cryptoKeys/*)")
	g.apihubAHICreateCmd.Flags().BoolVar(&g.flagAHIDisableSearch, "config-disable-search", false,
		"Disable search on the instance")
	g.apihubAHICreateCmd.Flags().StringVar(&g.flagAHIVertexLoc, "config-vertex-location", "",
		"Vertex AI location where the data store is stored")
	g.apihubAHICreateCmd.Flags().StringVar(&g.flagAHIEncryption, "config-encryption-type", "",
		"Encryption type: gmek or cmek")

	g.apihubAHICmd.AddCommand(all...)
	g.apihubCmd.AddCommand(g.apihubAHICmd)
}

func (g *apihubCmds) ahiParent() (string, error) {
	project, err := g.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, g.flagAHILocation), nil
}

func (g *apihubCmds) ahiName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := g.ahiParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/apiHubInstances/%s", parent, id), nil
}

// ahiEncryptionEnum maps the CLI-facing --config-encryption-type value to the
// API's enum representation.
func ahiEncryptionEnum(v string) (string, error) {
	switch strings.ToLower(v) {
	case "":
		return "", nil
	case "gmek":
		return "GMEK", nil
	case "cmek":
		return "CMEK", nil
	default:
		return "", fmt.Errorf("--config-encryption-type must be gmek or cmek (got %q)", v)
	}
}

func (g *apihubCmds) ahiLabelsFromFlag() map[string]string {
	if len(g.flagAHILabels) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, kv := range g.flagAHILabels {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		out[k] = v
	}
	return out
}

func (g *apihubCmds) runAHICreate(cmd *cobra.Command, args []string) error {
	parent, err := g.ahiParent()
	if err != nil {
		return err
	}
	enc, err := ahiEncryptionEnum(g.flagAHIEncryption)
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1ApiHubInstance{
		Description: g.flagAHIDescription,
		Labels:      g.ahiLabelsFromFlag(),
		Config: &apihub.GoogleCloudApihubV1Config{
			CmekKeyName:    g.flagAHICmekKey,
			DisableSearch:  g.flagAHIDisableSearch,
			EncryptionType: enc,
			VertexLocation: g.flagAHIVertexLoc,
		},
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ApiHubInstances.Create(parent, body).
		ApiHubInstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating api hub instance: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Create request issued for api hub instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(g.inv.stdout, op, g.flagAHIFormat)
}

func (g *apihubCmds) runAHIDelete(cmd *cobra.Command, args []string) error {
	name, err := g.ahiName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ApiHubInstances.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting api hub instance: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Delete request issued for api hub instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(g.inv.stdout, op, g.flagAHIFormat)
}

func (g *apihubCmds) runAHIDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.ahiName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ApiHubInstances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing api hub instance: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAHIFormat)
}

func (g *apihubCmds) runAHILookup(cmd *cobra.Command, args []string) error {
	parent, err := g.ahiParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ApiHubInstances.Lookup(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up api hub instance: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAHIFormat)
}
