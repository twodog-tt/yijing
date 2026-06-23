"use client";

import CastingOverlay from "@/components/casting/CastingOverlay";
import ErrorAlert from "@/components/ui/ErrorAlert";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import { createSession, getDailyFortuneToday } from "@/lib/api";
import { getSessionKey } from "@/lib/session";
import type { Divination } from "@/lib/types";
import { sleep } from "@/lib/utils";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";

export default function TodayPage() {
  const router = useRouter();
  const [pageLoading, setPageLoading] = useState(true);
  const [castingOpen, setCastingOpen] = useState(false);
  const [castingWaiting, setCastingWaiting] = useState(false);
  const [castingDivination, setCastingDivination] =
    useState<Divination | null>(null);
  const [pendingId, setPendingId] = useState<number | null>(null);
  const [error, setError] = useState("");
  const [hint, setHint] = useState("");
  const [ready, setReady] = useState(false);

  useEffect(() => {
    async function init() {
      try {
        await createSession(getSessionKey());
        setReady(true);
      } catch (e) {
        setError(e instanceof Error ? e.message : "初始化失败");
      } finally {
        setPageLoading(false);
      }
    }
    init();
  }, []);

  const handleCastingComplete = useCallback(() => {
    const id = pendingId ?? castingDivination?.id;
    if (id) {
      router.push(`/divination/${id}`);
    }
  }, [castingDivination?.id, pendingId, router]);

  async function handleStart() {
    setCastingOpen(true);
    setCastingWaiting(true);
    setCastingDivination(null);
    setPendingId(null);
    setError("");
    setHint("");

    try {
      const result = await getDailyFortuneToday(getSessionKey());

      if (result.daily_fortune.is_existing) {
        setCastingOpen(false);
        setCastingWaiting(false);
        setHint("你今天已经起过一卦，正在为你打开今日运势。");
        await sleep(800);
        router.push(`/divination/${result.divination.id}`);
        return;
      }

      setPendingId(result.divination.id);
      setCastingDivination(result.divination);
      setCastingWaiting(false);
    } catch (e) {
      setCastingOpen(false);
      setCastingWaiting(false);
      setCastingDivination(null);
      setError(e instanceof Error ? e.message : "获取今日运势失败");
    }
  }

  if (pageLoading && !error) {
    return (
      <main className="mx-auto w-full max-w-2xl flex-1 px-4 py-10">
        <LoadingSpinner label="准备今日运势…" />
      </main>
    );
  }

  const busy = castingOpen;

  return (
    <>
      <CastingOverlay
        open={castingOpen}
        waiting={castingWaiting}
        divination={castingDivination}
        onComplete={handleCastingComplete}
      />

      <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col px-4 py-8 sm:py-10">
        <div className="mb-6">
          <Link href="/" className="text-sm text-amber-800 hover:underline">
            ← 返回首页
          </Link>
          <h1 className="mt-4 text-2xl font-bold text-stone-900">今日一卦</h1>
          <p className="mt-2 text-sm leading-relaxed text-stone-600">
            用传统易经卦象，整理今天的状态、节奏与行动提醒。
          </p>
          <p className="mt-2 text-xs text-stone-500">
            每天只生成一次，重复进入会查看当天已有结果。
          </p>
        </div>

        {error && (
          <div className="mb-4">
            <ErrorAlert message={error} />
          </div>
        )}

        {hint && (
          <p className="mb-4 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
            {hint}
          </p>
        )}

        <section className="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm">
          <p className="text-sm leading-relaxed text-stone-700">
            今日一卦通过三枚硬币法起卦，帮你整理：
          </p>
          <ul className="mt-3 list-inside list-disc space-y-1 text-sm text-stone-600">
            <li>今日整体状态</li>
            <li>今日适合的节奏</li>
            <li>今日需要注意的风险</li>
            <li>今日行动建议</li>
            <li>今日情绪提醒</li>
          </ul>

          <button
            type="button"
            onClick={handleStart}
            disabled={busy || !ready}
            className="mt-6 w-full rounded-xl bg-stone-900 py-3.5 text-sm font-semibold text-white transition hover:bg-stone-800 disabled:opacity-60"
          >
            {busy ? "起卦中…" : "查看今日一卦"}
          </button>
        </section>

        <p className="mt-6 text-center text-xs leading-relaxed text-stone-500">
          仅供娱乐和传统文化参考，不构成医疗、法律、投资或决策建议。
        </p>
      </main>
    </>
  );
}
