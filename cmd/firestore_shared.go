package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

// firestoreDatabaseName returns projects/PROJECT/databases/DB. If db is already
// fully qualified it is returned as-is.
func firestoreDatabaseName(project, db string) string {
	if strings.HasPrefix(db, "projects/") {
		return db
	}
	return fmt.Sprintf("projects/%s/databases/%s", project, db)
}

// firestoreLocationName returns projects/PROJECT/locations/LOCATION.
func firestoreLocationName(project, location string) string {
	if strings.HasPrefix(location, "projects/") {
		return location
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

// firestoreAddDatabaseFlag registers --database on cmd.
func firestoreAddDatabaseFlag(cmd *cobra.Command, target *string, required bool) {
	cmd.Flags().StringVar(target, "database", "", "Firestore database ID (defaults to '(default)')")
	if required {
		_ = cmd.MarkFlagRequired("database")
	}
}

// firestoreWaitOp polls a Firestore admin operation to completion.
func firestoreWaitOp(ctx context.Context, svc *firestore.Service, op *GoogleLongrunningOperation) (*GoogleLongrunningOperation, error) {
	for !op.Done {
		got, err := svc.Projects.Databases.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = fromFirestoreOp(got)
	}
	if op.Error != "" {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error)
	}
	return op, nil
}

// GoogleLongrunningOperation collapses the firestore Operation to a shape our
// helpers can share across subgroups. Only the fields we actually use are
// retained.
type GoogleLongrunningOperation struct {
	Name     string
	Done     bool
	Error    string
	Response any
}

func fromFirestoreOp(op *firestore.GoogleLongrunningOperation) *GoogleLongrunningOperation {
	out := &GoogleLongrunningOperation{Name: op.Name, Done: op.Done}
	if op.Error != nil {
		out.Error = op.Error.Message
	}
	if op.Response != nil {
		out.Response = op.Response
	}
	return out
}

// firestoreFinishOp either reports the operation name (async) or waits.
func firestoreFinishOp(ctx context.Context, svc *firestore.Service, op *firestore.GoogleLongrunningOperation, verb, name string, async bool) error {
	wrapped := fromFirestoreOp(op)
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, wrapped.Name)
		return emitFormatted(op, "")
	}
	final, err := firestoreWaitOp(ctx, svc, wrapped)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}
