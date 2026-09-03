package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/flyingobsidian/gcloud-go/cmd"
)

func main() {
	executed, err := cmd.Execute()
	// If the failure was "API [...] not enabled on project [...]" and the user
	// confirms the prompt, MaybePromptEnableAPI enables the API and returns
	// retry=true — re-invoke Execute() so the same command runs against the
	// now-enabled API. Only one retry pass is allowed; a second "not enabled"
	// error falls through to the standard failure path.
	if err != nil {
		if retry, _ := cmd.MaybePromptEnableAPI(err); retry {
			executed, err = cmd.Execute()
		}
	}
	if err == nil {
		return
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}
	fmt.Fprintln(os.Stderr, cmd.FormatCLIError(executed, err))
	os.Exit(1)
}
