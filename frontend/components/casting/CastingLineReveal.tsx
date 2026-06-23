"use client";

import type { CastingLine } from "@/lib/casting";
import {
  getLineLabel,
  getLineMeaning,
  getPositionLabel,
} from "@/lib/casting";
import CoinTossVisual from "./CoinTossVisual";

interface CastingLineRevealProps {
  line: CastingLine;
  active?: boolean;
}

function LineBar({ line }: { line: CastingLine }) {
  const moving = line.is_moving === 1;
  const isYang = line.is_yang === 1;

  if (isYang) {
    return (
      <div className="flex flex-1 items-center gap-2">
        <div
          className={`h-2 w-full rounded-sm ${
            moving ? "bg-amber-500 ring-2 ring-amber-200" : "bg-stone-800"
          }`}
        />
        {moving && (
          <span className="shrink-0 rounded-md bg-amber-100 px-1.5 py-0.5 text-[10px] font-semibold text-amber-800">
            动
          </span>
        )}
      </div>
    );
  }

  return (
    <div className="flex flex-1 items-center gap-2">
      <div className="flex w-full gap-1.5 sm:gap-2">
        <div
          className={`h-2 flex-1 rounded-sm ${
            moving ? "bg-amber-500 ring-2 ring-amber-200" : "bg-stone-800"
          }`}
        />
        <div
          className={`h-2 flex-1 rounded-sm ${
            moving ? "bg-amber-500 ring-2 ring-amber-200" : "bg-stone-800"
          }`}
        />
      </div>
      {moving && (
        <span className="shrink-0 rounded-md bg-amber-100 px-1.5 py-0.5 text-[10px] font-semibold text-amber-800">
          动
        </span>
      )}
    </div>
  );
}

export default function CastingLineReveal({
  line,
  active = false,
}: CastingLineRevealProps) {
  return (
    <div
      className={`rounded-xl border px-3 py-3 transition-all duration-300 sm:px-4 ${
        active
          ? "border-amber-300 bg-amber-50/90 shadow-sm"
          : "border-transparent bg-transparent"
      }`}
    >
      <div className="flex items-center justify-between gap-2">
        <p className="text-sm font-semibold text-amber-100">
          {getPositionLabel(line.position)}
        </p>
        <p className="text-xs text-amber-200/80">
          {getLineLabel(line.value)} · {getLineMeaning(line.value)}
        </p>
      </div>
      <div className="mt-3">
        <CoinTossVisual lineValue={line.value} animating={active} />
      </div>
      <div className="mt-3 flex items-center gap-2">
        <span className="w-8 shrink-0 text-right text-[10px] text-stone-400">
          {line.position}爻
        </span>
        <LineBar line={line} />
      </div>
    </div>
  );
}
