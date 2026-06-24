const {
  getDivination,
  getFreeInterpretation,
  getFullInterpretation,
  unlockDivination,
} = require("../../utils/api");
const { getAdConfig, getCurrentEnvironment } = require("../../utils/config");
const { createRewardedAdController } = require("../../utils/rewarded-ad");
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

const AD_RESULT_MESSAGES = Object.freeze({
  cancelled: "完整观看后才能解锁",
  disabled: "当前环境暂未开启视频解锁",
  invalid_config: "视频解锁配置未完成",
  load_failed: "视频加载失败，请稍后再试",
  show_failed: "视频展示失败，请稍后再试",
  unsupported: "当前微信版本暂不支持视频解锁",
  busy: "正在处理中，请稍候",
  page_unloaded: "",
});

function getAdResultMessage(reason) {
  return AD_RESULT_MESSAGES[reason] || AD_RESULT_MESSAGES.cancelled;
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
    isDevEnv: false,
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
    this.setData({ id, isDevEnv: getCurrentEnvironment() === "dev" });
    this.pageUnloaded = false;
    this.unlockFlowToken = 0;
    this.scrollTimerId = null;
    this.rewardedAdController = createRewardedAdController({
      ...getAdConfig(),
      env: getCurrentEnvironment(),
    });
    this.loadResult();
  },

  onUnload() {
    this.pageUnloaded = true;
    this.unlockFlowToken += 1;
    this.unlockFlowRunning = false;
    this.clearScrollTimer();
    if (this.rewardedAdController) {
      this.rewardedAdController.dispose();
      this.rewardedAdController = null;
    }
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

  showAdResultToast(reason) {
    if (this.pageUnloaded || reason === "page_unloaded") return;
    const message = getAdResultMessage(reason);
    if (!message) return;
    wx.showToast({ title: message, icon: "none" });
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
    if (this.pageUnloaded) return;
    const parsed = parseFullContent(content);
    this.safeSetData({
      fullStatus: "loaded",
      fullReport: parsed.report,
      fullFallbackText: parsed.fallbackText,
      aiProvider,
      fullError: "",
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
        unlockType: "rewarded_video_mock",
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
      wx.showToast({ title: "完整解读已解锁", icon: "success" });

      this.clearScrollTimer();
      this.scrollTimerId = setTimeout(() => {
        if (!this.isFlowActive(flowToken)) return;
        wx.pageScrollTo({ selector: "#full-report", duration: 300 });
      }, 100);
    } catch (error) {
      if (!this.isFlowActive(flowToken)) return;
      this.safeSetData({
        fullStatus: "error",
        fullError: error?.message || "解锁失败，请稍后重试。",
      });
    } finally {
      this.endUnlockFlow(flowToken);
    }
  },

  handleUnlock() {
    if (!this.data.id) return;
    if (!this.rewardedAdController) {
      wx.showToast({ title: "广告模块暂不可用", icon: "none" });
      return;
    }

    const flowToken = this.beginUnlockFlow();
    if (flowToken === null) return;

    wx.showModal({
      title: "解锁完整解读",
      content:
        "观看一段视频，解锁完整解读。完整观看后可以解锁；中途退出不会解锁。",
      confirmText: "观看视频",
      cancelText: "取消",
      success: async (res) => {
        if (!this.isFlowActive(flowToken)) {
          this.endUnlockFlow(flowToken);
          return;
        }

        if (!res.confirm) {
          this.endUnlockFlow(flowToken);
          return;
        }

        try {
          const adResult = await this.rewardedAdController.show();
          if (!this.isFlowActive(flowToken)) return;

          if (adResult.completed !== true) {
            this.showAdResultToast(adResult.reason);
            this.endUnlockFlow(flowToken);
            return;
          }

          await this.performUnlock(flowToken);
        } catch (_error) {
          if (this.isFlowActive(flowToken)) {
            this.showAdResultToast("cancelled");
            this.endUnlockFlow(flowToken);
          }
        }
      },
      fail: () => {
        this.endUnlockFlow(flowToken);
      },
    });
  },

  async handleDevMockAdTest(event) {
    if (!this.data.isDevEnv || !this.rewardedAdController) return;

    const flowToken = this.beginUnlockFlow();
    if (flowToken === null) return;

    try {
      const outcome = event?.currentTarget?.dataset?.outcome || "completed";
      if (this.rewardedAdController.dispose) {
        this.rewardedAdController.dispose();
      }
      this.rewardedAdController = createRewardedAdController({
        ...getAdConfig(),
        env: getCurrentEnvironment(),
        mockOutcome: outcome,
      });

      const adResult = await this.rewardedAdController.show();
      if (!this.isFlowActive(flowToken)) return;

      if (adResult.completed !== true) {
        this.showAdResultToast(adResult.reason);
        this.endUnlockFlow(flowToken);
        return;
      }

      await this.performUnlock(flowToken);
    } catch (_error) {
      if (this.isFlowActive(flowToken)) {
        this.showAdResultToast("cancelled");
        this.endUnlockFlow(flowToken);
      }
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
