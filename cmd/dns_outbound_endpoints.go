package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns outbound-endpoints (gcloud-python 589.0.0, #1924) ---

var dnsOutboundEndpointsCmd = &cobra.Command{
	Use: "outbound-endpoints", Short: "Manage Cloud DNS outbound endpoints",
}

var (
	dnsOECreateCmd = &cobra.Command{
		Use: "create OUTBOUND_ENDPOINT", Short: "Create a Cloud DNS outbound endpoint",
		Args: cobra.ExactArgs(1), RunE: runDNSOECreate,
	}
	dnsOEDeleteCmd = &cobra.Command{
		Use: "delete OUTBOUND_ENDPOINT", Short: "Delete a Cloud DNS outbound endpoint",
		Args: cobra.ExactArgs(1), RunE: runDNSOEDelete,
	}
	dnsOEDescribeCmd = &cobra.Command{
		Use: "describe OUTBOUND_ENDPOINT", Short: "Describe a Cloud DNS outbound endpoint",
		Args: cobra.ExactArgs(1), RunE: runDNSOEDescribe,
	}
	dnsOEListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud DNS outbound endpoints in a location",
		Args: cobra.NoArgs, RunE: runDNSOEList,
	}
	dnsOEUpdateCmd = &cobra.Command{
		Use: "update OUTBOUND_ENDPOINT", Short: "Update the description or labels of a Cloud DNS outbound endpoint",
		Args: cobra.ExactArgs(1), RunE: runDNSOEUpdate,
	}
)

var (
	flagDNSOELocation     string
	flagDNSOEEndpointMode string
	flagDNSOEFormat       string
	flagDNSOEAsync        bool
	flagDNSOENetwork      string
	flagDNSOESubnetwork   string
	flagDNSOEEndpointIP   string
	flagDNSOEDescription  string
	flagDNSOELabels       map[string]string
	flagDNSOEUpdateLabels map[string]string
	flagDNSOERemoveLabels []string
	flagDNSOEClearLabels  bool
	flagDNSOELimit        int
	flagDNSOEPageSize     int64
	flagDNSOEURI          bool
)

func init() {
	all := []*cobra.Command{dnsOECreateCmd, dnsOEDeleteCmd, dnsOEDescribeCmd, dnsOEListCmd, dnsOEUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagDNSOELocation, "location", "", "Region of the outbound endpoint (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagDNSOEEndpointMode, "endpoint-mode", "",
			"Endpoint mode: global (default), regional, or regional-preferred")
		c.Flags().StringVar(&flagDNSOEFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dnsOECreateCmd, dnsOEDeleteCmd, dnsOEUpdateCmd} {
		c.Flags().BoolVar(&flagDNSOEAsync, "async", false,
			"Return immediately, without waiting for the operation in progress to complete")
	}
	for _, c := range []*cobra.Command{dnsOECreateCmd, dnsOEUpdateCmd} {
		c.Flags().StringVar(&flagDNSOEDescription, "description", "", "Description of the outbound endpoint")
	}
	dnsOECreateCmd.Flags().StringVar(&flagDNSOENetwork, "network", "",
		"Network for the outbound endpoint: a network name or projects/P/global/networks/N (required)")
	dnsOECreateCmd.Flags().StringVar(&flagDNSOESubnetwork, "subnetwork", "",
		"Subnetwork for the outbound endpoint: a subnetwork name or projects/P/regions/R/subnetworks/S (required)")
	dnsOECreateCmd.Flags().StringVar(&flagDNSOEEndpointIP, "endpoint-ip", "",
		"IP address of the outbound endpoint (required)")
	for _, f := range []string{"network", "subnetwork", "endpoint-ip"} {
		_ = dnsOECreateCmd.MarkFlagRequired(f)
	}
	dnsOECreateCmd.Flags().StringToStringVar(&flagDNSOELabels, "labels", nil, "List of label KEY=VALUE pairs to add")
	dnsOEUpdateCmd.Flags().StringToStringVar(&flagDNSOEUpdateLabels, "update-labels", nil,
		"List of label KEY=VALUE pairs to update; new keys are added")
	dnsOEUpdateCmd.Flags().StringSliceVar(&flagDNSOERemoveLabels, "remove-labels", nil, "List of label keys to remove")
	dnsOEUpdateCmd.Flags().BoolVar(&flagDNSOEClearLabels, "clear-labels", false, "Remove all labels")
	dnsOEUpdateCmd.MarkFlagsMutuallyExclusive("clear-labels", "remove-labels")
	dnsOEListCmd.Flags().IntVar(&flagDNSOELimit, "limit", 0, "Maximum number of outbound endpoints to list (0 = no limit)")
	dnsOEListCmd.Flags().Int64Var(&flagDNSOEPageSize, "page-size", 0, "Number of results per API page")
	dnsOEListCmd.Flags().BoolVar(&flagDNSOEURI, "uri", false, "Print resource names instead of the table")

	dnsOutboundEndpointsCmd.AddCommand(all...)
	dnsCmd.AddCommand(dnsOutboundEndpointsCmd)
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
func dnsOEService(ctx context.Context) (string, string, *dns.Service, error) {
	project, err := resolveProject()
	if err != nil {
		return "", "", nil, err
	}
	endpoint, err := dnsOEEndpoint(flagDNSOEEndpointMode, flagDNSOELocation)
	if err != nil {
		return "", "", nil, err
	}
	svc, err := gcp.DNSServiceAt(ctx, flagAccount, endpoint)
	if err != nil {
		return "", "", nil, err
	}
	return project, fmt.Sprintf("projects/%s/locations/%s", project, flagDNSOELocation), svc, nil
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
func dnsOEWait(ctx context.Context, svc *dns.Service, op *dns.GoogleLongrunningOperation, progress string) error {
	fmt.Fprintf(os.Stderr, "%s...", progress)
	interval := dnsOEPollInterval
	for !op.Done {
		time.Sleep(interval)
		if interval *= 2; interval > 10*time.Second {
			interval = 10 * time.Second
		}
		got, err := svc.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed.")
			return err
		}
		op = got
	}
	if op.Error != nil {
		fmt.Fprintln(os.Stderr, "failed.")
		return fmt.Errorf("%s", op.Error.Message)
	}
	fmt.Fprintln(os.Stderr, "done.")
	return nil
}

func runDNSOECreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	project, parent, svc, err := dnsOEService(ctx)
	if err != nil {
		return err
	}
	name := dnsOEName(parent, args[0])
	ep := &dns.OutboundEndpoint{
		Name:        name,
		Network:     dnsOEFormatNetwork(flagDNSOENetwork, project),
		Subnetwork:  dnsOEFormatSubnetwork(flagDNSOESubnetwork, project, flagDNSOELocation),
		EndpointIp:  flagDNSOEEndpointIP,
		Description: flagDNSOEDescription,
		Labels:      flagDNSOELabels,
	}
	op, err := svc.OutboundEndpoints.Create(parent, ep).OutboundEndpointId(path.Base(name)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating outbound endpoint: %w", err)
	}
	if flagDNSOEAsync {
		fmt.Fprintf(os.Stderr, "Create in progress for OutboundEndpoint [%s] with operation [%s].\n", name, op.Name)
		return emitFormatted(op, dnsOEOutputFormat())
	}
	if err := dnsOEWait(ctx, svc, op, fmt.Sprintf("Creating OutboundEndpoint [%s]", name)); err != nil {
		return err
	}
	got, err := svc.OutboundEndpoints.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing outbound endpoint: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Created OutboundEndpoint [%s].\n", name)
	return emitFormatted(got, dnsOEOutputFormat())
}

// dnsOEOutputFormat is --format, defaulting to json as in gcloud-python's
// create and update.
func dnsOEOutputFormat() string {
	if flagDNSOEFormat == "" {
		return "json"
	}
	return flagDNSOEFormat
}

func runDNSOEDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	_, parent, svc, err := dnsOEService(ctx)
	if err != nil {
		return err
	}
	name := dnsOEName(parent, args[0])
	if err := promptContinue(fmt.Sprintf("You are about to delete outbound endpoint [%s]", name), ""); err != nil {
		fmt.Fprintln(os.Stderr, "Aborted by user.")
		return nil
	}
	op, err := svc.OutboundEndpoints.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting outbound endpoint: %w", err)
	}
	if flagDNSOEAsync {
		fmt.Fprintf(os.Stderr, "Delete in progress for OutboundEndpoint [%s] with operation [%s].\n", name, op.Name)
		return emitFormatted(op, flagDNSOEFormat)
	}
	if err := dnsOEWait(ctx, svc, op, fmt.Sprintf("Deleting OutboundEndpoint [%s]", name)); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Deleted OutboundEndpoint [%s].\n", name)
	return nil
}

func runDNSOEDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	_, parent, svc, err := dnsOEService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.OutboundEndpoints.Get(dnsOEName(parent, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing outbound endpoint: %w", err)
	}
	return emitFormatted(got, flagDNSOEFormat)
}

func runDNSOEList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	_, parent, svc, err := dnsOEService(ctx)
	if err != nil {
		return err
	}
	var all []*dns.OutboundEndpoint
	pageToken := ""
	for {
		call := svc.OutboundEndpoints.List(parent).Context(ctx)
		if flagDNSOEPageSize > 0 {
			call = call.PageSize(flagDNSOEPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing outbound endpoints: %w", err)
		}
		all = append(all, resp.OutboundEndpoints...)
		if flagDNSOELimit > 0 && len(all) >= flagDNSOELimit {
			all = all[:flagDNSOELimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitDNSOEList(os.Stdout, os.Stderr, all, flagDNSOEFormat, flagDNSOEURI)
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

// dnsOEApplyLabels applies gcloud's labels Diff (clear, then remove, then
// update) to current and reports whether the result differs.
func dnsOEApplyLabels(current map[string]string, clear bool, remove []string, update map[string]string) (map[string]string, bool) {
	next := map[string]string{}
	if !clear {
		for k, v := range current {
			next[k] = v
		}
	}
	for _, k := range remove {
		delete(next, k)
	}
	for k, v := range update {
		next[k] = v
	}
	if len(next) != len(current) {
		return next, true
	}
	for k, v := range next {
		if cv, ok := current[k]; !ok || cv != v {
			return next, true
		}
	}
	return next, false
}

func runDNSOEUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	_, parent, svc, err := dnsOEService(ctx)
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
		ep.Description = flagDNSOEDescription
		mask = append(mask, "description")
	}
	if flagDNSOEClearLabels || len(flagDNSOERemoveLabels) > 0 || len(flagDNSOEUpdateLabels) > 0 {
		if labels, changed := dnsOEApplyLabels(ep.Labels, flagDNSOEClearLabels, flagDNSOERemoveLabels, flagDNSOEUpdateLabels); changed {
			ep.Labels = labels
			if len(labels) == 0 {
				ep.NullFields = append(ep.NullFields, "Labels")
			}
			mask = append(mask, "labels")
		}
	}
	if len(mask) == 0 {
		fmt.Fprintln(os.Stderr, "Nothing to update.")
		return emitFormatted(ep, dnsOEOutputFormat())
	}
	op, err := svc.OutboundEndpoints.Patch(name, ep).UpdateMask(strings.Join(mask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating outbound endpoint: %w", err)
	}
	if flagDNSOEAsync {
		fmt.Fprintf(os.Stderr, "Update in progress for OutboundEndpoint [%s] with operation [%s].\n", name, op.Name)
		return emitFormatted(op, dnsOEOutputFormat())
	}
	if err := dnsOEWait(ctx, svc, op, fmt.Sprintf("Updating OutboundEndpoint [%s]", name)); err != nil {
		return err
	}
	got, err := svc.OutboundEndpoints.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing outbound endpoint: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated OutboundEndpoint [%s].\n", name)
	return emitFormatted(got, dnsOEOutputFormat())
}
