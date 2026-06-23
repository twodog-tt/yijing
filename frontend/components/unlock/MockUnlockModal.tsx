"use client";

import { unlockDivination } from "@/lib/api";
import { getSessionKey } from "@/lib/session";
import type { FullReport } from "@/lib/types";
import { useState } from "react";

interface MockUnlockModalProps {
  divinationId: number;
  open: boolean;
  onClose: () => void;
  onSuccess: (report: FullReport | string) => void;
}

export default function MockUnlockModal({
  divinationId,
  open,
  onClose,
  onSuccess,
}: MockUnlockModalProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  if (!open) return null;

  async function handleConfirm() {
    setLoading(true);
    setError("");
    try {
      const result = await unlockDivination(
        divinationId,
        getSessionKey(),
        "mock_ad"
      );
      onSuccess(result.full_interpretation);
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "解锁失败，请重试");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/40 p-4 sm:items-center"
      role="dialog"
      aria-modal
      aria-labelledby="unlock-title"
    >
      <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl">
        <h2 id="unlock-title" className="text-lg font-bold text-stone-900">
          模拟观看激励广告
        </h2>
        <p className="mt-3 text-sm leading-relaxed text-stone-600">
          本地 MVP 阶段不会播放真实广告，点击确认后视为解锁成功。
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
            className="flex-1 rounded-xl border border-stone-300 py-3 text-sm font-medium text-stone-700"
          >
            取消
          </button>
          <button
            type="button"
            onClick={handleConfirm}
            disabled={loading}
            className="flex-1 rounded-xl bg-stone-900 py-3 text-sm font-semibold text-white disabled:opacity-60"
          >
            {loading ? "解锁中…" : "确认解锁"}
          </button>
        </div>
      </div>
    </div>
  );
}
