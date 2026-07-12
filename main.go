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
