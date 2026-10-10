// Package changelog reads the release notes in changelog/<tag>/CHANGELOG.md.
//
// A file holds an optional intro paragraph followed by "## <Section>" headings
// from SectionOrder, in that order. Each section lists "- " bullets whose
// wrapped lines are indented two spaces. Nothing else is accepted, so the
// GitHub release, the release gate and the game all read the same notes.
package changelog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// FileName is the notes file inside each version folder
const FileName = "CHANGELOG.md"

// SectionOrder lists the allowed section headings in the order they must appear
var SectionOrder = []string{"Security", "Added", "Changed", "Fixed", "Build"}

var versionPattern = regexp.MustCompile(`^v\d+(\.\d+){0,2}$`)

// Entry is the release notes for one version
type Entry struct {
	Version  string
	Intro    string
	Sections []Section
}

// Section is one heading and its bullets
type Section struct {
	Title string
	Items []string
}

// IsVersion reports whether name is a release tag such as v7, v7.1 or v7.1.2
func IsVersion(name string) bool {
	return versionPattern.MatchString(name)
}

// CompareVersions orders two release tags numerically, counting missing parts as 0
func CompareVersions(a, b string) int {
	pa, pb := versionParts(a), versionParts(b)
	for i := range pa {
		if pa[i] != pb[i] {
			if pa[i] < pb[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func versionParts(version string) [3]int {
	var parts [3]int
	for i, part := range strings.Split(strings.TrimPrefix(version, "v"), ".") {
		if i >= len(parts) {
			break
		}
		parts[i], _ = strconv.Atoi(part)
	}
	return parts
}

// Parse reads the notes for version from the contents of its CHANGELOG.md
func Parse(version string, data []byte) (Entry, error) {
	if !IsVersion(version) {
		return Entry{}, fmt.Errorf("invalid version %q: expected vMAJOR[.MINOR[.PATCH]]", version)
	}

	entry := Entry{Version: version}
	var intro []string
	var section *Section
	var item *string
	nextSection := 0

	closeSection := func(line int) error {
		if section != nil && len(section.Items) == 0 {
			return fmt.Errorf("line %d: section %q has no bullets", line, section.Title)
		}
		return nil
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for i, raw := range lines {
		lineNo := i + 1
		line := strings.TrimRight(raw, " \t")

		switch {
		case line == "":
			item = nil

		case strings.HasPrefix(line, "## "):
			if err := closeSection(lineNo); err != nil {
				return Entry{}, err
			}
			title := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			position := slices.Index(SectionOrder, title)
			if position < 0 {
				return Entry{}, fmt.Errorf("line %d: unknown section %q, expected one of %s", lineNo, title, strings.Join(SectionOrder, ", "))
			}
			if position < nextSection {
				return Entry{}, fmt.Errorf("line %d: section %q is repeated or out of order, expected order %s", lineNo, title, strings.Join(SectionOrder, ", "))
			}
			nextSection = position + 1
			entry.Sections = append(entry.Sections, Section{Title: title})
			section = &entry.Sections[len(entry.Sections)-1]
			item = nil

		case strings.HasPrefix(line, "#"):
			return Entry{}, fmt.Errorf("line %d: only \"## <Section>\" headings are allowed", lineNo)

		case section == nil:
			if len(intro) > 0 && i > 0 && strings.TrimSpace(lines[i-1]) == "" {
				return Entry{}, fmt.Errorf("line %d: the intro must be a single paragraph", lineNo)
			}
			intro = append(intro, strings.TrimSpace(line))

		case line == "-":
			return Entry{}, fmt.Errorf("line %d: empty bullet", lineNo)

		case strings.HasPrefix(line, "- "):
			section.Items = append(section.Items, strings.TrimSpace(strings.TrimPrefix(line, "- ")))
			item = &section.Items[len(section.Items)-1]

		case strings.HasPrefix(line, "  ") && item != nil:
			*item += " " + strings.TrimSpace(line)

		default:
			return Entry{}, fmt.Errorf("line %d: expected a \"- \" bullet or a line indented two spaces under one", lineNo)
		}
	}
	if err := closeSection(len(lines)); err != nil {
		return Entry{}, err
	}

	entry.Intro = strings.Join(intro, " ")
	if entry.Intro == "" && len(entry.Sections) == 0 {
		return Entry{}, errors.New("no intro and no sections")
	}
	return entry, nil
}

// Load reads and parses dir/<version>/CHANGELOG.md
func Load(dir, version string) (Entry, error) {
	path := filepath.Join(dir, version, FileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return Entry{}, fmt.Errorf("failed to read changelog for %s: %w", version, err)
	}
	entry, err := Parse(version, data)
	if err != nil {
		return Entry{}, fmt.Errorf("%s: %w", path, err)
	}
	return entry, nil
}

// LoadAll reads every version folder in dir, newest version first
func LoadAll(dir string) ([]Entry, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read changelog directory: %w", err)
	}

	var entries []Entry
	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			continue
		}
		if !IsVersion(dirEntry.Name()) {
			return nil, fmt.Errorf("invalid changelog folder %q: expected vMAJOR[.MINOR[.PATCH]]", dirEntry.Name())
		}
		entry, err := Load(dir, dirEntry.Name())
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	slices.SortFunc(entries, func(a, b Entry) int {
		return CompareVersions(b.Version, a.Version)
	})
	return entries, nil
}
