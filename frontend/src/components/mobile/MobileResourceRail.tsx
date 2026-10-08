import {
  ResourceTypeCredit,
  ResourceTypeEnergy,
  ResourceTypeGreeneryTile,
  ResourceTypeHeat,
  ResourceTypePlant,
  ResourceTypeSteel,
  ResourceTypeTemperature,
  ResourceTypeTitanium,
  type OtherPlayerDto,
  type PlayerDto,
  type ResourceType,
} from "@/types/generated/api-types.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";
import GameIcon from "../ui/display/GameIcon.tsx";
import ResourceAmount from "./ResourceAmount.tsx";
import { useMobileGame } from "./MobileGameContext.tsx";
import { getConversionAvailability } from "./conversionAvailability.ts";

type ResourceField = keyof PlayerDto["resources"];

interface RailResource {
  field: ResourceField;
  icon: ResourceType;
  label: string;
}

const RESOURCES: RailResource[] = [
  { field: "credits", icon: ResourceTypeCredit, label: "Credits" },
  { field: "steel", icon: ResourceTypeSteel, label: "Steel" },
  { field: "titanium", icon: ResourceTypeTitanium, label: "Titanium" },
  { field: "plants", icon: ResourceTypePlant, label: "Plants" },
  { field: "energy", icon: ResourceTypeEnergy, label: "Energy" },
  { field: "heat", icon: ResourceTypeHeat, label: "Heat" },
];

interface MobileResourceRailProps {
  spectatingPlayer?: PlayerDto | OtherPlayerDto | null;
  className?: string;
}

export default function MobileResourceRail({
  spectatingPlayer,
  className = "",
}: MobileResourceRailProps) {
  const {
    gameState,
    currentPlayer,
    isSpectator,
    onConvertPlantsToGreenery,
    onConvertHeatToTemperature,
  } = useMobileGame();

  const isSpectating = !!spectatingPlayer;
  const displayPlayer = isSpectating ? spectatingPlayer : currentPlayer;
  if ((isSpectator && !isSpectating) || !displayPlayer?.resources || !displayPlayer.production) {
    return null;
  }

  const conversions = getConversionAvailability(gameState, currentPlayer);
  const canConvertPlants = !isSpectating && conversions.plants;
  const canConvertHeat = !isSpectating && conversions.heat;

  const conversionFor = (field: ResourceField) => {
    if (field === "plants" && canConvertPlants) {
      return { icon: ResourceTypeGreeneryTile, onConvert: onConvertPlantsToGreenery };
    }
    if (field === "heat" && canConvertHeat) {
      return { icon: ResourceTypeTemperature, onConvert: onConvertHeatToTemperature };
    }
    return null;
  };

  return (
    <div
      data-testid="mobile-resource-rail"
      className={`fixed left-0 flex flex-col justify-center-safe overflow-y-auto overscroll-contain [scrollbar-width:none] [&::-webkit-scrollbar]:hidden bg-[rgba(3,3,4,0.82)] border-r border-white/10 pointer-events-auto ${className}`}
      style={{
        zIndex: Z_INDEX.MOBILE_HUD,
        top: "calc(var(--hud-top-h) + var(--safe-top))",
        bottom: "calc(var(--hud-dock-h) + var(--safe-bottom))",
        width: "calc(var(--hud-rail-left-w) + var(--safe-left))",
        paddingLeft: "var(--safe-left)",
      }}
    >
      {RESOURCES.map((resource) => {
        const amount = displayPlayer.resources[resource.field];
        const production = displayPlayer.production[resource.field];
        const conversion = conversionFor(resource.field);
        if (conversion) {
          return (
            <button
              key={resource.field}
              type="button"
              aria-label={`Convert ${resource.label.toLowerCase()}`}
              className="relative shrink-0 h-11 flex items-center px-1.5 cursor-pointer bg-amber-900/40 border-y border-amber-400/60 shadow-[inset_0_0_10px_rgba(245,158,11,0.35)] animate-pulse"
              onClick={conversion.onConvert}
            >
              <ResourceAmount
                resource={resource.icon}
                amount={amount}
                production={production}
                className="w-full justify-between"
              />
              <span
                className="absolute top-px left-1/2 -translate-x-1/2 flex items-center"
                aria-hidden="true"
              >
                <span className="text-[11px] font-bold leading-none text-white/90">+</span>
                <GameIcon iconType={conversion.icon} size="small" className="!w-3 !h-3" />
              </span>
            </button>
          );
        }
        return (
          <div
            key={resource.field}
            className="shrink-0 h-9 flex items-center px-1.5 border-b border-white/5 last:border-b-0"
            aria-label={`${resource.label} ${amount}, production ${production}`}
          >
            <ResourceAmount
              resource={resource.icon}
              amount={amount}
              production={production}
              className="w-full justify-between"
            />
          </div>
        );
      })}
    </div>
  );
}
