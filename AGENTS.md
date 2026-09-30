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
- `confile.ParseFromFile()` also shells out to `git` and populates built-in vars: `Branch`, `Sha`,
  `ShortSha`, `Dir`, `Os`, and `Project`.
- A step runs `pre` steps, then nested `steps`, `cmds`, or `cmd` (first match wins), then `post`
  steps; `--solo`, `--no-pre`, and `--no-post` skip hooks.
- Command strings and env values are templated through `vars.Vars.Process()`; unknown vars are left as-is.
- Child process env is step envs, then project envs, then the inherited process environment.
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
- Parsing fails outside a Git repo with at least one commit, and panics if `project.name` is missing.
- `config fmt` re-marshals the whole file and drops YAML comments and custom ordering.
- The README is minimal; treat `forge.yaml`, `cmd/`, and `confile/` as the behavior reference.
</KnownLandmines>
