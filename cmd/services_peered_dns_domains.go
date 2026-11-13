package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	sn "google.golang.org/api/servicenetworking/v1"
)

func (g *servicesCmds) initServicesPeeredDnsDomains() {
	scopeFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&g.flagPeeredDNSService, "service", "servicenetworking.googleapis.com", "Peered service (default: servicenetworking.googleapis.com)")
		c.Flags().StringVar(&g.flagPeeredDNSNetwork, "network", "", "Consumer VPC network name (required)")
		c.MarkFlagRequired("network")
	}
	scopeFlags(g.peeredDNSCreateCmd)
	g.peeredDNSCreateCmd.Flags().StringVar(&g.flagPeeredDNSDnsSuffix, "dns-suffix", "", "DNS suffix (must end with a trailing dot) (required)")
	g.peeredDNSCreateCmd.MarkFlagRequired("dns-suffix")

	scopeFlags(g.peeredDNSDeleteCmd)
	scopeFlags(g.peeredDNSListCmd)
	g.peeredDNSListCmd.Flags().StringVar(&g.flagPeeredDNSListFormat, "format", "", "Output format (json, yaml, or table)")

	g.servicesPeeredDNSCmd.AddCommand(g.peeredDNSCreateCmd, g.peeredDNSDeleteCmd, g.peeredDNSListCmd)
	g.servicesCmd.AddCommand(g.servicesPeeredDNSCmd)
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

func (g *servicesCmds) runPeeredDNSCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceNetworking(ctx)
	if err != nil {
		return err
	}
	parent := peeredDNSParent(g.flagPeeredDNSService, project, g.flagPeeredDNSNetwork)
	op, err := svc.Services.Projects.Global.Networks.PeeredDnsDomains.Create(parent, &sn.PeeredDnsDomain{
		Name:      args[0],
		DnsSuffix: g.flagPeeredDNSDnsSuffix,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating peered DNS domain: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Create peered DNS domain in progress (operation: %s).\n", op.Name)
	return g.inv.yaml(op)
}

func (g *servicesCmds) runPeeredDNSDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceNetworking(ctx)
	if err != nil {
		return err
	}
	name := peeredDNSParent(g.flagPeeredDNSService, project, g.flagPeeredDNSNetwork) + "/peeredDnsDomains/" + args[0]
	op, err := svc.Services.Projects.Global.Networks.PeeredDnsDomains.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting peered DNS domain: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Delete peered DNS domain in progress (operation: %s).\n", op.Name)
	return g.inv.yaml(op)
}

func (g *servicesCmds) runPeeredDNSList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceNetworking(ctx)
	if err != nil {
		return err
	}
	parent := peeredDNSParent(g.flagPeeredDNSService, project, g.flagPeeredDNSNetwork)
	resp, err := svc.Services.Projects.Global.Networks.PeeredDnsDomains.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing peered DNS domains: %w", err)
	}
	return g.inv.printListResults(resp.PeeredDnsDomains, g.flagPeeredDNSListFormat, func() {
		printTableHeadingTo(g.inv.stdout, len(resp.PeeredDnsDomains), "%-30s %s\n", "NAME", "DNS_SUFFIX")
		for _, d := range resp.PeeredDnsDomains {
			fmt.Fprintf(g.inv.stdout, "%-30s %s\n", d.Name, d.DnsSuffix)
		}
	})
}
