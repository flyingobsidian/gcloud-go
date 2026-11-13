package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	developerconnect "google.golang.org/api/developerconnect/v1"
)

// --- gcloud developer-connect connections (#1023) ---

func (s *developerConnectCmds) initDeveloperConnectConnections() {
	all := []*cobra.Command{
		s.developerConnectConnCreateCmd, s.developerConnectConnDeleteCmd,
		s.developerConnectConnDescribeCmd, s.developerConnectConnListCmd,
		s.developerConnectConnUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDcConnLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDcConnFormat, "format", "", "Output format")
	}
	s.developerConnectConnCreateCmd.Flags().StringVar(&s.flagDcConnConfigFile, "config-file", "", "YAML/JSON file with the Connection body (required)")
	_ = s.developerConnectConnCreateCmd.MarkFlagRequired("config-file")
	s.developerConnectConnListCmd.Flags().Int64Var(&s.flagDcConnPageSize, "page-size", 0, "Maximum results per page")
	s.developerConnectConnUpdateCmd.Flags().StringVar(&s.flagDcConnConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.developerConnectConnUpdateCmd.MarkFlagRequired("config-file")
	s.developerConnectConnUpdateCmd.Flags().StringVar(&s.flagDcConnUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.developerConnectConnectionsCmd.AddCommand(all...)
	s.developerConnectCmd.AddCommand(s.developerConnectConnectionsCmd)
}

func (s *developerConnectCmds) dcConnName(id string) (string, error) {
	return s.devConnResourceName(s.flagDcConnLocation, "connections", id)
}

func (s *developerConnectCmds) runDcConnCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.devConnLocationParent(s.flagDcConnLocation)
	if err != nil {
		return err
	}
	body := &developerconnect.Connection{}
	if err := loadYAMLOrJSONInto(s.flagDcConnConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Connections.Create(parent, body).ConnectionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating connection: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDcConnFormat)
}

func (s *developerConnectCmds) runDcConnDelete(cmd *cobra.Command, args []string) error {
	name, err := s.dcConnName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Connections.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting connection: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDcConnFormat)
}

func (s *developerConnectCmds) runDcConnDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.dcConnName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Connections.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing connection: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDcConnFormat)
}

func (s *developerConnectCmds) runDcConnList(cmd *cobra.Command, args []string) error {
	parent, err := s.devConnLocationParent(s.flagDcConnLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
	if err != nil {
		return err
	}
	var all []*developerconnect.Connection
	pageToken := ""
	for {
		call := svc.Projects.Locations.Connections.List(parent).Context(ctx)
		if s.flagDcConnPageSize > 0 {
			call = call.PageSize(s.flagDcConnPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing connections: %w", err)
		}
		all = append(all, resp.Connections...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDcConnFormat)
}

func (s *developerConnectCmds) runDcConnUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.dcConnName(args[0])
	if err != nil {
		return err
	}
	body := &developerconnect.Connection{}
	if err := loadYAMLOrJSONInto(s.flagDcConnConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagDcConnUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Connections.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating connection: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDcConnFormat)
}
