package cli

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud memorystore locations (#979) ---

func (s *memorystoreCmds) initMemorystoreLocations() {
	all := []*cobra.Command{s.memstoreLocDescribeCmd, s.memstoreLocListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagMemstoreLocFormat, "format", "", "Output format")
	}
	s.memstoreLocListCmd.Flags().Int64Var(&s.flagMemstoreLocPageSize, "page-size", 0, "Maximum results per page")

	s.memstoreLocCmd.AddCommand(all...)
	s.memorystoreCmd.AddCommand(s.memstoreLocCmd)
}

func (s *memorystoreCmds) memstoreLocationName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, id), nil
}

func (s *memorystoreCmds) runMemstoreLocDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreLocationName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMemstoreLocFormat)
}

func (s *memorystoreCmds) runMemstoreLocList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(memorystoreRest()).paginate(ctx, fmt.Sprintf("/projects/%s/locations", project), nil, "locations", s.flagMemstoreLocPageSize)
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagMemstoreLocFormat)
}
