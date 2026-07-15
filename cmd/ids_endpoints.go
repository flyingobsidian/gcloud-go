package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	ids "google.golang.org/api/ids/v1"
)

// --- gcloud ids endpoints (#1196) ---

var idsEndpointsCmd = &cobra.Command{Use: "endpoints", Short: "Manage Cloud IDS endpoints"}

var (
	flagIDSZone             string
	flagIDSNetwork          string
	flagIDSDescription      string
	flagIDSSeverity         string
	flagIDSThreatExceptions []string
	flagIDSTrafficLogs      bool
	flagIDSLabels           []string
	flagIDSFormat           string
	flagIDSFilter           string
	flagIDSPageSize         int64
	flagIDSUpdateMask       string
)

var (
	idsEndpointsCreateCmd = &cobra.Command{
		Use: "create ENDPOINT", Short: "Create a Cloud IDS endpoint",
		Args: cobra.ExactArgs(1), RunE: runIDSCreate,
	}
	idsEndpointsDeleteCmd = &cobra.Command{
		Use: "delete ENDPOINT", Short: "Delete a Cloud IDS endpoint",
		Args: cobra.ExactArgs(1), RunE: runIDSDelete,
	}
	idsEndpointsDescribeCmd = &cobra.Command{
		Use: "describe ENDPOINT", Short: "Describe a Cloud IDS endpoint",
		Args: cobra.ExactArgs(1), RunE: runIDSDescribe,
	}
	idsEndpointsListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud IDS endpoints",
		Args: cobra.NoArgs, RunE: runIDSList,
	}
	idsEndpointsUpdateCmd = &cobra.Command{
		Use: "update ENDPOINT", Short: "Update a Cloud IDS endpoint",
		Args: cobra.ExactArgs(1), RunE: runIDSUpdate,
	}
)

func init() {
	all := []*cobra.Command{idsEndpointsCreateCmd, idsEndpointsDeleteCmd, idsEndpointsDescribeCmd, idsEndpointsListCmd, idsEndpointsUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagIDSZone, "zone", "",
			"Zone containing the endpoint, e.g. us-central1-a (required)")
		_ = c.MarkFlagRequired("zone")
		c.Flags().StringVar(&flagIDSFormat, "format", "", "Output format")
	}
	idsEndpointsCreateCmd.Flags().StringVar(&flagIDSNetwork, "network", "",
		"Fully qualified URL of the VPC network (required)")
	_ = idsEndpointsCreateCmd.MarkFlagRequired("network")
	idsEndpointsCreateCmd.Flags().StringVar(&flagIDSDescription, "description", "",
		"Human-readable description")
	idsEndpointsCreateCmd.Flags().StringVar(&flagIDSSeverity, "severity", "",
		"Lowest threat severity to alert on: informational, low, medium, high, critical (required)")
	_ = idsEndpointsCreateCmd.MarkFlagRequired("severity")
	idsEndpointsCreateCmd.Flags().StringSliceVar(&flagIDSThreatExceptions, "threat-exceptions", nil,
		"Threat IDs to except from generating alerts")
	idsEndpointsCreateCmd.Flags().BoolVar(&flagIDSTrafficLogs, "enable-traffic-logs", false,
		"Whether the endpoint should report traffic logs in addition to threat logs")
	idsEndpointsCreateCmd.Flags().StringSliceVar(&flagIDSLabels, "labels", nil,
		"Labels as KEY=VALUE pairs")
	idsEndpointsUpdateCmd.Flags().StringSliceVar(&flagIDSThreatExceptions, "threat-exceptions", nil,
		"Threat IDs to except from generating alerts (empty to clear)")
	idsEndpointsUpdateCmd.Flags().StringVar(&flagIDSUpdateMask, "update-mask", "threatExceptions",
		"Comma-separated list of fields to update")
	idsEndpointsListCmd.Flags().StringVar(&flagIDSFilter, "filter", "", "Server-side filter expression")
	idsEndpointsListCmd.Flags().Int64Var(&flagIDSPageSize, "page-size", 0, "Maximum number of results per page")

	idsEndpointsCmd.AddCommand(all...)
	idsCmd.AddCommand(idsEndpointsCmd)
}

func idsSeverityEnum(v string) (string, error) {
	switch strings.ToUpper(v) {
	case "":
		return "", nil
	case "INFORMATIONAL":
		return "INFORMATIONAL", nil
	case "LOW":
		return "LOW", nil
	case "MEDIUM":
		return "MEDIUM", nil
	case "HIGH":
		return "HIGH", nil
	case "CRITICAL":
		return "CRITICAL", nil
	default:
		return "", fmt.Errorf("--severity must be one of informational, low, medium, high, critical (got %q)", v)
	}
}

func idsLabelsFromFlag() map[string]string {
	if len(flagIDSLabels) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, kv := range flagIDSLabels {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		out[k] = v
	}
	return out
}

func idsEndpointParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagIDSZone), nil
}

func idsEndpointName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := idsEndpointParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/endpoints/%s", parent, id), nil
}

func runIDSCreate(cmd *cobra.Command, args []string) error {
	parent, err := idsEndpointParent()
	if err != nil {
		return err
	}
	sev, err := idsSeverityEnum(flagIDSSeverity)
	if err != nil {
		return err
	}
	body := &ids.Endpoint{
		Description:      flagIDSDescription,
		Network:          flagIDSNetwork,
		Severity:         sev,
		ThreatExceptions: flagIDSThreatExceptions,
		TrafficLogs:      flagIDSTrafficLogs,
		Labels:           idsLabelsFromFlag(),
	}
	ctx := context.Background()
	svc, err := gcp.IDSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Endpoints.Create(parent, body).EndpointId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating endpoint: %w", err)
	}
	fmt.Printf("Create request issued for endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagIDSFormat)
}

func runIDSDelete(cmd *cobra.Command, args []string) error {
	name, err := idsEndpointName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IDSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Endpoints.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting endpoint: %w", err)
	}
	fmt.Printf("Delete request issued for endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagIDSFormat)
}

func runIDSDescribe(cmd *cobra.Command, args []string) error {
	name, err := idsEndpointName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IDSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Endpoints.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing endpoint: %w", err)
	}
	return emitFormatted(got, flagIDSFormat)
}

func runIDSList(cmd *cobra.Command, args []string) error {
	parent, err := idsEndpointParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IDSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*ids.Endpoint
	pageToken := ""
	for {
		call := svc.Projects.Locations.Endpoints.List(parent).Context(ctx)
		if flagIDSFilter != "" {
			call = call.Filter(flagIDSFilter)
		}
		if flagIDSPageSize > 0 {
			call = call.PageSize(flagIDSPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing endpoints: %w", err)
		}
		all = append(all, resp.Endpoints...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagIDSFormat)
}

func runIDSUpdate(cmd *cobra.Command, args []string) error {
	name, err := idsEndpointName(args[0])
	if err != nil {
		return err
	}
	body := &ids.Endpoint{
		ThreatExceptions: flagIDSThreatExceptions,
	}
	// The API distinguishes an explicit empty list from an unset one; force
	// the field so an empty --threat-exceptions clears the server value.
	body.ForceSendFields = []string{"ThreatExceptions"}
	ctx := context.Background()
	svc, err := gcp.IDSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Endpoints.Patch(name, body).Context(ctx)
	if flagIDSUpdateMask != "" {
		call = call.UpdateMask(flagIDSUpdateMask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating endpoint: %w", err)
	}
	fmt.Printf("Update request issued for endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagIDSFormat)
}
