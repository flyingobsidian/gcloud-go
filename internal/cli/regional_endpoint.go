package cli

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"
)

// Endpoint modes of the regional/endpoint_mode property and --endpoint-mode.
const (
	endpointModeGlobal            = "global"
	endpointModeRegional          = "regional"
	endpointModeRegionalPreferred = "regional-preferred"
)

// addEndpointModeFlag adds --endpoint-mode to c, as gcloud-python does for
// commands annotated regional_endpoint_compatibility: SUPPORTED.
func addEndpointModeFlag(c *cobra.Command, mode *string) {
	c.Flags().StringVar(mode, "endpoint-mode", "",
		"Endpoint mode: global (default), regional, or regional-preferred (overrides regional/endpoint_mode)")
}

// endpointMode returns flagMode, or the regional/endpoint_mode property when
// the flag is unset.
func (inv *invocation) endpointMode(flagMode string) (string, error) {
	if flagMode != "" {
		return flagMode, nil
	}
	return inv.env.Property("regional", "endpoint_mode")
}

// regionalEndpoint returns the endpoint base URL for service in location, or
// "" for the global endpoint, following gcloud-python's
// _ShouldUseRegionalEndpoints for a command that supports both: mode
// regional always selects SERVICE.LOCATION.rep.googleapis.com,
// regional-preferred does so only for a location in known (the regions
// gcloud-python's API map lists for service), and global or unset uses the
// global endpoint.
func regionalEndpoint(mode, service, location string, known []string) (string, error) {
	regional := fmt.Sprintf("https://%s.%s.rep.googleapis.com/", service, location)
	switch mode {
	case "", endpointModeGlobal:
		return "", nil
	case endpointModeRegional:
		return regional, nil
	case endpointModeRegionalPreferred:
		if slices.Contains(known, location) {
			return regional, nil
		}
		return "", nil
	}
	return "", fmt.Errorf("argument --endpoint-mode: Invalid choice: '%s'.\n\nValid choices are [%s, %s, %s].",
		mode, endpointModeGlobal, endpointModeRegional, endpointModeRegionalPreferred)
}

// checkGlobalOnlyEndpointMode returns gcloud-python's error for a command
// without regional endpoint support when regional/endpoint_mode is regional.
func (inv *invocation) checkGlobalOnlyEndpointMode() error {
	mode, err := inv.env.Property("regional", "endpoint_mode")
	if err != nil {
		return err
	}
	if mode == endpointModeRegional {
		return fmt.Errorf("The regional/endpoint_mode property is currently set to [%s], but this command does not support regional endpoints.", mode)
	}
	return nil
}
