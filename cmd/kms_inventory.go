package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	kmsinventory "google.golang.org/api/kmsinventory/v1"
)

// --- gcloud kms inventory (#1104) ---

var kmsInventoryCmd = &cobra.Command{
	Use:   "inventory",
	Short: "KMS inventory and key-tracking",
}

var (
	flagKmsInvFormat        string
	flagKmsInvPageSize      int64
	flagKmsInvScope         string
	flagKmsInvCryptoKey     string
	flagKmsInvResourceTypes []string
	flagKmsInvKey           string
	flagKmsInvFallbackScope string
)

var kmsInvSummaryCmd = &cobra.Command{
	Use:   "get-protected-resources-summary",
	Short: "Get an aggregate summary of resources protected by a CryptoKey",
	Args:  cobra.NoArgs,
	RunE:  runKmsInvSummary,
}

var kmsInvListKeysCmd = &cobra.Command{
	Use:   "list-keys",
	Short: "List Cloud KMS keys in a project (from inventory snapshots)",
	Args:  cobra.NoArgs,
	RunE:  runKmsInvListKeys,
}

var kmsInvSearchCmd = &cobra.Command{
	Use:   "search-protected-resources",
	Short: "Search resources protected by a Cloud KMS CryptoKey",
	Args:  cobra.NoArgs,
	RunE:  runKmsInvSearch,
}

func init() {
	for _, c := range []*cobra.Command{kmsInvSummaryCmd, kmsInvListKeysCmd, kmsInvSearchCmd} {
		c.Flags().StringVar(&flagKmsInvFormat, "format", "", "Output format")
	}

	kmsInvSummaryCmd.Flags().StringVar(&flagKmsInvKey, "crypto-key", "", "Fully-qualified CryptoKey resource name (required)")
	kmsInvSummaryCmd.Flags().StringVar(&flagKmsInvFallbackScope, "fallback-scope", "", "Fallback scope: FALLBACK_SCOPE_PROJECT")
	_ = kmsInvSummaryCmd.MarkFlagRequired("crypto-key")

	kmsInvListKeysCmd.Flags().Int64Var(&flagKmsInvPageSize, "page-size", 0, "Page size (max 1000)")

	kmsInvSearchCmd.Flags().StringVar(&flagKmsInvScope, "scope", "", "Scope: organizations/ORG or projects/PROJECT (required)")
	kmsInvSearchCmd.Flags().StringVar(&flagKmsInvCryptoKey, "crypto-key", "", "Fully-qualified CryptoKey resource name (required)")
	kmsInvSearchCmd.Flags().StringSliceVar(&flagKmsInvResourceTypes, "resource-types", nil, "Optional resource type filters")
	kmsInvSearchCmd.Flags().Int64Var(&flagKmsInvPageSize, "page-size", 0, "Page size (max 500)")
	_ = kmsInvSearchCmd.MarkFlagRequired("scope")
	_ = kmsInvSearchCmd.MarkFlagRequired("crypto-key")

	kmsInventoryCmd.AddCommand(kmsInvSummaryCmd, kmsInvListKeysCmd, kmsInvSearchCmd)
	kmsCmd.AddCommand(kmsInventoryCmd)
}

func runKmsInvSummary(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.KMSInventoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.KeyRings.CryptoKeys.GetProtectedResourcesSummary(flagKmsInvKey).Context(ctx)
	if flagKmsInvFallbackScope != "" {
		call = call.FallbackScope(flagKmsInvFallbackScope)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("getting protected resources summary: %w", err)
	}
	return emitFormatted(out, flagKmsInvFormat)
}

func runKmsInvListKeys(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSInventoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := "projects/" + project
	var all []*kmsinventory.GoogleCloudKmsV1CryptoKey
	token := ""
	for {
		call := svc.Projects.CryptoKeys.List(parent).Context(ctx)
		if flagKmsInvPageSize > 0 {
			call = call.PageSize(flagKmsInvPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing keys: %w", err)
		}
		all = append(all, resp.CryptoKeys...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagKmsInvFormat)
}

func runKmsInvSearch(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.KMSInventoryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	scope := flagKmsInvScope
	var all []*kmsinventory.GoogleCloudKmsInventoryV1ProtectedResource
	token := ""
	// Dispatch on scope prefix so we hit the right REST parent.
	if strings.HasPrefix(scope, "projects/") {
		for {
			call := svc.Projects.ProtectedResources.Search(scope).CryptoKey(flagKmsInvCryptoKey).Context(ctx)
			if len(flagKmsInvResourceTypes) > 0 {
				call = call.ResourceTypes(flagKmsInvResourceTypes...)
			}
			if flagKmsInvPageSize > 0 {
				call = call.PageSize(flagKmsInvPageSize)
			}
			if token != "" {
				call = call.PageToken(token)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("searching protected resources: %w", err)
			}
			all = append(all, resp.ProtectedResources...)
			if resp.NextPageToken == "" {
				break
			}
			token = resp.NextPageToken
		}
	} else {
		for {
			call := svc.Organizations.ProtectedResources.Search(scope).CryptoKey(flagKmsInvCryptoKey).Context(ctx)
			if len(flagKmsInvResourceTypes) > 0 {
				call = call.ResourceTypes(flagKmsInvResourceTypes...)
			}
			if flagKmsInvPageSize > 0 {
				call = call.PageSize(flagKmsInvPageSize)
			}
			if token != "" {
				call = call.PageToken(token)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("searching protected resources: %w", err)
			}
			all = append(all, resp.ProtectedResources...)
			if resp.NextPageToken == "" {
				break
			}
			token = resp.NextPageToken
		}
	}
	return emitFormatted(all, flagKmsInvFormat)
}
