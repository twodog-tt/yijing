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
      // 非 JSON 内容降级为文本展示。
    }
    return { report: null, fallbackText: String(content).trim() };
  }

  return { report: null, fallbackText: "" };
}

function buildPosterData(
  divination,
  freeContent,
  movingLinesDisplay,
  displayLines,
  fullReport,
  fullFallbackText
) {
  return {
    id: Number(divination?.id) || 0,
    categoryName: divination?.category?.name || "未分类",
    themeNote: "问事主题已用于本次解析",
    primaryHexagram: divination?.primary_hexagram || null,
    changedHexagram: divination?.changed_hexagram || null,
    movingLinesDisplay,
    lines: displayLines,
    freeContent: String(freeContent || "").trim(),
    fullReport: fullReport || null,
    fullFallbackText: String(fullFallbackText || "").trim(),
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
    unlockFlowRunning: false,
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
    this.pageUnloaded = false;
    this.unlockFlowToken = 0;
    this.scrollTimerId = null;
    this.loadResult();
  },

  onUnload() {
    this.pageUnloaded = true;
    this.unlockFlowToken += 1;
    this.unlockFlowRunning = false;
    this.clearScrollTimer();
  },

  safeSetData(data) {
    if (this.pageUnloaded) return;
    this.setData(data);
  },

  isFlowActive(flowToken) {
    return !this.pageUnloaded && flowToken === this.unlockFlowToken;
  },

  clearScrollTimer() {
    if (this.scrollTimerId != null) {
      clearTimeout(this.scrollTimerId);
      this.scrollTimerId = null;
    }
  },

  beginUnlockFlow() {
    if (this.pageUnloaded) return null;
    if (this.unlockFlowRunning) {
      wx.showToast({ title: "正在处理中，请稍候", icon: "none" });
      return null;
    }

    this.unlockFlowRunning = true;
    this.unlockFlowToken += 1;
    const flowToken = this.unlockFlowToken;
    this.safeSetData({ unlockFlowRunning: true });
    return flowToken;
  },

  endUnlockFlow(flowToken) {
    if (flowToken != null && flowToken !== this.unlockFlowToken) return;
    this.unlockFlowRunning = false;
    if (this.pageUnloaded) return;
    this.safeSetData({ unlockFlowRunning: false, unlocking: false });
  },

  refreshPosterData(overrides = {}) {
    const divination = overrides.divination || this.data.divination;
    if (!divination) return null;

    const posterData = buildPosterData(
      divination,
      overrides.freeContent ?? this.data.freeContent,
      overrides.movingLinesDisplay ?? this.data.movingLinesDisplay,
      overrides.displayLines ?? this.data.displayLines,
      overrides.fullReport ?? this.data.fullReport,
      overrides.fullFallbackText ?? this.data.fullFallbackText
    );
    return posterData;
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

      let fullStatus = "locked";
      let fullReport = null;
      let fullFallbackText = "";
      let aiProvider = "";
      let fullError = "";

      if (full?._loadError) {
        fullStatus = "error";
        fullError = full._loadError.message || "完整解读状态加载失败。";
      } else if (full?.unlocked) {
        const parsed = parseFullContent(full.full_content);
        fullStatus = "loaded";
        fullReport = parsed.report;
        fullFallbackText = parsed.fallbackText;
        aiProvider = full.ai_provider || "";
      }

      const posterData = buildPosterData(
        divination,
        freeContent,
        movingLinesDisplay,
        displayLines,
        fullReport,
        fullFallbackText
      );

      this.setData({
        divination,
        createdAtDisplay: formatDateTime(divination.created_at),
        movingLinesDisplay,
        displayLines,
        freeContent,
        fullStatus,
        fullReport,
        fullFallbackText,
        aiProvider,
        fullError,
        posterData,
      });
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
    if (this.pageUnloaded) return;
    const parsed = parseFullContent(content);
    const posterData = this.refreshPosterData({
      fullReport: parsed.report,
      fullFallbackText: parsed.fallbackText,
    });
    this.safeSetData({
      fullStatus: "loaded",
      fullReport: parsed.report,
      fullFallbackText: parsed.fallbackText,
      aiProvider,
      fullError: "",
      posterData,
    });
  },

  async performUnlock(flowToken) {
    if (!this.data.id || !this.isFlowActive(flowToken)) {
      this.endUnlockFlow(flowToken);
      return;
    }

    this.safeSetData({ unlocking: true, fullError: "" });

    try {
      const unlockResult = await unlockDivination(this.data.id, {
        unlockType: "mock_button",
      });
      if (!this.isFlowActive(flowToken)) return;

      let content = unlockResult?.full_interpretation;
      let aiProvider = "";

      try {
        const full = await getFullInterpretation(this.data.id);
        if (!this.isFlowActive(flowToken)) return;
        if (full.unlocked) {
          content = full.full_content;
          aiProvider = full.ai_provider || "";
        }
      } catch (_error) {
        // 解锁响应已经包含完整内容时，后续查询失败不阻塞展示。
      }

      if (!this.isFlowActive(flowToken)) return;
      if (!content) throw new Error("完整解读暂未返回，请稍后重新加载。");

      this.applyFullContent(content, aiProvider);
      wx.showToast({ title: "完整解析已加载", icon: "success" });

      this.clearScrollTimer();
      this.scrollTimerId = setTimeout(() => {
        if (!this.isFlowActive(flowToken)) return;
        wx.pageScrollTo({ selector: "#full-report", duration: 300 });
      }, 100);
    } catch (error) {
      if (!this.isFlowActive(flowToken)) return;
      this.safeSetData({
        fullStatus: "error",
        fullError: error?.message || "加载完整解析失败，请稍后重试。",
      });
    } finally {
      this.endUnlockFlow(flowToken);
    }
  },

  handleViewFullReport() {
    if (!this.data.id || this.data.fullStatus === "loaded") return;

    const flowToken = this.beginUnlockFlow();
    if (flowToken === null) return;

    this.performUnlock(flowToken);
  },

  async handleGeneratePoster() {
    if (
      this.data.posterGenerating ||
      !this.data.posterData ||
      this.data.fullStatus !== "loaded"
    ) {
      if (this.data.fullStatus !== "loaded") {
        wx.showToast({ title: "请先查看完整解析", icon: "none" });
      }
      return;
    }

    const poster = this.selectComponent("#sharePoster");
    if (!poster) {
      wx.showToast({ title: "长图组件暂不可用", icon: "none" });
      return;
    }

    this.setData({ posterGenerating: true });
    try {
      await poster.open(this.data.posterData);
    } finally {
      this.setData({ posterGenerating: false });
    }
  },

  onShareAppMessage() {
    const { id, divination, error, loading } = this.data;
    if (loading || error || !divination || !id) {
      return {
        title: "文易传统文化",
        path: "/pages/index/index",
      };
    }
    return {
      title: "一份基于传统文化的趣味解读",
      path: `/pages/result/result?id=${id}`,
    };
  },
});
