import type { CardDescriptionSectionDto } from "@/types/generated/api-types.ts";

type SectionType = CardDescriptionSectionDto["type"];

const SECTION_LABELS: Record<SectionType, string | null> = {
  generic: null,
  effect: "Effect",
  action: "Action",
  requirement: "Requirement",
};

export const getCardDescriptionSectionLabel = (type: SectionType): string | null =>
  SECTION_LABELS[type];

export const cardDescriptionToPlainText = (sections: CardDescriptionSectionDto[]): string =>
  sections
    .map((section) => {
      const label = SECTION_LABELS[section.type];
      return label ? `${label}: ${section.text}` : section.text;
    })
    .join("\n");
