# Agent Instructions

<Overview>
`forge` is a Go CLI that runs named project steps defined in a `forge.yaml` in the current
directory, giving projects one consistent entry point for build, test, and maintenance scripts.
</Overview>

<TechStack>
- Go 1.26, pinned by CI in `.github/workflows/build.yml` (`go.mod` still declares 1.23.3).
- `cobra` for commands, `viper` for global flag access, `gopkg.in/yaml.v3` for config parsing.
- `github.com/gomicro/scribe` for themed, nested step output; `testify` for test assertions.
- Commands execute via `bash -c`; GoReleaser builds release binaries and the `ghcr.io` image.
</TechStack>

<BuildValidate>
```sh
go fmt ./...
```

```sh
go vet ./...
```

```sh
go build ./...
```

```sh
go test ./...
```

```sh
golangci-lint run
```
</BuildValidate>

<RepositoryLayout>
- `cmd/` — Cobra root command (step runner), `completion`, and `version`.
- `cmd/config/` — `config` subcommands (`init`, `fmt`) that create and reformat `forge.yaml`.
- `confile/` — `forge.yaml` model, parsing, and recursive step execution.
- `vars/` — `{{.Name}}` template variable storage and substitution; tests live alongside.
- `forge.yaml` — this repo's own step config and the best working example of the format.
- `vendor/` — vendored dependencies.
</RepositoryLayout>

<ArchitectureNotes>
- `forge` reads only `./forge.yaml`; there is no config discovery beyond the working directory.
- `confile.Parse(path)` only reads and validates YAML. `(*File).ResolveVars(dir)` shells out to `git`
  for the built-in vars (`Branch`, `Sha`, `ShortSha`, `Dir`, `Os`, `Project`); only step execution calls it.
- A step runs `pre` steps, then exactly one of nested `steps`, `cmds`, or `cmd`, then `post`
  steps; `--solo`, `--no-pre`, and `--no-post` skip hooks.
- `Parse` validates the whole step graph up front (empty or ambiguous steps, undefined references,
  cycles through `pre`/`steps`/`post`) and reports every problem; any invalid step blocks all runs.
- Command strings and env values are templated through `vars.Vars.Process()`; unknown vars are left as-is.
- `cmd/` builds `confile.RunOptions`; every child receives the same vars, project envs, and flags.
  Steps hold only configuration; execution uses the caller's context and does not read Viper globals.
- Child process env precedence is step > project > inherited process environment, including empty values.
  Only config env values are templated; inherited values pass through unchanged.
- Root completion parses `forge.yaml` to suggest step names with their `help` text.
- `cmd.Version` is injected at build time via `-ldflags`; empty means `dev-local`.
</ArchitectureNotes>

<KeyConventions>
- Keep CLI wiring in `cmd/` and config/runtime behavior in `confile/`.
- Never call `fmt.Print*` outside `cmd/` (enforced by `forbidigo`); emit step output through `scribe.Scriber`.
- Preserve the execution order and env precedence above unless intentionally changing config semantics.
- Preserve the built-in template variable names; `forge.yaml` files in other repos depend on them.
- Add focused tests beside the changed package; coverage is light outside `vars/`.
- Never edit `vendor/` directly; change `go.mod` and run `go mod vendor`.
</KeyConventions>

<KnownLandmines>
- Running steps requires a Git repo with at least one commit; `config` commands and completion do not.
- Parsing errors if the `project` block is missing.
- `config fmt` re-marshals the whole file and drops YAML comments and custom ordering.
- The README is minimal; treat `forge.yaml`, `cmd/`, and `confile/` as the behavior reference.
</KnownLandmines>
