"use client";

import AnalysisFullContent from "@/components/analysis/AnalysisFullContent";
import BaziResultCards from "@/components/analysis/BaziResultCards";
import QimenResultCards from "@/components/analysis/QimenResultCards";
import ElementOrbit from "@/components/motion/ElementOrbit";
import QimenScanGrid from "@/components/motion/QimenScanGrid";
import SectionReveal from "@/components/motion/SectionReveal";
import FreeInterpretationCard from "@/components/interpretation/FreeInterpretationCard";
import AnalysisUnlockModal from "@/components/unlock/AnalysisUnlockModal";
import ErrorAlert from "@/components/ui/ErrorAlert";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import { deleteAnalysis, ensureSession, getAnalysis, unlockAnalysis } from "@/lib/api";
import { buildAnalysisView, MODULE_BAZI_LABEL } from "@/lib/bazi";
import {
  buildQimenView,
  isQimenRecord,
  MODULE_QIMEN_LABEL,
} from "@/lib/qimen";
import { getSessionKey } from "@/lib/session";
import { ModuleTypeBazi, ModuleTypeQimen, type AnalysisRecord } from "@/lib/types";
import { formatDateTime } from "@/lib/utils";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";

export default function AnalysisResultPage() {
  const params = useParams();
  const router = useRouter();
  const id = Number(params.id);
  const fullSectionRef = useRef<HTMLDivElement>(null);

  const [record, setRecord] = useState<AnalysisRecord | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [fullContent, setFullContent] = useState("");
  const [unlockModalOpen, setUnlockModalOpen] = useState(false);
  const [unlocking, setUnlocking] = useState(false);
  const [repairing, setRepairing] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState("");
  const [shouldScrollToFull, setShouldScrollToFull] = useState(false);
  const [fullRevealActive, setFullRevealActive] = useState(false);

  const loadData = useCallback(async () => {
    if (!id || Number.isNaN(id)) {
      setError("无效的记录 ID");
      setLoading(false);
      return;
    }

    setLoading(true);
    setError("");
    setDeleteError("");
    try {
      await ensureSession(getSessionKey());
      const data = await getAnalysis(id, getSessionKey());
      setRecord(data);
      const unlocked = data.unlock_status === 1;
      const content = unlocked ? String(data.full_content || "").trim() : "";
      setFullContent(content);
      setFullRevealActive(unlocked && Boolean(content));
    } catch (e) {
      setError(e instanceof Error ? e.message : "加载失败");
      setRecord(null);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- initial fetch by record id
    void loadData();
  }, [loadData]);

  useEffect(() => {
    if (shouldScrollToFull && fullContent && fullSectionRef.current) {
      const timer = setTimeout(() => {
        fullSectionRef.current?.scrollIntoView({
          behavior: "smooth",
          block: "start",
        });
        setShouldScrollToFull(false);
      }, 300);
      return () => clearTimeout(timer);
    }
  }, [shouldScrollToFull, fullContent]);

  function handleUnlockSuccess(content: string) {
    setFullContent(content);
    setFullRevealActive(true);
    setShouldScrollToFull(true);
    if (record) {
      setRecord({ ...record, unlock_status: 1, full_content: content });
    }
  }

  async function handleRepairFullReport() {
    if (
      !record ||
      record.unlock_status !== 1 ||
      Boolean(fullContent) ||
      repairing ||
      deleting
    ) {
      return;
    }

    setRepairing(true);
    setError("");
    try {
      await ensureSession(getSessionKey());
      const result = await unlockAnalysis(record.id, getSessionKey());
      const content = String(result.full_content || "").trim();
      if (!content) {
        throw new Error("完整报告暂未返回，请稍后重新加载。");
      }
      handleUnlockSuccess(content);
    } catch (e) {
      setError(e instanceof Error ? e.message : "重新获取完整报告失败");
    } finally {
      setRepairing(false);
    }
  }

  async function handleDelete() {
    if (deleting || !record) return;
    const confirmed = window.confirm("删除后不可恢复，是否确认删除？");
    if (!confirmed) return;

    setDeleting(true);
    setDeleteError("");
    try {
      await ensureSession(getSessionKey());
      await deleteAnalysis(record.id, getSessionKey());
      if (record.module_type === ModuleTypeBazi) {
        router.push("/bazi");
      } else if (record.module_type === ModuleTypeQimen) {
        router.push("/qimen");
      } else {
        router.push("/");
      }
    } catch (e) {
      setDeleteError(e instanceof Error ? e.message : "删除失败，请重试");
    } finally {
      setDeleting(false);
    }
  }

  if (loading) {
    return (
      <main className="mx-auto w-full max-w-2xl flex-1 px-4 py-10">
        <LoadingSpinner label="加载简析结果…" />
      </main>
    );
  }

  if (error && !record) {
    return (
      <main className="mx-auto w-full max-w-2xl flex-1 px-4 py-10">
        <ErrorAlert message={error} onRetry={loadData} />
        <Link href="/" className="mt-4 inline-block text-sm text-amber-800">
          返回首页
        </Link>
      </main>
    );
  }

  if (!record) return null;

  const isBazi = record.module_type === ModuleTypeBazi;
  const isQimen = isQimenRecord(record);
  const moduleLabel = isBazi ? MODULE_BAZI_LABEL : MODULE_QIMEN_LABEL;
  const pageTitle = isBazi
    ? "基于传统干支文化的性格与五行倾向学习"
    : "基于传统奇门文化的局势梳理与行动节奏参考";
  const backHref = isBazi ? "/bazi" : "/qimen";
  const baziView = isBazi ? buildAnalysisView(record) : null;
  const qimenView = isQimen ? buildQimenView(record) : null;
  const freeContent = baziView?.freeContent || qimenView?.freeContent || "";
  const isUnlocked = record.unlock_status === 1;
  const hasFull = Boolean(fullContent);
  const resultReady = true;
  const freeRevealDelay = isBazi ? 880 : 560;
  const unlockRevealDelay = isBazi ? 960 : 640;

  if (!isBazi && !isQimen) {
    return (
      <main className="mx-auto w-full max-w-2xl flex-1 px-4 py-10">
        <ErrorAlert message="暂不支持该类型的简析记录。" />
        <Link href="/" className="mt-4 inline-block text-sm text-amber-800">
          返回首页
        </Link>
      </main>
    );
  }

  return (
    <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-5 px-4 py-8 sm:py-10">
      <Link
        href={backHref}
        className="-ml-3 inline-flex min-h-11 items-center rounded-lg px-3 text-sm text-amber-800 hover:underline"
      >
        ← 返回{moduleLabel}
      </Link>

      <SectionReveal active={resultReady} delay={0}>
        <header className="relative overflow-hidden">
          {isBazi && <div className="header-aura header-aura--subtle" aria-hidden />}
          {isQimen && <QimenScanGrid />}
          <p className="text-sm tracking-[0.15em] text-amber-800">{moduleLabel}</p>
          <h1 className="mt-2 text-2xl font-bold text-stone-900">{pageTitle}</h1>
          <div className="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-xs text-stone-500">
            <span>{record.algorithm_version}</span>
            {qimenView?.categoryLabel && <span>{qimenView.categoryLabel}</span>}
            {qimenView?.timeBucketLabel && (
              <span>{qimenView.timeBucketLabel}</span>
            )}
            <span>{formatDateTime(record.created_at)}</span>
          </div>
          {isBazi && <ElementOrbit />}
        </header>
      </SectionReveal>

      {error && <ErrorAlert message={error} onRetry={loadData} />}

      {baziView && <BaziResultCards view={baziView} revealed={resultReady} />}
      {qimenView && <QimenResultCards view={qimenView} revealed={resultReady} />}

      {freeContent && (
        <SectionReveal active={resultReady} delay={freeRevealDelay}>
          <FreeInterpretationCard content={freeContent} />
        </SectionReveal>
      )}

      {!isUnlocked && !hasFull && (
        <SectionReveal active={resultReady} delay={unlockRevealDelay}>
          <div className="space-y-2">
            <button
              type="button"
              onClick={() => setUnlockModalOpen(true)}
              disabled={deleting || unlocking || repairing}
              className={`w-full rounded-xl border border-amber-300 bg-amber-50 py-3.5 text-sm font-semibold text-amber-900 transition hover:bg-amber-100 disabled:opacity-60 ${
                unlocking || unlockModalOpen ? "unlock-glow" : ""
              }`}
            >
              解锁完整报告（Web 内测模拟）
            </button>
            <p className="text-center text-xs leading-relaxed text-stone-500">
              完整报告仍基于简化规则，仅供传统文化学习与自我反思参考。
            </p>
          </div>
        </SectionReveal>
      )}

      {isUnlocked && !hasFull && (
        <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
          <h2 className="text-lg font-bold text-stone-900">完整报告</h2>
          <p className="mt-3 text-sm text-stone-600">
            完整报告内容暂不可用，请尝试重新获取。
          </p>
          <button
            type="button"
            onClick={handleRepairFullReport}
            disabled={repairing || deleting || unlocking}
            className="mt-4 min-h-11 rounded-xl border border-stone-300 bg-white px-4 py-2.5 text-sm font-medium text-stone-800 transition hover:bg-stone-50 disabled:opacity-60"
          >
            {repairing ? "获取中…" : "重新获取完整报告"}
          </button>
        </section>
      )}

      {hasFull && (
        <div ref={fullSectionRef}>
          <AnalysisFullContent content={fullContent} revealed={fullRevealActive} />
        </div>
      )}

      <SectionReveal staticSection>
        <section className="rounded-2xl border border-red-100 bg-red-50/40 p-5">
        <p className="text-xs font-medium uppercase tracking-wide text-red-800">
          危险操作
        </p>
        {deleteError && (
          <div className="mt-3">
            <ErrorAlert message={deleteError} />
          </div>
        )}
        <button
          type="button"
          onClick={handleDelete}
          disabled={deleting || unlockModalOpen || unlocking || repairing}
          className="mt-3 min-h-11 rounded-xl border border-red-300 bg-white px-4 py-2.5 text-sm font-medium text-red-700 transition hover:bg-red-50 disabled:opacity-60"
        >
          {deleting ? "删除中…" : "删除记录"}
        </button>
        </section>
      </SectionReveal>

      <p className="text-center text-xs leading-relaxed text-stone-500">
        内容仅用于传统文化学习、自我反思和行动整理，不构成现实决策建议。
      </p>

      <AnalysisUnlockModal
        analysisId={id}
        open={unlockModalOpen}
        onClose={() => setUnlockModalOpen(false)}
        onSuccess={handleUnlockSuccess}
        onLoadingChange={setUnlocking}
      />
    </main>
  );
}
