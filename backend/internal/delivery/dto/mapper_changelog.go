package dto

import "openmars/internal/changelog"

// ToChangelogResponse maps parsed release notes to the changelog response
func ToChangelogResponse(entries []changelog.Entry) ChangelogResponse {
	response := ChangelogResponse{Entries: make([]ChangelogEntry, 0, len(entries))}
	for _, entry := range entries {
		sections := make([]ChangelogSection, 0, len(entry.Sections))
		for _, section := range entry.Sections {
			sections = append(sections, ChangelogSection{Title: section.Title, Items: section.Items})
		}
		response.Entries = append(response.Entries, ChangelogEntry{
			Version:  entry.Version,
			Intro:    entry.Intro,
			Sections: sections,
		})
	}
	return response
}
