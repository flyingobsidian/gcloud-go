package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns record-sets (#1541) ---

func (s *dnsCmds) initDnsRecordSets() {
	base := []*cobra.Command{
		s.dnsRRCreateCmd, s.dnsRRDeleteCmd, s.dnsRRDescribeCmd, s.dnsRRListCmd, s.dnsRRUpdateCmd,
		s.dnsRRImportCmd, s.dnsRRExportCmd,
	}
	for _, c := range base {
		c.Flags().StringVar(&s.flagDNSRRZone, "zone", "", "Managed zone name (required)")
		_ = c.MarkFlagRequired("zone")
		c.Flags().StringVar(&s.flagDNSRRFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dnsRRCreateCmd, s.dnsRRUpdateCmd} {
		c.Flags().StringVar(&s.flagDNSRRType, "type", "", "Record type (e.g. A, AAAA, CNAME) (required)")
		_ = c.MarkFlagRequired("type")
		c.Flags().Int64Var(&s.flagDNSRRTTL, "ttl", 300, "TTL in seconds")
		c.Flags().StringSliceVar(&s.flagDNSRRRRDatas, "rrdatas", nil, "Record data (repeatable) (required)")
		_ = c.MarkFlagRequired("rrdatas")
		c.Flags().StringVar(&s.flagDNSRRConfigFile, "config-file", "",
			"Optional path to a YAML/JSON file with the ResourceRecordSet body (overrides individual flags)")
	}
	s.dnsRRDeleteCmd.Flags().StringVar(&s.flagDNSRRType, "type", "", "Record type (required)")
	_ = s.dnsRRDeleteCmd.MarkFlagRequired("type")
	s.dnsRRDescribeCmd.Flags().StringVar(&s.flagDNSRRType, "type", "", "Record type (required)")
	_ = s.dnsRRDescribeCmd.MarkFlagRequired("type")
	s.dnsRRListCmd.Flags().StringVar(&s.flagDNSRRName, "name", "", "Restrict results to record sets with this name")
	s.dnsRRListCmd.Flags().StringVar(&s.flagDNSRRType, "type", "", "Restrict results to record sets of this type")
	s.dnsRRListCmd.Flags().Int64Var(&s.flagDNSRRMaxResults, "limit", 0, "Maximum results per page")

	s.dnsRRImportCmd.Flags().StringVar(&s.flagDNSRRImportFile, "source", "",
		"Path to a YAML/JSON file with a list of ResourceRecordSet objects (required)")
	_ = s.dnsRRImportCmd.MarkFlagRequired("source")
	s.dnsRRImportCmd.Flags().BoolVar(&s.flagDNSRRReplaceAll, "replace-origin-ns", false,
		"Replace existing NS records at the origin")
	s.dnsRRImportCmd.Flags().BoolVar(&s.flagDNSRRZoneFmt, "zone-file-format", false,
		"Treat --source as a BIND zone file (unsupported; use JSON/YAML)")

	s.dnsRRExportCmd.Flags().StringVar(&s.flagDNSRRExportFile, "destination", "",
		"Path to write the exported record sets to (required)")
	_ = s.dnsRRExportCmd.MarkFlagRequired("destination")

	s.dnsRRCmd.AddCommand(base...)

	// changes subgroup
	for _, c := range []*cobra.Command{s.dnsRRChangesDescribeCmd, s.dnsRRChangesListCmd} {
		c.Flags().StringVar(&s.flagDNSRRZone, "zone", "", "Managed zone name (required)")
		_ = c.MarkFlagRequired("zone")
		c.Flags().StringVar(&s.flagDNSRRFormat, "format", "", "Output format")
	}
	s.dnsRRChangesListCmd.Flags().Int64Var(&s.flagDNSRRMaxResults, "limit", 0, "Maximum results per page")
	s.dnsRRChangesCmd.AddCommand(s.dnsRRChangesDescribeCmd, s.dnsRRChangesListCmd)
	s.dnsRRCmd.AddCommand(s.dnsRRChangesCmd)

	// transaction subgroup
	txnAll := []*cobra.Command{
		s.dnsRRTxnStartCmd, s.dnsRRTxnAddCmd, s.dnsRRTxnRemoveCmd,
		s.dnsRRTxnDescribeCmd, s.dnsRRTxnExecuteCmd, s.dnsRRTxnAbortCmd,
	}
	for _, c := range txnAll {
		c.Flags().StringVar(&s.flagDNSRRZone, "zone", "", "Managed zone name (required)")
		_ = c.MarkFlagRequired("zone")
		c.Flags().StringVar(&s.flagDNSRRTxnFile, "transaction-file", "transaction.yaml",
			"Local file used to stage the transaction")
	}
	for _, c := range []*cobra.Command{s.dnsRRTxnAddCmd, s.dnsRRTxnRemoveCmd} {
		c.Flags().StringVar(&s.flagDNSRRType, "type", "", "Record type (required)")
		_ = c.MarkFlagRequired("type")
		c.Flags().Int64Var(&s.flagDNSRRTTL, "ttl", 300, "TTL in seconds")
		c.Flags().StringSliceVar(&s.flagDNSRRRRDatas, "rrdatas", nil, "Record data (repeatable) (required)")
		_ = c.MarkFlagRequired("rrdatas")
	}
	s.dnsRRTxnExecuteCmd.Flags().StringVar(&s.flagDNSRRFormat, "format", "", "Output format")
	s.dnsRRTxnCmd.AddCommand(txnAll...)
	s.dnsRRCmd.AddCommand(s.dnsRRTxnCmd)

	s.dnsCmd.AddCommand(s.dnsRRCmd)
}

func (s *dnsCmds) dnsRRLoadOrBuild(name string) (*dns.ResourceRecordSet, error) {
	if s.flagDNSRRConfigFile != "" {
		body := &dns.ResourceRecordSet{}
		if err := loadYAMLOrJSONInto(s.flagDNSRRConfigFile, body); err != nil {
			return nil, err
		}
		return body, nil
	}
	return &dns.ResourceRecordSet{
		Name:    name,
		Type:    s.flagDNSRRType,
		Ttl:     s.flagDNSRRTTL,
		Rrdatas: s.flagDNSRRRRDatas,
	}, nil
}

func (s *dnsCmds) runDNSRRCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body, err := s.dnsRRLoadOrBuild(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ResourceRecordSets.Create(project, s.flagDNSRRZone, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating record set: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created record set [%s] type=%s.\n", args[0], body.Type)
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSRRFormat)
}

func (s *dnsCmds) runDNSRRDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.ResourceRecordSets.Delete(project, s.flagDNSRRZone, args[0], s.flagDNSRRType).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting record set: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted record set [%s] type=%s.\n", args[0], s.flagDNSRRType)
	return nil
}

func (s *dnsCmds) runDNSRRDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ResourceRecordSets.Get(project, s.flagDNSRRZone, args[0], s.flagDNSRRType).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing record set: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSRRFormat)
}

func (s *dnsCmds) runDNSRRList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	var all []*dns.ResourceRecordSet
	pageToken := ""
	for {
		call := svc.ResourceRecordSets.List(project, s.flagDNSRRZone).Context(ctx)
		if s.flagDNSRRName != "" {
			call = call.Name(s.flagDNSRRName)
		}
		if s.flagDNSRRType != "" {
			call = call.Type(s.flagDNSRRType)
		}
		if s.flagDNSRRMaxResults > 0 {
			call = call.MaxResults(s.flagDNSRRMaxResults)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagDNSRRFormat)
}

func (s *dnsCmds) runDNSRRUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body, err := s.dnsRRLoadOrBuild(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ResourceRecordSets.Patch(project, s.flagDNSRRZone, args[0], body.Type, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating record set: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated record set [%s] type=%s.\n", args[0], body.Type)
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSRRFormat)
}

func (s *dnsCmds) runDNSRRImport(cmd *cobra.Command, args []string) error {
	if s.flagDNSRRZoneFmt {
		return fmt.Errorf("--zone-file-format is not supported; provide YAML/JSON")
	}
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	var payload struct {
		Additions []*dns.ResourceRecordSet `json:"additions" yaml:"additions"`
		Deletions []*dns.ResourceRecordSet `json:"deletions" yaml:"deletions"`
	}
	if err := loadYAMLOrJSONInto(s.flagDNSRRImportFile, &payload); err != nil {
		return err
	}
	if len(payload.Additions) == 0 && len(payload.Deletions) == 0 {
		var rrs []*dns.ResourceRecordSet
		if err := loadYAMLOrJSONInto(s.flagDNSRRImportFile, &rrs); err != nil {
			return fmt.Errorf("source file must contain {additions:[...],deletions:[...]} or a list of ResourceRecordSet: %w", err)
		}
		payload.Additions = rrs
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Changes.Create(project, s.flagDNSRRZone, &dns.Change{
		Additions: payload.Additions,
		Deletions: payload.Deletions,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing record sets: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Applied %d additions and %d deletions (change ID: %s).\n",
		len(payload.Additions), len(payload.Deletions), got.Id)
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSRRFormat)
}

func (s *dnsCmds) runDNSRRExport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	var all []*dns.ResourceRecordSet
	pageToken := ""
	for {
		call := svc.ResourceRecordSets.List(project, s.flagDNSRRZone).Context(ctx)
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
	if err := os.WriteFile(filepath.Clean(s.flagDNSRRExportFile), buf, 0644); err != nil {
		return fmt.Errorf("writing export: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Wrote %d record sets to %s.\n", len(all), s.flagDNSRRExportFile)
	return nil
}

func (s *dnsCmds) runDNSRRChangesDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Changes.Get(project, s.flagDNSRRZone, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing change: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSRRFormat)
}

func (s *dnsCmds) runDNSRRChangesList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	var all []*dns.Change
	pageToken := ""
	for {
		call := svc.Changes.List(project, s.flagDNSRRZone).Context(ctx)
		if s.flagDNSRRMaxResults > 0 {
			call = call.MaxResults(s.flagDNSRRMaxResults)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagDNSRRFormat)
}

func (s *dnsCmds) dnsRRLoadTxn() (*dns.Change, error) {
	data, err := os.ReadFile(s.flagDNSRRTxnFile)
	if err != nil {
		return nil, fmt.Errorf("reading transaction file: %w", err)
	}
	change := &dns.Change{}
	if err := json.Unmarshal(data, change); err != nil {
		return nil, fmt.Errorf("parsing transaction file: %w", err)
	}
	return change, nil
}

func (s *dnsCmds) dnsRRSaveTxn(change *dns.Change) error {
	buf, err := json.MarshalIndent(change, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling transaction: %w", err)
	}
	return os.WriteFile(s.flagDNSRRTxnFile, buf, 0644)
}

func (s *dnsCmds) runDNSRRTxnStart(cmd *cobra.Command, args []string) error {
	if _, err := os.Stat(s.flagDNSRRTxnFile); err == nil {
		return fmt.Errorf("transaction file %s already exists; abort or execute first", s.flagDNSRRTxnFile)
	}
	return s.dnsRRSaveTxn(&dns.Change{})
}

func (s *dnsCmds) runDNSRRTxnAdd(cmd *cobra.Command, args []string) error {
	change, err := s.dnsRRLoadTxn()
	if err != nil {
		return err
	}
	change.Additions = append(change.Additions, &dns.ResourceRecordSet{
		Name:    args[0],
		Type:    s.flagDNSRRType,
		Ttl:     s.flagDNSRRTTL,
		Rrdatas: s.flagDNSRRRRDatas,
	})
	return s.dnsRRSaveTxn(change)
}

func (s *dnsCmds) runDNSRRTxnRemove(cmd *cobra.Command, args []string) error {
	change, err := s.dnsRRLoadTxn()
	if err != nil {
		return err
	}
	change.Deletions = append(change.Deletions, &dns.ResourceRecordSet{
		Name:    args[0],
		Type:    s.flagDNSRRType,
		Ttl:     s.flagDNSRRTTL,
		Rrdatas: s.flagDNSRRRRDatas,
	})
	return s.dnsRRSaveTxn(change)
}

func (s *dnsCmds) runDNSRRTxnDescribe(cmd *cobra.Command, args []string) error {
	change, err := s.dnsRRLoadTxn()
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, change, "")
}

func (s *dnsCmds) runDNSRRTxnExecute(cmd *cobra.Command, args []string) error {
	change, err := s.dnsRRLoadTxn()
	if err != nil {
		return err
	}
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Changes.Create(project, s.flagDNSRRZone, change).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("executing transaction: %w", err)
	}
	if err := os.Remove(s.flagDNSRRTxnFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing transaction file: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Executed transaction (change ID: %s).\n", got.Id)
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSRRFormat)
}

func (s *dnsCmds) runDNSRRTxnAbort(cmd *cobra.Command, args []string) error {
	if err := os.Remove(s.flagDNSRRTxnFile); err != nil {
		return fmt.Errorf("removing transaction file: %w", err)
	}
	fmt.Fprintln(s.inv.stdout, "Aborted transaction.")
	return nil
}
