package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
)

// --- gcloud spanner rows (#1212) ---

var spannerRowsCmd = &cobra.Command{Use: "rows", Short: "Manage Cloud Spanner rows via DML"}

var (
	flagSpRowsInstance string
	flagSpRowsDatabase string
	flagSpRowsTable    string
	flagSpRowsFormat   string
	flagSpRowsKeys     []string
	flagSpRowsData     map[string]string
)

var (
	spannerRowsInsertCmd = &cobra.Command{
		Use: "insert", Short: "Insert a row via DML",
		Args: cobra.NoArgs, RunE: runSpRowsInsert,
	}
	spannerRowsUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update rows via DML",
		Args: cobra.NoArgs, RunE: runSpRowsUpdate,
	}
	spannerRowsDeleteCmd = &cobra.Command{
		Use: "delete", Short: "Delete rows via DML",
		Args: cobra.NoArgs, RunE: runSpRowsDelete,
	}
)

func init() {
	all := []*cobra.Command{spannerRowsInsertCmd, spannerRowsUpdateCmd, spannerRowsDeleteCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagSpRowsInstance, "instance", "", "Spanner instance (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&flagSpRowsDatabase, "database", "", "Spanner database (required)")
		_ = c.MarkFlagRequired("database")
		c.Flags().StringVar(&flagSpRowsTable, "table", "", "Table to modify (required)")
		_ = c.MarkFlagRequired("table")
		c.Flags().StringVar(&flagSpRowsFormat, "format", "", "Output format")
	}
	spannerRowsInsertCmd.Flags().StringToStringVar(&flagSpRowsData, "data", nil,
		"Column=value pairs to insert (repeatable). Values are inserted as strings; cast in SQL as needed.")
	_ = spannerRowsInsertCmd.MarkFlagRequired("data")

	spannerRowsUpdateCmd.Flags().StringToStringVar(&flagSpRowsData, "data", nil,
		"Column=value pairs to set (repeatable)")
	_ = spannerRowsUpdateCmd.MarkFlagRequired("data")
	spannerRowsUpdateCmd.Flags().StringSliceVar(&flagSpRowsKeys, "keys", nil,
		"Primary-key column=value expressions used in the WHERE clause (repeatable)")
	_ = spannerRowsUpdateCmd.MarkFlagRequired("keys")

	spannerRowsDeleteCmd.Flags().StringSliceVar(&flagSpRowsKeys, "keys", nil,
		"Primary-key column=value expressions used in the WHERE clause (repeatable)")
	_ = spannerRowsDeleteCmd.MarkFlagRequired("keys")

	spannerRowsCmd.AddCommand(all...)
	spannerCmd.AddCommand(spannerRowsCmd)
}

// splitKV splits a "k=v" pair. Trims spaces.
func splitKV(s string) (string, string, error) {
	i := strings.IndexByte(s, '=')
	if i < 0 {
		return "", "", fmt.Errorf("expected KEY=VALUE, got %q", s)
	}
	return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:]), nil
}

// quoteLit renders a literal for embedding directly in a DML statement. Strings
// are single-quoted and internal single quotes are escaped. Numeric-looking
// values pass through unquoted.
func quoteLit(v string) string {
	// If it parses as a number, or is TRUE/FALSE/NULL, pass through.
	upper := strings.ToUpper(v)
	if upper == "NULL" || upper == "TRUE" || upper == "FALSE" {
		return upper
	}
	// Cheap numeric sniff.
	if isNumeric(v) {
		return v
	}
	return "'" + strings.ReplaceAll(v, "'", "''") + "'"
}

func isNumeric(v string) bool {
	if v == "" {
		return false
	}
	seenDot := false
	for i, r := range v {
		if r == '-' && i == 0 {
			continue
		}
		if r == '.' {
			if seenDot {
				return false
			}
			seenDot = true
			continue
		}
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func runSpRowsInsert(cmd *cobra.Command, args []string) error {
	database, err := spannerDatabase(flagSpRowsInstance, flagSpRowsDatabase)
	if err != nil {
		return err
	}
	if len(flagSpRowsData) == 0 {
		return fmt.Errorf("--data is required")
	}
	cols := make([]string, 0, len(flagSpRowsData))
	vals := make([]string, 0, len(flagSpRowsData))
	for k, v := range flagSpRowsData {
		cols = append(cols, k)
		vals = append(vals, quoteLit(v))
	}
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		flagSpRowsTable, strings.Join(cols, ","), strings.Join(vals, ","))
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	rs, err := spDbRunSingleUseSQL(ctx, svc, database, sql, true)
	if err != nil {
		return err
	}
	return emitFormatted(rs, flagSpRowsFormat)
}

func runSpRowsUpdate(cmd *cobra.Command, args []string) error {
	database, err := spannerDatabase(flagSpRowsInstance, flagSpRowsDatabase)
	if err != nil {
		return err
	}
	if len(flagSpRowsData) == 0 {
		return fmt.Errorf("--data is required")
	}
	if len(flagSpRowsKeys) == 0 {
		return fmt.Errorf("--keys is required")
	}
	setClauses := make([]string, 0, len(flagSpRowsData))
	for k, v := range flagSpRowsData {
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", k, quoteLit(v)))
	}
	whereClauses, err := keysWhere(flagSpRowsKeys)
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		flagSpRowsTable, strings.Join(setClauses, ", "), strings.Join(whereClauses, " AND "))
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	rs, err := spDbRunSingleUseSQL(ctx, svc, database, sql, true)
	if err != nil {
		return err
	}
	return emitFormatted(rs, flagSpRowsFormat)
}

func runSpRowsDelete(cmd *cobra.Command, args []string) error {
	database, err := spannerDatabase(flagSpRowsInstance, flagSpRowsDatabase)
	if err != nil {
		return err
	}
	if len(flagSpRowsKeys) == 0 {
		return fmt.Errorf("--keys is required")
	}
	whereClauses, err := keysWhere(flagSpRowsKeys)
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("DELETE FROM %s WHERE %s",
		flagSpRowsTable, strings.Join(whereClauses, " AND "))
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	rs, err := spDbRunSingleUseSQL(ctx, svc, database, sql, true)
	if err != nil {
		return err
	}
	return emitFormatted(rs, flagSpRowsFormat)
}

func keysWhere(keys []string) ([]string, error) {
	out := make([]string, 0, len(keys))
	for _, kv := range keys {
		k, v, err := splitKV(kv)
		if err != nil {
			return nil, err
		}
		out = append(out, fmt.Sprintf("%s = %s", k, quoteLit(v)))
	}
	return out, nil
}
