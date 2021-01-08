![running gophers](running_gophers.png)

# go-semver

Simple version string parser, bulider, bumper etc. Follows strictly [semantic Versioning 2.0.0 specification](https://semver.org/) and has unit tests to ensure it works for most edge cases.

> Note: this isn't stable version yet. It should do the job and all tests are passing bug API will be freezed only after version 1.0.0 tag. No promises till that point.

## Features

- Validate version stored in a string.
- Parse and unpack to standarized `Version` structure where it can be easily introspected or used for higher "business" logic.
- Bump parsed structure to next version.
- Operator to compare two versions: allows choosing max version, sorting etc.

## Install

Run `go get github.com/adamwasila/go-semver`

## Requirements

go 1.15 or newer is preferred. Run all unit tests first before using any older version of go.

## Examples

Create new semver struct:

```go
sv, _ := semver.New(
    semver.SetCore("1.2.3"),
    semver.Prerelease("rc.1"), semver.BuildMetadata("cafebabe"),
)
fmt.Printf("%s", sv.String())
```

Output:

```console
1.2.3-rc.1+cafebabe
```

Parse string with version (will panic if version is incorrect):

```go
v := "1.2.3-ver.12a+build.1234"
sv := semver.MustParse(v)
fmt.Printf("%s == %s", v, sv.String())
```

Output:

```console
1.2.3-ver.12a+build.1234 == 1.2.3-ver.12a+build.1234
```

Bump version:

```go
sv := semver.MustParse("1.2.3-rc.1+cafebabe")
sv, _ = sv.Bump(semver.BreakingChange)
fmt.Printf("%s", sv.String())
```

Output:

```console
2.0.0
```

## License

Distributed under Apache License Version 2.0. See [LICENSE](LICENSE) for more information.

## Contact

To discuss about bugs or features: please fill gitlab issue ticket.

Pull Requests except for trivial (eg. typo) fixes should also be discussed by opening issue first.
