package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudiam "google.golang.org/api/iam/v1"
)

// --- gcloud iam workforce-pools subjects (#1930) ---
//
// Declarative (YAML) commands in gcloud-python on the
// iam.locations.workforcePools.subjects collection; each returns an LRO
// that is waited on unless --async.

var iamWpSubjectsCmd = &cobra.Command{
	Use:   "subjects",
	Short: "Create and manage workforce pool subjects",
}

var (
	iamWpSubjDeleteCmd = &cobra.Command{
		Use: "delete SUBJECT", Short: "Delete a workforce pool subject",
		Args: cobra.ExactArgs(1), RunE: runIamWpSubjDelete,
	}
	iamWpSubjUndeleteCmd = &cobra.Command{
		Use: "undelete SUBJECT", Short: "Undelete a workforce pool subject",
		Args: cobra.ExactArgs(1), RunE: runIamWpSubjUndelete,
	}
	iamWpSubjRevokeSessionsCmd = &cobra.Command{
		Use: "revoke-sessions SUBJECT", Short: "Revoke all sessions for a workforce pool subject",
		Args: cobra.ExactArgs(1), RunE: runIamWpSubjRevokeSessions,
	}
)

var (
	flagIamWpSubjLocation string
	flagIamWpSubjPool     string
	flagIamWpSubjAsync    bool
	flagIamWpSubjFormat   string
)

func init() {
	for _, c := range []*cobra.Command{iamWpSubjDeleteCmd, iamWpSubjUndeleteCmd, iamWpSubjRevokeSessionsCmd} {
		c.Flags().StringVar(&flagIamWpSubjLocation, "location", "", "The location for the workforce pool")
		c.Flags().StringVar(&flagIamWpSubjPool, "workforce-pool", "", "The ID of the workforce pool")
		c.Flags().BoolVar(&flagIamWpSubjAsync, "async", false,
			"Return immediately, without waiting for the operation in progress to complete")
		c.Flags().StringVar(&flagIamWpSubjFormat, "format", "", "Output format")
	}
	iamWpSubjectsCmd.AddCommand(iamWpSubjDeleteCmd, iamWpSubjUndeleteCmd, iamWpSubjRevokeSessionsCmd)
	iamWorkforcePoolsCmd.AddCommand(iamWpSubjectsCmd)
}

// iamWpSubjectName resolves SUBJECT (an ID or a full
// locations/L/workforcePools/P/subjects/S name) to its relative name and ID.
func iamWpSubjectName(arg, location, pool string) (string, string, error) {
	if strings.HasPrefix(arg, "locations/") {
		parts := strings.Split(arg, "/")
		if len(parts) != 6 || parts[2] != "workforcePools" || parts[4] != "subjects" {
			return "", "", fmt.Errorf("invalid workforce pool subject name [%s]", arg)
		}
		return arg, parts[5], nil
	}
	if location == "" {
		return "", "", fmt.Errorf("--location is required (or pass a fully qualified subject name)")
	}
	if pool == "" {
		return "", "", fmt.Errorf("--workforce-pool is required (or pass a fully qualified subject name)")
	}
	return fmt.Sprintf("locations/%s/workforcePools/%s/subjects/%s", location, pool, arg), arg, nil
}

// iamWpSubjFinishOp prints gcloud's declarative LRO messages and waits for
// the operation unless --async. It returns the finished operation, or nil
// when --async was given.
func iamWpSubjFinishOp(ctx context.Context, svc *cloudiam.Service, op *cloudiam.Operation, requestMsg string) (*cloudiam.Operation, error) {
	fmt.Fprintln(os.Stderr, requestMsg)
	if flagIamWpSubjAsync {
		fmt.Fprintf(os.Stderr, "Check operation [%s] for status.\n", op.Name)
		return nil, nil
	}
	fmt.Fprintf(os.Stderr, "Waiting for operation [%s] to complete...", op.Name)
	for !op.Done {
		time.Sleep(time.Second)
		got, err := svc.Locations.WorkforcePools.Subjects.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed.")
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		fmt.Fprintln(os.Stderr, "failed.")
		return nil, fmt.Errorf("%s", op.Error.Message)
	}
	fmt.Fprintln(os.Stderr, "done.")
	return op, nil
}

func runIamWpSubjDelete(cmd *cobra.Command, args []string) error {
	name, id, err := iamWpSubjectName(args[0], flagIamWpSubjLocation, flagIamWpSubjPool)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Subjects.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting workforce pool subject: %w", err)
	}
	final, err := iamWpSubjFinishOp(ctx, svc, op, fmt.Sprintf("Delete request issued for: [%s]", id))
	if err != nil {
		return err
	}
	// gcloud-python's delete is a SilentCommand: no stdout unless --format.
	if final == nil {
		if flagIamWpSubjFormat != "" {
			return emitFormatted(op, flagIamWpSubjFormat)
		}
		return nil
	}
	fmt.Fprintf(os.Stderr, "Deleted workforce pool subject [%s].\n", id)
	if flagIamWpSubjFormat != "" && len(final.Response) > 0 {
		return emitFormatted(final.Response, flagIamWpSubjFormat)
	}
	return nil
}

func runIamWpSubjUndelete(cmd *cobra.Command, args []string) error {
	name, id, err := iamWpSubjectName(args[0], flagIamWpSubjLocation, flagIamWpSubjPool)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Subjects.Undelete(name, &cloudiam.UndeleteWorkforcePoolSubjectRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting workforce pool subject: %w", err)
	}
	return iamWpSubjGenericResult(ctx, svc, op, id)
}

func runIamWpSubjRevokeSessions(cmd *cobra.Command, args []string) error {
	name, id, err := iamWpSubjectName(args[0], flagIamWpSubjLocation, flagIamWpSubjPool)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Subjects.RevokeSessions(name, &cloudiam.RevokeWorkforcePoolSubjectSessionsRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("revoking workforce pool subject sessions: %w", err)
	}
	return iamWpSubjGenericResult(ctx, svc, op, id)
}

// iamWpSubjGenericResult finishes a generic (custom-method) declarative
// command: with --async the operation is printed, otherwise the operation's
// response, if any.
func iamWpSubjGenericResult(ctx context.Context, svc *cloudiam.Service, op *cloudiam.Operation, id string) error {
	final, err := iamWpSubjFinishOp(ctx, svc, op, fmt.Sprintf("Request issued for: [%s]", id))
	if err != nil {
		return err
	}
	if final == nil {
		return emitFormatted(op, flagIamWpSubjFormat)
	}
	if len(final.Response) > 0 {
		return emitFormatted(final.Response, flagIamWpSubjFormat)
	}
	return nil
}
