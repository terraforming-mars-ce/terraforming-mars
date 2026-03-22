import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import * as THREE from "three";
import { HexGrid2D } from "../../../utils/hex-grid-2d";
import Tile from "./Tile";
import { GameDto, TileDto, TileBonusDto } from "../../../types/generated/api-types";
import { usePreviousTiles } from "../../../hooks/usePreviousTiles";
import { useSoundEffects } from "../../../hooks/useSoundEffects";
import { useHoverSound } from "../../../hooks/useHoverSound";
import GreeneryRenderer from "./GreeneryRenderer";
import PrimitiveRenderer from "./PrimitiveManager";
import BirdRenderer from "./BirdRenderer";
import SpaceshipRenderer from "./SpaceshipRenderer";
import { SPHERE_RADIUS } from "./boardConstants";
import TileTooltip, { TileTooltipData } from "../../ui/display/TileTooltip";
import { Html } from "@react-three/drei";
import { panState } from "../controls/PanControls";
import { useVPCounting } from "../../../contexts/VPCountingContext";

const noop = () => {};
const VP_COLOR_SECONDARY: [number, number, number] = [0.4, 0.9, 0.4];
const VP_COLOR_PRIMARY: [number, number, number] = [0.95, 0.95, 1.0];

interface TileGridProps {
  gameState?: GameDto;
  onHexClick?: (hexCoordinate: string) => void;
  animateHexEntrance?: boolean;
  startHidden?: boolean;
}

// Local type for tiles with projected positions
interface ProjectedTile {
  backendTile?: TileDto;
  coordinate: { q: number; r: number; s: number };
  position: { x: number; y: number };
  spherePosition: THREE.Vector3;
  normal: THREE.Vector3;
  isOceanSpace: boolean;
  bonuses: { [key: string]: number };
}

// Type for the tile data returned by getTileData
type TileType =
  | "city"
  | "empty"
  | "ocean"
  | "greenery"
  | "special"
  | "volcano"
  | "nuclear-zone"
  | "mining"
  | "restricted"
  | "ecological-zone"
  | "natural-preserve"
  | "world-tree"
  | "mohole";

// Labels for special tile types (keyed by occupant type from backend)
const SPECIAL_TILE_LABELS: Record<string, string> = {};

interface TileData {
  type: TileType;
  ownerId: string | null;
  specialLabel: string | null;
}

export default function TileGrid({
  gameState,
  onHexClick,
  animateHexEntrance = false,
  startHidden = false,
}: TileGridProps) {
  const newlyPlacedTiles = usePreviousTiles(gameState?.board?.tiles);
  const { playWaterPlacementSound, playOxygenSound, playConstructionSound } = useSoundEffects();
  const { state: vpCountingState } = useVPCounting();

  // Play placement sounds for newly placed tiles
  useEffect(() => {
    if (newlyPlacedTiles.size === 0 || !gameState?.board?.tiles) return;

    for (const hexKey of newlyPlacedTiles) {
      const tile = gameState.board.tiles.find(
        (t) => HexGrid2D.coordinateToKey(t.coordinates) === hexKey,
      );
      if (!tile?.occupiedBy) continue;

      switch (tile.occupiedBy.type) {
        case "ocean-tile":
          void playWaterPlacementSound();
          break;
        case "greenery-tile":
        case "ecological-zone-tile":
        case "natural-preserve-tile":
          break;
        case "city-tile":
        case "special-tile":
        case "nuclear-zone-tile":
        case "mining-tile":
        case "restricted-tile":
          void playConstructionSound();
          break;
      }
    }
  }, [
    newlyPlacedTiles,
    gameState?.board?.tiles,
    playWaterPlacementSound,
    playOxygenSound,
    playConstructionSound,
  ]);

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
  const tooltipPositionRef = useRef<{ x: number; y: number }>({ x: 0, y: 0 });

  const handleTileHoverInfo = useCallback(
    (
      data: Omit<TileTooltipData, "ownerName" | "ownerColor" | "reservedByName"> & {
        position: { x: number; y: number };
        ownerId: string | null;
        reservedById: string | null;
      },
    ) => {
      tooltipPositionRef.current = data.position;
      setTooltipData({
        tileType: data.tileType,
        displayName: data.displayName,
        ownerName: data.ownerId ? playerNameMap.get(data.ownerId) : undefined,
        ownerColor: data.ownerId ? playerColorMap.get(data.ownerId) : undefined,
        reservedByName: data.reservedById ? playerNameMap.get(data.reservedById) : undefined,
        isOceanSpace: data.isOceanSpace,
        isVolcanic: data.isVolcanic,
        bonuses: data.bonuses,
      });
    },
    [playerNameMap, playerColorMap],
  );

  const handleTileHoverMove = useCallback((position: { x: number; y: number }) => {
    tooltipPositionRef.current = position;
  }, []);

  const handleTileHoverLeave = useCallback(() => {
    setTooltipData(null);
  }, []);

  // Convert hex coordinates to 2D pixel position (same as backend logic)
  const hexToPixel = (coord: { q: number; r: number; s: number }) => {
    const size = 0.3; // Same as HEX_SIZE in HexGrid2D
    const x = size * Math.sqrt(3) * (coord.q + coord.r / 2);
    const y = ((size * 3) / 2) * coord.r;
    return { x, y };
  };

  // Convert backend tile bonuses to legacy format
  const convertBackendBonuses = (bonuses: TileBonusDto[] | undefined) => {
    const converted: { [key: string]: number } = {};
    if (bonuses) {
      bonuses.forEach((bonus) => {
        converted[bonus.type] = bonus.amount;
      });
    }
    return converted;
  };

  // Use backend board tiles or fallback to hardcoded generation
  const projectedHexGrid = useMemo((): ProjectedTile[] => {
    // Use backend tiles if available (filter to Mars-only)
    if (gameState?.board?.tiles) {
      return gameState.board.tiles
        .filter((tile: TileDto) => tile.location !== "venus")
        .map((tile: TileDto): ProjectedTile => {
          // Convert hex coordinate to 2D position for projection
          const position2D = hexToPixel(tile.coordinates);
          const spherePosition = projectToSphere(position2D, SPHERE_RADIUS);

          return {
            backendTile: tile,
            coordinate: tile.coordinates,
            position: position2D,
            spherePosition,
            normal: spherePosition.clone().normalize(),
            // Convert backend tile data to legacy interface for compatibility
            isOceanSpace: tile.type === "ocean-space",
            bonuses: convertBackendBonuses(tile.bonuses),
          };
        });
    }

    // Fallback to hardcoded generation if backend data not available
    const hexGrid = HexGrid2D.generateGrid();
    return hexGrid.map((tile) => {
      const spherePosition = projectToSphere(tile.position, SPHERE_RADIUS);
      return {
        ...tile,
        spherePosition,
        normal: spherePosition.clone().normalize(),
      };
    });
  }, [gameState?.board?.tiles]);

  // Get tile type and occupancy data
  const getTileData = (tile: ProjectedTile): TileData => {
    if (tile.backendTile) {
      const backendTile: TileDto = tile.backendTile;

      // Determine tile type based on occupancy
      if (backendTile.occupiedBy) {
        switch (backendTile.occupiedBy.type) {
          case "ocean-tile":
            return {
              type: "ocean",
              ownerId: backendTile.ownerId || null,
              specialLabel: null,
            };
          case "city-tile":
            return {
              type: "city",
              ownerId: backendTile.ownerId || null,
              specialLabel: null,
            };
          case "greenery-tile":
            return {
              type: "greenery",
              ownerId: backendTile.ownerId || null,
              specialLabel: null,
            };
          case "volcano-tile":
            return {
              type: "volcano",
              ownerId: backendTile.ownerId || null,
              specialLabel: null,
            };
          case "nuclear-zone-tile":
            return {
              type: "nuclear-zone",
              ownerId: backendTile.ownerId || null,
              specialLabel: null,
            };
          case "ecological-zone-tile":
            return {
              type: "ecological-zone",
              ownerId: backendTile.ownerId || null,
              specialLabel: null,
            };
          case "natural-preserve-tile":
            return {
              type: "natural-preserve",
              ownerId: backendTile.ownerId || null,
              specialLabel: null,
            };
          case "mining-tile":
            return {
              type: "mining",
              ownerId: backendTile.ownerId || null,
              specialLabel: null,
            };
          case "restricted-tile":
            return {
              type: "restricted",
              ownerId: backendTile.ownerId || null,
              specialLabel: null,
            };
          case "world-tree-tile":
            return {
              type: "world-tree",
              ownerId: backendTile.ownerId || null,
              specialLabel: null,
            };
          case "mohole-tile":
            return {
              type: "mohole",
              ownerId: backendTile.ownerId || null,
              specialLabel: null,
            };
          default:
            return {
              type: "special",
              ownerId: backendTile.ownerId || null,
              specialLabel:
                SPECIAL_TILE_LABELS[backendTile.occupiedBy.type] || backendTile.occupiedBy.type,
            };
        }
      }

      // Reserved (land claim) but not occupied - show as fence tile
      if (backendTile.reservedBy) {
        return {
          type: "restricted",
          ownerId: backendTile.reservedBy,
          specialLabel: null,
        };
      }

      // Empty tile
      return {
        type: "empty",
        ownerId: backendTile.ownerId || null,
        specialLabel: null,
      };
    }

    // Fallback for hardcoded tiles
    return { type: "empty", ownerId: null, specialLabel: null };
  };

  // Get available hexes from current player's pending tile selection
  const availableHexes = gameState?.currentPlayer?.pendingTileSelection?.availableHexes || [];

  // Collect all greenery tiles for the GreeneryRenderer (includes living greenery types)
  const greeneryTiles = useMemo(() => {
    return projectedHexGrid
      .filter((tile) => {
        const tileData = getTileData(tile);
        return (
          tileData.type === "greenery" ||
          tileData.type === "ecological-zone" ||
          tileData.type === "natural-preserve"
        );
      })
      .map((tile) => ({
        coordinate: tile.coordinate,
        worldPosition: tile.spherePosition,
        normal: tile.normal,
      }));
  }, [projectedHexGrid]);

  // Collect living greenery tiles (ecological-zone, natural-preserve) for boosted vegetation
  const livingGreeneryTiles = useMemo(() => {
    return projectedHexGrid
      .filter((tile) => {
        const tileData = getTileData(tile);
        return tileData.type === "ecological-zone" || tileData.type === "natural-preserve";
      })
      .map((tile) => ({
        coordinate: tile.coordinate,
        worldPosition: tile.spherePosition,
        normal: tile.normal,
      }));
  }, [projectedHexGrid]);

  // Collect volcano tiles for vegetation around them
  const volcanoTiles = useMemo(() => {
    return projectedHexGrid
      .filter((tile) => {
        const tileData = getTileData(tile);
        return tileData.type === "volcano";
      })
      .map((tile) => ({
        coordinate: tile.coordinate,
        worldPosition: tile.spherePosition,
        normal: tile.normal,
      }));
  }, [projectedHexGrid]);

  // Detect newly placed greenery tiles
  const knownGreeneryRef = useRef<Set<string>>(new Set());
  const greeneryInitializedRef = useRef(false);
  const newGreeneryKeys = useMemo(() => {
    const currentKeys = new Set<string>();
    for (const tile of greeneryTiles) {
      currentKeys.add(`${tile.coordinate.q},${tile.coordinate.r},${tile.coordinate.s}`);
    }
    for (const tile of volcanoTiles) {
      currentKeys.add(`${tile.coordinate.q},${tile.coordinate.r},${tile.coordinate.s}`);
    }
    for (const tile of livingGreeneryTiles) {
      currentKeys.add(`${tile.coordinate.q},${tile.coordinate.r},${tile.coordinate.s}`);
    }

    if (!greeneryInitializedRef.current) {
      knownGreeneryRef.current = currentKeys;
      greeneryInitializedRef.current = true;
      return new Set<string>();
    }

    const added = new Set<string>();
    for (const key of currentKeys) {
      if (!knownGreeneryRef.current.has(key)) added.add(key);
    }
    knownGreeneryRef.current = currentKeys;
    return added;
  }, [greeneryTiles, volcanoTiles, livingGreeneryTiles]);

  // --- Centralized interaction sphere (single raycast target) ---
  const [hoveredHexKey, setHoveredHexKey] = useState<string | null>(null);
  const hoveredHexKeyRef = useRef<string | null>(null);
  const hoverSound = useHoverSound();
  const interactionSphereGeometry = useMemo(
    () => new THREE.SphereGeometry(SPHERE_RADIUS + 0.02, 64, 64),
    [],
  );

  const findNearestHex = useCallback(
    (hitPoint: THREE.Vector3): string | null => {
      let bestKey: string | null = null;
      let bestDist = Infinity;
      const HEX_HIT_RADIUS = 0.17;
      for (const tile of projectedHexGrid) {
        const dist = hitPoint.distanceTo(tile.spherePosition);
        if (dist < bestDist && dist < HEX_HIT_RADIUS) {
          bestDist = dist;
          bestKey = HexGrid2D.coordinateToKey(tile.coordinate);
        }
      }
      return bestKey;
    },
    [projectedHexGrid],
  );

  const handleSpherePointerMove = useCallback(
    (event: THREE.Event & { point: THREE.Vector3; nativeEvent: PointerEvent }) => {
      if (panState.isPanning) {
        if (hoveredHexKeyRef.current) {
          hoveredHexKeyRef.current = null;
          setHoveredHexKey(null);
          handleTileHoverLeave();
        }
        return;
      }
      const key = findNearestHex(event.point);
      if (key !== hoveredHexKeyRef.current) {
        hoveredHexKeyRef.current = key;
        setHoveredHexKey(key);
        if (key) {
          const tile = projectedHexGrid.find(
            (t) => HexGrid2D.coordinateToKey(t.coordinate) === key,
          );
          if (tile) {
            const tileData = getTileData(tile);
            const isAvailable = availableHexes.includes(key);
            if (isAvailable) {
              hoverSound.onMouseEnter?.();
            }
            handleTileHoverInfo({
              position: { x: event.nativeEvent.clientX, y: event.nativeEvent.clientY },
              tileType: tileData.type,
              displayName: tileData.specialLabel || tile.backendTile?.displayName,
              ownerId: tileData.ownerId,
              reservedById: tile.backendTile?.reservedBy || null,
              isOceanSpace: tile.isOceanSpace,
              isVolcanic: tile.backendTile?.tags?.includes("volcanic") ?? false,
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
      projectedHexGrid,
      availableHexes,
      findNearestHex,
      handleTileHoverInfo,
      handleTileHoverMove,
      handleTileHoverLeave,
      hoverSound,
    ],
  );

  const handleSpherePointerLeave = useCallback(() => {
    hoveredHexKeyRef.current = null;
    setHoveredHexKey(null);
    handleTileHoverLeave();
  }, [handleTileHoverLeave]);

  const handleSphereClick = useCallback(
    (event: THREE.Event & { point: THREE.Vector3 }) => {
      if (panState.isPanning || panState.hasDragged) return;
      const key = findNearestHex(event.point);
      if (key) {
        const isAvailable = availableHexes.includes(key);
        if (isAvailable) {
          hoverSound.onClick?.();
        }
        onHexClick?.(key);
      }
    },
    [findNearestHex, availableHexes, onHexClick, hoverSound],
  );

  return (
    <>
      {/* Single invisible sphere for all pointer interaction */}
      <mesh
        geometry={interactionSphereGeometry}
        onPointerMove={handleSpherePointerMove}
        onPointerLeave={handleSpherePointerLeave}
        onClick={handleSphereClick}
        visible={false}
      />

      {/* Tile hover tooltip — Html bridges R3F→DOM, TileTooltip portals to body */}
      <Html>
        <TileTooltip data={tooltipData} positionRef={tooltipPositionRef} />
      </Html>
      {/* Single GreeneryRenderer handles ALL greenery + volcano vegetation */}
      <GreeneryRenderer
        tiles={greeneryTiles}
        volcanoTiles={volcanoTiles}
        livingGreeneryTiles={livingGreeneryTiles}
        newTileKeys={newGreeneryKeys}
      />
      <PrimitiveRenderer />
      <BirdRenderer livingGreeneryTiles={livingGreeneryTiles} />
      {gameState?.settings?.cardPacks?.includes("colonies") && (
        <SpaceshipRenderer
          gameState={gameState}
          animateEntrance={animateHexEntrance}
          startHidden={startHidden}
        />
      )}
      {projectedHexGrid.map((tile, index) => {
        const hexKey = HexGrid2D.coordinateToKey(tile.coordinate);
        const tileData = getTileData(tile);
        const isAvailable = availableHexes.includes(hexKey);

        const isPrimaryHighlight = vpCountingState.highlightedTiles.has(hexKey);
        const isSecondaryHighlight = vpCountingState.secondaryHighlightedTiles.has(hexKey);
        const vpHighlightIntensity = isPrimaryHighlight ? 0.5 : isSecondaryHighlight ? 0.25 : 0;
        const vpHighlightColor = isSecondaryHighlight ? VP_COLOR_SECONDARY : VP_COLOR_PRIMARY;

        return (
          <Tile
            key={hexKey}
            tileData={tile}
            tileType={tileData.type}
            ownerId={tileData.ownerId}
            ownerColor={tileData.ownerId ? playerColorMap.get(tileData.ownerId) : undefined}
            reservedById={tile.backendTile?.reservedBy || null}
            displayName={tileData.specialLabel || tile.backendTile?.displayName}
            isVolcanic={tile.backendTile?.tags?.includes("volcanic") ?? false}
            isOceanSpace={tile.isOceanSpace}
            bonuses={tile.bonuses}
            onClick={noop}
            isHovered={hoveredHexKey === hexKey}
            isAvailableForPlacement={isAvailable}
            animateEntrance={animateHexEntrance}
            startHidden={startHidden}
            entranceDelay={index * 15}
            isNewlyPlaced={newlyPlacedTiles.has(hexKey)}
            vpHighlightIntensity={vpHighlightIntensity}
            vpHighlightColor={vpHighlightColor}
          />
        );
      })}
    </>
  );
}

/**
 * Project a 2D point onto the surface of a sphere
 * This simulates "wrapping" the flat hex grid around the sphere
 */
function projectToSphere(position2D: { x: number; y: number }, radius: number): THREE.Vector3 {
  // Scale the 2D coordinates to fit nicely on the sphere
  const scale = 0.4; // Reduced scale to bring hexagons closer together
  const x = position2D.x * scale;
  const y = position2D.y * scale;

  // Project onto sphere using azimuthal projection
  // This simulates "wrapping" the flat grid around the front hemisphere
  const r = Math.sqrt(x * x + y * y);

  if (r === 0) {
    // Center point
    return new THREE.Vector3(0, 0, radius);
  }

  // Convert to spherical coordinates
  const theta = Math.atan2(y, x); // Azimuth angle
  const phi = (r / radius) * (Math.PI / 2); // Polar angle (scaled to hemisphere)

  // Convert back to Cartesian coordinates on sphere
  const sphereX = radius * Math.sin(phi) * Math.cos(theta);
  const sphereY = radius * Math.sin(phi) * Math.sin(theta);
  const sphereZ = radius * Math.cos(phi);

  return new THREE.Vector3(sphereX, sphereY, sphereZ);
}
