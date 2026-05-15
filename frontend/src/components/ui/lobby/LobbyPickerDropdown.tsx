import React, { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Z_INDEX } from "@/constants/zIndex.ts";

export interface LobbyPickerItem {
  id: string;
  name: string;
  description?: string;
}

interface LobbyPickerDropdownProps {
  trigger: (state: { open: boolean; toggle: () => void }) => React.ReactNode;
  items: LobbyPickerItem[];
  selectedId?: string;
  onSelect: (id: string) => void;
  emptyMessage?: string;
  matchTriggerWidth?: boolean;
  minWidth?: number;
}

const LobbyPickerDropdown: React.FC<LobbyPickerDropdownProps> = ({
  trigger,
  items,
  selectedId,
  onSelect,
  emptyMessage = "No options",
  matchTriggerWidth = true,
  minWidth,
}) => {
  const [open, setOpen] = useState(false);
  const anchorRef = useRef<HTMLDivElement>(null);

  const toggle = () => setOpen((v) => !v);
  const close = () => setOpen(false);

  useEffect(() => {
    if (!open) {
      return;
    }
    const handleClick = (e: MouseEvent) => {
      const portalEl = document.getElementById("lobby-picker-dropdown-portal");
      const target = e.target as Node;
      if (anchorRef.current && !anchorRef.current.contains(target) && !portalEl?.contains(target)) {
        close();
      }
    };
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [open]);

  const handleSelect = (id: string) => {
    onSelect(id);
    close();
  };

  const rect = anchorRef.current?.getBoundingClientRect();

  return (
    <>
      <div ref={anchorRef}>{trigger({ open, toggle })}</div>
      {open &&
        rect &&
        createPortal(
          <div
            id="lobby-picker-dropdown-portal"
            className="fixed bg-space-black-darker/95 border border-space-blue-600/50 rounded-lg overflow-hidden shadow-[0_10px_40px_rgba(0,0,0,0.6)] max-h-[300px] overflow-y-auto backdrop-blur-space"
            style={{
              zIndex: Z_INDEX.POPOVER,
              top: rect.bottom + 4,
              left: rect.left,
              width: matchTriggerWidth ? rect.width : undefined,
              minWidth: minWidth ?? (matchTriggerWidth ? undefined : rect.width),
            }}
          >
            {items.length === 0 ? (
              <div className="px-3 py-3 text-white/40 text-xs italic">{emptyMessage}</div>
            ) : (
              items.map((item) => {
                const isSelected = item.id === selectedId;
                return (
                  <button
                    key={item.id}
                    onClick={() => handleSelect(item.id)}
                    className={`w-full flex flex-col gap-0.5 px-3 py-2.5 text-left transition-colors border-b border-white/10 last:border-b-0 cursor-pointer ${
                      isSelected
                        ? "text-white bg-space-blue-800/60"
                        : "text-white/80 hover:bg-white/10"
                    }`}
                  >
                    <div className="flex items-center justify-between w-full gap-2">
                      <span className="font-orbitron text-sm font-medium tracking-wide">
                        {item.name}
                      </span>
                      {isSelected && (
                        <span className="bg-space-blue-900 text-white py-0.5 px-1.5 rounded text-[10px] font-bold uppercase shrink-0">
                          Selected
                        </span>
                      )}
                    </div>
                    {item.description ? (
                      <span className="text-[11px] text-white/40 leading-tight">
                        {item.description}
                      </span>
                    ) : null}
                  </button>
                );
              })
            )}
          </div>,
          document.body,
        )}
    </>
  );
};

export default LobbyPickerDropdown;
