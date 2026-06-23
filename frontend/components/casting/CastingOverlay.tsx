"use client";

import type { Divination } from "@/lib/types";
import {
  CASTING_DISCLAIMER,
  CASTING_TIMING,
  hasCastingLines,
  sortLinesBottomToTop,
  type CastingLine,
} from "@/lib/casting";
import { useCallback, useEffect, useRef, useState } from "react";
import CastingHexagramPreview from "./CastingHexagramPreview";
import CastingLineReveal from "./CastingLineReveal";

export type CastingOverlayPhase =
  | "waiting"
  | "gather"
  | "line"
  | "primary"
  | "changed"
  | "interpret";

interface CastingOverlayProps {
  open: boolean;
  waiting: boolean;
  divination: Divination | null;
  onComplete: () => void;
}

function getPhaseMessage(
  phase: CastingOverlayPhase,
  divination: Divination | null,
  activeLine: CastingLine | null,
  revealedCount: number
): { title: string; subtitle?: string } {
  switch (phase) {
    case "waiting":
      return { title: "正在准备起卦" };
    case "gather":
      return {
        title: "正在收束你的问题",
        subtitle: "请在心中专注这一件事",
      };
    case "line":
      return {
        title: "正在生成六爻",
        subtitle: activeLine
          ? `第 ${revealedCount} 爻已现`
          : `已生成 ${revealedCount} / 6 爻`,
      };
    case "primary":
      return {
        title: "六爻已成，正在形成本卦",
        subtitle: divination?.primary_hexagram?.full_name
          ? `本卦：${divination.primary_hexagram.full_name}`
          : undefined,
      };
    case "changed": {
      const moving = divination?.moving_lines ?? [];
      if (moving.length > 0) {
        return {
          title: "动爻已现，正在推演变卦",
          subtitle: `动爻：第 ${moving.join("、")} 爻 · 变卦：${divination?.changed_hexagram?.full_name ?? ""}`,
        };
      }
      return {
        title: "本卦无动爻，卦象保持稳定",
      };
    }
    case "interpret":
      return { title: "正在整理卦象解读" };
    default:
      return { title: "正在起卦" };
  }
}

export default function CastingOverlay({
  open,
  waiting,
  divination,
  onComplete,
}: CastingOverlayProps) {
  const [phase, setPhase] = useState<CastingOverlayPhase>("waiting");
  const [revealedCount, setRevealedCount] = useState(0);
  const [activePosition, setActivePosition] = useState<number | undefined>();
  const mountedRef = useRef(true);
  const onCompleteRef = useRef(onComplete);

  useEffect(() => {
    onCompleteRef.current = onComplete;
  }, [onComplete]);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  const runSequence = useCallback(async (data: Divination, isCancelled: () => boolean) => {
    const lines = sortLinesBottomToTop(data.lines ?? []);

    const delays = (ms: number) =>
      new Promise<void>((resolve) => {
        window.setTimeout(() => {
          if (!isCancelled()) resolve();
        }, ms);
      });

    const safeSet = (fn: () => void) => {
      if (!isCancelled() && mountedRef.current) fn();
    };

    if (!hasCastingLines(data) || lines.length === 0) {
      if (!isCancelled()) onCompleteRef.current();
      return;
    }

    safeSet(() => {
      setPhase("gather");
      setRevealedCount(0);
      setActivePosition(undefined);
    });
    await delays(CASTING_TIMING.gather);
    if (isCancelled()) return;

    for (const line of lines) {
      safeSet(() => {
        setPhase("line");
        setActivePosition(line.position);
      });
      await delays(CASTING_TIMING.line * 0.45);
      if (isCancelled()) return;
      safeSet(() => {
        setRevealedCount(line.position);
      });
      await delays(CASTING_TIMING.line * 0.55);
      if (isCancelled()) return;
    }

    safeSet(() => {
      setActivePosition(undefined);
      setPhase("primary");
    });
    await delays(CASTING_TIMING.primary);
    if (isCancelled()) return;

    safeSet(() => setPhase("changed"));
    await delays(CASTING_TIMING.changed);
    if (isCancelled()) return;

    safeSet(() => setPhase("interpret"));
    await delays(CASTING_TIMING.interpret);
    if (isCancelled()) return;

    onCompleteRef.current();
  }, []);

  useEffect(() => {
    if (!open) {
      setPhase("waiting");
      setRevealedCount(0);
      setActivePosition(undefined);
      return;
    }

    if (waiting || !divination) {
      setPhase("waiting");
      return;
    }

    let cancelled = false;
    const isCancelled = () => cancelled || !mountedRef.current;

    void runSequence(divination, isCancelled);

    return () => {
      cancelled = true;
    };
  }, [open, waiting, divination, runSequence]);

  if (!open) return null;

  const lines = sortLinesBottomToTop(divination?.lines ?? []);
  const activeLine =
    activePosition != null
      ? lines.find((l) => l.position === activePosition) ?? null
      : null;
  const { title, subtitle } = getPhaseMessage(
    phase,
    divination,
    activeLine,
    revealedCount
  );

  return (
    <div
      className="fixed inset-0 z-50 flex flex-col bg-stone-900/92 px-4 pb-[max(1.5rem,env(safe-area-inset-bottom))] pt-[max(1.5rem,env(safe-area-inset-top))] backdrop-blur-sm sm:px-6"
      role="status"
      aria-live="polite"
      aria-label="起卦中"
    >
      <div className="mx-auto flex w-full max-w-lg flex-1 flex-col">
        <header className="text-center">
          <p className="text-xs tracking-[0.25em] text-amber-300/80">正在起卦</p>
          <h2 className="mt-2 text-lg font-semibold text-amber-50 sm:text-xl">
            {title}
          </h2>
          {subtitle && (
            <p className="mt-2 text-sm leading-relaxed text-stone-300">
              {subtitle}
            </p>
          )}
        </header>

        <div className="mt-6 flex flex-1 flex-col items-center justify-center gap-5">
          {phase === "waiting" && (
            <div className="flex flex-col items-center gap-4">
              <div className="h-12 w-12 animate-spin rounded-full border-2 border-stone-600 border-t-amber-400" />
              <p className="text-sm text-stone-400">等待卦象数据返回…</p>
            </div>
          )}

          {phase !== "waiting" && lines.length > 0 && (
            <>
              {phase === "line" && activeLine && (
                <div className="w-full max-w-sm">
                  <CastingLineReveal line={activeLine} active />
                </div>
              )}

              <CastingHexagramPreview
                lines={lines}
                revealedCount={revealedCount}
                activePosition={activePosition}
              />
            </>
          )}
        </div>

        <footer className="mt-6 text-center">
          <p className="text-xs leading-relaxed text-stone-500">
            {CASTING_DISCLAIMER}
          </p>
        </footer>
      </div>
    </div>
  );
}
