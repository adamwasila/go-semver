package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/adamwasila/semver"
)

var extraHelp string = "\n" +
	"  Bump to new version.\n" +
	"\n\n"

func main() {
	flag.CommandLine.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTIONS]... version\n", os.Args[0])
		fmt.Fprint(flag.CommandLine.Output(), extraHelp)
		flag.PrintDefaults()
	}

	var (
		major, minor, patch, release bool

		buildmetadata string
	)

	flag.BoolVar(&major, "major", false, "Bump to next major version")
	flag.BoolVar(&minor, "minor", false, "Bump to next minor version")
	flag.BoolVar(&patch, "patch", false, "Bump to next patch version")
	flag.BoolVar(&release, "release", false, "Strip prerelease from version")

	flag.StringVar(&buildmetadata, "meta", "", "Optional build metadata attached to new version. Can be used multiple times.")

	flag.Parse()
	versions := flag.Args()

	if len(versions) != 1 {
		fmt.Fprintf(flag.CommandLine.Output(), "Only one argument expected: version")
		os.Exit(1)
	}

	version := versions[0]

	parsedVersion, err := semver.Parse(version)
	if err != nil {
		fmt.Printf("Invalid version: '%s', %s\n", version, err)
		os.Exit(1)
	}

	var newVersion semver.Version
	var meta []string

	if buildmetadata != "" {
		meta = strings.Split(buildmetadata, ".")
	}

	var opts []semver.BumpOption

	if len(meta) > 0 {
		opts = append(opts, semver.BumpOptionBuildmetadata(meta))
	}

	switch {
	case major:
		opts = append(opts, semver.BreakingChange)
	case minor:
		opts = append(opts, semver.FeatureChange)
	case patch:
		opts = append(opts, semver.ImplementationChange)
	case release:
		opts = append(opts, semver.Release())
	default:
		opts = append(opts, semver.ImplementationChange)
	}

	newVersion, err = parsedVersion.Bump(opts...)
	if err != nil {
		fmt.Printf("Bump '%s' failed: %s\n", version, err)
		os.Exit(1)
	}

	fmt.Println(newVersion.String())

	os.Exit(0)
}