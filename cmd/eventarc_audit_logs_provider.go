package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

// The gcloud-python `eventarc audit-logs-provider` group is backed by a static
// JSON catalog hosted in the google-cloudevents repository. gcloud-go reads
// the same document so its output tracks upstream changes without a re-release.
const eventarcAuditLogsCatalogURL = "https://raw.githubusercontent.com/googleapis/google-cloudevents/master/json/audit/service_catalog.json"

var eventarcAuditLogsProviderCmd = &cobra.Command{
	Use:   "audit-logs-provider",
	Short: "Explore audit log providers for Eventarc",
	Long: `Explore provider serviceNames and methodNames for event type
google.cloud.audit.log.v1.written in Eventarc.`,
}

var eventarcAuditLogsMethodNamesCmd = &cobra.Command{
	Use:   "method-names",
	Short: "Explore values for the methodName attribute",
}

var eventarcAuditLogsServiceNamesCmd = &cobra.Command{
	Use:   "service-names",
	Short: "Explore values for the serviceName attribute",
}

var eventarcAuditLogsMethodNamesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List values for the methodName attribute",
	Args:  cobra.NoArgs,
	RunE:  runEventarcAuditLogsMethodNamesList,
}

var eventarcAuditLogsServiceNamesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List values for the serviceName attribute",
	Args:  cobra.NoArgs,
	RunE:  runEventarcAuditLogsServiceNamesList,
}

var flagEventarcAuditLogsServiceName string

func init() {
	eventarcAuditLogsMethodNamesListCmd.Flags().StringVar(&flagEventarcAuditLogsServiceName, "service-name", "",
		"Name of the service to list methodName values for (required)")
	_ = eventarcAuditLogsMethodNamesListCmd.MarkFlagRequired("service-name")

	eventarcAuditLogsMethodNamesCmd.AddCommand(eventarcAuditLogsMethodNamesListCmd)
	eventarcAuditLogsServiceNamesCmd.AddCommand(eventarcAuditLogsServiceNamesListCmd)
	eventarcAuditLogsProviderCmd.AddCommand(eventarcAuditLogsMethodNamesCmd, eventarcAuditLogsServiceNamesCmd)
	eventarcCmd.AddCommand(eventarcAuditLogsProviderCmd)
}

// auditLogService models one entry in the service catalog document.
type auditLogService struct {
	ServiceName string           `json:"serviceName"`
	DisplayName string           `json:"displayName"`
	Methods     []auditLogMethod `json:"methods"`
}

type auditLogMethod struct {
	MethodName  string `json:"methodName"`
	DisplayName string `json:"displayName"`
}

type auditLogCatalog struct {
	Services []auditLogService `json:"services"`
}

var httpGetAuditLogCatalog = fetchAuditLogCatalog

func fetchAuditLogCatalog(ctx context.Context) ([]auditLogService, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, eventarcAuditLogsCatalogURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching audit log service catalog: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("fetching audit log service catalog: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading audit log service catalog: %w", err)
	}
	return parseAuditLogCatalog(body)
}

func parseAuditLogCatalog(body []byte) ([]auditLogService, error) {
	var cat auditLogCatalog
	if err := json.Unmarshal(body, &cat); err != nil {
		return nil, fmt.Errorf("parsing audit log service catalog: %w", err)
	}
	return cat.Services, nil
}

func findAuditLogMethods(services []auditLogService, serviceName string) ([]auditLogMethod, error) {
	for _, s := range services {
		if s.ServiceName == serviceName {
			return s.Methods, nil
		}
	}
	return nil, fmt.Errorf("%q is not a known value for the serviceName attribute.", serviceName)
}

func runEventarcAuditLogsServiceNamesList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	services, err := httpGetAuditLogCatalog(ctx)
	if err != nil {
		return err
	}
	if flagFormat != "" {
		return emitFormatted(services, flagFormat)
	}
	fmt.Printf("%-50s %s\n", "SERVICE_NAME", "DISPLAY_NAME")
	for _, s := range services {
		fmt.Printf("%-50s %s\n", s.ServiceName, s.DisplayName)
	}
	return nil
}

func runEventarcAuditLogsMethodNamesList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	services, err := httpGetAuditLogCatalog(ctx)
	if err != nil {
		return err
	}
	methods, err := findAuditLogMethods(services, flagEventarcAuditLogsServiceName)
	if err != nil {
		return err
	}
	if flagFormat != "" {
		return emitFormatted(methods, flagFormat)
	}
	fmt.Println("METHOD_NAME")
	for _, m := range methods {
		fmt.Println(m.MethodName)
	}
	return nil
}
