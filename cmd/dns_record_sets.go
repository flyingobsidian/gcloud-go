package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns record-sets (#1541) ---

var dnsRRCmd = &cobra.Command{Use: "record-sets", Short: "Manage Cloud DNS record sets"}

var dnsRRChangesCmd = &cobra.Command{Use: "changes", Short: "Manage record-set changes"}
var dnsRRTxnCmd = &cobra.Command{Use: "transaction", Short: "Stage record-set changes as a transaction"}

var (
	flagDNSRRZone       string
	flagDNSRRFormat     string
	flagDNSRRName       string
	flagDNSRRType       string
	flagDNSRRTTL        int64
	flagDNSRRRRDatas    []string
	flagDNSRRConfigFile string
	flagDNSRRMaxResults int64
	flagDNSRRImportFile string
	flagDNSRRExportFile string
	flagDNSRRReplaceAll bool
	flagDNSRRZoneFmt    bool
	flagDNSRRTxnFile    string
)

var (
	dnsRRCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Create a resource record set",
		Args: cobra.ExactArgs(1), RunE: runDNSRRCreate,
	}
	dnsRRDeleteCmd = &cobra.Command{
		Use: "delete NAME", Short: "Delete a resource record set",
		Args: cobra.ExactArgs(1), RunE: runDNSRRDelete,
	}
	dnsRRDescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Describe a resource record set",
		Args: cobra.ExactArgs(1), RunE: runDNSRRDescribe,
	}
	dnsRRListCmd = &cobra.Command{
		Use: "list", Short: "List resource record sets in a zone",
		Args: cobra.NoArgs, RunE: runDNSRRList,
	}
	dnsRRUpdateCmd = &cobra.Command{
		Use: "update NAME", Short: "Update a resource record set",
		Args: cobra.ExactArgs(1), RunE: runDNSRRUpdate,
	}
	dnsRRImportCmd = &cobra.Command{
		Use: "import", Short: "Import record sets from a file (YAML/JSON)",
		Args: cobra.NoArgs, RunE: runDNSRRImport,
	}
	dnsRRExportCmd = &cobra.Command{
		Use: "export", Short: "Export record sets to a file (YAML/JSON)",
		Args: cobra.NoArgs, RunE: runDNSRRExport,
	}
	dnsRRChangesDescribeCmd = &cobra.Command{
		Use: "describe CHANGE_ID", Short: "Describe a change",
		Args: cobra.ExactArgs(1), RunE: runDNSRRChangesDescribe,
	}
	dnsRRChangesListCmd = &cobra.Command{
		Use: "list", Short: "List changes for a zone",
		Args: cobra.NoArgs, RunE: runDNSRRChangesList,
	}
	dnsRRTxnStartCmd = &cobra.Command{
		Use: "start", Short: "Start a transaction (writes an empty Change to --transaction-file)",
		Args: cobra.NoArgs, RunE: runDNSRRTxnStart,
	}
	dnsRRTxnAddCmd = &cobra.Command{
		Use: "add NAME", Short: "Add a record set to the current transaction",
		Args: cobra.ExactArgs(1), RunE: runDNSRRTxnAdd,
	}
	dnsRRTxnRemoveCmd = &cobra.Command{
		Use: "remove NAME", Short: "Remove a record set as part of the current transaction",
		Args: cobra.ExactArgs(1), RunE: runDNSRRTxnRemove,
	}
	dnsRRTxnDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the staged transaction",
		Args: cobra.NoArgs, RunE: runDNSRRTxnDescribe,
	}
	dnsRRTxnExecuteCmd = &cobra.Command{
		Use: "execute", Short: "Submit the staged transaction",
		Args: cobra.NoArgs, RunE: runDNSRRTxnExecute,
	}
	dnsRRTxnAbortCmd = &cobra.Command{
		Use: "abort", Short: "Discard the staged transaction",
		Args: cobra.NoArgs, RunE: runDNSRRTxnAbort,
	}
)

func init() {
	base := []*cobra.Command{
		dnsRRCreateCmd, dnsRRDeleteCmd, dnsRRDescribeCmd, dnsRRListCmd, dnsRRUpdateCmd,
		dnsRRImportCmd, dnsRRExportCmd,
	}
	for _, c := range base {
		c.Flags().StringVar(&flagDNSRRZone, "zone", "", "Managed zone name (required)")
		_ = c.MarkFlagRequired("zone")
		c.Flags().StringVar(&flagDNSRRFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dnsRRCreateCmd, dnsRRUpdateCmd} {
		c.Flags().StringVar(&flagDNSRRType, "type", "", "Record type (e.g. A, AAAA, CNAME) (required)")
		_ = c.MarkFlagRequired("type")
		c.Flags().Int64Var(&flagDNSRRTTL, "ttl", 300, "TTL in seconds")
		c.Flags().StringSliceVar(&flagDNSRRRRDatas, "rrdatas", nil, "Record data (repeatable) (required)")
		_ = c.MarkFlagRequired("rrdatas")
		c.Flags().StringVar(&flagDNSRRConfigFile, "config-file", "",
			"Optional path to a YAML/JSON file with the ResourceRecordSet body (overrides individual flags)")
	}
	dnsRRDeleteCmd.Flags().StringVar(&flagDNSRRType, "type", "", "Record type (required)")
	_ = dnsRRDeleteCmd.MarkFlagRequired("type")
	dnsRRDescribeCmd.Flags().StringVar(&flagDNSRRType, "type", "", "Record type (required)")
	_ = dnsRRDescribeCmd.MarkFlagRequired("type")
	dnsRRListCmd.Flags().StringVar(&flagDNSRRName, "name", "", "Restrict results to record sets with this name")
	dnsRRListCmd.Flags().StringVar(&flagDNSRRType, "type", "", "Restrict results to record sets of this type")
	dnsRRListCmd.Flags().Int64Var(&flagDNSRRMaxResults, "limit", 0, "Maximum results per page")

	dnsRRImportCmd.Flags().StringVar(&flagDNSRRImportFile, "source", "",
		"Path to a YAML/JSON file with a list of ResourceRecordSet objects (required)")
	_ = dnsRRImportCmd.MarkFlagRequired("source")
	dnsRRImportCmd.Flags().BoolVar(&flagDNSRRReplaceAll, "replace-origin-ns", false,
		"Replace existing NS records at the origin")
	dnsRRImportCmd.Flags().BoolVar(&flagDNSRRZoneFmt, "zone-file-format", false,
		"Treat --source as a BIND zone file (unsupported; use JSON/YAML)")

	dnsRRExportCmd.Flags().StringVar(&flagDNSRRExportFile, "destination", "",
		"Path to write the exported record sets to (required)")
	_ = dnsRRExportCmd.MarkFlagRequired("destination")

	dnsRRCmd.AddCommand(base...)

	// changes subgroup
	for _, c := range []*cobra.Command{dnsRRChangesDescribeCmd, dnsRRChangesListCmd} {
		c.Flags().StringVar(&flagDNSRRZone, "zone", "", "Managed zone name (required)")
		_ = c.MarkFlagRequired("zone")
		c.Flags().StringVar(&flagDNSRRFormat, "format", "", "Output format")
	}
	dnsRRChangesListCmd.Flags().Int64Var(&flagDNSRRMaxResults, "limit", 0, "Maximum results per page")
	dnsRRChangesCmd.AddCommand(dnsRRChangesDescribeCmd, dnsRRChangesListCmd)
	dnsRRCmd.AddCommand(dnsRRChangesCmd)

	// transaction subgroup
	txnAll := []*cobra.Command{
		dnsRRTxnStartCmd, dnsRRTxnAddCmd, dnsRRTxnRemoveCmd,
		dnsRRTxnDescribeCmd, dnsRRTxnExecuteCmd, dnsRRTxnAbortCmd,
	}
	for _, c := range txnAll {
		c.Flags().StringVar(&flagDNSRRZone, "zone", "", "Managed zone name (required)")
		_ = c.MarkFlagRequired("zone")
		c.Flags().StringVar(&flagDNSRRTxnFile, "transaction-file", "transaction.yaml",
			"Local file used to stage the transaction")
	}
	for _, c := range []*cobra.Command{dnsRRTxnAddCmd, dnsRRTxnRemoveCmd} {
		c.Flags().StringVar(&flagDNSRRType, "type", "", "Record type (required)")
		_ = c.MarkFlagRequired("type")
		c.Flags().Int64Var(&flagDNSRRTTL, "ttl", 300, "TTL in seconds")
		c.Flags().StringSliceVar(&flagDNSRRRRDatas, "rrdatas", nil, "Record data (repeatable) (required)")
		_ = c.MarkFlagRequired("rrdatas")
	}
	dnsRRTxnExecuteCmd.Flags().StringVar(&flagDNSRRFormat, "format", "", "Output format")
	dnsRRTxnCmd.AddCommand(txnAll...)
	dnsRRCmd.AddCommand(dnsRRTxnCmd)

	dnsCmd.AddCommand(dnsRRCmd)
}

func dnsRRLoadOrBuild(name string) (*dns.ResourceRecordSet, error) {
	if flagDNSRRConfigFile != "" {
		body := &dns.ResourceRecordSet{}
		if err := loadYAMLOrJSONInto(flagDNSRRConfigFile, body); err != nil {
			return nil, err
		}
		return body, nil
	}
	return &dns.ResourceRecordSet{
		Name:    name,
		Type:    flagDNSRRType,
		Ttl:     flagDNSRRTTL,
		Rrdatas: flagDNSRRRRDatas,
	}, nil
}

func runDNSRRCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body, err := dnsRRLoadOrBuild(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.ResourceRecordSets.Create(project, flagDNSRRZone, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating record set: %w", err)
	}
	fmt.Printf("Created record set [%s] type=%s.\n", args[0], body.Type)
	return emitFormatted(got, flagDNSRRFormat)
}

func runDNSRRDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.ResourceRecordSets.Delete(project, flagDNSRRZone, args[0], flagDNSRRType).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting record set: %w", err)
	}
	fmt.Printf("Deleted record set [%s] type=%s.\n", args[0], flagDNSRRType)
	return nil
}

func runDNSRRDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.ResourceRecordSets.Get(project, flagDNSRRZone, args[0], flagDNSRRType).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing record set: %w", err)
	}
	return emitFormatted(got, flagDNSRRFormat)
}

func runDNSRRList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*dns.ResourceRecordSet
	pageToken := ""
	for {
		call := svc.ResourceRecordSets.List(project, flagDNSRRZone).Context(ctx)
		if flagDNSRRName != "" {
			call = call.Name(flagDNSRRName)
		}
		if flagDNSRRType != "" {
			call = call.Type(flagDNSRRType)
		}
		if flagDNSRRMaxResults > 0 {
			call = call.MaxResults(flagDNSRRMaxResults)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing record sets: %w", err)
		}
		all = append(all, resp.Rrsets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDNSRRFormat)
}

func runDNSRRUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body, err := dnsRRLoadOrBuild(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.ResourceRecordSets.Patch(project, flagDNSRRZone, args[0], body.Type, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating record set: %w", err)
	}
	fmt.Printf("Updated record set [%s] type=%s.\n", args[0], body.Type)
	return emitFormatted(got, flagDNSRRFormat)
}

func runDNSRRImport(cmd *cobra.Command, args []string) error {
	if flagDNSRRZoneFmt {
		return fmt.Errorf("--zone-file-format is not supported; provide YAML/JSON")
	}
	project, err := resolveProject()
	if err != nil {
		return err
	}
	var payload struct {
		Additions []*dns.ResourceRecordSet `json:"additions" yaml:"additions"`
		Deletions []*dns.ResourceRecordSet `json:"deletions" yaml:"deletions"`
	}
	if err := loadYAMLOrJSONInto(flagDNSRRImportFile, &payload); err != nil {
		return err
	}
	if len(payload.Additions) == 0 && len(payload.Deletions) == 0 {
		var rrs []*dns.ResourceRecordSet
		if err := loadYAMLOrJSONInto(flagDNSRRImportFile, &rrs); err != nil {
			return fmt.Errorf("source file must contain {additions:[...],deletions:[...]} or a list of ResourceRecordSet: %w", err)
		}
		payload.Additions = rrs
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Changes.Create(project, flagDNSRRZone, &dns.Change{
		Additions: payload.Additions,
		Deletions: payload.Deletions,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing record sets: %w", err)
	}
	fmt.Printf("Applied %d additions and %d deletions (change ID: %s).\n",
		len(payload.Additions), len(payload.Deletions), got.Id)
	return emitFormatted(got, flagDNSRRFormat)
}

func runDNSRRExport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*dns.ResourceRecordSet
	pageToken := ""
	for {
		call := svc.ResourceRecordSets.List(project, flagDNSRRZone).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("exporting record sets: %w", err)
		}
		all = append(all, resp.Rrsets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	buf, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling: %w", err)
	}
	if err := os.WriteFile(filepath.Clean(flagDNSRRExportFile), buf, 0644); err != nil {
		return fmt.Errorf("writing export: %w", err)
	}
	fmt.Printf("Wrote %d record sets to %s.\n", len(all), flagDNSRRExportFile)
	return nil
}

func runDNSRRChangesDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Changes.Get(project, flagDNSRRZone, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing change: %w", err)
	}
	return emitFormatted(got, flagDNSRRFormat)
}

func runDNSRRChangesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*dns.Change
	pageToken := ""
	for {
		call := svc.Changes.List(project, flagDNSRRZone).Context(ctx)
		if flagDNSRRMaxResults > 0 {
			call = call.MaxResults(flagDNSRRMaxResults)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing changes: %w", err)
		}
		all = append(all, resp.Changes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDNSRRFormat)
}

func dnsRRLoadTxn() (*dns.Change, error) {
	data, err := os.ReadFile(flagDNSRRTxnFile)
	if err != nil {
		return nil, fmt.Errorf("reading transaction file: %w", err)
	}
	change := &dns.Change{}
	if err := json.Unmarshal(data, change); err != nil {
		return nil, fmt.Errorf("parsing transaction file: %w", err)
	}
	return change, nil
}

func dnsRRSaveTxn(change *dns.Change) error {
	buf, err := json.MarshalIndent(change, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling transaction: %w", err)
	}
	return os.WriteFile(flagDNSRRTxnFile, buf, 0644)
}

func runDNSRRTxnStart(cmd *cobra.Command, args []string) error {
	if _, err := os.Stat(flagDNSRRTxnFile); err == nil {
		return fmt.Errorf("transaction file %s already exists; abort or execute first", flagDNSRRTxnFile)
	}
	return dnsRRSaveTxn(&dns.Change{})
}

func runDNSRRTxnAdd(cmd *cobra.Command, args []string) error {
	change, err := dnsRRLoadTxn()
	if err != nil {
		return err
	}
	change.Additions = append(change.Additions, &dns.ResourceRecordSet{
		Name:    args[0],
		Type:    flagDNSRRType,
		Ttl:     flagDNSRRTTL,
		Rrdatas: flagDNSRRRRDatas,
	})
	return dnsRRSaveTxn(change)
}

func runDNSRRTxnRemove(cmd *cobra.Command, args []string) error {
	change, err := dnsRRLoadTxn()
	if err != nil {
		return err
	}
	change.Deletions = append(change.Deletions, &dns.ResourceRecordSet{
		Name:    args[0],
		Type:    flagDNSRRType,
		Ttl:     flagDNSRRTTL,
		Rrdatas: flagDNSRRRRDatas,
	})
	return dnsRRSaveTxn(change)
}

func runDNSRRTxnDescribe(cmd *cobra.Command, args []string) error {
	change, err := dnsRRLoadTxn()
	if err != nil {
		return err
	}
	return emitFormatted(change, "")
}

func runDNSRRTxnExecute(cmd *cobra.Command, args []string) error {
	change, err := dnsRRLoadTxn()
	if err != nil {
		return err
	}
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Changes.Create(project, flagDNSRRZone, change).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("executing transaction: %w", err)
	}
	if err := os.Remove(flagDNSRRTxnFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing transaction file: %w", err)
	}
	fmt.Printf("Executed transaction (change ID: %s).\n", got.Id)
	return emitFormatted(got, flagDNSRRFormat)
}

func runDNSRRTxnAbort(cmd *cobra.Command, args []string) error {
	if err := os.Remove(flagDNSRRTxnFile); err != nil {
		return fmt.Errorf("removing transaction file: %w", err)
	}
	fmt.Println("Aborted transaction.")
	return nil
}

