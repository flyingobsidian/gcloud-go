package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
)

// --- gcloud dns project-info (#1540) ---

var dnsPICmd = &cobra.Command{Use: "project-info", Short: "View Cloud DNS project info"}

var flagDNSPIFormat string

var dnsPIDescribeCmd = &cobra.Command{
	Use: "describe", Short: "Describe the current Cloud DNS project info",
	Args: cobra.NoArgs, RunE: runDNSPIDescribe,
}

func init() {
	dnsPIDescribeCmd.Flags().StringVar(&flagDNSPIFormat, "format", "", "Output format")
	dnsPICmd.AddCommand(dnsPIDescribeCmd)
	dnsCmd.AddCommand(dnsPICmd)
}

func runDNSPIDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Get(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing DNS project info: %w", err)
	}
	return emitFormatted(got, flagDNSPIFormat)
}
