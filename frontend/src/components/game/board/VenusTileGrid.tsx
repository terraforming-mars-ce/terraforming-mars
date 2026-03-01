import { useMemo, type RefObject } from "react";
import * as THREE from "three";
import { HexGrid2D } from "../../../utils/hex-grid-2d";
import Tile, { TileHighlightMode } from "./Tile";
import { GameDto, TileDto, TileBonusDto } from "../../../types/generated/api-types";
import { VENUS_RADIUS, VENUS_POSITION } from "./boardConstants";
import { getPlayerColor } from "../../../utils/playerColors";
import { usePreviousTiles } from "../../../hooks/usePreviousTiles";

interface VenusTileGridProps {
  gameState?: GameDto;
  onHexClick?: (hexCoordinate: string) => void;
  tileHighlightMode?: TileHighlightMode;
  tileOpacity?: RefObject<number>;
}

interface ProjectedVenusTile {
  backendTile: TileDto;
  coordinate: { q: number; r: number; s: number };
  position: { x: number; y: number };
  spherePosition: THREE.Vector3;
  normal: THREE.Vector3;
  isOceanSpace: boolean;
  bonuses: { [key: string]: number };
}

type TileType = "city" | "empty" | "special";

const VENUS_COORD_OFFSET = { q: 100, r: 0, s: -100 };

const hexToPixel = (coord: { q: number; r: number; s: number }) => {
  const q = coord.q - VENUS_COORD_OFFSET.q;
  const r = coord.r - VENUS_COORD_OFFSET.r;
  const size = 0.3;
  const x = size * Math.sqrt(3) * (q + r / 2);
  const y = ((size * 3) / 2) * r;
  return { x, y };
};

function projectToSphere(position2D: { x: number; y: number }, radius: number): THREE.Vector3 {
  const scale = 0.4;
  const x = position2D.x * scale;
  const y = position2D.y * scale;

  const r = Math.sqrt(x * x + y * y);

  if (r === 0) {
    return new THREE.Vector3(0, 0, radius);
  }

  const theta = Math.atan2(y, x);
  const phi = (r / radius) * (Math.PI / 2);

  const sphereX = radius * Math.sin(phi) * Math.cos(theta);
  const sphereY = radius * Math.sin(phi) * Math.sin(theta);
  const sphereZ = radius * Math.cos(phi);

  return new THREE.Vector3(sphereX, sphereY, sphereZ);
}

const convertBonuses = (bonuses: TileBonusDto[] | undefined) => {
  const converted: { [key: string]: number } = {};
  if (bonuses) {
    bonuses.forEach((bonus) => {
      converted[bonus.type] = bonus.amount;
    });
  }
  return converted;
};

export default function VenusTileGrid({
  gameState,
  onHexClick,
  tileHighlightMode,
  tileOpacity,
}: VenusTileGridProps) {
  const venusTilesOnly = useMemo(
    () => gameState?.board?.tiles?.filter((t) => t.location === "venus"),
    [gameState?.board?.tiles],
  );
  const newlyPlacedTiles = usePreviousTiles(venusTilesOnly);

  const venusWorldCenter = useMemo(
    () => new THREE.Vector3(VENUS_POSITION[0], VENUS_POSITION[1], VENUS_POSITION[2]),
    [],
  );

  const playerColorMap = useMemo(() => {
    const map = new Map<string, string>();
    if (gameState?.turnOrder) {
      gameState.turnOrder.forEach((playerId, index) => {
        map.set(playerId, getPlayerColor(index));
      });
    }
    return map;
  }, [gameState?.turnOrder]);

  const venusTiles = useMemo((): ProjectedVenusTile[] => {
    if (!gameState?.board?.tiles) return [];

    return gameState.board.tiles
      .filter((tile: TileDto) => tile.location === "venus")
      .map((tile: TileDto): ProjectedVenusTile => {
        const position2D = hexToPixel(tile.coordinates);
        const spherePosition = projectToSphere(position2D, VENUS_RADIUS);

        return {
          backendTile: tile,
          coordinate: tile.coordinates,
          position: position2D,
          spherePosition,
          normal: spherePosition.clone().normalize(),
          isOceanSpace: false,
          bonuses: convertBonuses(tile.bonuses),
        };
      });
  }, [gameState?.board?.tiles]);

  const availableHexes = gameState?.currentPlayer?.pendingTileSelection?.availableHexes || [];

  const getTileType = (tile: ProjectedVenusTile): TileType => {
    if (tile.backendTile.occupiedBy) {
      if (tile.backendTile.occupiedBy.type === "city-tile") return "city";
      return "special";
    }
    return "empty";
  };

  return (
    <>
      {venusTiles.map((tile) => {
        const hexKey = HexGrid2D.coordinateToKey(tile.coordinate);
        const tileType = getTileType(tile);
        const isAvailable = availableHexes.includes(hexKey);
        const ownerId = tile.backendTile.ownerId || null;

        return (
          <Tile
            key={hexKey}
            tileData={tile}
            tileType={tileType}
            ownerId={ownerId}
            ownerColor={ownerId ? playerColorMap.get(ownerId) : undefined}
            reservedById={tile.backendTile.reservedBy || null}
            displayName={tile.backendTile.displayName}
            isVolcanic={false}
            onClick={() => {
              onHexClick?.(hexKey);
            }}
            isAvailableForPlacement={isAvailable}
            isNewlyPlaced={newlyPlacedTiles.has(hexKey)}
            highlightMode={tileHighlightMode === "city" && tileType === "city" ? "city" : null}
            sphereRadius={VENUS_RADIUS}
            sphereCenter={venusWorldCenter}
            tileOpacity={tileOpacity}
          />
        );
      })}
    </>
  );
}
