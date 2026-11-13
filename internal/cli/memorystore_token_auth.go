package cli

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud memorystore instances token auth (#1935) ---
//
// `instances create-token-auth-user`, `instances token-auth-users` and
// `instances token-auth-users auth-tokens` were promoted to GA in
// gcloud-python 585.0.0. They are declarative commands; mutations return an
// operation that is waited on unless --async.

func (s *memorystoreCmds) initMemorystoreTokenAuth() {
	all := []*cobra.Command{
		s.memstoreCreateTAUCmd, s.memstoreTAUCreateTokenCmd, s.memstoreTAUDeleteCmd, s.memstoreTAUDescribeCmd,
		s.memstoreTAUListCmd, s.memstoreATDeleteCmd, s.memstoreATDescribeCmd, s.memstoreATListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagMemstoreTAULocation, "location", "", "Memorystore location")
		c.Flags().StringVar(&s.flagMemstoreTAUFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		s.memstoreTAUCreateTokenCmd, s.memstoreTAUDeleteCmd, s.memstoreTAUDescribeCmd, s.memstoreTAUListCmd,
		s.memstoreATDeleteCmd, s.memstoreATDescribeCmd, s.memstoreATListCmd,
	} {
		c.Flags().StringVar(&s.flagMemstoreTAUInstance, "instance", "", "The instance id")
	}
	for _, c := range []*cobra.Command{s.memstoreATDeleteCmd, s.memstoreATDescribeCmd, s.memstoreATListCmd} {
		c.Flags().StringVar(&s.flagMemstoreTAUUser, "token-auth-user", "", "The token auth user id")
	}
	for _, c := range []*cobra.Command{s.memstoreCreateTAUCmd, s.memstoreTAUCreateTokenCmd, s.memstoreTAUDeleteCmd, s.memstoreATDeleteCmd} {
		c.Flags().BoolVar(&s.flagMemstoreTAUAsync, "async", false,
			"Return immediately, without waiting for the operation in progress to complete")
	}
	s.memstoreCreateTAUCmd.Flags().StringVar(&s.flagMemstoreTAUUser, "token-auth-user", "",
		"The name of the token auth user to create (required)")
	_ = s.memstoreCreateTAUCmd.MarkFlagRequired("token-auth-user")
	s.memstoreTAUDeleteCmd.Flags().StringVar(&s.flagMemstoreTAURequest, "request-id", "",
		"An optional request ID (UUID) to identify the request")
	s.memstoreTAUDeleteCmd.Flags().BoolVar(&s.flagMemstoreTAUForce, "force", false,
		"If set to true, any auth tokens from this user will also be deleted")
	for _, c := range []*cobra.Command{s.memstoreTAUListCmd, s.memstoreATListCmd} {
		c.Flags().Int64Var(&s.flagMemstoreTAUPageSize, "page-size", 0, "Maximum results per page")
	}

	s.memstoreAuthTokensCmd.AddCommand(s.memstoreATDeleteCmd, s.memstoreATDescribeCmd, s.memstoreATListCmd)
	s.memstoreTAUCmd.AddCommand(s.memstoreTAUCreateTokenCmd, s.memstoreTAUDeleteCmd, s.memstoreTAUDescribeCmd,
		s.memstoreTAUListCmd, s.memstoreAuthTokensCmd)
	s.memstoreInstCmd.AddCommand(s.memstoreCreateTAUCmd, s.memstoreTAUCmd)
}

// memstoreTAUResource resolves a positional resource argument (an ID with
// its parent flags, or a full projects/... name) to its relative name and
// short ID. collections lists the path segments below the location, e.g.
// ["instances", "tokenAuthUsers"]; parentIDs holds the flag values for all
// but the last segment.
func memstoreTAUResource(project, location, arg string, collections, parentFlags, parentIDs []string) (string, string, error) {
	if strings.HasPrefix(arg, "projects/") {
		parts := strings.Split(arg, "/")
		if len(parts) != 4+2*len(collections) || parts[2] != "locations" {
			return "", "", fmt.Errorf("invalid resource name [%s]", arg)
		}
		for i, c := range collections {
			if parts[4+2*i] != c {
				return "", "", fmt.Errorf("invalid resource name [%s]", arg)
			}
		}
		return arg, parts[len(parts)-1], nil
	}
	if location == "" {
		return "", "", fmt.Errorf("--location is required (or pass a fully qualified resource name)")
	}
	name := fmt.Sprintf("projects/%s/locations/%s", project, location)
	for i, c := range collections[:len(collections)-1] {
		if parentIDs[i] == "" {
			return "", "", fmt.Errorf("--%s is required (or pass a fully qualified resource name)", parentFlags[i])
		}
		name += "/" + c + "/" + parentIDs[i]
	}
	return name + "/" + collections[len(collections)-1] + "/" + arg, arg, nil
}

func (s *memorystoreCmds) memstoreTAUInstanceName(arg string) (string, string, error) {
	project, err := s.resolveProjectIfNeeded(arg)
	if err != nil {
		return "", "", err
	}
	return memstoreTAUResource(project, s.flagMemstoreTAULocation, arg, []string{"instances"}, nil, nil)
}

func (s *memorystoreCmds) memstoreTAUUserName(arg string) (string, string, error) {
	project, err := s.resolveProjectIfNeeded(arg)
	if err != nil {
		return "", "", err
	}
	return memstoreTAUResource(project, s.flagMemstoreTAULocation, arg,
		[]string{"instances", "tokenAuthUsers"}, []string{"instance"}, []string{s.flagMemstoreTAUInstance})
}

func (s *memorystoreCmds) memstoreTAUTokenName(arg string) (string, string, error) {
	project, err := s.resolveProjectIfNeeded(arg)
	if err != nil {
		return "", "", err
	}
	return memstoreTAUResource(project, s.flagMemstoreTAULocation, arg,
		[]string{"instances", "tokenAuthUsers", "authTokens"},
		[]string{"instance", "token-auth-user"}, []string{s.flagMemstoreTAUInstance, s.flagMemstoreTAUUser})
}

// resolveProjectIfNeeded resolves the project unless arg is already a full
// resource name.
func (s *memorystoreCmds) resolveProjectIfNeeded(arg string) (string, error) {
	if strings.HasPrefix(arg, "projects/") {
		return "", nil
	}
	return s.inv.project()
}

// memstoreTAUWait prints gcloud's declarative LRO messages and waits for op
// unless --async. It returns the finished operation, or nil with --async.
func (s *memorystoreCmds) memstoreTAUWait(ctx context.Context, op map[string]any, requestMsg string) (map[string]any, error) {
	fmt.Fprintln(s.inv.stderr, requestMsg)
	opName, _ := op["name"].(string)
	if s.flagMemstoreTAUAsync {
		fmt.Fprintf(s.inv.stderr, "Check operation [%s] for status.\n", opName)
		return nil, nil
	}
	fmt.Fprintf(s.inv.stderr, "Waiting for operation [%s] to complete...", opName)
	for done, _ := op["done"].(bool); !done; done, _ = op["done"].(bool) {
		time.Sleep(time.Second)
		op = map[string]any{}
		if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodGet, "/"+opName, nil, nil, &op); err != nil {
			fmt.Fprintln(s.inv.stderr, "failed.")
			return nil, fmt.Errorf("polling operation %s: %w", opName, err)
		}
	}
	if e, ok := op["error"].(map[string]any); ok {
		fmt.Fprintln(s.inv.stderr, "failed.")
		msg, _ := e["message"].(string)
		return nil, fmt.Errorf("%s", msg)
	}
	fmt.Fprintln(s.inv.stderr, "done.")
	return op, nil
}

// memstoreTAUGenericResult finishes a generic declarative command: with
// --async the operation is printed, otherwise its response, if any.
func (s *memorystoreCmds) memstoreTAUGenericResult(ctx context.Context, op map[string]any, id string) error {
	final, err := s.memstoreTAUWait(ctx, op, fmt.Sprintf("Request issued for: [%s]", id))
	if err != nil {
		return err
	}
	if final == nil {
		return emitFormattedTo(s.inv.stdout, op, s.flagMemstoreTAUFormat)
	}
	if resp, ok := final["response"].(map[string]any); ok && len(resp) > 0 {
		return emitFormattedTo(s.inv.stdout, resp, s.flagMemstoreTAUFormat)
	}
	return nil
}

// memstoreTAUDeleteResult finishes a declarative delete: silent on stdout
// unless --format, with "Deleted KIND [ID]." once the operation is done.
func (s *memorystoreCmds) memstoreTAUDeleteResult(ctx context.Context, op map[string]any, kind, id string) error {
	final, err := s.memstoreTAUWait(ctx, op, fmt.Sprintf("Delete request issued for: [%s]", id))
	if err != nil {
		return err
	}
	if final != nil {
		fmt.Fprintf(s.inv.stderr, "Deleted %s [%s].\n", kind, id)
	}
	if s.flagMemstoreTAUFormat != "" {
		if final == nil {
			final = op
		}
		return emitFormattedTo(s.inv.stdout, final, s.flagMemstoreTAUFormat)
	}
	return nil
}

func (s *memorystoreCmds) runMemstoreCreateTAU(cmd *cobra.Command, args []string) error {
	name, id, err := s.memstoreTAUInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	body := map[string]any{"tokenAuthUser": s.flagMemstoreTAUUser}
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodPost, "/"+name+":addTokenAuthUser", nil, body, &op); err != nil {
		return fmt.Errorf("creating token auth user: %w", err)
	}
	return s.memstoreTAUGenericResult(ctx, op, id)
}

func (s *memorystoreCmds) runMemstoreTAUCreateToken(cmd *cobra.Command, args []string) error {
	name, id, err := s.memstoreTAUUserName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodPost, "/"+name+":addAuthToken", nil, map[string]any{}, &op); err != nil {
		return fmt.Errorf("creating auth token: %w", err)
	}
	return s.memstoreTAUGenericResult(ctx, op, id)
}

func (s *memorystoreCmds) runMemstoreTAUDelete(cmd *cobra.Command, args []string) error {
	name, id, err := s.memstoreTAUUserName(args[0])
	if err != nil {
		return err
	}
	q := url.Values{}
	if s.flagMemstoreTAURequest != "" {
		q.Set("requestId", s.flagMemstoreTAURequest)
	}
	if s.flagMemstoreTAUForce {
		q.Set("force", "true")
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodDelete, "/"+name, q, nil, &op); err != nil {
		return fmt.Errorf("deleting token auth user: %w", err)
	}
	return s.memstoreTAUDeleteResult(ctx, op, "tokenAuthUser", id)
}

func (s *memorystoreCmds) runMemstoreTAUDescribe(cmd *cobra.Command, args []string) error {
	name, _, err := s.memstoreTAUUserName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing token auth user: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMemstoreTAUFormat)
}

func (s *memorystoreCmds) runMemstoreTAUList(cmd *cobra.Command, args []string) error {
	if s.flagMemstoreTAUInstance == "" {
		return fmt.Errorf("--instance is required")
	}
	parent, _, err := s.memstoreTAUInstanceName(s.flagMemstoreTAUInstance)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(memorystoreRest()).paginate(ctx, "/"+parent+"/tokenAuthUsers", nil, "tokenAuthUsers", s.flagMemstoreTAUPageSize)
	if err != nil {
		return fmt.Errorf("listing token auth users: %w", err)
	}
	return s.memstoreTAUEmitList(items)
}

func (s *memorystoreCmds) runMemstoreATDelete(cmd *cobra.Command, args []string) error {
	name, id, err := s.memstoreTAUTokenName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting auth token: %w", err)
	}
	return s.memstoreTAUDeleteResult(ctx, op, "authToken", id)
}

func (s *memorystoreCmds) runMemstoreATDescribe(cmd *cobra.Command, args []string) error {
	name, _, err := s.memstoreTAUTokenName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing auth token: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMemstoreTAUFormat)
}

func (s *memorystoreCmds) runMemstoreATList(cmd *cobra.Command, args []string) error {
	if s.flagMemstoreTAUUser == "" {
		return fmt.Errorf("--token-auth-user is required")
	}
	parent, _, err := s.memstoreTAUUserName(s.flagMemstoreTAUUser)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(memorystoreRest()).paginate(ctx, "/"+parent+"/authTokens", nil, "authTokens", s.flagMemstoreTAUPageSize)
	if err != nil {
		return fmt.Errorf("listing auth tokens: %w", err)
	}
	return s.memstoreTAUEmitList(items)
}

func (s *memorystoreCmds) memstoreTAUEmitList(items []map[string]any) error {
	if items == nil {
		items = []map[string]any{}
	}
	if err := emitFormattedTo(s.inv.stdout, items, s.flagMemstoreTAUFormat); err != nil {
		return err
	}
	if s.flagMemstoreTAUFormat == "" && len(items) == 0 {
		fmt.Fprintln(s.inv.stderr, "Listed 0 items.")
	}
	return nil
}
