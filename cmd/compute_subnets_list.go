package cmd

import (
	"context"
	"encoding/json"
	"os"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	compute "google.golang.org/api/compute/v1"
)

// --- compute networks subnets list (#2002) ---

var (
	flagSubnetsListRegions []string
	flagSubnetsListNetwork string
	flagSubnetsListURI     bool
)

func init() {
	list := &cobra.Command{
		Use:   "list",
		Short: "List subnetworks",
		Args:  cobra.NoArgs,
		RunE:  runSubnetsList,
	}
	list.Flags().StringSliceVar(&flagSubnetsListRegions, "regions", nil,
		"If regions are given, only subnetworks in those regions are shown")
	list.Flags().StringVar(&flagSubnetsListNetwork, "network", "", "Only show subnetworks of a specific network")
	list.Flags().BoolVar(&flagSubnetsListURI, "uri", false, "Print a list of resource URIs instead of the default output")

	subnets := &cobra.Command{Use: "subnets", Short: "List, create, and delete Compute Engine subnetworks"}
	subnets.AddCommand(list)
	networksCmd.AddCommand(subnets)
}

var subnetsLister = computeLister[*compute.Subnetwork]{
	noun: "subnetworks",
	aggregated: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.Subnetwork, error) {
		var out []*compute.Subnetwork
		err := svc.Subnetworks.AggregatedList(project).Pages(ctx, func(r *compute.SubnetworkAggregatedList) error {
			for _, k := range sortedScopes(r.Items) {
				out = append(out, r.Items[k].Subnetworks...)
			}
			return nil
		})
		return out, err
	},
	regionalList: func(ctx context.Context, svc *compute.Service, project, region string) ([]*compute.Subnetwork, error) {
		var out []*compute.Subnetwork
		err := svc.Subnetworks.List(project, region).Pages(ctx, func(r *compute.SubnetworkList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	location: func(s *compute.Subnetwork) (string, string) { return "", s.Region },
	selfLink: func(s *compute.Subnetwork) string { return s.SelfLink },
	tableFormat: "table(name,region,network,range,stackType,ipv6AccessType,internalIpv6Prefix," +
		"externalIpv6Prefix,utilizationDetails)",
	tableRow: func(s *compute.Subnetwork) map[string]any {
		utilization := ""
		if s.UtilizationDetails != nil {
			if b, err := json.Marshal(s.UtilizationDetails); err == nil {
				utilization = string(b)
			}
		}
		return map[string]any{
			"name":               s.Name,
			"region":             computeBasename(s.Region),
			"network":            computeBasename(s.Network),
			"range":              s.IpCidrRange,
			"stackType":          s.StackType,
			"ipv6AccessType":     s.Ipv6AccessType,
			"internalIpv6Prefix": s.InternalIpv6Prefix,
			"externalIpv6Prefix": s.ExternalIpv6Prefix,
			"utilizationDetails": utilization,
		}
	},
}

// filterSubnetsByNetwork keeps the subnetworks of the named network, as
// gcloud's --network does; an empty network keeps them all.
func filterSubnetsByNetwork(subnets []*compute.Subnetwork, network string) []*compute.Subnetwork {
	if network == "" {
		return subnets
	}
	kept := []*compute.Subnetwork{}
	for _, s := range subnets {
		if s.Network != "" && computeBasename(s.Network) == network {
			kept = append(kept, s)
		}
	}
	return kept
}

func runSubnetsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := computeListRequest{regions: flagSubnetsListRegions, regionsSet: cmd.Flags().Changed("regions")}
	subnets, err := subnetsLister.list(ctx, svc, project, req)
	if err != nil {
		return err
	}
	return subnetsLister.emit(os.Stdout, os.Stderr, filterSubnetsByNetwork(subnets, flagSubnetsListNetwork),
		flagFormat, flagSubnetsListURI)
}
