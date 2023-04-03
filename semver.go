// Package semver provides methods to validate, parse, compare and modify semantic version compliant strings.
package semver

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type Version struct {
	Major         string
	Minor         string
	Patch         string
	Prerelease    []string
	Buildmetadata []string
}

// New creates new version struct instance
func New(options ...func(*Version) error) (*Version, error) {
	s := Version{
		Major: "0",
		Minor: "0",
		Patch: "0",
	}
	err := compose(options...)(&s)
	return &s, err
}

// Parse unpacks provided version string to predefined Version struct
func Parse(s string) (Version, error) {
	v := Version{}
	v.Prerelease = []string{}
	v.Buildmetadata = []string{}

	_, err := defaultParser(0, s, &v)
	if err != nil {
		return Version{}, err
	}

	return v, nil
}

// MustParse behaves like Parse but in case validation fails simply panics instead of returning an error
func MustParse(semver string) Version {
	s, err := Parse(semver)
	if err != nil {
		panic(err)
	}
	return s
}

// SetCore returns option to set version core: dot separated major, minor and patch number.
// Numbers must not have leading zeros.
func SetCore(core string) func(*Version) error {
	return func(s *Version) error {
		if core == "" {
			return nil
		}
		parser := sequence(
			major(),
			dot(),
			minor(),
			dot(),
			patch(),
			excess(),
		)

		_, err := parser(0, core, s)
		return err
	}
}

// Prerelease return option to set prerelease component. Empty string is silently ignored returning no error.
// Calling this option will not clear prerelease component so it may be used more than once and final result
// will be sum of all requests
//
// With current spec valid request string should be:
//
// * dot separated, nonempty components of [a-zA-Z0-9-] alphabet
// * if only numbers are used for particular component leading zeroes are forbidden
//
// Valid examples:
//
// `a.b.c`, `rc.1`, `rc01`, `alpha.beta.gamma` etc.
func Prerelease(pr string) func(*Version) error {
	return func(s *Version) error {
		if pr == "" {
			return nil
		}
		parser := sequence(prerelease(), repeat(dot(), prerelease()), excess())
		_, err := parser(0, pr, s)
		return err
	}
}

// BuildMetadata return option to set buildmetadata component. Empty string is silently ignored returning no error.
//
// In current spec buildmetadata must consist of:
//
// * dot separated, nonempty components of [a-zA-Z0-9-] alphabet.
func BuildMetadata(bl string) func(*Version) error {
	return func(s *Version) error {
		if bl == "" {
			return nil
		}
		parser := sequence(buildmetadata(), repeat(dot(), buildmetadata()), excess())
		_, err := parser(0, bl, s)
		return err
	}
}

func compose(opts ...func(*Version) error) func(*Version) error {
	return func(v *Version) error {
		for _, o := range opts {
			err := o(v)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// BumpOption is function option that changes version to newer
type BumpOption func(*Version) error

func increment(n string) string {
	const baseDec = 10
	bigN, ok := big.NewInt(0).SetString(n, baseDec)
	if !ok {
		panic("incrementing corrupted version number")
	}
	bigN.Add(bigN, big.NewInt(1))
	return bigN.String()
}

// BumpMajor is major version increment
func BumpMajor() BumpOption {
	return func(s *Version) error {
		s.Major = increment(s.Major)
		s.Minor = "0"
		s.Patch = "0"
		s.Prerelease = []string{}
		return nil
	}
}

// BumpMinor is minor version increment
func BumpMinor() BumpOption {
	return func(s *Version) error {
		s.Minor = increment(s.Minor)
		s.Patch = "0"
		s.Prerelease = []string{}
		return nil
	}
}

// BumpPatch is patch version increment
func BumpPatch() BumpOption {
	return func(s *Version) error {
		s.Patch = increment(s.Patch)
		s.Prerelease = []string{}
		return nil
	}
}

// BumpRelease is increment to next non-prerelease version
func BumpRelease() BumpOption {
	return func(v *Version) error {
		if len(v.Prerelease) == 0 {
			return fmt.Errorf("no prerelease set in version")
		}
		v.Prerelease = []string{}
		return nil
	}
}

// BumpPrerelease is incrementing last numeric prerelease component
func BumpPrelease() BumpOption {
	return func(v *Version) error {
		if len(v.Prerelease) == 0 {
			return fmt.Errorf("no prerelease set in version")
		}
		for i := len(v.Prerelease) - 1; i >= 0; i-- {
			const baseDec = 10
			bigN, ok := big.NewInt(0).SetString(v.Prerelease[i], baseDec)
			one := big.NewInt(1)
			if ok {
				bigN.Add(bigN, one)
				v.Prerelease[i] = bigN.String()
				break
			}
		}
		return nil
	}
}

// Bump changes version to newer using provided list of bump options
func (semver *Version) Bump(options ...BumpOption) (Version, error) {
	if len(options) == 0 {
		options = []BumpOption{
			BumpMajor(),
			BuildMetadata(""),
		}
	}

	newSemver := *semver
	newSemver.Buildmetadata = []string{}

	var err error
	for _, option := range options {
		err = option(&newSemver)
		if err != nil {
			return Version{}, err
		}
	}
	return newSemver, nil
}

// MustBump works like Bump but panics in cases where Bump returns an error
func (semver *Version) MustBump(options ...BumpOption) Version {
	nv, err := semver.Bump(options...)
	if err != nil {
		panic(err)
	}
	return nv
}

// Valid performs regular parse and returns whetever it was successful or not bug discarding actual result
func Valid(semver string) bool {
	_, err := Parse(semver)
	return err == nil
}

// Valid checks if version struct follows semver rules. Instances created by New must always be valid.
// Purpose of this method is to check directly created or modified structs.
func (semver *Version) Valid() bool {
	_, err := Parse(semver.String())
	return err == nil
}

// Less perform comparison of to specified versions. It strictly follows rules
// of semver specification, paragraph 11.: https://semver.org/#spec-item-11
func Less(s1, s2 *Version) bool {
	// 1.0.0 < 2.0.0
	if less, eq := lessOrEqual(s1.Major, s2.Major); !eq {
		return less
	}
	// 0.1.0 < 0.2.0
	if less, eq := lessOrEqual(s1.Minor, s2.Minor); !eq {
		return less
	}
	// 0.0.1 < 0.0.2
	if less, eq := lessOrEqual(s1.Patch, s2.Patch); !eq {
		return less
	}
	// 1.0.0-alpha.1 < 1.0.0
	// 1.0.0-alpha < 1.0.0-alpha.1
	if less, eq := lessOrEqualStrings(s1.Prerelease, s2.Prerelease); !eq {
		return less
	}
	// note: at this point false means both versions are equal
	return false
}

func lessOrEqual(a, b string) (less, eq bool) {
	if len(a) > len(b) {
		return false, false
	}
	if len(a) < len(b) {
		return true, false
	}
	return a < b, a == b
}

func lessOrEqualStrings(a, b []string) (less, eq bool) {
	if len(a) == 0 && len(b) == 0 {
		return false, true
	}
	if len(a) == 0 {
		return false, false
	}
	if len(b) == 0 {
		return true, false
	}
	n := min(len(a), len(b))
	for i := 0; i < n; i++ {
		aIsNum, aVal := isNum(a[i])
		bIsNum, bVal := isNum(b[i])
		if aIsNum && !bIsNum {
			return true, false
		}
		if !aIsNum && bIsNum {
			return false, false
		}
		if aIsNum && bIsNum {
			return aVal < bVal, aVal == bVal
		}
		if a[i] < b[i] {
			return true, false
		}
		if a[i] > b[i] {
			return false, false
		}
	}
	if len(a) < len(b) {
		return true, false
	}
	if len(a) > len(b) {
		return false, false
	}
	return false, true
}

func isNum(s string) (num bool, val int) {
	val, err := strconv.Atoi(s)
	return err == nil, val
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// String returns semver compliant version string. It is efectively a reverse of Parse/MustParse functions.
func (semver *Version) String() string {
	version := fmt.Sprintf("%s.%s.%s", semver.Major, semver.Minor, semver.Patch)
	if len(semver.Prerelease) > 0 {
		version = fmt.Sprintf("%s-%s", version, strings.Join(semver.Prerelease, "."))
	}
	if len(semver.Buildmetadata) > 0 {
		version = fmt.Sprintf("%s+%s", version, strings.Join(semver.Buildmetadata, "."))
	}
	return version
}

var defaultParser = semverParser()

type consumer func(pos int, stream string, v *Version) (remain string, err error)

type positionError struct {
	pos int
	msg string
}

func positionErr(pos int, format string, a ...interface{}) error {
	return &positionError{
		pos: pos,
		msg: fmt.Sprintf(format, a...),
	}
}

// Error returns error with stream position where error has occurred
func (e *positionError) Error() string {
	return fmt.Sprintf("error at position %d: %s", e.pos, e.msg)
}

func semverParser() consumer {
	var parser = []consumer{
		major(),
		dot(),
		minor(),
		dot(),
		patch(),
		optional(minus(), prerelease(), repeat(dot(), prerelease())),

		optional(plus(), buildmetadata(), repeat(dot(), buildmetadata())),
		excess(),
	}
	return sequence(parser...)
}

func literal(val string) consumer {
	return func(pos int, stream string, v *Version) (remain string, err error) {
		if stream == "" {
			return stream, positionErr(pos, "unexpected end of stream while %s was expected", nameFor(val))
		}
		if strings.HasPrefix(stream, val) {
			return stream[len(val):], nil
		}
		return stream, positionErr(pos, "unexpected character in place where %s was expected", nameFor(val))
	}
}

func dot() consumer {
	return literal(".")
}

func plus() consumer {
	return literal("+")
}

func minus() consumer {
	return literal("-")
}

func nameFor(val string) string {
	switch val {
	case ".":
		return "dot"
	case "+":
		return "plus"
	case "-":
		return "minus"
	default:
		return val
	}
}

func number(f func(v *Version, num string)) consumer {
	return func(pos int, stream string, v *Version) (remain string, err error) {
		var num string
		for i, s := range stream {
			if s < '0' || s > '9' {
				break
			}
			num = stream[0 : i+1]
			remain = stream[i+1:]
		}
		if num == "" {
			return stream, positionErr(pos+1, "unexpected non-numeric character")
		}
		if len(num) > 1 && num[0] == '0' {
			return stream, positionErr(pos+1, "unexpected leading zero")
		}

		f(v, num)

		return remain, nil
	}
}

func major() consumer {
	return number(func(v *Version, n string) { v.Major = n })
}

func minor() consumer {
	return number(func(v *Version, n string) { v.Minor = n })
}

func patch() consumer {
	return number(func(v *Version, n string) { v.Patch = n })
}

func sequence(cs ...consumer) consumer {
	return func(pos int, stream string, v *Version) (remain string, err error) {
		for _, c := range cs {
			remain, err := c(pos, stream, v)
			if err != nil {
				return stream, err
			}
			pos += (len(stream) - len(remain))
			stream = remain
		}
		return stream, nil
	}
}

func optional(cs consumer, c ...consumer) consumer {
	return func(pos int, stream string, v *Version) (remain string, err error) {
		remain, err = cs(pos, stream, v)
		if err != nil {
			return stream, nil
		}

		pos += (len(stream) - len(remain))

		for _, consumer := range c {
			remain, err = consumer(pos, remain, v)
			if err != nil {
				return stream, err
			}
			pos += (len(stream) - len(remain))
		}
		return remain, err
	}
}

func repeat(c ...consumer) consumer {
	return func(pos int, stream string, v *Version) (remain string, err error) {
		remain = stream
		for i := 0; ; i++ {
			oldRemainLen := len(remain)
			remain, err = c[i%len(c)](pos, remain, v)
			if err != nil {
				if i%len(c) == 0 {
					break
				}
				return stream, err
			}
			pos += (oldRemainLen - len(remain))
		}
		return remain, nil
	}
}

func prerelease() consumer {
	return func(pos int, stream string, v *Version) (remain string, err error) {
		if stream == "" {
			return "", positionErr(pos, "unexpected end of stream in prerelease")
		}
		remain = stream
		numberFlag := true
		var token string
		for i, s := range stream {
			if s == '.' || s == '+' {
				if i == 0 {
					return stream, positionErr(pos+i, "unexpected empty prerelease")
				}
				token = stream[:i]
				v.Prerelease = append(v.Prerelease, token)
				remain = stream[i:]
				break
			}
			if (s < '0' || s > '9') && (s < 'a' || s > 'z') && (s < 'A' || s > 'Z') && s != '-' {
				return stream, positionErr(pos+i-1, "invalid character in prerelease")
			}
			if s < '0' || s > '9' {
				numberFlag = false
			}
		}
		if token == "" {
			token = stream
			v.Prerelease = append(v.Prerelease, token)
			remain = ""
		}
		if token != "" && numberFlag {
			if len(token) > 1 && token[0] == '0' {
				return stream, positionErr(pos, "unexpected leading zero")
			}
		}
		return remain, err
	}
}

func buildmetadata() consumer {
	return func(pos int, stream string, v *Version) (remain string, err error) {
		if stream == "" {
			return "", positionErr(pos, "unexpected end of stream in buildmetadata")
		}
		for i, s := range stream {
			if s == '.' || s == '+' {
				if i == 0 {
					return stream, positionErr(pos+i, "unexpected empty prerelease")
				}
				v.Buildmetadata = append(v.Buildmetadata, stream[:i])
				return stream[i:], nil
			}
			if (s < '0' || s > '9') && (s < 'a' || s > 'z') && (s < 'A' || s > 'Z') && s != '-' {
				return stream, positionErr(pos+i, "invalid character in buildmetadata")
			}
		}
		v.Buildmetadata = append(v.Buildmetadata, stream)
		return "", nil
	}
}

func excess() consumer {
	return func(pos int, stream string, v *Version) (remain string, err error) {
		if stream != "" {
			return stream, positionErr(pos, "unexpected extra data")
		}
		return "", nil
	}
}
