import { useCallback, useMemo, useRef, useState, type RefObject } from "react";
import * as THREE from "three";
import { HexGrid2D } from "../../../utils/hex-grid-2d";
import Tile from "./Tile";
import { GameDto, TileDto, TileBonusDto } from "../../../types/generated/api-types";
import { VENUS_RADIUS, VENUS_POSITION } from "./boardConstants";
import { usePreviousTiles } from "../../../hooks/usePreviousTiles";
import TileTooltip, { TileTooltipData } from "../../ui/display/TileTooltip";
import { Html } from "@react-three/drei";
import { usePlanetFocus } from "../../../contexts/PlanetFocusContext";

interface VenusTileGridProps {
  gameState?: GameDto;
  onHexClick?: (hexCoordinate: string) => void;
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

export default function VenusTileGrid({ gameState, onHexClick, tileOpacity }: VenusTileGridProps) {
  const { activePlanet } = usePlanetFocus();

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
    if (gameState) {
      if (gameState.currentPlayer?.color) {
        map.set(gameState.currentPlayer.id, gameState.currentPlayer.color);
      }
      gameState.otherPlayers?.forEach((p) => {
        if (p.color) map.set(p.id, p.color);
      });
    }
    return map;
  }, [gameState]);

  const playerNameMap = useMemo(() => {
    const map = new Map<string, string>();
    if (gameState) {
      if (gameState.currentPlayer) {
        map.set(gameState.currentPlayer.id, gameState.currentPlayer.name);
      }
      gameState.otherPlayers?.forEach((p) => {
        map.set(p.id, p.name);
      });
    }
    return map;
  }, [gameState]);

  const [tooltipData, setTooltipData] = useState<TileTooltipData | null>(null);
  const hoverTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleTileHoverInfo = useCallback(
    (
      data: Omit<TileTooltipData, "ownerName" | "ownerColor" | "reservedByName"> & {
        ownerId: string | null;
        reservedById: string | null;
      },
    ) => {
      if (hoverTimerRef.current) clearTimeout(hoverTimerRef.current);
      hoverTimerRef.current = setTimeout(() => {
        setTooltipData({
          position: data.position,
          tileType: data.tileType,
          displayName: data.displayName,
          ownerName: data.ownerId ? playerNameMap.get(data.ownerId) : undefined,
          ownerColor: data.ownerId ? playerColorMap.get(data.ownerId) : undefined,
          reservedByName: data.reservedById ? playerNameMap.get(data.reservedById) : undefined,
          isOceanSpace: data.isOceanSpace,
          isVolcanic: data.isVolcanic,
          bonuses: data.bonuses,
        });
      }, 200);
    },
    [playerNameMap, playerColorMap],
  );

  const handleTileHoverMove = useCallback((position: { x: number; y: number }) => {
    setTooltipData((prev) => (prev ? { ...prev, position } : null));
  }, []);

  const handleTileHoverLeave = useCallback(() => {
    if (hoverTimerRef.current) {
      clearTimeout(hoverTimerRef.current);
      hoverTimerRef.current = null;
    }
    setTooltipData(null);
  }, []);

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

  const interactionSphereGeometry = useMemo(
    () => new THREE.SphereGeometry(VENUS_RADIUS + 0.02, 32, 32),
    [],
  );

  const findNearestHex = useCallback(
    (hitPoint: THREE.Vector3): string | null => {
      let bestKey: string | null = null;
      let bestDist = Infinity;
      const HEX_HIT_RADIUS = 0.17;
      for (const tile of venusTiles) {
        const dist = hitPoint.distanceTo(tile.spherePosition);
        if (dist < bestDist && dist < HEX_HIT_RADIUS) {
          bestDist = dist;
          bestKey = HexGrid2D.coordinateToKey(tile.coordinate);
        }
      }
      return bestKey;
    },
    [venusTiles],
  );

  const [hoveredHexKey, setHoveredHexKey] = useState<string | null>(null);

  const handleSpherePointerMove = useCallback(
    (
      event: THREE.Event & {
        point: THREE.Vector3;
        nativeEvent: PointerEvent;
        object: THREE.Object3D;
      },
    ) => {
      const localPoint = event.object.worldToLocal(event.point.clone());
      const key = findNearestHex(localPoint);
      if (key !== hoveredHexKey) {
        setHoveredHexKey(key);
        if (key) {
          const tile = venusTiles.find((t) => HexGrid2D.coordinateToKey(t.coordinate) === key);
          if (tile) {
            const ownerId = tile.backendTile.ownerId || null;
            handleTileHoverInfo({
              position: { x: event.nativeEvent.clientX, y: event.nativeEvent.clientY },
              tileType: getTileType(tile),
              displayName: tile.backendTile.displayName,
              ownerId,
              reservedById: tile.backendTile.reservedBy || null,
              isOceanSpace: false,
              isVolcanic: false,
              bonuses: tile.bonuses,
            });
          }
        } else {
          handleTileHoverLeave();
        }
      } else if (key) {
        handleTileHoverMove({ x: event.nativeEvent.clientX, y: event.nativeEvent.clientY });
      }
    },
    [
      hoveredHexKey,
      venusTiles,
      findNearestHex,
      handleTileHoverInfo,
      handleTileHoverMove,
      handleTileHoverLeave,
    ],
  );

  const handleSpherePointerLeave = useCallback(() => {
    setHoveredHexKey(null);
    handleTileHoverLeave();
  }, [handleTileHoverLeave]);

  const handleSphereClick = useCallback(
    (
      event: THREE.Event & {
        point: THREE.Vector3;
        object: THREE.Object3D;
        stopPropagation: () => void;
      },
    ) => {
      event.stopPropagation();
      const localPoint = event.object.worldToLocal(event.point.clone());
      const key = findNearestHex(localPoint);
      if (key) {
        onHexClick?.(key);
      }
    },
    [findNearestHex, onHexClick],
  );

  return (
    <>
      {/* Invisible interaction sphere for click/hover handling — disabled when viewing from Mars
         so it doesn't block raycasts to the VenusSphere click-to-travel handler */}
      {activePlanet === "venus" && (
        <mesh
          geometry={interactionSphereGeometry}
          onPointerMove={handleSpherePointerMove}
          onPointerLeave={handleSpherePointerLeave}
          onClick={handleSphereClick}
          visible={false}
        />
      )}

      <Html>
        <TileTooltip data={tooltipData} />
      </Html>
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
            sphereRadius={VENUS_RADIUS}
            sphereCenter={venusWorldCenter}
            tileOpacity={tileOpacity}
            onHoverInfo={handleTileHoverInfo}
            onHoverMove={handleTileHoverMove}
            onHoverLeave={handleTileHoverLeave}
            isHovered={hoveredHexKey === hexKey}
          />
        );
      })}
    </>
  );
}
