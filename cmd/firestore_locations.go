package cmd

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

func (g *firestoreCmds) initFirestoreLocations() {
	g.fsLocListCmd.Flags().StringVar(&g.flagFSLocFormat, "format", "", "Output format")
	g.fsLocListCmd.Flags().Int64Var(&g.flagFSLocPageSize, "page-size", 0, "Page size")
	g.fsLocListCmd.Flags().Int64Var(&g.flagFSLocLimit, "limit", 0, "Cap total results (0 = no cap)")
	g.fsLocDescribeCmd.Flags().StringVar(&g.flagFSLocFormat, "format", "", "Output format")

	g.firestoreLocationsCmd.AddCommand(g.fsLocDescribeCmd, g.fsLocListCmd)
	g.firestoreCmd.AddCommand(g.firestoreLocationsCmd)
}

func (g *firestoreCmds) runFSLocList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	var all []*firestore.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if g.flagFSLocPageSize > 0 {
			call = call.PageSize(g.flagFSLocPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if g.flagFSLocLimit > 0 && int64(len(all)) >= g.flagFSLocLimit {
			all = all[:g.flagFSLocLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if g.flagFSLocFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagFSLocFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range all {
		fmt.Fprintf(g.inv.stdout, "%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

func (g *firestoreCmds) runFSLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(firestoreLocationName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	if g.flagFSLocFormat != "" {
		return emitFormattedTo(g.inv.stdout, loc, g.flagFSLocFormat)
	}
	fmt.Fprintf(g.inv.stdout, "%s (%s)\n", path.Base(loc.Name), loc.DisplayName)
	return nil
}
