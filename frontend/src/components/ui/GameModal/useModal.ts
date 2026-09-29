import { useEffect, useRef, type RefObject } from "react";

interface UseModalOptions {
  modalRef: RefObject<HTMLDivElement | null>;
  isVisible: boolean;
  onClose: () => void;
  closeOnEscape?: boolean;
  lockScroll?: boolean;
  preventClose?: boolean;
  onPreventedClose?: () => void;
}

export function useModal({
  modalRef,
  isVisible,
  onClose,
  closeOnEscape = true,
  lockScroll = true,
  preventClose = false,
  onPreventedClose,
}: UseModalOptions) {
  const callbacks = useRef({ onClose, onPreventedClose });
  callbacks.current = { onClose, onPreventedClose };
  useEffect(() => {
    if (!isVisible) return;

    const previousFocus = document.activeElement;
    const modal = modalRef.current;
    const focusable = () =>
      Array.from(
        modal?.querySelectorAll<HTMLElement>(
          'button:not(:disabled), a[href], input:not(:disabled), select:not(:disabled), textarea:not(:disabled), [tabindex="0"]',
        ) ?? [],
      ).filter((element) => element.getClientRects().length > 0);
    (focusable()[0] ?? modal)?.focus();
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Tab") {
        const elements = focusable();
        const first = elements[0];
        const last = elements[elements.length - 1];
        if (!first) {
          e.preventDefault();
          modal?.focus();
        } else if (
          e.shiftKey &&
          (document.activeElement === first || document.activeElement === modal)
        ) {
          e.preventDefault();
          last.focus();
        } else if (!e.shiftKey && document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
      if (e.key === "Escape") {
        if (preventClose) {
          callbacks.current.onPreventedClose?.();
        } else if (closeOnEscape) {
          callbacks.current.onClose();
        }
      }
    };

    document.addEventListener("keydown", handleEscape);

    const previousOverflow = document.body.style.overflow;
    if (lockScroll) {
      document.body.style.overflow = "hidden";
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
      if (previousFocus instanceof HTMLElement && previousFocus.isConnected) {
        previousFocus.focus();
      }
      if (lockScroll) {
        document.body.style.overflow = previousOverflow;
      }
    };
  }, [isVisible, modalRef, closeOnEscape, lockScroll, preventClose]);
}
