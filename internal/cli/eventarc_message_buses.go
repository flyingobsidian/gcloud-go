package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

func (grp *eventarcCmds) initEventarcMessageBuses() {
	for _, c := range []*cobra.Command{grp.evMBCreateCmd, grp.evMBDeleteCmd, grp.evMBDescribeCmd, grp.evMBListCmd, grp.evMBUpdateCmd, grp.evMBListEnrCmd, grp.evMBPublishCmd} {
		eventarcAddRegionFlag(c, &grp.flagEvMBLocation, true)
	}
	for _, c := range []*cobra.Command{grp.evMBCreateCmd, grp.evMBUpdateCmd} {
		c.Flags().StringVar(&grp.flagEvMBConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the MessageBus message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	grp.evMBUpdateCmd.Flags().StringVar(&grp.flagEvMBUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{grp.evMBCreateCmd, grp.evMBDeleteCmd, grp.evMBUpdateCmd} {
		c.Flags().BoolVar(&grp.flagEvMBAsync, "async", false, "Return the long-running operation without waiting")
	}
	grp.evMBDescribeCmd.Flags().StringVar(&grp.flagEvMBFormat, "format", "", "Output format")
	grp.evMBListCmd.Flags().StringVar(&grp.flagEvMBFormat, "format", "", "Output format")
	grp.evMBListCmd.Flags().Int64Var(&grp.flagEvMBListPage, "page-size", 0, "Page size")
	grp.evMBListCmd.Flags().Int64Var(&grp.flagEvMBListLimit, "limit", 0, "Cap total results (0 = no cap)")
	grp.evMBListCmd.Flags().StringVar(&grp.flagEvMBListFilter, "filter", "", "Server-side filter expression")
	grp.evMBListCmd.Flags().BoolVar(&grp.flagEvMBListURI, "uri", false, "Print resource names only")

	grp.evMBListEnrCmd.Flags().Int64Var(&grp.flagEvMBListPage, "page-size", 0, "Page size")
	grp.evMBListEnrCmd.Flags().Int64Var(&grp.flagEvMBListLimit, "limit", 0, "Cap total results (0 = no cap)")
	grp.evMBListEnrCmd.Flags().StringVar(&grp.flagEvMBFormat, "format", "", "Output format")

	grp.evMBPublishCmd.Flags().StringVar(&grp.flagEvMBPublishEventID, "event-id", "", "CloudEvents id")
	grp.evMBPublishCmd.Flags().StringVar(&grp.flagEvMBPublishType, "event-type", "", "CloudEvents type")
	grp.evMBPublishCmd.Flags().StringVar(&grp.flagEvMBPublishSource, "event-source", "", "CloudEvents source URI")
	grp.evMBPublishCmd.Flags().StringVar(&grp.flagEvMBPublishData, "event-data", "", "CloudEvents payload data (JSON string)")
	grp.evMBPublishCmd.Flags().StringArrayVar(&grp.flagEvMBPublishAttrs, "event-attributes", nil,
		"Additional CloudEvent attributes; repeat as KEY=VALUE")
	grp.evMBPublishCmd.Flags().StringVar(&grp.flagEvMBPublishJSON, "json-message", "",
		"Full CloudEvent as a JSON string (mutually exclusive with the --event-* flags)")
	grp.evMBPublishCmd.Flags().StringVar(&grp.flagEvMBPublishEnroll, "destination-enrollment", "",
		"Fully qualified enrollment name to route the event to")

	grp.eventarcMessageBusesCmd.AddCommand(grp.evMBCreateCmd, grp.evMBDeleteCmd, grp.evMBDescribeCmd, grp.evMBListCmd, grp.evMBUpdateCmd, grp.evMBListEnrCmd, grp.evMBPublishCmd)
	grp.eventarcCmd.AddCommand(grp.eventarcMessageBusesCmd)
}

func (grp *eventarcCmds) runEvMBCreate(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	mb := &eventarc.MessageBus{}
	if err := loadYAMLOrJSONInto(grp.flagEvMBConfigFile, mb); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MessageBuses.Create(eventarcLocationParent(project, grp.flagEvMBLocation), mb).
		MessageBusId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating message bus: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Create message bus", args[0], grp.flagEvMBAsync)
}

func (grp *eventarcCmds) runEvMBDelete(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MessageBuses.Delete(eventarcResourceName("messageBuses", args[0], project, grp.flagEvMBLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting message bus: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Delete message bus", args[0], grp.flagEvMBAsync)
}

func (grp *eventarcCmds) runEvMBDescribe(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	mb, err := svc.Projects.Locations.MessageBuses.Get(eventarcResourceName("messageBuses", args[0], project, grp.flagEvMBLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing message bus: %w", err)
	}
	return emitFormattedTo(grp.inv.stdout, mb, grp.flagEvMBFormat)
}

func (grp *eventarcCmds) runEvMBList(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, grp.flagEvMBLocation)
	var all []*eventarc.MessageBus
	pageToken := ""
	for {
		call := svc.Projects.Locations.MessageBuses.List(parent).Context(ctx)
		if grp.flagEvMBListFilter != "" {
			call = call.Filter(grp.flagEvMBListFilter)
		}
		if grp.flagEvMBListPage > 0 {
			call = call.PageSize(grp.flagEvMBListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing message buses: %w", err)
		}
		all = append(all, resp.MessageBuses...)
		if grp.flagEvMBListLimit > 0 && int64(len(all)) >= grp.flagEvMBListLimit {
			all = all[:grp.flagEvMBListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if grp.flagEvMBListURI {
		for _, mb := range all {
			fmt.Fprintln(grp.inv.stdout, mb.Name)
		}
		return nil
	}
	if grp.flagEvMBFormat != "" {
		return emitFormattedTo(grp.inv.stdout, all, grp.flagEvMBFormat)
	}
	printTableHeadingTo(grp.inv.stdout, len(all), "%-40s %s\n", "NAME", "UPDATE_TIME")
	for _, mb := range all {
		fmt.Fprintf(grp.inv.stdout, "%-40s %s\n", path.Base(mb.Name), mb.UpdateTime)
	}
	return nil
}

func (grp *eventarcCmds) runEvMBUpdate(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	mb := &eventarc.MessageBus{}
	if err := loadYAMLOrJSONInto(grp.flagEvMBConfigFile, mb); err != nil {
		return err
	}
	mask := grp.flagEvMBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(mb))
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MessageBuses.Patch(eventarcResourceName("messageBuses", args[0], project, grp.flagEvMBLocation), mb).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating message bus: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Update message bus", args[0], grp.flagEvMBAsync)
}

func (grp *eventarcCmds) runEvMBListEnrollments(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	parent := eventarcResourceName("messageBuses", args[0], project, grp.flagEvMBLocation)
	var all []string
	pageToken := ""
	for {
		call := svc.Projects.Locations.MessageBuses.ListEnrollments(parent).Context(ctx)
		if grp.flagEvMBListPage > 0 {
			call = call.PageSize(grp.flagEvMBListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing enrollments for message bus: %w", err)
		}
		all = append(all, resp.Enrollments...)
		if grp.flagEvMBListLimit > 0 && int64(len(all)) >= grp.flagEvMBListLimit {
			all = all[:grp.flagEvMBListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if grp.flagEvMBFormat != "" {
		return emitFormattedTo(grp.inv.stdout, all, grp.flagEvMBFormat)
	}
	for _, e := range all {
		fmt.Fprintln(grp.inv.stdout, e)
	}
	return nil
}

// runEvMBPublish posts a CloudEvent to the eventarcpublishing API (which the
// generated Go client does not cover). The endpoint accepts either a full JSON
// CloudEvent via `jsonMessage`, or a discrete set of event fields packaged as
// a `protoMessage`; this mirrors gcloud-python's `eventarc message-buses
// publish` command.
func (grp *eventarcCmds) runEvMBPublish(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	if grp.flagEvMBPublishJSON == "" && grp.flagEvMBPublishEventID == "" {
		return fmt.Errorf("either --json-message or --event-id (plus --event-type/--event-source) is required")
	}
	messageBus := eventarcResourceName("messageBuses", args[0], project, grp.flagEvMBLocation)

	body := map[string]any{"messageBus": messageBus}
	if grp.flagEvMBPublishJSON != "" {
		body["jsonMessage"] = grp.flagEvMBPublishJSON
	} else {
		attrs := map[string]any{
			"time":            map[string]string{"ceTimestamp": time.Now().UTC().Format(time.RFC3339Nano)},
			"datacontenttype": map[string]string{"ceString": "application/json"},
		}
		for _, kv := range grp.flagEvMBPublishAttrs {
			k, v, ok := strings.Cut(kv, "=")
			if !ok {
				return fmt.Errorf("--event-attributes value %q must be KEY=VALUE", kv)
			}
			attrs[k] = map[string]string{"ceString": v}
		}
		body["protoMessage"] = map[string]any{
			"@type":       "type.googleapis.com/io.cloudevents.v1.CloudEvent",
			"id":          grp.flagEvMBPublishEventID,
			"source":      grp.flagEvMBPublishSource,
			"specVersion": "1.0",
			"type":        grp.flagEvMBPublishType,
			"attributes":  attrs,
			"textData":    grp.flagEvMBPublishData,
		}
	}
	if grp.flagEvMBPublishEnroll != "" {
		body["destinationEnrollment"] = grp.flagEvMBPublishEnroll
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	ts, err := grp.inv.clients().PlatformTokenSource(ctx)
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
	fmt.Fprintln(grp.inv.stderr, "Event published successfully")
	return nil
}
