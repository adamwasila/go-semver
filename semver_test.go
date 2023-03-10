package semver_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/adamwasila/go-semver"
)

func TestExamplesFromWebpage(t *testing.T) {
	validVersions := []string{
		"1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay",
		"1.0.0-rc.1+build.1",
		"2.0.0-rc.1+build.123",
		"1.2.3-beta",
		"10.2.3-DEV-SNAPSHOT",
		"1.2.3-SNAPSHOT-123",
		"1.0.0",
		"2.0.0",
		"1.1.7",
		"2.0.0+build.1848",
		"2.0.1-alpha.1227",
		"1.0.0-alpha+beta",
		"1.2.3----RC-SNAPSHOT.12.9.1--.12+788",
		"1.2.3----R-S.12.9.1--.12+meta",
		"1.2.3----RC-SNAPSHOT.12.9.1--.12",
		"1.0.0+0.build.1-rc.10000aaa-kk-0.1",
		"99999999999999999999999.999999999999999999.99999999999999999",
		"1.0.0-0A.is.legal",
	}
	for i, version := range validVersions {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			valid := semver.Valid(version)
			if !valid {
				t.Fatalf("expected successful parsing of a version: %s", version)
				t.FailNow()
			}

			v, err := semver.Parse(version)
			if err != nil {
				t.Fatalf("version '%s' should be valid but got '%s' instead", version, err)
				t.FailNow()
			}
			if v.String() != version {
				t.Fatalf("version '%s' and parsed then serialized '%s' should be equal", version, v.String())
				t.FailNow()
			}
		})
	}
}

func TestExamplesFromWebpageInvalid(t *testing.T) {
	invalidVersions := []string{
		"1",
		"1.2",
		"1.2.3-0123",
		"1.2.3-0123.0123",
		"1.1.2+.123",
		"+invalid",
		"-invalid",
		"-invalid+invalid",
		"-invalid.01",
		"alpha",
		"alpha.beta",
		"alpha.beta.1",
		"alpha.1",
		"alpha+beta",
		"alpha_beta",
		"alpha.",
		"alpha..",
		"beta",
		"1.0.0-alpha_beta",
		"-alpha.",
		"1.0.0-alpha..",
		"1.0.0-alpha..1",
		"1.0.0-alpha...1",
		"1.0.0-alpha....1",
		"1.0.0-alpha.....1",
		"1.0.0-alpha......1",
		"1.0.0-alpha.......1",
		"01.1.1",
		"1.01.1",
		"1.1.01",
		"1.2",
		"1.2.3.DEV",
		"1.2-SNAPSHOT",
		"1.2.31.2.3----RC-SNAPSHOT.12.09.1--..12+788",
		"1.2-RC-SNAPSHOT",
		"-1.0.3-gamma+b7718",
		"+justmeta",
		"9.8.7+meta+meta",
		"9.8.7-whatever+meta+meta",
		"99999999999999999999999.999999999999999999.99999999999999999----RC-SNAPSHOT.12.09.1--------------------------------..12",
	}
	for _, version := range invalidVersions {
		t.Run(version, func(t *testing.T) {
			valid := semver.Valid(version)
			if valid {
				t.Fatalf("expected to fail parsing as a version: %s", version)
				t.FailNow()
			}

			_, err := semver.Parse(version)
			if err == nil {
				t.Fatalf("version '%s' should be invalid", version)
				t.FailNow()
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    []func(*semver.Version) error
		want    string
		wantErr bool
	}{
		{
			"1.2.3",
			[]func(*semver.Version) error{
				semver.SetCore("1.2.3"),
			},
			"1.2.3",
			false,
		},
		{
			"0.0.0-alpha.1",
			[]func(*semver.Version) error{
				semver.Prerelease("alpha.1"),
			},
			"0.0.0-alpha.1",
			false,
		},
		{
			"0.0.0+buildxyz",
			[]func(*semver.Version) error{
				semver.BuildMetadata("buildxyz"),
			},
			"0.0.0+buildxyz",
			false,
		},
		{
			"0.0.0-beta.2+buildxyz",
			[]func(*semver.Version) error{
				semver.Prerelease("beta.2"),
				semver.BuildMetadata("buildxyz"),
			},
			"0.0.0-beta.2+buildxyz",
			false,
		},
		{
			"1.2.3-4.5.6+7.8.9", []func(*semver.Version) error{
				semver.SetCore("1.2.3"),
				semver.Prerelease("4"),
				semver.Prerelease("5"),
				semver.Prerelease("6"),
				semver.BuildMetadata("7"),
				semver.BuildMetadata("8"),
				semver.BuildMetadata("9"),
			}, "1.2.3-4.5.6+7.8.9", false,
		},
		{
			"1.2.3-4.5.6+7.8.9 (2)",
			[]func(*semver.Version) error{
				semver.SetCore("1.2.3"),
				semver.Prerelease("4.5"),
				semver.Prerelease("6"),
				semver.BuildMetadata("7.8"),
				semver.BuildMetadata("9"),
			},
			"1.2.3-4.5.6+7.8.9",
			false,
		},
		{
			"1.2.3-4.5.6+7.8.9 (2)",
			[]func(*semver.Version) error{
				semver.SetCore("1.2.3"),
				semver.Prerelease("4.5"),
				semver.Prerelease("6"),
				semver.BuildMetadata("7.8"),
				semver.BuildMetadata("9"),
			},
			"1.2.3-4.5.6+7.8.9",
			false,
		},
		{
			"Invalid prerelease",
			[]func(*semver.Version) error{
				semver.SetCore("1.2.3"),
				semver.Prerelease("1_2"),
			},
			"",
			true,
		},
		{
			"Invalid prerelease",
			[]func(*semver.Version) error{
				semver.SetCore("1.2.3"),
				semver.Prerelease("1.02"),
			},
			"",
			true,
		},
		{
			"Invalid buildmetadata",
			[]func(*semver.Version) error{
				semver.SetCore("1.2.3"),
				semver.BuildMetadata("a_b"),
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := semver.New(tt.opts...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error constructing version but there was none")
					t.FailNow()
				}
			} else {
				if err != nil {
					t.Errorf("expected no error constructing version but there was one: %v", err)
					t.FailNow()
				}

				if tt.want != got.String() {
					t.Errorf("version built: %s is different than expected: %s", got.String(), tt.want)
					t.FailNow()
				}
			}
		})
	}
}

func TestSortSemver(t *testing.T) {
	data := []struct {
		name           string
		sortedVersions []string
		allEquals      bool
	}{
		// Two examples from specification (11.2 and 11.4): https://semver.org/#spec-item-11
		{"sorting examples from semver spec 11.2",
			[]string{"1.0.0", "2.0.0", "2.1.0", "2.1.1"},
			false,
		},
		{"sorting examples from semver spec 11.4",
			[]string{"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta", "1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1", "1.0.0"},
			false,
		},
		// sanity check
		{"sorting with same version twice",
			[]string{"1.0.0-a.b.c.1.2.3+anything", "1.0.0-a.b.c.1.2.3+anything"},
			true,
		},
		// test if core components are compared as numbers not strings
		{"sorting core components - major",
			[]string{"1.0.0", "2.0.0", "10.0.0"},
			false,
		},
		{"sorting core components - minor",
			[]string{"2.1.0", "2.2.0", "2.10.0"},
			false,
		},
		{"sorting core components - patch",
			[]string{"3.6.1", "3.6.2", "3.6.10"},
			false,
		},
		// build metadata should be ignored
		{"sorting with build metadata - equals",
			[]string{"1.0.0+B", "1.0.0+A", "1.0.0+C", "1.0.0"},
			true,
		},
		{"sorting with build metadata",
			[]string{"1.0.0+D", "2.0.0+A", "3.0.0+B", "4.0.0"},
			false,
		},
	}

	for _, tt := range data {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < len(tt.sortedVersions); i++ {
				for j := i + 1; j < len(tt.sortedVersions); j++ {
					asc := less(tt.sortedVersions[i], tt.sortedVersions[j])
					desc := less(tt.sortedVersions[j], tt.sortedVersions[i])
					if tt.allEquals {
						if asc {
							t.Errorf("Should be: %s == %s", tt.sortedVersions[i], tt.sortedVersions[j])
							t.Fail()
						}
						if desc {
							t.Errorf("Should be: %s == %s", tt.sortedVersions[i], tt.sortedVersions[j])
							t.Fail()
						}
						break
					}
					if !asc {
						t.Errorf("Should be: %s < %s", tt.sortedVersions[i], tt.sortedVersions[j])
						t.Fail()
					}
					if desc {
						t.Errorf("Should be: %s >= %s", tt.sortedVersions[j], tt.sortedVersions[i])
						t.Fail()
					}
				}
			}
		})
	}
}

func less(v1, v2 string) bool {
	s1 := semver.MustParse(v1)
	s2 := semver.MustParse(v2)
	return semver.Less(&s1, &s2)
}

func BenchmarkParse(b *testing.B) {
	v := "1.2.3-alpha.1+build.7d97e98f8af710c7e7fe703abc8f639e0ee507c4"
	for i := 0; i < b.N; i++ {
		_, _ = semver.Parse(v)
	}
}

func BenchmarkValid(b *testing.B) {
	v := "7.8.9-beta.1+build.7d97e98f8af710c7e7fe703abc8f639e0ee507c4"
	for i := 0; i < b.N; i++ {
		_ = semver.Valid(v)
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = semver.New(
			semver.SetCore("1.2.3"),
			semver.Prerelease("alpha"),
			semver.Prerelease("1"),
			semver.BuildMetadata("build"),
			semver.BuildMetadata("7d97e98f8af710c7e7fe703abc8f639e0ee507c4"),
		)
	}
}

func ExampleNew() {
	sv, _ := semver.New(
		semver.SetCore("1.2.3"),
		semver.Prerelease("rc.1"), semver.BuildMetadata("cafebabe"),
	)
	fmt.Printf("%s", sv.String())
	// Output:
	// 1.2.3-rc.1+cafebabe
}

func ExampleMustParse() {
	v := "1.2.3-ver.12a+build.1234"
	sv := semver.MustParse(v)
	fmt.Printf("%s == %s", v, sv.String())
	// Output:
	// 1.2.3-ver.12a+build.1234 == 1.2.3-ver.12a+build.1234
}

func ExampleVersion_Bump() {
	sv := semver.MustParse("1.2.3-rc.1+cafebabe")
	sv, _ = sv.Bump(semver.BreakingChange)
	fmt.Printf("%s", sv.String())
	// Output:
	// 2.0.0
}

func ExampleLess() {
	versions := []string{
		"1.0.0+first",
		"1.0.0+second",
		"1.0.0+3rd",
		"1.0.0-rc.1",
		"1.0.0-beta.11",
		"1.0.0-beta.2",
		"1.0.0-beta",
		"1.0.0-alpha.beta",
		"1.0.0-alpha.1",
		"1.0.0-alpha",
	}

	sort.SliceStable(versions, func(i, j int) bool {
		v1 := semver.MustParse(versions[i])
		v2 := semver.MustParse(versions[j])
		return semver.Less(&v1, &v2)
	})

	for i, v := range versions {
		fmt.Printf("%d. %v\n", i, v)
	}
	// Output:
	// 0. 1.0.0-alpha
	// 1. 1.0.0-alpha.1
	// 2. 1.0.0-alpha.beta
	// 3. 1.0.0-beta
	// 4. 1.0.0-beta.2
	// 5. 1.0.0-beta.11
	// 6. 1.0.0-rc.1
	// 7. 1.0.0+first
	// 8. 1.0.0+second
	// 9. 1.0.0+3rd
}

func TestVersion_Bump(t *testing.T) {
	type opts = []semver.BumpOption

	type args struct {
		baseVersion string
		options     []semver.BumpOption
	}
	type result struct {
		expectedVersion string
		expectError     bool
	}
	data := []struct {
		name   string
		args   args
		result result
	}{
		{"by default version bumps from major version to next major version",
			args{baseVersion: "1.0.0"},
			result{expectedVersion: "2.0.0"},
		},
		{"by default version bumps from minor version to next major version",
			args{baseVersion: "1.2.0"},
			result{expectedVersion: "2.0.0"},
		},
		{"by default version bumps from patched version to next major version",
			args{baseVersion: "1.2.3"},
			result{expectedVersion: "2.0.0"},
		},
		{"with explicit option breaking change version bumps to next major version",
			args{baseVersion: "1.2.3", options: opts{semver.BreakingChange}},
			result{expectedVersion: "2.0.0"},
		},
		{"with explicit option feature add version bumps to next minor version",
			args{baseVersion: "1.2.3", options: opts{semver.FeatureChange}},
			result{expectedVersion: "1.3.0"},
		},
		{"with explicit option withbugfix/implementation change version bumps to next patch version",
			args{baseVersion: "1.2.3", options: opts{semver.ImplementationChange}},
			result{expectedVersion: "1.2.4"},
		},

		// TODO
		// {"by default prerelease version bumps to next prerelease version if last component is number",
		// 	args{baseVersion: "1.2.3-rc.1"}, result{expectedVersion: "1.2.3-rc.2"},
		// },
		// {"by default prerelease version bumps to next prerelease version but only last component",
		// 	args{baseVersion: "1.2.3-alpha.1.72"}, result{expectedVersion: "1.2.3-alpha.1.73"},
		// },
		{"by default version bumps and clears buildmetadata",
			args{baseVersion: "4.3.2+hello"},
			result{expectedVersion: "5.0.0"},
		},
		{"with explicit option bumps and sets buildmetadata",
			args{baseVersion: "1.0.0+test", options: opts{semver.BreakingChange, semver.BuildMetadata("other")}},
			result{expectedVersion: "2.0.0+other"},
		},
	}

	for _, tt := range data {
		t.Run(tt.name, func(t *testing.T) {
			sv := semver.MustParse(tt.args.baseVersion)

			resultSv, err := sv.Bump(tt.args.options...)

			if (err != nil) != tt.result.expectError {
				t.Errorf("bumping vesion returned error: %s which is unexpected", err)
				t.FailNow()
			}

			result := semver.MustParse(tt.result.expectedVersion)

			if result.String() != resultSv.String() {
				t.Errorf("bumped version: %s is different than expected: %s", resultSv, result)
				t.FailNow()
			}
		})
	}
}

func TestVersion_Valid(t *testing.T) {
	tests := []struct {
		name    string
		mutator func(*semver.Version)
		valid   bool
	}{
		{
			name:    "1. correct version",
			mutator: func(v *semver.Version) {},
			valid:   true,
		},
		{
			name:    "2a. invalid major version: not a number",
			mutator: func(v *semver.Version) { v.Major = "abc" },
		},
		{
			name:    "2b. invalid major version: leading zeros",
			mutator: func(v *semver.Version) { v.Major = "01" },
		},
		{
			name:    "3a. invalid minor version: not a number",
			mutator: func(v *semver.Version) { v.Minor = "xyz" },
		},
		{
			name:    "3b. invalid minor version: not a number",
			mutator: func(v *semver.Version) { v.Minor = "007" },
		},
		{
			name:    "4a. invalid patch version: not a number",
			mutator: func(v *semver.Version) { v.Patch = "hello_world" },
		},
		{
			name:    "4b. invalid patch version: not a number",
			mutator: func(v *semver.Version) { v.Patch = "00" },
		},
		{
			name:    "5a. invalid prerelease: character outside allowed alphabet",
			mutator: func(v *semver.Version) { v.Prerelease = []string{"invalid_alphabet!"} },
		},
		{
			name:    "5b. invalid prerelease: empty component",
			mutator: func(v *semver.Version) { v.Prerelease = []string{"", "rc", "1"} },
		},
		{
			name:    "5c. invalid prerelease: only empty component",
			mutator: func(v *semver.Version) { v.Prerelease = []string{""} },
		},
		{
			name:    "6a. invalid buildmetadata: character outside allowed alphabet",
			mutator: func(v *semver.Version) { v.Buildmetadata = []string{"invalid_alphabet!"} },
		},
		{
			name:    "6b. invalid buildmetadata: empty component",
			mutator: func(v *semver.Version) { v.Buildmetadata = []string{"", "abcdef"} },
		},
		{
			name:    "6c. invalid buildmetadata: only empty component",
			mutator: func(v *semver.Version) { v.Buildmetadata = []string{""} },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sv := semver.MustParse("1.2.3-rc.1+cafebabe")

			tt.mutator(&sv)

			if got := sv.Valid(); got != tt.valid {
				t.Errorf("Version.Valid() = %v, want %v", got, tt.valid)
			}
		})
	}
}
