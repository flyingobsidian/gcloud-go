package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/flyingobsidian/gcloud-go/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
