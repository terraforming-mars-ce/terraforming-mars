import { createContext, useContext, useState, useRef, ReactNode } from "react";

export const SKYBOX_OPTIONS = [
  {
    id: "starmap-2020-8k",
    label: "NASA Starmap 2020",
    path: "/assets/backgrounds/starmap_2020_8k.exr",
  },
] as const;

export type SkyboxId = (typeof SKYBOX_OPTIONS)[number]["id"];

export interface World3DSettings {
  sunDirectionX: number;
  sunDirectionY: number;
  sunDirectionZ: number;
  sunIntensity: number;
  sunColor: { r: number; g: number; b: number };
  waterColor: { r: number; g: number; b: number };
  reflectance: number;
  freeCameraEnabled: boolean;
  showCameraFrustum: boolean;
  skyboxId: SkyboxId;
  skyboxBrightness: number;
}

export interface StoredCameraState {
  position: { x: number; y: number; z: number };
  spherical: { radius: number; phi: number; theta: number };
}

export interface CameraDisplayState {
  position: { x: number; y: number; z: number };
  rotation: { x: number; y: number; z: number };
}

export interface PendingCameraTransform {
  position?: { x: number; y: number; z: number };
  rotation?: { x: number; y: number; z: number };
}

const defaultSettings: World3DSettings = {
  sunDirectionX: 0.9,
  sunDirectionY: 0.0,
  sunDirectionZ: 0.8,
  sunIntensity: 0.75,
  sunColor: { r: 1.0, g: 0.86, b: 0.72 },
  waterColor: { r: 0.05, g: 0.09, b: 0.1 },
  reflectance: 0.1,
  freeCameraEnabled: false,
  showCameraFrustum: false,
  skyboxId: "starmap-2020-8k" as SkyboxId,
  skyboxBrightness: 0.35,
};

interface World3DSettingsContextType {
  settings: World3DSettings;
  updateSettings: (partial: Partial<World3DSettings>) => void;
  resetSettings: () => void;
  storedCameraState: StoredCameraState | null;
  setStoredCameraState: (state: StoredCameraState | null) => void;
  cameraStateRef: React.MutableRefObject<CameraDisplayState>;
  pendingCameraTransformRef: React.MutableRefObject<PendingCameraTransform | null>;
}

const World3DSettingsContext = createContext<World3DSettingsContextType | null>(null);

export function World3DSettingsProvider({ children }: { children: ReactNode }) {
  const [settings, setSettings] = useState<World3DSettings>(defaultSettings);
  const [storedCameraState, setStoredCameraState] = useState<StoredCameraState | null>(null);
  const cameraStateRef = useRef<CameraDisplayState>({
    position: { x: 0, y: 0, z: 8 },
    rotation: { x: 0, y: 0, z: 0 },
  });
  const pendingCameraTransformRef = useRef<PendingCameraTransform | null>(null);

  const updateSettings = (partial: Partial<World3DSettings>) => {
    setSettings((prev) => ({ ...prev, ...partial }));
  };

  const resetSettings = () => {
    setSettings(defaultSettings);
  };

  return (
    <World3DSettingsContext.Provider
      value={{
        settings,
        updateSettings,
        resetSettings,
        storedCameraState,
        setStoredCameraState,
        cameraStateRef,
        pendingCameraTransformRef,
      }}
    >
      {children}
    </World3DSettingsContext.Provider>
  );
}

export function useWorld3DSettings() {
  const context = useContext(World3DSettingsContext);
  if (!context) {
    throw new Error("useWorld3DSettings must be used within World3DSettingsProvider");
  }
  return context;
}
