package dto

import "openmars/internal/changelog"

// ToChangelogResponse maps parsed player release notes to the changelog response
func ToChangelogResponse(entries []changelog.Entry) ChangelogResponse {
	response := ChangelogResponse{Entries: make([]ChangelogEntry, 0, len(entries))}
	for _, entry := range entries {
		sections := make([]ChangelogSection, 0, len(entry.Sections))
		for _, section := range entry.Sections {
			mapped := ChangelogSection{
				Title: section.Title,
				Major: section.Major,
				Intro: section.Intro,
				Items: section.Items,
			}
			if section.Image != nil {
				mapped.Image = &ChangelogImage{File: section.Image.File, Alt: section.Image.Alt}
			}
			sections = append(sections, mapped)
		}
		response.Entries = append(response.Entries, ChangelogEntry{
			Version:  entry.Version,
			Intro:    entry.Intro,
			Sections: sections,
		})
	}
	return response
}
