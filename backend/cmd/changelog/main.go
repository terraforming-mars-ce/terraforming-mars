// Command changelog works with the release notes in changelog/<tag>/: the
// developer CHANGELOG.md and the player CHANGELOG-USER.md.
//
//	changelog check <changelog-dir> <tag>
//	    Validate both files. The release recipe and the release workflow run it
//	    before a tag ships.
//	changelog release-notes <changelog-dir> <tag> <owner/repo>
//	    Print the GitHub release body: player notes first, then the developer
//	    notes under "Developer details".
package main

import (
	"fmt"
	"os"

	"openmars/internal/changelog"
)

func main() {
	switch {
	case len(os.Args) == 4 && os.Args[1] == "check":
		check(os.Args[2], os.Args[3])
	case len(os.Args) == 5 && os.Args[1] == "release-notes":
		releaseNotes(os.Args[2], os.Args[3], os.Args[4])
	default:
		fmt.Fprintln(os.Stderr, "Usage: changelog check <changelog-dir> <tag>")
		fmt.Fprintln(os.Stderr, "       changelog release-notes <changelog-dir> <tag> <owner/repo>")
		os.Exit(2)
	}
}

func check(dir, tag string) {
	failed := false
	for _, format := range []changelog.Format{changelog.Developer, changelog.Player} {
		entry, err := changelog.Load(format, dir, tag)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			failed = true
			continue
		}
		fmt.Printf("%s %s: %d sections\n", entry.Version, format.FileName, len(entry.Sections))
	}
	if failed {
		os.Exit(1)
	}
}

// Changelog images are stored in Git LFS, which raw.githubusercontent.com serves
// as pointer files, so release bodies link the LFS media URL instead
func releaseNotes(dir, tag, repo string) {
	notes, err := changelog.ReleaseNotes(dir, tag, func(file string) string {
		return fmt.Sprintf("https://media.githubusercontent.com/media/%s/%s/changelog/%s/%s", repo, tag, tag, file)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(notes)
}
