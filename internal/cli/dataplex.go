package cli

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	dataplex "google.golang.org/api/dataplex/v1"
)

// --- datascans run ---

// --- datascans jobs ---

// --- datascans describe ---

// --- datascans list ---

// --- datascans create ---

// --- datascans delete ---

func (g *dataplexCmds) initDataplex() {
	g.dataplexDatascansRunCmd.Flags().StringVar(&g.flagDataplexLocation, "location", "", "Location of the data scan")
	g.dataplexDatascansRunCmd.Flags().StringVar(&g.flagDataplexDatascan, "datascan", "", "Data scan ID (for jobs commands)")

	g.dataplexDatascansJobsListCmd.Flags().StringVar(&g.flagDataplexLocation, "location", "", "Location of the data scan")
	g.dataplexDatascansJobsListCmd.Flags().StringVar(&g.flagDataplexDatascan, "datascan", "", "Data scan ID")
	g.dataplexDatascansJobsListCmd.Flags().StringVar(&g.flagDataplexJobFormat, "format", "", "Output format (e.g. json)")
	g.dataplexDatascansJobsListCmd.Flags().BoolVar(&g.flagDataplexJobURI, "uri", false, "Print resource names")

	g.dataplexDatascansJobsDescribeCmd.Flags().StringVar(&g.flagDataplexLocation, "location", "", "Location of the data scan")
	g.dataplexDatascansJobsDescribeCmd.Flags().StringVar(&g.flagDataplexDatascan, "datascan", "", "Data scan ID")
	g.dataplexDatascansJobsDescribeCmd.Flags().StringVar(&g.flagDataplexJobView, "view", "", "Job view (BASIC or FULL)")

	g.dataplexDatascansDescribeCmd.Flags().StringVar(&g.flagDataplexLocation, "location", "", "Location of the data scan")
	g.dataplexDatascansListCmd.Flags().StringVar(&g.flagDataplexLocation, "location", "", "Location of the data scans")
	g.dataplexDatascansListCmd.Flags().StringVar(&g.flagDataplexListFormat, "format", "", "Output format (e.g. json)")
	g.dataplexDatascansListCmd.Flags().BoolVar(&g.flagDataplexListURI, "uri", false, "Print resource names")
	g.dataplexDatascansCreateCmd.Flags().StringVar(&g.flagDataplexLocation, "location", "", "Location of the data scan")
	g.dataplexDatascansCreateCmd.Flags().StringVar(&g.flagDataplexDataSource, "data-source", "", "Data source entity (required)")
	g.dataplexDatascansCreateCmd.MarkFlagRequired("data-source")
	g.dataplexDatascansCreateCmd.Flags().StringVar(&g.flagDataplexScanType, "type", "data-quality",
		"Scan type (data-quality, data-profile, or data-documentation)")
	g.dataplexDatascansCreateCmd.Flags().BoolVar(&g.flagDataplexCatalogPublish, "enable-catalog-publishing", false,
		"For data-documentation and data-profile scans, publish the latest scan results to Dataplex Catalog.")
	g.dataplexDatascansCreateCmd.Flags().StringVar(&g.flagDataplexProfileMode, "mode", "",
		"For data-profile scans, the execution mode (STANDARD or LIGHTWEIGHT).")
	g.dataplexDatascansDeleteCmd.Flags().StringVar(&g.flagDataplexLocation, "location", "", "Location of the data scan")
	// --quiet is provided by the global persistent flag

	g.dataplexDatascansJobsCmd.AddCommand(g.dataplexDatascansJobsListCmd)
	g.dataplexDatascansJobsCmd.AddCommand(g.dataplexDatascansJobsDescribeCmd)
	g.dataplexDatascansCmd.AddCommand(g.dataplexDatascansRunCmd)
	g.dataplexDatascansCmd.AddCommand(g.dataplexDatascansDescribeCmd)
	g.dataplexDatascansCmd.AddCommand(g.dataplexDatascansListCmd)
	g.dataplexDatascansCmd.AddCommand(g.dataplexDatascansCreateCmd)
	g.dataplexDatascansCmd.AddCommand(g.dataplexDatascansDeleteCmd)
	g.dataplexDatascansCmd.AddCommand(g.dataplexDatascansJobsCmd)
	g.dataplexCmd.AddCommand(g.dataplexDatascansCmd)

	g.registerDataplexContext(g.dataplexCmd)

	// gcloud-python dataplex subgroups not yet implemented (#541).
	for name, short := range map[string]string{
		"aspect-types":      "Manage aspect types",
		"assets":            "Manage Dataplex assets",
		"encryption-config": "Manage Dataplex encryption config",
		"entries":           "Manage Dataplex entries",
		"entry-groups":      "Manage entry groups",
		"entry-types":       "Manage entry types",
		"glossaries":        "Manage Dataplex glossaries",
		"lakes":             "Manage Dataplex lakes",
		"metadata-jobs":     "Manage metadata jobs",
		"tasks":             "Manage Dataplex tasks",
		"zones":             "Manage Dataplex zones",
	} {
		registerStubGroup(g.dataplexCmd, name, short, "list", "describe")
	}

}

// --- dataplex context lookup ---

func (g *dataplexCmds) registerDataplexContext(parent *cobra.Command) {
	ctxCmd := &cobra.Command{Use: "context", Short: "Manage Dataplex context resources"}
	lookupCmd := &cobra.Command{
		Use:   "lookup",
		Short: "Look up context metadata for the given Dataplex entries",
		Args:  cobra.NoArgs,
		RunE:  g.runDataplexContextLookup,
	}
	lookupCmd.Flags().StringVar(&g.flagDataplexLocation, "location", "", "Location of the Dataplex resources")
	lookupCmd.Flags().StringSliceVar(&g.flagDataplexCtxResources, "resources", nil,
		"Comma-separated list of entry names to look up context for (required)")
	_ = lookupCmd.MarkFlagRequired("resources")
	lookupCmd.Flags().StringVar(&g.flagDataplexCtxFormat, "context-format", "yaml",
		"Output format of the returned context (yaml, json, or xml)")
	lookupCmd.Flags().BoolVar(&g.flagDataplexCtxAllSchema, "all-schema-fields", false,
		"If set, include all schema fields in the returned context regardless of context_budget.")
	lookupCmd.Flags().StringVar(&g.flagDataplexCtxOptionsJSON, "options", "",
		"Additional lookup options as a JSON object (merged with well-known option flags).")
	ctxCmd.AddCommand(lookupCmd)
	parent.AddCommand(ctxCmd)
}

func (g *dataplexCmds) runDataplexContextLookup(cmd *cobra.Command, args []string) error {
	project, location, err := g.resolveDataplexLocation()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Dataplex(ctx)
	if err != nil {
		return err
	}
	options := map[string]string{}
	if g.flagDataplexCtxOptionsJSON != "" {
		if err := json.Unmarshal([]byte(g.flagDataplexCtxOptionsJSON), &options); err != nil {
			return fmt.Errorf("--options must be a JSON object of string keys and values: %w", err)
		}
	}
	if g.flagDataplexCtxFormat != "" {
		options["format"] = g.flagDataplexCtxFormat
	}
	if g.flagDataplexCtxAllSchema {
		options["all_schema_fields"] = "true"
	}
	req := &dataplex.GoogleCloudDataplexV1LookupContextRequest{
		Resources: g.flagDataplexCtxResources,
		Options:   options,
	}
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	resp, err := svc.Projects.Locations.LookupContext(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up context: %w", err)
	}
	fmt.Fprintln(g.inv.stdout, resp.Context)
	return nil
}

func (g *dataplexCmds) resolveDataplexLocation() (string, string, error) {
	project, err := g.inv.project()
	if err != nil {
		return "", "", err
	}
	location := g.flagDataplexLocation
	if location == "" {
		_, region, err := g.inv.region()
		if err != nil || region == "" {
			return "", "", fmt.Errorf("--location is required")
		}
		location = region
	}
	return project, location, nil
}

func datascanName(project, location, datascanID string) string {
	return fmt.Sprintf("projects/%s/locations/%s/dataScans/%s", project, location, datascanID)
}

func (g *dataplexCmds) runDataplexDatascansRun(cmd *cobra.Command, args []string) error {
	project, location, err := g.resolveDataplexLocation()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Dataplex(ctx)
	if err != nil {
		return err
	}

	name := datascanName(project, location, args[0])
	resp, err := svc.Projects.Locations.DataScans.Run(name, &dataplex.GoogleCloudDataplexV1RunDataScanRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running data scan: %w", err)
	}

	return g.inv.formatOutput(resp, "")
}

func (g *dataplexCmds) runDataplexDatascansJobsList(cmd *cobra.Command, args []string) error {
	project, location, err := g.resolveDataplexLocation()
	if err != nil {
		return err
	}

	datascanID := g.flagDataplexDatascan
	if datascanID == "" {
		return fmt.Errorf("--datascan is required")
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Dataplex(ctx)
	if err != nil {
		return err
	}

	parent := datascanName(project, location, datascanID)

	var allJobs []*dataplex.GoogleCloudDataplexV1DataScanJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.DataScans.Jobs.List(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing data scan jobs: %w", err)
		}
		allJobs = append(allJobs, resp.DataScanJobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if g.flagDataplexJobURI {
		for _, job := range allJobs {
			fmt.Fprintln(g.inv.stdout, job.Name)
		}
		return nil
	}

	if g.flagDataplexJobFormat == "json" {
		return g.inv.printJSON(allJobs)
	}

	printTableHeadingTo(g.inv.stdout, len(allJobs), "%-40s %-15s %-25s %-25s\n", "JOB_ID", "STATE", "START_TIME", "END_TIME")
	for _, j := range allJobs {
		fmt.Fprintf(g.inv.stdout, "%-40s %-15s %-25s %-25s\n", path.Base(j.Name), j.State, j.StartTime, j.EndTime)
	}
	return nil
}

func (g *dataplexCmds) runDataplexDatascansJobsDescribe(cmd *cobra.Command, args []string) error {
	project, location, err := g.resolveDataplexLocation()
	if err != nil {
		return err
	}

	datascanID := g.flagDataplexDatascan
	if datascanID == "" {
		return fmt.Errorf("--datascan is required")
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Dataplex(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%s/jobs/%s", datascanName(project, location, datascanID), args[0])
	call := svc.Projects.Locations.DataScans.Jobs.Get(name).Context(ctx)
	if g.flagDataplexJobView != "" {
		call = call.View(strings.ToUpper(g.flagDataplexJobView))
	}
	job, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing data scan job: %w", err)
	}

	return g.inv.formatOutput(job, "")
}

func (g *dataplexCmds) runDataplexDatascansDescribe(cmd *cobra.Command, args []string) error {
	project, location, err := g.resolveDataplexLocation()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Dataplex(ctx)
	if err != nil {
		return err
	}

	name := datascanName(project, location, args[0])
	scan, err := svc.Projects.Locations.DataScans.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing data scan: %w", err)
	}

	return g.inv.printJSON(scan)
}

func (g *dataplexCmds) runDataplexDatascansList(cmd *cobra.Command, args []string) error {
	project, location, err := g.resolveDataplexLocation()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Dataplex(ctx)
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	var allScans []*dataplex.GoogleCloudDataplexV1DataScan
	pageToken := ""
	for {
		call := svc.Projects.Locations.DataScans.List(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing data scans: %w", err)
		}
		allScans = append(allScans, resp.DataScans...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if g.flagDataplexListURI {
		for _, ds := range allScans {
			fmt.Fprintln(g.inv.stdout, ds.Name)
		}
		return nil
	}

	if g.flagDataplexListFormat == "json" {
		return g.inv.printJSON(allScans)
	}

	printTableHeadingTo(g.inv.stdout, len(allScans), "%-30s %-15s %-25s %-20s\n", "NAME", "STATE", "CREATE_TIME", "TYPE")
	for _, s := range allScans {
		scanType := ""
		if s.DataQualitySpec != nil {
			scanType = "DATA_QUALITY"
		} else if s.DataProfileSpec != nil {
			scanType = "DATA_PROFILE"
		}
		fmt.Fprintf(g.inv.stdout, "%-30s %-15s %-25s %-20s\n", path.Base(s.Name), s.State, s.CreateTime, scanType)
	}
	return nil
}

func (g *dataplexCmds) runDataplexDatascansCreate(cmd *cobra.Command, args []string) error {
	project, location, err := g.resolveDataplexLocation()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Dataplex(ctx)
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	scan := &dataplex.GoogleCloudDataplexV1DataScan{
		Data: &dataplex.GoogleCloudDataplexV1DataSource{
			Entity: g.flagDataplexDataSource,
		},
	}

	switch strings.ToLower(g.flagDataplexScanType) {
	case "data-quality":
		scan.DataQualitySpec = &dataplex.GoogleCloudDataplexV1DataQualitySpec{}
	case "data-profile":
		spec := &dataplex.GoogleCloudDataplexV1DataProfileSpec{}
		if g.flagDataplexCatalogPublish {
			spec.CatalogPublishingEnabled = true
		}
		if g.flagDataplexProfileMode != "" {
			spec.Mode = strings.ToUpper(g.flagDataplexProfileMode)
		}
		scan.DataProfileSpec = spec
	case "data-documentation":
		spec := &dataplex.GoogleCloudDataplexV1DataDocumentationSpec{}
		if g.flagDataplexCatalogPublish {
			spec.CatalogPublishingEnabled = true
		}
		scan.DataDocumentationSpec = spec
	default:
		return fmt.Errorf("unsupported scan type: %s (use data-quality, data-profile, or data-documentation)", g.flagDataplexScanType)
	}

	op, err := svc.Projects.Locations.DataScans.Create(parent, scan).DataScanId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating data scan: %w", err)
	}

	fmt.Fprintf(g.inv.stdout, "Created data scan [%s] (operation: %s).\n", args[0], op.Name)
	return nil
}

func (g *dataplexCmds) runDataplexDatascansDelete(cmd *cobra.Command, args []string) error {
	project, location, err := g.resolveDataplexLocation()
	if err != nil {
		return err
	}

	if !g.inv.flags.quiet {
		fmt.Fprintf(g.inv.stdout, "You are about to delete data scan [%s]. This action cannot be undone.\n", args[0])
		fmt.Fprint(g.inv.stdout, "Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Fprintln(g.inv.stdout, "Aborted.")
			return nil
		}
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Dataplex(ctx)
	if err != nil {
		return err
	}

	name := datascanName(project, location, args[0])
	if _, err := svc.Projects.Locations.DataScans.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting data scan: %w", err)
	}

	fmt.Fprintf(g.inv.stdout, "Deleted data scan [%s].\n", args[0])
	return nil
}

// dataplexCmds holds the commands and flag values of one invocation's "dataplex"
// command group, so invocations don't share state (#2070).
type dataplexCmds struct {
	inv *invocation

	dataplexCmd                      *cobra.Command
	dataplexDatascansCmd             *cobra.Command
	dataplexDatascansRunCmd          *cobra.Command
	dataplexDatascansJobsCmd         *cobra.Command
	dataplexDatascansJobsListCmd     *cobra.Command
	dataplexDatascansJobsDescribeCmd *cobra.Command
	dataplexDatascansDescribeCmd     *cobra.Command
	dataplexDatascansListCmd         *cobra.Command
	flagDataplexListFormat           string
	flagDataplexListURI              bool
	dataplexDatascansCreateCmd       *cobra.Command
	flagDataplexDataSource           string
	flagDataplexScanType             string
	flagDataplexCatalogPublish       bool
	flagDataplexProfileMode          string
	dataplexDatascansDeleteCmd       *cobra.Command
	flagDataplexLocation             string
	flagDataplexDatascan             string
	flagDataplexJobFormat            string
	flagDataplexJobView              string
	flagDataplexJobURI               bool
	flagDataplexCtxResources         []string
	flagDataplexCtxAllSchema         bool
	flagDataplexCtxFormat            string
	flagDataplexCtxOptionsJSON       string
	dataplexDataProductsCmd          *cobra.Command
	dataplexDataAssetsCmd            *cobra.Command
	dpCreateCmd                      *cobra.Command
	dpUpdateCmd                      *cobra.Command
	dpDeleteCmd                      *cobra.Command
	dpDescribeCmd                    *cobra.Command
	dpListCmd                        *cobra.Command
	dpGetIamCmd                      *cobra.Command
	dpSetIamCmd                      *cobra.Command
	dpAddIamCmd                      *cobra.Command
	dpRemoveIamCmd                   *cobra.Command
	daCreateCmd                      *cobra.Command
	daUpdateCmd                      *cobra.Command
	daDeleteCmd                      *cobra.Command
	daDescribeCmd                    *cobra.Command
	daListCmd                        *cobra.Command
	flagDPLocation                   string
	flagDPDataProduct                string
	flagDPFormat                     string
	flagDPAsync                      bool
	flagDPDisplayName                string
	flagDPDescription                string
	flagDPOwnerEmails                []string
	flagDPAccessGroups               []string
	flagDPLabels                     map[string]string
	flagDPUpdateLabels               map[string]string
	flagDPRemoveLabels               []string
	flagDPClearLabels                bool
	flagDPEtag                       string
	flagDPResource                   string
	flagDPMember                     string
	flagDPRole                       string
	flagDPLimit                      int
	flagDPPageSize                   int64
	flagDPURI                        bool
}

// newDataplexCmd returns the "dataplex" command group for inv.
func newDataplexCmd(inv *invocation) *cobra.Command {
	return newDataplexCmds(inv).dataplexCmd
}

// newDataplexCmds builds the "dataplex" command group for inv and returns its state.
func newDataplexCmds(inv *invocation) *dataplexCmds {
	g := &dataplexCmds{inv: inv}
	g.dataplexCmd = &cobra.Command{
		Use:   "dataplex",
		Short: "Manage Cloud Dataplex",
	}
	g.dataplexDatascansCmd = &cobra.Command{
		Use:   "datascans",
		Short: "Manage Dataplex data scans",
	}
	g.dataplexDatascansRunCmd = &cobra.Command{
		Use:   "run DATASCAN_ID",
		Short: "Run an on-demand data scan",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runDataplexDatascansRun,
	}
	g.dataplexDatascansJobsCmd = &cobra.Command{
		Use:   "jobs",
		Short: "Manage data scan jobs",
	}
	g.dataplexDatascansJobsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List data scan jobs",
		Args:  cobra.NoArgs,
		RunE:  g.runDataplexDatascansJobsList,
	}
	g.dataplexDatascansJobsDescribeCmd = &cobra.Command{
		Use:   "describe JOB_ID",
		Short: "Describe a data scan job",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runDataplexDatascansJobsDescribe,
	}
	g.dataplexDatascansDescribeCmd = &cobra.Command{
		Use:   "describe DATASCAN_ID",
		Short: "Describe a data scan",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runDataplexDatascansDescribe,
	}
	g.dataplexDatascansListCmd = &cobra.Command{
		Use:   "list",
		Short: "List data scans",
		Args:  cobra.NoArgs,
		RunE:  g.runDataplexDatascansList,
	}
	g.dataplexDatascansCreateCmd = &cobra.Command{
		Use:   "create DATASCAN_ID",
		Short: "Create a data scan",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runDataplexDatascansCreate,
	}
	g.dataplexDatascansDeleteCmd = &cobra.Command{
		Use:   "delete DATASCAN_ID",
		Short: "Delete a data scan",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runDataplexDatascansDelete,
	}
	g.dataplexDataProductsCmd = &cobra.Command{Use: "data-products", Short: "Manage Dataplex data products"}
	g.dataplexDataAssetsCmd = &cobra.Command{Use: "data-assets", Short: "Manage Dataplex data assets"}
	g.dpCreateCmd = &cobra.Command{Use: "create DATA_PRODUCT", Short: "Create a Dataplex data product",
		Args: cobra.ExactArgs(1), RunE: g.runDPCreate}
	g.dpUpdateCmd = &cobra.Command{Use: "update DATA_PRODUCT", Short: "Update a Dataplex data product",
		Args: cobra.ExactArgs(1), RunE: g.runDPUpdate}
	g.dpDeleteCmd = &cobra.Command{Use: "delete DATA_PRODUCT", Short: "Delete a Dataplex data product",
		Args: cobra.ExactArgs(1), RunE: g.runDPDelete}
	g.dpDescribeCmd = &cobra.Command{Use: "describe DATA_PRODUCT", Short: "Describe a Dataplex data product",
		Args: cobra.ExactArgs(1), RunE: g.runDPDescribe}
	g.dpListCmd = &cobra.Command{Use: "list", Short: "List Dataplex data products",
		Args: cobra.NoArgs, RunE: g.runDPList}
	g.dpGetIamCmd = &cobra.Command{Use: "get-iam-policy DATA_PRODUCT", Short: "Get the IAM policy of a Dataplex data product",
		Args: cobra.ExactArgs(1), RunE: g.runDPGetIam}
	g.dpSetIamCmd = &cobra.Command{Use: "set-iam-policy DATA_PRODUCT POLICY_FILE", Short: "Set the IAM policy of a Dataplex data product",
		Args: cobra.ExactArgs(2), RunE: g.runDPSetIam}
	g.dpAddIamCmd = &cobra.Command{Use: "add-iam-policy-binding DATA_PRODUCT", Short: "Add an IAM policy binding to a Dataplex data product",
		Args: cobra.ExactArgs(1), RunE: g.runDPAddIam}
	g.dpRemoveIamCmd = &cobra.Command{Use: "remove-iam-policy-binding DATA_PRODUCT", Short: "Remove an IAM policy binding from a Dataplex data product",
		Args: cobra.ExactArgs(1), RunE: g.runDPRemoveIam}
	g.daCreateCmd = &cobra.Command{Use: "create DATA_ASSET", Short: "Create a Dataplex data asset",
		Args: cobra.ExactArgs(1), RunE: g.runDACreate}
	g.daUpdateCmd = &cobra.Command{Use: "update DATA_ASSET", Short: "Update a Dataplex data asset",
		Args: cobra.ExactArgs(1), RunE: g.runDAUpdate}
	g.daDeleteCmd = &cobra.Command{Use: "delete DATA_ASSET", Short: "Delete a Dataplex data asset",
		Args: cobra.ExactArgs(1), RunE: g.runDADelete}
	g.daDescribeCmd = &cobra.Command{Use: "describe DATA_ASSET", Short: "Describe a Dataplex data asset",
		Args: cobra.ExactArgs(1), RunE: g.runDADescribe}
	g.daListCmd = &cobra.Command{Use: "list", Short: "List Dataplex data assets in a data product",
		Args: cobra.NoArgs, RunE: g.runDAList}
	g.initDataplex()
	g.initDataplexDataProducts()
	return g
}
