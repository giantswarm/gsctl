# gsctl Development Documentation

Please read this if you intend to develop on gsctl.

## Required Tools, Prerequisites

- Go environment (`brew install go`)
- [`dep`](https://github.com/golang/dep)
- GNU Make
- `git`
- Docker

## Cloning to the right location

Make sure you have the `$GOPATH` environment variable set.

```nohighlight
$ echo $GOPATH
/Users/johndoe/go
```

Go to right location, then check out:

```nohighlight
$ mkdir -p $GOPATH/src/github.com/giantswarm
$ cd $GOPATH/src/github.com/giantswarm
$ git clone https://github.com/giantswarm/gsctl.git
$ cd gsctl
```

So the repo content will end up in `$GOPATH/src/github.com/giantswarm/gsctl`.

## Dependencies

Dependencies are managed using [`go dep`](https://github.com/golang/dep).

## Executing gsctl during development

One option is to execute the program via `go run`, like in this example:

```nohighlight
# ensure packr binary is up-to-date
$ packr

$ go run main.go info
```

Or you can first build a binary and then execute it.

```nohighlight
$ go build && ./gsctl info
```

To build a binary for your platform like the release build would do, do this:

```nohighlight
$ make clean
$ make
$ make install
```

## Running tests

The `Makefile` provides a few shortcuts.

To execute all Go unit tests:

```nohighlight
make gotest
```

To quickly run a number of commands:

```nohighlight
make test
```

## Embedded HTML files (packr)

For the `sso` command, gsctl needs to run a local webserver and show nicely formated
html to the user. These html files are found in this folder.

The files get compiled into a binary file in the `oidc` package
by running the `packr` command.

To learn more about `packr`, visit: https://github.com/gobuffalo/packr


## Coding Style

Before pushing any changes, please:

- Let `gofmt` format your code
- Do a `golint .` check and act on recommendations before pushing.

## Conventions

See [Command Blueprint](https://github.com/giantswarm/gsctl/blob/master/docs/Command-Blueprint.md) for a scaffold of a command file.

### Typed Errors

We use specific error objects and dedicated matcher functions to assert them. Example:

```go
var NotLoggedInError = errgo.New("user not logged in")

// IsNotLoggedInError asserts NotLoggedInError.
func IsNotLoggedInError(err error) bool {
	return errgo.Cause(err) == NotLoggedInError
}
```

## Publishing a Release

See [RELEASE.md](https://github.com/giantswarm/gsctl/blob/master/docs/Release.md)
