package changelog_test

import (
	"strings"
	"testing"

	"openmars/internal/changelog"
	"openmars/test/testutil"
)

func imageURL(file string) string {
	return "https://example.test/" + file
}

func TestReleaseNotes_PlayerNotesFirstThenDeveloperDetails(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "v3", changelog.Player.FileName, "## Major update: Colonies\n\nColonies is now playable.\n\n![Trading with a colony](colonies.png)\n\n- Trade.\n")
	writeFile(t, dir, "v3", "colonies.png", "png")
	writeEntry(t, dir, "v3", "Highlights since v2.\n\n## Added\n- Colony system.\n")

	notes, err := changelog.ReleaseNotes(dir, "v3", imageURL)
	testutil.AssertNoError(t, err, "release notes")

	want := `# v3

## Major update: Colonies

Colonies is now playable.

![Trading with a colony](https://example.test/colonies.png)

- Trade.

## Developer details

Highlights since v2.

### Added
- Colony system.
`
	testutil.AssertEqual(t, want, notes, "release notes")
}

func TestReleaseNotes_WithoutPlayerNotes(t *testing.T) {
	dir := t.TempDir()
	writeEntry(t, dir, "v7.1.1", "Not published.\n")

	notes, err := changelog.ReleaseNotes(dir, "v7.1.1", imageURL)
	testutil.AssertNoError(t, err, "release notes")
	testutil.AssertEqual(t, "# v7.1.1\n\n## Developer details\n\nNot published.\n", notes, "release notes")
}

func TestReleaseNotes_RejectsInvalidNotes(t *testing.T) {
	dir := t.TempDir()
	writeEntry(t, dir, "v1", "## Added\n- A.\n")
	writeFile(t, dir, "v1", changelog.Player.FileName, "## Added\n- A.\n")

	_, err := changelog.ReleaseNotes(dir, "v1", imageURL)
	testutil.AssertError(t, err, "invalid player notes")
	testutil.AssertTrue(t, strings.Contains(err.Error(), changelog.Player.FileName), "error names the player file")
}
