package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging saved-queries (#920) ---
// --- gcloud logging recent-queries (#918) ---
// --- gcloud logging scopes (#921) ---

// saved-queries

func (g *loggingCmds) savedQueryFromFlags(body *logging.SavedQuery) {
	if g.flagLogSavedQueryDescription != "" {
		body.Description = g.flagLogSavedQueryDescription
	}
	if g.flagLogSavedQueryDisplayName != "" {
		body.DisplayName = g.flagLogSavedQueryDisplayName
	}
	if g.flagLogSavedQueryVisibility != "" {
		body.Visibility = g.flagLogSavedQueryVisibility
	}
	if g.flagLogSavedQueryQuery != "" {
		if body.LoggingQuery == nil {
			body.LoggingQuery = &logging.LoggingQuery{}
		}
		body.LoggingQuery.Filter = g.flagLogSavedQueryQuery
	}
}

func (g *loggingCmds) runLogSavedQueryCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	loc := loggingLocationParent(parent, g.loggingLocation())
	body := &logging.SavedQuery{}
	if g.flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
			return err
		}
	}
	g.savedQueryFromFlags(body)
	if body.DisplayName == "" {
		body.DisplayName = args[0]
	}
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.SavedQuery
	switch loggingScope(parent) {
	case "projects":
		got, err = svc.Projects.Locations.SavedQueries.Create(loc, body).SavedQueryId(args[0]).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.SavedQueries.Create(loc, body).SavedQueryId(args[0]).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.SavedQueries.Create(loc, body).SavedQueryId(args[0]).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.Locations.SavedQueries.Create(loc, body).SavedQueryId(args[0]).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("creating saved query: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogSavedQueryDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	name := loggingLocationChildName(parent, g.loggingLocation(), "savedQueries", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	switch loggingScope(parent) {
	case "projects":
		_, err = svc.Projects.Locations.SavedQueries.Delete(name).Context(ctx).Do()
	case "folders":
		_, err = svc.Folders.Locations.SavedQueries.Delete(name).Context(ctx).Do()
	case "organizations":
		_, err = svc.Organizations.Locations.SavedQueries.Delete(name).Context(ctx).Do()
	case "billingAccounts":
		_, err = svc.BillingAccounts.Locations.SavedQueries.Delete(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("deleting saved query: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted saved query [%s].\n", args[0])
	return nil
}

func (g *loggingCmds) runLogSavedQueryDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	name := loggingLocationChildName(parent, g.loggingLocation(), "savedQueries", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.SavedQuery
	switch loggingScope(parent) {
	case "projects":
		got, err = svc.Projects.Locations.SavedQueries.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.SavedQueries.Get(name).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.SavedQueries.Get(name).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.Locations.SavedQueries.Get(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("describing saved query: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogSavedQueryList(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	loc := loggingLocationParent(parent, g.loggingLocation())
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var all []*logging.SavedQuery
	pageToken := ""
	for {
		var (
			page []*logging.SavedQuery
			next string
		)
		switch loggingScope(parent) {
		case "projects":
			call := svc.Projects.Locations.SavedQueries.List(loc).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing saved queries: %w", err)
			}
			page, next = resp.SavedQueries, resp.NextPageToken
		case "folders":
			call := svc.Folders.Locations.SavedQueries.List(loc).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing saved queries: %w", err)
			}
			page, next = resp.SavedQueries, resp.NextPageToken
		case "organizations":
			call := svc.Organizations.Locations.SavedQueries.List(loc).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing saved queries: %w", err)
			}
			page, next = resp.SavedQueries, resp.NextPageToken
		case "billingAccounts":
			call := svc.BillingAccounts.Locations.SavedQueries.List(loc).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing saved queries: %w", err)
			}
			page, next = resp.SavedQueries, resp.NextPageToken
		default:
			return fmt.Errorf("invalid parent %q", parent)
		}
		all = append(all, page...)
		if next == "" {
			break
		}
		pageToken = next
	}
	if g.flagLogFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagLogFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %-15s %s\n", "NAME", "VISIBILITY", "DISPLAY_NAME")
	for _, q := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %-15s %s\n", loggingBasename(q.Name), q.Visibility, q.DisplayName)
	}
	return nil
}

func (g *loggingCmds) runLogSavedQueryUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	name := loggingLocationChildName(parent, g.loggingLocation(), "savedQueries", args[0])
	body := &logging.SavedQuery{}
	if g.flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
			return err
		}
	}
	g.savedQueryFromFlags(body)
	mask := g.loggingResolveMask(body)
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.SavedQuery
	switch loggingScope(parent) {
	case "projects":
		got, err = svc.Projects.Locations.SavedQueries.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.SavedQueries.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.SavedQueries.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.Locations.SavedQueries.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("updating saved query: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

// recent-queries

func (g *loggingCmds) runLogRecentQueryList(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	loc := loggingLocationParent(parent, g.loggingLocation())
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var all []*logging.RecentQuery
	pageToken := ""
	for {
		var (
			page []*logging.RecentQuery
			next string
		)
		switch loggingScope(parent) {
		case "projects":
			call := svc.Projects.Locations.RecentQueries.List(loc).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing recent queries: %w", err)
			}
			page, next = resp.RecentQueries, resp.NextPageToken
		case "folders":
			call := svc.Folders.Locations.RecentQueries.List(loc).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing recent queries: %w", err)
			}
			page, next = resp.RecentQueries, resp.NextPageToken
		case "organizations":
			call := svc.Organizations.Locations.RecentQueries.List(loc).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing recent queries: %w", err)
			}
			page, next = resp.RecentQueries, resp.NextPageToken
		case "billingAccounts":
			call := svc.BillingAccounts.Locations.RecentQueries.List(loc).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing recent queries: %w", err)
			}
			page, next = resp.RecentQueries, resp.NextPageToken
		default:
			return fmt.Errorf("invalid parent %q", parent)
		}
		all = append(all, page...)
		if next == "" {
			break
		}
		pageToken = next
	}
	if g.flagLogFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagLogFormat)
	}
	for _, q := range all {
		fmt.Fprintln(g.inv.stdout, loggingBasename(q.Name))
	}
	return nil
}

// scopes

func (g *loggingCmds) scopeFromFlags(body *logging.LogScope) {
	if g.flagLogScopeDescription != "" {
		body.Description = g.flagLogScopeDescription
	}
	if len(g.flagLogScopeResourceNames) > 0 {
		body.ResourceNames = append(body.ResourceNames, g.flagLogScopeResourceNames...)
	}
}

func scopesParentSupported(scope string) error {
	if scope == "projects" || scope == "folders" || scope == "organizations" {
		return nil
	}
	return fmt.Errorf("log scopes are only supported at project, folder, or organization scope")
}

func (g *loggingCmds) runLogScopeCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	scope := loggingScope(parent)
	if err := scopesParentSupported(scope); err != nil {
		return err
	}
	loc := loggingLocationParent(parent, g.loggingLocation())
	body := &logging.LogScope{}
	if g.flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
			return err
		}
	}
	g.scopeFromFlags(body)
	if len(body.ResourceNames) == 0 {
		return fmt.Errorf("--resource-name is required (or provide `resourceNames` via --config-file)")
	}
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.LogScope
	switch scope {
	case "projects":
		got, err = svc.Projects.Locations.LogScopes.Create(loc, body).LogScopeId(args[0]).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.LogScopes.Create(loc, body).LogScopeId(args[0]).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.LogScopes.Create(loc, body).LogScopeId(args[0]).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("creating log scope: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogScopeDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	scope := loggingScope(parent)
	if err := scopesParentSupported(scope); err != nil {
		return err
	}
	name := loggingLocationChildName(parent, g.loggingLocation(), "logScopes", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	switch scope {
	case "projects":
		_, err = svc.Projects.Locations.LogScopes.Delete(name).Context(ctx).Do()
	case "folders":
		_, err = svc.Folders.Locations.LogScopes.Delete(name).Context(ctx).Do()
	case "organizations":
		_, err = svc.Organizations.Locations.LogScopes.Delete(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("deleting log scope: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted log scope [%s].\n", args[0])
	return nil
}

func (g *loggingCmds) runLogScopeDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	scope := loggingScope(parent)
	if err := scopesParentSupported(scope); err != nil {
		return err
	}
	name := loggingLocationChildName(parent, g.loggingLocation(), "logScopes", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.LogScope
	switch scope {
	case "projects":
		got, err = svc.Projects.Locations.LogScopes.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.LogScopes.Get(name).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.LogScopes.Get(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("describing log scope: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogScopeList(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	scope := loggingScope(parent)
	if err := scopesParentSupported(scope); err != nil {
		return err
	}
	loc := loggingLocationParent(parent, g.loggingLocation())
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var all []*logging.LogScope
	pageToken := ""
	for {
		var (
			page []*logging.LogScope
			next string
		)
		switch scope {
		case "projects":
			call := svc.Projects.Locations.LogScopes.List(loc).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log scopes: %w", err)
			}
			page, next = resp.LogScopes, resp.NextPageToken
		case "folders":
			call := svc.Folders.Locations.LogScopes.List(loc).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log scopes: %w", err)
			}
			page, next = resp.LogScopes, resp.NextPageToken
		case "organizations":
			call := svc.Organizations.Locations.LogScopes.List(loc).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log scopes: %w", err)
			}
			page, next = resp.LogScopes, resp.NextPageToken
		}
		all = append(all, page...)
		if next == "" {
			break
		}
		pageToken = next
	}
	if g.flagLogFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagLogFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "DESCRIPTION")
	for _, s := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", loggingBasename(s.Name), s.Description)
	}
	return nil
}

func (g *loggingCmds) runLogScopeUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	scope := loggingScope(parent)
	if err := scopesParentSupported(scope); err != nil {
		return err
	}
	name := loggingLocationChildName(parent, g.loggingLocation(), "logScopes", args[0])
	body := &logging.LogScope{}
	if g.flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
			return err
		}
	}
	g.scopeFromFlags(body)
	mask := g.loggingResolveMask(body)
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.LogScope
	switch scope {
	case "projects":
		got, err = svc.Projects.Locations.LogScopes.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.LogScopes.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.LogScopes.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("updating log scope: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) initLoggingQueries() {
	// saved-queries
	sq := []*cobra.Command{g.loggingSavedQueriesCreateCmd, g.loggingSavedQueriesDeleteCmd,
		g.loggingSavedQueriesDescribeCmd, g.loggingSavedQueriesListCmd, g.loggingSavedQueriesUpdateCmd}
	g.addLogScopeFlags(sq...)
	g.addLogLocationFlag(sq...)
	g.addLogFormatFlag(g.loggingSavedQueriesCreateCmd, g.loggingSavedQueriesDescribeCmd,
		g.loggingSavedQueriesListCmd, g.loggingSavedQueriesUpdateCmd)
	g.addLogPageSizeFlag(g.loggingSavedQueriesListCmd)
	for _, c := range []*cobra.Command{g.loggingSavedQueriesCreateCmd, g.loggingSavedQueriesUpdateCmd} {
		c.Flags().StringVar(&g.flagLogSavedQueryDescription, "description", "", "A textual description")
		c.Flags().StringVar(&g.flagLogSavedQueryDisplayName, "display-name", "", "Display name of the saved query")
		c.Flags().StringVar(&g.flagLogSavedQueryVisibility, "visibility", "PRIVATE", "Visibility (PRIVATE|SHARED)")
		c.Flags().StringVar(&g.flagLogSavedQueryQuery, "query", "", "Query filter body")
		c.Flags().StringVar(&g.flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the SavedQuery body")
	}
	g.loggingSavedQueriesUpdateCmd.Flags().StringVar(&g.flagLogUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	g.loggingSavedQueriesCmd.AddCommand(sq...)
	g.loggingCmd.AddCommand(g.loggingSavedQueriesCmd)

	// recent-queries
	g.addLogScopeFlags(g.loggingRecentQueriesListCmd)
	g.addLogLocationFlag(g.loggingRecentQueriesListCmd)
	g.addLogFormatFlag(g.loggingRecentQueriesListCmd)
	g.addLogPageSizeFlag(g.loggingRecentQueriesListCmd)
	g.loggingRecentQueriesCmd.AddCommand(g.loggingRecentQueriesListCmd)
	g.loggingCmd.AddCommand(g.loggingRecentQueriesCmd)

	// scopes
	sc := []*cobra.Command{g.loggingScopesCreateCmd, g.loggingScopesDeleteCmd, g.loggingScopesDescribeCmd,
		g.loggingScopesListCmd, g.loggingScopesUpdateCmd}
	g.addLogScopeFlags(sc...)
	g.addLogLocationFlag(sc...)
	g.addLogFormatFlag(g.loggingScopesCreateCmd, g.loggingScopesDescribeCmd, g.loggingScopesListCmd, g.loggingScopesUpdateCmd)
	g.addLogPageSizeFlag(g.loggingScopesListCmd)
	for _, c := range []*cobra.Command{g.loggingScopesCreateCmd, g.loggingScopesUpdateCmd} {
		c.Flags().StringVar(&g.flagLogScopeDescription, "description", "", "A textual description for the scope")
		c.Flags().StringSliceVar(&g.flagLogScopeResourceNames, "resource-name", nil, "Resource name to include (repeatable)")
		c.Flags().StringVar(&g.flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the LogScope body")
	}
	g.loggingScopesUpdateCmd.Flags().StringVar(&g.flagLogUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	g.loggingScopesCmd.AddCommand(sc...)
	g.loggingCmd.AddCommand(g.loggingScopesCmd)
}
