const { deleteAnalysis, getAnalysis, unlockAnalysis } = require("../../utils/api");
const { formatDateTime } = require("../../utils/date");
const { isBusinessError } = require("../../utils/request");
const {
  CONTENT_LOAD_ERROR_MESSAGE,
  NETWORK_ERROR_MESSAGE,
  RECORD_OPEN_ERROR_MESSAGE,
  REPORT_GENERATE_ERROR_MESSAGE,
  isNetworkLikeError,
} = require("../../utils/ux-state");
const {
  MODULE_QIMEN_LABEL,
  buildQimenView,
  buildQimenLongPosterData,
  isQimenRecord,
} = require("../../utils/qimen");

function mapLoadError(error) {
  if (!error) return CONTENT_LOAD_ERROR_MESSAGE;
  if (isNetworkLikeError(error)) return NETWORK_ERROR_MESSAGE;
  if (isBusinessError(error, 40001)) {
    return "会话已失效，请重新进入页面。";
  }
  if (isBusinessError(error, 40401)) {
    return RECORD_OPEN_ERROR_MESSAGE;
  }
  return CONTENT_LOAD_ERROR_MESSAGE;
}

function mapDeleteError(error) {
  if (!error) return "删除失败，请稍后再试。";
  if (isNetworkLikeError(error)) return NETWORK_ERROR_MESSAGE;
  if (isBusinessError(error, 40001)) {
    return "会话已失效，请重新进入页面。";
  }
  if (isBusinessError(error, 40401)) {
    return "记录已不存在，请刷新后查看。";
  }
  return "删除失败，请稍后再试。";
}

function mapUnlockError(error) {
  if (!error) return REPORT_GENERATE_ERROR_MESSAGE;
  if (isNetworkLikeError(error)) return NETWORK_ERROR_MESSAGE;
  if (isBusinessError(error, 40001)) {
    return "会话已失效，请重新进入页面。";
  }
  if (isBusinessError(error, 40301)) {
    return "当前记录暂不支持生成完整报告。";
  }
  if (isBusinessError(error, 40401)) {
    return RECORD_OPEN_ERROR_MESSAGE;
  }
  return REPORT_GENERATE_ERROR_MESSAGE;
}

Page({
  data: {
    loading: true,
    error: "",
    recordId: "",
    createdAtDisplay: "",
    moduleLabel: MODULE_QIMEN_LABEL,
    privacyNote: "问事主题已用于本次局势梳理",
    view: null,
    isUnlocked: false,
    fullContent: "",
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
    this.loadResult();
  },

  onUnload() {
    this.pageUnloaded = true;
    this.unlockFlowToken += 1;
    this.unlockFlowRunning = false;
  },

  safeSetData(data) {
    if (this.pageUnloaded) return;
    this.setData(data);
  },

  isFlowActive(flowToken) {
    return !this.pageUnloaded && flowToken === this.unlockFlowToken;
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

  applyFullContent(content) {
    if (this.pageUnloaded) return;
    const fullContent = String(content || "").trim();
    if (!fullContent) return;
    this.safeSetData({
      fullContent,
      isUnlocked: true,
      fullRevealActive: true,
      unlockError: "",
      cardData: buildQimenLongPosterData(
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
      if (!isQimenRecord(record)) {
        this.setData({
          loading: false,
          error: "该记录不是奇门问事结果，无法在此页展示。",
          view: null,
          isUnlocked: false,
          fullContent: "",
          resultReady: false,
          fullRevealActive: false,
          cardData: null,
        });
        return;
      }

      const view = buildQimenView(record);
      const isUnlocked = Number(record.unlock_status) === 1;
      const fullContent = isUnlocked ? String(record.full_content || "").trim() : "";

      this.setData({
        loading: false,
        view,
        createdAtDisplay: formatDateTime(record.created_at) || "—",
        isUnlocked,
        fullContent,
        fullRevealActive: isUnlocked && !!fullContent,
        resultReady: true,
        cardData: isUnlocked
          ? buildQimenLongPosterData(this.data.recordId, view, fullContent)
          : null,
      });
    } catch (error) {
      this.setData({
        loading: false,
        error: mapLoadError(error),
        view: null,
        isUnlocked: false,
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
      this.data.deleting
    ) {
      return;
    }

    this.setData({ unlocking: true, unlockError: "" });
    try {
      const unlockResult = await unlockAnalysis(this.data.recordId, {
        unlockType: "free_unlock",
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
        unlockType: "free_unlock",
      });
      if (!this.isFlowActive(flowToken)) return;

      const content = unlockResult?.full_content;
      if (!content) {
        throw new Error("完整报告暂未返回，请稍后重新加载。");
      }

      this.applyFullContent(content);
      wx.showToast({ title: "完整报告已生成", icon: "success" });
    } catch (error) {
      if (!this.isFlowActive(flowToken)) return;
      this.safeSetData({
        unlockError: mapUnlockError(error),
      });
    } finally {
      this.endUnlockFlow(flowToken);
    }
  },

  async handleUnlock() {
    if (!this.data.recordId || this.data.isUnlocked) return;
    if (this.data.deleting) return;

    const flowToken = this.beginUnlockFlow();
    if (flowToken === null) return;

    await this.performUnlock(flowToken);
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
    if (this.data.deleting || this.data.unlockFlowRunning || this.data.cardGenerating || !this.data.recordId) return;

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
        wx.redirectTo({ url: "/pages/qimen/qimen" });
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
      buildQimenLongPosterData(
        this.data.recordId,
        this.data.view,
        this.data.fullContent
      ) || this.data.cardData;
    if (!cardData?.id) {
      wx.showToast({ title: "长图数据暂不可用，请刷新后重试", icon: "none" });
      return;
    }

    const card = this.selectComponent("#qimenShareCard");
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
    const { view, error, loading, recordId } = this.data;
    if (loading || error || !view || !recordId) {
      return {
        title: "文易传统文化",
        path: "/pages/qimen/qimen",
      };
    }
    return {
      title: "一份传统文化视角的奇门问事简析",
      path: `/pages/qimen-result/qimen-result?id=${recordId}`,
    };
  },
});
