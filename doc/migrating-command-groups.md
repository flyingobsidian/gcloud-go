# Migrating a command group off package globals

gcloud-go is being made usable as a library ([#2070](https://github.com/flyingobsidian/gcloud-go/issues/2070)).
A command group is *migrated* when it keeps all its state in an `invocation`
(see `cmd/invocation.go`). Migrated groups can run in-process and
concurrently through the public `gcloud` package. Unmigrated groups still run
in the CLI, but `gcloud.Runner.Run` returns `ErrNotInProcess` for them.

`cmd/projects.go` is the reference example.

## Large groups: `tools/migrateglobals`

For a large group, start with the rewriter rather than converting by hand:

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
- prints references from outside the group (usually tests) to fix by hand.

Run it with `-audit` to list helpers outside the group that the group calls
and that still use process state; give those an `io.Writer` or `*invocation`
form first. Then do steps 4 to 7 below. storage, logging and compute were
migrated this way.

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

   When a shared helper writes to the process streams or reads globals, add
   an `io.Writer` or `*invocation` form next to it, and keep the old form
   for the groups that still use it.
4. **Register the group** in `migratedGroups` (`cmd/inprocess.go`).
5. **Fix tests** that set flag globals or use the old command vars. Build
   commands with `newXxxCmd(inv)` using `newTestInvocation`, or run them end
   to end with `newFakeAPIInvocation` and `runInvocation` (see
   `cmd/projects_inprocess_test.go`).
6. **Add a `-race` test** that runs several invocations of the group
   concurrently and checks that their outputs stay separate.
7. **Check output parity.** Build `main` and the branch, and diff stdout,
   stderr and exit status for the group's commands, including `--help` and
   error cases. Output must be byte-identical.

## Done when

```sh
grep -nE 'flag(Project|Zone|Account|Quiet|Format|Filter|Flatten|Verbosity|LogHTTP)\b|context\.Background|os\.(Stdout|Stderr|Stdin)|fmt\.Print|gcp\.[A-Za-z]+Service\(' cmd/<group>*.go
```

prints nothing for the group's files, apart from deliberate legacy wrappers.
