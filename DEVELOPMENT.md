# gsctl Development Documentation

Please read this if you intend to develop on gsctl.

## Required Tools, Prerequisites

- Go environment (`brew install go`)
- [`glide`](https://github.com/Masterminds/glide) (`brew install glide`)
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

Dependencies are managed using [`glide`](https://github.com/Masterminds/glide).

**CAUTION:** The `vendor` directory is _flattened_. Make sure to use glide with the `-v` (or `--strip-vendor`) flag.

## Executing gsctl During Development

One option is to execute the program via `go run`, like in this example:

```nohighlight
$ go run main.go info
```

Or you can first build a binary and then execute it.

```nohighlight
$ go build && ./gsctl info
```

To build a binary for your platform like the release build would do, do this:

```nohighlight
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

## Coding Style

Before pushing any changes, please:

- Let `gofmt` format your code
- Do a `golint .` check and act on recommendations before pushing.

## Conventions

### Typed Errors

We use specific error objects and dedicated matcher functions to assert them. Example:

```go
var notLoggedInError = errgo.New("user not logged in")

// IsNotLoggedInError asserts notLoggedInError.
func IsNotLoggedInError(err error) bool {
	return errgo.Cause(err) == notLoggedInError
}
```

## Publishing a Release

See [RELEASE.md](https://github.com/giantswarm/gsctl/blob/master/RELEASE.md)
