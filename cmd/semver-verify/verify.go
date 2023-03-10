package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/adamwasila/go-semver"
)

var extraHelp string = "\n" +
	"  Validate range of versions given in argument list.\n" +
	"\n\n"

func main() {
	flag.CommandLine.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [VERSIONS]...\n", os.Args[0])
		fmt.Fprint(flag.CommandLine.Output(), extraHelp)
		flag.PrintDefaults()
	}

	flag.Parse()
	versions := flag.Args()
	wasInvalid := false

	for _, version := range versions {
		_, err := semver.Parse(version)
		if err != nil {
			fmt.Printf("Invalid version: '%s', %s\n", version, err)
			wasInvalid = true
		}
	}
	if wasInvalid {
		os.Exit(1)
	}
	os.Exit(0)
}
