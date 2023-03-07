package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/adamwasila/semver"
)

var extraHelp string = "\n" +
	"  Reads list of versions from standard input and returns sorted list of versions\n" +
	"  to standard output. Sorting uses rules defined by semver 2.0 specification." +
	"  See semver.org for details.\n" +
	"\n" +
	"  Expects versions to be separated with any number of unicode whitespaces but can be\n" +
	"  changed with separate flag. Further customization is possible with any flag\n" +
	"  documented below.\n" +
	"\n\n"

func main() {
	flag.CommandLine.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTION]...\n", os.Args[0])
		fmt.Fprint(flag.CommandLine.Output(), extraHelp)
		flag.PrintDefaults()
	}

	delim := flag.String("d", "", "delimiter used to separate input versions; default to 1+ of unicode whitespaces")
	origSep := flag.String("s", "\n", "versions delimiter used in output")
	noLn := flag.Bool("n", false, "do not output the trailing newline")
	onlyLast := flag.Bool("1", false, "return only last sorted version")
	reverse := flag.Bool("r", false, "return versions in reversed order meaning newest first")
	ignoreErr := flag.Bool("i", false, "skip versions that have invalid format")

	flag.Parse()

	sep, err := strconv.Unquote(`"` + *origSep + `"`)
	if err != nil {
		sep = *origSep
	}

	var vs versions

	var scanner interface {
		Scan() bool
		Text() string
	}

	if *delim != "" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf("error reading input: %v", err)
			os.Exit(1)
		}
		scanner = newSplitScanner(string(data), *delim)
	} else {
		bscan := bufio.NewScanner(os.Stdin)
		bscan.Split(bufio.ScanWords)
		scanner = bscan
	}

	for scanner.Scan() {
		v, err := semver.Parse(scanner.Text())
		if err != nil && !*ignoreErr {
			fmt.Printf("error parsing %s: %v", scanner.Text(), err)
			os.Exit(1)
		}
		if err != nil {
			continue
		}
		vs = append(vs, v)
	}

	if len(vs) == 0 {
		os.Exit(0)
	}

	sortVersions(vs, *reverse)

	if *onlyLast {
		vs = vs[len(vs)-1:]
	}

	fmt.Print(vs[0].String())
	for _, v := range vs[1:] {
		fmt.Print(sep, v.String())
	}
	if !*noLn {
		fmt.Println()
	}

	os.Exit(0)
}

type splitScanner struct {
	pos int
	str []string
}

func newSplitScanner(s, sep string) *splitScanner {
	return &splitScanner{
		pos: -1,
		str: strings.Split(s, sep),
	}
}

func (s *splitScanner) Scan() bool {
	s.pos++
	return s.pos < len(s.str)
}

func (s *splitScanner) Text() string {
	return strings.TrimSpace(s.str[s.pos])
}

type versions []semver.Version

func (v versions) Len() int {
	return len(v)
}

func (v versions) Less(i, j int) bool {
	return semver.Less(&v[i], &v[j])
}

func (v versions) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func sortVersions(data sort.Interface, reverse bool) {
	if reverse {
		data = sort.Reverse(data)
	}

	sort.Stable(data)
}
