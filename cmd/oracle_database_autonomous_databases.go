package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database autonomous-databases (#1263) ---

func (s *oracleDatabaseCmds) initOracleDatabaseAutonomousDatabases() {
	all := []*cobra.Command{
		s.oracleAdbCreateCmd, s.oracleAdbDeleteCmd, s.oracleAdbDescribeCmd,
		s.oracleAdbListCmd, s.oracleAdbUpdateCmd, s.oracleAdbFailoverCmd,
		s.oracleAdbGenerateWalletCmd, s.oracleAdbRestartCmd, s.oracleAdbRestoreCmd,
		s.oracleAdbStartCmd, s.oracleAdbStopCmd, s.oracleAdbSwitchoverCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbAdbLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbAdbFormat, "format", "", "Output format")
	}
	s.oracleAdbCreateCmd.Flags().StringVar(&s.flagOdbAdbConfigFile, "config-file", "", "YAML/JSON file with the AutonomousDatabase body (required)")
	_ = s.oracleAdbCreateCmd.MarkFlagRequired("config-file")
	s.oracleAdbUpdateCmd.Flags().StringVar(&s.flagOdbAdbConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.oracleAdbUpdateCmd.MarkFlagRequired("config-file")
	s.oracleAdbUpdateCmd.Flags().StringVar(&s.flagOdbAdbUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	s.oracleAdbListCmd.Flags().StringVar(&s.flagOdbAdbFilter, "filter", "", "Server-side filter expression")
	s.oracleAdbListCmd.Flags().Int64Var(&s.flagOdbAdbPageSize, "page-size", 0, "Maximum results per page")
	for _, c := range []*cobra.Command{s.oracleAdbGenerateWalletCmd, s.oracleAdbRestoreCmd, s.oracleAdbSwitchoverCmd} {
		c.Flags().StringVar(&s.flagOdbAdbConfigFile, "config-file", "", "YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}

	s.oracleAutonomousDatabasesCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleAutonomousDatabasesCmd)
}

func (s *oracleDatabaseCmds) oracleAdbName(id string) (string, error) {
	return s.odbResource(s.flagOdbAdbLocation, "autonomousDatabases", id)
}

func (s *oracleDatabaseCmds) runOracleAdbCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbAdbLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.AutonomousDatabase{}
	if err := loadYAMLOrJSONInto(s.flagOdbAdbConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Create(parent, body).AutonomousDatabaseId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating autonomous database: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbAdbFormat)
}

func (s *oracleDatabaseCmds) runOracleAdbDelete(cmd *cobra.Command, args []string) error {
	name, err := s.oracleAdbName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting autonomous database: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbAdbFormat)
}

func (s *oracleDatabaseCmds) runOracleAdbDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.oracleAdbName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AutonomousDatabases.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing autonomous database: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagOdbAdbFormat)
}

func (s *oracleDatabaseCmds) runOracleAdbList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbAdbLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.AutonomousDatabase
	pageToken := ""
	for {
		call := svc.Projects.Locations.AutonomousDatabases.List(parent).Context(ctx)
		if s.flagOdbAdbFilter != "" {
			call = call.Filter(s.flagOdbAdbFilter)
		}
		if s.flagOdbAdbPageSize > 0 {
			call = call.PageSize(s.flagOdbAdbPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing autonomous databases: %w", err)
		}
		all = append(all, resp.AutonomousDatabases...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbAdbFormat)
}

func (s *oracleDatabaseCmds) runOracleAdbUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.oracleAdbName(args[0])
	if err != nil {
		return err
	}
	body := &oracledatabase.AutonomousDatabase{}
	if err := loadYAMLOrJSONInto(s.flagOdbAdbConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagOdbAdbUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.AutonomousDatabases.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating autonomous database: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbAdbFormat)
}

func (s *oracleDatabaseCmds) runOracleAdbFailover(cmd *cobra.Command, args []string) error {
	name, err := s.oracleAdbName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Failover(name, &oracledatabase.FailoverAutonomousDatabaseRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failing over autonomous database: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Failover autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbAdbFormat)
}

func (s *oracleDatabaseCmds) runOracleAdbGenerateWallet(cmd *cobra.Command, args []string) error {
	name, err := s.oracleAdbName(args[0])
	if err != nil {
		return err
	}
	body := &oracledatabase.GenerateAutonomousDatabaseWalletRequest{}
	if err := loadYAMLOrJSONInto(s.flagOdbAdbConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.AutonomousDatabases.GenerateWallet(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("generating wallet: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagOdbAdbFormat)
}

func (s *oracleDatabaseCmds) runOracleAdbRestart(cmd *cobra.Command, args []string) error {
	name, err := s.oracleAdbName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Restart(name, &oracledatabase.RestartAutonomousDatabaseRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restarting autonomous database: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Restart autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbAdbFormat)
}

func (s *oracleDatabaseCmds) runOracleAdbRestore(cmd *cobra.Command, args []string) error {
	name, err := s.oracleAdbName(args[0])
	if err != nil {
		return err
	}
	body := &oracledatabase.RestoreAutonomousDatabaseRequest{}
	if err := loadYAMLOrJSONInto(s.flagOdbAdbConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Restore(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring autonomous database: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Restore autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbAdbFormat)
}

func (s *oracleDatabaseCmds) runOracleAdbStart(cmd *cobra.Command, args []string) error {
	name, err := s.oracleAdbName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Start(name, &oracledatabase.StartAutonomousDatabaseRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting autonomous database: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Start autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbAdbFormat)
}

func (s *oracleDatabaseCmds) runOracleAdbStop(cmd *cobra.Command, args []string) error {
	name, err := s.oracleAdbName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Stop(name, &oracledatabase.StopAutonomousDatabaseRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping autonomous database: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Stop autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbAdbFormat)
}

func (s *oracleDatabaseCmds) runOracleAdbSwitchover(cmd *cobra.Command, args []string) error {
	name, err := s.oracleAdbName(args[0])
	if err != nil {
		return err
	}
	body := &oracledatabase.SwitchoverAutonomousDatabaseRequest{}
	if err := loadYAMLOrJSONInto(s.flagOdbAdbConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Switchover(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("switching over autonomous database: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Switchover autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbAdbFormat)
}
