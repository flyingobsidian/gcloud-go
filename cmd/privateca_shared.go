package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	privateca "google.golang.org/api/privateca/v1"
)

// privatecaLocationParent returns projects/PROJECT/locations/LOCATION.
func privatecaLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

// pcaWaitOp polls the given operation to completion.
func pcaWaitOp(ctx context.Context, svc *privateca.Service, op *privateca.Operation) (*privateca.Operation, error) {
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

func pcaFinishOp(ctx context.Context, svc *privateca.Service, op *privateca.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := pcaWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

// pcaAddIAMFlags registers --member and --role on cmd (both required).
func pcaAddIAMFlags(cmd *cobra.Command, member, role *string) {
	cmd.Flags().StringVar(member, "member", "", "IAM member, e.g. user:foo@example.com (required)")
	cmd.Flags().StringVar(role, "role", "", "IAM role, e.g. roles/privateca.viewer (required)")
	_ = cmd.MarkFlagRequired("member")
	_ = cmd.MarkFlagRequired("role")
}

// pcaAddBinding adds member to the given role in policy, creating the binding
// if necessary.
func pcaAddBinding(policy *privateca.Policy, role, member string) {
	for _, b := range policy.Bindings {
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
	policy.Bindings = append(policy.Bindings, &privateca.Binding{Role: role, Members: []string{member}})
}

// pcaRemoveBinding removes member from the given role in policy.
func pcaRemoveBinding(policy *privateca.Policy, role, member string) {
	for _, b := range policy.Bindings {
		if b.Role != role {
			continue
		}
		filtered := b.Members[:0]
		for _, m := range b.Members {
			if m != member {
				filtered = append(filtered, m)
			}
		}
		b.Members = filtered
	}
}

// pcaStripPrefix removes trailing "projects/…" prefix disambiguation: if id is
// already fully qualified, use it; else compose with parent + collection.
func pcaResourceName(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return parent + "/" + collection + "/" + id
}
