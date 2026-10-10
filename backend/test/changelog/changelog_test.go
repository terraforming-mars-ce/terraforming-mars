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
	entry, err := changelog.Parse(changelog.Developer, "v1.1.2", []byte(data))
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
	entry, err := changelog.Parse(changelog.Developer, "v7.0.1", []byte("Not published.\n"))
	testutil.AssertNoError(t, err, "parse")
	testutil.AssertEqual(t, "Not published.", entry.Intro, "intro")
	testutil.AssertEqual(t, 0, len(entry.Sections), "sections")
}

func TestParse_SectionsOnly(t *testing.T) {
	entry, err := changelog.Parse(changelog.Developer, "v2", []byte("## Build\n- Faster builds.\n"))
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
		{"major update in developer notes", "v1.0.0", "## Major update: Colonies\nIntro.\n", `unknown section "Major update: Colonies"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := changelog.Parse(changelog.Developer, tt.version, []byte(tt.data))
			testutil.AssertErrorContains(t, err, tt.wantErr, "parse")
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

	entries, err := changelog.LoadAll(changelog.Developer, dir)
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

	_, err := changelog.LoadAll(changelog.Developer, dir)
	testutil.AssertErrorContains(t, err, `invalid changelog folder "next"`, "load all")
}

func TestLoadAll_ReportsFileAndLine(t *testing.T) {
	dir := t.TempDir()
	writeEntry(t, dir, "v1", "## Added\nNot a bullet.\n")

	_, err := changelog.LoadAll(changelog.Developer, dir)
	testutil.AssertErrorContains(t, err, filepath.Join("v1", changelog.Developer.FileName), "error names the file")
	testutil.AssertErrorContains(t, err, "line 2", "error names the line")
}

func TestLoad_MissingEntry(t *testing.T) {
	_, err := changelog.Load(changelog.Developer, t.TempDir(), "v9.9.9")
	testutil.AssertErrorContains(t, err, "failed to read CHANGELOG.md for v9.9.9", "load missing")
}

func TestParse_PlayerMajorUpdate(t *testing.T) {
	data := `## Major update: Multiple servers

Pick between servers like EU 1 and EU 2. Your choice
is remembered.

- Game links take friends to the right server.
- Switch anytime with Change server in the
  menu.

## Fixed
- A bug.
`
	entry, err := changelog.Parse(changelog.Player, "v7", []byte(data))
	testutil.AssertNoError(t, err, "parse")

	testutil.AssertEqual(t, 2, len(entry.Sections), "section count")
	major := entry.Sections[0]
	testutil.AssertEqual(t, "Major update: Multiple servers", major.Title, "major title")
	testutil.AssertTrue(t, major.Major, "major flag")
	testutil.AssertEqual(t, "Pick between servers like EU 1 and EU 2. Your choice is remembered.", major.Intro, "major intro")
	testutil.AssertEqual(t, "Switch anytime with Change server in the menu.", major.Items[1], "wrapped bullet")
	testutil.AssertFalse(t, entry.Sections[1].Major, "fixed is not major")
}

func TestParse_PlayerMajorUpdateImage(t *testing.T) {
	data := `## Major update: Colonies

Colonies is now playable.

![Trading with a colony](colonies.png)

- Build colonies and trade with them.
`
	entry, err := changelog.Parse(changelog.Player, "v3", []byte(data))
	testutil.AssertNoError(t, err, "parse")

	image := entry.Sections[0].Image
	testutil.AssertTrue(t, image != nil, "image parsed")
	testutil.AssertEqual(t, "colonies.png", image.File, "image file")
	testutil.AssertEqual(t, "Trading with a colony", image.Alt, "image alt")
	testutil.AssertEqual(t, "Colonies is now playable.", entry.Sections[0].Intro, "intro unchanged")
	testutil.AssertEqual(t, 1, len(entry.Sections[0].Items), "bullets")
}

func TestLoad_PlayerImageMustExist(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "v3", changelog.Player.FileName, "## Major update: Colonies\nIntro.\n![Colonies](colonies.png)\n")

	_, err := changelog.Load(changelog.Player, dir, "v3")
	testutil.AssertErrorContains(t, err, "image colonies.png", "missing image")

	writeFile(t, dir, "v3", "colonies.png", "png")
	entry, err := changelog.Load(changelog.Player, dir, "v3")
	testutil.AssertNoError(t, err, "image present")
	testutil.AssertEqual(t, "colonies.png", entry.Sections[0].Image.File, "image file")
}

func TestParse_PlayerIntroOnly(t *testing.T) {
	entry, err := changelog.Parse(changelog.Player, "v7.1.0", []byte("Behind-the-scenes improvements.\n"))
	testutil.AssertNoError(t, err, "parse")
	testutil.AssertEqual(t, "Behind-the-scenes improvements.", entry.Intro, "intro")
}

func TestParse_PlayerRejects(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr string
	}{
		{"developer section", "## Added\n- A.\n", `unknown section "Added"`},
		{"major after regular section", "## New\n- A.\n\n## Major update: Colonies\nIntro.\n", "must come before"},
		{"major without intro", "## Major update: Colonies\n- A.\n", "needs an intro paragraph"},
		{"major without name", "## Major update: \nIntro.\n", "needs a name"},
		{"major intro with two paragraphs", "## Major update: Colonies\nOne.\n\nTwo.\n", "single paragraph"},
		{"text after major bullets", "## Major update: Colonies\nIntro.\n- A.\nMore.\n", "expected a \"- \" bullet"},
		{"intro text in regular section", "## New\nIntro.\n- A.\n", "expected a \"- \" bullet"},
		{"image in regular section", "## New\n![A](a.png)\n- A.\n", "only allowed in major update sections"},
		{"image before intro", "## Major update: Colonies\n![A](a.png)\nIntro.\n", "between its intro and its bullets"},
		{"image after bullets", "## Major update: Colonies\nIntro.\n- A.\n![A](a.png)\n", "between its intro and its bullets"},
		{"two images", "## Major update: Colonies\nIntro.\n![A](a.png)\n![B](b.png)\n", "between its intro and its bullets"},
		{"image without alt", "## Major update: Colonies\nIntro.\n![](a.png)\n", "needs alt text"},
		{"image outside the folder", "## Major update: Colonies\nIntro.\n![A](../a.png)\n", "must be a file in the version folder"},
		{"image with other extension", "## Major update: Colonies\nIntro.\n![A](a.gif)\n", "must be a file in the version folder"},
		{"text after image", "## Major update: Colonies\nIntro.\n![A](a.png)\nMore.\n", "expected a \"- \" bullet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := changelog.Parse(changelog.Player, "v1.0.0", []byte(tt.data))
			testutil.AssertErrorContains(t, err, tt.wantErr, "parse")
		})
	}
}

func TestLoadAll_PlayerSkipsVersionsWithoutPlayerNotes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "v7", changelog.Player.FileName, "Shipped.\n")
	writeEntry(t, dir, "v7.0.1", "Never shipped.\n")

	entries, err := changelog.LoadAll(changelog.Player, dir)
	testutil.AssertNoError(t, err, "load all")
	testutil.AssertEqual(t, 1, len(entries), "entry count")
	testutil.AssertEqual(t, "v7", entries[0].Version, "version")
}

func TestLoad_PlayerNotesRequiredForATag(t *testing.T) {
	dir := t.TempDir()
	writeEntry(t, dir, "v7.0.1", "Developer notes only.\n")

	_, err := changelog.Load(changelog.Player, dir, "v7.0.1")
	testutil.AssertErrorContains(t, err, "failed to read CHANGELOG-USER.md for v7.0.1", "load player notes")
}

// Every committed changelog must parse, so a malformed file fails CI before it can block a release
func TestRepoChangelog_AllEntriesParse(t *testing.T) {
	developer, err := changelog.LoadAll(changelog.Developer, repoChangelogDir)
	testutil.AssertNoError(t, err, "load developer changelog")
	testutil.AssertTrue(t, len(developer) > 0, "developer changelog has entries")

	player, err := changelog.LoadAll(changelog.Player, repoChangelogDir)
	testutil.AssertNoError(t, err, "load player changelog")
	testutil.AssertTrue(t, len(player) > 0, "player changelog has entries")
}

func writeEntry(t *testing.T, dir, version, content string) {
	t.Helper()
	writeFile(t, dir, version, changelog.Developer.FileName, content)
}

func writeFile(t *testing.T, dir, version, fileName, content string) {
	t.Helper()
	folder := filepath.Join(dir, version)
	testutil.AssertNoError(t, os.MkdirAll(folder, 0o755), "mkdir")
	testutil.AssertNoError(t, os.WriteFile(filepath.Join(folder, fileName), []byte(content), 0o644), "write")
}
