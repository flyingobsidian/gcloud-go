package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

var eventarcMessageBusesCmd = &cobra.Command{
	Use:   "message-buses",
	Short: "Manage Eventarc message buses",
}

var (
	evMBCreateCmd = &cobra.Command{
		Use: "create MESSAGE_BUS", Short: "Create a message bus from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runEvMBCreate,
	}
	evMBDeleteCmd = &cobra.Command{
		Use: "delete MESSAGE_BUS", Short: "Delete a message bus",
		Args: cobra.ExactArgs(1), RunE: runEvMBDelete,
	}
	evMBDescribeCmd = &cobra.Command{
		Use: "describe MESSAGE_BUS", Short: "Describe a message bus",
		Args: cobra.ExactArgs(1), RunE: runEvMBDescribe,
	}
	evMBListCmd = &cobra.Command{
		Use: "list", Short: "List message buses in a location",
		Args: cobra.NoArgs, RunE: runEvMBList,
	}
	evMBUpdateCmd = &cobra.Command{
		Use: "update MESSAGE_BUS", Short: "Update a message bus from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runEvMBUpdate,
	}
	evMBListEnrCmd = &cobra.Command{
		Use: "list-enrollments MESSAGE_BUS", Short: "List enrollments attached to a message bus",
		Args: cobra.ExactArgs(1), RunE: runEvMBListEnrollments,
	}
	evMBPublishCmd = &cobra.Command{
		Use: "publish MESSAGE_BUS", Short: "Publish an event to a message bus",
		Args: cobra.ExactArgs(1), RunE: runEvMBPublish,
	}
)

var (
	flagEvMBLocation        string
	flagEvMBConfigFile      string
	flagEvMBUpdateMask      string
	flagEvMBFormat          string
	flagEvMBAsync           bool
	flagEvMBListLimit       int64
	flagEvMBListPage        int64
	flagEvMBListFilter      string
	flagEvMBListURI         bool
	flagEvMBPublishEventID  string
	flagEvMBPublishType     string
	flagEvMBPublishSource   string
	flagEvMBPublishData     string
	flagEvMBPublishAttrs    []string
	flagEvMBPublishJSON     string
	flagEvMBPublishEnroll   string
)

func init() {
	for _, c := range []*cobra.Command{evMBCreateCmd, evMBDeleteCmd, evMBDescribeCmd, evMBListCmd, evMBUpdateCmd, evMBListEnrCmd, evMBPublishCmd} {
		eventarcAddRegionFlag(c, &flagEvMBLocation, true)
	}
	for _, c := range []*cobra.Command{evMBCreateCmd, evMBUpdateCmd} {
		c.Flags().StringVar(&flagEvMBConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the MessageBus message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	evMBUpdateCmd.Flags().StringVar(&flagEvMBUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{evMBCreateCmd, evMBDeleteCmd, evMBUpdateCmd} {
		c.Flags().BoolVar(&flagEvMBAsync, "async", false, "Return the long-running operation without waiting")
	}
	evMBDescribeCmd.Flags().StringVar(&flagEvMBFormat, "format", "", "Output format")
	evMBListCmd.Flags().StringVar(&flagEvMBFormat, "format", "", "Output format")
	evMBListCmd.Flags().Int64Var(&flagEvMBListPage, "page-size", 0, "Page size")
	evMBListCmd.Flags().Int64Var(&flagEvMBListLimit, "limit", 0, "Cap total results (0 = no cap)")
	evMBListCmd.Flags().StringVar(&flagEvMBListFilter, "filter", "", "Server-side filter expression")
	evMBListCmd.Flags().BoolVar(&flagEvMBListURI, "uri", false, "Print resource names only")

	evMBListEnrCmd.Flags().Int64Var(&flagEvMBListPage, "page-size", 0, "Page size")
	evMBListEnrCmd.Flags().Int64Var(&flagEvMBListLimit, "limit", 0, "Cap total results (0 = no cap)")
	evMBListEnrCmd.Flags().StringVar(&flagEvMBFormat, "format", "", "Output format")

	evMBPublishCmd.Flags().StringVar(&flagEvMBPublishEventID, "event-id", "", "CloudEvents id")
	evMBPublishCmd.Flags().StringVar(&flagEvMBPublishType, "event-type", "", "CloudEvents type")
	evMBPublishCmd.Flags().StringVar(&flagEvMBPublishSource, "event-source", "", "CloudEvents source URI")
	evMBPublishCmd.Flags().StringVar(&flagEvMBPublishData, "event-data", "", "CloudEvents payload data (JSON string)")
	evMBPublishCmd.Flags().StringArrayVar(&flagEvMBPublishAttrs, "event-attributes", nil,
		"Additional CloudEvent attributes; repeat as KEY=VALUE")
	evMBPublishCmd.Flags().StringVar(&flagEvMBPublishJSON, "json-message", "",
		"Full CloudEvent as a JSON string (mutually exclusive with the --event-* flags)")
	evMBPublishCmd.Flags().StringVar(&flagEvMBPublishEnroll, "destination-enrollment", "",
		"Fully qualified enrollment name to route the event to")

	eventarcMessageBusesCmd.AddCommand(evMBCreateCmd, evMBDeleteCmd, evMBDescribeCmd, evMBListCmd, evMBUpdateCmd, evMBListEnrCmd, evMBPublishCmd)
	eventarcCmd.AddCommand(eventarcMessageBusesCmd)
}

func runEvMBCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	mb := &eventarc.MessageBus{}
	if err := loadYAMLOrJSONInto(flagEvMBConfigFile, mb); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MessageBuses.Create(eventarcLocationParent(project, flagEvMBLocation), mb).
		MessageBusId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating message bus: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Create message bus", args[0], flagEvMBAsync)
}

func runEvMBDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MessageBuses.Delete(eventarcResourceName("messageBuses", args[0], project, flagEvMBLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting message bus: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Delete message bus", args[0], flagEvMBAsync)
}

func runEvMBDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	mb, err := svc.Projects.Locations.MessageBuses.Get(eventarcResourceName("messageBuses", args[0], project, flagEvMBLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing message bus: %w", err)
	}
	return emitFormatted(mb, flagEvMBFormat)
}

func runEvMBList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, flagEvMBLocation)
	var all []*eventarc.MessageBus
	pageToken := ""
	for {
		call := svc.Projects.Locations.MessageBuses.List(parent).Context(ctx)
		if flagEvMBListFilter != "" {
			call = call.Filter(flagEvMBListFilter)
		}
		if flagEvMBListPage > 0 {
			call = call.PageSize(flagEvMBListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing message buses: %w", err)
		}
		all = append(all, resp.MessageBuses...)
		if flagEvMBListLimit > 0 && int64(len(all)) >= flagEvMBListLimit {
			all = all[:flagEvMBListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagEvMBListURI {
		for _, mb := range all {
			fmt.Println(mb.Name)
		}
		return nil
	}
	if flagEvMBFormat != "" {
		return emitFormatted(all, flagEvMBFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "UPDATE_TIME")
	for _, mb := range all {
		fmt.Printf("%-40s %s\n", path.Base(mb.Name), mb.UpdateTime)
	}
	return nil
}

func runEvMBUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	mb := &eventarc.MessageBus{}
	if err := loadYAMLOrJSONInto(flagEvMBConfigFile, mb); err != nil {
		return err
	}
	mask := flagEvMBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(mb))
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MessageBuses.Patch(eventarcResourceName("messageBuses", args[0], project, flagEvMBLocation), mb).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating message bus: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Update message bus", args[0], flagEvMBAsync)
}

func runEvMBListEnrollments(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := eventarcResourceName("messageBuses", args[0], project, flagEvMBLocation)
	var all []string
	pageToken := ""
	for {
		call := svc.Projects.Locations.MessageBuses.ListEnrollments(parent).Context(ctx)
		if flagEvMBListPage > 0 {
			call = call.PageSize(flagEvMBListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing enrollments for message bus: %w", err)
		}
		all = append(all, resp.Enrollments...)
		if flagEvMBListLimit > 0 && int64(len(all)) >= flagEvMBListLimit {
			all = all[:flagEvMBListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagEvMBFormat != "" {
		return emitFormatted(all, flagEvMBFormat)
	}
	for _, e := range all {
		fmt.Println(e)
	}
	return nil
}

// runEvMBPublish posts a CloudEvent to the eventarcpublishing API (which the
// generated Go client does not cover). The endpoint accepts either a full JSON
// CloudEvent via `jsonMessage`, or a discrete set of event fields packaged as
// a `protoMessage`; this mirrors gcloud-python's `eventarc message-buses
// publish` command.
func runEvMBPublish(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if flagEvMBPublishJSON == "" && flagEvMBPublishEventID == "" {
		return fmt.Errorf("either --json-message or --event-id (plus --event-type/--event-source) is required")
	}
	messageBus := eventarcResourceName("messageBuses", args[0], project, flagEvMBLocation)

	body := map[string]any{"messageBus": messageBus}
	if flagEvMBPublishJSON != "" {
		body["jsonMessage"] = flagEvMBPublishJSON
	} else {
		attrs := map[string]any{
			"time":            map[string]string{"ceTimestamp": time.Now().UTC().Format(time.RFC3339Nano)},
			"datacontenttype": map[string]string{"ceString": "application/json"},
		}
		for _, kv := range flagEvMBPublishAttrs {
			k, v, ok := strings.Cut(kv, "=")
			if !ok {
				return fmt.Errorf("--event-attributes value %q must be KEY=VALUE", kv)
			}
			attrs[k] = map[string]string{"ceString": v}
		}
		body["protoMessage"] = map[string]any{
			"@type":       "type.googleapis.com/io.cloudevents.v1.CloudEvent",
			"id":          flagEvMBPublishEventID,
			"source":      flagEvMBPublishSource,
			"specVersion": "1.0",
			"type":        flagEvMBPublishType,
			"attributes":  attrs,
			"textData":    flagEvMBPublishData,
		}
	}
	if flagEvMBPublishEnroll != "" {
		body["destinationEnrollment"] = flagEvMBPublishEnroll
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	ctx := context.Background()
	ts, err := gcp.PlatformTokenSource(ctx, flagAccount)
	if err != nil {
		return err
	}
	tok, err := ts.Token()
	if err != nil {
		return fmt.Errorf("obtaining access token: %w", err)
	}

	url := fmt.Sprintf("https://eventarcpublishing.googleapis.com/v1/%s:publish", messageBus)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	tok.SetAuthHeader(req)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("publishing event: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("publishing event: HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	fmt.Fprintln(os.Stderr, "Event published successfully")
	return nil
}
