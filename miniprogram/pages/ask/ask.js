const { createDivination, getCategories } = require("../../utils/api");
const { isBusinessError } = require("../../utils/request");

Page({
  data: {
    categories: [],
    categoryIndex: 0,
    selectedCategoryId: 0,
    selectedCategoryName: "",
    question: "",
    questionLength: 0,
    confirmed: false,
    loading: true,
    loadError: "",
    submitError: "",
    submitCanRetry: false,
    fieldError: "",
    submitting: false,
    castingVisible: false,
    castingRecord: null,
  },

  onLoad() {
    this.loadCategories();
  },

  onUnload() {
    this.pageUnloaded = true;
    this.flowInProgress = false;
  },

  async loadCategories() {
    this.setData({ loading: true, loadError: "" });
    try {
      const result = await getCategories();
      const categories = (Array.isArray(result) ? result : []).filter(
        (item) => Number(item.id) !== 6 && item.name !== "今日运势"
      );
      if (!categories.length) throw new Error("暂无可用事项类型，请稍后重试。");
      this.setData({ categories });
    } catch (error) {
      this.setData({
        loadError: error?.message || "事项类型加载失败，请稍后重试。",
      });
    } finally {
      this.setData({ loading: false });
    }
  },

  onCategoryChange(event) {
    const categoryIndex = Number(event.detail.value);
    const category = this.data.categories[categoryIndex];
    if (!category) return;
    this.setData({
      categoryIndex,
      selectedCategoryId: Number(category.id),
      selectedCategoryName: category.name,
      fieldError: "",
    });
  },

  onQuestionInput(event) {
    const question = event.detail.value || "";
    this.setData({
      question,
      questionLength: [...question].length,
      fieldError: "",
    });
  },

  onDisclaimerChange(event) {
    this.setData({
      confirmed: event.detail.value.includes("confirmed"),
      fieldError: "",
    });
  },

  validate() {
    if (!this.data.selectedCategoryId) return "请选择事项类型。";
    const question = this.data.question.trim();
    const length = [...question].length;
    if (!question) return "请输入你想整理的问题。";
    if (length < 5 || length > 200) return "问题长度需在 5 到 200 字之间。";
    if (!this.data.confirmed) return "请先阅读并确认免责声明。";
    return "";
  },

  async submitForm() {
    if (this.flowInProgress || this.data.submitting) return;
    const fieldError = this.validate();
    if (fieldError) {
      this.setData({ fieldError });
      return;
    }

    this.flowInProgress = true;
    this.setData({
      submitting: true,
      fieldError: "",
      submitError: "",
      submitCanRetry: false,
    });

    let animationStarted = false;
    try {
      const result = await createDivination({
        category_id: this.data.selectedCategoryId,
        question: this.data.question.trim(),
      });
      if (!result?.id) throw new Error("起卦结果缺少记录 ID，请稍后重试。");
      this.navigationStarted = false;
      this.setData({
        castingRecord: result,
        castingVisible: true,
      });
      animationStarted = true;
    } catch (error) {
      let message = error?.message || "提交失败，请稍后重试。";
      let submitCanRetry = true;
      if (isBusinessError(error, 40002)) {
        message = "这个问题不适合使用卦象方式解读，请换成自我反思或行动选择类问题。";
        submitCanRetry = false;
      } else if (isBusinessError(error, 42901)) {
        message = "请求过于频繁，请稍后再试。";
        submitCanRetry = false;
      }
      this.setData({ submitError: message, submitCanRetry });
    } finally {
      if (!animationStarted) {
        this.flowInProgress = false;
        this.setData({ submitting: false });
      }
    }
  },

  handleCastingFinish(event) {
    this.openResult(event.detail?.recordId || this.data.castingRecord?.id);
  },

  handleCastingCancel(event) {
    wx.showToast({ title: "已跳过动画，正在打开结果", icon: "none" });
    this.openResult(event.detail?.recordId || this.data.castingRecord?.id);
  },

  openResult(recordId) {
    const id = Number(recordId);
    if (!id || this.navigationStarted || this.pageUnloaded) return;
    this.navigationStarted = true;
    this.setData({ castingVisible: false });
    wx.redirectTo({
      url: `/pages/result/result?id=${id}`,
      fail: () => {
        this.navigationStarted = false;
        this.flowInProgress = false;
        if (this.pageUnloaded) return;
        this.setData({
          submitting: false,
          submitError: "结果页打开失败，记录已保存，可从历史记录重新进入。",
          submitCanRetry: false,
        });
      },
    });
  },
});
