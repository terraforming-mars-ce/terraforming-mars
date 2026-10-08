import type { ReactNode } from "react";
import GameButton from "@/components/ui/buttons/GameButton.tsx";
import { useLocation } from "react-router-dom";
import { useNotifications, type NotificationSeverity } from "@/contexts/NotificationContext.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";

const SEVERITY_CLASSES: Record<NotificationSeverity, string> = {
  error: "bg-red-900/90 border-red-500/50 text-red-100",
  warning: "bg-yellow-900/90 border-yellow-500/50 text-yellow-100",
  info: "bg-[#0b0e12]/95 border-white/15 text-white/90",
};

const SEVERITY_ICON_PATHS: Record<NotificationSeverity, ReactNode> = {
  error: (
    <>
      <circle cx="12" cy="12" r="10" />
      <line x1="12" y1="8" x2="12" y2="12" />
      <line x1="12" y1="16" x2="12.01" y2="16" />
    </>
  ),
  warning: (
    <>
      <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
      <line x1="12" y1="9" x2="12" y2="13" />
      <line x1="12" y1="17" x2="12.01" y2="17" />
    </>
  ),
  info: (
    <>
      <circle cx="12" cy="12" r="10" />
      <line x1="12" y1="16" x2="12" y2="12" />
      <line x1="12" y1="8" x2="12.01" y2="8" />
    </>
  ),
};

const ENTER_ANIMATION =
  "animate-[notificationSlideIn_0.3s_ease-out] compact:animate-[notificationDropIn_0.3s_ease-out]";
const EXIT_ANIMATION =
  "animate-[notificationSlideOut_0.2s_ease-out_forwards] compact:animate-[notificationDropOut_0.2s_ease-out_forwards]";

export default function NotificationContainer() {
  const location = useLocation();
  const { notifications, dismissNotification } = useNotifications();

  if (location.pathname.startsWith("/game")) {
    return null;
  }

  return (
    <div
      className="menu-notifications flex flex-col gap-3 compact:gap-2"
      style={{ zIndex: Z_INDEX.SYSTEM_NOTIFICATIONS }}
    >
      {notifications.map((notification) => (
        <div
          key={notification.id}
          className={`flex items-center gap-3 px-4 py-3 rounded-none border backdrop-blur-sm ${
            notification.isExiting ? EXIT_ANIMATION : ENTER_ANIMATION
          } ${SEVERITY_CLASSES[notification.type]}`}
        >
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className="flex-shrink-0"
          >
            {SEVERITY_ICON_PATHS[notification.type]}
          </svg>
          <span className="flex-1 min-w-0 text-sm font-medium">{notification.message}</span>
          <GameButton
            emphasis="quiet"
            onClick={() => dismissNotification(notification.id)}
            className="ml-2 p-1 rounded-none hover:bg-white/10 transition-colors"
            aria-label="Dismiss notification"
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
            >
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </GameButton>
        </div>
      ))}
    </div>
  );
}
