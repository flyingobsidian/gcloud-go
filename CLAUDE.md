Read the [readme](/README.md) for an overview of the repo.

# Development

## git branches

Do not commit to the `main` branch. Use feature/bugfix branches, named as follows:

- Issue number followed by `-` (if addressing an Issue)
- short representation of the feature/bug name, all lowercase, with dashes

Examples:
- `123-improve-some-feature`
- `456-fix-some-bug`

## git commits

Do not commit to the `main` branch.

Make commits incremental and keep them small. If in doubt, split changes into more smaller commits rather than fewer larger commits.

## git commit messages

Keep commit messages to one line only. No paragraphs. No emoji. No bullet point lists. No list of files changed.

## git push

Before running `git push`, run tests. See [Makefile](/Makefile) for details.
