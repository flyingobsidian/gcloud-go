package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	cloudiam "google.golang.org/api/iam/v1"
)

// --- gcloud iam workforce-pools subjects (#1930) ---
//
// Declarative (YAML) commands in gcloud-python on the
// iam.locations.workforcePools.subjects collection; each returns an LRO
// that is waited on unless --async.

func (s *iamCmds) initIamWorkforcePoolsSubjects() {
	for _, c := range []*cobra.Command{s.iamWpSubjDeleteCmd, s.iamWpSubjUndeleteCmd, s.iamWpSubjRevokeSessionsCmd} {
		c.Flags().StringVar(&s.flagIamWpSubjLocation, "location", "", "The location for the workforce pool")
		c.Flags().StringVar(&s.flagIamWpSubjPool, "workforce-pool", "", "The ID of the workforce pool")
		c.Flags().BoolVar(&s.flagIamWpSubjAsync, "async", false,
			"Return immediately, without waiting for the operation in progress to complete")
		c.Flags().StringVar(&s.flagIamWpSubjFormat, "format", "", "Output format")
	}
	s.iamWpSubjectsCmd.AddCommand(s.iamWpSubjDeleteCmd, s.iamWpSubjUndeleteCmd, s.iamWpSubjRevokeSessionsCmd)
	s.iamWorkforcePoolsCmd.AddCommand(s.iamWpSubjectsCmd)
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
func (s *iamCmds) iamWpSubjFinishOp(ctx context.Context, svc *cloudiam.Service, op *cloudiam.Operation, requestMsg string) (*cloudiam.Operation, error) {
	fmt.Fprintln(s.inv.stderr, requestMsg)
	if s.flagIamWpSubjAsync {
		fmt.Fprintf(s.inv.stderr, "Check operation [%s] for status.\n", op.Name)
		return nil, nil
	}
	fmt.Fprintf(s.inv.stderr, "Waiting for operation [%s] to complete...", op.Name)
	for !op.Done {
		time.Sleep(time.Second)
		got, err := svc.Locations.WorkforcePools.Subjects.Operations.Get(op.Name).Context(ctx).Do()
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

func (s *iamCmds) runIamWpSubjDelete(cmd *cobra.Command, args []string) error {
	name, id, err := iamWpSubjectName(args[0], s.flagIamWpSubjLocation, s.flagIamWpSubjPool)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Subjects.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting workforce pool subject: %w", err)
	}
	final, err := s.iamWpSubjFinishOp(ctx, svc, op, fmt.Sprintf("Delete request issued for: [%s]", id))
	if err != nil {
		return err
	}
	// gcloud-python's delete is a SilentCommand: no stdout unless --format.
	if final == nil {
		if s.flagIamWpSubjFormat != "" {
			return emitFormattedTo(s.inv.stdout, op, s.flagIamWpSubjFormat)
		}
		return nil
	}
	fmt.Fprintf(s.inv.stderr, "Deleted workforce pool subject [%s].\n", id)
	if s.flagIamWpSubjFormat != "" && len(final.Response) > 0 {
		return emitFormattedTo(s.inv.stdout, final.Response, s.flagIamWpSubjFormat)
	}
	return nil
}

func (s *iamCmds) runIamWpSubjUndelete(cmd *cobra.Command, args []string) error {
	name, id, err := iamWpSubjectName(args[0], s.flagIamWpSubjLocation, s.flagIamWpSubjPool)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Subjects.Undelete(name, &cloudiam.UndeleteWorkforcePoolSubjectRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting workforce pool subject: %w", err)
	}
	return s.iamWpSubjGenericResult(ctx, svc, op, id)
}

func (s *iamCmds) runIamWpSubjRevokeSessions(cmd *cobra.Command, args []string) error {
	name, id, err := iamWpSubjectName(args[0], s.flagIamWpSubjLocation, s.flagIamWpSubjPool)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Subjects.RevokeSessions(name, &cloudiam.RevokeWorkforcePoolSubjectSessionsRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("revoking workforce pool subject sessions: %w", err)
	}
	return s.iamWpSubjGenericResult(ctx, svc, op, id)
}

// iamWpSubjGenericResult finishes a generic (custom-method) declarative
// command: with --async the operation is printed, otherwise the operation's
// response, if any.
func (s *iamCmds) iamWpSubjGenericResult(ctx context.Context, svc *cloudiam.Service, op *cloudiam.Operation, id string) error {
	final, err := s.iamWpSubjFinishOp(ctx, svc, op, fmt.Sprintf("Request issued for: [%s]", id))
	if err != nil {
		return err
	}
	if final == nil {
		return emitFormattedTo(s.inv.stdout, op, s.flagIamWpSubjFormat)
	}
	if len(final.Response) > 0 {
		return emitFormattedTo(s.inv.stdout, final.Response, s.flagIamWpSubjFormat)
	}
	return nil
}
