package cmd

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns outbound-endpoints (gcloud-python 589.0.0, #1924) ---

func (s *dnsCmds) initDnsOutboundEndpoints() {
	all := []*cobra.Command{s.dnsOECreateCmd, s.dnsOEDeleteCmd, s.dnsOEDescribeCmd, s.dnsOEListCmd, s.dnsOEUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDNSOELocation, "location", "", "Region of the outbound endpoint (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDNSOEEndpointMode, "endpoint-mode", "",
			"Endpoint mode: global (default), regional, or regional-preferred")
		c.Flags().StringVar(&s.flagDNSOEFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dnsOECreateCmd, s.dnsOEDeleteCmd, s.dnsOEUpdateCmd} {
		c.Flags().BoolVar(&s.flagDNSOEAsync, "async", false,
			"Return immediately, without waiting for the operation in progress to complete")
	}
	for _, c := range []*cobra.Command{s.dnsOECreateCmd, s.dnsOEUpdateCmd} {
		c.Flags().StringVar(&s.flagDNSOEDescription, "description", "", "Description of the outbound endpoint")
	}
	s.dnsOECreateCmd.Flags().StringVar(&s.flagDNSOENetwork, "network", "",
		"Network for the outbound endpoint: a network name or projects/P/global/networks/N (required)")
	s.dnsOECreateCmd.Flags().StringVar(&s.flagDNSOESubnetwork, "subnetwork", "",
		"Subnetwork for the outbound endpoint: a subnetwork name or projects/P/regions/R/subnetworks/S (required)")
	s.dnsOECreateCmd.Flags().StringVar(&s.flagDNSOEEndpointIP, "endpoint-ip", "",
		"IP address of the outbound endpoint (required)")
	for _, f := range []string{"network", "subnetwork", "endpoint-ip"} {
		_ = s.dnsOECreateCmd.MarkFlagRequired(f)
	}
	s.dnsOECreateCmd.Flags().StringToStringVar(&s.flagDNSOELabels, "labels", nil, "List of label KEY=VALUE pairs to add")
	s.dnsOEUpdateCmd.Flags().StringToStringVar(&s.flagDNSOEUpdateLabels, "update-labels", nil,
		"List of label KEY=VALUE pairs to update; new keys are added")
	s.dnsOEUpdateCmd.Flags().StringSliceVar(&s.flagDNSOERemoveLabels, "remove-labels", nil, "List of label keys to remove")
	s.dnsOEUpdateCmd.Flags().BoolVar(&s.flagDNSOEClearLabels, "clear-labels", false, "Remove all labels")
	s.dnsOEUpdateCmd.MarkFlagsMutuallyExclusive("clear-labels", "remove-labels")
	s.dnsOEListCmd.Flags().IntVar(&s.flagDNSOELimit, "limit", 0, "Maximum number of outbound endpoints to list (0 = no limit)")
	s.dnsOEListCmd.Flags().Int64Var(&s.flagDNSOEPageSize, "page-size", 0, "Number of results per API page")
	s.dnsOEListCmd.Flags().BoolVar(&s.flagDNSOEURI, "uri", false, "Print resource names instead of the table")

	s.dnsOutboundEndpointsCmd.AddCommand(all...)
	s.dnsCmd.AddCommand(s.dnsOutboundEndpointsCmd)
}

// dnsOEEndpoint mirrors gcloud-python's regional-endpoint selection for
// commands annotated RegionalEndpointsSupported: only an explicit
// --endpoint-mode=regional selects dns.LOCATION.rep.googleapis.com, as the
// 589.0.0 API map lists no known regional endpoints for dns.
func dnsOEEndpoint(mode, location string) (string, error) {
	switch mode {
	case "", "global", "regional-preferred":
		return "", nil
	case "regional":
		return fmt.Sprintf("https://dns.%s.rep.googleapis.com/", location), nil
	default:
		return "", fmt.Errorf("argument --endpoint-mode: Invalid choice: '%s'. Valid choices are [global, regional, regional-preferred]", mode)
	}
}

// dnsOEService returns the project, the projects/P/locations/L parent and a
// DNS client for the selected endpoint mode.
func (s *dnsCmds) dnsOEService(ctx context.Context) (string, string, *dns.Service, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", "", nil, err
	}
	endpoint, err := dnsOEEndpoint(s.flagDNSOEEndpointMode, s.flagDNSOELocation)
	if err != nil {
		return "", "", nil, err
	}
	svc, err := s.inv.clients().DNSAt(ctx, endpoint)
	if err != nil {
		return "", "", nil, err
	}
	return project, fmt.Sprintf("projects/%s/locations/%s", project, s.flagDNSOELocation), svc, nil
}

// dnsOEName expands a short endpoint id under parent; full names pass through.
func dnsOEName(parent, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return parent + "/outboundEndpoints/" + id
}

// dnsOEFormatNetwork and dnsOEFormatSubnetwork expand bare names as
// gcloud-python's _FormatNetwork / _FormatSubnetwork do.
func dnsOEFormatNetwork(network, project string) string {
	if network == "" || strings.Contains(network, "/") {
		return network
	}
	return fmt.Sprintf("projects/%s/global/networks/%s", project, network)
}

func dnsOEFormatSubnetwork(subnetwork, project, location string) string {
	if subnetwork == "" || strings.Contains(subnetwork, "/") {
		return subnetwork
	}
	return fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", project, location, subnetwork)
}

// dnsOEPollInterval is the initial operation poll delay (doubled per poll).
var dnsOEPollInterval = 2 * time.Second

// dnsOEWait polls a DNS operation with gcloud's "<progress>...done." line.
func (s *dnsCmds) dnsOEWait(ctx context.Context, svc *dns.Service, op *dns.GoogleLongrunningOperation, progress string) error {
	fmt.Fprintf(s.inv.stderr, "%s...", progress)
	interval := dnsOEPollInterval
	for !op.Done {
		time.Sleep(interval)
		if interval *= 2; interval > 10*time.Second {
			interval = 10 * time.Second
		}
		got, err := svc.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			fmt.Fprintln(s.inv.stderr, "failed.")
			return err
		}
		op = got
	}
	if op.Error != nil {
		fmt.Fprintln(s.inv.stderr, "failed.")
		return fmt.Errorf("%s", op.Error.Message)
	}
	fmt.Fprintln(s.inv.stderr, "done.")
	return nil
}

func (s *dnsCmds) runDNSOECreate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	project, parent, svc, err := s.dnsOEService(ctx)
	if err != nil {
		return err
	}
	name := dnsOEName(parent, args[0])
	ep := &dns.OutboundEndpoint{
		Name:        name,
		Network:     dnsOEFormatNetwork(s.flagDNSOENetwork, project),
		Subnetwork:  dnsOEFormatSubnetwork(s.flagDNSOESubnetwork, project, s.flagDNSOELocation),
		EndpointIp:  s.flagDNSOEEndpointIP,
		Description: s.flagDNSOEDescription,
		Labels:      s.flagDNSOELabels,
	}
	op, err := svc.OutboundEndpoints.Create(parent, ep).OutboundEndpointId(path.Base(name)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating outbound endpoint: %w", err)
	}
	if s.flagDNSOEAsync {
		fmt.Fprintf(s.inv.stderr, "Create in progress for OutboundEndpoint [%s] with operation [%s].\n", name, op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.dnsOEOutputFormat())
	}
	if err := s.dnsOEWait(ctx, svc, op, fmt.Sprintf("Creating OutboundEndpoint [%s]", name)); err != nil {
		return err
	}
	got, err := svc.OutboundEndpoints.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing outbound endpoint: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Created OutboundEndpoint [%s].\n", name)
	return emitFormattedTo(s.inv.stdout, got, s.dnsOEOutputFormat())
}

// dnsOEOutputFormat is --format, defaulting to json as in gcloud-python's
// create and update.
func (s *dnsCmds) dnsOEOutputFormat() string {
	if s.flagDNSOEFormat == "" {
		return "json"
	}
	return s.flagDNSOEFormat
}

func (s *dnsCmds) runDNSOEDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	_, parent, svc, err := s.dnsOEService(ctx)
	if err != nil {
		return err
	}
	name := dnsOEName(parent, args[0])
	if err := s.inv.promptContinue(fmt.Sprintf("You are about to delete outbound endpoint [%s]", name), ""); err != nil {
		fmt.Fprintln(s.inv.stderr, "Aborted by user.")
		return nil
	}
	op, err := svc.OutboundEndpoints.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting outbound endpoint: %w", err)
	}
	if s.flagDNSOEAsync {
		fmt.Fprintf(s.inv.stderr, "Delete in progress for OutboundEndpoint [%s] with operation [%s].\n", name, op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagDNSOEFormat)
	}
	if err := s.dnsOEWait(ctx, svc, op, fmt.Sprintf("Deleting OutboundEndpoint [%s]", name)); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "Deleted OutboundEndpoint [%s].\n", name)
	return nil
}

func (s *dnsCmds) runDNSOEDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	_, parent, svc, err := s.dnsOEService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.OutboundEndpoints.Get(dnsOEName(parent, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing outbound endpoint: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSOEFormat)
}

func (s *dnsCmds) runDNSOEList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	_, parent, svc, err := s.dnsOEService(ctx)
	if err != nil {
		return err
	}
	var all []*dns.OutboundEndpoint
	pageToken := ""
	for {
		call := svc.OutboundEndpoints.List(parent).Context(ctx)
		if s.flagDNSOEPageSize > 0 {
			call = call.PageSize(s.flagDNSOEPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing outbound endpoints: %w", err)
		}
		all = append(all, resp.OutboundEndpoints...)
		if s.flagDNSOELimit > 0 && len(all) >= s.flagDNSOELimit {
			all = all[:s.flagDNSOELimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitDNSOEList(s.inv.stdout, s.inv.stderr, all, s.flagDNSOEFormat, s.flagDNSOEURI)
}

// emitDNSOEList prints gcloud-python's NAME/ENDPOINT_IP/NETWORK/SUBNETWORK/
// DESCRIPTION table (basenames) unless --format or --uri is given.
func emitDNSOEList(w, errW io.Writer, all []*dns.OutboundEndpoint, format string, uri bool) error {
	if uri {
		for _, e := range all {
			if _, err := fmt.Fprintln(w, e.Name); err != nil {
				return err
			}
		}
		return nil
	}
	if format != "" {
		if all == nil {
			all = []*dns.OutboundEndpoint{}
		}
		return emitFormattedTo(w, all, format)
	}
	if len(all) == 0 {
		_, err := fmt.Fprintln(errW, "Listed 0 items.")
		return err
	}
	base := func(s string) string {
		if s == "" {
			return ""
		}
		return path.Base(s)
	}
	rows := make([]map[string]any, 0, len(all))
	for _, e := range all {
		rows = append(rows, map[string]any{
			"name": base(e.Name), "endpointIp": e.EndpointIp, "network": base(e.Network),
			"subnetwork": base(e.Subnetwork), "description": e.Description,
		})
	}
	return emitFormattedTo(w, rows, "table(name,endpointIp,network,subnetwork,description)")
}

func (s *dnsCmds) runDNSOEUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	_, parent, svc, err := s.dnsOEService(ctx)
	if err != nil {
		return err
	}
	name := dnsOEName(parent, args[0])
	ep, err := svc.OutboundEndpoints.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing outbound endpoint: %w", err)
	}
	var mask []string
	if cmd.Flags().Changed("description") {
		ep.Description = s.flagDNSOEDescription
		mask = append(mask, "description")
	}
	if s.flagDNSOEClearLabels || len(s.flagDNSOERemoveLabels) > 0 || len(s.flagDNSOEUpdateLabels) > 0 {
		if labels, changed := applyLabelsDiff(ep.Labels, s.flagDNSOEClearLabels, s.flagDNSOERemoveLabels, s.flagDNSOEUpdateLabels); changed {
			ep.Labels = labels
			if len(labels) == 0 {
				ep.NullFields = append(ep.NullFields, "Labels")
			}
			mask = append(mask, "labels")
		}
	}
	if len(mask) == 0 {
		fmt.Fprintln(s.inv.stderr, "Nothing to update.")
		return emitFormattedTo(s.inv.stdout, ep, s.dnsOEOutputFormat())
	}
	op, err := svc.OutboundEndpoints.Patch(name, ep).UpdateMask(strings.Join(mask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating outbound endpoint: %w", err)
	}
	if s.flagDNSOEAsync {
		fmt.Fprintf(s.inv.stderr, "Update in progress for OutboundEndpoint [%s] with operation [%s].\n", name, op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.dnsOEOutputFormat())
	}
	if err := s.dnsOEWait(ctx, svc, op, fmt.Sprintf("Updating OutboundEndpoint [%s]", name)); err != nil {
		return err
	}
	got, err := svc.OutboundEndpoints.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing outbound endpoint: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated OutboundEndpoint [%s].\n", name)
	return emitFormattedTo(s.inv.stdout, got, s.dnsOEOutputFormat())
}
