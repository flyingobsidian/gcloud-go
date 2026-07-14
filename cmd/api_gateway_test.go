package cmd

import (
	"testing"

	"github.com/spf13/cobra"
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
