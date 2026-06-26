const {
  buildDivinationPosterSummary,
  limitPosterText,
  normalizeText,
  pickDivinationChangeObservations,
  pickDivinationReflectionQuestions,
  pickPosterActionPoints,
} = require("./long-poster-canvas");

const QUESTION_SUMMARY = "用户问题已用于本次卦象梳理";

const METHOD_NOTE =
  "卦象解读仅供传统文化学习与自我观察参考，不等同于专业判断，不构成现实决策依据。";

function hexagramLabel(hexagram, fallback) {
  return normalizeText(hexagram?.full_name || hexagram?.name || fallback);
}

function hasHexagramChange(primary, changed, movingLines) {
  const primaryName = hexagramLabel(primary, "");
  const changedName = hexagramLabel(changed, "");
  const movingCount = Array.isArray(movingLines) ? movingLines.length : 0;
  if (movingCount === 0) return false;
  if (!primaryName || !changedName) return movingCount > 0;
  return primaryName !== changedName;
}

function buildMovingHint(movingLinesDisplay, movingLines) {
  const count = Array.isArray(movingLines) ? movingLines.length : 0;
  if (count > 0) {
    return `动爻 ${count} 处 · ${normalizeText(movingLinesDisplay || "有变化")}`;
  }
  return "无明显动爻 · 卦象以本卦为主";
}

function buildFullContentText(fullReport, fullFallbackText) {
  if (fullReport && typeof fullReport === "object") {
    return [
      fullReport.summary,
      fullReport.overall,
      fullReport.current_state,
      fullReport.opportunity,
      fullReport.risk,
      fullReport.emotion_reminder,
      ...(Array.isArray(fullReport.action_steps) ? fullReport.action_steps : []),
      ...(Array.isArray(fullReport.reflection_questions)
        ? fullReport.reflection_questions
        : []),
    ]
      .filter(Boolean)
      .join("\n");
  }
  return String(fullFallbackText || "").trim();
}

function buildDivinationLongPosterData(
  recordId,
  divination,
  {
    freeContent = "",
    movingLinesDisplay = "无动爻",
    displayLines = [],
    movingLines = [],
    fullReport = null,
    fullFallbackText = "",
  } = {}
) {
  if (!recordId || !divination) return null;

  const primaryHexagram = divination.primary_hexagram || null;
  const changedHexagram = divination.changed_hexagram || null;
  const hasChange = hasHexagramChange(primaryHexagram, changedHexagram, movingLines);
  const fullContentText = buildFullContentText(fullReport, fullFallbackText);

  const situationSummary = buildDivinationPosterSummary({
    freeContent,
    fullReport,
    fullFallbackText,
  });

  const changeObservations = pickDivinationChangeObservations({
    fullReport,
    freeContent,
    maxItems: 3,
  });

  const actionPoints = pickPosterActionPoints({
    suggestions: Array.isArray(fullReport?.action_steps) ? fullReport.action_steps : [],
    fullContent: fullContentText,
    sectionHints: ["行动", "建议", "近期行动"],
    fallback: [
      "先把问题拆小，安排一件今天能完成的小事。",
      "记录一次沟通或推进感受，再决定是否调整节奏。",
      "留意信息是否补齐，避免在不确定时一次做满决策。",
    ],
    maxItems: 3,
    maxLength: 72,
  });

  const reflectionQuestions = pickDivinationReflectionQuestions({
    fullReport,
    fullContent: fullContentText,
    freeContent,
    maxItems: 2,
  });

  return {
    id: String(recordId),
    moduleTitle: "问事起卦",
    methodNote: METHOD_NOTE,
    categoryName: normalizeText(divination.category?.name || "未分类"),
    questionSummary: QUESTION_SUMMARY,
    primaryHexagram,
    changedHexagram,
    changedHexagramLabel: hasChange
      ? hexagramLabel(changedHexagram, "变卦")
      : "无明显变卦",
    hasHexagramChange: hasChange,
    movingLinesDisplay: normalizeText(movingLinesDisplay || "无动爻"),
    movingHint: buildMovingHint(movingLinesDisplay, movingLines),
    lines: Array.isArray(displayLines) ? displayLines : [],
    situationSummary: limitPosterText(situationSummary, 120),
    changeObservations,
    actionPoints,
    reflectionQuestions,
  };
}

module.exports = {
  METHOD_NOTE,
  QUESTION_SUMMARY,
  buildDivinationLongPosterData,
};
