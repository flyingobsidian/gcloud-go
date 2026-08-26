package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// The emulator subcommands wrap standalone binaries shipped with the Python
// Cloud SDK (cbtemulator, cloud_datastore_emulator, pubsub-emulator etc.).
// gcloud-go doesn't bundle those, so `start` shells out to whichever copy is
// on PATH or under $CLOUDSDK_ROOT_DIR/platform. `env-init` prints the same
// shell exports gcloud-python does so scripts that source the output work
// unchanged.

// emulatorLookupPath returns the resolved path to bin, searching PATH first
// and then a handful of common Cloud SDK install locations. Returns an error
// with usage guidance if nothing is found.
func emulatorLookupPath(bin string, sdkSubdir string) (string, error) {
	if p, err := exec.LookPath(bin); err == nil {
		return p, nil
	}
	// Common Cloud SDK layouts. Users typically have exactly one of these.
	candidates := []string{
		os.Getenv("CLOUDSDK_ROOT_DIR"),
		"/usr/lib/google-cloud-sdk",
		"/opt/google-cloud-sdk",
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, home+"/google-cloud-sdk")
	}
	for _, root := range candidates {
		if root == "" {
			continue
		}
		p := root + "/platform/" + sdkSubdir + "/" + bin
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p, nil
		}
	}
	return "", fmt.Errorf(
		"%s not found on PATH or under a known Cloud SDK install. "+
			"Install the emulator via `gcloud components install %s` "+
			"(Python Cloud SDK) or point $CLOUDSDK_ROOT_DIR at your install",
		bin, sdkComponentFor(bin))
}

// sdkComponentFor maps an emulator binary name to the gcloud component the
// Python SDK uses to distribute it. This drives the "install via" hint in the
// not-found error.
func sdkComponentFor(bin string) string {
	switch bin {
	case "cbtemulator":
		return "bigtable"
	case "cloud_datastore_emulator", "cloud-datastore-emulator":
		return "cloud-datastore-emulator"
	case "cloud_firestore_emulator", "cloud-firestore-emulator":
		return "cloud-firestore-emulator"
	case "pubsub-emulator", "pubsub_emulator":
		return "pubsub-emulator"
	case "cloud_spanner_emulator", "cloud-spanner-emulator":
		return "cloud-spanner-emulator"
	}
	return bin
}

// runEmulator runs bin (resolved via emulatorLookupPath) with the given argv,
// forwarding stdio. Ctrl-C in a foreground terminal reaches the child via the
// process group so both exit together.
func runEmulator(bin, sdkSubdir string, argv []string) error {
	path, err := emulatorLookupPath(bin, sdkSubdir)
	if err != nil {
		return err
	}
	cmd := exec.Command(path, argv...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// printEnvInit prints the shell exports for an emulator's env variables. The
// output format mirrors `gcloud emulators <x> env-init` -- one `export`
// statement per line so `eval "$(gcloud-go emulators X env-init)"` works.
func printEnvInit(vars map[string]string) error {
	for k, v := range vars {
		if _, err := fmt.Fprintf(os.Stdout, "export %s=%s\n", k, shellQuote(v)); err != nil {
			return err
		}
	}
	return nil
}

// printEnvUnset prints unset commands for the given env variables.
func printEnvUnset(keys []string) error {
	for _, k := range keys {
		if _, err := fmt.Fprintf(os.Stdout, "unset %s\n", k); err != nil {
			return err
		}
	}
	return nil
}

// hostFrom returns the host portion of "host:port". A bare host returns
// that string unchanged.
func hostFrom(hostPort string) string {
	if i := strings.LastIndexByte(hostPort, ':'); i >= 0 {
		return hostPort[:i]
	}
	return hostPort
}

// portFrom returns the port portion of "host:port" (empty when no ':').
func portFrom(hostPort string) string {
	if i := strings.LastIndexByte(hostPort, ':'); i >= 0 {
		return hostPort[i+1:]
	}
	return ""
}

// shellQuote returns s wrapped in single quotes if it contains characters
// that would be interpreted by a POSIX shell.
func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " \t\n\"'\\$`") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// registerEmulatorGroup replaces the earlier stub registration for `emulators
// X` when the real subcommands are ready. It removes the stub group of the
// same name (added in emulators.go's init) so cobra doesn't see two children
// with matching names.
func registerEmulatorGroup(g *cobra.Command) {
	// Cobra doesn't provide a public Remove, but each stub group registered
	// in emulators.go is added by pointer via registerStubGroup. We rebuild
	// emulatorsCmd.commands by filtering out anything whose name matches.
	current := emulatorsCmd.Commands()
	emulatorsCmd.ResetCommands()
	for _, c := range current {
		if c.Name() == g.Name() {
			continue
		}
		emulatorsCmd.AddCommand(c)
	}
	emulatorsCmd.AddCommand(g)
}
