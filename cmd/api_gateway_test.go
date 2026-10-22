package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	apigateway "google.golang.org/api/apigateway/v1"
)

func apiGatewaySubgroup(name string) *cobra.Command {
	for _, c := range apiGatewayCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestApiGatewayApisSubcommands(t *testing.T) {
	g := apiGatewaySubgroup("apis")
	if g == nil {
		t.Fatal("api-gateway apis missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestApiGatewayApiConfigsSubcommands(t *testing.T) {
	g := apiGatewaySubgroup("api-configs")
	if g == nil {
		t.Fatal("api-gateway api-configs missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestApiGatewayGatewaysSubcommands(t *testing.T) {
	g := apiGatewaySubgroup("gateways")
	if g == nil {
		t.Fatal("api-gateway gateways missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestApiGatewayOperationsSubcommands(t *testing.T) {
	g := apiGatewaySubgroup("operations")
	if g == nil {
		t.Fatal("api-gateway operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "describe", "list", "wait"})
}

func TestAGGWCreateEnableStreaming(t *testing.T) {
	if agGWCreateCmd.Flags().Lookup("enable-streaming") == nil {
		t.Fatal("gateways create missing --enable-streaming")
	}
	if agGWUpdateCmd.Flags().Lookup("enable-streaming") != nil {
		t.Error("--enable-streaming is create-only")
	}
	gw := &apigateway.ApigatewayGateway{}
	applyAGGWEnableStreaming(gw, false)
	if gw.StreamingMode != "" {
		t.Errorf("unset flag changed StreamingMode to %q", gw.StreamingMode)
	}
	applyAGGWEnableStreaming(gw, true)
	if gw.StreamingMode != "STREAMING_MODE_ENABLED" {
		t.Errorf("StreamingMode = %q, want STREAMING_MODE_ENABLED", gw.StreamingMode)
	}
}
