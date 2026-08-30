# Compare previous `gcloud-python` to new `gcloud-python`

## Definitions

- `previous_version` is `568.0.0`
- `new_version` is `582.0.0`

## Resources

Use `~` to mean the Linux home directory of your user account.

- **gcloud-python-previous** - version `previous_version` of the original Python implementation of the Google Cloud SDK CLI
  - Source: `~/src/github.com/flyingobsidian/gcloud-go/reference/google-cloud-sdk-<previous_version>/` - read-only. Do not modify files here.
- **gcloud-python-new** - version `new_version` of the original Python implementation of the Google Cloud SDK CLI
  - Source: `~/src/github.com/flyingobsidian/gcloud-go/reference/google-cloud-sdk-<new_version>/` - read-only. Do not modify files here.
- **gcloud-go** - our Golang implementation of the Google Cloud SDK CLI, currently based on version `previous_version`.
  - Source: `~/.openclaw/workspace/src/github.com/flyingobsidian/gcloud-go/` - your private read-write workspace. Do all work here.
  - Binary: `make build` -> `bin/gcloud-go`

## Binaries

- `/usr/bin/gcloud` - version 575.0.0 - do not use this.

## Goal

`gcloud-go` was written with reference to version `previous_version` of `gcloud-python`. This is now out of date, since the most recent version is `new_version`.

`gcloud-go` should match version `new_version` of the reference `gcloud-python`.

# Task

Going subdirectory by subdirectory, file by file, compare **gcloud-python-previous** to **gcloud-python-new**. Look for new functionality added, old functionality removed, or functionality changed. For each change, use the `gh` CLI to create an Issue so that the addition/removal/change can be implemented in `gcloud-go`.

Present a summary report to the user at the end, listing Issues created and what each Issue addresses.

Note that some commands in `gcloud-go`, e.g. `scc ...`, have extended functionality which is in neither `previous_version` nor `new_version` of `gcloud-python`.
