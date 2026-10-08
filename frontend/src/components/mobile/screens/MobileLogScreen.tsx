import { useEffect, useRef } from "react";
import { useMobileUiStore, type MobileLogTab } from "@/stores/mobileUiStore.ts";
import { useGameLogs } from "@/hooks/useGameLogs.ts";
import GameLogList from "../../ui/popover/content/GameLogList.tsx";
import ChatOverlay from "../../ui/overlay/ChatOverlay.tsx";
import MobileScreen, { type MobileScreenTab } from "../MobileScreen.tsx";
import { LogIcon } from "../screenIcons.tsx";
import { useMobileGame } from "../MobileGameContext.tsx";

export default function MobileLogScreen() {
  const close = useMobileUiStore((s) => s.close);
  const logTab = useMobileUiStore((s) => s.logTab);
  const setLogTab = useMobileUiStore((s) => s.setLogTab);
  const lastSeenChatCount = useMobileUiStore((s) => s.lastSeenChatCount);
  const markChatSeen = useMobileUiStore((s) => s.markChatSeen);
  const { gameState, chatMessages, onSendChatMessage, playerColorMap } = useMobileGame();
  const logs = useGameLogs(gameState.id);
  const logScrollRef = useRef<HTMLDivElement>(null);
  const chatCount = chatMessages.length;
  const isChat = logTab === "chat";

  useEffect(() => {
    if (isChat) {
      markChatSeen(chatCount);
    }
  }, [isChat, chatCount, markChatSeen]);

  const tabs: (MobileScreenTab & { id: MobileLogTab })[] = [
    { id: "log", label: "Log" },
    { id: "chat", label: "Chat", badge: !isChat && chatCount > lastSeenChatCount },
  ];

  return (
    <MobileScreen
      icon={LogIcon}
      title="Log"
      tabs={tabs}
      activeTab={logTab}
      onTabChange={(tab) => setLogTab(tab as MobileLogTab)}
      onClose={close}
      keyboardAware
    >
      {isChat ? (
        <div className="h-full px-2 pt-1 pb-1">
          <ChatOverlay
            messages={chatMessages}
            onSendMessage={onSendChatMessage}
            playerColorMap={playerColorMap}
            embedded
            fill
          />
        </div>
      ) : (
        <div ref={logScrollRef} className="h-full overflow-y-auto overscroll-contain">
          <GameLogList
            logs={logs}
            gameState={gameState}
            density="screen"
            scrollContainerRef={logScrollRef}
            active
          />
        </div>
      )}
    </MobileScreen>
  );
}
