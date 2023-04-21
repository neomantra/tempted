# tempted - Temporal Text User Interface

`tempted` is a TUI (Textual User Interface) for [Temporal](https://temporal.io/).  It is an CLI alternative to [`tctl`](https://github.com/temporalio/tctl), seeking to give an interative experince like the [Temporal Web UI](https://docs.temporal.io/web-ui).

 * Currently tracking [this GitHub issue](https://github.com/temporalio/tctl/issues/359)

## Usage

```
$ tempted --help
```

## Environment Variables

The following environment variables affect program runtime:

| Name  | Default | Description |
| --- | --- |
| `TEMPORAL_CLI_ADDRESS` |"localhost:7233:7234" | `host:port` for Temporal frontend service |

## Installing

Binaries for multiple platforms are [released on GitHub](https://github.com/neomantra/tempted/releases) through [GitHub Actions](https://github.com/neomantra/tempted/actions).

You can also install for various platforms with [Homebrew](https://brew.sh) from [`neomantra/homebrew-tap`](https://github.com/neomantra/homebrew-tap):

```
brew tap neomantra/homebrew-tap
brew install neomantra/homebrew-tap/tempted
```

----

## Example Usage

```
TODO
```

----

## SSH App

`tempted` can be served via ssh application. For example, you could host an internal ssh application for your company such that anyone on the internal network can `ssh -p <your-port> <your-host>` and immediately access `tempted` without installing or configuring anything.

Serve the ssh app with `tempted serve`.

----

## Building

Building is performed with [task](https://taskfile.dev/):

```
$ task
```

----

## Credits and License

This Text Application is not only inspired by [`wander`](https://github.com/robinovitch61/wander), a similar tool for [HashiCorp Nomad](https://nomadproject.io).  The entire initial implementation was copied, then I ported an initial spike working in a few hours!

This software is released with the same license as [Temporal](https://github.com/temporalio/temporal/blob/master/LICENSE), with gratitude to and no affiliation with [Temporal.io](https://temporal.io) and [Charm.sh](https://charm.sh) and [robinovitch61](https://github.com/robinovitch61).  

Copyright (c) 2023 [Neomantra BV](https://www.neomantra.com).  Authored by Evan Wies.

Released under the [MIT License](https://en.wikipedia.org/wiki/MIT_License), see [LICENSE.txt](./LICENSE.txt).
