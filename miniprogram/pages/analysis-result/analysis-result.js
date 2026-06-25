const { getAnalysis, deleteAnalysis } = require("../../utils/api");
const {
  buildAnalysisView,
  ELEMENT_LABELS,
  MODULE_BAZI_LABEL,
} = require("../../utils/bazi");
const { formatDateTime } = require("../../utils/date");
const { ERROR_TYPES, isBusinessError } = require("../../utils/request");

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
    deleting: false,
    deleteError: "",
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
    this.loadResult();
  },

  buildElementRows(elements) {
    return ["wood", "fire", "earth", "metal", "water"].map((key) => ({
      key,
      label: ELEMENT_LABELS[key],
      value: elements[key] || 0,
    }));
  },

  async loadResult() {
    this.setData({
      loading: true,
      error: "",
      deleteError: "",
    });

    try {
      const record = await getAnalysis(this.data.recordId);
      const view = buildAnalysisView(record);

      this.setData({
        loading: false,
        view,
        createdAtDisplay: formatDateTime(record.created_at) || "—",
        elementRows: this.buildElementRows(view.elements),
      });
    } catch (error) {
      this.setData({
        loading: false,
        error: mapLoadError(error),
        view: null,
      });
    }
  },

  handleDelete() {
    if (this.data.deleting || !this.data.recordId) return;

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
    if (this.data.deleting) return;

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
});
