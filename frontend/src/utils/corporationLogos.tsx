import React from "react";
import { assetImage, corporations } from "@/assets";

export function getCorporationLogo(
  corporationName: string,
  className = "w-[220px] h-[110px]",
  sizes = "264px",
): React.ReactNode | null {
  const key = corporationName
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
  const logo = corporations[key];
  if (!logo) {
    return null;
  }
  return (
    <img
      {...assetImage(logo.id, sizes, 256)}
      alt={logo.name}
      className={`block max-w-full object-contain ${className}`}
    />
  );
}
