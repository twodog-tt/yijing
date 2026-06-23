"use client";

import type { CastingLine } from "@/lib/casting";
import { sortLinesBottomToTop } from "@/lib/casting";

interface CastingHexagramPreviewProps {
  lines: CastingLine[];
  revealedCount: number;
  activePosition?: number;
}

function PlaceholderLine() {
  return (
    <div className="flex items-center gap-2 opacity-30">
      <span className="w-8 shrink-0 text-right text-[10px] text-stone-500">
        ·
      </span>
      <div className="h-1.5 flex-1 rounded-sm bg-stone-600/40" />
    </div>
  );
}

function PreviewLine({
  line,
  state,
}: {
  line: CastingLine;
  state: "pending" | "active" | "done";
}) {
  const moving = line.is_moving === 1;
  const isYang = line.is_yang === 1;

  if (state === "pending") {
    return <PlaceholderLine />;
  }

  const barClass =
    state === "active"
      ? moving
        ? "bg-amber-400 ring-1 ring-amber-300"
        : "bg-amber-100"
      : moving
        ? "bg-amber-500"
        : "bg-stone-200";

  return (
    <div
      className={`flex items-center gap-2 rounded-md px-1 py-0.5 transition-all duration-300 ${
        state === "active" ? "bg-amber-500/15" : ""
      }`}
    >
      <span
        className={`w-8 shrink-0 text-right text-[10px] ${
          state === "active" ? "font-semibold text-amber-200" : "text-stone-400"
        }`}
      >
        {line.position}爻
      </span>
      {isYang ? (
        <div className="flex flex-1 items-center gap-1.5">
          <div className={`h-1.5 w-full rounded-sm ${barClass}`} />
          {moving && (
            <span className="text-[9px] font-semibold text-amber-300">动</span>
          )}
        </div>
      ) : (
        <div className="flex flex-1 items-center gap-1.5">
          <div className="flex w-full gap-1">
            <div className={`h-1.5 flex-1 rounded-sm ${barClass}`} />
            <div className={`h-1.5 flex-1 rounded-sm ${barClass}`} />
          </div>
          {moving && (
            <span className="text-[9px] font-semibold text-amber-300">动</span>
          )}
        </div>
      )}
    </div>
  );
}

export default function CastingHexagramPreview({
  lines,
  revealedCount,
  activePosition,
}: CastingHexagramPreviewProps) {
  const sorted = sortLinesBottomToTop(lines);
  const display = [...sorted].reverse();

  return (
    <div className="w-full max-w-xs rounded-2xl border border-stone-600/60 bg-stone-800/50 p-4">
      <p className="mb-3 text-center text-xs text-stone-400">六爻（下 → 上）</p>
      <div className="flex flex-col gap-2.5">
        {display.map((line) => {
          let state: "pending" | "active" | "done" = "pending";
          if (line.position <= revealedCount) {
            state = line.position === activePosition ? "active" : "done";
          }
          return (
            <PreviewLine key={line.position} line={line} state={state} />
          );
        })}
      </div>
    </div>
  );
}
