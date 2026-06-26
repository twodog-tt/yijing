const {
  pickPosterActionPoints,
  limitPosterText,
  getQimenCategoryHighlight,
} = require("./long-poster-canvas");

const MODULE_QIMEN_TYPE = 2;
const MODULE_QIMEN_LABEL = "奇门问事";
const QUESTION_SUMMARY = "用户问题已用于本次局势梳理";
const ALGORITHM_QIMEN_SIMPLE_V1 = "qimen-simple-v1";

const QIMEN_CATEGORIES = Object.freeze([
  { value: "career", label: "事业/计划" },
  { value: "relationship", label: "人际/关系" },
  { value: "study", label: "学习/成长" },
  { value: "decision", label: "决策/选择" },
  { value: "general", label: "综合问题" },
]);

const CATEGORY_LABELS = Object.freeze(
  QIMEN_CATEGORIES.reduce((acc, item) => {
    acc[item.value] = item.label;
    return acc;
  }, {})
);

const TIME_BUCKET_LABELS = Object.freeze({
  morning: "上午",
  day: "白天",
  evening: "傍晚",
  night: "夜间",
});

const METHOD_NOTE =
  "本功能采用 qimen-simple-v1 简化规则，仅供传统文化学习与自我反思参考，不等同于专业奇门排盘，也不构成现实决策依据。";

const FREE_CONTENT_POSTER_NOTE =
  "以上局势梳理、风险观察、行动节奏、自我反思与行动建议构成本次免费解读要点。";

const MIN_QUESTION_LENGTH = 4;
const MAX_QUESTION_LENGTH = 120;

function parseJSONField(raw) {
  if (!raw) return {};
  if (typeof raw === "object") return raw;
  if (typeof raw !== "string") return {};
  try {
    return JSON.parse(raw);
  } catch (_error) {
    return {};
  }
}

function getCategoryLabel(category) {
  return CATEGORY_LABELS[category] || CATEGORY_LABELS.general;
}

function getTimeBucketLabel(bucket) {
  return TIME_BUCKET_LABELS[bucket] || "";
}

function buildQimenView(record) {
  const result = parseJSONField(record?.result_payload);
  const category = result.category || "general";
  const risks = Array.isArray(result.risk_observations)
    ? result.risk_observations
    : [];
  const reflections = Array.isArray(result.reflection_questions)
    ? result.reflection_questions
    : [];
  const suggestions = Array.isArray(result.action_suggestions)
    ? result.action_suggestions
    : [];
  const limits = result.calculation_meta?.limits;
  const timeBucket = result.time_context?.time_bucket || "";
  const lens = result.qimen_lens || {};
  const questionProfile = result.question_profile || {};

  return {
    algorithmVersion:
      result.algorithm_version || record?.algorithm_version || ALGORITHM_QIMEN_SIMPLE_V1,
    methodNote: result.method_note || METHOD_NOTE,
    category,
    categoryLabel: getCategoryLabel(category),
    questionSummary: QUESTION_SUMMARY,
    timeBucket,
    timeBucketLabel: getTimeBucketLabel(timeBucket),
    qimenLens: {
      focusTheme: lens.focus_theme || "",
      supportTheme: lens.support_theme || "",
      cautionTheme: lens.caution_theme || "",
      pacingTheme: lens.pacing_theme || "",
    },
    questionProfile: {
      intentType: questionProfile.intent_type || "",
      timeHorizon: questionProfile.time_horizon || "",
      decisionPressure: questionProfile.decision_pressure || "",
      relationScope: questionProfile.relation_scope || "",
      riskTone: questionProfile.risk_tone || "",
    },
    safeQuestionSummary: result.safe_question_summary || QUESTION_SUMMARY,
    situationOverview: result.situation_overview || "",
    riskObservations: risks,
    actionPacing: result.action_pacing || "",
    reflectionQuestions: reflections,
    actionSuggestions: suggestions,
    limits: Array.isArray(limits) ? limits : [],
    freeContent: String(record?.free_content || "").trim(),
  };
}

function isQimenRecord(record) {
  return Number(record?.module_type) === MODULE_QIMEN_TYPE;
}

function listRecordSubtitle(item) {
  const summary = String(item?.question || "").trim();
  if (!summary || summary === QUESTION_SUMMARY) {
    return QUESTION_SUMMARY;
  }
  return QUESTION_SUMMARY;
}

function buildQimenLongPosterData(recordId, view, fullContent) {
  if (!recordId || !view) return null;

  const lens = view.qimenLens || {};
  const profile = view.questionProfile || {};
  const combinedSuggestions = [
    ...(Array.isArray(view.actionSuggestions) ? view.actionSuggestions : []),
    ...(Array.isArray(view.reflectionQuestions) ? view.reflectionQuestions : []),
  ];

  const actionPoints = pickPosterActionPoints({
    suggestions: combinedSuggestions,
    fullContent,
    sectionHints: ["反思", "行动", "节奏"],
    fallback: [
      "先把问题拆小，安排一件今天能完成的小事。",
      "记录一次互动或学习感受，再决定是否调整节奏。",
    ],
    maxItems: 3,
    maxLength: 72,
  });

  return {
    id: String(recordId),
    methodNote: view.methodNote || METHOD_NOTE,
    category: view.category || "general",
    categoryLabel: view.categoryLabel || getCategoryLabel(view.category),
    timeBucketLabel: view.timeBucketLabel || "",
    questionSummary: QUESTION_SUMMARY,
    qimenLens: {
      focusTheme: limitPosterText(lens.focusTheme, 72),
      supportTheme: limitPosterText(lens.supportTheme, 72),
      cautionTheme: limitPosterText(lens.cautionTheme, 72),
      pacingTheme: limitPosterText(lens.pacingTheme, 72),
    },
    questionProfile: {
      intentType: limitPosterText(profile.intentType, 32),
      timeHorizon: limitPosterText(profile.timeHorizon, 24),
      decisionPressure: limitPosterText(profile.decisionPressure, 16),
      relationScope: limitPosterText(profile.relationScope, 24),
      riskTone: limitPosterText(profile.riskTone, 24),
    },
    categoryHighlight: getQimenCategoryHighlight(view.category || "general"),
    actionPoints,
  };
}

module.exports = {
  ALGORITHM_QIMEN_SIMPLE_V1,
  MAX_QUESTION_LENGTH,
  FREE_CONTENT_POSTER_NOTE,
  METHOD_NOTE,
  MIN_QUESTION_LENGTH,
  MODULE_QIMEN_LABEL,
  MODULE_QIMEN_TYPE,
  QIMEN_CATEGORIES,
  QUESTION_SUMMARY,
  buildQimenView,
  buildQimenLongPosterData,
  getCategoryLabel,
  isQimenRecord,
  listRecordSubtitle,
};
