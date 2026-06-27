"use client";

import { forwardRef, useEffect, useState } from "react";
import type { Divination, FullReport } from "@/lib/types";
import { buildPosterSummary } from "@/components/divination/ResultSummaryCard";
import { getDivinationUrl } from "@/lib/site";
import { formatDateTime, truncateText } from "@/lib/utils";

export interface SharePosterData {
  divination: Divination;
  freeContent: string;
  fullReport: FullReport | string | null;
}

interface SharePosterProps {
  data: SharePosterData;
}

const SharePoster = forwardRef<HTMLDivElement, SharePosterProps>(
  function SharePoster({ data }, ref) {
    const { divination, freeContent, fullReport } = data;
    const [qrDataUrl, setQrDataUrl] = useState<string | null>(null);

    useEffect(() => {
      let cancelled = false;
      const url = getDivinationUrl(divination.id);

      import("qrcode")
        .then((QRCode) =>
          QRCode.toDataURL(url, { width: 112, margin: 1, errorCorrectionLevel: "M" })
        )
        .then((dataUrl) => {
          if (!cancelled) setQrDataUrl(dataUrl);
        })
        .catch(() => {
          if (!cancelled) setQrDataUrl(null);
        });

      return () => {
        cancelled = true;
      };
    }, [divination.id]);

    const summary = buildPosterSummary(
      freeContent,
      fullReport,
      divination.primary_hexagram?.summary
    );
    const movingText =
      divination.moving_lines?.length > 0
        ? `动爻：第 ${divination.moving_lines.join("、")} 爻`
        : "动爻：无";

    const isDailyFortune = divination.category?.name === "今日运势";

    return (
      <div
        ref={ref}
        className="w-full max-w-[320px] overflow-hidden rounded-2xl border border-stone-200 bg-[#faf8f5] shadow-xl"
        style={{ fontFamily: "serif" }}
      >
        <div className="border-b border-amber-200/60 bg-gradient-to-r from-amber-100 to-stone-100 px-6 py-5">
          <p className="text-xs tracking-[0.3em] text-amber-800">文易传统文化</p>
          <p className="mt-1 text-lg font-bold text-stone-900">
            {isDailyFortune ? "今日一卦" : "卦象分享卡"}
          </p>
        </div>

        <div className="space-y-4 px-6 py-5">
          <div>
            <p className="text-xs text-stone-500">
              {isDailyFortune ? "今日主题" : "所问之事"}
            </p>
            <p className="mt-1 text-sm font-medium leading-relaxed text-stone-800">
              {truncateText(divination.question, 30)}
            </p>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="rounded-xl border border-amber-200/80 bg-white p-3">
              <p className="text-xs text-amber-800">本卦</p>
              <p className="mt-1 text-sm font-bold text-stone-900">
                {divination.primary_hexagram?.full_name}
              </p>
            </div>
            <div className="rounded-xl border border-stone-200 bg-white p-3">
              <p className="text-xs text-stone-500">变卦</p>
              <p className="mt-1 text-sm font-bold text-stone-900">
                {divination.changed_hexagram?.full_name}
              </p>
            </div>
          </div>

          <p className="text-xs text-stone-600">{movingText}</p>

          <div className="rounded-xl border border-stone-200 bg-white p-4">
            <p className="text-xs text-stone-500">一句话提示</p>
            <p className="mt-2 text-sm leading-relaxed text-stone-800">
              {truncateText(summary, 80)}
            </p>
          </div>

          <p className="text-xs text-stone-400">
            {formatDateTime(divination.created_at)}
          </p>
        </div>

        <div className="flex items-end justify-between gap-3 border-t border-stone-200 bg-stone-50 px-6 py-4">
          <p className="flex-1 text-[10px] leading-relaxed text-stone-500">
            仅供娱乐和传统文化参考，不构成医疗、法律、投资或决策建议。
          </p>
          <div className="flex h-14 w-14 shrink-0 items-center justify-center overflow-hidden rounded border border-stone-200 bg-white">
            {qrDataUrl ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={qrDataUrl}
                alt="结果页二维码"
                width={56}
                height={56}
                className="h-full w-full object-contain"
              />
            ) : (
              <div className="flex flex-col items-center justify-center text-[9px] text-stone-400">
                <span>二维码</span>
                <span>占位</span>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }
);

export default SharePoster;
