package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database autonomous-databases (#1263) ---

var oracleAutonomousDatabasesCmd = &cobra.Command{Use: "autonomous-databases", Short: "Manage Oracle Autonomous Databases"}

var (
	flagOdbAdbLocation   string
	flagOdbAdbFormat     string
	flagOdbAdbConfigFile string
	flagOdbAdbUpdateMask string
	flagOdbAdbPageSize   int64
	flagOdbAdbFilter     string
)

var (
	oracleAdbCreateCmd = &cobra.Command{
		Use: "create AUTONOMOUS_DATABASE", Short: "Create an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: runOracleAdbCreate,
	}
	oracleAdbDeleteCmd = &cobra.Command{
		Use: "delete AUTONOMOUS_DATABASE", Short: "Delete an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: runOracleAdbDelete,
	}
	oracleAdbDescribeCmd = &cobra.Command{
		Use: "describe AUTONOMOUS_DATABASE", Short: "Describe an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: runOracleAdbDescribe,
	}
	oracleAdbListCmd = &cobra.Command{
		Use: "list", Short: "List Autonomous Databases in a location",
		Args: cobra.NoArgs, RunE: runOracleAdbList,
	}
	oracleAdbUpdateCmd = &cobra.Command{
		Use: "update AUTONOMOUS_DATABASE", Short: "Update an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: runOracleAdbUpdate,
	}
	oracleAdbFailoverCmd = &cobra.Command{
		Use: "failover AUTONOMOUS_DATABASE", Short: "Failover an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: runOracleAdbFailover,
	}
	oracleAdbGenerateWalletCmd = &cobra.Command{
		Use: "generate-wallet AUTONOMOUS_DATABASE", Short: "Generate a wallet for an Autonomous Database (loads request from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runOracleAdbGenerateWallet,
	}
	oracleAdbRestartCmd = &cobra.Command{
		Use: "restart AUTONOMOUS_DATABASE", Short: "Restart an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: runOracleAdbRestart,
	}
	oracleAdbRestoreCmd = &cobra.Command{
		Use: "restore AUTONOMOUS_DATABASE", Short: "Restore an Autonomous Database from a backup (loads request from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runOracleAdbRestore,
	}
	oracleAdbStartCmd = &cobra.Command{
		Use: "start AUTONOMOUS_DATABASE", Short: "Start an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: runOracleAdbStart,
	}
	oracleAdbStopCmd = &cobra.Command{
		Use: "stop AUTONOMOUS_DATABASE", Short: "Stop an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: runOracleAdbStop,
	}
	oracleAdbSwitchoverCmd = &cobra.Command{
		Use: "switchover AUTONOMOUS_DATABASE", Short: "Switchover an Autonomous Database (loads request from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runOracleAdbSwitchover,
	}
)

func init() {
	all := []*cobra.Command{
		oracleAdbCreateCmd, oracleAdbDeleteCmd, oracleAdbDescribeCmd,
		oracleAdbListCmd, oracleAdbUpdateCmd, oracleAdbFailoverCmd,
		oracleAdbGenerateWalletCmd, oracleAdbRestartCmd, oracleAdbRestoreCmd,
		oracleAdbStartCmd, oracleAdbStopCmd, oracleAdbSwitchoverCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbAdbLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbAdbFormat, "format", "", "Output format")
	}
	oracleAdbCreateCmd.Flags().StringVar(&flagOdbAdbConfigFile, "config-file", "", "YAML/JSON file with the AutonomousDatabase body (required)")
	_ = oracleAdbCreateCmd.MarkFlagRequired("config-file")
	oracleAdbUpdateCmd.Flags().StringVar(&flagOdbAdbConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = oracleAdbUpdateCmd.MarkFlagRequired("config-file")
	oracleAdbUpdateCmd.Flags().StringVar(&flagOdbAdbUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	oracleAdbListCmd.Flags().StringVar(&flagOdbAdbFilter, "filter", "", "Server-side filter expression")
	oracleAdbListCmd.Flags().Int64Var(&flagOdbAdbPageSize, "page-size", 0, "Maximum results per page")
	for _, c := range []*cobra.Command{oracleAdbGenerateWalletCmd, oracleAdbRestoreCmd, oracleAdbSwitchoverCmd} {
		c.Flags().StringVar(&flagOdbAdbConfigFile, "config-file", "", "YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}

	oracleAutonomousDatabasesCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleAutonomousDatabasesCmd)
}

func oracleAdbName(id string) (string, error) {
	return odbResource(flagOdbAdbLocation, "autonomousDatabases", id)
}

func runOracleAdbCreate(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbAdbLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.AutonomousDatabase{}
	if err := loadYAMLOrJSONInto(flagOdbAdbConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Create(parent, body).AutonomousDatabaseId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating autonomous database: %w", err)
	}
	fmt.Printf("Create autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbAdbFormat)
}

func runOracleAdbDelete(cmd *cobra.Command, args []string) error {
	name, err := oracleAdbName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting autonomous database: %w", err)
	}
	fmt.Printf("Delete autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbAdbFormat)
}

func runOracleAdbDescribe(cmd *cobra.Command, args []string) error {
	name, err := oracleAdbName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AutonomousDatabases.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing autonomous database: %w", err)
	}
	return emitFormatted(got, flagOdbAdbFormat)
}

func runOracleAdbList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbAdbLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.AutonomousDatabase
	pageToken := ""
	for {
		call := svc.Projects.Locations.AutonomousDatabases.List(parent).Context(ctx)
		if flagOdbAdbFilter != "" {
			call = call.Filter(flagOdbAdbFilter)
		}
		if flagOdbAdbPageSize > 0 {
			call = call.PageSize(flagOdbAdbPageSize)
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
	return emitFormatted(all, flagOdbAdbFormat)
}

func runOracleAdbUpdate(cmd *cobra.Command, args []string) error {
	name, err := oracleAdbName(args[0])
	if err != nil {
		return err
	}
	body := &oracledatabase.AutonomousDatabase{}
	if err := loadYAMLOrJSONInto(flagOdbAdbConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagOdbAdbUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
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
	fmt.Printf("Update autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbAdbFormat)
}

func runOracleAdbFailover(cmd *cobra.Command, args []string) error {
	name, err := oracleAdbName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Failover(name, &oracledatabase.FailoverAutonomousDatabaseRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failing over autonomous database: %w", err)
	}
	fmt.Printf("Failover autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbAdbFormat)
}

func runOracleAdbGenerateWallet(cmd *cobra.Command, args []string) error {
	name, err := oracleAdbName(args[0])
	if err != nil {
		return err
	}
	body := &oracledatabase.GenerateAutonomousDatabaseWalletRequest{}
	if err := loadYAMLOrJSONInto(flagOdbAdbConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.AutonomousDatabases.GenerateWallet(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("generating wallet: %w", err)
	}
	return emitFormatted(resp, flagOdbAdbFormat)
}

func runOracleAdbRestart(cmd *cobra.Command, args []string) error {
	name, err := oracleAdbName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Restart(name, &oracledatabase.RestartAutonomousDatabaseRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restarting autonomous database: %w", err)
	}
	fmt.Printf("Restart autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbAdbFormat)
}

func runOracleAdbRestore(cmd *cobra.Command, args []string) error {
	name, err := oracleAdbName(args[0])
	if err != nil {
		return err
	}
	body := &oracledatabase.RestoreAutonomousDatabaseRequest{}
	if err := loadYAMLOrJSONInto(flagOdbAdbConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Restore(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring autonomous database: %w", err)
	}
	fmt.Printf("Restore autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbAdbFormat)
}

func runOracleAdbStart(cmd *cobra.Command, args []string) error {
	name, err := oracleAdbName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Start(name, &oracledatabase.StartAutonomousDatabaseRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting autonomous database: %w", err)
	}
	fmt.Printf("Start autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbAdbFormat)
}

func runOracleAdbStop(cmd *cobra.Command, args []string) error {
	name, err := oracleAdbName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Stop(name, &oracledatabase.StopAutonomousDatabaseRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping autonomous database: %w", err)
	}
	fmt.Printf("Stop autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbAdbFormat)
}

func runOracleAdbSwitchover(cmd *cobra.Command, args []string) error {
	name, err := oracleAdbName(args[0])
	if err != nil {
		return err
	}
	body := &oracledatabase.SwitchoverAutonomousDatabaseRequest{}
	if err := loadYAMLOrJSONInto(flagOdbAdbConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AutonomousDatabases.Switchover(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("switching over autonomous database: %w", err)
	}
	fmt.Printf("Switchover autonomous database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbAdbFormat)
}
