package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/adamwasila/go-semver"
)

var extraHelp = "\n" +
	"  Bump to new version.\n" +
	"\n\n"

func main() {
	flag.CommandLine.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTIONS]... version\n", os.Args[0])
		fmt.Fprint(flag.CommandLine.Output(), extraHelp)
		flag.PrintDefaults()
	}

	var (
		major, minor, patch, prerelease, release bool

		buildmetadata string
		keepMetadata  bool
	)

	flag.BoolVar(&major, "major", false, "Bump to next major version")
	flag.BoolVar(&minor, "minor", false, "Bump to next minor version")
	flag.BoolVar(&patch, "patch", false, "Bump to next patch version")
	flag.BoolVar(&prerelease, "prerelease", false,
		"Try to upgrade to next prerelese version by incrementing"+
			" last number component in prerelease tag")
	flag.BoolVar(&release, "release", false, "Strip prerelease from version")

	flag.StringVar(&buildmetadata, "meta", "", "Optional build metadata attached to new version. Can be used multiple times.")
	flag.BoolVar(&keepMetadata, "keep-meta", false, "Do not reset originam metadata when bumping to new version")

	flag.Parse()
	versions := flag.Args()

	if len(versions) != 1 {
		fmt.Fprintf(flag.CommandLine.Output(), "expected single argument: version\n")
		os.Exit(1)
	}

	version := versions[0]

	parsedVersion, err := semver.Parse(version)
	if err != nil {
		fmt.Printf("Invalid version: '%s', %s\n", version, err)
		os.Exit(1)
	}

	var newVersion semver.Version
	var opts []semver.BumpOption

	if keepMetadata {
		for _, bm := range parsedVersion.Buildmetadata {
			opts = append(opts, semver.BuildMetadata(bm))
		}
	}

	if buildmetadata != "" {
		opts = append(opts, semver.BuildMetadata(buildmetadata))
	}

	switch {
	case major:
		opts = append(opts, semver.BumpMajor())
	case minor:
		opts = append(opts, semver.BumpMinor())
	case patch:
		opts = append(opts, semver.BumpPatch())
	case prerelease:
		opts = append(opts, semver.BumpPrelease())
	case release:
		opts = append(opts, semver.BumpRelease())
	default:
		opts = append(opts, semver.BumpPatch())
	}

	newVersion, err = parsedVersion.Bump(opts...)
	if err != nil {
		fmt.Printf("Bump '%s' failed: %s\n", version, err)
		os.Exit(1)
	}

	fmt.Println(newVersion.String())

	os.Exit(0)
}
