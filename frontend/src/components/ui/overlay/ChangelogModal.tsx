import React, { useEffect, useRef, useState } from "react";
import GameMenuModal from "./GameMenuModal.tsx";
import GameButton from "../buttons/GameButton.tsx";
import { APP_VERSION } from "@/config.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { useBackDismiss } from "@/hooks/useBackDismiss.ts";
import { apiService } from "@/services/apiService.ts";
import { useVersionStore } from "@/stores/versionStore.ts";
import type { ChangelogEntry } from "@/types/generated/api-types.ts";
import { compareVersions, isLocalBuild } from "@/utils/version.ts";

type ChangelogState =
  | { status: "loading" }
  | { status: "error" }
  | { status: "ready"; entries: ChangelogEntry[] };

function InlineText({ text }: { text: string }) {
  return (
    <>
      {text.split("`").map((part, index) =>
        index % 2 === 1 ? (
          <code key={index} className="rounded-sm bg-white/10 px-1 font-mono text-[0.95em]">
            {part}
          </code>
        ) : (
          <React.Fragment key={index}>{part}</React.Fragment>
        ),
      )}
    </>
  );
}

function isNotInstalled(version: string): boolean {
  return !isLocalBuild(APP_VERSION) && compareVersions(version, APP_VERSION) > 0;
}

const EntrySection = React.forwardRef<HTMLElement, { entry: ChangelogEntry }>(function EntrySection(
  { entry },
  ref,
) {
  const notInstalled = isNotInstalled(entry.version);
  return (
    <section
      ref={ref}
      className="scroll-mt-2 border-t border-white/10 py-4 first:border-t-0 first:pt-0"
    >
      <div className="flex items-center justify-between gap-3">
        <h3 className="m-0 font-orbitron text-base font-bold text-white">{entry.version}</h3>
        {entry.version === APP_VERSION && (
          <span className="font-orbitron text-xs text-white/40">Installed</span>
        )}
      </div>
      {notInstalled && (
        <div className="mt-2 flex items-center justify-between gap-3 text-sm text-white/70">
          <span>Not installed yet</span>
          <GameButton size="sm" height={36} onClick={() => window.location.reload()}>
            Update
          </GameButton>
        </div>
      )}
      {entry.intro && (
        <p className="m-0 mt-2 text-sm text-white/70">
          <InlineText text={entry.intro} />
        </p>
      )}
      {entry.sections.map((section) => (
        <div key={section.title} className="mt-3">
          {section.major ? (
            <h4 className="m-0 mb-1 font-orbitron text-sm font-bold text-white">{section.title}</h4>
          ) : (
            <h4 className="m-0 mb-1 font-orbitron text-xs uppercase tracking-wider text-white/50">
              {section.title}
            </h4>
          )}
          {section.intro && (
            <p className="m-0 mb-1 text-sm text-white/85">
              <InlineText text={section.intro} />
            </p>
          )}
          {section.items.length > 0 && (
            <ul className="m-0 list-disc space-y-1 pl-4 text-sm text-white/85">
              {section.items.map((item, index) => (
                <li key={index}>
                  <InlineText text={item} />
                </li>
              ))}
            </ul>
          )}
        </div>
      ))}
    </section>
  );
});

function ChangelogDialog() {
  const focus = useVersionStore((s) => s.changelogFocus);
  const closeChangelog = useVersionStore((s) => s.closeChangelog);
  const [state, setState] = useState<ChangelogState>({ status: "loading" });
  const focusRef = useRef<HTMLElement>(null);
  useBackDismiss(closeChangelog);

  useEffect(() => {
    const controller = new AbortController();
    apiService.getChangelog(controller.signal).then(
      (changelog) => setState({ status: "ready", entries: changelog.entries }),
      () => {
        if (!controller.signal.aborted) {
          setState({ status: "error" });
        }
      },
    );
    return () => controller.abort();
  }, []);

  useEffect(() => {
    if (state.status === "ready") {
      focusRef.current?.scrollIntoView({ block: "start" });
    }
  }, [state.status]);

  let body: React.ReactNode;
  if (state.status === "loading") {
    body = <p className="m-0 text-center text-sm text-white/60">Loading changelog…</p>;
  } else if (state.status === "error") {
    body = <p className="m-0 text-center text-sm text-white/60">Could not load the changelog.</p>;
  } else if (state.entries.length === 0) {
    body = <p className="m-0 text-center text-sm text-white/60">No release notes yet.</p>;
  } else {
    body = state.entries.map((entry) => (
      <EntrySection
        key={entry.version}
        ref={entry.version === focus ? focusRef : undefined}
        entry={entry}
      />
    ));
  }

  return (
    <GameMenuModal
      title="Changelog"
      showBackdrop={true}
      zIndex={Z_INDEX.APP_OVERLAY}
      onClose={closeChangelog}
      showCloseButton={true}
    >
      <div className="text-left">{body}</div>
    </GameMenuModal>
  );
}

const ChangelogModal: React.FC = () => {
  const open = useVersionStore((s) => s.changelogOpen);
  return open ? <ChangelogDialog /> : null;
};

export default ChangelogModal;
