package main

import (
	"context"
	"os"

	"github.com/flyingobsidian/gcloud-go/cmd"
)

func main() {
	os.Exit(cmd.Main(context.Background(), os.Args[1:]))
}
