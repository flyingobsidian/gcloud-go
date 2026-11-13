package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	ids "google.golang.org/api/ids/v1"
)

// --- gcloud ids endpoints (#1196) ---

func (s *idsCmds) initIdsEndpoints() {
	all := []*cobra.Command{s.idsEndpointsCreateCmd, s.idsEndpointsDeleteCmd, s.idsEndpointsDescribeCmd, s.idsEndpointsListCmd, s.idsEndpointsUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagIDSZone, "zone", "",
			"Zone containing the endpoint, e.g. us-central1-a (required)")
		_ = c.MarkFlagRequired("zone")
		c.Flags().StringVar(&s.flagIDSFormat, "format", "", "Output format")
	}
	s.idsEndpointsCreateCmd.Flags().StringVar(&s.flagIDSNetwork, "network", "",
		"Fully qualified URL of the VPC network (required)")
	_ = s.idsEndpointsCreateCmd.MarkFlagRequired("network")
	s.idsEndpointsCreateCmd.Flags().StringVar(&s.flagIDSDescription, "description", "",
		"Human-readable description")
	s.idsEndpointsCreateCmd.Flags().StringVar(&s.flagIDSSeverity, "severity", "",
		"Lowest threat severity to alert on: informational, low, medium, high, critical (required)")
	_ = s.idsEndpointsCreateCmd.MarkFlagRequired("severity")
	s.idsEndpointsCreateCmd.Flags().StringSliceVar(&s.flagIDSThreatExceptions, "threat-exceptions", nil,
		"Threat IDs to except from generating alerts")
	s.idsEndpointsCreateCmd.Flags().BoolVar(&s.flagIDSTrafficLogs, "enable-traffic-logs", false,
		"Whether the endpoint should report traffic logs in addition to threat logs")
	s.idsEndpointsCreateCmd.Flags().StringSliceVar(&s.flagIDSLabels, "labels", nil,
		"Labels as KEY=VALUE pairs")
	s.idsEndpointsUpdateCmd.Flags().StringSliceVar(&s.flagIDSThreatExceptions, "threat-exceptions", nil,
		"Threat IDs to except from generating alerts (empty to clear)")
	s.idsEndpointsUpdateCmd.Flags().StringVar(&s.flagIDSUpdateMask, "update-mask", "threatExceptions",
		"Comma-separated list of fields to update")
	s.idsEndpointsListCmd.Flags().StringVar(&s.flagIDSFilter, "filter", "", "Server-side filter expression")
	s.idsEndpointsListCmd.Flags().Int64Var(&s.flagIDSPageSize, "page-size", 0, "Maximum number of results per page")

	s.idsEndpointsCmd.AddCommand(all...)
	s.idsCmd.AddCommand(s.idsEndpointsCmd)
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

func (s *idsCmds) idsLabelsFromFlag() map[string]string {
	if len(s.flagIDSLabels) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, kv := range s.flagIDSLabels {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		out[k] = v
	}
	return out
}

func (s *idsCmds) idsEndpointParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagIDSZone), nil
}

func (s *idsCmds) idsEndpointName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.idsEndpointParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/endpoints/%s", parent, id), nil
}

func (s *idsCmds) runIDSCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.idsEndpointParent()
	if err != nil {
		return err
	}
	sev, err := idsSeverityEnum(s.flagIDSSeverity)
	if err != nil {
		return err
	}
	body := &ids.Endpoint{
		Description:      s.flagIDSDescription,
		Network:          s.flagIDSNetwork,
		Severity:         sev,
		ThreatExceptions: s.flagIDSThreatExceptions,
		TrafficLogs:      s.flagIDSTrafficLogs,
		Labels:           s.idsLabelsFromFlag(),
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IDS(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Endpoints.Create(parent, body).EndpointId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating endpoint: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagIDSFormat)
}

func (s *idsCmds) runIDSDelete(cmd *cobra.Command, args []string) error {
	name, err := s.idsEndpointName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IDS(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Endpoints.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting endpoint: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagIDSFormat)
}

func (s *idsCmds) runIDSDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.idsEndpointName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IDS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Endpoints.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing endpoint: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIDSFormat)
}

func (s *idsCmds) runIDSList(cmd *cobra.Command, args []string) error {
	parent, err := s.idsEndpointParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IDS(ctx)
	if err != nil {
		return err
	}
	var all []*ids.Endpoint
	pageToken := ""
	for {
		call := svc.Projects.Locations.Endpoints.List(parent).Context(ctx)
		if s.flagIDSFilter != "" {
			call = call.Filter(s.flagIDSFilter)
		}
		if s.flagIDSPageSize > 0 {
			call = call.PageSize(s.flagIDSPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagIDSFormat)
}

func (s *idsCmds) runIDSUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.idsEndpointName(args[0])
	if err != nil {
		return err
	}
	body := &ids.Endpoint{
		ThreatExceptions: s.flagIDSThreatExceptions,
	}
	// The API distinguishes an explicit empty list from an unset one; force
	// the field so an empty --threat-exceptions clears the server value.
	body.ForceSendFields = []string{"ThreatExceptions"}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IDS(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Endpoints.Patch(name, body).Context(ctx)
	if s.flagIDSUpdateMask != "" {
		call = call.UpdateMask(s.flagIDSUpdateMask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating endpoint: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagIDSFormat)
}
