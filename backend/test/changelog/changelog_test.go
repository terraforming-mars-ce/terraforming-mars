package changelog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"openmars/internal/changelog"
	"openmars/test/testutil"
)

const repoChangelogDir = "../../../changelog"

func TestParse_FullEntry(t *testing.T) {
	data := `Fixes since v1.1.0. Supersedes v1.1.1, whose images
never built.

## Added
- A new thing with ` + "`just check`" + `
  wrapped onto a second line.
- Another thing.

## Fixed
- A bug.
`
	entry, err := changelog.Parse("v1.1.2", []byte(data))
	testutil.AssertNoError(t, err, "parse")

	testutil.AssertEqual(t, "v1.1.2", entry.Version, "version")
	testutil.AssertEqual(t, "Fixes since v1.1.0. Supersedes v1.1.1, whose images never built.", entry.Intro, "intro")
	testutil.AssertEqual(t, 2, len(entry.Sections), "section count")
	testutil.AssertEqual(t, "Added", entry.Sections[0].Title, "first section")
	testutil.AssertEqual(t, 2, len(entry.Sections[0].Items), "added items")
	testutil.AssertEqual(t, "A new thing with `just check` wrapped onto a second line.", entry.Sections[0].Items[0], "wrapped bullet")
	testutil.AssertEqual(t, "Fixed", entry.Sections[1].Title, "second section")
	testutil.AssertEqual(t, "A bug.", entry.Sections[1].Items[0], "fixed bullet")
}

func TestParse_IntroOnly(t *testing.T) {
	entry, err := changelog.Parse("v7.0.1", []byte("Not published.\n"))
	testutil.AssertNoError(t, err, "parse")
	testutil.AssertEqual(t, "Not published.", entry.Intro, "intro")
	testutil.AssertEqual(t, 0, len(entry.Sections), "sections")
}

func TestParse_SectionsOnly(t *testing.T) {
	entry, err := changelog.Parse("v2", []byte("## Build\n- Faster builds.\n"))
	testutil.AssertNoError(t, err, "parse")
	testutil.AssertEqual(t, "", entry.Intro, "intro")
	testutil.AssertEqual(t, "Build", entry.Sections[0].Title, "section")
}

func TestParse_Rejects(t *testing.T) {
	tests := []struct {
		name    string
		version string
		data    string
		wantErr string
	}{
		{"empty file", "v1.0.0", "", "no intro and no sections"},
		{"blank file", "v1.0.0", "\n\n", "no intro and no sections"},
		{"invalid version", "1.0.0", "Intro.", "invalid version"},
		{"unknown section", "v1.0.0", "## Removed\n- Gone.\n", `unknown section "Removed"`},
		{"out of order", "v1.0.0", "## Fixed\n- A.\n\n## Added\n- B.\n", `"Added" is repeated or out of order`},
		{"duplicate", "v1.0.0", "## Added\n- A.\n\n## Added\n- B.\n", `"Added" is repeated or out of order`},
		{"empty section", "v1.0.0", "## Added\n\n## Fixed\n- A.\n", `section "Added" has no bullets`},
		{"empty last section", "v1.0.0", "## Added\n- A.\n\n## Fixed\n", `section "Fixed" has no bullets`},
		{"other heading level", "v1.0.0", "# Title\n", "only \"## <Section>\" headings"},
		{"stray text in section", "v1.0.0", "## Added\nNot a bullet.\n", "expected a \"- \" bullet"},
		{"continuation after blank line", "v1.0.0", "## Added\n- A.\n\n  more\n", "expected a \"- \" bullet"},
		{"empty bullet", "v1.0.0", "## Added\n- \n", "empty bullet"},
		{"two intro paragraphs", "v1.0.0", "One.\n\nTwo.\n", "single paragraph"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := changelog.Parse(tt.version, []byte(tt.data))
			testutil.AssertError(t, err, "parse")
			testutil.AssertTrue(t, strings.Contains(err.Error(), tt.wantErr), "error "+err.Error()+" mentions "+tt.wantErr)
		})
	}
}

func TestCompareVersions(t *testing.T) {
	ordered := []string{"v0.0.1", "v1", "v2", "v7", "v7.0.1", "v7.1", "v7.1.2", "v7.10.0", "v10"}
	for i := 1; i < len(ordered); i++ {
		testutil.AssertEqual(t, -1, changelog.CompareVersions(ordered[i-1], ordered[i]), ordered[i-1]+" < "+ordered[i])
		testutil.AssertEqual(t, 1, changelog.CompareVersions(ordered[i], ordered[i-1]), ordered[i]+" > "+ordered[i-1])
	}
	testutil.AssertEqual(t, 0, changelog.CompareVersions("v7", "v7.0.0"), "v7 == v7.0.0")
}

func TestIsVersion(t *testing.T) {
	for _, valid := range []string{"v1", "v7.1", "v7.1.2", "v0.0.1"} {
		testutil.AssertTrue(t, changelog.IsVersion(valid), valid)
	}
	for _, invalid := range []string{"7.1.2", "v7.1.2.3", "v7.1.2-rc1", "latest", "v"} {
		testutil.AssertFalse(t, changelog.IsVersion(invalid), invalid)
	}
}

func TestLoadAll_SortsNewestFirst(t *testing.T) {
	dir := t.TempDir()
	for _, version := range []string{"v7", "v7.1.0", "v1", "v7.0.1"} {
		writeEntry(t, dir, version, "Release "+version+".\n")
	}

	entries, err := changelog.LoadAll(dir)
	testutil.AssertNoError(t, err, "load all")

	var versions []string
	for _, entry := range entries {
		versions = append(versions, entry.Version)
	}
	testutil.AssertEqual(t, "v7.1.0,v7.0.1,v7,v1", strings.Join(versions, ","), "order")
}

func TestLoadAll_RejectsInvalidFolderName(t *testing.T) {
	dir := t.TempDir()
	writeEntry(t, dir, "next", "Unreleased.\n")

	_, err := changelog.LoadAll(dir)
	testutil.AssertError(t, err, "load all")
}

func TestLoadAll_ReportsFileAndLine(t *testing.T) {
	dir := t.TempDir()
	writeEntry(t, dir, "v1", "## Added\nNot a bullet.\n")

	_, err := changelog.LoadAll(dir)
	testutil.AssertError(t, err, "load all")
	testutil.AssertTrue(t, strings.Contains(err.Error(), filepath.Join("v1", changelog.FileName)), "error names the file")
	testutil.AssertTrue(t, strings.Contains(err.Error(), "line 2"), "error names the line")
}

func TestLoad_MissingEntry(t *testing.T) {
	_, err := changelog.Load(t.TempDir(), "v9.9.9")
	testutil.AssertError(t, err, "load missing")
}

// Every committed changelog must parse, so a malformed file fails CI before it can block a release
func TestRepoChangelog_AllEntriesParse(t *testing.T) {
	entries, err := changelog.LoadAll(repoChangelogDir)
	testutil.AssertNoError(t, err, "load repo changelog")
	testutil.AssertTrue(t, len(entries) > 0, "repo changelog has entries")
}

func writeEntry(t *testing.T, dir, version, content string) {
	t.Helper()
	folder := filepath.Join(dir, version)
	testutil.AssertNoError(t, os.MkdirAll(folder, 0o755), "mkdir")
	testutil.AssertNoError(t, os.WriteFile(filepath.Join(folder, changelog.FileName), []byte(content), 0o644), "write")
}
