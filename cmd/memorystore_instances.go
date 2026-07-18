package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud memorystore instances (#978) ---

var memstoreInstCmd = &cobra.Command{Use: "instances", Short: "Manage Memorystore instances"}

var (
	flagMemstoreInstLocation   string
	flagMemstoreInstFormat     string
	flagMemstoreInstConfigFile string
	flagMemstoreInstUpdateMask string
	flagMemstoreInstPageSize   int64
)

var (
	memstoreInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: runMemstoreInstCreate,
	}
	memstoreInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: runMemstoreInstDelete,
	}
	memstoreInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: runMemstoreInstDescribe,
	}
	memstoreInstListCmd = &cobra.Command{
		Use: "list", Short: "List Memorystore instances",
		Args: cobra.NoArgs, RunE: runMemstoreInstList,
	}
	memstoreInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: runMemstoreInstUpdate,
	}
	memstoreInstBackupCmd = &cobra.Command{
		Use: "backup INSTANCE", Short: "Trigger a backup for a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: runMemstoreInstBackup,
	}
	memstoreInstFinishMigrationCmd = &cobra.Command{
		Use: "finish-migration INSTANCE", Short: "Finish an in-progress Memorystore migration",
		Args: cobra.ExactArgs(1), RunE: runMemstoreInstFinishMigration,
	}
	memstoreInstStartMigrationCmd = &cobra.Command{
		Use: "start-migration INSTANCE", Short: "Start a Memorystore migration",
		Args: cobra.ExactArgs(1), RunE: runMemstoreInstStartMigration,
	}
	memstoreInstGetCACmd = &cobra.Command{
		Use: "get-certificate-authority INSTANCE", Short: "Get the certificate authority for a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: runMemstoreInstGetCA,
	}
	memstoreInstGetSharedCACmd = &cobra.Command{
		Use: "get-shared-regional-certificate-authority",
		Short: "Get the shared regional certificate authority for a Memorystore location",
		Args: cobra.NoArgs, RunE: runMemstoreInstGetSharedCA,
	}
	memstoreInstRescheduleMaintCmd = &cobra.Command{
		Use: "reschedule-maintenance INSTANCE", Short: "Reschedule maintenance for a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: runMemstoreInstRescheduleMaint,
	}
)

func init() {
	all := []*cobra.Command{
		memstoreInstCreateCmd, memstoreInstDeleteCmd, memstoreInstDescribeCmd, memstoreInstListCmd, memstoreInstUpdateCmd,
		memstoreInstBackupCmd, memstoreInstFinishMigrationCmd, memstoreInstStartMigrationCmd,
		memstoreInstGetCACmd, memstoreInstGetSharedCACmd, memstoreInstRescheduleMaintCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagMemstoreInstLocation, "location", "", "Memorystore location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagMemstoreInstFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{memstoreInstCreateCmd, memstoreInstUpdateCmd, memstoreInstStartMigrationCmd, memstoreInstRescheduleMaintCmd} {
		c.Flags().StringVar(&flagMemstoreInstConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	// backup and finish-migration accept optional config, defaulting to {}.
	memstoreInstBackupCmd.Flags().StringVar(&flagMemstoreInstConfigFile, "config-file", "",
		"Optional path to a YAML/JSON file with the backup request body")
	memstoreInstFinishMigrationCmd.Flags().StringVar(&flagMemstoreInstConfigFile, "config-file", "",
		"Optional path to a YAML/JSON file with the finish-migration request body")
	memstoreInstUpdateCmd.Flags().StringVar(&flagMemstoreInstUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	memstoreInstListCmd.Flags().Int64Var(&flagMemstoreInstPageSize, "page-size", 0, "Maximum results per page")

	memstoreInstCmd.AddCommand(all...)
	memorystoreCmd.AddCommand(memstoreInstCmd)
}

func memstoreInstParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagMemstoreInstLocation), nil
}

func memstoreInstName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := memstoreInstParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/instances/%s", parent, id), nil
}

// memstoreInstOptionalBody returns the parsed --config-file body, or {} if the flag
// was not set. Used by RPCs where the body is optional.
func memstoreInstOptionalBody() (map[string]any, error) {
	body := map[string]any{}
	if flagMemstoreInstConfigFile == "" {
		return body, nil
	}
	if err := loadYAMLOrJSONInto(flagMemstoreInstConfigFile, &body); err != nil {
		return nil, err
	}
	return body, nil
}

func runMemstoreInstCreate(cmd *cobra.Command, args []string) error {
	parent, err := memstoreInstParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagMemstoreInstConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("instanceId", args[0])
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodPost, "/"+parent+"/instances", q, body, &op); err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	fmt.Printf("Create request issued for instance [%s].\n", args[0])
	return emitFormatted(op, flagMemstoreInstFormat)
}

func runMemstoreInstDelete(cmd *cobra.Command, args []string) error {
	name, err := memstoreInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	fmt.Printf("Delete request issued for instance [%s].\n", args[0])
	return emitFormatted(op, flagMemstoreInstFormat)
}

func runMemstoreInstDescribe(cmd *cobra.Command, args []string) error {
	name, err := memstoreInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := memorystoreRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormatted(got, flagMemstoreInstFormat)
}

func runMemstoreInstList(cmd *cobra.Command, args []string) error {
	parent, err := memstoreInstParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := memorystoreRest.paginate(ctx, "/"+parent+"/instances", nil, "instances", flagMemstoreInstPageSize)
	if err != nil {
		return fmt.Errorf("listing instances: %w", err)
	}
	return emitFormatted(items, flagMemstoreInstFormat)
}

func runMemstoreInstUpdate(cmd *cobra.Command, args []string) error {
	name, err := memstoreInstName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagMemstoreInstConfigFile, &body); err != nil {
		return err
	}
	mask := flagMemstoreInstUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	q := url.Values{}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodPatch, "/"+name, q, body, &op); err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	fmt.Printf("Update request issued for instance [%s].\n", args[0])
	return emitFormatted(op, flagMemstoreInstFormat)
}

func runMemstoreInstBackup(cmd *cobra.Command, args []string) error {
	name, err := memstoreInstName(args[0])
	if err != nil {
		return err
	}
	body, err := memstoreInstOptionalBody()
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodPost, "/"+name+":backup", nil, body, &op); err != nil {
		return fmt.Errorf("backing up instance: %w", err)
	}
	fmt.Printf("Backup request issued for instance [%s].\n", args[0])
	return emitFormatted(op, flagMemstoreInstFormat)
}

func runMemstoreInstFinishMigration(cmd *cobra.Command, args []string) error {
	name, err := memstoreInstName(args[0])
	if err != nil {
		return err
	}
	body, err := memstoreInstOptionalBody()
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodPost, "/"+name+":finishMigration", nil, body, &op); err != nil {
		return fmt.Errorf("finishing migration for instance: %w", err)
	}
	fmt.Printf("Finish-migration request issued for instance [%s].\n", args[0])
	return emitFormatted(op, flagMemstoreInstFormat)
}

func runMemstoreInstStartMigration(cmd *cobra.Command, args []string) error {
	name, err := memstoreInstName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagMemstoreInstConfigFile, &body); err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodPost, "/"+name+":startMigration", nil, body, &op); err != nil {
		return fmt.Errorf("starting migration for instance: %w", err)
	}
	fmt.Printf("Start-migration request issued for instance [%s].\n", args[0])
	return emitFormatted(op, flagMemstoreInstFormat)
}

func runMemstoreInstGetCA(cmd *cobra.Command, args []string) error {
	name, err := memstoreInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := memorystoreRest.do(ctx, http.MethodGet, "/"+name+"/certificateAuthority", nil, nil, &got); err != nil {
		return fmt.Errorf("getting certificate authority: %w", err)
	}
	return emitFormatted(got, flagMemstoreInstFormat)
}

func runMemstoreInstGetSharedCA(cmd *cobra.Command, args []string) error {
	parent, err := memstoreInstParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := memorystoreRest.do(ctx, http.MethodGet, "/"+parent+":getSharedRegionalCertificateAuthority", nil, nil, &got); err != nil {
		return fmt.Errorf("getting shared regional certificate authority: %w", err)
	}
	return emitFormatted(got, flagMemstoreInstFormat)
}

func runMemstoreInstRescheduleMaint(cmd *cobra.Command, args []string) error {
	name, err := memstoreInstName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagMemstoreInstConfigFile, &body); err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodPost, "/"+name+":rescheduleMaintenance", nil, body, &op); err != nil {
		return fmt.Errorf("rescheduling maintenance: %w", err)
	}
	fmt.Printf("Reschedule-maintenance request issued for instance [%s].\n", args[0])
	return emitFormatted(op, flagMemstoreInstFormat)
}
