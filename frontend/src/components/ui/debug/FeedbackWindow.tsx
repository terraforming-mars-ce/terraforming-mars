import FloatingWindow from "./FloatingWindow.tsx";
import GameButton from "@/components/ui/buttons/GameButton.tsx";
import React, { useCallback, useEffect, useRef, useState } from "react";
import { useWindowDrag, useWindowManager } from "./WindowManager.tsx";
import { apiService } from "@/services/apiService.ts";
import { FeedbackDto, GameDto } from "@/types/generated/api-types.ts";

interface FeedbackWindowProps {
  isVisible: boolean;
  onClose: () => void;
  gameState: GameDto | null;
}

const WINDOW_ID = "feedback";
const WINDOW_WIDTH = 400;
const ACCENT = "#8aa7ed";
const EXCLUDE_SELECTORS = [".feedback-content-area"];
const POLL_INTERVAL_MS = 2000;

type FeedbackTag = "bug" | "feature-request";

type ServiceStatus =
  | { state: "loading" }
  | { state: "available" }
  | { state: "unavailable"; reason: string };

const FeedbackWindow: React.FC<FeedbackWindowProps> = ({ isVisible, onClose, gameState }) => {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [author, setAuthor] = useState("");
  const [tags, setTags] = useState<FeedbackTag[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [report, setReport] = useState<FeedbackDto | null>(null);
  const [status, setStatus] = useState<ServiceStatus>({ state: "loading" });
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const { position, handleMouseDown } = useWindowDrag({
    windowId: WINDOW_ID,
    width: WINDOW_WIDTH,
    height: 480,
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
    setTitle("");
    setDescription("");
    setAuthor("");
    setTags([]);

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

  const toggleTag = (tag: FeedbackTag) => {
    setTags((prev) => (prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag]));
  };

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
            setErrorMessage("Lost connection while processing feedback");
          });
      }, POLL_INTERVAL_MS);
    },
    [stopPolling],
  );

  const handleSubmit = async () => {
    if (title.trim().length === 0 || status.state !== "available") {
      return;
    }

    setErrorMessage(null);

    try {
      setIsSubmitting(true);

      const result = await apiService.submitFeedback({
        title: title.trim(),
        description: description.trim(),
        tags,
        author: author.trim() || gameState?.currentPlayer?.name,
        gameState: gameState ?? undefined,
      });

      setReport(result);
      startPolling(result.id);
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : "Failed to submit feedback");
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isVisible) {
    return null;
  }

  const isProcessing = report !== null && report.status === "processing";
  const isCompleted = report !== null && report.status === "completed";
  const isFailed = report !== null && report.status === "failed";
  const showForm = status.state === "available" && !report && !isSubmitting;
  const canSubmit = showForm && title.trim().length > 0;

  return (
    <FloatingWindow
      title="Feedback"
      onClose={onClose}
      onMouseDown={handleMouseDown}
      style={{
        top: position.y,
        left: position.x,
        width: WINDOW_WIDTH,
        maxHeight: "calc(100dvh - 100px)",
        zIndex: getZIndex(WINDOW_ID),
      }}
    >
      <div
        className="feedback-content-area"
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          gap: "10px",
        }}
      >
        {status.state === "loading" && <CenteredMessage text="Checking availability..." />}

        {status.state === "unavailable" && (
          <CenteredMessage text={`Feedback is not available: ${status.reason}`} />
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
            <div style={{ display: "flex", gap: "8px" }}>
              <TagChip
                label="Bug"
                selected={tags.includes("bug")}
                onClick={() => toggleTag("bug")}
                color="#f87171"
              />
              <TagChip
                label="Feature Request"
                selected={tags.includes("feature-request")}
                onClick={() => toggleTag("feature-request")}
                color="#60a5fa"
              />
            </div>

            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Title"
              spellCheck={true}
              autoCorrect="off"
              autoComplete="off"
              maxLength={200}
              className="game-input w-full text-sm"
            />

            <textarea
              className="game-input w-full text-sm"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Add details..."
              spellCheck={true}
              autoCorrect="off"
              autoComplete="off"
              maxLength={30000}
              style={{
                minHeight: "100px",
                resize: "vertical",
              }}
            />

            <input
              type="text"
              value={author}
              onChange={(e) => setAuthor(e.target.value)}
              placeholder={`Your name${gameState?.currentPlayer?.name ? ` (default: ${gameState.currentPlayer.name})` : ""}`}
              spellCheck={false}
              autoCorrect="off"
              autoComplete="off"
              maxLength={100}
              className="game-input w-full text-sm"
            />

            <GameButton
              onClick={() => void handleSubmit()}
              disabled={!canSubmit}
              className="w-full"
            >
              Submit
            </GameButton>

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
    </FloatingWindow>
  );
};

const TagChip: React.FC<{
  label: string;
  selected: boolean;
  onClick: () => void;
  color: string;
}> = ({ label, selected, onClick, color }) => (
  <GameButton
    emphasis="secondary"
    size="sm"
    selected={selected}
    aria-pressed={selected}
    accent={color}
    onClick={onClick}
  >
    {label}
  </GameButton>
);

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

export default FeedbackWindow;
