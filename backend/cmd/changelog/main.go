// Command changelog validates the release notes in changelog/<tag>/: the
// developer CHANGELOG.md and the player CHANGELOG-USER.md. The release recipe
// and the release workflow run it before a tag ships.
package main

import (
	"fmt"
	"os"

	"openmars/internal/changelog"
)

func main() {
	if len(os.Args) != 4 || os.Args[1] != "check" {
		fmt.Fprintln(os.Stderr, "Usage: changelog check <changelog-dir> <tag>")
		os.Exit(2)
	}

	dir, tag := os.Args[2], os.Args[3]
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
