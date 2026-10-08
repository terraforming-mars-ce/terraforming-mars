import { useMobileUiStore, type MobileScreenId } from "@/stores/mobileUiStore.ts";
import MobileHandScreen from "./screens/MobileHandScreen.tsx";
import MobileActionsScreen from "./screens/MobileActionsScreen.tsx";
import MobileProjectsScreen from "./screens/MobileProjectsScreen.tsx";
import MobileTableauScreen from "./screens/MobileTableauScreen.tsx";
import MobilePlayersScreen from "./screens/MobilePlayersScreen.tsx";
import MobileLogScreen from "./screens/MobileLogScreen.tsx";

function renderScreen(screen: MobileScreenId) {
  switch (screen) {
    case "hand":
      return <MobileHandScreen />;
    case "actions":
      return <MobileActionsScreen />;
    case "projects":
      return <MobileProjectsScreen />;
    case "tableau":
      return <MobileTableauScreen />;
    case "players":
      return <MobilePlayersScreen />;
    case "log":
      return <MobileLogScreen />;
    default: {
      const unreachable: never = screen;
      return unreachable;
    }
  }
}

export default function MobileScreenHost() {
  const activeScreen = useMobileUiStore((s) => s.activeScreen);
  if (!activeScreen) {
    return null;
  }
  return renderScreen(activeScreen);
}
