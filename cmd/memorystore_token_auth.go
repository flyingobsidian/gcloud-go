package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
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

var (
	memstoreTAUCmd = &cobra.Command{
		Use: "token-auth-users", Short: "Manage token auth users of a Memorystore instance",
	}
	memstoreAuthTokensCmd = &cobra.Command{
		Use: "auth-tokens", Short: "Manage auth tokens of a Memorystore token auth user",
	}

	memstoreCreateTAUCmd = &cobra.Command{
		Use: "create-token-auth-user INSTANCE", Short: "Create token auth user for a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: runMemstoreCreateTAU,
	}
	memstoreTAUCreateTokenCmd = &cobra.Command{
		Use: "create-auth-token TOKEN_AUTH_USER", Short: "Create an auth token for a token auth user",
		Args: cobra.ExactArgs(1), RunE: runMemstoreTAUCreateToken,
	}
	memstoreTAUDeleteCmd = &cobra.Command{
		Use: "delete TOKEN_AUTH_USER", Short: "Delete a token auth user",
		Args: cobra.ExactArgs(1), RunE: runMemstoreTAUDelete,
	}
	memstoreTAUDescribeCmd = &cobra.Command{
		Use: "describe TOKEN_AUTH_USER", Short: "Describe a token auth user",
		Args: cobra.ExactArgs(1), RunE: runMemstoreTAUDescribe,
	}
	memstoreTAUListCmd = &cobra.Command{
		Use: "list", Short: "List token auth users of an instance",
		Args: cobra.NoArgs, RunE: runMemstoreTAUList,
	}
	memstoreATDeleteCmd = &cobra.Command{
		Use: "delete AUTH_TOKEN", Short: "Delete an auth token",
		Args: cobra.ExactArgs(1), RunE: runMemstoreATDelete,
	}
	memstoreATDescribeCmd = &cobra.Command{
		Use: "describe AUTH_TOKEN", Short: "Describe an auth token",
		Args: cobra.ExactArgs(1), RunE: runMemstoreATDescribe,
	}
	memstoreATListCmd = &cobra.Command{
		Use: "list", Short: "List auth tokens of a token auth user",
		Args: cobra.NoArgs, RunE: runMemstoreATList,
	}
)

var (
	flagMemstoreTAULocation string
	flagMemstoreTAUInstance string
	flagMemstoreTAUUser     string
	flagMemstoreTAUFormat   string
	flagMemstoreTAUAsync    bool
	flagMemstoreTAURequest  string
	flagMemstoreTAUForce    bool
	flagMemstoreTAUPageSize int64
)

func init() {
	all := []*cobra.Command{
		memstoreCreateTAUCmd, memstoreTAUCreateTokenCmd, memstoreTAUDeleteCmd, memstoreTAUDescribeCmd,
		memstoreTAUListCmd, memstoreATDeleteCmd, memstoreATDescribeCmd, memstoreATListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagMemstoreTAULocation, "location", "", "Memorystore location")
		c.Flags().StringVar(&flagMemstoreTAUFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		memstoreTAUCreateTokenCmd, memstoreTAUDeleteCmd, memstoreTAUDescribeCmd, memstoreTAUListCmd,
		memstoreATDeleteCmd, memstoreATDescribeCmd, memstoreATListCmd,
	} {
		c.Flags().StringVar(&flagMemstoreTAUInstance, "instance", "", "The instance id")
	}
	for _, c := range []*cobra.Command{memstoreATDeleteCmd, memstoreATDescribeCmd, memstoreATListCmd} {
		c.Flags().StringVar(&flagMemstoreTAUUser, "token-auth-user", "", "The token auth user id")
	}
	for _, c := range []*cobra.Command{memstoreCreateTAUCmd, memstoreTAUCreateTokenCmd, memstoreTAUDeleteCmd, memstoreATDeleteCmd} {
		c.Flags().BoolVar(&flagMemstoreTAUAsync, "async", false,
			"Return immediately, without waiting for the operation in progress to complete")
	}
	memstoreCreateTAUCmd.Flags().StringVar(&flagMemstoreTAUUser, "token-auth-user", "",
		"The name of the token auth user to create (required)")
	_ = memstoreCreateTAUCmd.MarkFlagRequired("token-auth-user")
	memstoreTAUDeleteCmd.Flags().StringVar(&flagMemstoreTAURequest, "request-id", "",
		"An optional request ID (UUID) to identify the request")
	memstoreTAUDeleteCmd.Flags().BoolVar(&flagMemstoreTAUForce, "force", false,
		"If set to true, any auth tokens from this user will also be deleted")
	for _, c := range []*cobra.Command{memstoreTAUListCmd, memstoreATListCmd} {
		c.Flags().Int64Var(&flagMemstoreTAUPageSize, "page-size", 0, "Maximum results per page")
	}

	memstoreAuthTokensCmd.AddCommand(memstoreATDeleteCmd, memstoreATDescribeCmd, memstoreATListCmd)
	memstoreTAUCmd.AddCommand(memstoreTAUCreateTokenCmd, memstoreTAUDeleteCmd, memstoreTAUDescribeCmd,
		memstoreTAUListCmd, memstoreAuthTokensCmd)
	memstoreInstCmd.AddCommand(memstoreCreateTAUCmd, memstoreTAUCmd)
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

func memstoreTAUInstanceName(arg string) (string, string, error) {
	project, err := resolveProjectIfNeeded(arg)
	if err != nil {
		return "", "", err
	}
	return memstoreTAUResource(project, flagMemstoreTAULocation, arg, []string{"instances"}, nil, nil)
}

func memstoreTAUUserName(arg string) (string, string, error) {
	project, err := resolveProjectIfNeeded(arg)
	if err != nil {
		return "", "", err
	}
	return memstoreTAUResource(project, flagMemstoreTAULocation, arg,
		[]string{"instances", "tokenAuthUsers"}, []string{"instance"}, []string{flagMemstoreTAUInstance})
}

func memstoreTAUTokenName(arg string) (string, string, error) {
	project, err := resolveProjectIfNeeded(arg)
	if err != nil {
		return "", "", err
	}
	return memstoreTAUResource(project, flagMemstoreTAULocation, arg,
		[]string{"instances", "tokenAuthUsers", "authTokens"},
		[]string{"instance", "token-auth-user"}, []string{flagMemstoreTAUInstance, flagMemstoreTAUUser})
}

// resolveProjectIfNeeded resolves the project unless arg is already a full
// resource name.
func resolveProjectIfNeeded(arg string) (string, error) {
	if strings.HasPrefix(arg, "projects/") {
		return "", nil
	}
	return resolveProject()
}

// memstoreTAUWait prints gcloud's declarative LRO messages and waits for op
// unless --async. It returns the finished operation, or nil with --async.
func memstoreTAUWait(ctx context.Context, op map[string]any, requestMsg string) (map[string]any, error) {
	fmt.Fprintln(os.Stderr, requestMsg)
	opName, _ := op["name"].(string)
	if flagMemstoreTAUAsync {
		fmt.Fprintf(os.Stderr, "Check operation [%s] for status.\n", opName)
		return nil, nil
	}
	fmt.Fprintf(os.Stderr, "Waiting for operation [%s] to complete...", opName)
	for done, _ := op["done"].(bool); !done; done, _ = op["done"].(bool) {
		time.Sleep(time.Second)
		op = map[string]any{}
		if err := memorystoreRest.do(ctx, http.MethodGet, "/"+opName, nil, nil, &op); err != nil {
			fmt.Fprintln(os.Stderr, "failed.")
			return nil, fmt.Errorf("polling operation %s: %w", opName, err)
		}
	}
	if e, ok := op["error"].(map[string]any); ok {
		fmt.Fprintln(os.Stderr, "failed.")
		msg, _ := e["message"].(string)
		return nil, fmt.Errorf("%s", msg)
	}
	fmt.Fprintln(os.Stderr, "done.")
	return op, nil
}

// memstoreTAUGenericResult finishes a generic declarative command: with
// --async the operation is printed, otherwise its response, if any.
func memstoreTAUGenericResult(ctx context.Context, op map[string]any, id string) error {
	final, err := memstoreTAUWait(ctx, op, fmt.Sprintf("Request issued for: [%s]", id))
	if err != nil {
		return err
	}
	if final == nil {
		return emitFormatted(op, flagMemstoreTAUFormat)
	}
	if resp, ok := final["response"].(map[string]any); ok && len(resp) > 0 {
		return emitFormatted(resp, flagMemstoreTAUFormat)
	}
	return nil
}

// memstoreTAUDeleteResult finishes a declarative delete: silent on stdout
// unless --format, with "Deleted KIND [ID]." once the operation is done.
func memstoreTAUDeleteResult(ctx context.Context, op map[string]any, kind, id string) error {
	final, err := memstoreTAUWait(ctx, op, fmt.Sprintf("Delete request issued for: [%s]", id))
	if err != nil {
		return err
	}
	if final != nil {
		fmt.Fprintf(os.Stderr, "Deleted %s [%s].\n", kind, id)
	}
	if flagMemstoreTAUFormat != "" {
		if final == nil {
			final = op
		}
		return emitFormatted(final, flagMemstoreTAUFormat)
	}
	return nil
}

func runMemstoreCreateTAU(cmd *cobra.Command, args []string) error {
	name, id, err := memstoreTAUInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	body := map[string]any{"tokenAuthUser": flagMemstoreTAUUser}
	if err := memorystoreRest.do(ctx, http.MethodPost, "/"+name+":addTokenAuthUser", nil, body, &op); err != nil {
		return fmt.Errorf("creating token auth user: %w", err)
	}
	return memstoreTAUGenericResult(ctx, op, id)
}

func runMemstoreTAUCreateToken(cmd *cobra.Command, args []string) error {
	name, id, err := memstoreTAUUserName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodPost, "/"+name+":addAuthToken", nil, map[string]any{}, &op); err != nil {
		return fmt.Errorf("creating auth token: %w", err)
	}
	return memstoreTAUGenericResult(ctx, op, id)
}

func runMemstoreTAUDelete(cmd *cobra.Command, args []string) error {
	name, id, err := memstoreTAUUserName(args[0])
	if err != nil {
		return err
	}
	q := url.Values{}
	if flagMemstoreTAURequest != "" {
		q.Set("requestId", flagMemstoreTAURequest)
	}
	if flagMemstoreTAUForce {
		q.Set("force", "true")
	}
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodDelete, "/"+name, q, nil, &op); err != nil {
		return fmt.Errorf("deleting token auth user: %w", err)
	}
	return memstoreTAUDeleteResult(ctx, op, "tokenAuthUser", id)
}

func runMemstoreTAUDescribe(cmd *cobra.Command, args []string) error {
	name, _, err := memstoreTAUUserName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := memorystoreRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing token auth user: %w", err)
	}
	return emitFormatted(got, flagMemstoreTAUFormat)
}

func runMemstoreTAUList(cmd *cobra.Command, args []string) error {
	if flagMemstoreTAUInstance == "" {
		return fmt.Errorf("--instance is required")
	}
	parent, _, err := memstoreTAUInstanceName(flagMemstoreTAUInstance)
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := memorystoreRest.paginate(ctx, "/"+parent+"/tokenAuthUsers", nil, "tokenAuthUsers", flagMemstoreTAUPageSize)
	if err != nil {
		return fmt.Errorf("listing token auth users: %w", err)
	}
	return memstoreTAUEmitList(items)
}

func runMemstoreATDelete(cmd *cobra.Command, args []string) error {
	name, id, err := memstoreTAUTokenName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting auth token: %w", err)
	}
	return memstoreTAUDeleteResult(ctx, op, "authToken", id)
}

func runMemstoreATDescribe(cmd *cobra.Command, args []string) error {
	name, _, err := memstoreTAUTokenName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := memorystoreRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing auth token: %w", err)
	}
	return emitFormatted(got, flagMemstoreTAUFormat)
}

func runMemstoreATList(cmd *cobra.Command, args []string) error {
	if flagMemstoreTAUUser == "" {
		return fmt.Errorf("--token-auth-user is required")
	}
	parent, _, err := memstoreTAUUserName(flagMemstoreTAUUser)
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := memorystoreRest.paginate(ctx, "/"+parent+"/authTokens", nil, "authTokens", flagMemstoreTAUPageSize)
	if err != nil {
		return fmt.Errorf("listing auth tokens: %w", err)
	}
	return memstoreTAUEmitList(items)
}

func memstoreTAUEmitList(items []map[string]any) error {
	if items == nil {
		items = []map[string]any{}
	}
	if err := emitFormatted(items, flagMemstoreTAUFormat); err != nil {
		return err
	}
	if flagMemstoreTAUFormat == "" && len(items) == 0 {
		fmt.Fprintln(os.Stderr, "Listed 0 items.")
	}
	return nil
}
