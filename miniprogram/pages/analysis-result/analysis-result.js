const { getAnalysis, deleteAnalysis, unlockAnalysis } = require("../../utils/api");
const { getAdConfig, getCurrentEnvironment } = require("../../utils/config");
const { createRewardedAdController } = require("../../utils/rewarded-ad");
const {
  buildAnalysisView,
  buildBaziLongPosterData,
  ELEMENT_LABELS,
  MODULE_BAZI_LABEL,
} = require("../../utils/bazi");
const { formatDateTime } = require("../../utils/date");
const { ERROR_TYPES, isBusinessError } = require("../../utils/request");

const AD_RESULT_MESSAGES = Object.freeze({
  cancelled: "需要完整观看后才能解锁",
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

function mapLoadError(error) {
  if (!error) return "加载失败，请稍后重试。";
  if (error.type === ERROR_TYPES.NETWORK) {
    return "网络异常，请检查网络后重试。";
  }
  if (isBusinessError(error, 40001)) {
    return "会话已失效，请重新进入页面。";
  }
  if (isBusinessError(error, 40401)) {
    return "记录不存在或已被删除。";
  }
  return error.message || "加载失败，请稍后重试。";
}

function mapDeleteError(error) {
  if (!error) return "删除失败，请稍后重试。";
  if (error.type === ERROR_TYPES.NETWORK) {
    return "网络异常，请检查网络后重试。";
  }
  if (isBusinessError(error, 40001)) {
    return "会话已失效，请重新进入页面。";
  }
  if (isBusinessError(error, 40401)) {
    return "记录不存在或已被删除。";
  }
  return error.message || "删除失败，请稍后重试。";
}

function mapUnlockError(error) {
  if (!error) return "解锁失败，请稍后重试。";
  if (error.type === ERROR_TYPES.NETWORK) {
    return "网络异常，请检查网络后重试。";
  }
  if (isBusinessError(error, 40001)) {
    return "会话已失效，请重新进入页面。";
  }
  if (isBusinessError(error, 40301)) {
    return "当前记录暂不支持解锁完整报告。";
  }
  if (isBusinessError(error, 40401)) {
    return "记录不存在或已被删除。";
  }
  return error.message || "解锁失败，请稍后重试。";
}

Page({
  data: {
    loading: true,
    error: "",
    recordId: "",
    createdAtDisplay: "",
    moduleLabel: MODULE_BAZI_LABEL,
    privacyNote: "出生信息已用于本次简析",
    view: null,
    elementRows: [],
    fullStatus: "locked",
    fullContent: "",
    isUnlocked: false,
    unlockError: "",
    unlockFlowRunning: false,
    unlocking: false,
    deleting: false,
    deleteError: "",
    cardData: null,
    cardGenerating: false,
    resultReady: false,
    fullRevealActive: false,
  },

  onLoad(options) {
    const recordId = String(options?.id || "").trim();
    if (!recordId) {
      this.setData({
        loading: false,
        error: "缺少记录编号，无法加载结果。",
      });
      return;
    }

    this.setData({ recordId });
    this.pageUnloaded = false;
    this.unlockFlowToken = 0;
    this.unlockFlowRunning = false;
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

  showAdResultToast(reason) {
    if (this.pageUnloaded || reason === "page_unloaded") return;
    const message = getAdResultMessage(reason);
    if (!message) return;
    wx.showToast({ title: message, icon: "none" });
  },

  beginUnlockFlow() {
    if (this.pageUnloaded || this.data.deleting || this.data.cardGenerating) return null;
    if (this.unlockFlowRunning) {
      wx.showToast({ title: "正在处理中，请稍候", icon: "none" });
      return null;
    }

    this.unlockFlowRunning = true;
    this.unlockFlowToken += 1;
    const flowToken = this.unlockFlowToken;
    this.safeSetData({ unlockFlowRunning: true, unlockError: "" });
    return flowToken;
  },

  endUnlockFlow(flowToken) {
    if (flowToken != null && flowToken !== this.unlockFlowToken) return;
    this.unlockFlowRunning = false;
    if (this.pageUnloaded) return;
    this.safeSetData({ unlockFlowRunning: false, unlocking: false });
  },

  buildElementRows(elements) {
    return ["wood", "fire", "earth", "metal", "water"].map((key) => ({
      key,
      label: ELEMENT_LABELS[key],
      value: elements[key] || 0,
    }));
  },

  applyFullContent(content) {
    if (this.pageUnloaded) return;
    const fullContent = String(content || "").trim();
    if (!fullContent) return;
    this.safeSetData({
      fullStatus: "loaded",
      fullContent,
      isUnlocked: true,
      fullRevealActive: true,
      unlockError: "",
      cardData: buildBaziLongPosterData(
        this.data.recordId,
        this.data.view,
        fullContent
      ),
    });
  },

  async loadResult() {
    this.setData({
      loading: true,
      error: "",
      deleteError: "",
      unlockError: "",
      resultReady: false,
      fullRevealActive: false,
    });

    try {
      const record = await getAnalysis(this.data.recordId);
      const view = buildAnalysisView(record);
      const isUnlocked = Number(record.unlock_status) === 1;
      const fullContent = isUnlocked ? String(record.full_content || "").trim() : "";

      this.setData({
        loading: false,
        view,
        createdAtDisplay: formatDateTime(record.created_at) || "—",
        elementRows: this.buildElementRows(view.elements),
        isUnlocked,
        fullStatus: isUnlocked ? "loaded" : "locked",
        fullContent,
        fullRevealActive: isUnlocked && !!fullContent,
        resultReady: true,
        cardData: isUnlocked
          ? buildBaziLongPosterData(this.data.recordId, view, fullContent)
          : null,
      });
    } catch (error) {
      this.setData({
        loading: false,
        error: mapLoadError(error),
        view: null,
        isUnlocked: false,
        fullStatus: "locked",
        fullContent: "",
        resultReady: false,
        fullRevealActive: false,
        cardData: null,
      });
    }
  },

  async handleRepairFullReport() {
    if (
      !this.data.recordId ||
      !this.data.isUnlocked ||
      String(this.data.fullContent || "").trim() ||
      this.data.unlocking ||
      this.data.deleting ||
      this.data.cardGenerating
    ) {
      return;
    }

    this.setData({ unlocking: true, unlockError: "" });
    try {
      const unlockResult = await unlockAnalysis(this.data.recordId, {
        unlockType: "rewarded_video_mock",
      });
      const content = unlockResult?.full_content;
      if (!content) {
        throw new Error("完整报告暂未返回，请稍后重新加载。");
      }
      this.applyFullContent(content);
      wx.showToast({ title: "完整报告已恢复", icon: "success" });
    } catch (error) {
      this.setData({
        unlockError: mapUnlockError(error),
      });
    } finally {
      this.setData({ unlocking: false });
    }
  },

  async performUnlock(flowToken) {
    if (!this.data.recordId || !this.isFlowActive(flowToken)) {
      this.endUnlockFlow(flowToken);
      return;
    }

    this.safeSetData({ unlocking: true, unlockError: "" });

    try {
      const unlockResult = await unlockAnalysis(this.data.recordId, {
        unlockType: "rewarded_video_mock",
      });
      if (!this.isFlowActive(flowToken)) return;

      const content = unlockResult?.full_content;
      if (!content) {
        throw new Error("完整报告暂未返回，请稍后重新加载。");
      }

      this.applyFullContent(content);
      wx.showToast({ title: "完整报告已解锁", icon: "success" });
    } catch (error) {
      if (!this.isFlowActive(flowToken)) return;
      this.safeSetData({
        fullStatus: "locked",
        unlockError: mapUnlockError(error),
      });
    } finally {
      this.endUnlockFlow(flowToken);
    }
  },

  handleUnlock() {
    if (!this.data.recordId || this.data.isUnlocked) return;
    if (this.data.deleting || this.data.cardGenerating) return;
    if (!this.rewardedAdController) {
      wx.showToast({ title: "广告模块暂不可用", icon: "none" });
      return;
    }

    const flowToken = this.beginUnlockFlow();
    if (flowToken === null) return;

    wx.showModal({
      title: "解锁完整报告",
      content:
        "观看一段视频，解锁完整报告。完整观看后可以解锁；中途退出不会解锁。",
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

  handleDelete() {
    if (this.data.deleting || this.data.unlockFlowRunning || this.data.cardGenerating || !this.data.recordId) {
      return;
    }

    wx.showModal({
      title: "确认删除",
      content: "删除后不可恢复，是否确认删除？",
      confirmColor: "#b91c1c",
      success: async (result) => {
        if (!result.confirm) return;
        await this.deleteRecord();
      },
    });
  },

  async deleteRecord() {
    if (this.data.deleting || this.data.unlockFlowRunning || this.data.cardGenerating) return;

    this.setData({
      deleting: true,
      deleteError: "",
    });

    try {
      await deleteAnalysis(this.data.recordId);
      this.setData({ deleting: false });
      wx.showToast({ title: "已删除", icon: "success" });

      const pages = getCurrentPages();
      if (pages.length > 1) {
        wx.navigateBack({ delta: 1 });
      } else {
        wx.redirectTo({ url: "/pages/bazi/bazi" });
      }
    } catch (error) {
      this.setData({
        deleting: false,
        deleteError: mapDeleteError(error),
      });
    }
  },

  async handleGenerateCard() {
    if (
      this.data.loading ||
      this.data.cardGenerating ||
      this.data.deleting ||
      this.data.unlockFlowRunning ||
      this.data.unlocking ||
      !this.data.isUnlocked ||
      !this.data.view ||
      !this.data.recordId
    ) {
      return;
    }

    const cardData =
      buildBaziLongPosterData(
        this.data.recordId,
        this.data.view,
        this.data.fullContent
      ) || this.data.cardData;
    if (!cardData?.id) {
      wx.showToast({ title: "长图数据暂不可用，请刷新后重试", icon: "none" });
      return;
    }

    const card = this.selectComponent("#baziShareCard");
    if (!card || typeof card.open !== "function") {
      wx.showToast({ title: "长图画布初始化失败，请重新进入页面", icon: "none" });
      return;
    }

    this.setData({ cardGenerating: true });
    try {
      await card.open(cardData);
    } finally {
      if (!this.pageUnloaded) {
        this.setData({ cardGenerating: false });
      }
    }
  },

  onShareAppMessage() {
    const { view, error, loading } = this.data;
    if (loading || error || !view) {
      return {
        title: "文易传统文化",
        path: "/pages/bazi/bazi",
      };
    }
    return {
      title: "一份传统文化视角的八字简析",
      path: "/pages/bazi/bazi",
    };
  },
});
