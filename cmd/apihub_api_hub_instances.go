package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub api-hub-instances (#1167) ---

var apihubAHICmd = &cobra.Command{Use: "api-hub-instances", Short: "Manage API Hub instances"}

var (
	flagAHILocation      string
	flagAHIFormat        string
	flagAHIDescription   string
	flagAHILabels        []string
	flagAHICmekKey       string
	flagAHIDisableSearch bool
	flagAHIVertexLoc     string
	flagAHIEncryption    string
)

var (
	apihubAHICreateCmd = &cobra.Command{
		Use: "create API_HUB_INSTANCE", Short: "Create an API Hub instance",
		Args: cobra.ExactArgs(1), RunE: runAHICreate,
	}
	apihubAHIDeleteCmd = &cobra.Command{
		Use: "delete API_HUB_INSTANCE", Short: "Delete an API Hub instance",
		Args: cobra.ExactArgs(1), RunE: runAHIDelete,
	}
	apihubAHIDescribeCmd = &cobra.Command{
		Use: "describe API_HUB_INSTANCE", Short: "Describe an API Hub instance",
		Args: cobra.ExactArgs(1), RunE: runAHIDescribe,
	}
	apihubAHILookupCmd = &cobra.Command{
		Use: "lookup", Short: "Look up the API Hub instance for a project location",
		Args: cobra.NoArgs, RunE: runAHILookup,
	}
)

func init() {
	all := []*cobra.Command{apihubAHICreateCmd, apihubAHIDeleteCmd, apihubAHIDescribeCmd, apihubAHILookupCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagAHILocation, "location", "",
			"Location that owns the API Hub instance (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagAHIFormat, "format", "", "Output format")
	}
	apihubAHICreateCmd.Flags().StringVar(&flagAHIDescription, "description", "",
		"Description of the API Hub instance")
	apihubAHICreateCmd.Flags().StringSliceVar(&flagAHILabels, "labels", nil,
		"Labels as KEY=VALUE pairs")
	apihubAHICreateCmd.Flags().StringVar(&flagAHICmekKey, "config-cmek-key-name", "",
		"Fully qualified CMEK key (projects/*/locations/*/keyRings/*/cryptoKeys/*)")
	apihubAHICreateCmd.Flags().BoolVar(&flagAHIDisableSearch, "config-disable-search", false,
		"Disable search on the instance")
	apihubAHICreateCmd.Flags().StringVar(&flagAHIVertexLoc, "config-vertex-location", "",
		"Vertex AI location where the data store is stored")
	apihubAHICreateCmd.Flags().StringVar(&flagAHIEncryption, "config-encryption-type", "",
		"Encryption type: gmek or cmek")

	apihubAHICmd.AddCommand(all...)
	apihubCmd.AddCommand(apihubAHICmd)
}

func ahiParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagAHILocation), nil
}

func ahiName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := ahiParent()
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

func ahiLabelsFromFlag() map[string]string {
	if len(flagAHILabels) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, kv := range flagAHILabels {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		out[k] = v
	}
	return out
}

func runAHICreate(cmd *cobra.Command, args []string) error {
	parent, err := ahiParent()
	if err != nil {
		return err
	}
	enc, err := ahiEncryptionEnum(flagAHIEncryption)
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1ApiHubInstance{
		Description: flagAHIDescription,
		Labels:      ahiLabelsFromFlag(),
		Config: &apihub.GoogleCloudApihubV1Config{
			CmekKeyName:    flagAHICmekKey,
			DisableSearch:  flagAHIDisableSearch,
			EncryptionType: enc,
			VertexLocation: flagAHIVertexLoc,
		},
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ApiHubInstances.Create(parent, body).
		ApiHubInstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating api hub instance: %w", err)
	}
	fmt.Printf("Create request issued for api hub instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAHIFormat)
}

func runAHIDelete(cmd *cobra.Command, args []string) error {
	name, err := ahiName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ApiHubInstances.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting api hub instance: %w", err)
	}
	fmt.Printf("Delete request issued for api hub instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAHIFormat)
}

func runAHIDescribe(cmd *cobra.Command, args []string) error {
	name, err := ahiName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ApiHubInstances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing api hub instance: %w", err)
	}
	return emitFormatted(got, flagAHIFormat)
}

func runAHILookup(cmd *cobra.Command, args []string) error {
	parent, err := ahiParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ApiHubInstances.Lookup(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up api hub instance: %w", err)
	}
	return emitFormatted(got, flagAHIFormat)
}
