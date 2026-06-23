import type { Divination, Line } from "./types";

export type CastingLine = Line;

export const CASTING_DISCLAIMER =
  "仅供娱乐和传统文化参考，不构成现实决策建议。";

/** 动画阶段时长（毫秒）；整段约 5–5.5 秒，便于看清每爻 */
export const CASTING_TIMING = {
  gather: 500,
  line: 500,
  primary: 600,
  changed: 600,
  interpret: 500,
} as const;

const POSITION_LABELS = [
  "",
  "第一爻",
  "第二爻",
  "第三爻",
  "第四爻",
  "第五爻",
  "第六爻",
] as const;

export function getPositionLabel(position: number): string {
  return POSITION_LABELS[position] ?? `第${position}爻`;
}

export function getLineLabel(value: number): string {
  switch (value) {
    case 6:
      return "老阴";
    case 7:
      return "少阳";
    case 8:
      return "少阴";
    case 9:
      return "老阳";
    default:
      return "未知";
  }
}

export function getLineMeaning(value: number): string {
  switch (value) {
    case 6:
      return "阴爻，动爻";
    case 7:
      return "阳爻";
    case 8:
      return "阴爻";
    case 9:
      return "阳爻，动爻";
    default:
      return "";
  }
}

/** 三枚硬币阳面（字）数量：0 头=老阴，3 头=老阳 */
export function getCoinYangCount(value: number): number {
  switch (value) {
    case 6:
      return 0;
    case 7:
      return 1;
    case 8:
      return 2;
    case 9:
      return 3;
    default:
      return 0;
  }
}

export function sortLinesBottomToTop(lines: CastingLine[]): CastingLine[] {
  return [...lines].sort((a, b) => a.position - b.position);
}

export function hasCastingLines(divination: Divination | null | undefined): boolean {
  return Boolean(divination?.lines?.length);
}

export function estimateCastingDurationMs(lineCount: number): number {
  const lines = Math.max(0, Math.min(lineCount, 6));
  return (
    CASTING_TIMING.gather +
    lines * CASTING_TIMING.line +
    CASTING_TIMING.primary +
    CASTING_TIMING.changed +
    CASTING_TIMING.interpret
  );
}
