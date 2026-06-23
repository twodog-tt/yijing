"use client";

import ErrorAlert from "@/components/ui/ErrorAlert";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import { getDivinationHistory } from "@/lib/api";
import { getSessionKey } from "@/lib/session";
import type { DivinationListItem } from "@/lib/types";
import { formatDateShort, formatDateOnly, isDailyFortuneCategory, truncateText } from "@/lib/utils";
import Link from "next/link";
import { useCallback, useEffect, useState } from "react";

const PAGE_SIZE = 20;

export default function HistoryPage() {
  const [items, setItems] = useState<DivinationListItem[]>([]);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState("");

  const applyPagination = useCallback(
    (newItems: DivinationListItem[], currentPage: number, total: number) => {
      setHasMore(currentPage * PAGE_SIZE < total && newItems.length > 0);
    },
    []
  );

  const loadHistory = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await getDivinationHistory(getSessionKey(), 1, PAGE_SIZE);
      const list = res.items ?? [];
      setItems(list);
      setPage(1);
      applyPagination(list, 1, res.total ?? list.length);
    } catch (e) {
      setError(e instanceof Error ? e.message : "加载历史记录失败");
    } finally {
      setLoading(false);
    }
  }, [applyPagination]);

  async function loadMore() {
    const nextPage = page + 1;
    setLoadingMore(true);
    setError("");
    try {
      const res = await getDivinationHistory(
        getSessionKey(),
        nextPage,
        PAGE_SIZE
      );
      const list = res.items ?? [];
      setItems((prev) => [...prev, ...list]);
      setPage(nextPage);
      applyPagination(list, nextPage, res.total ?? items.length + list.length);
      if (list.length < PAGE_SIZE) {
        setHasMore(false);
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "加载更多失败");
    } finally {
      setLoadingMore(false);
    }
  }

  useEffect(() => {
    loadHistory();
  }, [loadHistory]);

  return (
    <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col px-4 py-8 sm:py-10">
      <div className="mb-6">
        <Link href="/" className="-ml-3 inline-flex min-h-11 items-center rounded-lg px-3 text-sm text-amber-800 hover:underline">
          ← 返回首页
        </Link>
        <h1 className="mt-4 text-2xl font-bold text-stone-900">历史记录</h1>
      </div>

      {loading && <LoadingSpinner label="加载历史…" />}

      {error && !loading && (
        <ErrorAlert message={error} onRetry={loadHistory} />
      )}

      {!loading && !error && items.length === 0 && (
        <section className="rounded-2xl border border-dashed border-stone-300 bg-white p-6 text-center sm:p-10">
          <p className="text-stone-600">你还没有起过卦，先去问一件事吧。</p>
          <Link
            href="/ask"
            className="mt-6 inline-block rounded-xl bg-stone-900 px-6 py-3 text-sm font-semibold text-white"
          >
            立即起卦
          </Link>
        </section>
      )}

      {!loading && items.length > 0 && (
        <>
          <ul className="space-y-3">
            {items.map((item) => (
              <li key={item.id}>
                <Link
                  href={`/divination/${item.id}`}
                  className="block rounded-xl border border-stone-200 bg-white p-4 shadow-sm transition hover:border-amber-300 hover:shadow"
                >
                  <div className="flex items-start justify-between gap-3">
                    <p className="flex-1 text-sm font-medium leading-relaxed text-stone-900">
                      {isDailyFortuneCategory(item.category?.name)
                        ? `今日运势｜${formatDateOnly(item.created_at)}`
                        : truncateText(item.question, 40)}
                    </p>
                    <span
                      className={`shrink-0 rounded-full px-2 py-0.5 text-xs ${
                        item.unlock_status === 1
                          ? "bg-green-100 text-green-800"
                          : "bg-stone-100 text-stone-600"
                      }`}
                    >
                      {item.unlock_status === 1 ? "已解锁" : "未解锁"}
                    </span>
                  </div>
                  <p className="mt-2 text-xs text-stone-500">
                    {item.category?.name} · 本卦 {item.primary_hexagram?.name}{" "}
                    → 变卦 {item.changed_hexagram?.name}
                  </p>
                  <p className="mt-1 text-xs text-stone-400">
                    {formatDateShort(item.created_at)}
                  </p>
                </Link>
              </li>
            ))}
          </ul>

          <div className="mt-6 text-center">
            {hasMore ? (
              <button
                type="button"
                onClick={loadMore}
                disabled={loadingMore}
                className="min-h-11 rounded-xl border border-stone-300 bg-white px-8 py-3 text-sm font-medium text-stone-700 disabled:opacity-60"
              >
                {loadingMore ? "加载中…" : "加载更多"}
              </button>
            ) : (
              items.length >= PAGE_SIZE && (
                <p className="text-sm text-stone-400">没有更多了</p>
              )
            )}
          </div>
        </>
      )}
    </main>
  );
}
