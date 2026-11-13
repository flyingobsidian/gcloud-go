# Migrating a command group off package globals

gcloud-go is being made usable as a library ([#2070](https://github.com/flyingobsidian/gcloud-go/issues/2070)).
A command group is *migrated* when it keeps all its state in an `invocation`
(see `internal/cli/invocation.go`). Migrated groups can run in-process and
concurrently through the public `gcloud` package.

Every command group is now migrated, and the process-wide path they used
before has been removed: there is no global command tree, no `flag*`
globals and no process-wide forms of the helpers in the table below, so code
that uses them no longer compiles. New command groups must be written in the
migrated form from the start; `make lint` rejects package-level variables
and `init` functions (#2063). The command code lives in `internal/cli` (it
was `cmd/`), so the only public package is `gcloud`. This document records
how the groups were migrated; the "Before" column names helpers that no
longer exist.

`internal/cli/projects.go` is the reference example.

## Large groups: `tools/migrateglobals`

The rewriter was deleted once every group was migrated; it is in the git
history before #2063. For a large group, it was run rather than converting
by hand:

```sh
cd tools/migrateglobals
go run . -dir ../.. -files storage.go,storage_hmac.go,… \
    -type storageCmds -root storageCmd -ctor newStorageCmd
```

It keeps the group's code where it is, but:

- moves the command and flag variables onto a `storageCmds` struct and
  generates `newStorageCmd(inv)`, which builds the commands and then runs the
  former `init` functions (now methods) in file order, as Go did;
- turns every function that uses those variables, global flags or process
  state into a method, and applies the table below (`context.Background()`
  becomes the command's `cmd.Context()`, or `inv.context()` where there is no
  command);
- binds package-level REST clients (`*restClient`) to the invocation with
  `inv.rest(c)`, and turns `newRESTClient(endpoint)` into
  `inv.restClient(endpoint)`;
- with `-fixtests`, rewrites test functions that use the group's variables or
  converted functions to build their own group state for a
  `legacyTestInvocation()` (which keeps `captureStdout` and the REST test
  hooks working), and prints references outside functions to fix by hand;
- adds the group to `migratedGroups`;
- prints references from outside the group to fix by hand.

Things it does not handle, which `go vet ./cmd` or the tests showed:

- methods on other types, such as generic CRUD helpers (`ncCRUD[T]`): give
  the type a field holding the group state and route its globals through it;
- test helpers that set up group state for a test: `-fixtests` gives them
  their own state, so pass the test's state in instead;
- tests that call a run function directly: `cmd.Context()` is nil until the
  command runs, so call `setContexts` on the group first.

Run it with `-audit` to list helpers outside the group that the group calls
and that still use process state; give those an `io.Writer` or `*invocation`
form first. Then do steps 4 to 7 below. Every group from storage, logging
and compute on was migrated this way.

## Steps

1. **Commands are built by constructors.** Replace each `var xxxCmd = &cobra.Command{…}`
   and the `init()` that wires it with `func newXxxCmd(inv *invocation) *cobra.Command`.
   The group constructor adds its subcommands. Remove `rootCmd.AddCommand(xxxCmd)`.
2. **Flags bind to locals.** Declare the flag variables, or an options struct
   for more than two or three flags, inside the constructor and bind flags to them.
   `RunE` is a closure that passes them, with `c.Context()` and `inv`, to the
   run function:

   ```go
   RunE: func(c *cobra.Command, args []string) error {
       return runProjectList(c.Context(), inv, o)
   },
   ```

3. **Run functions take `(ctx context.Context, inv *invocation, …)`.** In them:

   | Before | After |
   |---|---|
   | `context.Background()` | the `ctx` parameter |
   | `gcp.XService(ctx, flagAccount)` | `inv.clients().X(ctx)` |
   | `fmt.Print*`, `os.Stdout` | `fmt.Fprint*(inv.stdout, …)` |
   | `os.Stderr` | `inv.stderr` |
   | `os.Stdin` | `inv.stdin` |
   | `emitFormatted(v, flagFormat)` | `inv.emit(v)` |
   | `emitFormatted(v, localFormat)` | `emitFormattedTo(inv.stdout, v, localFormat)` |
   | `yamlEncode(v)` / `printJSON(v)` | `inv.yaml(v)` / `inv.printJSON(v)` |
   | `printTableHeading(…)` | `printTableHeadingTo(inv.stdout, …)` |
   | `emitWithPipeline(v)` | `inv.emitWithPipeline(v)` |
   | `promptYesNo(…)` | `inv.confirm(…)` |
   | `resolveProject()` | `inv.project()` |
   | `resolveProjectZone()` | `inv.projectZone()` |
   | `config.Load()` | `inv.properties()` |
   | `config.Resolve(…)` / `config.Property(…)` | `inv.resolve(…)` / `inv.env.Property(…)` |
   | `flagProject`, `flagFormat`, `flagQuiet`, … | `inv.flags.project`, `inv.flags.format`, `inv.flags.quiet`, … |
   | `httpDebugWriter()` | `inv.httpDebugWriter()` (already applied by `inv.clients()`) |

   When a shared helper writes to the process streams or reads globals, give
   it an `io.Writer` or `*invocation` form.
4. **Register the group** in `migratedGroups` (`internal/cli/inprocess.go`).
5. **Fix tests** that set flag globals or use the old command vars. Build
   commands with `newXxxCmd(inv)` using `newTestInvocation`, or run them end
   to end with `newFakeAPIInvocation` and `runInvocation` (see
   `internal/cli/projects_inprocess_test.go`).
6. **Add a `-race` test** that runs several invocations of the group
   concurrently and checks that their outputs stay separate.
7. **Check output parity.** Build `main` and the branch, and diff stdout,
   stderr and exit status for the group's commands, including `--help` and
   error cases. Output must be byte-identical.

## Done when

```sh
grep -nE 'flag(Project|Zone|Account|Quiet|Format|Filter|Flatten|Verbosity|LogHTTP)\b|context\.Background|os\.(Stdout|Stderr|Stdin)|fmt\.Print|gcp\.[A-Za-z]+Service\(' internal/cli/<group>*.go
```

prints nothing for the group's files.
