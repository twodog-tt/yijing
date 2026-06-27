import type { AnalysisListItem } from "./types";
import { QUESTION_SUMMARY } from "./qimen";
import { sanitizeInternalTerms } from "./display-text";

export function safeAnalysisListSubtitle(
  item: AnalysisListItem,
  getSubtitle: (item: AnalysisListItem) => string
): string {
  const text = sanitizeInternalTerms(getSubtitle(item));
  const rawQuestion = item.question?.trim();
  if (rawQuestion && text === rawQuestion) {
    return QUESTION_SUMMARY;
  }
  return text || "记录摘要";
}
