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
    if (!isVisible) return;

    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      if (target.closest?.("[data-overlay-layer]")) return;
      const outsidePopover = popoverRef.current && !popoverRef.current.contains(target);
      const outsideAnchor = !anchorRef?.current || !anchorRef.current.contains(target);
      if (outsidePopover && outsideAnchor) onClose();
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isVisible, onClose, popoverRef, anchorRef]);
}
