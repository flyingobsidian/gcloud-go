package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	gkeonprem "google.golang.org/api/gkeonprem/v1"
)

// --- gcloud container bare-metal (#1134) ---
//
// Manages Anthos-on-bare-metal user clusters via the gkeonprem v1 client.

func (s *containerCmds) initContainerBareMetal() {
	all := []*cobra.Command{
		s.containerBmClCreateCmd, s.containerBmClDeleteCmd, s.containerBmClDescribeCmd,
		s.containerBmClListCmd, s.containerBmClUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagCbmLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagCbmFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.containerBmClCreateCmd, s.containerBmClUpdateCmd} {
		c.Flags().StringVar(&s.flagCbmConfigFile, "config-file", "", "YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.containerBmClUpdateCmd.Flags().StringVar(&s.flagCbmUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	s.containerBmClListCmd.Flags().Int64Var(&s.flagCbmPageSize, "page-size", 0, "Maximum results per page")

	s.containerBareMetalClustersCmd.AddCommand(all...)
	s.containerBareMetalCmd.AddCommand(s.containerBareMetalClustersCmd)
	s.containerCmd.AddCommand(s.containerBareMetalCmd)
}

func (s *containerCmds) cbmLocationParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if s.flagCbmLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagCbmLocation), nil
}

func (s *containerCmds) cbmClusterName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.cbmLocationParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/bareMetalClusters/%s", parent, id), nil
}

func (s *containerCmds) runCbmClCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.cbmLocationParent()
	if err != nil {
		return err
	}
	body := &gkeonprem.BareMetalCluster{}
	if err := loadYAMLOrJSONInto(s.flagCbmConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEOnPrem(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BareMetalClusters.Create(parent, body).BareMetalClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating bare-metal cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create bare-metal cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCbmFormat)
}

func (s *containerCmds) runCbmClDelete(cmd *cobra.Command, args []string) error {
	name, err := s.cbmClusterName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEOnPrem(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BareMetalClusters.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting bare-metal cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete bare-metal cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCbmFormat)
}

func (s *containerCmds) runCbmClDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.cbmClusterName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEOnPrem(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.BareMetalClusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing bare-metal cluster: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCbmFormat)
}

func (s *containerCmds) runCbmClList(cmd *cobra.Command, args []string) error {
	parent, err := s.cbmLocationParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEOnPrem(ctx)
	if err != nil {
		return err
	}
	var all []*gkeonprem.BareMetalCluster
	pageToken := ""
	for {
		call := svc.Projects.Locations.BareMetalClusters.List(parent).Context(ctx)
		if s.flagCbmPageSize > 0 {
			call = call.PageSize(s.flagCbmPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing bare-metal clusters: %w", err)
		}
		all = append(all, resp.BareMetalClusters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagCbmFormat)
}

func (s *containerCmds) runCbmClUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.cbmClusterName(args[0])
	if err != nil {
		return err
	}
	body := &gkeonprem.BareMetalCluster{}
	if err := loadYAMLOrJSONInto(s.flagCbmConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagCbmUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEOnPrem(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.BareMetalClusters.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating bare-metal cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update bare-metal cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCbmFormat)
}
