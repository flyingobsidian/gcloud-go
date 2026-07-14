package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

// eventarcLocationParent returns projects/PROJECT/locations/REGION.
func eventarcLocationParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

// eventarcResourceName returns a fully qualified resource name for a per-region
// Eventarc collection. If id is already fully qualified it is returned as-is.
func eventarcResourceName(collection, id, project, region string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/locations/%s/%s/%s", project, region, collection, id)
}

// eventarcAddRegionFlag registers the --location flag on cmd. Eventarc's Python
// surface uses --location (not --region) as the primary form.
func eventarcAddRegionFlag(cmd *cobra.Command, target *string, required bool) {
	cmd.Flags().StringVar(target, "location", "", "Location of the Eventarc resource (required)")
	if required {
		_ = cmd.MarkFlagRequired("location")
	}
}

// eventarcWaitOp polls a long-running operation until it completes.
func eventarcWaitOp(ctx context.Context, svc *eventarc.Service, op *eventarc.GoogleLongrunningOperation) (*eventarc.GoogleLongrunningOperation, error) {
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

// eventarcFinishOp reports the outcome of an LRO, mirroring the pattern used by
// the database-migration commands.
func eventarcFinishOp(ctx context.Context, svc *eventarc.Service, op *eventarc.GoogleLongrunningOperation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	op, err := eventarcWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if op.Response != nil {
		return emitFormatted(op.Response, "")
	}
	return nil
}
