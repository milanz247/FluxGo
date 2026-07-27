# FluxGo CLI

The FluxGo CLI provides the local development workflow.

## Install

From the project root:

```sh
go install ./cmd/flux
```

Ensure Go's binary directory is on `PATH`. It is commonly `$HOME/go/bin`:

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

The CLI can also run without installation:

```sh
go run ./cmd/flux dev
```

## Hot reload

Start the development server:

```sh
flux dev
```

The command finds the project root, builds `./bootstrap`, starts the resulting
application, and watches:

- Go source files (`.go`)
- GoHTML templates (`.gohtml`)
- `.env`
- `go.mod` and `go.sum`

When a watched file changes, the CLI builds a new temporary binary. The running
application is restarted only after that build succeeds. If compilation fails,
the errors are printed and the last successful application continues running.

Generated binaries are stored under `.flux/tmp/`, which is ignored by Git.
Press `Ctrl+C` to stop the application and development runner.

The watcher ignores `.git`, `.flux`, `vendor`, and `node_modules`.

## Other commands

```sh
flux version
flux help
```
