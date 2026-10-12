# Forge

A CLI tool for executing, in a consistent manner, scripts and commands for building and maintaining projects.

## Versioning

The project will be versioned in accordance with [Semver 2.0.0](https://semver.org). See the [releases](https://github.com/gomicro/forge/releases) section for the latest version. Until version 1.0.0 the project is considered to be unstable.

It is always highly recommended to vendor the version you are using.

## Docker Image

The published `ghcr.io/gomicro/forge` image is built `FROM scratch` and contains only the
statically linked `forge` binary. Use it as a binary carrier in a multi-stage Docker build with
`COPY --from`; it does not include the `bash` or `git` executables required to run project steps.

## License
See [LICENSE.md](./LICENSE.md) for more information.
