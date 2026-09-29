import { corporations } from "@/assets";

export function getCorporationBorderColor(corporationName: string): string {
  const key = corporationName
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
  return corporations[key]?.color ?? "#ffc107";
}
