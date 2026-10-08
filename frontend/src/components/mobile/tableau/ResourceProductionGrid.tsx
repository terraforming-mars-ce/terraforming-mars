import {
  ResourceTypeCredit,
  ResourceTypeEnergy,
  ResourceTypeHeat,
  ResourceTypePlant,
  ResourceTypeSteel,
  ResourceTypeTitanium,
  type ResourceType,
} from "@/types/generated/api-types.ts";
import ResourceAmount from "../ResourceAmount.tsx";
import type { TableauPlayer } from "./tableauPlayers.ts";

type ResourceField = keyof TableauPlayer["resources"];

const RESOURCES: { field: ResourceField; icon: ResourceType; label: string }[] = [
  { field: "credits", icon: ResourceTypeCredit, label: "Credits" },
  { field: "steel", icon: ResourceTypeSteel, label: "Steel" },
  { field: "titanium", icon: ResourceTypeTitanium, label: "Titanium" },
  { field: "plants", icon: ResourceTypePlant, label: "Plants" },
  { field: "energy", icon: ResourceTypeEnergy, label: "Energy" },
  { field: "heat", icon: ResourceTypeHeat, label: "Heat" },
];

interface ResourceProductionGridProps {
  player: TableauPlayer;
  size: "compact" | "regular";
  layout?: "grid" | "row";
}

function containerClass(size: ResourceProductionGridProps["size"], layout: "grid" | "row") {
  if (layout === "row") {
    return "flex items-center gap-2";
  }
  if (size === "regular") {
    return "grid grid-cols-3 gap-x-5 gap-y-2";
  }
  return "grid grid-cols-3 gap-x-3 gap-y-1";
}

export default function ResourceProductionGrid({
  player,
  size,
  layout = "grid",
}: ResourceProductionGridProps) {
  if (!player.resources || !player.production) {
    return null;
  }
  const isRegular = size === "regular";

  return (
    <div className={containerClass(size, layout)}>
      {RESOURCES.map((resource) => {
        const amount = player.resources[resource.field];
        const production = player.production[resource.field];
        return (
          <div
            key={resource.field}
            aria-label={`${resource.label} ${amount}, production ${production}`}
            className={
              layout === "row"
                ? "flex items-center px-1.5 py-1 bg-white/[0.04] border border-white/10"
                : undefined
            }
          >
            <ResourceAmount
              resource={resource.icon}
              amount={amount}
              production={production}
              size={isRegular ? "md" : "sm"}
            />
          </div>
        );
      })}
    </div>
  );
}
