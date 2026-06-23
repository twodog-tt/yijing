"use client";

import CastingOverlay from "@/components/casting/CastingOverlay";
import ErrorAlert from "@/components/ui/ErrorAlert";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import Select from "@/components/ui/Select";
import {
  ApiError,
  createDivination,
  createSession,
  getCategories,
  isSensitiveBlockError,
} from "@/lib/api";
import { getSessionKey } from "@/lib/session";
import type { Category, Divination } from "@/lib/types";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useCallback, useEffect, useState } from "react";

const DISCLAIMER =
  "本内容仅供娱乐和传统文化参考，不构成现实决策建议。";

export default function AskPage() {
  const router = useRouter();
  const [categories, setCategories] = useState<Category[]>([]);
  const [categoryId, setCategoryId] = useState<number | "">("");
  const [question, setQuestion] = useState("");
  const [confirmed, setConfirmed] = useState(false);
  const [loading, setLoading] = useState(true);
  const [castingOpen, setCastingOpen] = useState(false);
  const [castingWaiting, setCastingWaiting] = useState(false);
  const [castingDivination, setCastingDivination] =
    useState<Divination | null>(null);
  const [apiError, setApiError] = useState("");
  const [fieldError, setFieldError] = useState("");

  useEffect(() => {
    async function init() {
      try {
        const sessionKey = getSessionKey();
        await createSession(sessionKey);
        const items = await getCategories();
        setCategories(items);
      } catch (e) {
        setApiError(e instanceof Error ? e.message : "加载失败");
      } finally {
        setLoading(false);
      }
    }
    init();
  }, []);

  const handleCastingComplete = useCallback(() => {
    if (castingDivination) {
      router.push(`/divination/${castingDivination.id}`);
    }
  }, [castingDivination, router]);

  function validate(): string {
    if (categoryId === "") return "请选择事项类型";
    const q = question.trim();
    if (!q) return "请输入你的问题";
    const len = [...q].length;
    if (len < 5 || len > 200) return "问题长度需在 5 到 200 字之间";
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
    setCastingOpen(true);
    setCastingWaiting(true);
    setCastingDivination(null);

    try {
      const result = await createDivination({
        session_key: getSessionKey(),
        category_id: categoryId as number,
        question: question.trim(),
        confirm_disclaimer: true,
      });
      setCastingDivination(result);
      setCastingWaiting(false);
    } catch (err) {
      setCastingOpen(false);
      setCastingWaiting(false);
      setCastingDivination(null);
      if (isSensitiveBlockError(err)) {
        setApiError((err as ApiError).message);
      } else {
        setApiError(err instanceof Error ? err.message : "起卦失败，请重试");
      }
    }
  }

  if (loading) {
    return (
      <main className="mx-auto w-full max-w-2xl flex-1 px-4 py-10">
        <LoadingSpinner label="准备问事…" />
      </main>
    );
  }

  const casting = castingOpen;

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
          <h1 className="mt-4 text-2xl font-bold text-stone-900">问事起卦</h1>
          <p className="mt-2 text-sm text-stone-600">
            选择事项类型，输入你想整理的问题。
          </p>
        </div>

        {apiError && (
          <div className="mb-4">
            <ErrorAlert message={apiError} />
          </div>
        )}

        <form
          onSubmit={handleSubmit}
          className="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm"
        >
          <label className="block">
            <span className="text-sm font-medium text-stone-800">事项类型</span>
            <Select
              value={categoryId}
              onChange={(v) =>
                setCategoryId(v === "" ? "" : Number(v))
              }
              options={categories.map((c) => ({
                value: c.id,
                label: c.name,
              }))}
              placeholder="请选择"
              disabled={casting}
            />
          </label>

          <label className="mt-5 block">
            <span className="text-sm font-medium text-stone-800">你的问题</span>
            <textarea
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
              disabled={casting}
              rows={5}
              placeholder="例如：我现在适不适合继续推进这个 AI 易经小程序？"
              className="mt-2 w-full resize-none rounded-xl border border-stone-300 px-4 py-3 text-sm leading-relaxed text-stone-900 outline-none focus:border-amber-600 disabled:opacity-60"
            />
            <p className="mt-1 text-right text-xs text-stone-400">
              {[...question.trim()].length} / 200 字
            </p>
          </label>

          <label className="mt-4 flex items-start gap-3">
            <input
              type="checkbox"
              checked={confirmed}
              onChange={(e) => setConfirmed(e.target.checked)}
              disabled={casting}
              className="mt-1 h-4 w-4 rounded border-stone-300"
            />
            <span className="text-sm leading-relaxed text-stone-600">
              我已阅读并理解：本工具仅供娱乐和传统文化参考，不构成医疗、法律、投资或重大决策建议。
            </span>
          </label>

          {fieldError && (
            <p className="mt-3 text-sm text-red-600" role="alert">
              {fieldError}
            </p>
          )}

          <button
            type="submit"
            disabled={casting}
            className="mt-6 w-full rounded-xl bg-stone-900 py-3.5 text-sm font-semibold text-white transition hover:bg-stone-800 disabled:opacity-60"
          >
            {casting ? "起卦中…" : "开始起卦"}
          </button>
        </form>

        <p className="mt-6 text-center text-xs leading-relaxed text-stone-500">
          {DISCLAIMER}
        </p>
      </main>
    </>
  );
}
