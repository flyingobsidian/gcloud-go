# Addressing Issues One by One

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

Name branches `1234-issue-name-slug` where:
- `1234` is the correct number of the Issue being worked on.
- `issue-name-slug` is the lowercase dash-separated Issue name, with common words (like "a", "the" etc) removed, and with non-alphabetic characters removed.

## Process

Get the Issue information by using the `gh` CLI.

Read the Issue description.

Implement the feature (or fix the bug) on a separate branch.

Commit changes, push to `origin`.

Create a PR. In the PR description, use "Closes #" plus the Issue number so that merging the PR closes the Issue.

Wait for CI to pass, then merge the PR. If CI fails, read and address the problems. After merging the PR, switch back to main, pull the latest, so you are ready for the next task.
