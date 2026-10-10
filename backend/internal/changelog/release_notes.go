package changelog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReleaseNotes builds the GitHub release body for a version: the version as the
// title, the player notes, then the developer notes under "Developer details".
// A release body cannot reference files in the repo, so imageURL turns each
// image file name into an absolute URL.
func ReleaseNotes(dir, version string, imageURL func(file string) string) (string, error) {
	if _, err := Load(Developer, dir, version); err != nil {
		return "", err
	}
	developer, err := os.ReadFile(filepath.Join(dir, version, Developer.FileName))
	if err != nil {
		return "", fmt.Errorf("failed to read %s for %s: %w", Developer.FileName, version, err)
	}

	var notes strings.Builder
	fmt.Fprintf(&notes, "# %s\n\n", version)

	if _, err := os.Stat(filepath.Join(dir, version, Player.FileName)); err == nil {
		if _, err := Load(Player, dir, version); err != nil {
			return "", err
		}
		player, err := os.ReadFile(filepath.Join(dir, version, Player.FileName))
		if err != nil {
			return "", fmt.Errorf("failed to read %s for %s: %w", Player.FileName, version, err)
		}
		notes.WriteString(absoluteImages(strings.TrimSpace(string(player)), imageURL))
		notes.WriteString("\n\n")
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("failed to read %s for %s: %w", Player.FileName, version, err)
	}

	notes.WriteString("## Developer details\n\n")
	notes.WriteString(demoteHeadings(strings.TrimSpace(string(developer))))
	notes.WriteString("\n")
	return notes.String(), nil
}

func absoluteImages(text string, imageURL func(file string) string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		match := imageLinePattern.FindStringSubmatch(strings.TrimSpace(line))
		if match != nil {
			lines[i] = fmt.Sprintf("![%s](%s)", strings.TrimSpace(match[1]), imageURL(strings.TrimSpace(match[2])))
		}
	}
	return strings.Join(lines, "\n")
}

func demoteHeadings(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "## ") {
			lines[i] = "#" + line
		}
	}
	return strings.Join(lines, "\n")
}
