import type { AnalysisRecord } from "./types";
import { ModuleTypeQimen } from "./types";
import { sanitizeInternalTermList, sanitizeInternalTerms } from "./display-text";

export const MODULE_QIMEN_LABEL = "奇门问事";
export const QUESTION_SUMMARY = "用户问题已用于本次局势梳理";
export const METHOD_NOTE =
  "本功能采用简化奇门文化规则，仅供传统文化学习与自我反思参考，不等同于专业奇门排盘，也不构成现实决策依据。";

export const QIMEN_CATEGORIES = [
  { value: "career", label: "事业/计划" },
  { value: "relationship", label: "人际/关系" },
  { value: "study", label: "学习/成长" },
  { value: "decision", label: "决策/选择" },
  { value: "general", label: "综合问题" },
] as const;

export const MIN_QUESTION_LENGTH = 4;
export const MAX_QUESTION_LENGTH = 120;

const CATEGORY_LABELS: Record<string, string> = QIMEN_CATEGORIES.reduce(
  (acc, item) => {
    acc[item.value] = item.label;
    return acc;
  },
  {} as Record<string, string>
);

const TIME_BUCKET_LABELS: Record<string, string> = {
  morning: "上午",
  day: "白天",
  evening: "傍晚",
  night: "夜间",
};

export interface QimenLensView {
  focusTheme: string;
  supportTheme: string;
  cautionTheme: string;
  pacingTheme: string;
}

export interface QimenAnalysisView {
  algorithmVersion: string;
  methodNote: string;
  category: string;
  categoryLabel: string;
  questionSummary: string;
  timeBucket: string;
  timeBucketLabel: string;
  qimenLens: QimenLensView;
  situationOverview: string;
  riskObservations: string[];
  actionPacing: string;
  reflectionQuestions: string[];
  actionSuggestions: string[];
  limits: string[];
  freeContent: string;
}

function parseJSONField(raw: unknown): Record<string, unknown> {
  if (!raw) return {};
  if (typeof raw === "object") return raw as Record<string, unknown>;
  if (typeof raw !== "string") return {};
  try {
    return JSON.parse(raw) as Record<string, unknown>;
  } catch {
    return {};
  }
}

function displayTextField(value: unknown, fallback = ""): string {
  const text = sanitizeInternalTerms(value);
  return text || fallback;
}

export function buildQimenView(record: AnalysisRecord): QimenAnalysisView {
  const result = parseJSONField(record.result_payload);
  const category = (result.category as string) || "general";
  const risks = Array.isArray(result.risk_observations)
    ? (result.risk_observations as string[])
    : [];
  const reflections = Array.isArray(result.reflection_questions)
    ? (result.reflection_questions as string[])
    : [];
  const suggestions = Array.isArray(result.action_suggestions)
    ? (result.action_suggestions as string[])
    : [];
  const calcMeta = (result.calculation_meta as Record<string, unknown>) || {};
  const limits = calcMeta.limits;
  const timeContext = (result.time_context as Record<string, string>) || {};
  const timeBucket = timeContext.time_bucket || "";
  const lensRaw = (result.qimen_lens as Record<string, string>) || {};

  return {
    algorithmVersion:
      (result.algorithm_version as string) ||
      record.algorithm_version ||
      "qimen-simple-v1",
    methodNote: displayTextField(result.method_note) || METHOD_NOTE,
    category,
    categoryLabel: CATEGORY_LABELS[category] || CATEGORY_LABELS.general,
    questionSummary: QUESTION_SUMMARY,
    timeBucket,
    timeBucketLabel: TIME_BUCKET_LABELS[timeBucket] || "",
    qimenLens: {
      focusTheme: displayTextField(lensRaw.focus_theme),
      supportTheme: displayTextField(lensRaw.support_theme),
      cautionTheme: displayTextField(lensRaw.caution_theme),
      pacingTheme: displayTextField(lensRaw.pacing_theme),
    },
    situationOverview: displayTextField(result.situation_overview),
    riskObservations: sanitizeInternalTermList(risks),
    actionPacing: displayTextField(result.action_pacing),
    reflectionQuestions: sanitizeInternalTermList(reflections),
    actionSuggestions: sanitizeInternalTermList(suggestions),
    limits: sanitizeInternalTermList(limits),
    freeContent: displayTextField(record.free_content),
  };
}

export function isQimenRecord(record: AnalysisRecord): boolean {
  return record.module_type === ModuleTypeQimen;
}

export function listRecordSubtitle(): string {
  return QUESTION_SUMMARY;
}
