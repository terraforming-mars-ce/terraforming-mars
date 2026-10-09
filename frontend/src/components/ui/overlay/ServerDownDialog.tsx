import GameButton from "@/components/ui/buttons/GameButton.tsx";
import GameMenuModal from "@/components/ui/overlay/GameMenuModal.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { useUIOverlayStore } from "@/stores/uiOverlayStore.ts";
import { changeServer, gatewayServer } from "@/utils/gateway.ts";

/**
 * Shown behind a gateway when the server stops answering. Everything it needs
 * is already loaded, and both actions go through the gateway, not the server.
 */
export default function ServerDownDialog() {
  const show = useUIOverlayStore((s) => s.showServerDown);
  if (!show || !gatewayServer) {
    return null;
  }
  return (
    <GameMenuModal
      title={`Lost connection to ${gatewayServer.name}`}
      showBackdrop={true}
      zIndex={Z_INDEX.CONFIRMATION_MODAL}
    >
      <p className="text-white/80 text-center mb-6">
        The server isn't responding. Games on a server that goes down can't continue.
      </p>
      <div className="flex gap-4 justify-center">
        <GameButton emphasis="secondary" onClick={() => window.location.reload()}>
          Retry
        </GameButton>
        <GameButton onClick={changeServer}>Switch server</GameButton>
      </div>
    </GameMenuModal>
  );
}
