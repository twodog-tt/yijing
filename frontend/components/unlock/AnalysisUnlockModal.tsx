"use client";

import { ensureSession, unlockAnalysis } from "@/lib/api";
import { getSessionKey } from "@/lib/session";
import { useEffect, useState } from "react";

interface AnalysisUnlockModalProps {
  analysisId: number;
  open: boolean;
  onClose: () => void;
  onSuccess: (fullContent: string) => void;
  onLoadingChange?: (loading: boolean) => void;
}

export default function AnalysisUnlockModal({
  analysisId,
  open,
  onClose,
  onSuccess,
  onLoadingChange,
}: AnalysisUnlockModalProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect -- reset stale modal errors on reopen
    setError("");
  }, [open]);

  if (!open) return null;

  async function handleConfirm() {
    setLoading(true);
    onLoadingChange?.(true);
    setError("");
    try {
      await ensureSession(getSessionKey());
      const result = await unlockAnalysis(
        analysisId,
        getSessionKey(),
        "rewarded_video_mock"
      );
      const content = String(result.full_content || "").trim();
      if (!content) {
        throw new Error("完整报告暂未返回，请稍后重新加载。");
      }
      onSuccess(content);
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "解锁失败，请重试");
    } finally {
      setLoading(false);
      onLoadingChange?.(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/40 p-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] sm:items-center sm:p-4"
      role="dialog"
      aria-modal
      aria-labelledby="analysis-unlock-title"
    >
      <div className="w-full max-w-md rounded-2xl bg-white p-5 shadow-xl sm:p-6">
        <h2 id="analysis-unlock-title" className="text-lg font-bold text-stone-900">
          解锁完整报告
        </h2>
        <p className="mt-3 text-sm leading-relaxed text-stone-600">
          Web 内测阶段不会播放真实广告，点击确认后视为解锁成功。完整报告仍基于简化规则，仅供传统文化学习与自我反思参考。
        </p>

        {error && (
          <p className="mt-3 text-sm text-red-600" role="alert">
            {error}
          </p>
        )}

        <div className="mt-6 flex gap-3">
          <button
            type="button"
            onClick={onClose}
            disabled={loading}
            className="min-h-12 flex-1 rounded-xl border border-stone-300 py-3 text-sm font-medium text-stone-700"
          >
            取消
          </button>
          <button
            type="button"
            onClick={handleConfirm}
            disabled={loading}
            className="min-h-12 flex-1 rounded-xl bg-stone-900 py-3 text-sm font-semibold text-white disabled:opacity-60"
          >
            {loading ? "解锁中…" : "确认解锁"}
          </button>
        </div>
      </div>
    </div>
  );
}
