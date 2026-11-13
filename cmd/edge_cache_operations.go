package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud edge-cache operations (#1058) ---

func (s *edgeCacheCmds) initEdgeCacheOperations() {
	all := []*cobra.Command{s.edgeCacheOperationsDescribeCmd, s.edgeCacheOperationsListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagEdgeCacheOperationsLocation, "location", "", "Edge Cache location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagEdgeCacheOperationsFormat, "format", "", "Output format")
	}
	s.edgeCacheOperationsListCmd.Flags().Int64Var(&s.flagEdgeCacheOperationsPageSize, "page-size", 0, "Maximum results per page")

	s.edgeCacheOperationsCmd.AddCommand(all...)
	s.edgeCacheCmd.AddCommand(s.edgeCacheOperationsCmd)
}

func (s *edgeCacheCmds) edgeCacheOperationsParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagEdgeCacheOperationsLocation), nil
}

func (s *edgeCacheCmds) edgeCacheOperationsName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.edgeCacheOperationsParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/operations/%s", parent, id), nil
}

func (s *edgeCacheCmds) runEdgeCacheOperationsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheOperationsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(edgeCacheRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing edge-cache operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagEdgeCacheOperationsFormat)
}

func (s *edgeCacheCmds) runEdgeCacheOperationsList(cmd *cobra.Command, args []string) error {
	parent, err := s.edgeCacheOperationsParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(edgeCacheRest()).paginate(ctx, "/"+parent+"/operations", nil, "operations", s.flagEdgeCacheOperationsPageSize)
	if err != nil {
		return fmt.Errorf("listing edge-cache operations: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagEdgeCacheOperationsFormat)
}
