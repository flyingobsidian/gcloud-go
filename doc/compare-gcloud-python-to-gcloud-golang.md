# Compare `gcloud-go` to `gcloud-python`

## Resources

Use `~` to mean the Linux home directory of your user account.

- **gcloud-python** - the original Python implementation of the Google Cloud SDK CLI
  - Source: `~/src/github.com/flyingobsidian/gcloud-go/reference/google-cloud-sdk-568.0.0/` - read-only. Do not modify files here.
  - Binary: `/usr/bin/gcloud`
- **gcloud-go** - our Golang implementation of the Google Cloud SDK CLI
  - Source: `~/.openclaw/workspace/src/github.com/flyingobsidian/gcloud-go/` - your private read-write workspace. Do all work here.
  - Binary: `make build` -> `bin/gcloud-go`

## Goal

These two implementations should be identical in functionality, though the Golang implementation may have additions so that it works in some extra situations.

# Task

Analyse the reference **gcloud-python**, and compare it against the new **gcloud-go**. Look for inconsistencies in behaviour.

Using `gcloud --help`, look at each command, then for each command, look at each subcommand. Continue recursing as necessary. Also consider global arguments that apply to all commands, and per-command/per-subcommand arguments.

Using the `gh` CLI:
- Read a list of existing open Issues
- If there is not aleady an existing open Issue, create a Github Issue for each occurrence of the following:
  - an inconsistency between **gcloud-python** and **gcloud-go**
  - a missing command, or subcommand, or argument in **gcloud-go**
  - an unimplemented command or subcommand. Look in particular for `Not yet implemented` and `registerStub`. Use the shell command `find -name '*.go' |xargs grep -E '(Not yet impl|registerStub)'` to find occurrences.

Present a summary report to the user at the end, listing Issues created and what each Issue addresses.
