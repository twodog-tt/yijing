"use client";

import HexagramCard from "@/components/divination/HexagramCard";
import LineChart from "@/components/divination/LineChart";
import ResultSummaryCard from "@/components/divination/ResultSummaryCard";
import FreeInterpretationCard from "@/components/interpretation/FreeInterpretationCard";
import FullInterpretation from "@/components/interpretation/FullInterpretation";
import SharePosterModal from "@/components/share/SharePosterModal";
import type { SharePosterData } from "@/components/share/SharePoster";
import MockUnlockModal from "@/components/unlock/MockUnlockModal";
import ErrorAlert from "@/components/ui/ErrorAlert";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import {
  getDivination,
  getFreeInterpretation,
  getFullInterpretation,
  isNotUnlockedError,
} from "@/lib/api";
import { getSessionKey } from "@/lib/session";
import type { Divination, FullReport } from "@/lib/types";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";

export default function DivinationResultPage() {
  const params = useParams();
  const id = Number(params.id);
  const fullSectionRef = useRef<HTMLDivElement>(null);

  const [divination, setDivination] = useState<Divination | null>(null);
  const [freeContent, setFreeContent] = useState("");
  const [fullReport, setFullReport] = useState<FullReport | string | null>(
    null
  );
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [unlockModalOpen, setUnlockModalOpen] = useState(false);
  const [posterModalOpen, setPosterModalOpen] = useState(false);
  const [loadingFull, setLoadingFull] = useState(false);
  const [shouldScrollToFull, setShouldScrollToFull] = useState(false);

  const loadData = useCallback(async () => {
    if (!id || Number.isNaN(id)) {
      setError("无效的记录 ID");
      setLoading(false);
      return;
    }

    setLoading(true);
    setError("");
    try {
      const [detail, free] = await Promise.all([
        getDivination(id),
        getFreeInterpretation(id),
      ]);
      setDivination(detail);
      setFreeContent(
        free.free_content || detail.free_interpretation || ""
      );

      if (detail.unlock_status === 1) {
        setLoadingFull(true);
        try {
          const res = await getFullInterpretation(id, getSessionKey());
          setFullReport(res.full_content);
        } catch (e) {
          if (!isNotUnlockedError(e)) {
            setError(
              e instanceof Error ? e.message : "加载完整解读失败"
            );
          }
        } finally {
          setLoadingFull(false);
        }
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "加载失败");
    } finally {
      setLoading(false);
    }
  }, [id]);

  async function loadFullInterpretation(scroll = false) {
    setLoadingFull(true);
    try {
      const res = await getFullInterpretation(id, getSessionKey());
      setFullReport(res.full_content);
      if (scroll) setShouldScrollToFull(true);
    } catch (e) {
      if (!isNotUnlockedError(e)) {
        setError(e instanceof Error ? e.message : "加载完整解读失败");
      }
    } finally {
      setLoadingFull(false);
    }
  }

  useEffect(() => {
    loadData();
  }, [loadData]);

  useEffect(() => {
    if (shouldScrollToFull && fullReport && fullSectionRef.current) {
      const timer = setTimeout(() => {
        fullSectionRef.current?.scrollIntoView({
          behavior: "smooth",
          block: "start",
        });
        setShouldScrollToFull(false);
      }, 300);
      return () => clearTimeout(timer);
    }
  }, [shouldScrollToFull, fullReport]);

  function handleUnlockSuccess(report: FullReport | string) {
    setFullReport(report);
    setShouldScrollToFull(true);
    if (divination) {
      setDivination({ ...divination, unlock_status: 1 });
    }
  }

  if (loading) {
    return (
      <main className="mx-auto w-full max-w-2xl flex-1 px-4 py-10">
        <LoadingSpinner label="加载卦象…" />
      </main>
    );
  }

  if (error && !divination) {
    return (
      <main className="mx-auto w-full max-w-2xl flex-1 px-4 py-10">
        <ErrorAlert message={error} onRetry={loadData} />
        <Link href="/ask" className="mt-4 inline-block text-sm text-amber-800">
          返回问事
        </Link>
      </main>
    );
  }

  if (!divination) return null;

  const unlocked = divination.unlock_status === 1;
  const hasFull = fullReport !== null;
  const posterData: SharePosterData = {
    divination,
    freeContent,
    fullReport,
  };

  return (
    <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-5 px-4 py-8 sm:py-10">
      <div className="flex items-center justify-between gap-2">
        <Link href="/history" className="-ml-3 inline-flex min-h-11 items-center rounded-lg px-3 text-sm text-amber-800 hover:underline">
          ← 历史记录
        </Link>
        <button
          type="button"
          onClick={() => setPosterModalOpen(true)}
          className="inline-flex min-h-11 items-center rounded-lg px-2 text-sm font-medium text-stone-600 underline-offset-2 hover:text-amber-800 hover:underline"
        >
          生成分享海报
        </button>
      </div>

      {error && <ErrorAlert message={error} onRetry={loadData} />}

      <ResultSummaryCard divination={divination} />

      <div className="grid gap-4 sm:grid-cols-2">
        <HexagramCard
          title="本卦"
          hexagram={divination.primary_hexagram}
          accent="primary"
        />
        <HexagramCard
          title="变卦"
          hexagram={divination.changed_hexagram}
          accent="changed"
        />
      </div>

      <LineChart lines={divination.lines} title="六爻图（下→上）" />

      <FreeInterpretationCard content={freeContent} />

      {!unlocked && !hasFull && (
        <button
          type="button"
          onClick={() => setUnlockModalOpen(true)}
          className="w-full rounded-xl border border-amber-300 bg-amber-50 py-3.5 text-sm font-semibold text-amber-900 transition hover:bg-amber-100"
        >
          观看广告解锁完整解读（模拟）
        </button>
      )}

      {unlocked && !hasFull && !loadingFull && (
        <button
          type="button"
          onClick={() => loadFullInterpretation(true)}
          className="w-full rounded-xl bg-stone-900 py-3.5 text-sm font-semibold text-white"
        >
          查看完整解读
        </button>
      )}

      {loadingFull && <LoadingSpinner label="加载完整解读…" />}

      {hasFull && (
        <div ref={fullSectionRef} id="full-interpretation" className="scroll-mt-4">
          <h2 className="mb-4 text-lg font-bold text-stone-900">完整解读</h2>
          <FullInterpretation content={fullReport} />
        </div>
      )}

      <MockUnlockModal
        divinationId={id}
        open={unlockModalOpen}
        onClose={() => setUnlockModalOpen(false)}
        onSuccess={handleUnlockSuccess}
      />

      <SharePosterModal
        open={posterModalOpen}
        onClose={() => setPosterModalOpen(false)}
        data={posterData}
      />
    </main>
  );
}
