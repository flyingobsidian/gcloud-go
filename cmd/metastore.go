package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	metastore "google.golang.org/api/metastore/v1"
)

// --- gcloud metastore (#356) ---

var metastoreCmd = &cobra.Command{Use: "metastore", Short: "Manage Dataproc Metastore"}

func msLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func msChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func msWaitOp(ctx context.Context, svc *metastore.APIService, op *metastore.Operation) (*metastore.Operation, error) {
	for !op.Done {
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func msFinishOp(ctx context.Context, svc *metastore.APIService, op *metastore.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := msWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

var (
	flagMSLocation   string
	flagMSConfigFile string
	flagMSUpdateMask string
	flagMSFormat     string
	flagMSAsync      bool
	flagMSService    string
	flagMSIamMember  string
	flagMSIamRole    string
	flagMSBackup     string

	// alter-location
	flagMSResourceName string
	flagMSLocationURI  string
	// alter-table-properties
	flagMSTable      string
	flagMSProperties map[string]string
	// move-table
	flagMSDBSource string
	flagMSDBDest   string
	flagMSTblName  string
	// query-metadata
	flagMSQuery string
	// export
	flagMSDestGCS string
	flagMSExportDB string
	flagMSDBType  string
	// import
	flagMSImportGCS string
	flagMSImportDesc string
	flagMSImportDBType string
	// restore
	flagMSBackupName    string
	flagMSRestoreType   string
	flagMSRestoreServ string
)

// --- federations ---

var metastoreFederationsCmd = &cobra.Command{Use: "federations", Short: "Manage Metastore federations"}

var (
	msFedCreateCmd = &cobra.Command{
		Use: "create FEDERATION", Short: "Create a federation from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMSFedCreate,
	}
	msFedDeleteCmd = &cobra.Command{
		Use: "delete FEDERATION", Short: "Delete a federation",
		Args: cobra.ExactArgs(1), RunE: runMSFedDelete,
	}
	msFedDescribeCmd = &cobra.Command{
		Use: "describe FEDERATION", Short: "Describe a federation",
		Args: cobra.ExactArgs(1), RunE: runMSFedDescribe,
	}
	msFedListCmd = &cobra.Command{
		Use: "list", Short: "List federations in a location",
		Args: cobra.NoArgs, RunE: runMSFedList,
	}
	msFedUpdateCmd = &cobra.Command{
		Use: "update FEDERATION", Short: "Update a federation from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMSFedUpdate,
	}
	msFedGetIamCmd = &cobra.Command{
		Use: "get-iam-policy FEDERATION", Short: "Get the IAM policy for a federation",
		Args: cobra.ExactArgs(1), RunE: runMSFedGetIam,
	}
	msFedSetIamCmd = &cobra.Command{
		Use: "set-iam-policy FEDERATION POLICY_FILE", Short: "Replace the IAM policy",
		Args: cobra.ExactArgs(2), RunE: runMSFedSetIam,
	}
	msFedAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding FEDERATION", Short: "Add an IAM binding",
		Args: cobra.ExactArgs(1), RunE: runMSFedAddIam,
	}
	msFedRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding FEDERATION", Short: "Remove an IAM binding",
		Args: cobra.ExactArgs(1), RunE: runMSFedRemoveIam,
	}
)

// --- locations ---

var metastoreLocationsCmd = &cobra.Command{Use: "locations", Short: "Manage Metastore locations"}

var (
	msLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Metastore location",
		Args: cobra.ExactArgs(1), RunE: runMSLocDescribe,
	}
	msLocListCmd = &cobra.Command{
		Use: "list", Short: "List Metastore locations",
		Args: cobra.NoArgs, RunE: runMSLocList,
	}
)

// --- operations ---

var metastoreOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Metastore operations"}

var (
	msOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: runMSOpCancel,
	}
	msOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an operation",
		Args: cobra.ExactArgs(1), RunE: runMSOpDelete,
	}
	msOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: runMSOpDescribe,
	}
	msOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a location",
		Args: cobra.NoArgs, RunE: runMSOpList,
	}
	msOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for an operation to complete",
		Args: cobra.ExactArgs(1), RunE: runMSOpWait,
	}
)

// --- services ---

var metastoreServicesCmd = &cobra.Command{Use: "services", Short: "Manage Metastore services"}

var (
	msSvcCreateCmd = &cobra.Command{
		Use: "create SERVICE", Short: "Create a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMSSvcCreate,
	}
	msSvcDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete a service",
		Args: cobra.ExactArgs(1), RunE: runMSSvcDelete,
	}
	msSvcDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a service",
		Args: cobra.ExactArgs(1), RunE: runMSSvcDescribe,
	}
	msSvcListCmd = &cobra.Command{
		Use: "list", Short: "List services in a location",
		Args: cobra.NoArgs, RunE: runMSSvcList,
	}
	msSvcUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMSSvcUpdate,
	}
	msSvcGetIamCmd = &cobra.Command{
		Use: "get-iam-policy SERVICE", Short: "Get the IAM policy for a service",
		Args: cobra.ExactArgs(1), RunE: runMSSvcGetIam,
	}
	msSvcSetIamCmd = &cobra.Command{
		Use: "set-iam-policy SERVICE POLICY_FILE", Short: "Replace the IAM policy",
		Args: cobra.ExactArgs(2), RunE: runMSSvcSetIam,
	}
	msSvcAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding SERVICE", Short: "Add an IAM binding",
		Args: cobra.ExactArgs(1), RunE: runMSSvcAddIam,
	}
	msSvcRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding SERVICE", Short: "Remove an IAM binding",
		Args: cobra.ExactArgs(1), RunE: runMSSvcRemoveIam,
	}
	msSvcAlterLocCmd = &cobra.Command{
		Use: "alter-metadata-resource-location SERVICE", Short: "Alter a metadata resource location",
		Args: cobra.ExactArgs(1), RunE: runMSSvcAlterLoc,
	}
	msSvcAlterTblCmd = &cobra.Command{
		Use: "alter-table-properties SERVICE", Short: "Alter table properties",
		Args: cobra.ExactArgs(1), RunE: runMSSvcAlterTbl,
	}
	msSvcMoveTblCmd = &cobra.Command{
		Use: "move-table-to-database SERVICE", Short: "Move a table between databases",
		Args: cobra.ExactArgs(1), RunE: runMSSvcMoveTbl,
	}
	msSvcQueryCmd = &cobra.Command{
		Use: "query-metadata SERVICE", Short: "Run a metadata query",
		Args: cobra.ExactArgs(1), RunE: runMSSvcQuery,
	}
	msSvcExportCmd = &cobra.Command{
		Use: "export SERVICE", Short: "Export metadata to Cloud Storage",
		Args: cobra.ExactArgs(1), RunE: runMSSvcExport,
	}
	msSvcImportCmd = &cobra.Command{
		Use: "import SERVICE IMPORT_ID", Short: "Import metadata from Cloud Storage",
		Args: cobra.ExactArgs(2), RunE: runMSSvcImport,
	}
	msSvcRestoreCmd = &cobra.Command{
		Use: "restore SERVICE", Short: "Restore a service from a backup",
		Args: cobra.ExactArgs(1), RunE: runMSSvcRestore,
	}
)

// backups (nested)

var metastoreServicesBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage Metastore service backups"}

var (
	msBackupCreateCmd = &cobra.Command{
		Use: "create BACKUP", Short: "Create a backup of a service",
		Args: cobra.ExactArgs(1), RunE: runMSBackupCreate,
	}
	msBackupDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a service backup",
		Args: cobra.ExactArgs(1), RunE: runMSBackupDelete,
	}
	msBackupDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a service backup",
		Args: cobra.ExactArgs(1), RunE: runMSBackupDescribe,
	}
	msBackupListCmd = &cobra.Command{
		Use: "list", Short: "List service backups",
		Args: cobra.NoArgs, RunE: runMSBackupList,
	}
)

func init() {
	// federations
	fedAll := []*cobra.Command{msFedCreateCmd, msFedDeleteCmd, msFedDescribeCmd, msFedListCmd, msFedUpdateCmd,
		msFedGetIamCmd, msFedSetIamCmd, msFedAddIamCmd, msFedRemoveIamCmd}
	for _, c := range fedAll {
		c.Flags().StringVar(&flagMSLocation, "location", "", "Location containing the federation (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{msFedCreateCmd, msFedUpdateCmd} {
		c.Flags().StringVar(&flagMSConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Federation body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	msFedUpdateCmd.Flags().StringVar(&flagMSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{msFedCreateCmd, msFedDeleteCmd, msFedUpdateCmd} {
		c.Flags().BoolVar(&flagMSAsync, "async", false, "Return the LRO without waiting")
	}
	for _, c := range []*cobra.Command{msFedDescribeCmd, msFedListCmd, msFedGetIamCmd} {
		c.Flags().StringVar(&flagMSFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{msFedAddIamCmd, msFedRemoveIamCmd} {
		c.Flags().StringVar(&flagMSIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&flagMSIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	metastoreFederationsCmd.AddCommand(fedAll...)
	metastoreCmd.AddCommand(metastoreFederationsCmd)

	// locations
	msLocDescribeCmd.Flags().StringVar(&flagMSFormat, "format", "", "Output format")
	msLocListCmd.Flags().StringVar(&flagMSFormat, "format", "", "Output format")
	metastoreLocationsCmd.AddCommand(msLocDescribeCmd, msLocListCmd)
	metastoreCmd.AddCommand(metastoreLocationsCmd)

	// operations
	opAll := []*cobra.Command{msOpCancelCmd, msOpDeleteCmd, msOpDescribeCmd, msOpListCmd, msOpWaitCmd}
	for _, c := range opAll {
		c.Flags().StringVar(&flagMSLocation, "location", "", "Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{msOpDescribeCmd, msOpListCmd, msOpWaitCmd} {
		c.Flags().StringVar(&flagMSFormat, "format", "", "Output format")
	}
	metastoreOperationsCmd.AddCommand(opAll...)
	metastoreCmd.AddCommand(metastoreOperationsCmd)

	// services
	svcAll := []*cobra.Command{msSvcCreateCmd, msSvcDeleteCmd, msSvcDescribeCmd, msSvcListCmd, msSvcUpdateCmd,
		msSvcGetIamCmd, msSvcSetIamCmd, msSvcAddIamCmd, msSvcRemoveIamCmd,
		msSvcAlterLocCmd, msSvcAlterTblCmd, msSvcMoveTblCmd, msSvcQueryCmd,
		msSvcExportCmd, msSvcImportCmd, msSvcRestoreCmd}
	for _, c := range svcAll {
		c.Flags().StringVar(&flagMSLocation, "location", "", "Location containing the service (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{msSvcCreateCmd, msSvcUpdateCmd} {
		c.Flags().StringVar(&flagMSConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Service body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	msSvcUpdateCmd.Flags().StringVar(&flagMSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{msSvcCreateCmd, msSvcDeleteCmd, msSvcUpdateCmd,
		msSvcAlterLocCmd, msSvcAlterTblCmd, msSvcMoveTblCmd,
		msSvcExportCmd, msSvcImportCmd, msSvcRestoreCmd} {
		c.Flags().BoolVar(&flagMSAsync, "async", false, "Return the LRO without waiting")
	}
	for _, c := range []*cobra.Command{msSvcDescribeCmd, msSvcListCmd, msSvcGetIamCmd, msSvcQueryCmd} {
		c.Flags().StringVar(&flagMSFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{msSvcAddIamCmd, msSvcRemoveIamCmd} {
		c.Flags().StringVar(&flagMSIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&flagMSIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	msSvcAlterLocCmd.Flags().StringVar(&flagMSResourceName, "resource-name", "", "Metadata resource name (required)")
	msSvcAlterLocCmd.Flags().StringVar(&flagMSLocationURI, "location-uri", "", "New location URI (required)")
	_ = msSvcAlterLocCmd.MarkFlagRequired("resource-name")
	_ = msSvcAlterLocCmd.MarkFlagRequired("location-uri")
	msSvcAlterTblCmd.Flags().StringVar(&flagMSTable, "table", "", "Fully qualified table name (required)")
	msSvcAlterTblCmd.Flags().StringToStringVar(&flagMSProperties, "properties", nil, "Table properties (key=value); may repeat")
	_ = msSvcAlterTblCmd.MarkFlagRequired("table")
	msSvcMoveTblCmd.Flags().StringVar(&flagMSDBSource, "source-database", "", "Source database (required)")
	msSvcMoveTblCmd.Flags().StringVar(&flagMSDBDest, "destination-database", "", "Destination database (required)")
	msSvcMoveTblCmd.Flags().StringVar(&flagMSTblName, "table-name", "", "Table name (required)")
	_ = msSvcMoveTblCmd.MarkFlagRequired("source-database")
	_ = msSvcMoveTblCmd.MarkFlagRequired("destination-database")
	_ = msSvcMoveTblCmd.MarkFlagRequired("table-name")
	msSvcQueryCmd.Flags().StringVar(&flagMSQuery, "query", "", "HQL query to run (required)")
	_ = msSvcQueryCmd.MarkFlagRequired("query")
	msSvcExportCmd.Flags().StringVar(&flagMSDestGCS, "destination-gcs-folder", "", "gs:// destination folder (required)")
	msSvcExportCmd.Flags().StringVar(&flagMSExportDB, "database-dump-type", "MYSQL", "Database dump type")
	_ = msSvcExportCmd.MarkFlagRequired("destination-gcs-folder")
	msSvcImportCmd.Flags().StringVar(&flagMSImportGCS, "database-dump", "", "gs:// URI of the dump (required)")
	msSvcImportCmd.Flags().StringVar(&flagMSImportDesc, "description", "", "Description for the import")
	msSvcImportCmd.Flags().StringVar(&flagMSImportDBType, "database-type", "MYSQL", "Database type of the dump")
	_ = msSvcImportCmd.MarkFlagRequired("database-dump")
	msSvcRestoreCmd.Flags().StringVar(&flagMSBackupName, "backup", "", "Backup name to restore from (required)")
	msSvcRestoreCmd.Flags().StringVar(&flagMSRestoreType, "restore-type", "FULL", "Restore type (FULL or METADATA_ONLY)")
	msSvcRestoreCmd.Flags().StringVar(&flagMSRestoreServ, "source-service", "", "Source service (required if backup path is not fully-qualified)")
	_ = msSvcRestoreCmd.MarkFlagRequired("backup")
	metastoreServicesCmd.AddCommand(svcAll...)

	// backups (nested)
	bkAll := []*cobra.Command{msBackupCreateCmd, msBackupDeleteCmd, msBackupDescribeCmd, msBackupListCmd}
	for _, c := range bkAll {
		c.Flags().StringVar(&flagMSLocation, "location", "", "Location of the parent service (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagMSService, "service", "", "Parent service (required)")
		_ = c.MarkFlagRequired("service")
	}
	for _, c := range []*cobra.Command{msBackupCreateCmd, msBackupDeleteCmd} {
		c.Flags().BoolVar(&flagMSAsync, "async", false, "Return the LRO without waiting")
	}
	msBackupCreateCmd.Flags().StringVar(&flagMSConfigFile, "config-file", "",
		"Path to an optional JSON/YAML file with additional Backup body fields")
	for _, c := range []*cobra.Command{msBackupDescribeCmd, msBackupListCmd} {
		c.Flags().StringVar(&flagMSFormat, "format", "", "Output format")
	}
	metastoreServicesBackupsCmd.AddCommand(bkAll...)
	metastoreServicesCmd.AddCommand(metastoreServicesBackupsCmd)

	metastoreCmd.AddCommand(metastoreServicesCmd)

	rootCmd.AddCommand(metastoreCmd)
}

// --- federations impl ---

func msFedName(id, project, location string) string {
	return msChild("federations", id, msLocationParent(project, location))
}

func runMSFedCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	f := &metastore.Federation{}
	if err := loadYAMLOrJSONInto(flagMSConfigFile, f); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Federations.Create(msLocationParent(project, flagMSLocation), f).
		FederationId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating federation: %w", err)
	}
	return msFinishOp(ctx, svc, op, "Create federation", args[0], flagMSAsync)
}

func runMSFedDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Federations.Delete(msFedName(args[0], project, flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting federation: %w", err)
	}
	return msFinishOp(ctx, svc, op, "Delete federation", args[0], flagMSAsync)
}

func runMSFedDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Federations.Get(msFedName(args[0], project, flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing federation: %w", err)
	}
	return emitFormatted(got, flagMSFormat)
}

func runMSFedList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Federations.List(msLocationParent(project, flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing federations: %w", err)
	}
	if flagMSFormat != "" {
		return emitFormatted(resp.Federations, flagMSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, f := range resp.Federations {
		fmt.Printf("%-40s %s\n", path.Base(f.Name), f.State)
	}
	return nil
}

func runMSFedUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	f := &metastore.Federation{}
	if err := loadYAMLOrJSONInto(flagMSConfigFile, f); err != nil {
		return err
	}
	mask := flagMSUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(f))
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Federations.Patch(msFedName(args[0], project, flagMSLocation), f).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating federation: %w", err)
	}
	return msFinishOp(ctx, svc, op, "Update federation", args[0], flagMSAsync)
}

func runMSFedGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Federations.GetIamPolicy(msFedName(args[0], project, flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagMSFormat)
}

func runMSFedSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &metastore.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Federations.SetIamPolicy(msFedName(args[0], project, flagMSLocation),
		&metastore.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runMSFedAddIam(cmd *cobra.Command, args []string) error {
	return msFedModifyIam(args[0], func(p *metastore.Policy) { msAddBinding(p, flagMSIamRole, flagMSIamMember) })
}

func runMSFedRemoveIam(cmd *cobra.Command, args []string) error {
	return msFedModifyIam(args[0], func(p *metastore.Policy) { msRemoveBinding(p, flagMSIamRole, flagMSIamMember) })
}

func msFedModifyIam(fedID string, mutate func(*metastore.Policy)) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := msFedName(fedID, project, flagMSLocation)
	policy, err := svc.Projects.Locations.Federations.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.Federations.SetIamPolicy(resource,
		&metastore.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func msAddBinding(p *metastore.Policy, role, member string) {
	for _, b := range p.Bindings {
		if b.Role == role {
			for _, m := range b.Members {
				if m == member {
					return
				}
			}
			b.Members = append(b.Members, member)
			return
		}
	}
	p.Bindings = append(p.Bindings, &metastore.Binding{Role: role, Members: []string{member}})
}

func msRemoveBinding(p *metastore.Policy, role, member string) {
	for _, b := range p.Bindings {
		if b.Role != role {
			continue
		}
		out := b.Members[:0]
		for _, m := range b.Members {
			if m != member {
				out = append(out, m)
			}
		}
		b.Members = out
	}
}

// --- locations impl ---

func runMSLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Get(msLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormatted(got, flagMSFormat)
}

func runMSLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	if flagMSFormat != "" {
		return emitFormatted(resp.Locations, flagMSFormat)
	}
	fmt.Printf("%-30s %s\n", "LOCATION_ID", "NAME")
	for _, l := range resp.Locations {
		fmt.Printf("%-30s %s\n", l.LocationId, l.Name)
	}
	return nil
}

// --- operations impl ---

func msOpName(id, project, location string) string {
	return msChild("operations", id, msLocationParent(project, location))
}

func runMSOpCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(msOpName(args[0], project, flagMSLocation),
		&metastore.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runMSOpDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(msOpName(args[0], project, flagMSLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runMSOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(msOpName(args[0], project, flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagMSFormat)
}

func runMSOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(msLocationParent(project, flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagMSFormat != "" {
		return emitFormatted(resp.Operations, flagMSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

func runMSOpWait(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := msOpName(args[0], project, flagMSLocation)
	op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting operation: %w", err)
	}
	final, err := msWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagMSFormat)
}

// --- services impl ---

func msSvcName(id, project, location string) string {
	return msChild("services", id, msLocationParent(project, location))
}

func runMSSvcCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	s := &metastore.Service{}
	if err := loadYAMLOrJSONInto(flagMSConfigFile, s); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Create(msLocationParent(project, flagMSLocation), s).
		ServiceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service: %w", err)
	}
	return msFinishOp(ctx, svc, op, "Create service", args[0], flagMSAsync)
}

func runMSSvcDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Delete(msSvcName(args[0], project, flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	return msFinishOp(ctx, svc, op, "Delete service", args[0], flagMSAsync)
}

func runMSSvcDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Services.Get(msSvcName(args[0], project, flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	return emitFormatted(got, flagMSFormat)
}

func runMSSvcList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Services.List(msLocationParent(project, flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing services: %w", err)
	}
	if flagMSFormat != "" {
		return emitFormatted(resp.Services, flagMSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, s := range resp.Services {
		fmt.Printf("%-40s %s\n", path.Base(s.Name), s.State)
	}
	return nil
}

func runMSSvcUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	s := &metastore.Service{}
	if err := loadYAMLOrJSONInto(flagMSConfigFile, s); err != nil {
		return err
	}
	mask := flagMSUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(s))
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Patch(msSvcName(args[0], project, flagMSLocation), s).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating service: %w", err)
	}
	return msFinishOp(ctx, svc, op, "Update service", args[0], flagMSAsync)
}

func runMSSvcGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Services.GetIamPolicy(msSvcName(args[0], project, flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagMSFormat)
}

func runMSSvcSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &metastore.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Services.SetIamPolicy(msSvcName(args[0], project, flagMSLocation),
		&metastore.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runMSSvcAddIam(cmd *cobra.Command, args []string) error {
	return msSvcModifyIam(args[0], func(p *metastore.Policy) { msAddBinding(p, flagMSIamRole, flagMSIamMember) })
}

func runMSSvcRemoveIam(cmd *cobra.Command, args []string) error {
	return msSvcModifyIam(args[0], func(p *metastore.Policy) { msRemoveBinding(p, flagMSIamRole, flagMSIamMember) })
}

func msSvcModifyIam(svcID string, mutate func(*metastore.Policy)) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := msSvcName(svcID, project, flagMSLocation)
	policy, err := svc.Projects.Locations.Services.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.Services.SetIamPolicy(resource,
		&metastore.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runMSSvcAlterLoc(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.AlterLocation(msSvcName(args[0], project, flagMSLocation),
		&metastore.AlterMetadataResourceLocationRequest{
			ResourceName: flagMSResourceName,
			LocationUri:  flagMSLocationURI,
		}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("altering metadata resource location: %w", err)
	}
	return msFinishOp(ctx, svc, op, "Alter metadata resource location", args[0], flagMSAsync)
}

func runMSSvcAlterTbl(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.AlterTableProperties(msSvcName(args[0], project, flagMSLocation),
		&metastore.AlterTablePropertiesRequest{
			TableName:  flagMSTable,
			Properties: flagMSProperties,
		}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("altering table properties: %w", err)
	}
	return msFinishOp(ctx, svc, op, "Alter table properties", args[0], flagMSAsync)
}

func runMSSvcMoveTbl(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.MoveTableToDatabase(msSvcName(args[0], project, flagMSLocation),
		&metastore.MoveTableToDatabaseRequest{
			DbName:            flagMSDBSource,
			DestinationDbName: flagMSDBDest,
			TableName:         flagMSTblName,
		}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("moving table: %w", err)
	}
	return msFinishOp(ctx, svc, op, "Move table", args[0], flagMSAsync)
}

func runMSSvcQuery(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.QueryMetadata(msSvcName(args[0], project, flagMSLocation),
		&metastore.QueryMetadataRequest{Query: flagMSQuery}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("querying metadata: %w", err)
	}
	final, err := msWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagMSFormat)
}

func runMSSvcExport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.ExportMetadata(msSvcName(args[0], project, flagMSLocation),
		&metastore.ExportMetadataRequest{
			DestinationGcsFolder: flagMSDestGCS,
			DatabaseDumpType:     flagMSExportDB,
		}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting metadata: %w", err)
	}
	return msFinishOp(ctx, svc, op, "Export metadata", args[0], flagMSAsync)
}

func runMSSvcImport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &metastore.MetadataImport{
		Description: flagMSImportDesc,
		DatabaseDump: &metastore.DatabaseDump{
			GcsUri: flagMSImportGCS,
			Type:   flagMSImportDBType,
		},
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.MetadataImports.Create(msSvcName(args[0], project, flagMSLocation), body).
		MetadataImportId(args[1]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing metadata: %w", err)
	}
	return msFinishOp(ctx, svc, op, "Import metadata", args[1], flagMSAsync)
}

func runMSSvcRestore(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	// Fully-qualify backup path if only ID given.
	backup := flagMSBackupName
	if !strings.HasPrefix(backup, "projects/") {
		source := flagMSRestoreServ
		if source == "" {
			source = args[0]
		}
		backup = fmt.Sprintf("%s/backups/%s", msSvcName(source, project, flagMSLocation), backup)
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Restore(msSvcName(args[0], project, flagMSLocation),
		&metastore.RestoreServiceRequest{Backup: backup, RestoreType: flagMSRestoreType}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring service: %w", err)
	}
	return msFinishOp(ctx, svc, op, "Restore service", args[0], flagMSAsync)
}

// --- backups impl ---

func msBackupParent(project, location, service string) string {
	return msSvcName(service, project, location)
}

func msBackupName(id, project, location, service string) string {
	return msChild("backups", id, msBackupParent(project, location, service))
}

func runMSBackupCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	b := &metastore.Backup{}
	if flagMSConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagMSConfigFile, b); err != nil {
			return err
		}
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Backups.Create(
		msBackupParent(project, flagMSLocation, flagMSService), b).
		BackupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	return msFinishOp(ctx, svc, op, "Create backup", args[0], flagMSAsync)
}

func runMSBackupDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Backups.Delete(
		msBackupName(args[0], project, flagMSLocation, flagMSService)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	return msFinishOp(ctx, svc, op, "Delete backup", args[0], flagMSAsync)
}

func runMSBackupDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Services.Backups.Get(
		msBackupName(args[0], project, flagMSLocation, flagMSService)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormatted(got, flagMSFormat)
}

func runMSBackupList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MetastoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Services.Backups.List(
		msBackupParent(project, flagMSLocation, flagMSService)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}
	if flagMSFormat != "" {
		return emitFormatted(resp.Backups, flagMSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, b := range resp.Backups {
		fmt.Printf("%-40s %s\n", path.Base(b.Name), b.State)
	}
	return nil
}
