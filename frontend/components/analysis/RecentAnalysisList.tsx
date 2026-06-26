"use client";

import ErrorAlert from "@/components/ui/ErrorAlert";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import type { AnalysisListItem } from "@/lib/types";
import { safeAnalysisListSubtitle } from "@/lib/analysis-ui";
import { formatDateTime } from "@/lib/utils";
import Link from "next/link";

interface RecentAnalysisListProps {
  items: AnalysisListItem[];
  loading: boolean;
  error: string;
  moduleLabel: string;
  emptyText: string;
  getSubtitle: (item: AnalysisListItem) => string;
  onRetry: () => void;
}

export default function RecentAnalysisList({
  items,
  loading,
  error,
  moduleLabel,
  emptyText,
  getSubtitle,
  onRetry,
}: RecentAnalysisListProps) {
  return (
    <section className="mt-8">
      <h2 className="text-sm font-semibold text-stone-900">最近记录</h2>

      {loading && (
        <div className="mt-4">
          <LoadingSpinner label="正在加载最近记录…" />
        </div>
      )}

      {error && !loading && (
        <div className="mt-4">
          <ErrorAlert message={error} onRetry={onRetry} />
        </div>
      )}

      {!loading && !error && items.length === 0 && (
        <p className="mt-4 text-sm text-stone-500">{emptyText}</p>
      )}

      {!loading && !error && items.length > 0 && (
        <ul className="mt-4 divide-y divide-stone-100 rounded-2xl border border-stone-200 bg-white shadow-sm">
          {items.map((item) => (
            <li key={item.id}>
              <Link
                href={`/analysis/${item.id}`}
                className="flex items-center justify-between gap-3 px-4 py-4 transition hover:bg-stone-50"
              >
                <div className="min-w-0">
                  <p className="text-sm font-medium text-stone-900">
                    {moduleLabel}
                  </p>
                  <p className="mt-1 truncate text-xs text-stone-500">
                    {safeAnalysisListSubtitle(item, getSubtitle)}
                  </p>
                </div>
                <time className="shrink-0 text-xs text-stone-400">
                  {formatDateTime(item.created_at)}
                </time>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
