package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud telco-automation operations (#847) ---

var taOpCmd = &cobra.Command{Use: "operations", Short: "Manage Telco Automation operations"}

var (
	flagTAOpLocation string
	flagTAOpFormat   string
	flagTAOpTimeout  time.Duration
)

var (
	taOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Telco Automation operation",
		Args: cobra.ExactArgs(1), RunE: runTAOpDescribe,
	}
	taOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Telco Automation operation to finish",
		Args: cobra.ExactArgs(1), RunE: runTAOpWait,
	}
)

func init() {
	all := []*cobra.Command{taOpDescribeCmd, taOpWaitCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagTAOpLocation, "location", "", "Telco Automation location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagTAOpFormat, "format", "", "Output format")
	}
	taOpWaitCmd.Flags().DurationVar(&flagTAOpTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the operation to finish")

	taOpCmd.AddCommand(all...)
	telcoAutomationCmd.AddCommand(taOpCmd)
}

func taOpQualifiedName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s/operations/%s", project, flagTAOpLocation, id), nil
}

func runTAOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := taOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := telcoAutomationRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagTAOpFormat)
}

func runTAOpWait(cmd *cobra.Command, args []string) error {
	name, err := taOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	op, err := telcoAutomationRest.waitOperation(ctx, name, flagTAOpTimeout)
	if err != nil {
		return err
	}
	fmt.Printf("Operation %s completed.\n", args[0])
	return emitFormatted(op, flagTAOpFormat)
}
