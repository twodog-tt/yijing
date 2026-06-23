"use client";

import { getCoinYangCount } from "@/lib/casting";

interface CoinTossVisualProps {
  lineValue: number;
  animating?: boolean;
}

function CoinFace({ isYang, spinning }: { isYang: boolean; spinning: boolean }) {
  return (
    <div
      className={`flex h-9 w-9 items-center justify-center rounded-full border-2 text-xs font-semibold shadow-sm transition-all duration-300 sm:h-10 sm:w-10 ${
        spinning ? "animate-bounce" : ""
      } ${
        isYang
          ? "border-amber-300 bg-amber-50 text-amber-900"
          : "border-stone-300 bg-stone-100 text-stone-600"
      }`}
    >
      {isYang ? "字" : "背"}
    </div>
  );
}

export default function CoinTossVisual({
  lineValue,
  animating = false,
}: CoinTossVisualProps) {
  const yangCount = getCoinYangCount(lineValue);
  const faces = [0, 1, 2].map((i) => i < yangCount);

  return (
    <div className="flex items-center justify-center gap-2 sm:gap-3">
      {faces.map((isYang, i) => (
        <CoinFace key={i} isYang={isYang} spinning={animating} />
      ))}
    </div>
  );
}
