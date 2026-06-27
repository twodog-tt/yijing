"use client";

import RecentAnalysisList from "@/components/analysis/RecentAnalysisList";
import ElementOrbit from "@/components/motion/ElementOrbit";
import ErrorAlert from "@/components/ui/ErrorAlert";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import Select from "@/components/ui/Select";
import {
  ApiError,
  createBaziAnalysis,
  ensureSession,
  getAnalysisList,
  getLocalDateString,
  isSensitiveBlockError,
} from "@/lib/api";
import { HOUR_BRANCHES, MODULE_BAZI_LABEL } from "@/lib/bazi";
import { getSessionKey } from "@/lib/session";
import type { AnalysisListItem } from "@/lib/types";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useCallback, useEffect, useState } from "react";

const DISCLAIMER =
  "内容仅用于传统文化学习、自我反思和行动整理，不构成现实决策建议。";

export default function BaziPage() {
  const router = useRouter();
  const [birthDate, setBirthDate] = useState("");
  const [hourBranch, setHourBranch] = useState<string>("");
  const [birthHourUnknown, setBirthHourUnknown] = useState(false);
  const [confirmed, setConfirmed] = useState(false);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [apiError, setApiError] = useState("");
  const [fieldError, setFieldError] = useState("");
  const [recentList, setRecentList] = useState<AnalysisListItem[]>([]);
  const [listLoading, setListLoading] = useState(false);
  const [listError, setListError] = useState("");

  const maxDate = getLocalDateString();

  const loadRecentList = useCallback(async () => {
    setListLoading(true);
    setListError("");
    try {
      await ensureSession(getSessionKey());
      const result = await getAnalysisList(getSessionKey(), "bazi");
      setRecentList(result.items || []);
    } catch (e) {
      setListError(e instanceof Error ? e.message : "加载记录失败");
      setRecentList([]);
    } finally {
      setListLoading(false);
    }
  }, []);

  useEffect(() => {
    async function init() {
      try {
        await loadRecentList();
      } catch (e) {
        setApiError(e instanceof Error ? e.message : "加载失败");
      } finally {
        setLoading(false);
      }
    }
    init();
  }, [loadRecentList]);

  useEffect(() => {
    function handlePageShow(event: PageTransitionEvent) {
      if (event.persisted) {
        void loadRecentList();
      }
    }
    window.addEventListener("pageshow", handlePageShow);
    return () => window.removeEventListener("pageshow", handlePageShow);
  }, [loadRecentList]);

  function validate(): string {
    if (!birthDate) return "请选择出生日期";
    if (!birthHourUnknown && !hourBranch) {
      return "请选择出生时辰，或勾选「时辰未知」";
    }
    if (!confirmed) return "请勾选免责声明";
    return "";
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const msg = validate();
    if (msg) {
      setFieldError(msg);
      return;
    }
    setFieldError("");
    setApiError("");
    setSubmitting(true);

    try {
      const record = await createBaziAnalysis({
        session_key: getSessionKey(),
        birth_date: birthDate,
        birth_hour_unknown: birthHourUnknown,
        birth_hour_branch: birthHourUnknown ? undefined : hourBranch,
        confirm_disclaimer: true,
      });
      router.push(`/analysis/${record.id}`);
    } catch (err) {
      if (isSensitiveBlockError(err)) {
        setApiError((err as ApiError).message);
      } else {
        setApiError(err instanceof Error ? err.message : "提交失败，请重试");
      }
    } finally {
      setSubmitting(false);
    }
  }

  if (loading) {
    return (
      <main className="mx-auto w-full max-w-2xl flex-1 px-4 py-10">
        <LoadingSpinner label="准备八字简析…" />
      </main>
    );
  }

  return (
    <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col px-4 py-8 sm:py-10">
      <div className="relative mb-6 overflow-hidden">
        <div className="header-aura" aria-hidden />
        <Link
          href="/"
          className="-ml-3 inline-flex min-h-11 items-center rounded-lg px-3 text-sm text-amber-800 hover:underline"
        >
          ← 返回首页
        </Link>
        <p className="mt-4 text-sm tracking-[0.15em] text-amber-800">
          八字简析
        </p>
        <h1 className="mt-2 text-2xl font-bold text-stone-900">
          基于传统干支文化的性格与五行倾向学习
        </h1>
        <p className="mt-2 text-sm leading-relaxed text-stone-600">
          本功能采用简化干支文化规则，仅用于传统文化学习和自我反思，不等同于专业八字排盘，也不构成现实决策依据。
        </p>
        <ElementOrbit />
      </div>

      {apiError && (
        <div className="mb-4">
          <ErrorAlert message={apiError} />
        </div>
      )}

      <form
        onSubmit={handleSubmit}
        className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm sm:p-6"
      >
        <label className="block">
          <span className="text-sm font-medium text-stone-800">出生日期</span>
          <input
            type="date"
            value={birthDate}
            min="1900-01-01"
            max={maxDate}
            onChange={(e) => {
              setBirthDate(e.target.value);
              setFieldError("");
            }}
            disabled={submitting}
            className="mt-2 w-full rounded-xl border border-stone-300 px-4 py-3 text-base text-stone-900 outline-none focus:border-amber-600 disabled:opacity-60 sm:text-sm"
          />
        </label>

        <label className="mt-4 flex items-start gap-3">
          <input
            type="checkbox"
            checked={birthHourUnknown}
            onChange={(e) => {
              setBirthHourUnknown(e.target.checked);
              if (e.target.checked) setHourBranch("");
              setFieldError("");
            }}
            disabled={submitting}
            className="mt-0.5 h-5 w-5 shrink-0 rounded border-stone-300"
          />
          <span className="text-sm leading-relaxed text-stone-600">
            时辰未知（本次不生成时柱）
          </span>
        </label>

        {!birthHourUnknown && (
          <label className="mt-5 block">
            <span className="text-sm font-medium text-stone-800">出生时辰</span>
            <Select
              value={hourBranch}
              onChange={(v) => {
                setHourBranch(String(v));
                setFieldError("");
              }}
              options={HOUR_BRANCHES.map((item) => ({
                value: item.value,
                label: item.label,
              }))}
              placeholder="请选择出生时辰"
              disabled={submitting}
            />
          </label>
        )}

        <label className="mt-4 flex items-start gap-3">
          <input
            type="checkbox"
            checked={confirmed}
            onChange={(e) => setConfirmed(e.target.checked)}
            disabled={submitting}
            className="mt-0.5 h-5 w-5 shrink-0 rounded border-stone-300"
          />
          <span className="text-sm leading-relaxed text-stone-600">
            我理解本内容仅用于传统文化学习和自我反思，不作为现实决策依据。
          </span>
        </label>

        {fieldError && (
          <p className="mt-3 text-sm text-red-600" role="alert">
            {fieldError}
          </p>
        )}

        <button
          type="submit"
          disabled={submitting}
          className={`mt-6 min-h-12 w-full rounded-xl bg-stone-900 py-3.5 text-sm font-semibold text-white transition hover:bg-stone-800 disabled:opacity-60 ${
            submitting ? "submit-breathe" : ""
          }`}
        >
          {submitting ? "生成中…" : "生成八字简析"}
        </button>
      </form>

      <RecentAnalysisList
        items={recentList}
        loading={listLoading}
        error={listError}
        moduleLabel={MODULE_BAZI_LABEL}
        emptyText="暂无八字简析记录，提交后将显示在这里。"
        getSubtitle={() => "五行倾向与行动整理"}
        onRetry={loadRecentList}
      />

      <p className="mt-6 text-center text-xs leading-relaxed text-stone-500">
        {DISCLAIMER}
      </p>
    </main>
  );
}
