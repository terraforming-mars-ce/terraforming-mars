import { useEffect, useId, useRef, useState } from "react";
import GameButton from "./buttons/GameButton.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";

interface Props {
  id: string;
  label: string;
  value: string;
  options: { id: string; name: string }[];
  disabled?: boolean;
  onChange: (value: string) => void;
}

export default function GameSelect({ id, label, value, options, disabled, onChange }: Props) {
  const listId = useId();
  const root = useRef<HTMLDivElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const list = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(0);
  const selected = options.findIndex((option) => option.id === value);

  useEffect(() => {
    if (!open) {
      return;
    }
    list.current?.focus();
    const dismiss = (event: PointerEvent) => {
      if (!root.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("pointerdown", dismiss);
    return () => document.removeEventListener("pointerdown", dismiss);
  }, [open]);

  useEffect(() => {
    if (open) {
      const container = list.current;
      const option = container?.children[active] as HTMLElement | undefined;
      if (container && option) {
        if (option.offsetTop < container.scrollTop) {
          container.scrollTop = option.offsetTop;
        } else if (
          option.offsetTop + option.offsetHeight >
          container.scrollTop + container.clientHeight
        ) {
          container.scrollTop = option.offsetTop + option.offsetHeight - container.clientHeight;
        }
      }
    }
  }, [open, active]);

  const close = () => {
    setOpen(false);
    trigger.current?.focus();
  };
  const choose = (index: number) => {
    if (options[index]) {
      onChange(options[index].id);
    }
    close();
  };
  const show = () => {
    setActive(Math.max(0, selected));
    setOpen(true);
  };

  return (
    <div
      ref={root}
      className="relative min-w-0 w-60 max-w-full"
      onBlur={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget)) {
          setOpen(false);
        }
      }}
    >
      <GameButton
        ref={trigger}
        id={id}
        emphasis="secondary"
        className="w-full !justify-between text-left"
        aria-label={label}
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={listId}
        disabled={disabled}
        onClick={() => {
          if (open) {
            close();
          } else {
            show();
          }
        }}
        onKeyDown={(event) => {
          if (event.key === "ArrowDown" || event.key === "ArrowUp") {
            event.preventDefault();
            show();
          }
        }}
      >
        <span className="truncate">{options[selected]?.name ?? value}</span>
        <svg aria-hidden="true" width="12" height="8" viewBox="0 0 12 8" fill="none">
          <path d="m1 1 5 5 5-5" stroke="currentColor" strokeWidth="1.5" />
        </svg>
      </GameButton>
      {open && (
        <div
          ref={list}
          id={listId}
          role="listbox"
          aria-label={label}
          tabIndex={-1}
          aria-activedescendant={`${listId}-${active}`}
          className="absolute top-full mt-1 right-0 w-full max-h-64 overflow-y-auto overscroll-contain border border-white/25 bg-[#050506] text-white py-1 menu-enter outline-none"
          style={{ zIndex: Z_INDEX.POPOVER }}
          onKeyDown={(event) => {
            switch (event.key) {
              case "ArrowDown":
                event.preventDefault();
                setActive((index) => Math.min(options.length - 1, index + 1));
                break;
              case "ArrowUp":
                event.preventDefault();
                setActive((index) => Math.max(0, index - 1));
                break;
              case "Home":
                event.preventDefault();
                setActive(0);
                break;
              case "End":
                event.preventDefault();
                setActive(options.length - 1);
                break;
              case "Enter":
              case " ":
                event.preventDefault();
                choose(active);
                break;
              case "Escape":
                event.preventDefault();
                event.stopPropagation();
                close();
                break;
              case "Tab":
                setOpen(false);
                break;
              default: {
                if (event.key.length === 1) {
                  const index = options.findIndex((option) =>
                    option.name.toLowerCase().startsWith(event.key.toLowerCase()),
                  );
                  if (index >= 0) {
                    setActive(index);
                  }
                }
              }
            }
          }}
        >
          {options.map((option, index) => (
            <div
              key={option.id}
              id={`${listId}-${index}`}
              role="option"
              aria-selected={option.id === value}
              className={`px-4 py-3 text-left font-orbitron text-xs cursor-pointer transition-colors ${index === active ? "bg-white/10" : ""} ${option.id === value ? "text-[#a9c0ff]" : "text-white/80"}`}
              onPointerMove={() => setActive(index)}
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => choose(index)}
            >
              {option.name}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
