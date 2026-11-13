package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	gkeonprem "google.golang.org/api/gkeonprem/v1"
)

// --- gcloud container vmware (#1143) ---
//
// Manages Anthos-on-VMware user clusters via the gkeonprem v1 client.

func (s *containerCmds) initContainerVmware() {
	all := []*cobra.Command{
		s.containerVmClCreateCmd, s.containerVmClDeleteCmd, s.containerVmClDescribeCmd,
		s.containerVmClListCmd, s.containerVmClUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagCvmLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagCvmFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.containerVmClCreateCmd, s.containerVmClUpdateCmd} {
		c.Flags().StringVar(&s.flagCvmConfigFile, "config-file", "", "YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.containerVmClUpdateCmd.Flags().StringVar(&s.flagCvmUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	s.containerVmClListCmd.Flags().Int64Var(&s.flagCvmPageSize, "page-size", 0, "Maximum results per page")

	s.containerVmwareClustersCmd.AddCommand(all...)
	s.containerVmwareCmd.AddCommand(s.containerVmwareClustersCmd)
	s.containerCmd.AddCommand(s.containerVmwareCmd)
}

func (s *containerCmds) cvmLocationParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if s.flagCvmLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagCvmLocation), nil
}

func (s *containerCmds) cvmClusterName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.cvmLocationParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/vmwareClusters/%s", parent, id), nil
}

func (s *containerCmds) runCvmClCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.cvmLocationParent()
	if err != nil {
		return err
	}
	body := &gkeonprem.VmwareCluster{}
	if err := loadYAMLOrJSONInto(s.flagCvmConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEOnPrem(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VmwareClusters.Create(parent, body).VmwareClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating VMware cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create VMware cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCvmFormat)
}

func (s *containerCmds) runCvmClDelete(cmd *cobra.Command, args []string) error {
	name, err := s.cvmClusterName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEOnPrem(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VmwareClusters.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting VMware cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete VMware cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCvmFormat)
}

func (s *containerCmds) runCvmClDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.cvmClusterName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEOnPrem(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.VmwareClusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing VMware cluster: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCvmFormat)
}

func (s *containerCmds) runCvmClList(cmd *cobra.Command, args []string) error {
	parent, err := s.cvmLocationParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEOnPrem(ctx)
	if err != nil {
		return err
	}
	var all []*gkeonprem.VmwareCluster
	pageToken := ""
	for {
		call := svc.Projects.Locations.VmwareClusters.List(parent).Context(ctx)
		if s.flagCvmPageSize > 0 {
			call = call.PageSize(s.flagCvmPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing VMware clusters: %w", err)
		}
		all = append(all, resp.VmwareClusters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagCvmFormat)
}

func (s *containerCmds) runCvmClUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.cvmClusterName(args[0])
	if err != nil {
		return err
	}
	body := &gkeonprem.VmwareCluster{}
	if err := loadYAMLOrJSONInto(s.flagCvmConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagCvmUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEOnPrem(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.VmwareClusters.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating VMware cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update VMware cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCvmFormat)
}
