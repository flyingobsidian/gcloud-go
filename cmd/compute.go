package cmd

import (
	"github.com/spf13/cobra"
)

var computeCmd = &cobra.Command{
	Use:   "compute",
	Short: "Compute Engine commands",
}

var instancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Manage Compute Engine instances",
}

var instanceGroupsCmd = &cobra.Command{
	Use:   "instance-groups",
	Short: "Manage instance groups",
}

var unmanagedCmd = &cobra.Command{
	Use:   "unmanaged",
	Short: "Manage unmanaged instance groups",
}

func init() {
	instanceGroupsCmd.AddCommand(unmanagedCmd)
	computeCmd.AddCommand(instancesCmd)
	computeCmd.AddCommand(instanceGroupsCmd)

	// Stub registrations for compute subgroups present in gcloud-python but
	// not yet implemented in gcloud-go (#539). Each is a placeholder so
	// `--help` lists the surface; per-subgroup PRs replace these with real
	// behavior. Grouped in registration to keep the diff scannable.
	for name, short := range map[string]string{
		"accelerator-types":                 "GPU / accelerator types",
		"advice":                            "Advisor recommendations",
		"backend-buckets":                   "Manage backend buckets",
		"backend-services":                  "Manage backend services",
		"commitments":                       "Manage committed use discounts",
		"composite-health-checks":           "Manage composite health checks",
		"connect-to-serial-port":            "Connect to a VM's serial port",
		"copy-files":                        "Copy files to/from a VM (deprecated in favor of scp)",
		"diagnose":                          "Diagnose instance issues",
		"disks":                             "Manage Compute Engine disks",
		"disk-types":                        "List available disk types",
		"external-vpn-gateways":             "Manage external VPN gateways",
		"firewall-policies":                 "Manage hierarchical firewall policies",
		"firewall-rules":                    "Manage VPC firewall rules",
		"future-reservations":               "Manage future capacity reservations",
		"global-vm-extension-policies":      "Manage global VM extension policies",
		"health-aggregation-policies":       "Manage health aggregation policies",
		"health-checks":                     "Manage health checks",
		"health-sources":                    "Manage health sources",
		"http-health-checks":                "Manage legacy HTTP health checks",
		"https-health-checks":               "Manage legacy HTTPS health checks",
		"images":                            "Manage disk images",
		"instance-templates":                "Manage instance templates",
		"instant-snapshot-groups":           "Manage instant snapshot groups",
		"instant-snapshots":                 "Manage instant snapshots",
		"interconnects":                     "Manage Cloud Interconnect",
		"machine-images":                    "Manage machine images",
		"machine-types":                     "List machine types",
		"migration":                         "Migrate to Google Cloud",
		"network-attachments":               "Manage network attachments",
		"network-edge-security-services":    "Manage network edge security services",
		"network-endpoint-groups":           "Manage network endpoint groups",
		"network-firewall-policies":         "Manage network firewall policies",
		"network-profiles":                  "Manage network profiles",
		"operations":                        "Manage long-running operations",
		"org-security-policies":             "Manage organization security policies",
		"os-config":                         "Manage OS Config resources",
		"os-login":                          "Manage OS Login",
		"packet-mirrorings":                 "Manage packet mirroring",
		"preview-features":                  "Manage preview features",
		"project-zonal-metadata":            "Manage project zonal metadata",
		"public-advertised-prefixes":        "Manage public advertised prefixes",
		"public-delegated-prefixes":         "Manage public delegated prefixes",
		"regions":                           "List Compute Engine regions",
		"reservations":                      "Manage capacity reservations",
		"resource-policies":                 "Manage resource policies",
		"rollout-plans":                     "Manage rollout plans",
		"rollouts":                          "Manage rollouts",
		"routers":                           "Manage Cloud Routers",
		"routes":                            "Manage VPC routes",
		"security-policies":                 "Manage Cloud Armor security policies",
		"service-attachments":               "Manage service attachments",
		"shared-vpc":                        "Manage Shared VPC",
		"snapshots":                         "Manage disk snapshots",
		"snapshot-settings":                 "Manage snapshot settings",
		"sole-tenancy":                      "Manage sole-tenant nodes",
		"ssl-certificates":                  "Manage SSL certificates",
		"ssl-policies":                      "Manage SSL policies",
		"storage-pools":                     "Manage storage pools",
		"storage-pool-types":                "List storage pool types",
		"target-grpc-proxies":               "Manage target gRPC proxies",
		"target-http-proxies":               "Manage target HTTP proxies",
		"target-https-proxies":              "Manage target HTTPS proxies",
		"target-instances":                  "Manage target instances",
		"target-pools":                      "Manage target pools",
		"target-ssl-proxies":                "Manage target SSL proxies",
		"target-tcp-proxies":                "Manage target TCP proxies",
		"target-vpn-gateways":               "Manage Classic VPN gateways",
		"tpus":                              "Manage Cloud TPU nodes",
		"url-maps":                          "Manage URL maps",
		"vpn-gateways":                      "Manage HA VPN gateways",
		"vpn-tunnels":                       "Manage VPN tunnels",
		"zones":                             "List Compute Engine zones",
		"zone-vm-extension-policies":        "Manage zonal VM extension policies",
	} {
		registerStubGroup(computeCmd, name, short, "list", "describe")
	}
	for name, short := range map[string]string{
		"config-ssh":                "Populate ~/.ssh/config with SSH aliases for VMs",
		"reset-windows-password":    "Reset a Windows VM's password",
		"sign-url":                  "Sign a URL for Cloud CDN backend buckets",
		"start-iap-tunnel":          "Open an IAP TCP tunnel to a VM",
	} {
		registerStubCommand(computeCmd, name, short)
	}

	rootCmd.AddCommand(computeCmd)
}
