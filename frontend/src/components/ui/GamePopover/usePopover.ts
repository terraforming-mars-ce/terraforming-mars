import { useEffect, RefObject } from "react";

export function usePopover({
  isVisible,
  onClose,
  popoverRef,
  anchorRef,
}: {
  isVisible: boolean;
  onClose: () => void;
  popoverRef: RefObject<HTMLElement | null>;
  anchorRef?: RefObject<HTMLElement | null>;
}) {
  useEffect(() => {
    if (!isVisible) {
      return;
    }

    const handleClickOutside = (e: PointerEvent) => {
      const target = e.target as HTMLElement;
      if (target.closest?.("[data-overlay-layer]")) {
        return;
      }
      const outsidePopover = popoverRef.current && !popoverRef.current.contains(target);
      const outsideAnchor = !anchorRef?.current || !anchorRef.current.contains(target);
      if (outsidePopover && outsideAnchor) {
        onClose();
      }
    };

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key !== "Escape" || e.defaultPrevented) {
        return;
      }
      e.preventDefault();
      onClose();
    };

    document.addEventListener("pointerdown", handleClickOutside);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("pointerdown", handleClickOutside);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [isVisible, onClose, popoverRef, anchorRef]);
}
