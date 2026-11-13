package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database cloud-vm-clusters (#1265) ---

func (s *oracleDatabaseCmds) initOracleDatabaseCloudVmClusters() {
	all := []*cobra.Command{
		s.oracleVmClCreateCmd, s.oracleVmClDeleteCmd,
		s.oracleVmClDescribeCmd, s.oracleVmClListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbVmClLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbVmClFormat, "format", "", "Output format")
	}
	s.oracleVmClCreateCmd.Flags().StringVar(&s.flagOdbVmClConfigFile, "config-file", "", "YAML/JSON file with cloud VM cluster body (required)")
	_ = s.oracleVmClCreateCmd.MarkFlagRequired("config-file")
	s.oracleVmClListCmd.Flags().Int64Var(&s.flagOdbVmClPageSize, "page-size", 0, "Maximum results per page")
	s.oracleVmClListCmd.Flags().StringVar(&s.flagOdbVmClFilter, "filter", "", "Server-side filter expression")

	s.oracleCloudVmClustersCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleCloudVmClustersCmd)
}

func (s *oracleDatabaseCmds) oracleVmClName(id string) (string, error) {
	return s.odbResource(s.flagOdbVmClLocation, "cloudVmClusters", id)
}

func (s *oracleDatabaseCmds) runOracleVmClCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbVmClLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.CloudVmCluster{}
	if err := loadYAMLOrJSONInto(s.flagOdbVmClConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CloudVmClusters.Create(parent, body).CloudVmClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cloud VM cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create cloud VM cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbVmClFormat)
}

func (s *oracleDatabaseCmds) runOracleVmClDelete(cmd *cobra.Command, args []string) error {
	name, err := s.oracleVmClName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CloudVmClusters.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cloud VM cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete cloud VM cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbVmClFormat)
}

func (s *oracleDatabaseCmds) runOracleVmClDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.oracleVmClName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CloudVmClusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cloud VM cluster: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagOdbVmClFormat)
}

func (s *oracleDatabaseCmds) runOracleVmClList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbVmClLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.CloudVmCluster
	pageToken := ""
	for {
		call := svc.Projects.Locations.CloudVmClusters.List(parent).Context(ctx)
		if s.flagOdbVmClPageSize > 0 {
			call = call.PageSize(s.flagOdbVmClPageSize)
		}
		if s.flagOdbVmClFilter != "" {
			call = call.Filter(s.flagOdbVmClFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing cloud VM clusters: %w", err)
		}
		all = append(all, resp.CloudVmClusters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbVmClFormat)
}
