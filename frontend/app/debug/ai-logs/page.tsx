"use client";

import ErrorAlert from "@/components/ui/ErrorAlert";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import { getAILogs, getAIStats, isDebugDisabledError } from "@/lib/api";
import { sanitizeInternalTerms } from "@/lib/display-text";
import type { AIGenerationLog, AIStats } from "@/lib/types";
import { formatDateTime } from "@/lib/utils";
import Link from "next/link";
import { useCallback, useEffect, useState } from "react";

const PAGE_SIZE = 20;

function statusLabel(status: number): { text: string; className: string } {
  switch (status) {
    case 1:
      return {
        text: "成功",
        className: "bg-emerald-100 text-emerald-800 border-emerald-200",
      };
    case 2:
      return {
        text: "失败",
        className: "bg-red-100 text-red-800 border-red-200",
      };
    case 3:
      return {
        text: "降级成功",
        className: "bg-amber-100 text-amber-800 border-amber-200",
      };
    default:
      return {
        text: `未知(${status})`,
        className: "bg-stone-100 text-stone-600 border-stone-200",
      };
  }
}

function StatsCards({ stats }: { stats: AIStats }) {
  const cards = [
    { label: "总调用", value: stats.total_count },
    { label: "成功", value: stats.success_count },
    { label: "失败", value: stats.fail_count },
    { label: "降级", value: stats.fallback_count },
    {
      label: "平均耗时",
      value: `${Math.round(stats.avg_duration_ms)} 毫秒`,
    },
  ];

  return (
    <div className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-3">
      {cards.map((card) => (
        <div
          key={card.label}
          className="rounded-xl border border-stone-200 bg-white p-4 shadow-sm"
        >
          <p className="text-xs text-stone-500">{card.label}</p>
          <p className="mt-1 text-lg font-bold text-stone-900">{card.value}</p>
        </div>
      ))}
      {stats.latest_created_at ? (
        <div className="col-span-2 rounded-xl border border-stone-200 bg-white p-4 shadow-sm sm:col-span-3">
          <p className="text-xs text-stone-500">最近调用</p>
          <p className="mt-1 text-sm font-medium text-stone-800">
            {formatDateTime(stats.latest_created_at)}
          </p>
        </div>
      ) : null}
    </div>
  );
}

function LogRow({ log }: { log: AIGenerationLog }) {
  const badge = statusLabel(log.status);

  return (
    <div className="rounded-xl border border-stone-200 bg-white p-4 shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex flex-wrap items-center gap-2">
          <span
            className={`rounded-full border px-2.5 py-0.5 text-xs font-medium ${badge.className}`}
          >
            {badge.text}
          </span>
          <span className="text-xs text-stone-500">#{log.id}</span>
        </div>
        <span className="text-xs text-stone-400">
          {formatDateTime(log.created_at)}
        </span>
      </div>

      <dl className="mt-3 grid gap-2 text-sm sm:grid-cols-2">
        <div>
          <dt className="text-xs text-stone-500">问事记录编号</dt>
          <dd className="font-medium text-stone-800">
            <Link
              href={`/divination/${log.divination_id}`}
              className="text-amber-800 hover:underline"
            >
              {log.divination_id}
            </Link>
          </dd>
        </div>
        <div>
          <dt className="text-xs text-stone-500">生成方式 / 模型</dt>
          <dd className="text-stone-800">
            {sanitizeInternalTerms(log.ai_provider)} ·{" "}
            {sanitizeInternalTerms(log.model_name)}
          </dd>
        </div>
        <div>
          <dt className="text-xs text-stone-500">耗时</dt>
          <dd className="text-stone-800">{log.duration_ms} 毫秒</dd>
        </div>
        <div>
          <dt className="text-xs text-stone-500">是否降级</dt>
          <dd className="text-stone-800">
            {log.fallback_used === 1 ? "是" : "否"}
          </dd>
        </div>
      </dl>

      {log.error_message ? (
        <p className="mt-3 rounded-lg bg-stone-50 px-3 py-2 text-xs leading-relaxed text-stone-600">
          {sanitizeInternalTerms(log.error_message)}
        </p>
      ) : null}
    </div>
  );
}

export default function AILogsDebugPage() {
  const [items, setItems] = useState<AIGenerationLog[]>([]);
  const [stats, setStats] = useState<AIStats | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState("");
  const [debugDisabled, setDebugDisabled] = useState(false);

  const loadLogs = useCallback(async () => {
    setLoading(true);
    setError("");
    setDebugDisabled(false);
    try {
      const [statsRes, logsRes] = await Promise.all([
        getAIStats(),
        getAILogs(1, PAGE_SIZE),
      ]);
      const list = logsRes.items ?? [];
      setStats(statsRes);
      setItems(list);
      setPage(1);
      setHasMore((logsRes.total ?? 0) > list.length);
    } catch (e) {
      if (isDebugDisabledError(e)) {
        setDebugDisabled(true);
        setError("调试接口未启用，仅本地开发环境可用。");
      } else {
        setError(e instanceof Error ? e.message : "加载智能生成日志失败");
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- fetch debug logs after mount
    loadLogs();
  }, [loadLogs]);

  async function loadMore() {
    const nextPage = page + 1;
    setLoadingMore(true);
    setError("");
    try {
      const res = await getAILogs(nextPage, PAGE_SIZE);
      const list = res.items ?? [];
      setItems((prev) => [...prev, ...list]);
      setPage(nextPage);
      setHasMore(nextPage * PAGE_SIZE < (res.total ?? 0));
    } catch (e) {
      setError(e instanceof Error ? e.message : "加载更多失败");
    } finally {
      setLoadingMore(false);
    }
  }

  return (
    <main className="mx-auto min-h-screen max-w-2xl px-4 py-8">
      <div className="mb-6 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
        仅本地调试使用 — 生产环境请设置{" "}
        <code className="text-xs">ENABLE_DEBUG_ROUTES=false</code>。
      </div>

      <h1 className="text-2xl font-bold text-stone-900">智能生成调用日志</h1>
      <p className="mt-1 text-sm text-stone-500">
        查看完整解读生成时的生成方式、耗时与降级情况。
      </p>

      {error ? (
        <div className="mt-4">
          <ErrorAlert message={error} />
        </div>
      ) : null}

      {loading ? (
        <div className="mt-12 flex justify-center">
          <LoadingSpinner />
        </div>
      ) : debugDisabled ? null : (
        <>
          {stats ? <StatsCards stats={stats} /> : null}

          {items.length === 0 ? (
            <p className="mt-8 text-center text-sm text-stone-500">
              暂无记录。请先完成一次完整解读。
            </p>
          ) : (
            <div className="mt-6 space-y-3">
              {items.map((log) => (
                <LogRow key={log.id} log={log} />
              ))}
            </div>
          )}

          {hasMore ? (
            <div className="mt-6 text-center">
              <button
                type="button"
                onClick={loadMore}
                disabled={loadingMore}
                className="rounded-full border border-stone-300 px-6 py-2 text-sm text-stone-700 hover:bg-stone-50 disabled:opacity-50"
              >
                {loadingMore ? "加载中…" : "加载更多"}
              </button>
            </div>
          ) : null}
        </>
      )}
    </main>
  );
}
