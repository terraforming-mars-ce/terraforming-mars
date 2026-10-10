// Command changelog validates release notes in changelog/<tag>/CHANGELOG.md.
// The release recipe and the release workflow run it before a tag ships.
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
	entry, err := changelog.Load(dir, tag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%s: %d sections\n", entry.Version, len(entry.Sections))
}
