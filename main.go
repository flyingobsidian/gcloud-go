package main

import (
	"os"

	"github.com/flyingobsidian/gcloud-go/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
