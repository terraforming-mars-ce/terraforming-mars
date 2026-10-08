import React, { createContext, useCallback, useContext, useState } from "react";

export type NotificationSeverity = "error" | "warning" | "info";

interface Notification {
  id: string;
  message: string;
  type: NotificationSeverity;
  duration: number;
  isExiting: boolean;
}

interface ShowNotificationOptions {
  message: string;
  type: NotificationSeverity;
  duration?: number;
}

interface NotificationContextType {
  notifications: Notification[];
  showNotification: (options: ShowNotificationOptions) => string;
  dismissNotification: (id: string) => void;
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

export function NotificationProvider({ children }: { children: React.ReactNode }) {
  const [notifications, setNotifications] = useState<Notification[]>([]);

  const showNotification = useCallback((options: ShowNotificationOptions): string => {
    const id = crypto.randomUUID();
    const duration = options.duration ?? 3000;

    const notification: Notification = {
      id,
      message: options.message,
      type: options.type,
      duration,
      isExiting: false,
    };

    setNotifications((prev) => [...prev, notification]);

    if (duration > 0) {
      setTimeout(() => {
        setNotifications((prev) => prev.map((n) => (n.id === id ? { ...n, isExiting: true } : n)));
        setTimeout(() => {
          setNotifications((prev) => prev.filter((n) => n.id !== id));
        }, 200);
      }, duration);
    }

    return id;
  }, []);

  const dismissNotification = useCallback((id: string) => {
    setNotifications((prev) => prev.map((n) => (n.id === id ? { ...n, isExiting: true } : n)));
    setTimeout(() => {
      setNotifications((prev) => prev.filter((n) => n.id !== id));
    }, 200);
  }, []);

  const contextValue: NotificationContextType = {
    notifications,
    showNotification,
    dismissNotification,
  };

  return (
    <NotificationContext.Provider value={contextValue}>{children}</NotificationContext.Provider>
  );
}

export function useNotifications() {
  const context = useContext(NotificationContext);
  if (context === undefined) {
    throw new Error("useNotifications must be used within a NotificationProvider");
  }
  return context;
}
