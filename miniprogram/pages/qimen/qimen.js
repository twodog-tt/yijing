const { createQimenAnalysis, getQimenAnalysisList } = require("../../utils/api");
const { formatDateTime } = require("../../utils/date");
const { ERROR_TYPES, isBusinessError } = require("../../utils/request");
const {
  MAX_QUESTION_LENGTH,
  METHOD_NOTE,
  MIN_QUESTION_LENGTH,
  MODULE_QIMEN_LABEL,
  QIMEN_CATEGORIES,
  QUESTION_SUMMARY,
  listRecordSubtitle,
} = require("../../utils/qimen");

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
    return "这个问题不适合用奇门简化解读，请换成自我反思、局势整理或行动节奏类问题。";
  }
  return error.message || "提交失败，请稍后重试。";
}

Page({
  data: {
    categories: QIMEN_CATEGORIES,
    categoryIndex: 0,
    selectedCategory: QIMEN_CATEGORIES[0].value,
    selectedCategoryLabel: QIMEN_CATEGORIES[0].label,
    question: "",
    questionLength: 0,
    confirmDisclaimer: false,
    methodNote: METHOD_NOTE,
    fieldError: "",
    submitError: "",
    submitCanRetry: false,
    submitting: false,
    listLoading: false,
    listError: "",
    recentList: [],
    moduleLabel: MODULE_QIMEN_LABEL,
    questionSummary: QUESTION_SUMMARY,
    categoryPulse: false,
  },

  onLoad() {
    this.pageUnloaded = false;
  },

  onShow() {
    this.loadRecentList();
  },

  onUnload() {
    this.pageUnloaded = true;
    if (this.categoryPulseTimer) {
      clearTimeout(this.categoryPulseTimer);
      this.categoryPulseTimer = null;
    }
  },

  pulseCategoryField() {
    if (this.categoryPulseTimer) {
      clearTimeout(this.categoryPulseTimer);
    }
    this.setData({ categoryPulse: true });
    this.categoryPulseTimer = setTimeout(() => {
      this.categoryPulseTimer = null;
      if (!this.pageUnloaded) {
        this.setData({ categoryPulse: false });
      }
    }, 480);
  },

  onCategoryChange(event) {
    const categoryIndex = Number(event.detail.value);
    const category = this.data.categories[categoryIndex];
    if (!category) return;
    this.setData({
      categoryIndex,
      selectedCategory: category.value,
      selectedCategoryLabel: category.label,
      fieldError: "",
      submitError: "",
    });
    this.pulseCategoryField();
  },

  onQuestionInput(event) {
    const question = event.detail.value || "";
    this.setData({
      question,
      questionLength: [...question].length,
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
    const question = this.data.question.trim();
    const length = [...question].length;
    if (!question) {
      return "请输入你想整理的问题。";
    }
    if (length < MIN_QUESTION_LENGTH || length > MAX_QUESTION_LENGTH) {
      return `问题长度需在 ${MIN_QUESTION_LENGTH} 到 ${MAX_QUESTION_LENGTH} 字之间。`;
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
      const result = await getQimenAnalysisList({ page: 1, page_size: 20 });
      const items = Array.isArray(result?.items) ? result.items : [];
      const recentList = items.map((item) => ({
        id: item.id,
        createdAtDisplay: formatDateTime(item.created_at) || "—",
        moduleLabel: MODULE_QIMEN_LABEL,
        subtitle: listRecordSubtitle(item),
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

    this.setData({
      submitting: true,
      fieldError: "",
      submitError: "",
      submitCanRetry: false,
    });

    try {
      const record = await createQimenAnalysis({
        question: this.data.question.trim(),
        category: this.data.selectedCategory,
      });

      this.setData({ submitting: false });
      wx.navigateTo({
        url: `/pages/qimen-result/qimen-result?id=${record.id}`,
      });
    } catch (error) {
      this.setData({
        submitting: false,
        submitError: mapSubmitError(error),
        submitCanRetry:
          error?.type === ERROR_TYPES.NETWORK && !isBusinessError(error, 40002),
      });
    }
  },

  openRecord(event) {
    const id = event.currentTarget.dataset.id;
    if (!id) return;
    wx.navigateTo({
      url: `/pages/qimen-result/qimen-result?id=${id}`,
    });
  },
});
