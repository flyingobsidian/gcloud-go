package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud artifacts operations (#1076) ---

func (s *artifactsCmds) initArtifactsOperations() {
	s.artifactsOperationsDescribeCmd.Flags().StringVar(&s.flagArtOpsLocation, "location", "", "Location of the operation")
	s.artifactsOperationsDescribeCmd.Flags().StringVar(&s.flagArtOpsFormat, "format", "", "Output format")

	s.artifactsOperationsCmd.AddCommand(s.artifactsOperationsDescribeCmd)
	s.artifactsCmd.AddCommand(s.artifactsOperationsCmd)
}

func (s *artifactsCmds) runArtifactsOperationsDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]
	if s.flagArtOpsLocation == "" && !isFullOperationName(name) {
		return fmt.Errorf("--location is required (or pass a full projects/.../operations/... name)")
	}
	if !isFullOperationName(name) {
		project, err := s.inv.project()
		if err != nil {
			return err
		}
		name = fmt.Sprintf("%s/operations/%s", artLocationParent(project, s.flagArtOpsLocation), name)
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagArtOpsFormat)
}

func isFullOperationName(s string) bool {
	return len(s) > 9 && s[:9] == "projects/"
}
