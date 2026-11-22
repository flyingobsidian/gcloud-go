package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/googleapi"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database ... update (#1968) ---
//
// gcloud-python 589.0.0 added labels-only update commands (declarative,
// read_modify_update) to several oracle-database groups. Each gets the
// resource, applies --labels or --update-labels/--remove-labels/--clear-labels,
// patches it with updateMask=labels and, unless --async, waits for the
// operation and prints the updated resource.

// odbLabelsTarget describes one resource type with a labels-only update.
type odbLabelsTarget struct {
	kind       string // gcloud-python's resource name, e.g. cloudVmCluster
	collection string // e.g. cloudVmClusters
	get        func(ctx context.Context, svc *oracledatabase.Service, name string) (resource any, labels map[string]string, err error)
	patch      func(ctx context.Context, svc *oracledatabase.Service, name string, labels map[string]string, opts ...googleapi.CallOption) (*oracledatabase.Operation, error)
}

// odbLabelsUpdateFlags holds one update command's flag values.
type odbLabelsUpdateFlags struct {
	location     string
	requestID    string
	async        bool
	labels       map[string]string
	updateLabels map[string]string
	removeLabels []string
	clearLabels  bool
}

// newODBLabelsUpdateCmd returns the "update RESOURCE" command for t.
func (s *oracleDatabaseCmds) newODBLabelsUpdateCmd(t odbLabelsTarget) *cobra.Command {
	f := &odbLabelsUpdateFlags{}
	c := &cobra.Command{
		Use:   "update " + strings.ToUpper(camelToSnake(t.kind)),
		Short: "Update labels of a " + t.kind,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.runODBLabelsUpdate(cmd, args[0], t, f)
		},
	}
	fl := c.Flags()
	fl.StringVar(&f.location, "location", "", "The location id of the "+t.kind+" resource")
	fl.StringVar(&f.requestID, "request-id", "", "An optional ID to identify the request (a UUID)")
	fl.BoolVar(&f.async, "async", false, "Return immediately, without waiting for the operation in progress to complete")
	fl.StringToStringVar(&f.labels, "labels", nil, "Set the labels to these KEY=VALUE pairs, replacing any existing labels")
	fl.StringToStringVar(&f.updateLabels, "update-labels", nil, "List of label KEY=VALUE pairs to update; new keys are added")
	fl.StringSliceVar(&f.removeLabels, "remove-labels", nil, "List of label keys to remove")
	fl.BoolVar(&f.clearLabels, "clear-labels", false, "Remove all labels")
	return c
}

func (s *oracleDatabaseCmds) runODBLabelsUpdate(cmd *cobra.Command, id string, t odbLabelsTarget, f *odbLabelsUpdateFlags) error {
	changed := cmd.Flags().Changed
	// gcloud-python's labels group: --labels | (--update-labels,
	// --clear-labels | --remove-labels).
	if changed("labels") && (changed("update-labels") || changed("clear-labels") || changed("remove-labels")) {
		return errors.New("argument --labels: At most one of --labels | --update-labels --clear-labels | --remove-labels can be specified.")
	}
	if changed("clear-labels") && changed("remove-labels") {
		return errors.New("argument --clear-labels: At most one of --clear-labels | --remove-labels can be specified.")
	}
	if !changed("labels") && !changed("update-labels") && !changed("remove-labels") && !changed("clear-labels") {
		return errors.New("Must specify at least one valid parameter to update.")
	}
	name, err := s.odbResource(f.location, t.collection, id)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	_, current, err := t.get(ctx, svc, name)
	if err != nil {
		return fmt.Errorf("describing %s: %w", t.kind, err)
	}
	labels := f.labels
	if !changed("labels") {
		labels, _ = applyLabelsDiff(current, f.clearLabels, f.removeLabels, f.updateLabels)
	}
	var opts []googleapi.CallOption
	if f.requestID != "" {
		opts = append(opts, googleapi.QueryParameter("requestId", f.requestID))
	}
	op, err := t.patch(ctx, svc, name, labels, opts...)
	if err != nil {
		return fmt.Errorf("updating %s: %w", t.kind, err)
	}
	id = cdLastSegment(name)
	fmt.Fprintf(s.inv.stderr, "Request issued for: [%s]\n", id)
	if f.async {
		fmt.Fprintf(s.inv.stderr, "Check operation [%s] for status.\n", op.Name)
		fmt.Fprintf(s.inv.stderr, "Updated %s [%s].\n", t.kind, id)
		return emitFormattedTo(s.inv.stdout, op, s.inv.flags.format)
	}
	if err := s.odbWaitOp(ctx, svc, op); err != nil {
		return err
	}
	got, _, err := t.get(ctx, svc, name)
	if err != nil {
		return fmt.Errorf("describing %s: %w", t.kind, err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated %s [%s].\n", t.kind, id)
	return emitFormattedTo(s.inv.stdout, got, s.inv.flags.format)
}

// odbWaitOp prints gcloud's "Waiting for operation" progress line and polls
// op until it is done.
func (s *oracleDatabaseCmds) odbWaitOp(ctx context.Context, svc *oracledatabase.Service, op *oracledatabase.Operation) error {
	fmt.Fprintf(s.inv.stderr, "Waiting for operation [%s] to complete...", op.Name)
	interval := time.Second
	for !op.Done {
		select {
		case <-ctx.Done():
			fmt.Fprintln(s.inv.stderr, "failed.")
			return ctx.Err()
		case <-time.After(interval):
		}
		if interval *= 2; interval > 10*time.Second {
			interval = 10 * time.Second
		}
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			fmt.Fprintln(s.inv.stderr, "failed.")
			return fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		fmt.Fprintln(s.inv.stderr, "failed.")
		return errors.New(op.Error.Message)
	}
	fmt.Fprintln(s.inv.stderr, "done.")
	return nil
}

// odbLabelsTargets returns the resource types whose update commands
// gcloud-python 589.0.0 added, keyed by command group.
func odbLabelsTargets() map[string]odbLabelsTarget {
	return map[string]odbLabelsTarget{
		"cloud-exadata-infrastructures": {
			kind: "cloudExadataInfrastructure", collection: "cloudExadataInfrastructures",
			get: func(ctx context.Context, svc *oracledatabase.Service, name string) (any, map[string]string, error) {
				r, err := svc.Projects.Locations.CloudExadataInfrastructures.Get(name).Context(ctx).Do()
				if err != nil {
					return nil, nil, err
				}
				return r, r.Labels, nil
			},
			patch: func(ctx context.Context, svc *oracledatabase.Service, name string, labels map[string]string, opts ...googleapi.CallOption) (*oracledatabase.Operation, error) {
				return svc.Projects.Locations.CloudExadataInfrastructures.Patch(name, &oracledatabase.CloudExadataInfrastructure{Labels: labels}).
					UpdateMask("labels").Context(ctx).Do(opts...)
			},
		},
		"cloud-vm-clusters": {
			kind: "cloudVmCluster", collection: "cloudVmClusters",
			get: func(ctx context.Context, svc *oracledatabase.Service, name string) (any, map[string]string, error) {
				r, err := svc.Projects.Locations.CloudVmClusters.Get(name).Context(ctx).Do()
				if err != nil {
					return nil, nil, err
				}
				return r, r.Labels, nil
			},
			patch: func(ctx context.Context, svc *oracledatabase.Service, name string, labels map[string]string, opts ...googleapi.CallOption) (*oracledatabase.Operation, error) {
				return svc.Projects.Locations.CloudVmClusters.Patch(name, &oracledatabase.CloudVmCluster{Labels: labels}).
					UpdateMask("labels").Context(ctx).Do(opts...)
			},
		},
		"db-systems": {
			kind: "dbSystem", collection: "dbSystems",
			get: func(ctx context.Context, svc *oracledatabase.Service, name string) (any, map[string]string, error) {
				r, err := svc.Projects.Locations.DbSystems.Get(name).Context(ctx).Do()
				if err != nil {
					return nil, nil, err
				}
				return r, r.Labels, nil
			},
			patch: func(ctx context.Context, svc *oracledatabase.Service, name string, labels map[string]string, opts ...googleapi.CallOption) (*oracledatabase.Operation, error) {
				return svc.Projects.Locations.DbSystems.Patch(name, &oracledatabase.DbSystem{Labels: labels}).
					UpdateMask("labels").Context(ctx).Do(opts...)
			},
		},
		"exascale-db-storage-vaults": {
			kind: "exascaleDbStorageVault", collection: "exascaleDbStorageVaults",
			get: func(ctx context.Context, svc *oracledatabase.Service, name string) (any, map[string]string, error) {
				r, err := svc.Projects.Locations.ExascaleDbStorageVaults.Get(name).Context(ctx).Do()
				if err != nil {
					return nil, nil, err
				}
				return r, r.Labels, nil
			},
			patch: func(ctx context.Context, svc *oracledatabase.Service, name string, labels map[string]string, opts ...googleapi.CallOption) (*oracledatabase.Operation, error) {
				return svc.Projects.Locations.ExascaleDbStorageVaults.Patch(name, &oracledatabase.ExascaleDbStorageVault{Labels: labels}).
					UpdateMask("labels").Context(ctx).Do(opts...)
			},
		},
		"goldengate-connections": {
			kind: "goldengateConnection", collection: "goldengateConnections",
			get: func(ctx context.Context, svc *oracledatabase.Service, name string) (any, map[string]string, error) {
				r, err := svc.Projects.Locations.GoldengateConnections.Get(name).Context(ctx).Do()
				if err != nil {
					return nil, nil, err
				}
				return r, r.Labels, nil
			},
			patch: func(ctx context.Context, svc *oracledatabase.Service, name string, labels map[string]string, opts ...googleapi.CallOption) (*oracledatabase.Operation, error) {
				return svc.Projects.Locations.GoldengateConnections.Patch(name, &oracledatabase.GoldengateConnection{Labels: labels}).
					UpdateMask("labels").Context(ctx).Do(opts...)
			},
		},
		"goldengate-connection-assignments": {
			kind: "goldengateConnectionAssignment", collection: "goldengateConnectionAssignments",
			get: func(ctx context.Context, svc *oracledatabase.Service, name string) (any, map[string]string, error) {
				r, err := svc.Projects.Locations.GoldengateConnectionAssignments.Get(name).Context(ctx).Do()
				if err != nil {
					return nil, nil, err
				}
				return r, r.Labels, nil
			},
			patch: func(ctx context.Context, svc *oracledatabase.Service, name string, labels map[string]string, opts ...googleapi.CallOption) (*oracledatabase.Operation, error) {
				return svc.Projects.Locations.GoldengateConnectionAssignments.Patch(name, &oracledatabase.GoldengateConnectionAssignment{Labels: labels}).
					UpdateMask("labels").Context(ctx).Do(opts...)
			},
		},
		"goldengate-deployments": {
			kind: "goldengateDeployment", collection: "goldengateDeployments",
			get: func(ctx context.Context, svc *oracledatabase.Service, name string) (any, map[string]string, error) {
				r, err := svc.Projects.Locations.GoldengateDeployments.Get(name).Context(ctx).Do()
				if err != nil {
					return nil, nil, err
				}
				return r, r.Labels, nil
			},
			patch: func(ctx context.Context, svc *oracledatabase.Service, name string, labels map[string]string, opts ...googleapi.CallOption) (*oracledatabase.Operation, error) {
				return svc.Projects.Locations.GoldengateDeployments.Patch(name, &oracledatabase.GoldengateDeployment{Labels: labels}).
					UpdateMask("labels").Context(ctx).Do(opts...)
			},
		},
	}
}

// initOracleDatabaseLabelsUpdates adds the labels-only update command to each
// group in odbLabelsTargets; it runs after the groups are populated.
func (s *oracleDatabaseCmds) initOracleDatabaseLabelsUpdates() {
	targets := odbLabelsTargets()
	for _, g := range s.oracleDatabaseCmd.Commands() {
		if t, ok := targets[g.Name()]; ok {
			g.AddCommand(s.newODBLabelsUpdateCmd(t))
		}
	}
}
