package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud memorystore operations (#980) ---

func (s *memorystoreCmds) initMemorystoreOperations() {
	all := []*cobra.Command{s.memstoreOpCancelCmd, s.memstoreOpDeleteCmd, s.memstoreOpDescribeCmd, s.memstoreOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagMemstoreOpLocation, "location", "", "Memorystore location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagMemstoreOpFormat, "format", "", "Output format")
	}
	s.memstoreOpListCmd.Flags().Int64Var(&s.flagMemstoreOpPageSize, "page-size", 0, "Maximum results per page")

	s.memstoreOpCmd.AddCommand(all...)
	s.memorystoreCmd.AddCommand(s.memstoreOpCmd)
}

func (s *memorystoreCmds) memstoreOpParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagMemstoreOpLocation), nil
}

func (s *memorystoreCmds) memstoreOpName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.memstoreOpParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/operations/%s", parent, id), nil
}

func (s *memorystoreCmds) runMemstoreOpCancel(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodPost, "/"+name+":cancel", nil, map[string]any{}, nil); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancel request issued for operation [%s].\n", args[0])
	return nil
}

func (s *memorystoreCmds) runMemstoreOpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodDelete, "/"+name, nil, nil, nil); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (s *memorystoreCmds) runMemstoreOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMemstoreOpFormat)
}

func (s *memorystoreCmds) runMemstoreOpList(cmd *cobra.Command, args []string) error {
	parent, err := s.memstoreOpParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(memorystoreRest()).paginate(ctx, "/"+parent+"/operations", nil, "operations", s.flagMemstoreOpPageSize)
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagMemstoreOpFormat)
}
