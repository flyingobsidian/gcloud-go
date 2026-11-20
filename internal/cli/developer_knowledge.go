package cli

import (
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gfilter"
	"github.com/spf13/cobra"
	developerknowledge "google.golang.org/api/developerknowledge/v1"
)

// --- gcloud developer-knowledge (#1954) ---
//
// gcloud-python promoted developer-knowledge to GA in 586.0.0 (beta in
// 584.0.0). Both tracks use the developerknowledge v1 API.

// dkSearchChunksDefaultLimit is gcloud-python's default --limit for
// documents search-chunks.
const dkSearchChunksDefaultLimit = 5

// dkView maps a documents describe --view choice to its DocumentView value.
func dkView(choice string) (string, bool) {
	switch choice {
	case "basic", "content", "full":
		return "DOCUMENT_VIEW_" + strings.ToUpper(choice), true
	}
	return "", false
}

func (s *developerKnowledgeCmds) initDeveloperKnowledge() {
	s.dkAnswerQueryCmd.Flags().StringVar(&s.flagDKQuery, "query", "", "The query to answer (required)")
	_ = s.dkAnswerQueryCmd.MarkFlagRequired("query")
	s.dkAnswerQueryCmd.Flags().StringVar(&s.flagDKQueryFilter, "query-filter", "",
		"Apply a strict filter to the search results used to ground the answer. Supported filter fields: "+
			"`content_length_bytes`, `data_source`, `update_time`, `uri`")

	s.dkDocDescribeCmd.Flags().StringVar(&s.flagDKView, "view", "content",
		"Fields of the document included in the response: basic, content, full")

	f := s.dkDocSearchChunksCmd.Flags()
	f.StringVar(&s.flagDKQuery, "query", "",
		`Raw query string provided by the user, such as "How to create a Cloud Storage bucket?" (required)`)
	_ = s.dkDocSearchChunksCmd.MarkFlagRequired("query")
	f.StringVar(&s.flagDKQueryFilter, "query-filter", "",
		"Apply a strict filter to the search results. Supported filter fields: `data_source`, `update_time`, `uri`")
	f.IntVar(&s.flagDKLimit, "limit", dkSearchChunksDefaultLimit, "Maximum number of resources to list")
	f.Int64Var(&s.flagDKPageSize, "page-size", 0, "Maximum results per page")
	f.StringSliceVar(&s.flagDKSortBy, "sort-by", nil, "Comma-separated list of resource field keys to sort by")

	s.dkDocumentsCmd.AddCommand(s.dkDocDescribeCmd, s.dkDocSearchChunksCmd)
	s.developerKnowledgeCmd.AddCommand(s.dkAnswerQueryCmd, s.dkDocumentsCmd)
}

func (s *developerKnowledgeCmds) runDKAnswerQuery(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperKnowledge(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.V1.AnswerQuery(&developerknowledge.AnswerQueryRequest{
		Query:  s.flagDKQuery,
		Filter: s.flagDKQueryFilter,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("answering query: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.inv.flags.format)
}

func (s *developerKnowledgeCmds) runDKDocDescribe(cmd *cobra.Command, args []string) error {
	view, ok := dkView(s.flagDKView)
	if !ok {
		return fmt.Errorf("argument --view: Invalid choice: '%s'.\n\nValid choices are [basic, content, full].", s.flagDKView)
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperKnowledge(ctx)
	if err != nil {
		return err
	}
	doc, err := svc.Documents.Get(args[0]).View(view).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing document: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, doc, s.inv.flags.format)
}

func (s *developerKnowledgeCmds) runDKDocSearchChunks(cmd *cobra.Command, args []string) error {
	filter, err := gfilter.Compile(s.inv.flags.filter)
	if err != nil {
		return fmt.Errorf("invalid --filter expression: %w", err)
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperKnowledge(ctx)
	if err != nil {
		return err
	}
	// --limit applies after --filter and --sort-by, so with neither the
	// pager can stop as soon as it has --limit results, as gcloud-python's
	// does.
	stopEarly := s.inv.flags.filter == "" && len(s.flagDKSortBy) == 0
	var records []map[string]any
	pageToken := ""
	for {
		call := svc.Documents.SearchDocumentChunks().Query(s.flagDKQuery).Context(ctx)
		if s.flagDKQueryFilter != "" {
			call = call.Filter(s.flagDKQueryFilter)
		}
		if s.flagDKPageSize > 0 {
			call = call.PageSize(s.flagDKPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("searching document chunks: %w", err)
		}
		rows, err := jsonRows(resp.Results)
		if err != nil {
			return err
		}
		for _, r := range rows {
			if filter.Match(r) {
				records = append(records, r)
			}
		}
		if resp.NextPageToken == "" || (stopEarly && s.flagDKLimit > 0 && len(records) >= s.flagDKLimit) {
			break
		}
		pageToken = resp.NextPageToken
	}
	sortRecordsBy(records, s.flagDKSortBy)
	if s.flagDKLimit > 0 && len(records) > s.flagDKLimit {
		records = records[:s.flagDKLimit]
	}
	return emitFormattedTo(s.inv.stdout, records, s.inv.flags.format)
}

// developerKnowledgeCmds holds the commands and flag values of one
// invocation's "developer-knowledge" command group, so invocations don't share
// state (#2070).
type developerKnowledgeCmds struct {
	inv *invocation

	developerKnowledgeCmd *cobra.Command
	dkAnswerQueryCmd      *cobra.Command
	dkDocumentsCmd        *cobra.Command
	dkDocDescribeCmd      *cobra.Command
	dkDocSearchChunksCmd  *cobra.Command

	flagDKQuery       string
	flagDKQueryFilter string
	flagDKView        string
	flagDKLimit       int
	flagDKPageSize    int64
	flagDKSortBy      []string
}

// newDeveloperKnowledgeCmd returns the "developer-knowledge" command group for
// inv.
func newDeveloperKnowledgeCmd(inv *invocation) *cobra.Command {
	return newDeveloperKnowledgeCmds(inv).developerKnowledgeCmd
}

// newDeveloperKnowledgeCmds builds the "developer-knowledge" command group for
// inv and returns its state.
func newDeveloperKnowledgeCmds(inv *invocation) *developerKnowledgeCmds {
	s := &developerKnowledgeCmds{inv: inv}
	s.developerKnowledgeCmd = &cobra.Command{
		Use: "developer-knowledge", Short: "Manage Developer Knowledge resources and query knowledge base",
	}
	s.dkAnswerQueryCmd = &cobra.Command{
		Use: "answer-query", Short: "Answer a natural language query using grounded generation",
		Args: cobra.NoArgs, RunE: s.runDKAnswerQuery,
	}
	s.dkDocumentsCmd = &cobra.Command{Use: "documents", Short: "Explore and search Developer Knowledge documents"}
	s.dkDocDescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Retrieve a single document with its full Markdown content",
		Args: cobra.ExactArgs(1), RunE: s.runDKDocDescribe,
	}
	s.dkDocSearchChunksCmd = &cobra.Command{
		Use:   "search-chunks",
		Short: "Search for developer knowledge across Google's developer documentation",
		Args:  cobra.NoArgs, RunE: s.runDKDocSearchChunks,
	}
	s.initDeveloperKnowledge()
	return s
}
