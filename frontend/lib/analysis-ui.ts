import type { AnalysisListItem } from "./types";
import { QUESTION_SUMMARY } from "./qimen";

export function safeAnalysisListSubtitle(
  item: AnalysisListItem,
  getSubtitle: (item: AnalysisListItem) => string
): string {
  const text = getSubtitle(item).trim();
  const rawQuestion = item.question?.trim();
  if (rawQuestion && text === rawQuestion) {
    return QUESTION_SUMMARY;
  }
  return text;
}
