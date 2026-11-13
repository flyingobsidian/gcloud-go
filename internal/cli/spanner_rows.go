package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud spanner rows (#1212) ---

func (s *spannerCmds) initSpannerRows() {
	all := []*cobra.Command{s.spannerRowsInsertCmd, s.spannerRowsUpdateCmd, s.spannerRowsDeleteCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagSpRowsInstance, "instance", "", "Spanner instance (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&s.flagSpRowsDatabase, "database", "", "Spanner database (required)")
		_ = c.MarkFlagRequired("database")
		c.Flags().StringVar(&s.flagSpRowsTable, "table", "", "Table to modify (required)")
		_ = c.MarkFlagRequired("table")
		c.Flags().StringVar(&s.flagSpRowsFormat, "format", "", "Output format")
	}
	s.spannerRowsInsertCmd.Flags().StringToStringVar(&s.flagSpRowsData, "data", nil,
		"Column=value pairs to insert (repeatable). Values are inserted as strings; cast in SQL as needed.")
	_ = s.spannerRowsInsertCmd.MarkFlagRequired("data")

	s.spannerRowsUpdateCmd.Flags().StringToStringVar(&s.flagSpRowsData, "data", nil,
		"Column=value pairs to set (repeatable)")
	_ = s.spannerRowsUpdateCmd.MarkFlagRequired("data")
	s.spannerRowsUpdateCmd.Flags().StringSliceVar(&s.flagSpRowsKeys, "keys", nil,
		"Primary-key column=value expressions used in the WHERE clause (repeatable)")
	_ = s.spannerRowsUpdateCmd.MarkFlagRequired("keys")

	s.spannerRowsDeleteCmd.Flags().StringSliceVar(&s.flagSpRowsKeys, "keys", nil,
		"Primary-key column=value expressions used in the WHERE clause (repeatable)")
	_ = s.spannerRowsDeleteCmd.MarkFlagRequired("keys")

	s.spannerRowsCmd.AddCommand(all...)
	s.spannerCmd.AddCommand(s.spannerRowsCmd)
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

func (s *spannerCmds) runSpRowsInsert(cmd *cobra.Command, args []string) error {
	database, err := s.spannerDatabase(s.flagSpRowsInstance, s.flagSpRowsDatabase)
	if err != nil {
		return err
	}
	if len(s.flagSpRowsData) == 0 {
		return fmt.Errorf("--data is required")
	}
	cols := make([]string, 0, len(s.flagSpRowsData))
	vals := make([]string, 0, len(s.flagSpRowsData))
	for k, v := range s.flagSpRowsData {
		cols = append(cols, k)
		vals = append(vals, quoteLit(v))
	}
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		s.flagSpRowsTable, strings.Join(cols, ","), strings.Join(vals, ","))
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	rs, err := spDbRunSingleUseSQL(ctx, svc, database, sql, true)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, rs, s.flagSpRowsFormat)
}

func (s *spannerCmds) runSpRowsUpdate(cmd *cobra.Command, args []string) error {
	database, err := s.spannerDatabase(s.flagSpRowsInstance, s.flagSpRowsDatabase)
	if err != nil {
		return err
	}
	if len(s.flagSpRowsData) == 0 {
		return fmt.Errorf("--data is required")
	}
	if len(s.flagSpRowsKeys) == 0 {
		return fmt.Errorf("--keys is required")
	}
	setClauses := make([]string, 0, len(s.flagSpRowsData))
	for k, v := range s.flagSpRowsData {
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", k, quoteLit(v)))
	}
	whereClauses, err := keysWhere(s.flagSpRowsKeys)
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		s.flagSpRowsTable, strings.Join(setClauses, ", "), strings.Join(whereClauses, " AND "))
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	rs, err := spDbRunSingleUseSQL(ctx, svc, database, sql, true)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, rs, s.flagSpRowsFormat)
}

func (s *spannerCmds) runSpRowsDelete(cmd *cobra.Command, args []string) error {
	database, err := s.spannerDatabase(s.flagSpRowsInstance, s.flagSpRowsDatabase)
	if err != nil {
		return err
	}
	if len(s.flagSpRowsKeys) == 0 {
		return fmt.Errorf("--keys is required")
	}
	whereClauses, err := keysWhere(s.flagSpRowsKeys)
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("DELETE FROM %s WHERE %s",
		s.flagSpRowsTable, strings.Join(whereClauses, " AND "))
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	rs, err := spDbRunSingleUseSQL(ctx, svc, database, sql, true)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, rs, s.flagSpRowsFormat)
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
