package main

import (
	"context"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/cli"
)

func main() {
	os.Exit(cli.Main(context.Background(), os.Args[1:]))
}
