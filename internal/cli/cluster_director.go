package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	hypercomputecluster "google.golang.org/api/hypercomputecluster/v1"
	"gopkg.in/yaml.v3"
)

// --- gcloud cluster-director (#1946) ---
//
// gcloud-python promoted cluster-director to GA in 583.0.0. The commands use
// hypercomputecluster v1. clusters create and update take the cluster body
// from --config (a JSON object or a file); Python's granular flags
// (--create-network, --slurm-node-sets, --quickstart-cluster, ...) are not
// ported.

func (s *clusterDirectorCmds) initClusterDirector() {
	clusters := []*cobra.Command{s.cdClCreateCmd, s.cdClDeleteCmd, s.cdClDescribeCmd, s.cdClListCmd, s.cdClUpdateCmd}
	for _, c := range clusters {
		c.Flags().StringVar(&s.flagCDLocation, "location", "", "Location ID of the cluster resource")
		c.Flags().StringVar(&s.flagCDFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.cdClCreateCmd, s.cdClDeleteCmd, s.cdClUpdateCmd} {
		c.Flags().BoolVar(&s.flagCDAsync, "async", false,
			"Return immediately, without waiting for the operation in progress to complete")
	}
	for _, c := range []*cobra.Command{s.cdClCreateCmd, s.cdClUpdateCmd} {
		c.Flags().StringVar(&s.flagCDConfig, "config", "",
			"Configuration of the cluster specs in the form of a JSON object (or a file containing one) (required)")
		_ = c.MarkFlagRequired("config")
	}
	s.cdClUpdateCmd.Flags().StringVar(&s.flagCDUpdateMask, "update-mask", "",
		`Update mask to specify the fields to update, for example "description,labels" (required)`)
	_ = s.cdClUpdateCmd.MarkFlagRequired("update-mask")
	for _, c := range []*cobra.Command{s.cdClListCmd, s.cdLocListCmd} {
		c.Flags().Int64Var(&s.flagCDPageSize, "page-size", 0, "Maximum results per page")
		c.Flags().IntVar(&s.flagCDLimit, "limit", 0, "Maximum number of resources to list")
	}
	s.cdOpDescribeCmd.Flags().StringVar(&s.flagCDLocation, "location", "", "The location id of the operation resource")
	for _, c := range []*cobra.Command{s.cdOpDescribeCmd, s.cdLocDescribeCmd, s.cdLocListCmd} {
		c.Flags().StringVar(&s.flagCDFormat, "format", "", "Output format")
	}

	s.cdClustersCmd.AddCommand(clusters...)
	s.cdOperationsCmd.AddCommand(s.cdOpDescribeCmd)
	s.cdLocationsCmd.AddCommand(s.cdLocDescribeCmd, s.cdLocListCmd)
	s.clusterDirectorCmd.AddCommand(s.cdClustersCmd, s.cdOperationsCmd, s.cdLocationsCmd)
}

// cdName resolves a positional (an ID with --location, or a full name) to
// projects/P/locations/L/COLLECTION/ID.
func (s *clusterDirectorCmds) cdName(collection, arg string) (string, error) {
	if strings.HasPrefix(arg, "projects/") {
		return arg, nil
	}
	parent, err := s.cdParent()
	if err != nil {
		return "", err
	}
	return parent + "/" + collection + "/" + arg, nil
}

// cdParent returns projects/P/locations/L from --location.
func (s *clusterDirectorCmds) cdParent() (string, error) {
	if s.flagCDLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagCDLocation), nil
}

// cdLastSegment returns the ID at the end of a resource name.
func cdLastSegment(name string) string {
	return name[strings.LastIndex(name, "/")+1:]
}

// cdConfig decodes --config, a JSON (or YAML) object or the path of a file
// holding one, into a Cluster, rejecting unknown fields as Python's
// DictToMessageWithErrorCheck does.
func (s *clusterDirectorCmds) cdConfig() (*hypercomputecluster.Cluster, error) {
	data := []byte(s.flagCDConfig)
	if info, err := os.Stat(s.flagCDConfig); err == nil && info.Mode().IsRegular() {
		if data, err = os.ReadFile(s.flagCDConfig); err != nil {
			return nil, fmt.Errorf("reading --config: %w", err)
		}
	}
	var generic any
	if err := yaml.Unmarshal(data, &generic); err != nil {
		return nil, fmt.Errorf("invalid --config: %w", err)
	}
	if _, ok := convertYAMLKeys(generic).(map[string]any); !ok {
		return nil, fmt.Errorf("invalid --config: expected a JSON object")
	}
	jsonBytes, err := json.Marshal(convertYAMLKeys(generic))
	if err != nil {
		return nil, fmt.Errorf("invalid --config: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(jsonBytes))
	dec.DisallowUnknownFields()
	cluster := &hypercomputecluster.Cluster{}
	if err := dec.Decode(cluster); err != nil {
		return nil, fmt.Errorf("invalid --config: %w", err)
	}
	return cluster, nil
}

// cdWait prints gcloud's "<verb> request issued" line and, unless --async,
// waits for op. It returns the finished operation, or nil with --async.
func (s *clusterDirectorCmds) cdWait(ctx context.Context, svc *hypercomputecluster.Service, op *hypercomputecluster.Operation, verb, id string) (*hypercomputecluster.Operation, error) {
	fmt.Fprintf(s.inv.stderr, "%s request issued for: [%s]\n", verb, id)
	if s.flagCDAsync {
		return nil, nil
	}
	fmt.Fprintf(s.inv.stderr, "Waiting for operation [%s] to complete...", op.Name)
	interval := time.Second
	for !op.Done {
		time.Sleep(interval)
		if interval *= 2; interval > 10*time.Second {
			interval = 10 * time.Second
		}
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			fmt.Fprintln(s.inv.stderr, "failed.")
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		fmt.Fprintln(s.inv.stderr, "failed.")
		return nil, fmt.Errorf("%s", op.Error.Message)
	}
	fmt.Fprintln(s.inv.stderr, "done.")
	return op, nil
}

// cdCreateOrUpdate finishes a create or update: with --async it prints the
// operation, otherwise it waits and prints the resulting cluster.
func (s *clusterDirectorCmds) cdCreateOrUpdate(ctx context.Context, svc *hypercomputecluster.Service, op *hypercomputecluster.Operation, verb, name string) error {
	id := cdLastSegment(name)
	done, err := s.cdWait(ctx, svc, op, verb, id)
	if err != nil {
		return err
	}
	if done == nil {
		return emitFormattedTo(s.inv.stdout, op, s.flagCDFormat)
	}
	got := &hypercomputecluster.Cluster{}
	if err := json.Unmarshal(done.Response, got); err != nil {
		return fmt.Errorf("decoding operation response: %w", err)
	}
	past := map[string]string{"Create": "Created", "Update": "Updated"}[verb]
	fmt.Fprintf(s.inv.stderr, "%s cluster [%s].\n", past, id)
	return emitFormattedTo(s.inv.stdout, got, s.flagCDFormat)
}

func (s *clusterDirectorCmds) runCDClCreate(cmd *cobra.Command, args []string) error {
	name, err := s.cdName("clusters", args[0])
	if err != nil {
		return err
	}
	cluster, err := s.cdConfig()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().HypercomputeCluster(ctx)
	if err != nil {
		return err
	}
	parent := name[:strings.LastIndex(name, "/clusters/")]
	op, err := svc.Projects.Locations.Clusters.Create(parent, cluster).ClusterId(cdLastSegment(name)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	return s.cdCreateOrUpdate(ctx, svc, op, "Create", name)
}

func (s *clusterDirectorCmds) runCDClUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.cdName("clusters", args[0])
	if err != nil {
		return err
	}
	cluster, err := s.cdConfig()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().HypercomputeCluster(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Patch(name, cluster).UpdateMask(s.flagCDUpdateMask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating cluster: %w", err)
	}
	return s.cdCreateOrUpdate(ctx, svc, op, "Update", name)
}

func (s *clusterDirectorCmds) runCDClDelete(cmd *cobra.Command, args []string) error {
	name, err := s.cdName("clusters", args[0])
	if err != nil {
		return err
	}
	id := cdLastSegment(name)
	if err := s.inv.promptContinue(fmt.Sprintf("You are about to delete cluster [%s]", id), ""); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().HypercomputeCluster(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	done, err := s.cdWait(ctx, svc, op, "Delete", id)
	if err != nil {
		return err
	}
	if done == nil {
		fmt.Fprintf(s.inv.stderr, "Check operation [%s] for status.\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagCDFormat)
	}
	fmt.Fprintf(s.inv.stderr, "Deleted cluster [%s].\n", id)
	return nil
}

func (s *clusterDirectorCmds) runCDClDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.cdName("clusters", args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().HypercomputeCluster(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cluster: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCDFormat)
}

func (s *clusterDirectorCmds) runCDClList(cmd *cobra.Command, args []string) error {
	parent, err := s.cdParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().HypercomputeCluster(ctx)
	if err != nil {
		return err
	}
	var all []*hypercomputecluster.Cluster
	pageToken := ""
	for {
		call := svc.Projects.Locations.Clusters.List(parent).Context(ctx)
		if s.flagCDPageSize > 0 {
			call = call.PageSize(s.flagCDPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing clusters: %w", err)
		}
		all = append(all, resp.Clusters...)
		if s.flagCDLimit > 0 && len(all) >= s.flagCDLimit {
			all = all[:s.flagCDLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagCDFormat)
}

func (s *clusterDirectorCmds) runCDOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.cdName("operations", args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().HypercomputeCluster(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCDFormat)
}

func (s *clusterDirectorCmds) runCDLocDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]
	if !strings.HasPrefix(name, "projects/") {
		project, err := s.inv.project()
		if err != nil {
			return err
		}
		name = fmt.Sprintf("projects/%s/locations/%s", project, name)
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().HypercomputeCluster(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCDFormat)
}

func (s *clusterDirectorCmds) runCDLocList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().HypercomputeCluster(ctx)
	if err != nil {
		return err
	}
	var all []*hypercomputecluster.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List("projects/" + project).Context(ctx)
		if s.flagCDPageSize > 0 {
			call = call.PageSize(s.flagCDPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if s.flagCDLimit > 0 && len(all) >= s.flagCDLimit {
			all = all[:s.flagCDLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagCDFormat)
}

// clusterDirectorCmds holds the commands and flag values of one invocation's
// "cluster-director" command group, so invocations don't share state (#2070).
type clusterDirectorCmds struct {
	inv *invocation

	clusterDirectorCmd *cobra.Command
	cdClustersCmd      *cobra.Command
	cdClCreateCmd      *cobra.Command
	cdClDeleteCmd      *cobra.Command
	cdClDescribeCmd    *cobra.Command
	cdClListCmd        *cobra.Command
	cdClUpdateCmd      *cobra.Command
	cdOperationsCmd    *cobra.Command
	cdOpDescribeCmd    *cobra.Command
	cdLocationsCmd     *cobra.Command
	cdLocDescribeCmd   *cobra.Command
	cdLocListCmd       *cobra.Command

	flagCDLocation   string
	flagCDFormat     string
	flagCDAsync      bool
	flagCDConfig     string
	flagCDUpdateMask string
	flagCDPageSize   int64
	flagCDLimit      int
}

// newClusterDirectorCmd returns the "cluster-director" command group for inv.
func newClusterDirectorCmd(inv *invocation) *cobra.Command {
	return newClusterDirectorCmds(inv).clusterDirectorCmd
}

// newClusterDirectorCmds builds the "cluster-director" command group for inv
// and returns its state.
func newClusterDirectorCmds(inv *invocation) *clusterDirectorCmds {
	s := &clusterDirectorCmds{inv: inv}
	s.clusterDirectorCmd = &cobra.Command{Use: "cluster-director", Short: "Manage Cluster Director resources"}
	s.cdClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage Cluster Director clusters"}
	s.cdClCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Creates a Cluster Director resource (cluster spec from --config)",
		Args: cobra.ExactArgs(1), RunE: s.runCDClCreate,
	}
	s.cdClDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Deletes a Cluster Director resource",
		Args: cobra.ExactArgs(1), RunE: s.runCDClDelete,
	}
	s.cdClDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Gets details of a Cluster Director resource",
		Args: cobra.ExactArgs(1), RunE: s.runCDClDescribe,
	}
	s.cdClListCmd = &cobra.Command{
		Use: "list", Short: "List Cluster Director resources",
		Args: cobra.NoArgs, RunE: s.runCDClList,
	}
	s.cdClUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Updates a Cluster Director resource (fields from --config and --update-mask)",
		Args: cobra.ExactArgs(1), RunE: s.runCDClUpdate,
	}
	s.cdOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Cluster Director operations"}
	s.cdOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: s.runCDOpDescribe,
	}
	s.cdLocationsCmd = &cobra.Command{Use: "locations", Short: "Manage Cluster Director locations"}
	s.cdLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a location",
		Args: cobra.ExactArgs(1), RunE: s.runCDLocDescribe,
	}
	s.cdLocListCmd = &cobra.Command{
		Use: "list", Short: "List locations",
		Args: cobra.NoArgs, RunE: s.runCDLocList,
	}
	s.initClusterDirector()
	return s
}
