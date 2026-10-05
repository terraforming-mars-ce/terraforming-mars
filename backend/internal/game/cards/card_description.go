package cards

import "strings"

// DescriptionSectionType identifies how a card description section is labelled
type DescriptionSectionType string

const (
	DescriptionSectionGeneric     DescriptionSectionType = "generic"
	DescriptionSectionEffect      DescriptionSectionType = "effect"
	DescriptionSectionAction      DescriptionSectionType = "action"
	DescriptionSectionRequirement DescriptionSectionType = "requirement"
)

// DescriptionSection is one paragraph of a card description. Text excludes the
// generated label and may contain inline **bold** emphasis.
type DescriptionSection struct {
	Type DescriptionSectionType `json:"type"`
	Text string                 `json:"text"`
}

// CardDescription is the ordered list of sections making up a card's description
type CardDescription []DescriptionSection

// Label returns the generated prefix for a section type, or "" for generic sections
func (t DescriptionSectionType) Label() string {
	switch t {
	case DescriptionSectionEffect:
		return "Effect"
	case DescriptionSectionAction:
		return "Action"
	case DescriptionSectionRequirement:
		return "Requirement"
	default:
		return ""
	}
}

// IsValid reports whether the section type is one of the known types
func (t DescriptionSectionType) IsValid() bool {
	switch t {
	case DescriptionSectionGeneric, DescriptionSectionEffect, DescriptionSectionAction, DescriptionSectionRequirement:
		return true
	default:
		return false
	}
}

// PlainText flattens the description into labelled lines, one per section
func (d CardDescription) PlainText() string {
	lines := make([]string, 0, len(d))
	for _, section := range d {
		if label := section.Type.Label(); label != "" {
			lines = append(lines, label+": "+section.Text)
		} else {
			lines = append(lines, section.Text)
		}
	}
	return strings.Join(lines, "\n")
}

// DeepCopy returns an independent copy of the description
func (d CardDescription) DeepCopy() CardDescription {
	if d == nil {
		return nil
	}
	sections := make(CardDescription, len(d))
	copy(sections, d)
	return sections
}
