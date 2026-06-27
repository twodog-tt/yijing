const HOUR_BRANCHES = Object.freeze([
  { value: "zi", label: "子时 23:00-00:59" },
  { value: "chou", label: "丑时 01:00-02:59" },
  { value: "yin", label: "寅时 03:00-04:59" },
  { value: "mao", label: "卯时 05:00-06:59" },
  { value: "chen", label: "辰时 07:00-08:59" },
  { value: "si", label: "巳时 09:00-10:59" },
  { value: "wu", label: "午时 11:00-12:59" },
  { value: "wei", label: "未时 13:00-14:59" },
  { value: "shen", label: "申时 15:00-16:59" },
  { value: "you", label: "酉时 17:00-18:59" },
  { value: "xu", label: "戌时 19:00-20:59" },
  { value: "hai", label: "亥时 21:00-22:59" },
]);

const ELEMENT_LABELS = Object.freeze({
  wood: "木",
  fire: "火",
  earth: "土",
  metal: "金",
  water: "水",
});

const MODULE_BAZI_TYPE = 1;
const MODULE_BAZI_LABEL = "八字简析";
const ALGORITHM_BAZI_SIMPLE_V1 = "bazi-simple-v1";
const ALGORITHM_BAZI_V2_POC = "bazi-v2-poc";

const BAZI_V2_PREVIEW_NOTE =
  "节气口径预览 · 立春换年与节气月柱试行观察 · 传统文化学习参考 · 不构成现实决策依据";

const BAZI_V2_POSTER_NOTE =
  "本记录采用节气观察口径，详情页可查看节气与四柱结构说明。";

const BAZI_V2_BOUNDARY_NOTES = Object.freeze([
  "当前为节气观察试行口径，不等同于最终权威专业排盘",
  "真太阳时未实现",
  "节气时刻仍为公式近似",
  "建议结合现实情况判断",
  "不构成现实决策依据",
]);

const YEAR_BOUNDARY_LABELS = Object.freeze({
  lichun: "立春换年",
});

const MONTH_BOUNDARY_LABELS = Object.freeze({
  solar_terms_jie: "十二节令月柱",
});

const DAY_PILLAR_BASIS_LABELS = Object.freeze({
  fixed_epoch_v1: "固定基准日",
});

const { pickPosterActionPoints, limitPosterText } = require("./long-poster-canvas");
const { formatDateTime } = require("./date");
const { sanitizeInternalTerms, sanitizeInternalTermList } = require("./display-text");

function parseJSONField(raw) {
  if (!raw) return {};
  if (typeof raw === "object") return raw;
  if (typeof raw !== "string") return {};
  try {
    return JSON.parse(raw);
  } catch (error) {
    return {};
  }
}

function normalizeTextField(value, fallback = "") {
  const text = String(value == null ? "" : value).trim();
  return text || fallback;
}

function displayTextField(value, fallback = "") {
  const text = sanitizeInternalTerms(value).trim();
  return text || fallback;
}

function buildFiveElementsView(raw) {
  const elements = raw && typeof raw === "object" ? raw : {};
  return {
    wood: Number(elements.wood) || 0,
    fire: Number(elements.fire) || 0,
    earth: Number(elements.earth) || 0,
    metal: Number(elements.metal) || 0,
    water: Number(elements.water) || 0,
  };
}

function buildCalendarBasisView(raw) {
  const basis = raw && typeof raw === "object" ? raw : {};
  const yearKey = normalizeTextField(basis.year_boundary);
  const monthKey = normalizeTextField(basis.month_boundary);
  const dayKey = normalizeTextField(basis.day_pillar_basis);
  return {
    yearBoundaryLabel: YEAR_BOUNDARY_LABELS[yearKey] || displayTextField(yearKey, "—"),
    monthBoundaryLabel: MONTH_BOUNDARY_LABELS[monthKey] || displayTextField(monthKey, "—"),
    dayPillarBasisLabel: DAY_PILLAR_BASIS_LABELS[dayKey] || displayTextField(dayKey, "—"),
    trueSolarTimeLabel:
      basis.true_solar_time === true
        ? "已启用"
        : basis.true_solar_time === false
          ? "未实现"
          : "—",
    note: displayTextField(basis.note),
  };
}

function buildPillarsV2View(pillarsV2, pillarsFallback) {
  const v2 = pillarsV2 && typeof pillarsV2 === "object" ? pillarsV2 : {};
  const fallback = pillarsFallback && typeof pillarsFallback === "object" ? pillarsFallback : {};
  const hour = normalizeTextField(v2.hour || fallback.hour);
  return {
    year: normalizeTextField(v2.year || fallback.year, "—"),
    month: normalizeTextField(v2.month || fallback.month, "—"),
    day: normalizeTextField(v2.day || fallback.day, "—"),
    hour,
    hourUnknown: !hour,
  };
}

function buildBaziV2View(result) {
  if (normalizeTextField(result?.algorithm_version) !== ALGORITHM_BAZI_V2_POC) {
    return null;
  }

  const limits = Array.isArray(result?.calculation_meta?.limits)
    ? result.calculation_meta.limits.map((item) => normalizeTextField(item)).filter(Boolean)
    : Array.isArray(result?.limits)
      ? result.limits.map((item) => normalizeTextField(item)).filter(Boolean)
      : [];

  return {
    isBaziV2: true,
    algorithmVersion: ALGORITHM_BAZI_V2_POC,
    previewNote: BAZI_V2_PREVIEW_NOTE,
    calendarBasisView: buildCalendarBasisView(result.calendar_basis),
    pillarsV2View: buildPillarsV2View(result.pillars_v2, result.pillars),
    fiveElementsView: buildFiveElementsView(result.five_elements),
    compatibilityNote: displayTextField(result.compatibility_note),
    boundaryNotes: BAZI_V2_BOUNDARY_NOTES.slice(),
    limits: sanitizeInternalTermList(limits),
  };
}

function buildAnalysisView(record) {
  const result = parseJSONField(record?.result_payload);
  const pillars = result.pillars || {};
  const elements = result.five_elements || {};
  const hourUnknown = !normalizeTextField(pillars.hour);
  const suggestions = Array.isArray(result.action_suggestions)
    ? result.action_suggestions
    : [];
  const profile = result.bazi_profile || {};
  const lens = result.interpretation_lens || {};
  const baziV2 = buildBaziV2View(result);

  return {
    algorithmVersion:
      result.algorithm_version || record?.algorithm_version || ALGORITHM_BAZI_SIMPLE_V1,
    methodNote:
      displayTextField(result.method_note) ||
      "本功能采用简化干支文化规则，不等同于专业八字排盘。",
    pillars: {
      year: displayTextField(pillars.year, "—"),
      month: displayTextField(pillars.month, "—"),
      day: displayTextField(pillars.day, "—"),
      hour: displayTextField(pillars.hour),
    },
    hourUnknown,
    dayMaster: displayTextField(result.day_master, "—"),
    elements: buildFiveElementsView(elements),
    baziProfile: {
      dayMasterObservation: displayTextField(profile.day_master_observation),
      seasonTendency: displayTextField(profile.season_tendency),
      elementBalanceType: displayTextField(profile.element_balance_type),
      actionStyle: displayTextField(profile.action_style),
      reflectionTheme: displayTextField(profile.reflection_theme),
    },
    interpretationLens: {
      strengthHint: displayTextField(lens.strength_hint),
      cautionHint: displayTextField(lens.caution_hint),
      pacingHint: displayTextField(lens.pacing_hint),
      relationshipWithElements: displayTextField(lens.relationship_with_elements),
    },
    reflectionFocus: displayTextField(result.reflection_focus),
    actionSuggestions: sanitizeInternalTermList(suggestions),
    freeContent: displayTextField(record?.free_content),
    isBaziV2: !!(baziV2 && baziV2.isBaziV2),
    baziV2: baziV2 || null,
  };
}

function summarizeText(text, maxLength = 72) {
  const normalized = String(text || "").replace(/\s+/g, " ").trim();
  if (!normalized) return "";
  if (normalized.length <= maxLength) return normalized;
  return `${normalized.slice(0, maxLength)}…`;
}

function buildElementSummary(elements) {
  return ["wood", "fire", "earth", "metal", "water"]
    .map((key) => `${ELEMENT_LABELS[key]} ${Number(elements?.[key]) || 0}`)
    .join(" · ");
}

function buildBaziCardData(recordId, view) {
  if (!recordId || !view) return null;

  const pillars = {
    year: view.pillars?.year || "—",
    month: view.pillars?.month || "—",
    day: view.pillars?.day || "—",
    hour: view.pillars?.hour || "",
  };
  const hourUnknown =
    Boolean(view.hourUnknown) || !String(pillars.hour || "").trim();

  const suggestions = Array.isArray(view.actionSuggestions)
    ? view.actionSuggestions
        .slice(0, 2)
        .map((item) => summarizeText(item, 56))
        .filter(Boolean)
    : [];

  const reflectionFocus =
    summarizeText(view.reflectionFocus, 80) ||
    "可从简化干支示意与五行倾向出发，做自我观察与行动整理。";

  return {
    id: String(recordId),
    pillars,
    hourUnknown,
    elementSummary: buildElementSummary(view.elements),
    reflectionFocus,
    actionSuggestions:
      suggestions.length > 0
        ? suggestions
        : ["记录一周内的精力变化，做行动整理。"],
  };
}

function buildBaziLongPosterData(recordId, view, fullContent) {
  if (!recordId || !view) return null;

  const pillars = {
    year: view.pillars?.year || "—",
    month: view.pillars?.month || "—",
    day: view.pillars?.day || "—",
    hour: view.pillars?.hour || "",
  };
  const hourUnknown =
    Boolean(view.hourUnknown) || !String(pillars.hour || "").trim();
  const profile = view.baziProfile || {};
  const lens = view.interpretationLens || {};

  const actionPoints = pickPosterActionPoints({
    suggestions: view.actionSuggestions,
    fullContent,
    sectionHints: ["行动", "近期行动"],
    fallback: [
      profile.reflectionTheme
        ? `围绕「${limitPosterText(profile.reflectionTheme, 24)}」安排一件本周可完成的小事。`
        : "",
      "记录一周内的精力变化，做行动整理。",
    ].filter(Boolean),
    maxItems: 3,
    maxLength: 72,
  });

  return {
    id: String(recordId),
    methodNote:
      view.methodNote || "本功能采用简化干支文化规则，不等同于专业八字排盘。",
    v2PosterNote: view.isBaziV2 ? BAZI_V2_POSTER_NOTE : "",
    pillars,
    hourUnknown,
    dayMaster: view.dayMaster || "—",
    elementSummary: buildElementSummary(view.elements),
    baziProfile: {
      elementBalanceType: limitPosterText(profile.elementBalanceType, 32),
      actionStyle: limitPosterText(profile.actionStyle, 32),
      reflectionTheme: limitPosterText(profile.reflectionTheme, 32),
      seasonTendency: limitPosterText(profile.seasonTendency, 40),
    },
    interpretationLens: {
      strengthHint: limitPosterText(lens.strengthHint, 72),
      cautionHint: limitPosterText(lens.cautionHint, 72),
      pacingHint: limitPosterText(lens.pacingHint, 72),
    },
    actionPoints,
  };
}

function buildBaziHistoryListItem(item) {
  if (!item || !item.id) return null;

  return {
    key: `bazi-${item.id}`,
    recordType: "bazi",
    id: item.id,
    typeLabel: MODULE_BAZI_LABEL,
    title: "八字简析记录",
    summary: "解读视角 · 五行倾向与行动整理",
    statusText: Number(item.unlock_status) === 1 ? "已查看完整报告" : "已生成",
    created_at: item.created_at || "",
    createdAtDisplay: formatDateTime(item.created_at) || "—",
    detailUrl: `/pages/analysis-result/analysis-result?id=${item.id}`,
    canDelete: true,
  };
}

module.exports = {
  ALGORITHM_BAZI_SIMPLE_V1,
  ALGORITHM_BAZI_V2_POC,
  BAZI_V2_POSTER_NOTE,
  ELEMENT_LABELS,
  HOUR_BRANCHES,
  MODULE_BAZI_LABEL,
  MODULE_BAZI_TYPE,
  buildAnalysisView,
  buildBaziCardData,
  buildBaziHistoryListItem,
  buildBaziLongPosterData,
  buildBaziV2View,
  parseJSONField,
};
