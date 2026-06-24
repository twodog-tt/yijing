const {
  getDivination,
  getFreeInterpretation,
  getFullInterpretation,
  unlockDivination,
} = require("../../utils/api");
const { formatDateTime } = require("../../utils/date");

const POSITION_LABELS = ["", "初爻", "二爻", "三爻", "四爻", "五爻", "上爻"];

function lineMeaning(value) {
  const meanings = {
    6: "老阴",
    7: "少阳",
    8: "少阴",
    9: "老阳",
  };
  return meanings[Number(value)] || "爻象";
}

function prepareLines(lines) {
  return (Array.isArray(lines) ? lines : [])
    .map((line) => ({
      ...line,
      position: Number(line.position),
      position_label: POSITION_LABELS[Number(line.position)] || `第${line.position}爻`,
      is_yang: Number(line.is_yang) === 1,
      is_moving: Number(line.is_moving) === 1,
      meaning: lineMeaning(line.value),
    }))
    .sort((a, b) => b.position - a.position);
}

function summarizeText(text, maxLength = 500) {
  const normalized = String(text || "").replace(/\s+/g, " ").trim();
  if (normalized.length <= maxLength) return normalized;
  return `${normalized.slice(0, maxLength)}…`;
}

function normalizeReport(report) {
  if (!report || typeof report !== "object" || Array.isArray(report)) return null;
  return {
    summary: String(report.summary || ""),
    overall: String(report.overall || ""),
    current_state: String(report.current_state || ""),
    opportunity: String(report.opportunity || ""),
    risk: String(report.risk || ""),
    action_steps: Array.isArray(report.action_steps)
      ? report.action_steps.map((item) => String(item))
      : [],
    emotion_reminder: String(report.emotion_reminder || ""),
    reflection_questions: Array.isArray(report.reflection_questions)
      ? report.reflection_questions.map((item) => String(item))
      : [],
    disclaimer: String(report.disclaimer || ""),
  };
}

function parseFullContent(content) {
  if (content && typeof content === "object") {
    return { report: normalizeReport(content), fallbackText: "" };
  }

  if (typeof content === "string" && content.trim()) {
    try {
      const parsed = JSON.parse(content);
      const report = normalizeReport(parsed);
      if (report) return { report, fallbackText: "" };
    } catch (_error) {
      // 非 JSON 内容降级为文本摘要，避免影响结果页。
    }
    return { report: null, fallbackText: summarizeText(content) };
  }

  return { report: null, fallbackText: "" };
}

function buildPosterData(divination, freeContent, movingLinesDisplay, displayLines) {
  return {
    id: Number(divination?.id) || 0,
    categoryName: divination?.category?.name || "未分类",
    question: summarizeText(divination?.question || "一次卦象记录", 72),
    primaryHexagram: divination?.primary_hexagram || null,
    changedHexagram: divination?.changed_hexagram || null,
    movingLinesDisplay,
    lines: displayLines,
    freeSummary: summarizeText(freeContent, 150),
  };
}

Page({
  data: {
    id: 0,
    loading: true,
    error: "",
    divination: null,
    createdAtDisplay: "",
    movingLinesDisplay: "无动爻",
    displayLines: [],
    freeContent: "",
    fullStatus: "locked",
    fullError: "",
    fullReport: null,
    fullFallbackText: "",
    aiProvider: "",
    unlocking: false,
    loadingFull: false,
    posterData: null,
    posterGenerating: false,
  },

  onLoad(options) {
    const id = Number(options.id);
    if (!Number.isInteger(id) || id <= 0) {
      this.setData({
        loading: false,
        error: "记录 ID 无效，请从历史记录或起卦流程重新进入。",
      });
      return;
    }
    this.setData({ id });
    this.loadResult();
  },

  async loadResult() {
    if (!this.data.id) return;
    this.setData({ loading: true, error: "", fullError: "" });

    try {
      const fullPromise = getFullInterpretation(this.data.id).catch((error) => ({
        _loadError: error,
      }));
      const [divination, free, full] = await Promise.all([
        getDivination(this.data.id),
        getFreeInterpretation(this.data.id),
        fullPromise,
      ]);

      const movingLines = Array.isArray(divination.moving_lines)
        ? divination.moving_lines
        : [];
      const movingLinesDisplay = movingLines.length
        ? `第 ${movingLines.join("、")} 爻`
        : "无动爻";
      const displayLines = prepareLines(divination.lines);
      const freeContent =
        free?.free_content || divination.free_interpretation || "暂无免费解读。";
      const nextData = {
        divination,
        createdAtDisplay: formatDateTime(divination.created_at),
        movingLinesDisplay,
        displayLines,
        freeContent,
        posterData: buildPosterData(
          divination,
          freeContent,
          movingLinesDisplay,
          displayLines
        ),
      };

      if (full?._loadError) {
        nextData.fullStatus = "error";
        nextData.fullError = full._loadError.message || "完整解读状态加载失败。";
      } else if (full?.unlocked) {
        const parsed = parseFullContent(full.full_content);
        nextData.fullStatus = "loaded";
        nextData.fullReport = parsed.report;
        nextData.fullFallbackText = parsed.fallbackText;
        nextData.aiProvider = full.ai_provider || "";
      } else {
        nextData.fullStatus = "locked";
      }

      this.setData(nextData);
    } catch (error) {
      this.setData({
        error: error?.message || "结果加载失败，请稍后重试。",
      });
    } finally {
      this.setData({ loading: false });
    }
  },

  async loadFullOnly() {
    if (!this.data.id || this.data.loadingFull) return;
    this.setData({ loadingFull: true, fullError: "" });
    try {
      const full = await getFullInterpretation(this.data.id);
      if (!full.unlocked) {
        this.setData({ fullStatus: "locked" });
        return;
      }
      this.applyFullContent(full.full_content, full.ai_provider || "");
    } catch (error) {
      this.setData({
        fullStatus: "error",
        fullError: error?.message || "完整解读加载失败，请稍后重试。",
      });
    } finally {
      this.setData({ loadingFull: false });
    }
  },

  applyFullContent(content, aiProvider = "") {
    const parsed = parseFullContent(content);
    this.setData({
      fullStatus: "loaded",
      fullReport: parsed.report,
      fullFallbackText: parsed.fallbackText,
      aiProvider,
      fullError: "",
    });
  },

  async handleUnlock() {
    if (!this.data.id || this.data.unlocking) return;
    this.setData({ unlocking: true, fullError: "" });

    try {
      const unlockResult = await unlockDivination(this.data.id);
      let content = unlockResult?.full_interpretation;
      let aiProvider = "";

      try {
        const full = await getFullInterpretation(this.data.id);
        if (full.unlocked) {
          content = full.full_content;
          aiProvider = full.ai_provider || "";
        }
      } catch (_error) {
        // 解锁响应已经包含完整内容时，后续查询失败不阻塞展示。
      }

      if (!content) throw new Error("完整解读暂未返回，请稍后重新加载。" );
      this.applyFullContent(content, aiProvider);
      wx.showToast({ title: "完整解读已解锁", icon: "success" });
      setTimeout(() => {
        wx.pageScrollTo({ selector: "#full-report", duration: 300 });
      }, 100);
    } catch (error) {
      this.setData({
        fullStatus: "error",
        fullError: error?.message || "模拟解锁失败，请稍后重试。",
      });
    } finally {
      this.setData({ unlocking: false });
    }
  },

  async handleGeneratePoster() {
    if (this.data.posterGenerating || !this.data.posterData) return;
    const poster = this.selectComponent("#sharePoster");
    if (!poster) {
      wx.showToast({ title: "海报组件暂不可用", icon: "none" });
      return;
    }

    this.setData({ posterGenerating: true });
    try {
      await poster.open();
    } finally {
      this.setData({ posterGenerating: false });
    }
  },

  onShareAppMessage() {
    return {
      title: "一份基于传统文化的趣味解读",
      path: `/pages/result/result?id=${this.data.id}`,
    };
  },
});
