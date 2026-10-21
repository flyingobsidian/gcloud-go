package cmd

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	compute "google.golang.org/api/compute/v1"
)

// --- Compute Engine list commands (#1998) ---
//
// Each command mirrors its gcloud-python counterpart's scopes and default
// table (surface/compute/<group>/list.py and the group's flags.py).

func init() {
	for _, g := range []struct {
		name, short string
		list        *cobra.Command
	}{
		{"instance-templates", "Manage instance templates", newComputeListCmd("List Compute Engine instance templates", instanceTemplatesLister)},
		{"snapshots", "Manage disk snapshots", newComputeListCmd("List Compute Engine snapshots", snapshotsLister)},
		{"firewall-rules", "Manage VPC firewall rules", newComputeListCmd("List Compute Engine firewall rules", firewallRulesLister)},
		{"routers", "Manage Cloud Routers", newComputeListCmd("List Compute Engine routers", routersLister)},
		{"backend-services", "Manage backend services", newComputeListCmd("List backend services", backendServicesLister)},
		{"health-checks", "Manage health checks", newComputeListCmd("List health checks", healthChecksLister)},
		{"url-maps", "Manage URL maps", newComputeListCmd("List URL maps", urlMapsLister)},
		{"ssl-certificates", "Manage SSL certificates", newComputeListCmd("List Compute Engine SSL certificates", sslCertificatesLister)},
	} {
		group := &cobra.Command{Use: g.name, Short: g.short}
		group.AddCommand(g.list)
		registerStubCommand(group, "describe", "Not yet implemented")
		computeCmd.AddCommand(group)
	}
	disksCmd.AddCommand(newComputeListCmd("List Compute Engine disks", disksLister))
}

// --- default-table transforms, after api_lib/compute/transforms.py ---

// computeBasename is gcloud's basename() transform.
func computeBasename(url string) string {
	if url == "" {
		return ""
	}
	return path.Base(url)
}

// computeLocation and computeLocationScope are gcloud's location() and
// location_scope() transforms: the zone or region name, and which it is.
func computeLocation(zone, region string) string {
	if zone != "" {
		return computeBasename(zone)
	}
	return computeBasename(region)
}

func computeLocationScope(zone, region string) string {
	switch {
	case zone != "":
		return "zone"
	case region != "":
		return "region"
	}
	return ""
}

// computeScope is gcloud's scope() transform: the part of url after
// /regions/ or /zones/, else its last segment.
func computeScope(url string) string {
	if !strings.Contains(url, "/") {
		return url
	}
	for _, seg := range []string{"/regions/", "/zones/"} {
		if i := strings.LastIndex(url, seg); i >= 0 {
			return url[i+len(seg):]
		}
	}
	if strings.HasPrefix(url, "https://") {
		return path.Base(url)
	}
	return url
}

// computeScopedSuffix is gcloud's path_simplifier.ScopedSuffix.
func computeScopedSuffix(url string) string {
	for _, seg := range []string{"/zones/", "/regions/"} {
		if i := strings.LastIndex(url, seg); i >= 0 {
			return url[i+len(seg):]
		}
	}
	return path.Base(url)
}

// computeTypeSuffix is gcloud's type_suffix() transform: the last two path
// segments of url.
func computeTypeSuffix(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return url
	}
	return strings.Join(parts[len(parts)-2:], "/")
}

var computeCustomMachineType = regexp.MustCompile(`([a-zA-Z0-9]+)-custom-([0-9]+)-([0-9]+)`)

// computeMachineType is gcloud's machine_type() transform, which spells out
// custom machine types.
func computeMachineType(name string) string {
	m := computeCustomMachineType.FindStringSubmatch(name)
	if m == nil {
		return name
	}
	mib, err := strconv.ParseFloat(m[3], 64)
	if err != nil {
		return name
	}
	return fmt.Sprintf("custom (%s, %s vCPU, %.2f GiB)", m[1], m[2], mib/1024)
}

// computeFirewallRule is gcloud's firewall_rule() transform over a list of
// rules, joined with commas: PROTOCOL:PORT_RANGE, or PROTOCOL alone.
func computeFirewallRule(protocol string, ports []string) string {
	if len(ports) == 0 {
		return protocol
	}
	rules := make([]string, len(ports))
	for i, p := range ports {
		rules[i] = protocol + ":" + p
	}
	return strings.Join(rules, ",")
}

// computeYesNo is gcloud's yesno(yes=..., no=...) for booleans.
func computeYesNo(b bool, yes, no string) string {
	if b {
		return yes
	}
	return no
}

// --- instance-templates ---

var instanceTemplatesLister = computeLister[*compute.InstanceTemplate]{
	noun: "instance templates",
	aggregated: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.InstanceTemplate, error) {
		var out []*compute.InstanceTemplate
		err := svc.InstanceTemplates.AggregatedList(project).Pages(ctx, func(r *compute.InstanceTemplateAggregatedList) error {
			for _, k := range sortedScopes(r.Items) {
				out = append(out, r.Items[k].InstanceTemplates...)
			}
			return nil
		})
		return out, err
	},
	regionalList: func(ctx context.Context, svc *compute.Service, project, region string) ([]*compute.InstanceTemplate, error) {
		var out []*compute.InstanceTemplate
		err := svc.RegionInstanceTemplates.List(project, region).Pages(ctx, func(r *compute.InstanceTemplateList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	globalList: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.InstanceTemplate, error) {
		var out []*compute.InstanceTemplate
		err := svc.InstanceTemplates.List(project).Pages(ctx, func(r *compute.InstanceTemplateList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	location:    func(t *compute.InstanceTemplate) (string, string) { return "", t.Region },
	selfLink:    func(t *compute.InstanceTemplate) string { return t.SelfLink },
	tableFormat: "table(name,machineType,preemptible,creationTimestamp)",
	tableRow: func(t *compute.InstanceTemplate) map[string]any {
		machineType, preemptible := "", ""
		if p := t.Properties; p != nil {
			machineType = computeMachineType(p.MachineType)
			if p.Scheduling != nil {
				preemptible = computeYesNo(p.Scheduling.Preemptible, "true", "")
			}
		}
		return map[string]any{
			"name":              t.Name,
			"machineType":       machineType,
			"preemptible":       preemptible,
			"creationTimestamp": t.CreationTimestamp,
		}
	},
}

// --- disks ---

var disksLister = computeLister[*compute.Disk]{
	noun: "disks",
	aggregated: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.Disk, error) {
		var out []*compute.Disk
		err := svc.Disks.AggregatedList(project).Pages(ctx, func(r *compute.DiskAggregatedList) error {
			for _, k := range sortedScopes(r.Items) {
				out = append(out, r.Items[k].Disks...)
			}
			return nil
		})
		return out, err
	},
	zonalList: func(ctx context.Context, svc *compute.Service, project, zone string) ([]*compute.Disk, error) {
		var out []*compute.Disk
		err := svc.Disks.List(project, zone).Pages(ctx, func(r *compute.DiskList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	regionalList: func(ctx context.Context, svc *compute.Service, project, region string) ([]*compute.Disk, error) {
		var out []*compute.Disk
		err := svc.RegionDisks.List(project, region).Pages(ctx, func(r *compute.DiskList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	location:    func(d *compute.Disk) (string, string) { return d.Zone, d.Region },
	selfLink:    func(d *compute.Disk) string { return d.SelfLink },
	tableFormat: "table(name,location,locationScope,sizeGb,type,status)",
	tableRow: func(d *compute.Disk) map[string]any {
		return map[string]any{
			"name":          d.Name,
			"location":      computeLocation(d.Zone, d.Region),
			"locationScope": computeLocationScope(d.Zone, d.Region),
			"sizeGb":        strconv.FormatInt(d.SizeGb, 10),
			"type":          computeBasename(d.Type),
			"status":        d.Status,
		}
	},
}

// --- snapshots (global only in gcloud GA) ---

var snapshotsLister = computeLister[*compute.Snapshot]{
	noun: "snapshots",
	globalList: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.Snapshot, error) {
		var out []*compute.Snapshot
		err := svc.Snapshots.List(project).Pages(ctx, func(r *compute.SnapshotList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	location:    func(s *compute.Snapshot) (string, string) { return "", "" },
	selfLink:    func(s *compute.Snapshot) string { return s.SelfLink },
	tableFormat: "table(name,diskSizeGb,srcDisk,status)",
	tableRow: func(s *compute.Snapshot) map[string]any {
		return map[string]any{
			"name":       s.Name,
			"diskSizeGb": strconv.FormatInt(s.DiskSizeGb, 10),
			"srcDisk":    computeScope(s.SourceDisk),
			"status":     s.Status,
		}
	},
}

// --- firewall-rules (global) ---

// firewallRulesListNotice is gcloud's epilog for the default firewall table.
const firewallRulesListNotice = "\nTo show all fields of the firewall, please show in JSON format: --format=json\n" +
	"To show all fields in table format, please see the examples in --help.\n\n"

var firewallRulesLister = computeLister[*compute.Firewall]{
	noun: "firewall rules",
	globalList: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.Firewall, error) {
		var out []*compute.Firewall
		err := svc.Firewalls.List(project).Pages(ctx, func(r *compute.FirewallList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	location:    func(f *compute.Firewall) (string, string) { return "", "" },
	selfLink:    func(f *compute.Firewall) string { return f.SelfLink },
	tableFormat: "table(name,network,direction,priority,allow,deny,disabled)",
	notice:      firewallRulesListNotice,
	tableRow: func(f *compute.Firewall) map[string]any {
		var allow, deny []string
		for _, a := range f.Allowed {
			allow = append(allow, computeFirewallRule(a.IPProtocol, a.Ports))
		}
		for _, d := range f.Denied {
			deny = append(deny, computeFirewallRule(d.IPProtocol, d.Ports))
		}
		return map[string]any{
			"name":      f.Name,
			"network":   computeBasename(f.Network),
			"direction": f.Direction,
			"priority":  strconv.FormatInt(f.Priority, 10),
			"allow":     strings.Join(allow, ","),
			"deny":      strings.Join(deny, ","),
			"disabled":  strconv.FormatBool(f.Disabled),
		}
	},
}

// --- routers (regional) ---

var routersLister = computeLister[*compute.Router]{
	noun: "routers",
	aggregated: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.Router, error) {
		var out []*compute.Router
		err := svc.Routers.AggregatedList(project).Pages(ctx, func(r *compute.RouterAggregatedList) error {
			for _, k := range sortedScopes(r.Items) {
				out = append(out, r.Items[k].Routers...)
			}
			return nil
		})
		return out, err
	},
	regionalList: func(ctx context.Context, svc *compute.Service, project, region string) ([]*compute.Router, error) {
		var out []*compute.Router
		err := svc.Routers.List(project, region).Pages(ctx, func(r *compute.RouterList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	location:    func(r *compute.Router) (string, string) { return "", r.Region },
	selfLink:    func(r *compute.Router) string { return r.SelfLink },
	tableFormat: "table(name,region,network)",
	tableRow: func(r *compute.Router) map[string]any {
		network := computeBasename(r.Network)
		if network == "" {
			network = "N/A"
		}
		return map[string]any{
			"name":    r.Name,
			"region":  computeBasename(r.Region),
			"network": network,
		}
	},
}

// --- backend-services (global and regional) ---

var backendServicesLister = computeLister[*compute.BackendService]{
	noun: "backend services",
	aggregated: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.BackendService, error) {
		var out []*compute.BackendService
		err := svc.BackendServices.AggregatedList(project).Pages(ctx, func(r *compute.BackendServiceAggregatedList) error {
			for _, k := range sortedScopes(r.Items) {
				out = append(out, r.Items[k].BackendServices...)
			}
			return nil
		})
		return out, err
	},
	regionalList: func(ctx context.Context, svc *compute.Service, project, region string) ([]*compute.BackendService, error) {
		var out []*compute.BackendService
		err := svc.RegionBackendServices.List(project, region).Pages(ctx, func(r *compute.BackendServiceList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	globalList: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.BackendService, error) {
		var out []*compute.BackendService
		err := svc.BackendServices.List(project).Pages(ctx, func(r *compute.BackendServiceList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	location:    func(b *compute.BackendService) (string, string) { return "", b.Region },
	selfLink:    func(b *compute.BackendService) string { return b.SelfLink },
	tableFormat: "table(name,backends,protocol)",
	tableRow: func(b *compute.BackendService) map[string]any {
		var backends []string
		for _, be := range b.Backends {
			backends = append(backends, computeScopedSuffix(be.Group))
		}
		sort.Strings(backends)
		return map[string]any{
			"name":     b.Name,
			"backends": strings.Join(backends, ","),
			"protocol": b.Protocol,
		}
	},
}

// --- health-checks (global and regional) ---

var healthChecksLister = computeLister[*compute.HealthCheck]{
	noun: "health checks",
	aggregated: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.HealthCheck, error) {
		var out []*compute.HealthCheck
		err := svc.HealthChecks.AggregatedList(project).Pages(ctx, func(r *compute.HealthChecksAggregatedList) error {
			for _, k := range sortedScopes(r.Items) {
				out = append(out, r.Items[k].HealthChecks...)
			}
			return nil
		})
		return out, err
	},
	regionalList: func(ctx context.Context, svc *compute.Service, project, region string) ([]*compute.HealthCheck, error) {
		var out []*compute.HealthCheck
		err := svc.RegionHealthChecks.List(project, region).Pages(ctx, func(r *compute.HealthCheckList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	globalList: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.HealthCheck, error) {
		var out []*compute.HealthCheck
		err := svc.HealthChecks.List(project).Pages(ctx, func(r *compute.HealthCheckList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	location:    func(h *compute.HealthCheck) (string, string) { return "", h.Region },
	selfLink:    func(h *compute.HealthCheck) string { return h.SelfLink },
	tableFormat: "table(name,region,protocol)",
	tableRow: func(h *compute.HealthCheck) map[string]any {
		return map[string]any{
			"name":     h.Name,
			"region":   computeBasename(h.Region),
			"protocol": h.Type,
		}
	},
}

// --- url-maps (global and regional) ---

var urlMapsLister = computeLister[*compute.UrlMap]{
	noun: "URL maps",
	aggregated: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.UrlMap, error) {
		var out []*compute.UrlMap
		err := svc.UrlMaps.AggregatedList(project).Pages(ctx, func(r *compute.UrlMapsAggregatedList) error {
			for _, k := range sortedScopes(r.Items) {
				out = append(out, r.Items[k].UrlMaps...)
			}
			return nil
		})
		return out, err
	},
	regionalList: func(ctx context.Context, svc *compute.Service, project, region string) ([]*compute.UrlMap, error) {
		var out []*compute.UrlMap
		err := svc.RegionUrlMaps.List(project, region).Pages(ctx, func(r *compute.UrlMapList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	globalList: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.UrlMap, error) {
		var out []*compute.UrlMap
		err := svc.UrlMaps.List(project).Pages(ctx, func(r *compute.UrlMapList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	location:    func(u *compute.UrlMap) (string, string) { return "", u.Region },
	selfLink:    func(u *compute.UrlMap) string { return u.SelfLink },
	tableFormat: "table(name,defaultService)",
	tableRow: func(u *compute.UrlMap) map[string]any {
		defaultService := ""
		if u.DefaultService != "" {
			defaultService = computeTypeSuffix(u.DefaultService)
		}
		return map[string]any{
			"name":           u.Name,
			"defaultService": defaultService,
		}
	},
}

// --- ssl-certificates (global and regional) ---

var sslCertificatesLister = computeLister[*compute.SslCertificate]{
	noun: "SSL certificates",
	aggregated: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.SslCertificate, error) {
		var out []*compute.SslCertificate
		err := svc.SslCertificates.AggregatedList(project).Pages(ctx, func(r *compute.SslCertificateAggregatedList) error {
			for _, k := range sortedScopes(r.Items) {
				out = append(out, r.Items[k].SslCertificates...)
			}
			return nil
		})
		return out, err
	},
	regionalList: func(ctx context.Context, svc *compute.Service, project, region string) ([]*compute.SslCertificate, error) {
		var out []*compute.SslCertificate
		err := svc.RegionSslCertificates.List(project, region).Pages(ctx, func(r *compute.SslCertificateList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	globalList: func(ctx context.Context, svc *compute.Service, project string) ([]*compute.SslCertificate, error) {
		var out []*compute.SslCertificate
		err := svc.SslCertificates.List(project).Pages(ctx, func(r *compute.SslCertificateList) error {
			out = append(out, r.Items...)
			return nil
		})
		return out, err
	},
	location:    func(c *compute.SslCertificate) (string, string) { return "", c.Region },
	selfLink:    func(c *compute.SslCertificate) string { return c.SelfLink },
	tableFormat: "table(name,type,creationTimestamp,expireTime,region,managedStatus,domainStatus)",
	tableRow: func(c *compute.SslCertificate) map[string]any {
		managedStatus, domainStatus := "", ""
		if c.Managed != nil {
			managedStatus = c.Managed.Status
			domains := make([]string, 0, len(c.Managed.DomainStatus))
			for _, d := range sortedScopes(c.Managed.DomainStatus) {
				domains = append(domains, d+": "+c.Managed.DomainStatus[d])
			}
			domainStatus = strings.Join(domains, ", ")
		}
		return map[string]any{
			"name":              c.Name,
			"type":              c.Type,
			"creationTimestamp": c.CreationTimestamp,
			"expireTime":        c.ExpireTime,
			"region":            computeBasename(c.Region),
			"managedStatus":     managedStatus,
			"domainStatus":      domainStatus,
		}
	},
}
