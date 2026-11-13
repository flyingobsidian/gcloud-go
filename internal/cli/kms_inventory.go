package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	kmsinventory "google.golang.org/api/kmsinventory/v1"
)

// --- gcloud kms inventory (#1104) ---

func (s *kmsCmds) initKmsInventory() {
	for _, c := range []*cobra.Command{s.kmsInvSummaryCmd, s.kmsInvListKeysCmd, s.kmsInvSearchCmd} {
		c.Flags().StringVar(&s.flagKmsInvFormat, "format", "", "Output format")
	}

	s.kmsInvSummaryCmd.Flags().StringVar(&s.flagKmsInvKey, "crypto-key", "", "Fully-qualified CryptoKey resource name (required)")
	s.kmsInvSummaryCmd.Flags().StringVar(&s.flagKmsInvFallbackScope, "fallback-scope", "", "Fallback scope: FALLBACK_SCOPE_PROJECT")
	_ = s.kmsInvSummaryCmd.MarkFlagRequired("crypto-key")

	s.kmsInvListKeysCmd.Flags().Int64Var(&s.flagKmsInvPageSize, "page-size", 0, "Page size (max 1000)")

	s.kmsInvSearchCmd.Flags().StringVar(&s.flagKmsInvScope, "scope", "", "Scope: organizations/ORG or projects/PROJECT (required)")
	s.kmsInvSearchCmd.Flags().StringVar(&s.flagKmsInvCryptoKey, "crypto-key", "", "Fully-qualified CryptoKey resource name (required)")
	s.kmsInvSearchCmd.Flags().StringSliceVar(&s.flagKmsInvResourceTypes, "resource-types", nil, "Optional resource type filters")
	s.kmsInvSearchCmd.Flags().Int64Var(&s.flagKmsInvPageSize, "page-size", 0, "Page size (max 500)")
	_ = s.kmsInvSearchCmd.MarkFlagRequired("scope")
	_ = s.kmsInvSearchCmd.MarkFlagRequired("crypto-key")

	s.kmsInventoryCmd.AddCommand(s.kmsInvSummaryCmd, s.kmsInvListKeysCmd, s.kmsInvSearchCmd)
	s.kmsCmd.AddCommand(s.kmsInventoryCmd)
}

func (s *kmsCmds) runKmsInvSummary(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMSInventory(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.KeyRings.CryptoKeys.GetProtectedResourcesSummary(s.flagKmsInvKey).Context(ctx)
	if s.flagKmsInvFallbackScope != "" {
		call = call.FallbackScope(s.flagKmsInvFallbackScope)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("getting protected resources summary: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsInvFormat)
}

func (s *kmsCmds) runKmsInvListKeys(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMSInventory(ctx)
	if err != nil {
		return err
	}
	parent := "projects/" + project
	var all []*kmsinventory.GoogleCloudKmsV1CryptoKey
	token := ""
	for {
		call := svc.Projects.CryptoKeys.List(parent).Context(ctx)
		if s.flagKmsInvPageSize > 0 {
			call = call.PageSize(s.flagKmsInvPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagKmsInvFormat)
}

func (s *kmsCmds) runKmsInvSearch(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMSInventory(ctx)
	if err != nil {
		return err
	}
	scope := s.flagKmsInvScope
	var all []*kmsinventory.GoogleCloudKmsInventoryV1ProtectedResource
	token := ""
	// Dispatch on scope prefix so we hit the right REST parent.
	if strings.HasPrefix(scope, "projects/") {
		for {
			call := svc.Projects.ProtectedResources.Search(scope).CryptoKey(s.flagKmsInvCryptoKey).Context(ctx)
			if len(s.flagKmsInvResourceTypes) > 0 {
				call = call.ResourceTypes(s.flagKmsInvResourceTypes...)
			}
			if s.flagKmsInvPageSize > 0 {
				call = call.PageSize(s.flagKmsInvPageSize)
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
			call := svc.Organizations.ProtectedResources.Search(scope).CryptoKey(s.flagKmsInvCryptoKey).Context(ctx)
			if len(s.flagKmsInvResourceTypes) > 0 {
				call = call.ResourceTypes(s.flagKmsInvResourceTypes...)
			}
			if s.flagKmsInvPageSize > 0 {
				call = call.PageSize(s.flagKmsInvPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagKmsInvFormat)
}
