const { createBaziAnalysis, getAnalysisList } = require("../../utils/api");
const { getChinaTodayDate } = require("../../utils/date");
const { HOUR_BRANCHES, MODULE_BAZI_LABEL } = require("../../utils/bazi");
const { formatDateTime } = require("../../utils/date");
const { ERROR_TYPES, isBusinessError } = require("../../utils/request");

function mapListError(error) {
  if (!error) return "加载记录失败，请稍后重试。";
  if (error.type === ERROR_TYPES.NETWORK) {
    return "网络异常，请检查网络后重试。";
  }
  if (isBusinessError(error, 40001)) {
    return "会话已失效，请重新进入页面。";
  }
  return error.message || "加载记录失败，请稍后重试。";
}

function mapSubmitError(error) {
  if (!error) return "提交失败，请稍后重试。";
  if (error.type === ERROR_TYPES.NETWORK) {
    return "网络异常，请检查网络后重试。";
  }
  if (isBusinessError(error, 40001)) {
    return "会话已失效，请重新进入页面。";
  }
  if (isBusinessError(error, 40002)) {
    return "请检查出生日期、时辰与免责声明后重试。";
  }
  return error.message || "提交失败，请稍后重试。";
}

Page({
  data: {
    maxDate: getChinaTodayDate(),
    birthDate: "",
    hourBranches: HOUR_BRANCHES,
    hourBranchIndex: -1,
    birthHourUnknown: false,
    confirmDisclaimer: false,
    fieldError: "",
    submitError: "",
    submitCanRetry: false,
    submitting: false,
    listLoading: false,
    listError: "",
    recentList: [],
    moduleLabel: MODULE_BAZI_LABEL,
  },

  onShow() {
    this.loadRecentList();
  },

  onBirthDateChange(event) {
    this.setData({
      birthDate: event.detail.value,
      fieldError: "",
      submitError: "",
    });
  },

  onHourBranchChange(event) {
    this.setData({
      hourBranchIndex: Number(event.detail.value),
      fieldError: "",
      submitError: "",
    });
  },

  onHourUnknownChange(event) {
    const checked = event.detail.value.includes("unknown");
    this.setData({
      birthHourUnknown: checked,
      hourBranchIndex: checked ? -1 : this.data.hourBranchIndex,
      fieldError: "",
      submitError: "",
    });
  },

  onDisclaimerChange(event) {
    this.setData({
      confirmDisclaimer: event.detail.value.includes("confirmed"),
      fieldError: "",
      submitError: "",
    });
  },

  validateForm() {
    if (!this.data.birthDate) {
      return "请选择出生日期。";
    }
    if (!this.data.birthHourUnknown && this.data.hourBranchIndex < 0) {
      return "请选择出生时辰，或勾选「时辰未知」。";
    }
    if (!this.data.confirmDisclaimer) {
      return "请勾选免责声明后再提交。";
    }
    return "";
  },

  async loadRecentList() {
    this.setData({
      listLoading: true,
      listError: "",
    });

    try {
      const result = await getAnalysisList({ page: 1, page_size: 20 });
      const items = Array.isArray(result?.items) ? result.items : [];
      const recentList = items.map((item) => ({
        id: item.id,
        createdAtDisplay: formatDateTime(item.created_at) || "—",
        moduleLabel: MODULE_BAZI_LABEL,
        subtitle: item.algorithm_version || "bazi-simple-v1",
      }));

      this.setData({
        recentList,
        listLoading: false,
      });
    } catch (error) {
      this.setData({
        listLoading: false,
        listError: mapListError(error),
        recentList: [],
      });
    }
  },

  async submitForm() {
    if (this.data.submitting) return;

    const fieldError = this.validateForm();
    if (fieldError) {
      this.setData({
        fieldError,
        submitError: "",
      });
      return;
    }

    const hourBranch =
      this.data.hourBranchIndex >= 0
        ? HOUR_BRANCHES[this.data.hourBranchIndex].value
        : "";

    this.setData({
      submitting: true,
      fieldError: "",
      submitError: "",
      submitCanRetry: false,
    });

    try {
      const record = await createBaziAnalysis({
        birth_date: this.data.birthDate,
        birth_hour_branch: hourBranch,
        birth_hour_unknown: this.data.birthHourUnknown,
      });

      this.setData({ submitting: false });
      wx.navigateTo({
        url: `/pages/analysis-result/analysis-result?id=${record.id}`,
      });
    } catch (error) {
      this.setData({
        submitting: false,
        submitError: mapSubmitError(error),
        submitCanRetry: error?.type === ERROR_TYPES.NETWORK,
      });
    }
  },

  openRecord(event) {
    const id = event.currentTarget.dataset.id;
    if (!id) return;
    wx.navigateTo({
      url: `/pages/analysis-result/analysis-result?id=${id}`,
    });
  },
});
