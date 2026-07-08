package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	sn "google.golang.org/api/servicenetworking/v1"
)

var servicesPeeredDNSCmd = &cobra.Command{
	Use:   "peered-dns-domains",
	Short: "Manage peered DNS domains for private service connections",
}

var peeredDNSCreateCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "Create a peered DNS domain",
	Args:  cobra.ExactArgs(1),
	RunE:  runPeeredDNSCreate,
}

var peeredDNSDeleteCmd = &cobra.Command{
	Use:   "delete NAME",
	Short: "Delete a peered DNS domain",
	Args:  cobra.ExactArgs(1),
	RunE:  runPeeredDNSDelete,
}

var peeredDNSListCmd = &cobra.Command{
	Use:   "list",
	Short: "List peered DNS domains",
	Args:  cobra.NoArgs,
	RunE:  runPeeredDNSList,
}

var (
	flagPeeredDNSService   string
	flagPeeredDNSNetwork   string
	flagPeeredDNSDnsSuffix string
	flagPeeredDNSListFormat string
)

func init() {
	scopeFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&flagPeeredDNSService, "service", "servicenetworking.googleapis.com", "Peered service (default: servicenetworking.googleapis.com)")
		c.Flags().StringVar(&flagPeeredDNSNetwork, "network", "", "Consumer VPC network name (required)")
		c.MarkFlagRequired("network")
	}
	scopeFlags(peeredDNSCreateCmd)
	peeredDNSCreateCmd.Flags().StringVar(&flagPeeredDNSDnsSuffix, "dns-suffix", "", "DNS suffix (must end with a trailing dot) (required)")
	peeredDNSCreateCmd.MarkFlagRequired("dns-suffix")

	scopeFlags(peeredDNSDeleteCmd)
	scopeFlags(peeredDNSListCmd)
	peeredDNSListCmd.Flags().StringVar(&flagPeeredDNSListFormat, "format", "", "Output format (json, yaml, or table)")

	servicesPeeredDNSCmd.AddCommand(peeredDNSCreateCmd, peeredDNSDeleteCmd, peeredDNSListCmd)
	servicesCmd.AddCommand(servicesPeeredDNSCmd)
}

// peeredDNSParent returns the peered DNS parent resource path:
// `services/{service}/projects/{projectNumber}/global/networks/{network}`.
// The Service Networking API requires a project number (not a project ID)
// here; callers must supply the numeric form via --project.
func peeredDNSParent(service, project, network string) string {
	svc := service
	if !strings.HasPrefix(svc, "services/") {
		svc = "services/" + svc
	}
	return fmt.Sprintf("%s/projects/%s/global/networks/%s", svc, project, network)
}

func runPeeredDNSCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceNetworkingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := peeredDNSParent(flagPeeredDNSService, project, flagPeeredDNSNetwork)
	op, err := svc.Services.Projects.Global.Networks.PeeredDnsDomains.Create(parent, &sn.PeeredDnsDomain{
		Name:      args[0],
		DnsSuffix: flagPeeredDNSDnsSuffix,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating peered DNS domain: %w", err)
	}
	fmt.Printf("Create peered DNS domain in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runPeeredDNSDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceNetworkingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := peeredDNSParent(flagPeeredDNSService, project, flagPeeredDNSNetwork) + "/peeredDnsDomains/" + args[0]
	op, err := svc.Services.Projects.Global.Networks.PeeredDnsDomains.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting peered DNS domain: %w", err)
	}
	fmt.Printf("Delete peered DNS domain in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runPeeredDNSList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceNetworkingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := peeredDNSParent(flagPeeredDNSService, project, flagPeeredDNSNetwork)
	resp, err := svc.Services.Projects.Global.Networks.PeeredDnsDomains.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing peered DNS domains: %w", err)
	}
	return printListResults(resp.PeeredDnsDomains, flagPeeredDNSListFormat, func() {
		fmt.Printf("%-30s %s\n", "NAME", "DNS_SUFFIX")
		for _, d := range resp.PeeredDnsDomains {
			fmt.Printf("%-30s %s\n", d.Name, d.DnsSuffix)
		}
	})
}
