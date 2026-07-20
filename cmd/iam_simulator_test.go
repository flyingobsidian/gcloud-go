package cmd

import "testing"

func TestIamSimulatorSubgroups(t *testing.T) {
	g := iamSubgroup("simulator")
	if g == nil {
		t.Fatal("iam simulator missing")
	}
	assertSubcommands(t, g, []string{"replays"})
	replays := findSub(g, "replays")
	if replays == nil {
		t.Fatal("iam simulator replays missing")
	}
	assertSubcommands(t, replays, []string{"create", "describe"})
}
