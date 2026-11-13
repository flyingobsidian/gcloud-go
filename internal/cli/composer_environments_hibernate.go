package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	composerbeta "google.golang.org/api/composer/v1beta1"
)

// --- gcloud composer environments hibernate|resume (#1923) ---
//
// gcloud-python 583.0.0 added these alpha/beta commands; the Hibernate and
// Resume methods exist only on the Composer v1beta1 API.

// composerPowerAction describes one of hibernate/resume.
type composerPowerAction struct {
	participle, gerund string
	call               func(ctx context.Context, svc *composerbeta.Service, name string) (*composerbeta.Operation, error)
}

func composerHibernate() composerPowerAction {
	return composerPowerAction{
		participle: "hibernated", gerund: "hibernating",
		call: func(ctx context.Context, svc *composerbeta.Service, name string) (*composerbeta.Operation, error) {
			return svc.Projects.Locations.Environments.Hibernate(name, &composerbeta.HibernateEnvironmentRequest{}).Context(ctx).Do()
		},
	}
}

func composerResume() composerPowerAction {
	return composerPowerAction{
		participle: "resumed", gerund: "resuming",
		call: func(ctx context.Context, svc *composerbeta.Service, name string) (*composerbeta.Operation, error) {
			return svc.Projects.Locations.Environments.Resume(name, &composerbeta.ResumeEnvironmentRequest{}).Context(ctx).Do()
		},
	}
}

func (s *composerCmds) initComposerEnvironmentsHibernate() {
	for _, c := range []*cobra.Command{s.composerEnvHibernateCmd, s.composerEnvResumeCmd} {
		c.Flags().StringVar(&s.flagComposerEnvLocation, "location", "", "Composer location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().BoolVar(&s.flagComposerEnvPowerAsync, "async", false,
			"Return immediately, without waiting for the operation in progress to complete")
		c.Flags().StringVar(&s.flagComposerEnvFormat, "format", "", "Output format")
	}
	s.composerEnvCmd.AddCommand(s.composerEnvHibernateCmd, s.composerEnvResumeCmd)
}

// composerOpPollInterval is the initial delay between operation polls; it
// doubles up to gcloud-python's 15s ceiling.
const composerOpPollInterval = 2 * time.Second

func (s *composerCmds) composerEnvPowerFn(a composerPowerAction) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		name, err := s.composerEnvResolvedName(args[0])
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		svc, err := s.inv.clients().ComposerBeta(ctx)
		if err != nil {
			return err
		}
		op, err := a.call(ctx, svc, name)
		if err != nil {
			return fmt.Errorf("Error %s environment [%s]: %w", a.gerund, name, err)
		}
		if s.flagComposerEnvPowerAsync {
			fmt.Fprintf(s.inv.stderr, "Update in progress for environment [%s] with operation [%s].\n", name, op.Name)
			return emitFormattedTo(s.inv.stdout, op, s.flagComposerEnvFormat)
		}
		fmt.Fprintf(s.inv.stderr, "Waiting for [%s] to be %s with [%s]...", name, a.participle, op.Name)
		if err := composerBetaWaitOp(ctx, svc, op, composerOpPollInterval); err != nil {
			fmt.Fprintln(s.inv.stderr, "failed.")
			return fmt.Errorf("Error %s environment [%s]: %w", a.gerund, name, err)
		}
		fmt.Fprintln(s.inv.stderr, "done.")
		return nil
	}
}

// composerBetaWaitOp polls op until done (at most one hour, as in Python)
// and returns gcloud-python's OperationError text on failure.
func composerBetaWaitOp(ctx context.Context, svc *composerbeta.Service, op *composerbeta.Operation, interval time.Duration) error {
	deadline := time.Now().Add(time.Hour)
	for !op.Done {
		if time.Now().After(deadline) {
			return fmt.Errorf("Operation [%s] timed out", op.Name)
		}
		time.Sleep(interval)
		if interval *= 2; interval > 15*time.Second {
			interval = 15 * time.Second
		}
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return err
		}
		op = got
	}
	if op.Error != nil {
		return fmt.Errorf("Operation [%s] failed: %s", op.Name, op.Error.Message)
	}
	return nil
}
