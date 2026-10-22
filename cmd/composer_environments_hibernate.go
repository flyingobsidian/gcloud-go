package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	composerbeta "google.golang.org/api/composer/v1beta1"
)

// --- gcloud composer environments hibernate|resume (#1923) ---
//
// gcloud-python 583.0.0 added these alpha/beta commands; the Hibernate and
// Resume methods exist only on the Composer v1beta1 API.

var (
	composerEnvHibernateCmd = &cobra.Command{
		Use: "hibernate ENVIRONMENT", Short: "Hibernate a Cloud Composer environment",
		Args: cobra.ExactArgs(1), RunE: composerEnvPowerFn(composerHibernate),
	}
	composerEnvResumeCmd = &cobra.Command{
		Use: "resume ENVIRONMENT", Short: "Resume a Cloud Composer environment",
		Args: cobra.ExactArgs(1), RunE: composerEnvPowerFn(composerResume),
	}
	flagComposerEnvPowerAsync bool
)

// composerPowerAction describes one of hibernate/resume.
type composerPowerAction struct {
	participle, gerund string
	call               func(ctx context.Context, svc *composerbeta.Service, name string) (*composerbeta.Operation, error)
}

var (
	composerHibernate = composerPowerAction{
		participle: "hibernated", gerund: "hibernating",
		call: func(ctx context.Context, svc *composerbeta.Service, name string) (*composerbeta.Operation, error) {
			return svc.Projects.Locations.Environments.Hibernate(name, &composerbeta.HibernateEnvironmentRequest{}).Context(ctx).Do()
		},
	}
	composerResume = composerPowerAction{
		participle: "resumed", gerund: "resuming",
		call: func(ctx context.Context, svc *composerbeta.Service, name string) (*composerbeta.Operation, error) {
			return svc.Projects.Locations.Environments.Resume(name, &composerbeta.ResumeEnvironmentRequest{}).Context(ctx).Do()
		},
	}
)

func init() {
	for _, c := range []*cobra.Command{composerEnvHibernateCmd, composerEnvResumeCmd} {
		c.Flags().StringVar(&flagComposerEnvLocation, "location", "", "Composer location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().BoolVar(&flagComposerEnvPowerAsync, "async", false,
			"Return immediately, without waiting for the operation in progress to complete")
		c.Flags().StringVar(&flagComposerEnvFormat, "format", "", "Output format")
	}
	composerEnvCmd.AddCommand(composerEnvHibernateCmd, composerEnvResumeCmd)
}

// composerOpPollInterval is the initial delay between operation polls; it
// doubles up to gcloud-python's 15s ceiling.
var composerOpPollInterval = 2 * time.Second

func composerEnvPowerFn(a composerPowerAction) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		name, err := composerEnvResolvedName(args[0])
		if err != nil {
			return err
		}
		ctx := context.Background()
		svc, err := gcp.ComposerBetaService(ctx, flagAccount)
		if err != nil {
			return err
		}
		op, err := a.call(ctx, svc, name)
		if err != nil {
			return fmt.Errorf("Error %s environment [%s]: %w", a.gerund, name, err)
		}
		if flagComposerEnvPowerAsync {
			fmt.Fprintf(os.Stderr, "Update in progress for environment [%s] with operation [%s].\n", name, op.Name)
			return emitFormatted(op, flagComposerEnvFormat)
		}
		fmt.Fprintf(os.Stderr, "Waiting for [%s] to be %s with [%s]...", name, a.participle, op.Name)
		if err := composerBetaWaitOp(ctx, svc, op); err != nil {
			fmt.Fprintln(os.Stderr, "failed.")
			return fmt.Errorf("Error %s environment [%s]: %w", a.gerund, name, err)
		}
		fmt.Fprintln(os.Stderr, "done.")
		return nil
	}
}

// composerBetaWaitOp polls op until done (at most one hour, as in Python)
// and returns gcloud-python's OperationError text on failure.
func composerBetaWaitOp(ctx context.Context, svc *composerbeta.Service, op *composerbeta.Operation) error {
	deadline := time.Now().Add(time.Hour)
	interval := composerOpPollInterval
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
