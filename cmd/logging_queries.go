package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging saved-queries (#920) ---
// --- gcloud logging recent-queries (#918) ---
// --- gcloud logging scopes (#921) ---

// saved-queries

var loggingSavedQueriesCmd = &cobra.Command{Use: "saved-queries", Short: "Manage saved queries"}

var (
	flagLogSavedQueryDescription string
	flagLogSavedQueryDisplayName string
	flagLogSavedQueryVisibility  string
	flagLogSavedQueryQuery       string
)

var (
	loggingSavedQueriesCreateCmd = &cobra.Command{
		Use: "create SAVED_QUERY_ID", Short: "Create a saved query",
		Args: cobra.ExactArgs(1), RunE: runLogSavedQueryCreate,
	}
	loggingSavedQueriesDeleteCmd = &cobra.Command{
		Use: "delete SAVED_QUERY_ID", Short: "Delete a saved query",
		Args: cobra.ExactArgs(1), RunE: runLogSavedQueryDelete,
	}
	loggingSavedQueriesDescribeCmd = &cobra.Command{
		Use: "describe SAVED_QUERY_ID", Short: "Describe a saved query",
		Args: cobra.ExactArgs(1), RunE: runLogSavedQueryDescribe,
	}
	loggingSavedQueriesListCmd = &cobra.Command{
		Use: "list", Short: "List saved queries",
		Args: cobra.NoArgs, RunE: runLogSavedQueryList,
	}
	loggingSavedQueriesUpdateCmd = &cobra.Command{
		Use: "update SAVED_QUERY_ID", Short: "Update a saved query",
		Args: cobra.ExactArgs(1), RunE: runLogSavedQueryUpdate,
	}
)

func savedQueryFromFlags(body *logging.SavedQuery) {
	if flagLogSavedQueryDescription != "" {
		body.Description = flagLogSavedQueryDescription
	}
	if flagLogSavedQueryDisplayName != "" {
		body.DisplayName = flagLogSavedQueryDisplayName
	}
	if flagLogSavedQueryVisibility != "" {
		body.Visibility = flagLogSavedQueryVisibility
	}
	if flagLogSavedQueryQuery != "" {
		if body.LoggingQuery == nil {
			body.LoggingQuery = &logging.LoggingQuery{}
		}
		body.LoggingQuery.Filter = flagLogSavedQueryQuery
	}
}

func runLogSavedQueryCreate(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	loc := loggingLocationParent(parent, loggingLocation())
	body := &logging.SavedQuery{}
	if flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
			return err
		}
	}
	savedQueryFromFlags(body)
	if body.DisplayName == "" {
		body.DisplayName = args[0]
	}
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
	return emitFormatted(got, flagLogFormat)
}

func runLogSavedQueryDelete(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	name := loggingLocationChildName(parent, loggingLocation(), "savedQueries", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
	fmt.Printf("Deleted saved query [%s].\n", args[0])
	return nil
}

func runLogSavedQueryDescribe(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	name := loggingLocationChildName(parent, loggingLocation(), "savedQueries", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
	return emitFormatted(got, flagLogFormat)
}

func runLogSavedQueryList(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	loc := loggingLocationParent(parent, loggingLocation())
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
	if flagLogFormat != "" {
		return emitFormatted(all, flagLogFormat)
	}
	fmt.Printf("%-40s %-15s %s\n", "NAME", "VISIBILITY", "DISPLAY_NAME")
	for _, q := range all {
		fmt.Printf("%-40s %-15s %s\n", loggingBasename(q.Name), q.Visibility, q.DisplayName)
	}
	return nil
}

func runLogSavedQueryUpdate(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	name := loggingLocationChildName(parent, loggingLocation(), "savedQueries", args[0])
	body := &logging.SavedQuery{}
	if flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
			return err
		}
	}
	savedQueryFromFlags(body)
	mask := loggingResolveMask(body)
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
	return emitFormatted(got, flagLogFormat)
}

// recent-queries

var loggingRecentQueriesCmd = &cobra.Command{Use: "recent-queries", Short: "Manage recent queries"}

var loggingRecentQueriesListCmd = &cobra.Command{
	Use: "list", Short: "List recent queries",
	Args: cobra.NoArgs, RunE: runLogRecentQueryList,
}

func runLogRecentQueryList(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	loc := loggingLocationParent(parent, loggingLocation())
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
	if flagLogFormat != "" {
		return emitFormatted(all, flagLogFormat)
	}
	for _, q := range all {
		fmt.Println(loggingBasename(q.Name))
	}
	return nil
}

// scopes

var loggingScopesCmd = &cobra.Command{Use: "scopes", Short: "Manage log scopes"}

var (
	flagLogScopeDescription   string
	flagLogScopeResourceNames []string
)

var (
	loggingScopesCreateCmd = &cobra.Command{
		Use: "create LOG_SCOPE_ID", Short: "Create a log scope",
		Args: cobra.ExactArgs(1), RunE: runLogScopeCreate,
	}
	loggingScopesDeleteCmd = &cobra.Command{
		Use: "delete LOG_SCOPE_ID", Short: "Delete a log scope",
		Args: cobra.ExactArgs(1), RunE: runLogScopeDelete,
	}
	loggingScopesDescribeCmd = &cobra.Command{
		Use: "describe LOG_SCOPE_ID", Short: "Describe a log scope",
		Args: cobra.ExactArgs(1), RunE: runLogScopeDescribe,
	}
	loggingScopesListCmd = &cobra.Command{
		Use: "list", Short: "List log scopes",
		Args: cobra.NoArgs, RunE: runLogScopeList,
	}
	loggingScopesUpdateCmd = &cobra.Command{
		Use: "update LOG_SCOPE_ID", Short: "Update a log scope",
		Args: cobra.ExactArgs(1), RunE: runLogScopeUpdate,
	}
)

func scopeFromFlags(body *logging.LogScope) {
	if flagLogScopeDescription != "" {
		body.Description = flagLogScopeDescription
	}
	if len(flagLogScopeResourceNames) > 0 {
		body.ResourceNames = append(body.ResourceNames, flagLogScopeResourceNames...)
	}
}

func scopesParentSupported(scope string) error {
	if scope == "projects" || scope == "folders" || scope == "organizations" {
		return nil
	}
	return fmt.Errorf("log scopes are only supported at project, folder, or organization scope")
}

func runLogScopeCreate(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	scope := loggingScope(parent)
	if err := scopesParentSupported(scope); err != nil {
		return err
	}
	loc := loggingLocationParent(parent, loggingLocation())
	body := &logging.LogScope{}
	if flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
			return err
		}
	}
	scopeFromFlags(body)
	if len(body.ResourceNames) == 0 {
		return fmt.Errorf("--resource-name is required (or provide `resourceNames` via --config-file)")
	}
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
	return emitFormatted(got, flagLogFormat)
}

func runLogScopeDelete(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	scope := loggingScope(parent)
	if err := scopesParentSupported(scope); err != nil {
		return err
	}
	name := loggingLocationChildName(parent, loggingLocation(), "logScopes", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
	fmt.Printf("Deleted log scope [%s].\n", args[0])
	return nil
}

func runLogScopeDescribe(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	scope := loggingScope(parent)
	if err := scopesParentSupported(scope); err != nil {
		return err
	}
	name := loggingLocationChildName(parent, loggingLocation(), "logScopes", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
	return emitFormatted(got, flagLogFormat)
}

func runLogScopeList(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	scope := loggingScope(parent)
	if err := scopesParentSupported(scope); err != nil {
		return err
	}
	loc := loggingLocationParent(parent, loggingLocation())
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
	if flagLogFormat != "" {
		return emitFormatted(all, flagLogFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DESCRIPTION")
	for _, s := range all {
		fmt.Printf("%-40s %s\n", loggingBasename(s.Name), s.Description)
	}
	return nil
}

func runLogScopeUpdate(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	scope := loggingScope(parent)
	if err := scopesParentSupported(scope); err != nil {
		return err
	}
	name := loggingLocationChildName(parent, loggingLocation(), "logScopes", args[0])
	body := &logging.LogScope{}
	if flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
			return err
		}
	}
	scopeFromFlags(body)
	mask := loggingResolveMask(body)
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
	return emitFormatted(got, flagLogFormat)
}

func init() {
	// saved-queries
	sq := []*cobra.Command{loggingSavedQueriesCreateCmd, loggingSavedQueriesDeleteCmd,
		loggingSavedQueriesDescribeCmd, loggingSavedQueriesListCmd, loggingSavedQueriesUpdateCmd}
	addLogScopeFlags(sq...)
	addLogLocationFlag(sq...)
	addLogFormatFlag(loggingSavedQueriesCreateCmd, loggingSavedQueriesDescribeCmd,
		loggingSavedQueriesListCmd, loggingSavedQueriesUpdateCmd)
	addLogPageSizeFlag(loggingSavedQueriesListCmd)
	for _, c := range []*cobra.Command{loggingSavedQueriesCreateCmd, loggingSavedQueriesUpdateCmd} {
		c.Flags().StringVar(&flagLogSavedQueryDescription, "description", "", "A textual description")
		c.Flags().StringVar(&flagLogSavedQueryDisplayName, "display-name", "", "Display name of the saved query")
		c.Flags().StringVar(&flagLogSavedQueryVisibility, "visibility", "PRIVATE", "Visibility (PRIVATE|SHARED)")
		c.Flags().StringVar(&flagLogSavedQueryQuery, "query", "", "Query filter body")
		c.Flags().StringVar(&flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the SavedQuery body")
	}
	loggingSavedQueriesUpdateCmd.Flags().StringVar(&flagLogUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	loggingSavedQueriesCmd.AddCommand(sq...)
	loggingCmd.AddCommand(loggingSavedQueriesCmd)

	// recent-queries
	addLogScopeFlags(loggingRecentQueriesListCmd)
	addLogLocationFlag(loggingRecentQueriesListCmd)
	addLogFormatFlag(loggingRecentQueriesListCmd)
	addLogPageSizeFlag(loggingRecentQueriesListCmd)
	loggingRecentQueriesCmd.AddCommand(loggingRecentQueriesListCmd)
	loggingCmd.AddCommand(loggingRecentQueriesCmd)

	// scopes
	sc := []*cobra.Command{loggingScopesCreateCmd, loggingScopesDeleteCmd, loggingScopesDescribeCmd,
		loggingScopesListCmd, loggingScopesUpdateCmd}
	addLogScopeFlags(sc...)
	addLogLocationFlag(sc...)
	addLogFormatFlag(loggingScopesCreateCmd, loggingScopesDescribeCmd, loggingScopesListCmd, loggingScopesUpdateCmd)
	addLogPageSizeFlag(loggingScopesListCmd)
	for _, c := range []*cobra.Command{loggingScopesCreateCmd, loggingScopesUpdateCmd} {
		c.Flags().StringVar(&flagLogScopeDescription, "description", "", "A textual description for the scope")
		c.Flags().StringSliceVar(&flagLogScopeResourceNames, "resource-name", nil, "Resource name to include (repeatable)")
		c.Flags().StringVar(&flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the LogScope body")
	}
	loggingScopesUpdateCmd.Flags().StringVar(&flagLogUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	loggingScopesCmd.AddCommand(sc...)
	loggingCmd.AddCommand(loggingScopesCmd)
}
