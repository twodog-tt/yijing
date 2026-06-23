"use client";

import SharePoster, { type SharePosterData } from "./SharePoster";
import { useCallback, useRef, useState } from "react";

interface SharePosterModalProps {
  open: boolean;
  onClose: () => void;
  data: SharePosterData | null;
}

export default function SharePosterModal({
  open,
  onClose,
  data,
}: SharePosterModalProps) {
  const posterRef = useRef<HTMLDivElement>(null);
  const [downloading, setDownloading] = useState(false);
  const [error, setError] = useState("");

  const handleDownload = useCallback(async () => {
    if (!posterRef.current) return;
    setDownloading(true);
    setError("");
    try {
      const { toPng } = await import("html-to-image");
      const dataUrl = await toPng(posterRef.current, {
        pixelRatio: 2,
        cacheBust: true,
      });
      const link = document.createElement("a");
      link.download = `yijing-poster-${data?.divination.id ?? "share"}.png`;
      link.href = dataUrl;
      link.click();
    } catch {
      setError("海报生成失败，请稍后重试或截图保存预览。");
    } finally {
      setDownloading(false);
    }
  }, [data]);

  if (!open || !data) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/50 p-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] sm:items-center sm:p-4"
      role="dialog"
      aria-modal
      aria-labelledby="poster-title"
    >
      <div className="max-h-[calc(100dvh-1.5rem)] w-full max-w-md overflow-y-auto rounded-2xl bg-white p-4 shadow-xl sm:max-h-[90vh] sm:p-6">
        <h2 id="poster-title" className="text-lg font-bold text-stone-900">
          分享海报预览
        </h2>
        <p className="mt-1 text-sm text-stone-500">
          可保存图片分享至朋友圈、小红书或微信群。
        </p>

        <div className="mt-5 flex min-w-0 justify-center py-2">
          <SharePoster ref={posterRef} data={data} />
        </div>

        {error && (
          <p className="mt-3 text-sm text-red-600" role="alert">
            {error}
          </p>
        )}

        <div className="mt-6 flex gap-3">
          <button
            type="button"
            onClick={onClose}
            className="min-h-12 flex-1 rounded-xl border border-stone-300 py-3 text-sm font-medium text-stone-700"
          >
            关闭
          </button>
          <button
            type="button"
            onClick={handleDownload}
            disabled={downloading}
            className="min-h-12 flex-1 rounded-xl bg-stone-900 py-3 text-sm font-semibold text-white disabled:opacity-60"
          >
            {downloading ? "生成中…" : "下载海报"}
          </button>
        </div>
      </div>
    </div>
  );
}
