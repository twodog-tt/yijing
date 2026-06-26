"use client";

import RecentAnalysisList from "@/components/analysis/RecentAnalysisList";
import QimenScanGrid from "@/components/motion/QimenScanGrid";
import ErrorAlert from "@/components/ui/ErrorAlert";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import Select from "@/components/ui/Select";
import {
  createQimenAnalysis,
  ensureSession,
  getAnalysisList,
  isSensitiveBlockError,
} from "@/lib/api";
import {
  MAX_QUESTION_LENGTH,
  METHOD_NOTE,
  MIN_QUESTION_LENGTH,
  MODULE_QIMEN_LABEL,
  QIMEN_CATEGORIES,
  listRecordSubtitle,
} from "@/lib/qimen";
import { getSessionKey } from "@/lib/session";
import type { AnalysisListItem } from "@/lib/types";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useCallback, useEffect, useRef, useState } from "react";

const DISCLAIMER =
  "内容仅用于传统文化学习、自我反思和行动整理，不构成现实决策建议。";

export default function QimenPage() {
  const router = useRouter();
  const [category, setCategory] = useState<string>(QIMEN_CATEGORIES[0].value);
  const [question, setQuestion] = useState("");
  const [confirmed, setConfirmed] = useState(false);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [apiError, setApiError] = useState("");
  const [fieldError, setFieldError] = useState("");
  const [recentList, setRecentList] = useState<AnalysisListItem[]>([]);
  const [listLoading, setListLoading] = useState(false);
  const [listError, setListError] = useState("");
  const [categoryPulse, setCategoryPulse] = useState(false);
  const categoryPulseTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const loadRecentList = useCallback(async () => {
    setListLoading(true);
    setListError("");
    try {
      await ensureSession(getSessionKey());
      const result = await getAnalysisList(getSessionKey(), "qimen");
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

  useEffect(() => {
    return () => {
      if (categoryPulseTimer.current) {
        clearTimeout(categoryPulseTimer.current);
      }
    };
  }, []);

  function pulseCategoryField() {
    if (categoryPulseTimer.current) {
      clearTimeout(categoryPulseTimer.current);
    }
    setCategoryPulse(true);
    categoryPulseTimer.current = setTimeout(() => {
      categoryPulseTimer.current = null;
      setCategoryPulse(false);
    }, 480);
  }

  function handleCategoryChange(next: string) {
    setCategory(next);
    setFieldError("");
    pulseCategoryField();
  }

  function validate(): string {
    const q = question.trim();
    if (!q) return "请输入你想整理的问题";
    const len = [...q].length;
    if (len < MIN_QUESTION_LENGTH || len > MAX_QUESTION_LENGTH) {
      return `问题长度需在 ${MIN_QUESTION_LENGTH} 到 ${MAX_QUESTION_LENGTH} 字之间`;
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
      const record = await createQimenAnalysis({
        session_key: getSessionKey(),
        question: question.trim(),
        category,
        confirm_disclaimer: true,
      });
      router.push(`/analysis/${record.id}`);
    } catch (err) {
      if (isSensitiveBlockError(err)) {
        setApiError(
          "这个问题不适合用奇门简化解读，请换成自我反思、局势整理或行动节奏类问题。"
        );
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
        <LoadingSpinner label="准备奇门问事…" />
      </main>
    );
  }

  return (
    <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col px-4 py-8 sm:py-10">
      <div className="mb-6">
        <Link
          href="/"
          className="-ml-3 inline-flex min-h-11 items-center rounded-lg px-3 text-sm text-amber-800 hover:underline"
        >
          ← 返回首页
        </Link>
        <div className="qimen-hero mt-4">
          <div className="qimen-hero__grid">
            <QimenScanGrid />
          </div>
          <div className="qimen-hero__copy">
            <p className="text-sm tracking-[0.15em] text-amber-800">奇门问事</p>
            <h1 className="mt-2 text-2xl font-bold text-stone-900">
              基于传统奇门文化的局势梳理与行动节奏参考
            </h1>
            <p className="mt-2 text-sm leading-relaxed text-stone-600">
              {METHOD_NOTE}
            </p>
          </div>
        </div>
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
          <span className="text-sm font-medium text-stone-800">问事分类</span>
          <div className={categoryPulse ? "category-pulse" : ""}>
            <Select
              value={category}
              onChange={(v) => handleCategoryChange(String(v))}
              options={QIMEN_CATEGORIES.map((item) => ({
                value: item.value,
                label: item.label,
              }))}
              disabled={submitting}
            />
          </div>
        </label>

        <label className="mt-5 block">
          <span className="text-sm font-medium text-stone-800">你的问题</span>
          <textarea
            value={question}
            onChange={(e) => {
              setQuestion(e.target.value);
              setFieldError("");
            }}
            disabled={submitting}
            rows={5}
            maxLength={MAX_QUESTION_LENGTH}
            placeholder="例如：我最近适合如何安排这项计划的推进节奏？"
            className="mt-2 w-full resize-none rounded-xl border border-stone-300 px-4 py-3 text-base leading-relaxed text-stone-900 outline-none focus:border-amber-600 disabled:opacity-60 sm:text-sm"
          />
          <p className="mt-1 text-right text-xs text-stone-400">
            {[...question.trim()].length} / {MAX_QUESTION_LENGTH} 字
          </p>
        </label>

        <p className="mt-3 rounded-xl bg-stone-50 px-3 py-2 text-xs leading-relaxed text-stone-500">
          请勿输入医疗、法律、投资、赌博、伤害自己或他人等问题。本工具不会处理此类内容。
        </p>

        <label className="mt-4 flex items-start gap-3">
          <input
            type="checkbox"
            checked={confirmed}
            onChange={(e) => setConfirmed(e.target.checked)}
            disabled={submitting}
            className="mt-0.5 h-5 w-5 shrink-0 rounded border-stone-300"
          />
          <span className="text-sm leading-relaxed text-stone-600">
            {METHOD_NOTE}
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
          {submitting ? "整理局势中…" : "生成奇门简析"}
        </button>
      </form>

      <RecentAnalysisList
        items={recentList}
        loading={listLoading}
        error={listError}
        moduleLabel={MODULE_QIMEN_LABEL}
        emptyText="暂无奇门问事记录，提交后将显示在这里。"
        getSubtitle={listRecordSubtitle}
        onRetry={loadRecentList}
      />

      <p className="mt-6 text-center text-xs leading-relaxed text-stone-500">
        {DISCLAIMER}
      </p>
    </main>
  );
}
