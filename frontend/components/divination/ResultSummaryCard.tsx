import type { Divination, FullReport } from "@/lib/types";
import { formatDateTime } from "@/lib/utils";

interface ResultSummaryCardProps {
  divination: Divination;
  hint?: string;
}

export default function ResultSummaryCard({
  divination,
  hint,
}: ResultSummaryCardProps) {
  const defaultHint =
    divination.moving_lines?.length > 0
      ? `动爻在第 ${divination.moving_lines.join("、")} 爻，可关注变化处的提醒。`
      : "本次无动爻，可重点理解本卦整体含义。";

  return (
    <section className="overflow-hidden rounded-2xl border border-amber-200/80 bg-gradient-to-br from-amber-50 via-white to-stone-50 p-5 shadow-sm sm:p-6">
      <div className="flex items-center justify-between gap-2">
        <span className="rounded-full bg-amber-100 px-3 py-1 text-xs font-medium text-amber-900">
          {divination.category?.name}
        </span>
        <span className="text-xs text-stone-400">
          {formatDateTime(divination.created_at)}
        </span>
      </div>
      <h1 className="mt-4 text-lg font-bold leading-relaxed text-stone-900 sm:text-xl">
        {divination.question}
      </h1>
      <p className="mt-4 border-t border-amber-100 pt-4 text-sm leading-relaxed text-stone-600">
        {hint ?? defaultHint}
      </p>
    </section>
  );
}

export function buildPosterSummary(
  freeContent: string,
  fullReport: FullReport | string | null,
  primarySummary?: string
): string {
  if (fullReport && typeof fullReport === "object" && fullReport.summary) {
    return fullReport.summary;
  }
  const parsed =
    typeof fullReport === "string"
      ? (() => {
          try {
            return JSON.parse(fullReport) as FullReport;
          } catch {
            return null;
          }
        })()
      : null;
  if (parsed?.summary) return parsed.summary;

  const fromFree = freeContent.split(/[。！？\n]/)[0]?.trim();
  if (fromFree && fromFree.length > 4) return fromFree;
  if (primarySummary) return primarySummary;
  return "传统文化参考，助力自我反思与行动整理。";
}
