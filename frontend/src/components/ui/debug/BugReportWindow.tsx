import React, { useCallback, useEffect, useRef, useState } from "react";
import html2canvas from "html2canvas";
import { useWindowDrag, useWindowManager } from "./WindowManager.tsx";
import { apiService } from "@/services/apiService.ts";
import { BugReportDto, GameDto } from "@/types/generated/api-types.ts";

interface BugReportWindowProps {
  isVisible: boolean;
  onClose: () => void;
  gameState: GameDto | null;
}

const WINDOW_ID = "bug-report";
const WINDOW_WIDTH = 380;
const ACCENT = "#f59e0b";
const ACCENT_SHADOW = "rgba(245, 158, 11, 0.3)";
const EXCLUDE_SELECTORS = [".bug-report-content-area"];
const POLL_INTERVAL_MS = 2000;

type ServiceStatus =
  | { state: "loading" }
  | { state: "available" }
  | { state: "unavailable"; reason: string };

const BugReportWindow: React.FC<BugReportWindowProps> = ({ isVisible, onClose, gameState }) => {
  const windowRef = useRef<HTMLDivElement>(null);
  const [description, setDescription] = useState("");
  const [author, setAuthor] = useState("");
  const [includeScreenshot, setIncludeScreenshot] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [report, setReport] = useState<BugReportDto | null>(null);
  const [status, setStatus] = useState<ServiceStatus>({ state: "loading" });
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const { position, isDragging, handleMouseDown } = useWindowDrag({
    windowId: WINDOW_ID,
    width: WINDOW_WIDTH,
    height: 420,
    defaultPosition:
      typeof window !== "undefined"
        ? { x: Math.max(20, window.innerWidth / 2 - WINDOW_WIDTH / 2), y: 80 }
        : undefined,
    excludeSelectors: EXCLUDE_SELECTORS,
    isVisible,
  });

  const { getZIndex } = useWindowManager();

  const stopPolling = useCallback(() => {
    if (pollRef.current) {
      clearInterval(pollRef.current);
      pollRef.current = null;
    }
  }, []);

  useEffect(() => {
    if (!isVisible) {
      stopPolling();
      return;
    }

    setStatus({ state: "loading" });
    setReport(null);
    setErrorMessage(null);
    setDescription("");
    setAuthor("");

    apiService
      .getBugReportStatus()
      .then((res) => {
        if (res.available) {
          setStatus({ state: "available" });
        } else {
          setStatus({ state: "unavailable", reason: res.reason || "Unknown reason" });
        }
      })
      .catch(() => {
        setStatus({ state: "unavailable", reason: "Could not reach server" });
      });

    return stopPolling;
  }, [isVisible, stopPolling]);

  const startPolling = useCallback(
    (reportId: string) => {
      stopPolling();
      pollRef.current = setInterval(() => {
        apiService
          .getBugReport(reportId)
          .then((r) => {
            setReport(r);
            if (r.status === "completed" || r.status === "failed") {
              stopPolling();
            }
          })
          .catch(() => {
            stopPolling();
            setReport(null);
            setErrorMessage("Lost connection while processing bug report");
          });
      }, POLL_INTERVAL_MS);
    },
    [stopPolling],
  );

  const handleSubmit = async () => {
    if (description.trim().length === 0 || status.state !== "available") return;

    setErrorMessage(null);

    try {
      let screenshot: string | undefined;

      if (includeScreenshot) {
        const bugWindow = document.querySelector("[data-bug-report-window]") as HTMLElement;
        if (bugWindow) bugWindow.style.display = "none";
        try {
          const canvas = await html2canvas(document.body, {
            useCORS: true,
            logging: false,
            scale: 0.5,
            onclone: (clonedDoc) => {
              clonedDoc.querySelectorAll("*").forEach((el) => {
                const style = (el as HTMLElement).style;
                if (style?.colorScheme) style.colorScheme = "";
              });
            },
          });
          screenshot = canvas.toDataURL("image/png");
        } catch {
          console.warn("Screenshot capture failed, submitting without screenshot");
        } finally {
          if (bugWindow) bugWindow.style.display = "";
        }
      }

      setIsSubmitting(true);

      const result = await apiService.submitBugReport({
        description: description.trim(),
        author: author.trim() || gameState?.currentPlayer?.name,
        includeScreenshot,
        screenshot,
        gameState: gameState ?? undefined,
      });

      setReport(result);
      startPolling(result.id);
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : "Failed to submit bug report");
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isVisible) return null;

  const isProcessing = report !== null && report.status === "processing";
  const isCompleted = report !== null && report.status === "completed";
  const isFailed = report !== null && report.status === "failed";
  const showForm = status.state === "available" && !report && !isSubmitting;
  const canSubmit = showForm && description.trim().length > 0;

  return (
    <div
      ref={windowRef}
      data-bug-report-window
      onMouseDown={handleMouseDown}
      style={{
        position: "fixed",
        top: `${position.y}px`,
        left: `${position.x}px`,
        width: `${WINDOW_WIDTH}px`,
        background: "rgba(0, 0, 0, 0.95)",
        border: `2px solid ${ACCENT}`,
        borderRadius: "8px",
        padding: "12px 16px",
        zIndex: getZIndex(WINDOW_ID),
        overflow: "hidden",
        display: "flex",
        flexDirection: "column",
        boxShadow: `0 4px 20px ${ACCENT_SHADOW}`,
        cursor: isDragging ? "grabbing" : "default",
        transition: isDragging ? "none" : "top 0.2s ease-out, left 0.2s ease-out",
      }}
    >
      {/* Title bar */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "10px",
          paddingBottom: "8px",
          borderBottom: "1px solid #333",
          userSelect: "none",
          cursor: "grab",
        }}
      >
        <h3
          style={{
            margin: 0,
            color: ACCENT,
            fontSize: "14px",
            display: "flex",
            alignItems: "center",
            gap: "8px",
          }}
          className="font-orbitron"
        >
          <svg
            width="10"
            height="14"
            viewBox="0 0 10 14"
            fill="currentColor"
            style={{ opacity: 0.5 }}
          >
            <circle cx="2" cy="2" r="1.5" />
            <circle cx="8" cy="2" r="1.5" />
            <circle cx="2" cy="7" r="1.5" />
            <circle cx="8" cy="7" r="1.5" />
            <circle cx="2" cy="12" r="1.5" />
            <circle cx="8" cy="12" r="1.5" />
          </svg>
          Report Bug
        </h3>
        <button
          onClick={onClose}
          onMouseDown={(e) => e.stopPropagation()}
          style={{
            background: "none",
            border: "none",
            color: "#abb2bf",
            fontSize: "18px",
            cursor: "pointer",
            padding: "0 4px",
            lineHeight: 1,
          }}
        >
          ×
        </button>
      </div>

      {/* Content area */}
      <div
        className="bug-report-content-area"
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          gap: "10px",
        }}
      >
        {status.state === "loading" && <CenteredMessage text="Checking availability..." />}

        {status.state === "unavailable" && (
          <CenteredMessage text={`Bug reporting is not available: ${status.reason}`} />
        )}

        {(isSubmitting || isProcessing) && (
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              justifyContent: "center",
              minHeight: "160px",
              gap: "16px",
            }}
          >
            <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
              <circle
                cx="24"
                cy="24"
                r="22"
                stroke="rgba(255, 255, 255, 0.15)"
                strokeWidth="2.5"
                fill="none"
              />
              <circle
                cx="24"
                cy="24"
                r="22"
                stroke={ACCENT}
                strokeWidth="2.5"
                fill="none"
                strokeLinecap="round"
                strokeDasharray="100 38"
              >
                <animateTransform
                  attributeName="transform"
                  type="rotate"
                  from="0 24 24"
                  to="360 24 24"
                  dur="1s"
                  repeatCount="indefinite"
                />
              </circle>
            </svg>

            <span
              style={{
                color: "rgba(255, 255, 255, 0.6)",
                fontSize: "13px",
                textAlign: "center",
              }}
            >
              {report?.statusMessage || "Submitting..."}
            </span>
          </div>
        )}

        {isCompleted && report.issueUrl && (
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              justifyContent: "center",
              minHeight: "160px",
              gap: "16px",
            }}
          >
            <span
              style={{
                color: "rgba(255, 255, 255, 0.8)",
                fontSize: "14px",
              }}
              className="font-orbitron"
            >
              {report.statusMessage}
            </span>

            <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
              <circle cx="24" cy="24" r="22" stroke="#4ade80" strokeWidth="2.5" fill="none" />
              <path
                d="M14 24.5L21 31.5L34 18.5"
                stroke="#4ade80"
                strokeWidth="2.5"
                strokeLinecap="round"
                strokeLinejoin="round"
                fill="none"
              />
            </svg>

            <a
              href={report.issueUrl}
              target="_blank"
              rel="noopener noreferrer"
              style={{
                color: ACCENT,
                fontSize: "12px",
                wordBreak: "break-all",
              }}
            >
              {report.issueUrl}
            </a>
          </div>
        )}

        {isFailed && (
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              justifyContent: "center",
              minHeight: "160px",
              gap: "16px",
            }}
          >
            <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
              <circle cx="24" cy="24" r="22" stroke="#f87171" strokeWidth="2.5" fill="none" />
              <path
                d="M16 16L32 32M32 16L16 32"
                stroke="#f87171"
                strokeWidth="2.5"
                strokeLinecap="round"
                fill="none"
              />
            </svg>

            <span
              style={{
                color: "#f87171",
                fontSize: "13px",
                textAlign: "center",
              }}
            >
              {report.statusMessage}
            </span>
          </div>
        )}

        {showForm && (
          <>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Describe the bug..."
              spellCheck={false}
              autoCorrect="off"
              autoComplete="off"
              maxLength={30000}
              style={{
                width: "100%",
                minHeight: "120px",
                resize: "vertical",
                background: "rgba(255, 255, 255, 0.05)",
                border: "1px solid #444",
                borderRadius: "4px",
                padding: "8px",
                color: "#fff",
                fontSize: "13px",
                fontFamily: "inherit",
                outline: "none",
              }}
              onFocus={(e) => (e.target.style.borderColor = ACCENT)}
              onBlur={(e) => (e.target.style.borderColor = "#444")}
            />

            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                fontSize: "11px",
                color: "rgba(255, 255, 255, 0.3)",
              }}
            >
              <span>{description.length.toLocaleString()}/30,000</span>
            </div>

            <input
              type="text"
              value={author}
              onChange={(e) => setAuthor(e.target.value)}
              placeholder={`Your name${gameState?.currentPlayer?.name ? ` (default: ${gameState.currentPlayer.name})` : ""}`}
              spellCheck={false}
              autoCorrect="off"
              autoComplete="off"
              maxLength={100}
              style={{
                width: "100%",
                background: "rgba(255, 255, 255, 0.05)",
                border: "1px solid #444",
                borderRadius: "4px",
                padding: "8px",
                color: "#fff",
                fontSize: "13px",
                fontFamily: "inherit",
                outline: "none",
              }}
              onFocus={(e) => (e.target.style.borderColor = ACCENT)}
              onBlur={(e) => (e.target.style.borderColor = "#444")}
            />

            <label
              style={{
                display: "flex",
                alignItems: "center",
                gap: "8px",
                color: "rgba(255, 255, 255, 0.7)",
                fontSize: "13px",
                cursor: "pointer",
              }}
            >
              <input
                type="checkbox"
                checked={includeScreenshot}
                onChange={(e) => setIncludeScreenshot(e.target.checked)}
                style={{ accentColor: ACCENT }}
              />
              Include screenshot
            </label>

            <button
              onClick={() => void handleSubmit()}
              disabled={!canSubmit}
              style={{
                width: "100%",
                padding: "10px",
                background: ACCENT,
                color: "#000",
                border: "none",
                borderRadius: "4px",
                fontSize: "14px",
                fontWeight: "bold",
                opacity: canSubmit ? 1 : 0.3,
                cursor: canSubmit ? "pointer" : "default",
              }}
              className="font-orbitron"
            >
              Submit
            </button>

            {errorMessage && (
              <div
                style={{
                  fontSize: "12px",
                  color: "#f87171",
                  wordBreak: "break-word",
                }}
              >
                {errorMessage}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};

const CenteredMessage: React.FC<{ text: string }> = ({ text }) => (
  <div
    style={{
      display: "flex",
      alignItems: "center",
      justifyContent: "center",
      minHeight: "160px",
      color: "rgba(255, 255, 255, 0.4)",
      fontSize: "13px",
      textAlign: "center",
      padding: "0 20px",
    }}
  >
    {text}
  </div>
);

export default BugReportWindow;
