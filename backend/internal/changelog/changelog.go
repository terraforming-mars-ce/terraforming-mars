// Package changelog reads the release notes in changelog/<tag>/.
//
// Each version folder holds two files. CHANGELOG.md is for developers and is
// the GitHub release body; CHANGELOG-USER.md is for players and is what the
// game shows. Both start with an optional one-paragraph intro followed by
// "## <Section>" headings from the format's fixed list, in that order. Each
// section lists "- " bullets whose wrapped lines are indented two spaces. The
// player format also allows "## Major update: <Name>" sections before the
// others, each with its own intro paragraph, then optionally one
// "![Alt text](image.png)" line naming an image in the version folder, then
// its bullets.
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

// Format describes one kind of changelog file
type Format struct {
	FileName    string
	Sections    []string
	MajorPrefix string
	// Optional formats may be absent for a version; LoadAll skips those versions
	Optional bool
}

// Developer is the technical changelog used as the GitHub release body
var Developer = Format{
	FileName: "CHANGELOG.md",
	Sections: []string{"Security", "Added", "Changed", "Fixed", "Build"},
}

// Player is the changelog shown in the game. Versions that never reached
// players have none.
var Player = Format{
	FileName:    "CHANGELOG-USER.md",
	Sections:    []string{"New", "Improved", "Fixed"},
	MajorPrefix: "Major update: ",
	Optional:    true,
}

var (
	versionPattern   = regexp.MustCompile(`^v\d+(\.\d+){0,2}$`)
	imageLinePattern = regexp.MustCompile(`^!\[(.*)\]\((.*)\)$`)
	imageFilePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	imageExtensions  = []string{".png", ".jpg", ".jpeg", ".webp"}
)

// Entry is the release notes for one version
type Entry struct {
	Version  string
	Intro    string
	Sections []Section
}

// Section is one heading and its bullets. Major update sections also have an
// intro and may have an image.
type Section struct {
	Title string
	Major bool
	Intro string
	Image *Image
	Items []string
}

// Image is a picture stored next to the notes in the version folder
type Image struct {
	File string
	Alt  string
}

func parseImage(line string) (*Image, error) {
	match := imageLinePattern.FindStringSubmatch(line)
	alt, file := strings.TrimSpace(match[1]), strings.TrimSpace(match[2])
	if alt == "" {
		return nil, errors.New("an image needs alt text: ![What it shows](file.png)")
	}
	if !imageFilePattern.MatchString(file) || !slices.Contains(imageExtensions, strings.ToLower(filepath.Ext(file))) {
		return nil, fmt.Errorf("image %q must be a file in the version folder ending in %s", file, strings.Join(imageExtensions, ", "))
	}
	return &Image{File: file, Alt: alt}, nil
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

// Parse reads the notes for version from the contents of a file in the given format
func Parse(format Format, version string, data []byte) (Entry, error) {
	if !IsVersion(version) {
		return Entry{}, fmt.Errorf("invalid version %q: expected vMAJOR[.MINOR[.PATCH]]", version)
	}

	entry := Entry{Version: version}
	var section *Section
	var item *string
	previousBlank := false
	nextSection := 0

	closeSection := func(line int) error {
		if section == nil {
			return nil
		}
		if section.Major && section.Intro == "" {
			return fmt.Errorf("line %d: %q needs an intro paragraph before its bullets", line, section.Title)
		}
		if !section.Major && len(section.Items) == 0 {
			return fmt.Errorf("line %d: section %q has no bullets", line, section.Title)
		}
		return nil
	}

	appendParagraph := func(paragraph *string, text string, lineNo int) error {
		if *paragraph != "" && previousBlank {
			return fmt.Errorf("line %d: an intro must be a single paragraph", lineNo)
		}
		*paragraph = strings.TrimSpace(*paragraph + " " + text)
		return nil
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for i, raw := range lines {
		lineNo := i + 1
		line := strings.TrimRight(raw, " \t")
		blank := line == ""

		switch {
		case blank:
			item = nil

		case strings.HasPrefix(line, "## "):
			if err := closeSection(lineNo); err != nil {
				return Entry{}, err
			}
			next, err := format.heading(strings.TrimSpace(strings.TrimPrefix(line, "## ")), nextSection)
			if err != nil {
				return Entry{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
			if !next.Major {
				nextSection = slices.Index(format.Sections, next.Title) + 1
			}
			entry.Sections = append(entry.Sections, next)
			section = &entry.Sections[len(entry.Sections)-1]
			item = nil

		case strings.HasPrefix(line, "#"):
			return Entry{}, fmt.Errorf("line %d: only \"## <Section>\" headings are allowed", lineNo)

		case imageLinePattern.MatchString(line):
			if section == nil || !section.Major {
				return Entry{}, fmt.Errorf("line %d: images are only allowed in major update sections", lineNo)
			}
			if section.Intro == "" || len(section.Items) > 0 || section.Image != nil {
				return Entry{}, fmt.Errorf("line %d: a major update has at most one image, between its intro and its bullets", lineNo)
			}
			image, err := parseImage(line)
			if err != nil {
				return Entry{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
			section.Image = image

		case section == nil:
			if err := appendParagraph(&entry.Intro, line, lineNo); err != nil {
				return Entry{}, err
			}

		case line == "-":
			return Entry{}, fmt.Errorf("line %d: empty bullet", lineNo)

		case strings.HasPrefix(line, "- "):
			section.Items = append(section.Items, strings.TrimSpace(strings.TrimPrefix(line, "- ")))
			item = &section.Items[len(section.Items)-1]

		case strings.HasPrefix(line, "  ") && item != nil:
			*item += " " + strings.TrimSpace(line)

		case section.Major && len(section.Items) == 0 && section.Image == nil:
			if err := appendParagraph(&section.Intro, line, lineNo); err != nil {
				return Entry{}, err
			}

		default:
			return Entry{}, fmt.Errorf("line %d: expected a \"- \" bullet or a line indented two spaces under one", lineNo)
		}
		previousBlank = blank
	}
	if err := closeSection(len(lines)); err != nil {
		return Entry{}, err
	}

	if entry.Intro == "" && len(entry.Sections) == 0 {
		return Entry{}, errors.New("no intro and no sections")
	}
	return entry, nil
}

func (f Format) heading(title string, nextSection int) (Section, error) {
	majorPrefix := strings.TrimSpace(f.MajorPrefix)
	if majorPrefix != "" && strings.HasPrefix(title, majorPrefix) {
		if strings.TrimSpace(strings.TrimPrefix(title, majorPrefix)) == "" {
			return Section{}, fmt.Errorf("%q needs a name", title)
		}
		if nextSection > 0 {
			return Section{}, fmt.Errorf("%q must come before %s", title, strings.Join(f.Sections, ", "))
		}
		return Section{Title: title, Major: true}, nil
	}

	position := slices.Index(f.Sections, title)
	if position < 0 {
		allowed := strings.Join(f.Sections, ", ")
		if f.MajorPrefix != "" {
			allowed = f.MajorPrefix + "<Name>, " + allowed
		}
		return Section{}, fmt.Errorf("unknown section %q, expected one of %s", title, allowed)
	}
	if position < nextSection {
		return Section{}, fmt.Errorf("section %q is repeated or out of order, expected order %s", title, strings.Join(f.Sections, ", "))
	}
	return Section{Title: title}, nil
}

// Load reads and parses dir/<version>/<format file>
func Load(format Format, dir, version string) (Entry, error) {
	path := filepath.Join(dir, version, format.FileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return Entry{}, fmt.Errorf("failed to read %s for %s: %w", format.FileName, version, err)
	}
	entry, err := Parse(format, version, data)
	if err != nil {
		return Entry{}, fmt.Errorf("%s: %w", path, err)
	}
	for _, section := range entry.Sections {
		if section.Image == nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, version, section.Image.File)); err != nil {
			return Entry{}, fmt.Errorf("%s: image %s: %w", path, section.Image.File, err)
		}
	}
	return entry, nil
}

// LoadAll reads the format's file in every version folder in dir, newest version first
func LoadAll(format Format, dir string) ([]Entry, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read changelog directory: %w", err)
	}

	var entries []Entry
	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			continue
		}
		version := dirEntry.Name()
		if !IsVersion(version) {
			return nil, fmt.Errorf("invalid changelog folder %q: expected vMAJOR[.MINOR[.PATCH]]", version)
		}
		if format.Optional {
			if _, err := os.Stat(filepath.Join(dir, version, format.FileName)); errors.Is(err, os.ErrNotExist) {
				continue
			}
		}
		entry, err := Load(format, dir, version)
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
