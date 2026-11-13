package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	metastore "google.golang.org/api/metastore/v1"
)

// --- gcloud metastore (#356) ---

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

func (g *metastoreCmds) msFinishOp(ctx context.Context, svc *metastore.APIService, op *metastore.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(g.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(g.inv.stdout, op, "")
	}
	final, err := msWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(g.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(g.inv.stdout, final.Response, "")
	}
	return nil
}

var (
	flagMSBackup string

	flagMSDBType string
)

// --- federations ---

// --- locations ---

// --- operations ---

// --- services ---

// backups (nested)

func (g *metastoreCmds) initMetastore() {
	// federations
	fedAll := []*cobra.Command{g.msFedCreateCmd, g.msFedDeleteCmd, g.msFedDescribeCmd, g.msFedListCmd, g.msFedUpdateCmd,
		g.msFedGetIamCmd, g.msFedSetIamCmd, g.msFedAddIamCmd, g.msFedRemoveIamCmd}
	for _, c := range fedAll {
		c.Flags().StringVar(&g.flagMSLocation, "location", "", "Location containing the federation (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{g.msFedCreateCmd, g.msFedUpdateCmd} {
		c.Flags().StringVar(&g.flagMSConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Federation body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.msFedUpdateCmd.Flags().StringVar(&g.flagMSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{g.msFedCreateCmd, g.msFedDeleteCmd, g.msFedUpdateCmd} {
		c.Flags().BoolVar(&g.flagMSAsync, "async", false, "Return the LRO without waiting")
	}
	for _, c := range []*cobra.Command{g.msFedDescribeCmd, g.msFedListCmd, g.msFedGetIamCmd} {
		c.Flags().StringVar(&g.flagMSFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{g.msFedAddIamCmd, g.msFedRemoveIamCmd} {
		c.Flags().StringVar(&g.flagMSIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&g.flagMSIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	g.metastoreFederationsCmd.AddCommand(fedAll...)
	g.metastoreCmd.AddCommand(g.metastoreFederationsCmd)

	// locations
	g.msLocDescribeCmd.Flags().StringVar(&g.flagMSFormat, "format", "", "Output format")
	g.msLocListCmd.Flags().StringVar(&g.flagMSFormat, "format", "", "Output format")
	g.metastoreLocationsCmd.AddCommand(g.msLocDescribeCmd, g.msLocListCmd)
	g.metastoreCmd.AddCommand(g.metastoreLocationsCmd)

	// operations
	opAll := []*cobra.Command{g.msOpCancelCmd, g.msOpDeleteCmd, g.msOpDescribeCmd, g.msOpListCmd, g.msOpWaitCmd}
	for _, c := range opAll {
		c.Flags().StringVar(&g.flagMSLocation, "location", "", "Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{g.msOpDescribeCmd, g.msOpListCmd, g.msOpWaitCmd} {
		c.Flags().StringVar(&g.flagMSFormat, "format", "", "Output format")
	}
	g.metastoreOperationsCmd.AddCommand(opAll...)
	g.metastoreCmd.AddCommand(g.metastoreOperationsCmd)

	// services
	svcAll := []*cobra.Command{g.msSvcCreateCmd, g.msSvcDeleteCmd, g.msSvcDescribeCmd, g.msSvcListCmd, g.msSvcUpdateCmd,
		g.msSvcGetIamCmd, g.msSvcSetIamCmd, g.msSvcAddIamCmd, g.msSvcRemoveIamCmd,
		g.msSvcAlterLocCmd, g.msSvcAlterTblCmd, g.msSvcMoveTblCmd, g.msSvcQueryCmd,
		g.msSvcExportCmd, g.msSvcImportCmd, g.msSvcRestoreCmd}
	for _, c := range svcAll {
		c.Flags().StringVar(&g.flagMSLocation, "location", "", "Location containing the service (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{g.msSvcCreateCmd, g.msSvcUpdateCmd} {
		c.Flags().StringVar(&g.flagMSConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Service body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.msSvcUpdateCmd.Flags().StringVar(&g.flagMSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{g.msSvcCreateCmd, g.msSvcDeleteCmd, g.msSvcUpdateCmd,
		g.msSvcAlterLocCmd, g.msSvcAlterTblCmd, g.msSvcMoveTblCmd,
		g.msSvcExportCmd, g.msSvcImportCmd, g.msSvcRestoreCmd} {
		c.Flags().BoolVar(&g.flagMSAsync, "async", false, "Return the LRO without waiting")
	}
	for _, c := range []*cobra.Command{g.msSvcDescribeCmd, g.msSvcListCmd, g.msSvcGetIamCmd, g.msSvcQueryCmd} {
		c.Flags().StringVar(&g.flagMSFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{g.msSvcAddIamCmd, g.msSvcRemoveIamCmd} {
		c.Flags().StringVar(&g.flagMSIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&g.flagMSIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	g.msSvcAlterLocCmd.Flags().StringVar(&g.flagMSResourceName, "resource-name", "", "Metadata resource name (required)")
	g.msSvcAlterLocCmd.Flags().StringVar(&g.flagMSLocationURI, "location-uri", "", "New location URI (required)")
	_ = g.msSvcAlterLocCmd.MarkFlagRequired("resource-name")
	_ = g.msSvcAlterLocCmd.MarkFlagRequired("location-uri")
	g.msSvcAlterTblCmd.Flags().StringVar(&g.flagMSTable, "table", "", "Fully qualified table name (required)")
	g.msSvcAlterTblCmd.Flags().StringToStringVar(&g.flagMSProperties, "properties", nil, "Table properties (key=value); may repeat")
	_ = g.msSvcAlterTblCmd.MarkFlagRequired("table")
	g.msSvcMoveTblCmd.Flags().StringVar(&g.flagMSDBSource, "source-database", "", "Source database (required)")
	g.msSvcMoveTblCmd.Flags().StringVar(&g.flagMSDBDest, "destination-database", "", "Destination database (required)")
	g.msSvcMoveTblCmd.Flags().StringVar(&g.flagMSTblName, "table-name", "", "Table name (required)")
	_ = g.msSvcMoveTblCmd.MarkFlagRequired("source-database")
	_ = g.msSvcMoveTblCmd.MarkFlagRequired("destination-database")
	_ = g.msSvcMoveTblCmd.MarkFlagRequired("table-name")
	g.msSvcQueryCmd.Flags().StringVar(&g.flagMSQuery, "query", "", "HQL query to run (required)")
	_ = g.msSvcQueryCmd.MarkFlagRequired("query")
	g.msSvcExportCmd.Flags().StringVar(&g.flagMSDestGCS, "destination-gcs-folder", "", "gs:// destination folder (required)")
	g.msSvcExportCmd.Flags().StringVar(&g.flagMSExportDB, "database-dump-type", "MYSQL", "Database dump type")
	_ = g.msSvcExportCmd.MarkFlagRequired("destination-gcs-folder")
	g.msSvcImportCmd.Flags().StringVar(&g.flagMSImportGCS, "database-dump", "", "gs:// URI of the dump (required)")
	g.msSvcImportCmd.Flags().StringVar(&g.flagMSImportDesc, "description", "", "Description for the import")
	g.msSvcImportCmd.Flags().StringVar(&g.flagMSImportDBType, "database-type", "MYSQL", "Database type of the dump")
	_ = g.msSvcImportCmd.MarkFlagRequired("database-dump")
	g.msSvcRestoreCmd.Flags().StringVar(&g.flagMSBackupName, "backup", "", "Backup name to restore from (required)")
	g.msSvcRestoreCmd.Flags().StringVar(&g.flagMSRestoreType, "restore-type", "FULL", "Restore type (FULL or METADATA_ONLY)")
	g.msSvcRestoreCmd.Flags().StringVar(&g.flagMSRestoreServ, "source-service", "", "Source service (required if backup path is not fully-qualified)")
	_ = g.msSvcRestoreCmd.MarkFlagRequired("backup")
	g.metastoreServicesCmd.AddCommand(svcAll...)

	// backups (nested)
	bkAll := []*cobra.Command{g.msBackupCreateCmd, g.msBackupDeleteCmd, g.msBackupDescribeCmd, g.msBackupListCmd}
	for _, c := range bkAll {
		c.Flags().StringVar(&g.flagMSLocation, "location", "", "Location of the parent service (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagMSService, "service", "", "Parent service (required)")
		_ = c.MarkFlagRequired("service")
	}
	for _, c := range []*cobra.Command{g.msBackupCreateCmd, g.msBackupDeleteCmd} {
		c.Flags().BoolVar(&g.flagMSAsync, "async", false, "Return the LRO without waiting")
	}
	g.msBackupCreateCmd.Flags().StringVar(&g.flagMSConfigFile, "config-file", "",
		"Path to an optional JSON/YAML file with additional Backup body fields")
	for _, c := range []*cobra.Command{g.msBackupDescribeCmd, g.msBackupListCmd} {
		c.Flags().StringVar(&g.flagMSFormat, "format", "", "Output format")
	}
	g.metastoreServicesBackupsCmd.AddCommand(bkAll...)
	g.metastoreServicesCmd.AddCommand(g.metastoreServicesBackupsCmd)

	g.metastoreCmd.AddCommand(g.metastoreServicesCmd)

}

// --- federations impl ---

func msFedName(id, project, location string) string {
	return msChild("federations", id, msLocationParent(project, location))
}

func (g *metastoreCmds) runMSFedCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	f := &metastore.Federation{}
	if err := loadYAMLOrJSONInto(g.flagMSConfigFile, f); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Federations.Create(msLocationParent(project, g.flagMSLocation), f).
		FederationId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating federation: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Create federation", args[0], g.flagMSAsync)
}

func (g *metastoreCmds) runMSFedDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Federations.Delete(msFedName(args[0], project, g.flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting federation: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Delete federation", args[0], g.flagMSAsync)
}

func (g *metastoreCmds) runMSFedDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Federations.Get(msFedName(args[0], project, g.flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing federation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagMSFormat)
}

func (g *metastoreCmds) runMSFedList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Federations.List(msLocationParent(project, g.flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing federations: %w", err)
	}
	if g.flagMSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Federations, g.flagMSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Federations), "%-40s %s\n", "NAME", "STATE")
	for _, f := range resp.Federations {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(f.Name), f.State)
	}
	return nil
}

func (g *metastoreCmds) runMSFedUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	f := &metastore.Federation{}
	if err := loadYAMLOrJSONInto(g.flagMSConfigFile, f); err != nil {
		return err
	}
	mask := g.flagMSUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(f))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Federations.Patch(msFedName(args[0], project, g.flagMSLocation), f).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating federation: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Update federation", args[0], g.flagMSAsync)
}

func (g *metastoreCmds) runMSFedGetIam(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Federations.GetIamPolicy(msFedName(args[0], project, g.flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, policy, g.flagMSFormat)
}

func (g *metastoreCmds) runMSFedSetIam(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	policy := &metastore.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Federations.SetIamPolicy(msFedName(args[0], project, g.flagMSLocation),
		&metastore.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *metastoreCmds) runMSFedAddIam(cmd *cobra.Command, args []string) error {
	return g.msFedModifyIam(args[0], func(p *metastore.Policy) { msAddBinding(p, g.flagMSIamRole, g.flagMSIamMember) })
}

func (g *metastoreCmds) runMSFedRemoveIam(cmd *cobra.Command, args []string) error {
	return g.msFedModifyIam(args[0], func(p *metastore.Policy) { msRemoveBinding(p, g.flagMSIamRole, g.flagMSIamMember) })
}

func (g *metastoreCmds) msFedModifyIam(fedID string, mutate func(*metastore.Policy)) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := g.inv.context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	resource := msFedName(fedID, project, g.flagMSLocation)
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
	return emitFormattedTo(g.inv.stdout, got, "")
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

func (g *metastoreCmds) runMSLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Get(msLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagMSFormat)
}

func (g *metastoreCmds) runMSLocList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	if g.flagMSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Locations, g.flagMSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Locations), "%-30s %s\n", "LOCATION_ID", "NAME")
	for _, l := range resp.Locations {
		fmt.Fprintf(g.inv.stdout, "%-30s %s\n", l.LocationId, l.Name)
	}
	return nil
}

// --- operations impl ---

func msOpName(id, project, location string) string {
	return msChild("operations", id, msLocationParent(project, location))
}

func (g *metastoreCmds) runMSOpCancel(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(msOpName(args[0], project, g.flagMSLocation),
		&metastore.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (g *metastoreCmds) runMSOpDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(msOpName(args[0], project, g.flagMSLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (g *metastoreCmds) runMSOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(msOpName(args[0], project, g.flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagMSFormat)
}

func (g *metastoreCmds) runMSOpList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(msLocationParent(project, g.flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if g.flagMSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Operations, g.flagMSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(g.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

func (g *metastoreCmds) runMSOpWait(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	name := msOpName(args[0], project, g.flagMSLocation)
	op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting operation: %w", err)
	}
	final, err := msWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(g.inv.stdout, final, g.flagMSFormat)
}

// --- services impl ---

func msSvcName(id, project, location string) string {
	return msChild("services", id, msLocationParent(project, location))
}

func (g *metastoreCmds) runMSSvcCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	s := &metastore.Service{}
	if err := loadYAMLOrJSONInto(g.flagMSConfigFile, s); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Create(msLocationParent(project, g.flagMSLocation), s).
		ServiceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Create service", args[0], g.flagMSAsync)
}

func (g *metastoreCmds) runMSSvcDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Delete(msSvcName(args[0], project, g.flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Delete service", args[0], g.flagMSAsync)
}

func (g *metastoreCmds) runMSSvcDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Services.Get(msSvcName(args[0], project, g.flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagMSFormat)
}

func (g *metastoreCmds) runMSSvcList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Services.List(msLocationParent(project, g.flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing services: %w", err)
	}
	if g.flagMSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Services, g.flagMSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Services), "%-40s %s\n", "NAME", "STATE")
	for _, s := range resp.Services {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(s.Name), s.State)
	}
	return nil
}

func (g *metastoreCmds) runMSSvcUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	s := &metastore.Service{}
	if err := loadYAMLOrJSONInto(g.flagMSConfigFile, s); err != nil {
		return err
	}
	mask := g.flagMSUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(s))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Patch(msSvcName(args[0], project, g.flagMSLocation), s).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating service: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Update service", args[0], g.flagMSAsync)
}

func (g *metastoreCmds) runMSSvcGetIam(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Services.GetIamPolicy(msSvcName(args[0], project, g.flagMSLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, policy, g.flagMSFormat)
}

func (g *metastoreCmds) runMSSvcSetIam(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	policy := &metastore.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Services.SetIamPolicy(msSvcName(args[0], project, g.flagMSLocation),
		&metastore.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *metastoreCmds) runMSSvcAddIam(cmd *cobra.Command, args []string) error {
	return g.msSvcModifyIam(args[0], func(p *metastore.Policy) { msAddBinding(p, g.flagMSIamRole, g.flagMSIamMember) })
}

func (g *metastoreCmds) runMSSvcRemoveIam(cmd *cobra.Command, args []string) error {
	return g.msSvcModifyIam(args[0], func(p *metastore.Policy) { msRemoveBinding(p, g.flagMSIamRole, g.flagMSIamMember) })
}

func (g *metastoreCmds) msSvcModifyIam(svcID string, mutate func(*metastore.Policy)) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := g.inv.context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	resource := msSvcName(svcID, project, g.flagMSLocation)
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
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *metastoreCmds) runMSSvcAlterLoc(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.AlterLocation(msSvcName(args[0], project, g.flagMSLocation),
		&metastore.AlterMetadataResourceLocationRequest{
			ResourceName: g.flagMSResourceName,
			LocationUri:  g.flagMSLocationURI,
		}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("altering metadata resource location: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Alter metadata resource location", args[0], g.flagMSAsync)
}

func (g *metastoreCmds) runMSSvcAlterTbl(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.AlterTableProperties(msSvcName(args[0], project, g.flagMSLocation),
		&metastore.AlterTablePropertiesRequest{
			TableName:  g.flagMSTable,
			Properties: g.flagMSProperties,
		}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("altering table properties: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Alter table properties", args[0], g.flagMSAsync)
}

func (g *metastoreCmds) runMSSvcMoveTbl(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.MoveTableToDatabase(msSvcName(args[0], project, g.flagMSLocation),
		&metastore.MoveTableToDatabaseRequest{
			DbName:            g.flagMSDBSource,
			DestinationDbName: g.flagMSDBDest,
			TableName:         g.flagMSTblName,
		}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("moving table: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Move table", args[0], g.flagMSAsync)
}

func (g *metastoreCmds) runMSSvcQuery(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.QueryMetadata(msSvcName(args[0], project, g.flagMSLocation),
		&metastore.QueryMetadataRequest{Query: g.flagMSQuery}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("querying metadata: %w", err)
	}
	final, err := msWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(g.inv.stdout, final, g.flagMSFormat)
}

func (g *metastoreCmds) runMSSvcExport(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.ExportMetadata(msSvcName(args[0], project, g.flagMSLocation),
		&metastore.ExportMetadataRequest{
			DestinationGcsFolder: g.flagMSDestGCS,
			DatabaseDumpType:     g.flagMSExportDB,
		}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting metadata: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Export metadata", args[0], g.flagMSAsync)
}

func (g *metastoreCmds) runMSSvcImport(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	body := &metastore.MetadataImport{
		Description: g.flagMSImportDesc,
		DatabaseDump: &metastore.DatabaseDump{
			GcsUri: g.flagMSImportGCS,
			Type:   g.flagMSImportDBType,
		},
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.MetadataImports.Create(msSvcName(args[0], project, g.flagMSLocation), body).
		MetadataImportId(args[1]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing metadata: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Import metadata", args[1], g.flagMSAsync)
}

func (g *metastoreCmds) runMSSvcRestore(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	// Fully-qualify backup path if only ID given.
	backup := g.flagMSBackupName
	if !strings.HasPrefix(backup, "projects/") {
		source := g.flagMSRestoreServ
		if source == "" {
			source = args[0]
		}
		backup = fmt.Sprintf("%s/backups/%s", msSvcName(source, project, g.flagMSLocation), backup)
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Restore(msSvcName(args[0], project, g.flagMSLocation),
		&metastore.RestoreServiceRequest{Backup: backup, RestoreType: g.flagMSRestoreType}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring service: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Restore service", args[0], g.flagMSAsync)
}

// --- backups impl ---

func msBackupParent(project, location, service string) string {
	return msSvcName(service, project, location)
}

func msBackupName(id, project, location, service string) string {
	return msChild("backups", id, msBackupParent(project, location, service))
}

func (g *metastoreCmds) runMSBackupCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	b := &metastore.Backup{}
	if g.flagMSConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagMSConfigFile, b); err != nil {
			return err
		}
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Backups.Create(
		msBackupParent(project, g.flagMSLocation, g.flagMSService), b).
		BackupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Create backup", args[0], g.flagMSAsync)
}

func (g *metastoreCmds) runMSBackupDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Backups.Delete(
		msBackupName(args[0], project, g.flagMSLocation, g.flagMSService)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Delete backup", args[0], g.flagMSAsync)
}

func (g *metastoreCmds) runMSBackupDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Services.Backups.Get(
		msBackupName(args[0], project, g.flagMSLocation, g.flagMSService)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagMSFormat)
}

func (g *metastoreCmds) runMSBackupList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Services.Backups.List(
		msBackupParent(project, g.flagMSLocation, g.flagMSService)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}
	if g.flagMSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Backups, g.flagMSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Backups), "%-40s %s\n", "NAME", "STATE")
	for _, b := range resp.Backups {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(b.Name), b.State)
	}
	return nil
}

// metastoreCmds holds the commands and flag values of one invocation's "metastore"
// command group, so invocations don't share state (#2070).
type metastoreCmds struct {
	inv *invocation

	metastoreCmd     *cobra.Command
	flagMSLocation   string
	flagMSConfigFile string
	flagMSUpdateMask string
	flagMSFormat     string
	flagMSAsync      bool
	flagMSService    string
	flagMSIamMember  string
	flagMSIamRole    string
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
	flagMSDestGCS  string
	flagMSExportDB string
	// import
	flagMSImportGCS    string
	flagMSImportDesc   string
	flagMSImportDBType string
	// restore
	flagMSBackupName            string
	flagMSRestoreType           string
	flagMSRestoreServ           string
	metastoreFederationsCmd     *cobra.Command
	msFedCreateCmd              *cobra.Command
	msFedDeleteCmd              *cobra.Command
	msFedDescribeCmd            *cobra.Command
	msFedListCmd                *cobra.Command
	msFedUpdateCmd              *cobra.Command
	msFedGetIamCmd              *cobra.Command
	msFedSetIamCmd              *cobra.Command
	msFedAddIamCmd              *cobra.Command
	msFedRemoveIamCmd           *cobra.Command
	metastoreLocationsCmd       *cobra.Command
	msLocDescribeCmd            *cobra.Command
	msLocListCmd                *cobra.Command
	metastoreOperationsCmd      *cobra.Command
	msOpCancelCmd               *cobra.Command
	msOpDeleteCmd               *cobra.Command
	msOpDescribeCmd             *cobra.Command
	msOpListCmd                 *cobra.Command
	msOpWaitCmd                 *cobra.Command
	metastoreServicesCmd        *cobra.Command
	msSvcCreateCmd              *cobra.Command
	msSvcDeleteCmd              *cobra.Command
	msSvcDescribeCmd            *cobra.Command
	msSvcListCmd                *cobra.Command
	msSvcUpdateCmd              *cobra.Command
	msSvcGetIamCmd              *cobra.Command
	msSvcSetIamCmd              *cobra.Command
	msSvcAddIamCmd              *cobra.Command
	msSvcRemoveIamCmd           *cobra.Command
	msSvcAlterLocCmd            *cobra.Command
	msSvcAlterTblCmd            *cobra.Command
	msSvcMoveTblCmd             *cobra.Command
	msSvcQueryCmd               *cobra.Command
	msSvcExportCmd              *cobra.Command
	msSvcImportCmd              *cobra.Command
	msSvcRestoreCmd             *cobra.Command
	metastoreServicesBackupsCmd *cobra.Command
	msBackupCreateCmd           *cobra.Command
	msBackupDeleteCmd           *cobra.Command
	msBackupDescribeCmd         *cobra.Command
	msBackupListCmd             *cobra.Command
}

// newMetastoreCmd returns the "metastore" command group for inv.
func newMetastoreCmd(inv *invocation) *cobra.Command {
	return newMetastoreCmds(inv).metastoreCmd
}

// newMetastoreCmds builds the "metastore" command group for inv and returns its state.
func newMetastoreCmds(inv *invocation) *metastoreCmds {
	g := &metastoreCmds{inv: inv}
	g.metastoreCmd = &cobra.Command{Use: "metastore", Short: "Manage Dataproc Metastore"}
	g.metastoreFederationsCmd = &cobra.Command{Use: "federations", Short: "Manage Metastore federations"}
	g.msFedCreateCmd = &cobra.Command{
		Use: "create FEDERATION", Short: "Create a federation from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runMSFedCreate,
	}
	g.msFedDeleteCmd = &cobra.Command{
		Use: "delete FEDERATION", Short: "Delete a federation",
		Args: cobra.ExactArgs(1), RunE: g.runMSFedDelete,
	}
	g.msFedDescribeCmd = &cobra.Command{
		Use: "describe FEDERATION", Short: "Describe a federation",
		Args: cobra.ExactArgs(1), RunE: g.runMSFedDescribe,
	}
	g.msFedListCmd = &cobra.Command{
		Use: "list", Short: "List federations in a location",
		Args: cobra.NoArgs, RunE: g.runMSFedList,
	}
	g.msFedUpdateCmd = &cobra.Command{
		Use: "update FEDERATION", Short: "Update a federation from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runMSFedUpdate,
	}
	g.msFedGetIamCmd = &cobra.Command{
		Use: "get-iam-policy FEDERATION", Short: "Get the IAM policy for a federation",
		Args: cobra.ExactArgs(1), RunE: g.runMSFedGetIam,
	}
	g.msFedSetIamCmd = &cobra.Command{
		Use: "set-iam-policy FEDERATION POLICY_FILE", Short: "Replace the IAM policy",
		Args: cobra.ExactArgs(2), RunE: g.runMSFedSetIam,
	}
	g.msFedAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding FEDERATION", Short: "Add an IAM binding",
		Args: cobra.ExactArgs(1), RunE: g.runMSFedAddIam,
	}
	g.msFedRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding FEDERATION", Short: "Remove an IAM binding",
		Args: cobra.ExactArgs(1), RunE: g.runMSFedRemoveIam,
	}
	g.metastoreLocationsCmd = &cobra.Command{Use: "locations", Short: "Manage Metastore locations"}
	g.msLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Metastore location",
		Args: cobra.ExactArgs(1), RunE: g.runMSLocDescribe,
	}
	g.msLocListCmd = &cobra.Command{
		Use: "list", Short: "List Metastore locations",
		Args: cobra.NoArgs, RunE: g.runMSLocList,
	}
	g.metastoreOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Metastore operations"}
	g.msOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: g.runMSOpCancel,
	}
	g.msOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an operation",
		Args: cobra.ExactArgs(1), RunE: g.runMSOpDelete,
	}
	g.msOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: g.runMSOpDescribe,
	}
	g.msOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a location",
		Args: cobra.NoArgs, RunE: g.runMSOpList,
	}
	g.msOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for an operation to complete",
		Args: cobra.ExactArgs(1), RunE: g.runMSOpWait,
	}
	g.metastoreServicesCmd = &cobra.Command{Use: "services", Short: "Manage Metastore services"}
	g.msSvcCreateCmd = &cobra.Command{
		Use: "create SERVICE", Short: "Create a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runMSSvcCreate,
	}
	g.msSvcDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete a service",
		Args: cobra.ExactArgs(1), RunE: g.runMSSvcDelete,
	}
	g.msSvcDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a service",
		Args: cobra.ExactArgs(1), RunE: g.runMSSvcDescribe,
	}
	g.msSvcListCmd = &cobra.Command{
		Use: "list", Short: "List services in a location",
		Args: cobra.NoArgs, RunE: g.runMSSvcList,
	}
	g.msSvcUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runMSSvcUpdate,
	}
	g.msSvcGetIamCmd = &cobra.Command{
		Use: "get-iam-policy SERVICE", Short: "Get the IAM policy for a service",
		Args: cobra.ExactArgs(1), RunE: g.runMSSvcGetIam,
	}
	g.msSvcSetIamCmd = &cobra.Command{
		Use: "set-iam-policy SERVICE POLICY_FILE", Short: "Replace the IAM policy",
		Args: cobra.ExactArgs(2), RunE: g.runMSSvcSetIam,
	}
	g.msSvcAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding SERVICE", Short: "Add an IAM binding",
		Args: cobra.ExactArgs(1), RunE: g.runMSSvcAddIam,
	}
	g.msSvcRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding SERVICE", Short: "Remove an IAM binding",
		Args: cobra.ExactArgs(1), RunE: g.runMSSvcRemoveIam,
	}
	g.msSvcAlterLocCmd = &cobra.Command{
		Use: "alter-metadata-resource-location SERVICE", Short: "Alter a metadata resource location",
		Args: cobra.ExactArgs(1), RunE: g.runMSSvcAlterLoc,
	}
	g.msSvcAlterTblCmd = &cobra.Command{
		Use: "alter-table-properties SERVICE", Short: "Alter table properties",
		Args: cobra.ExactArgs(1), RunE: g.runMSSvcAlterTbl,
	}
	g.msSvcMoveTblCmd = &cobra.Command{
		Use: "move-table-to-database SERVICE", Short: "Move a table between databases",
		Args: cobra.ExactArgs(1), RunE: g.runMSSvcMoveTbl,
	}
	g.msSvcQueryCmd = &cobra.Command{
		Use: "query-metadata SERVICE", Short: "Run a metadata query",
		Args: cobra.ExactArgs(1), RunE: g.runMSSvcQuery,
	}
	g.msSvcExportCmd = &cobra.Command{
		Use: "export SERVICE", Short: "Export metadata to Cloud Storage",
		Args: cobra.ExactArgs(1), RunE: g.runMSSvcExport,
	}
	g.msSvcImportCmd = &cobra.Command{
		Use: "import SERVICE IMPORT_ID", Short: "Import metadata from Cloud Storage",
		Args: cobra.ExactArgs(2), RunE: g.runMSSvcImport,
	}
	g.msSvcRestoreCmd = &cobra.Command{
		Use: "restore SERVICE", Short: "Restore a service from a backup",
		Args: cobra.ExactArgs(1), RunE: g.runMSSvcRestore,
	}
	g.metastoreServicesBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage Metastore service backups"}
	g.msBackupCreateCmd = &cobra.Command{
		Use: "create BACKUP", Short: "Create a backup of a service",
		Args: cobra.ExactArgs(1), RunE: g.runMSBackupCreate,
	}
	g.msBackupDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a service backup",
		Args: cobra.ExactArgs(1), RunE: g.runMSBackupDelete,
	}
	g.msBackupDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a service backup",
		Args: cobra.ExactArgs(1), RunE: g.runMSBackupDescribe,
	}
	g.msBackupListCmd = &cobra.Command{
		Use: "list", Short: "List service backups",
		Args: cobra.NoArgs, RunE: g.runMSBackupList,
	}
	g.initMetastore()
	return g
}
