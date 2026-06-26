const POSTER_WIDTH = 600;
const CONTENT_WIDTH = 504;
const PADDING_X = 48;
const MIN_POSTER_HEIGHT = 960;
const MAX_CANVAS_HEIGHT = 12000;
const FOOTER_RESERVE = 132;
const TRUNCATION_NOTICE =
  "内容较长，长图仅展示前半部分内容；完整内容请在小程序结果页查看。";

const POSTER_DISCLAIMER =
  "内容仅供传统文化学习、自我观察与行动节奏整理，不构成现实决策依据，建议结合现实情况判断。";

const FORBIDDEN_POSTER_PHRASES = Object.freeze([
  "精准预测",
  "必成",
  "必败",
  "大吉",
  "大凶",
  "必发财",
  "必复合",
  "改运",
  "化灾",
  "转运",
  "投资建议",
  "医疗建议",
  "法律建议",
  "赌博建议",
  "军事行动建议",
]);

const PRIVACY_POSTER_PATTERNS = Object.freeze([
  /session_key/i,
  /input_payload/i,
  /result_payload/i,
  /\bprompt\b/i,
  /\d{4}-\d{2}-\d{2}/,
]);

function normalizeText(value) {
  return String(value || "").replace(/\s+/g, " ").trim();
}

function containsForbiddenPosterPhrase(text) {
  const normalized = normalizeText(text);
  if (!normalized) return false;
  return FORBIDDEN_POSTER_PHRASES.some((phrase) => normalized.includes(phrase));
}

function containsPosterPrivacyLeak(text) {
  const normalized = String(text || "");
  if (!normalized) return false;
  return PRIVACY_POSTER_PATTERNS.some((pattern) => pattern.test(normalized));
}

function sanitizePosterText(text) {
  const normalized = normalizeText(text);
  if (!normalized) return "";
  if (containsForbiddenPosterPhrase(normalized)) return "";
  if (containsPosterPrivacyLeak(normalized)) return "";
  return normalized;
}

function limitPosterText(text, maxLength = 80) {
  const cleaned = sanitizePosterText(text);
  if (!cleaned) return "";
  if (cleaned.length <= maxLength) return cleaned;
  return `${cleaned.slice(0, maxLength)}…`;
}

function normalizePosterLines(items, options) {
  const { maxItems = 3, maxLength = 80 } = options || {};
  const seen = new Set();
  const lines = [];
  (Array.isArray(items) ? items : []).forEach((item) => {
    const line = limitPosterText(item, maxLength);
    if (!line || seen.has(line)) return;
    seen.add(line);
    lines.push(line);
  });
  return lines.slice(0, maxItems);
}

function extractReportSection(fullContent, sectionHints) {
  const content = String(fullContent || "").trim();
  if (!content || !Array.isArray(sectionHints) || !sectionHints.length) return "";

  const hints = sectionHints.map((hint) => String(hint || "").trim()).filter(Boolean);
  const sectionPattern =
    /(?:【?\s*[一二三四五六七八九十\d]+[、.]?\s*([^】\n]+)】?\s*\n)([\s\S]*?)(?=(?:\n\s*(?:【?\s*[一二三四五六七八九十\d]+[、.]?)|$))/g;

  let match;
  while ((match = sectionPattern.exec(content)) !== null) {
    const title = normalizeText(match[1]);
    const body = normalizeText(match[2]);
    if (!body) continue;
    if (hints.some((hint) => title.includes(hint))) {
      return body;
    }
  }
  return "";
}

function extractReportHighlights(fullContent, sectionHints, options) {
  const body = extractReportSection(fullContent, sectionHints);
  if (!body) return [];

  const { maxItems = 3, maxLength = 80 } = options || {};
  const bulletMatches = body.match(/(?:^|\n)\s*(?:[-·•]|\d+[.、)])\s*([^\n]+)/g);
  if (bulletMatches && bulletMatches.length) {
    return normalizePosterLines(
      bulletMatches.map((line) => line.replace(/^\s*(?:[-·•]|\d+[.、)])\s*/, "")),
      { maxItems, maxLength }
    );
  }

  const sentences = body
    .split(/[。；;]\s*/)
    .map((part) => limitPosterText(part, maxLength))
    .filter(Boolean);
  return normalizePosterLines(sentences, { maxItems, maxLength });
}

function pickPosterActionPoints(options) {
  const {
    suggestions = [],
    fullContent = "",
    sectionHints = ["行动"],
    fallback = [],
    maxItems = 3,
    maxLength = 72,
  } = options || {};

  const fromStructured = normalizePosterLines(suggestions, { maxItems, maxLength });
  if (fromStructured.length >= maxItems) return fromStructured;

  const fromReport = extractReportHighlights(fullContent, sectionHints, {
    maxItems: maxItems - fromStructured.length,
    maxLength,
  });

  const merged = normalizePosterLines([...fromStructured, ...fromReport], {
    maxItems,
    maxLength,
  });
  if (merged.length) return merged;

  return normalizePosterLines(fallback, { maxItems, maxLength });
}

function buildDivinationPosterSummary(options) {
  const { freeContent = "", fullReport = null, fullFallbackText = "" } = options || {};

  if (fullReport && typeof fullReport === "object") {
    const summary = limitPosterText(fullReport.summary, 120);
    if (summary) return summary;
  }

  const freeSentences = normalizePosterLines(
    String(freeContent || "")
      .split(/[。！？；;\n]/)
      .map((part) => limitPosterText(part, 80)),
    { maxItems: 2, maxLength: 80 }
  );
  if (freeSentences.length) {
    return limitPosterText(freeSentences.join("。"), 120);
  }

  const fallbackLine = limitPosterText(
    String(fullFallbackText || "")
      .split(/[。！？；;\n]/)[0],
    120
  );
  if (fallbackLine) return fallbackLine;

  return "本次卦象梳理侧重局势观察与行动节奏整理，建议结合现实情况判断。";
}

function pickDivinationChangeObservations(options) {
  const { fullReport = null, freeContent = "", maxItems = 3 } = options || {};
  const items = [];

  if (fullReport && typeof fullReport === "object") {
    if (fullReport.opportunity) {
      items.push(`可借助 · ${limitPosterText(fullReport.opportunity, 64)}`);
    }
    if (fullReport.risk) {
      items.push(`需留意 · ${limitPosterText(fullReport.risk, 64)}`);
    }
    if (fullReport.current_state) {
      items.push(limitPosterText(fullReport.current_state, 72));
    }
    if (items.length < maxItems && fullReport.overall) {
      items.push(limitPosterText(fullReport.overall, 72));
    }
  }

  if (items.length < maxItems) {
    const freeHints = extractReportHighlights(freeContent, ["变化", "局势", "处境"], {
      maxItems: maxItems - items.length,
      maxLength: 72,
    });
    items.push(...freeHints);
  }

  return normalizePosterLines(items, { maxItems, maxLength: 72 });
}

function pickDivinationReflectionQuestions(options) {
  const {
    fullReport = null,
    fullContent = "",
    freeContent = "",
    maxItems = 2,
  } = options || {};

  const structured = normalizePosterLines(fullReport?.reflection_questions, {
    maxItems,
    maxLength: 72,
  });
  if (structured.length) return structured;

  const fromReport = extractReportHighlights(fullContent, ["反思", "问题"], {
    maxItems,
    maxLength: 72,
  });
  if (fromReport.length) return fromReport;

  const fromFree = extractReportHighlights(freeContent, ["反思", "问题"], {
    maxItems,
    maxLength: 72,
  });
  if (fromFree.length) return fromFree;

  return normalizePosterLines(
    ["这次卦象变化，对你当前最相关的一点是什么？", "若只调整一个行动节奏，你会先做什么？"],
    { maxItems, maxLength: 72 }
  );
}

function normalizeDivinationPosterLines(items, options) {
  return normalizePosterLines(items, options);
}

const QIMEN_CATEGORY_HIGHLIGHTS = Object.freeze({
  career: "推进顺序 · 资源协调 · 执行风险",
  relationship: "沟通边界 · 误解修复 · 关系节奏",
  study: "复盘节奏 · 专注方法 · 阶段目标",
  decision: "信息补齐 · 小步试探 · 备用方案",
  general: "问题整理 · 风险收敛 · 小步行动",
});

function getQimenCategoryHighlight(category) {
  return QIMEN_CATEGORY_HIGHLIGHTS[category] || QIMEN_CATEGORY_HIGHLIGHTS.general;
}

function preserveNewlines(text) {
  return String(text || "")
    .replace(/\r\n/g, "\n")
    .split("\n")
    .map((line) => line.replace(/\s+/g, " ").trim())
    .join("\n");
}

function wrapAllLines(ctx, text, maxWidth) {
  const normalized = preserveNewlines(text);
  if (!normalized) return [];

  const lines = [];
  normalized.split("\n").forEach((paragraph) => {
    if (!paragraph) {
      lines.push("");
      return;
    }
    let current = "";
    for (const character of paragraph) {
      const candidate = `${current}${character}`;
      if (current && ctx.measureText(candidate).width > maxWidth) {
        lines.push(current);
        current = character;
      } else {
        current = candidate;
      }
    }
    if (current) lines.push(current);
  });
  return lines;
}

function measureWrappedText(ctx, text, maxWidth, lineHeight, font) {
  ctx.font = font;
  const lines = wrapAllLines(ctx, text, maxWidth);
  return lines.length * lineHeight;
}

function drawWrappedParagraph(ctx, text, x, y, options) {
  const {
    maxWidth = CONTENT_WIDTH,
    lineHeight = 24,
    color = "#57534e",
    font = "15px sans-serif",
    maxLines = 0,
  } = options || {};

  ctx.fillStyle = color;
  ctx.font = font;
  let lines = wrapAllLines(ctx, text, maxWidth);
  if (maxLines > 0 && lines.length > maxLines) {
    lines = lines.slice(0, maxLines);
    const last = lines[maxLines - 1];
    let output = last;
    while (output && ctx.measureText(`${output}…`).width > maxWidth) {
      output = output.slice(0, -1);
    }
    lines[maxLines - 1] = `${output}…`;
  }

  lines.forEach((line, index) => {
    ctx.fillText(line, x, y + index * lineHeight);
  });
  return y + lines.length * lineHeight;
}

function drawTitle(ctx, text, x, y, options) {
  const { font = "bold 30px sans-serif", color = "#292524" } = options || {};
  ctx.fillStyle = color;
  ctx.font = font;
  ctx.fillText(text, x, y);
  return y + 36;
}

function drawSectionTitle(ctx, text, x, y, options) {
  const { font = "bold 16px sans-serif", color = "#292524" } = options || {};
  ctx.fillStyle = color;
  ctx.font = font;
  ctx.fillText(text, x, y);
  return y + 28;
}

function drawDivider(ctx, x, y, width = CONTENT_WIDTH) {
  ctx.strokeStyle = "#e7e5e4";
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.moveTo(x, y);
  ctx.lineTo(x + width, y);
  ctx.stroke();
  return y + 20;
}

function drawRoundedRect(ctx, x, y, width, height, radius, fillStyle) {
  const safeRadius = Math.min(radius, width / 2, height / 2);
  ctx.beginPath();
  ctx.moveTo(x + safeRadius, y);
  ctx.lineTo(x + width - safeRadius, y);
  ctx.quadraticCurveTo(x + width, y, x + width, y + safeRadius);
  ctx.lineTo(x + width, y + height - safeRadius);
  ctx.quadraticCurveTo(x + width, y + height, x + width - safeRadius, y + height);
  ctx.lineTo(x + safeRadius, y + height);
  ctx.quadraticCurveTo(x, y + height, x, y + height - safeRadius);
  ctx.lineTo(x, y + safeRadius);
  ctx.quadraticCurveTo(x, y, x + safeRadius, y);
  ctx.closePath();
  ctx.fillStyle = fillStyle;
  ctx.fill();
}

function estimateParagraphHeight(text, lineHeight, charsPerLine = 28) {
  const normalized = preserveNewlines(text);
  if (!normalized) return 0;
  let lines = 0;
  normalized.split("\n").forEach((paragraph) => {
    if (!paragraph) {
      lines += 1;
      return;
    }
    lines += Math.max(1, Math.ceil(paragraph.length / charsPerLine));
  });
  return lines * lineHeight;
}

function clampHeight(height) {
  return Math.min(Math.max(height, MIN_POSTER_HEIGHT), MAX_CANVAS_HEIGHT);
}

function computePosterDimensions(rawHeight) {
  const normalizedRaw = Math.max(Number(rawHeight) || 0, MIN_POSTER_HEIGHT);
  const isTruncated = normalizedRaw > MAX_CANVAS_HEIGHT;
  return {
    rawHeight: normalizedRaw,
    canvasHeight: clampHeight(normalizedRaw),
    isTruncated,
    truncatedHint: isTruncated
      ? "内容较长，长图仅展示前半部分；完整内容请在小程序内查看"
      : "",
  };
}

function remainingLines(maxY, currentY, lineHeight) {
  if (lineHeight <= 0) return 0;
  return Math.max(0, Math.floor((maxY - currentY) / lineHeight));
}

function resolveExportPixelRatio(canvasHeight, systemPixelRatio) {
  const ratio = Number(systemPixelRatio) || 2;
  if (canvasHeight >= 10000) return 1;
  if (canvasHeight >= 8000) return Math.min(ratio, 1.5);
  if (canvasHeight >= 5000) return Math.min(ratio, 2);
  return Math.min(ratio, 3);
}

function exportCanvasToTempFile(canvas, width, height, component, pixelRatio) {
  const ratio =
    Number(pixelRatio) ||
    resolveExportPixelRatio(height, wx.getSystemInfoSync().pixelRatio || 2);
  return new Promise((resolve, reject) => {
    wx.canvasToTempFilePath(
      {
        canvas,
        x: 0,
        y: 0,
        width,
        height,
        destWidth: width * ratio,
        destHeight: height * ratio,
        fileType: "png",
        quality: 1,
        success: (result) => resolve(result.tempFilePath),
        fail: reject,
      },
      component
    );
  });
}

function getAlbumPermissionHelpers(component) {
  return {
    getSetting() {
      return new Promise((resolve, reject) => {
        wx.getSetting({ success: resolve, fail: reject });
      });
    },
    authorizeAlbum() {
      return new Promise((resolve, reject) => {
        wx.authorize({ scope: "scope.writePhotosAlbum", success: resolve, fail: reject });
      });
    },
    async guideToSettings() {
      const modal = await new Promise((resolve) => {
        wx.showModal({
          title: "需要相册权限",
          content: "保存长图需要相册权限。你可以前往设置开启，也可以取消后仅查看预览。",
          confirmText: "打开设置",
          success: resolve,
          fail: () => resolve({ confirm: false }),
        });
      });
      if (!modal.confirm) return false;
      const setting = await new Promise((resolve) => {
        wx.openSetting({ success: resolve, fail: () => resolve({ authSetting: {} }) });
      });
      return setting.authSetting?.["scope.writePhotosAlbum"] === true;
    },
    async ensureAlbumPermission() {
      const setting = await this.getSetting();
      const permission = setting.authSetting?.["scope.writePhotosAlbum"];
      if (permission === true) return true;
      if (permission === false) return this.guideToSettings();
      try {
        await this.authorizeAlbum();
        return true;
      } catch (_error) {
        return this.guideToSettings();
      }
    },
    async saveImageToAlbum(filePath) {
      const allowed = await this.ensureAlbumPermission();
      if (!allowed) {
        wx.showToast({ title: "未获得相册权限，可稍后在设置中开启", icon: "none" });
        return false;
      }
      await new Promise((resolve, reject) => {
        wx.saveImageToPhotosAlbum({ filePath, success: resolve, fail: reject });
      });
      wx.showToast({ title: "长图已保存到相册", icon: "success" });
      return true;
    },
  };
}

module.exports = {
  POSTER_WIDTH,
  CONTENT_WIDTH,
  PADDING_X,
  MIN_POSTER_HEIGHT,
  MAX_CANVAS_HEIGHT,
  FOOTER_RESERVE,
  TRUNCATION_NOTICE,
  POSTER_DISCLAIMER,
  FORBIDDEN_POSTER_PHRASES,
  normalizeText,
  sanitizePosterText,
  limitPosterText,
  normalizePosterLines,
  extractReportSection,
  extractReportHighlights,
  pickPosterActionPoints,
  buildDivinationPosterSummary,
  pickDivinationChangeObservations,
  pickDivinationReflectionQuestions,
  normalizeDivinationPosterLines,
  getQimenCategoryHighlight,
  containsForbiddenPosterPhrase,
  preserveNewlines,
  wrapAllLines,
  measureWrappedText,
  drawWrappedParagraph,
  drawTitle,
  drawSectionTitle,
  drawDivider,
  drawRoundedRect,
  estimateParagraphHeight,
  clampHeight,
  computePosterDimensions,
  remainingLines,
  resolveExportPixelRatio,
  exportCanvasToTempFile,
  getAlbumPermissionHelpers,
};
