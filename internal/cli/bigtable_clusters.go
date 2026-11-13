package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable clusters (#1304) ---

func (s *bigtableCmds) initBigtableClusters() {
	all := []*cobra.Command{
		s.bigtableClCreateCmd, s.bigtableClDeleteCmd, s.bigtableClDescribeCmd,
		s.bigtableClListCmd, s.bigtableClUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBtClInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&s.flagBtClFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.bigtableClCreateCmd, s.bigtableClUpdateCmd} {
		c.Flags().StringVar(&s.flagBtClConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Cluster body (required)")
		_ = c.MarkFlagRequired("config-file")
	}

	s.bigtableClustersCmd.AddCommand(all...)
	s.bigtableCmd.AddCommand(s.bigtableClustersCmd)
}

func (s *bigtableCmds) runBtClCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.btInstanceName(s.flagBtClInstance)
	if err != nil {
		return err
	}
	body := &bigtableadmin.Cluster{}
	if err := loadYAMLOrJSONInto(s.flagBtClConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Clusters.Create(parent, body).ClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBtClFormat)
}

func (s *bigtableCmds) runBtClDelete(cmd *cobra.Command, args []string) error {
	name, err := s.btClusterName(s.flagBtClInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Clusters.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted cluster [%s].\n", args[0])
	return nil
}

func (s *bigtableCmds) runBtClDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.btClusterName(s.flagBtClInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Clusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cluster: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBtClFormat)
}

func (s *bigtableCmds) runBtClList(cmd *cobra.Command, args []string) error {
	parent, err := s.btInstanceName(s.flagBtClInstance)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.Cluster
	pageToken := ""
	for {
		call := svc.Projects.Instances.Clusters.List(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing clusters: %w", err)
		}
		all = append(all, resp.Clusters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagBtClFormat)
}

func (s *bigtableCmds) runBtClUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.btClusterName(s.flagBtClInstance, args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.Cluster{}
	if err := loadYAMLOrJSONInto(s.flagBtClConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Clusters.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBtClFormat)
}
