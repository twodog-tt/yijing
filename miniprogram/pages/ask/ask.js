const { createDivination, getCategories } = require("../../utils/api");
const { isBusinessError } = require("../../utils/request");
const {
  CONTENT_LOAD_ERROR_MESSAGE,
  NETWORK_ERROR_MESSAGE,
  RECORD_OPEN_ERROR_MESSAGE,
  isNetworkLikeError,
  networkOr,
} = require("../../utils/ux-state");

const DEFAULT_PAGE_COPY = Object.freeze({
  title: "整理你正在思考的问题",
  description: "选择事项类型，写下一个具体问题，用卦象帮助梳理处境与行动方向。",
  placeholder: "例如：我应该如何安排本周学习与休息的节奏？",
});

const RELATIONSHIP_TEMPLATES = Object.freeze([
  "我想梳理这段关系目前的状态",
  "我想了解现在是否适合继续主动沟通",
  "我想看看我们之间的问题主要卡在哪里",
  "我想整理这段关系接下来适合怎么处理",
  "我想知道现在是否应该先保持一点距离",
  "我想观察这段关系里的沟通边界",
]);

const SCENE_COPY = Object.freeze({
  relationship: {
    title: "感情关系观察",
    description: "可以围绕一段关系的状态、沟通、边界和下一步行动提出一个具体问题。",
    placeholder: "例如：我想梳理这段关系目前的状态",
    boundary: "结果仅供传统文化学习参考与自我观察，不用于判断对方真实想法，也不替代你的现实沟通与判断。",
    templates: RELATIONSHIP_TEMPLATES,
  },
});

function resolveSceneCopy(scene) {
  if (!scene) return null;
  return SCENE_COPY[String(scene)] || null;
}

function findRelationshipCategoryIndex(categories) {
  return categories.findIndex((item) => /关系|感情|情感|人际/.test(String(item.name || "")));
}

Page({
  data: {
    categories: [],
    categoryIndex: 0,
    selectedCategoryId: 0,
    selectedCategoryName: "",
    pageTitle: DEFAULT_PAGE_COPY.title,
    pageDescription: DEFAULT_PAGE_COPY.description,
    questionPlaceholder: DEFAULT_PAGE_COPY.placeholder,
    scene: "",
    isRelationshipScene: false,
    sceneBoundary: "",
    questionTemplates: [],
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

  onLoad(options = {}) {
    this.pageUnloaded = false;
    this.flowInProgress = false;
    this.navigationStarted = false;
    const sceneCopy = resolveSceneCopy(options.scene);
    this.activeScene = sceneCopy ? String(options.scene) : "";
    if (sceneCopy) {
      this.setData({
        pageTitle: sceneCopy.title,
        pageDescription: sceneCopy.description,
        questionPlaceholder: sceneCopy.placeholder,
        scene: this.activeScene,
        isRelationshipScene: this.activeScene === "relationship",
        sceneBoundary: sceneCopy.boundary,
        questionTemplates: sceneCopy.templates,
      });
    }
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
      if (!categories.length) throw new Error("暂无可用事项类型，请稍后再试。");
      const nextData = { categories };
      if (this.activeScene === "relationship") {
        const categoryIndex = findRelationshipCategoryIndex(categories);
        const category = categories[categoryIndex];
        if (category) {
          nextData.categoryIndex = categoryIndex;
          nextData.selectedCategoryId = Number(category.id);
          nextData.selectedCategoryName = category.name;
        }
      }
      this.setData(nextData);
    } catch (error) {
      this.setData({
        loadError: networkOr(error, error?.message || CONTENT_LOAD_ERROR_MESSAGE),
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

  onTemplateTap(event) {
    const question = String(event.currentTarget.dataset.template || "");
    if (!question) return;
    this.setData({
      question,
      questionLength: [...question].length,
      fieldError: "",
      submitError: "",
      submitCanRetry: false,
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
      if (!result?.id) throw new Error("起卦结果缺少记录 ID，请稍后再试。");
      this.navigationStarted = false;
      this.setData({
        castingRecord: result,
        castingVisible: true,
      });
      animationStarted = true;
    } catch (error) {
      let message = isNetworkLikeError(error)
        ? NETWORK_ERROR_MESSAGE
        : error?.message || "提交失败，请稍后再试。";
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
          submitError: RECORD_OPEN_ERROR_MESSAGE,
          submitCanRetry: false,
        });
      },
    });
  },
});
