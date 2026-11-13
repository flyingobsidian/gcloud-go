package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub operations (#1163) ---

func (g *apihubCmds) initApihubOperations() {
	all := []*cobra.Command{g.apihubOpCancelCmd, g.apihubOpDeleteCmd, g.apihubOpDescribeCmd, g.apihubOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagApihubOpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagApihubOpFormat, "format", "", "Output format")
	}
	g.apihubOpListCmd.Flags().StringVar(&g.flagApihubOpFilter, "filter", "", "Server-side filter expression")
	g.apihubOpListCmd.Flags().Int64Var(&g.flagApihubOpPageSize, "page-size", 0, "Maximum results per page")

	g.apihubOperationsCmd.AddCommand(all...)
	g.apihubCmd.AddCommand(g.apihubOperationsCmd)
}

func (g *apihubCmds) apihubOpName(id string) (string, error) {
	return g.apihubResource(g.flagApihubOpLocation, "operations", id)
}

func (g *apihubCmds) runApihubOpCancel(cmd *cobra.Command, args []string) error {
	name, err := g.apihubOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &apihub.GoogleLongrunningCancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (g *apihubCmds) runApihubOpDelete(cmd *cobra.Command, args []string) error {
	name, err := g.apihubOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (g *apihubCmds) runApihubOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.apihubOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubOpFormat)
}

func (g *apihubCmds) runApihubOpList(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubOpLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleLongrunningOperation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if g.flagApihubOpFilter != "" {
			call = call.Filter(g.flagApihubOpFilter)
		}
		if g.flagApihubOpPageSize > 0 {
			call = call.PageSize(g.flagApihubOpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagApihubOpFormat)
}
