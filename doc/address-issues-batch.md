# Addressing Issues in Batches

## Resources

- **gcloud-python** - the original Python implementation of the Google Cloud SDK CLI
  - Source: `~/src/github.com/flyingobsidian/gcloud-go/reference/google-cloud-sdk-568.0.0/` - read-only. Do not modify files here.
  - Binary: `/usr/bin/gcloud`
- **gcloud-go** - our Golang implementation of the Google Cloud SDK CLI
  - Source: `~/.openclaw/workspace/src/github.com/flyingobsidian/gcloud-go/` - your private read-write workspace. Do all work here.
  - Binary: `make build` -> `bin/gcloud-go`

## Rules

The Golang implementation should match the reference Python implementation.

# Branches

Name branches `batch-1234-1235-1236` where:
- `batch` is just the word "batch"
- `1234-1235-1236` is a dash-separated list of *all* the Issue numbers being worked on.

# PRs

In the PR description, include a "Closes #" plus Issue number for each Issue that the PR addresses.

Wait for CI to pass, then merge the PR. If CI fails, read and address the problems. After merging the PR, switch back to main, pull the latest, so you are ready for the next task.

## Process

Get a list of open Issues by
- reading `issues.tsv` (not committed to git)
- using the `gh` CLI, if the file above is not present

If the list of open Issues is empty, stop. If not, continue.

If possible, batch Issues into small groups. This way, CI is run once per group, not once per Issue. Issues can be batched if the title looks like `Implement [command] [subcommand] subcommands` and the top-level `command` matches between Issues.

If addressing an Issue is not possible with the current version of `google.golang.org/api`, check for a comment on the Issue noting this fact. If a comment is not found, make a comment, and leave the Issue open.

Go back to the start and repeat until all Issues are closed.
