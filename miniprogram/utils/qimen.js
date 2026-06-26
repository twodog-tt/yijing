const {
  pickPosterActionPoints,
  limitPosterText,
  getQimenCategoryHighlight,
} = require("./long-poster-canvas");
const { formatDateTime } = require("./date");

const MODULE_QIMEN_TYPE = 2;
const MODULE_QIMEN_LABEL = "奇门问事";
const QUESTION_SUMMARY = "用户问题已用于本次局势梳理";
const ALGORITHM_QIMEN_SIMPLE_V1 = "qimen-simple-v1";
const ALGORITHM_QIMEN_V2_PROFESSIONAL = "qimen-v2-professional";

const PROFESSIONAL_PALACE_GRID_ORDER = Object.freeze([4, 9, 2, 3, 5, 7, 8, 1, 6]);

const LAYOUT_ROLE_LABELS = Object.freeze({
  center: "中宫",
  chief: "值符宫",
  palace: "",
});

const DUN_YUAN_LABELS = Object.freeze({
  upper: "上元",
  middle: "中元",
  lower: "下元",
});

const PROFESSIONAL_PREVIEW_NOTE =
  "奇门 v2 professional 第一版 · 传统文化学习参考 · 结构化观察 · 不构成现实决策依据";

const PROFESSIONAL_POSTER_NOTE =
  "本记录含九宫结构化观察，详情页可查看排盘口径与宫位摘要。";

const PROFESSIONAL_BOUNDARY_NOTES = Object.freeze([
  "当前为第一版排盘口径，不等同于最终权威专业排盘",
  "置闰法尚未实现",
  "寄宫流派仍待校准",
  "不构成现实决策依据",
]);

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

function normalizeTextField(value, fallback = "") {
  const text = String(value == null ? "" : value).trim();
  return text || fallback;
}

function normalizePalaceView(raw) {
  if (!raw || typeof raw !== "object") return null;
  const index = Number(raw.index);
  if (!Number.isInteger(index) || index < 1 || index > 9) return null;

  const layoutRole = normalizeTextField(raw.layout_role);
  return {
    index,
    name: normalizeTextField(raw.name, `第${index}宫`),
    earthPlateStem: normalizeTextField(raw.earth_plate_stem, "—"),
    heavenPlateStem: normalizeTextField(raw.heaven_plate_stem, "—"),
    star: normalizeTextField(raw.star, "—"),
    door: normalizeTextField(raw.door, "—"),
    deity: normalizeTextField(raw.deity, "—"),
    layoutRole,
    layoutRoleLabel: LAYOUT_ROLE_LABELS[layoutRole] || "",
    summary: normalizeTextField(raw.summary),
  };
}

function buildProfessionalQimenView(result) {
  if (normalizeTextField(result?.algorithm_version) !== ALGORITHM_QIMEN_V2_PROFESSIONAL) {
    return null;
  }

  const palaces = Array.isArray(result.palaces) ? result.palaces : [];
  if (palaces.length !== 9) return null;

  const byIndex = {};
  for (let i = 0; i < palaces.length; i += 1) {
    const palace = normalizePalaceView(palaces[i]);
    if (!palace) return null;
    byIndex[palace.index] = palace;
  }

  const palaceGrid = PROFESSIONAL_PALACE_GRID_ORDER.map((index) => byIndex[index]).filter(Boolean);
  if (palaceGrid.length !== 9) return null;

  const chief = result.chief && typeof result.chief === "object" ? result.chief : {};
  const dun = result.dun && typeof result.dun === "object" ? result.dun : {};
  const calendar =
    result.calendar_basis && typeof result.calendar_basis === "object"
      ? result.calendar_basis
      : {};
  const ganzhi = result.ganzhi && typeof result.ganzhi === "object" ? result.ganzhi : {};
  const limits = Array.isArray(result.limits)
    ? result.limits.map((item) => normalizeTextField(item)).filter(Boolean)
    : [];

  const dunType = normalizeTextField(dun.type);
  const yuanKey = normalizeTextField(dun.yuan);
  const ganzhiParts = ["year", "month", "day", "hour"]
    .map((key) => normalizeTextField(ganzhi[key]))
    .filter(Boolean);

  return {
    isProfessionalQimen: true,
    algorithmVersion: ALGORITHM_QIMEN_V2_PROFESSIONAL,
    layoutVersion: normalizeTextField(result.layout_version || result.layout_basis),
    previewNote: PROFESSIONAL_PREVIEW_NOTE,
    palaceGrid,
    chiefView: {
      zhiFu: normalizeTextField(chief.zhi_fu, "—"),
      zhiShi: normalizeTextField(chief.zhi_shi, "—"),
      zhiFuPalace: normalizeTextField(chief.zhi_fu_palace),
      zhiShiPalace: normalizeTextField(chief.zhi_shi_palace),
    },
    dunView: {
      typeLabel:
        dunType === "yang" ? "阳遁" : dunType === "yin" ? "阴遁" : normalizeTextField(dunType, "—"),
      ju: dun.ju != null && dun.ju !== "" ? String(dun.ju) : "—",
      yuanLabel: DUN_YUAN_LABELS[yuanKey] || yuanKey || "—",
      method: normalizeTextField(dun.method || dun.source),
    },
    calendarBasisView: {
      solarTerm: normalizeTextField(calendar.solar_term, "—"),
      jieqiBasis: normalizeTextField(calendar.jieqi_basis),
      timeBasis: normalizeTextField(calendar.time_basis),
    },
    ganzhiView: {
      year: normalizeTextField(ganzhi.year),
      month: normalizeTextField(ganzhi.month),
      day: normalizeTextField(ganzhi.day),
      hour: normalizeTextField(ganzhi.hour),
      display: ganzhiParts.length ? ganzhiParts.join(" · ") : "—",
    },
    boundaryNotes: PROFESSIONAL_BOUNDARY_NOTES.slice(),
    professionalLimits: limits,
  };
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
  const professional = buildProfessionalQimenView(result);

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
    isProfessionalQimen: !!(professional && professional.isProfessionalQimen),
    professional: professional || null,
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
    professionalPosterNote: view.isProfessionalQimen ? PROFESSIONAL_POSTER_NOTE : "",
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

function buildQimenHistoryListItem(item) {
  if (!item || !item.id) return null;

  return {
    key: `qimen-${item.id}`,
    recordType: "qimen",
    id: item.id,
    typeLabel: MODULE_QIMEN_LABEL,
    title: "奇门问事记录",
    summary: "关注主题 · 行动节奏与需留意整理",
    statusText: Number(item.unlock_status) === 1 ? "已查看完整报告" : "已生成",
    created_at: item.created_at || "",
    createdAtDisplay: formatDateTime(item.created_at) || "—",
    detailUrl: `/pages/qimen-result/qimen-result?id=${item.id}`,
    canDelete: true,
  };
}

module.exports = {
  ALGORITHM_QIMEN_SIMPLE_V1,
  ALGORITHM_QIMEN_V2_PROFESSIONAL,
  MAX_QUESTION_LENGTH,
  FREE_CONTENT_POSTER_NOTE,
  METHOD_NOTE,
  MIN_QUESTION_LENGTH,
  MODULE_QIMEN_LABEL,
  MODULE_QIMEN_TYPE,
  QIMEN_CATEGORIES,
  QUESTION_SUMMARY,
  buildProfessionalQimenView,
  buildQimenView,
  buildQimenHistoryListItem,
  buildQimenLongPosterData,
  getCategoryLabel,
  isQimenRecord,
  listRecordSubtitle,
};
