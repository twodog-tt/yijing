const {
  POSTER_WIDTH,
  CONTENT_WIDTH,
  PADDING_X,
  FOOTER_RESERVE,
  MIN_POSTER_HEIGHT,
  TRUNCATION_NOTICE,
  computePosterDimensions,
  drawDivider,
  drawRoundedRect,
  drawSectionTitle,
  drawTitle,
  drawWrappedParagraph,
  estimateParagraphHeight,
  exportCanvasToTempFile,
  getAlbumPermissionHelpers,
  normalizeText,
  remainingLines,
  resolveExportPixelRatio,
  wrapAllLines,
} = require("../../utils/long-poster-canvas");

const QIMEN_INTRO_NOTE =
  "qimen-simple-v1 简化学习版，仅供传统文化学习与自我反思。不等同于专业奇门排盘，不生成完整九宫盘，不构成现实决策依据。";
const SAFE_QUESTION_SUMMARY = "用户问题已用于本次局势梳理";

function estimateMetaHeight(data) {
  const metaParts = [
    data.categoryLabel ? `分类 · ${data.categoryLabel}` : "",
    data.timeBucketLabel ? `时段 · ${data.timeBucketLabel}` : "",
    `摘要 · ${SAFE_QUESTION_SUMMARY}`,
  ].filter(Boolean);
  return estimateParagraphHeight(metaParts.join("\n"), 22, 30);
}

function estimateRawLongPosterHeight(data) {
  let height = 120;
  height += 80;
  height += estimateParagraphHeight(QIMEN_INTRO_NOTE, 22, 30);
  height += estimateParagraphHeight(data.methodNote, 22, 30);
  height += estimateMetaHeight(data);
  height += 80;
  height += estimateParagraphHeight(data.situationOverview, 24, 30);
  if (Array.isArray(data.riskObservations) && data.riskObservations.length) {
    height += 40;
    data.riskObservations.forEach((item) => {
      height += estimateParagraphHeight(item, 24, 30);
    });
  }
  if (data.actionPacing) {
    height += 40 + estimateParagraphHeight(data.actionPacing, 24, 30);
  }
  if (Array.isArray(data.reflectionQuestions) && data.reflectionQuestions.length) {
    height += 40;
    data.reflectionQuestions.forEach((item) => {
      height += estimateParagraphHeight(item, 24, 30);
    });
  }
  if (Array.isArray(data.actionSuggestions) && data.actionSuggestions.length) {
    height += 40;
    data.actionSuggestions.forEach((item) => {
      height += estimateParagraphHeight(item, 24, 30);
    });
  }
  if (data.freeContent) {
    height += 40 + estimateParagraphHeight(data.freeContent, 24, 28);
  }
  if (data.fullContent) {
    height += 40 + estimateParagraphHeight(data.fullContent, 24, 28);
  }
  height += 80;
  return height;
}

function layoutLongPoster(ctx, data, canvasHeight, options) {
  const { isTruncated = false } = options || {};
  const contentStopY = canvasHeight - FOOTER_RESERVE;

  ctx.fillStyle = "#f4efe5";
  ctx.fillRect(0, 0, POSTER_WIDTH, canvasHeight);
  drawRoundedRect(ctx, 24, 24, 552, canvasHeight - 48, 22, "#fffdf8");

  let y = 72;
  y = drawTitle(ctx, "文易传统文化", PADDING_X, y);
  ctx.fillStyle = "#92400e";
  ctx.font = "14px sans-serif";
  ctx.fillText("传统文化学习 · 自我观察 · 行动整理", PADDING_X, y);
  y += 28;
  y = drawDivider(ctx, PADDING_X, y);
  y = drawTitle(ctx, "奇门问事", PADDING_X, y, { font: "bold 26px sans-serif" });
  y = drawWrappedParagraph(
    ctx,
    QIMEN_INTRO_NOTE,
    PADDING_X,
    y,
    { lineHeight: 22, font: "14px sans-serif", color: "#57534e" }
  );
  y += 12;
  y = drawWrappedParagraph(ctx, normalizeText(data.methodNote), PADDING_X, y, {
    lineHeight: 22,
    font: "14px sans-serif",
    color: "#78716c",
  });
  y += 12;

  const metaParts = [
    data.categoryLabel ? `分类 · ${normalizeText(data.categoryLabel)}` : "",
    data.timeBucketLabel ? `时段 · ${normalizeText(data.timeBucketLabel)}` : "",
    `摘要 · ${SAFE_QUESTION_SUMMARY}`,
  ].filter(Boolean);
  y = drawWrappedParagraph(ctx, metaParts.join("\n"), PADDING_X, y, {
    lineHeight: 22,
    font: "14px sans-serif",
    color: "#57534e",
  });
  y += 16;

  y = drawSectionTitle(ctx, "局势梳理", PADDING_X, y);
  y = drawWrappedParagraph(ctx, data.situationOverview || "—", PADDING_X, y);
  y += 8;

  if (Array.isArray(data.riskObservations) && data.riskObservations.length) {
    y = drawSectionTitle(ctx, "风险观察", PADDING_X, y);
    data.riskObservations.forEach((item, index) => {
      y = drawWrappedParagraph(ctx, `${index + 1}. ${item}`, PADDING_X, y);
      y += 4;
    });
    y += 4;
  }

  if (data.actionPacing) {
    y = drawSectionTitle(ctx, "行动节奏", PADDING_X, y);
    y = drawWrappedParagraph(ctx, data.actionPacing, PADDING_X, y);
    y += 8;
  }

  if (Array.isArray(data.reflectionQuestions) && data.reflectionQuestions.length) {
    y = drawSectionTitle(ctx, "自我反思问题", PADDING_X, y);
    data.reflectionQuestions.forEach((item, index) => {
      y = drawWrappedParagraph(ctx, `${index + 1}. ${item}`, PADDING_X, y);
      y += 4;
    });
    y += 4;
  }

  if (Array.isArray(data.actionSuggestions) && data.actionSuggestions.length) {
    y = drawSectionTitle(ctx, "行动建议", PADDING_X, y);
    data.actionSuggestions.forEach((item, index) => {
      y = drawWrappedParagraph(ctx, `${index + 1}. ${item}`, PADDING_X, y);
      y += 4;
    });
    y += 4;
  }

  if (data.freeContent) {
    y = drawSectionTitle(ctx, "免费解读", PADDING_X, y);
    y = drawWrappedParagraph(ctx, data.freeContent, PADDING_X, y);
    y += 8;
  }

  let contentTruncated = false;
  if (data.fullContent) {
    y = drawSectionTitle(ctx, "完整报告", PADDING_X, y);
    if (isTruncated) {
      const maxLines = remainingLines(contentStopY, y, 28);
      const totalLines = wrapAllLines(ctx, data.fullContent, CONTENT_WIDTH);
      contentTruncated = totalLines.length > maxLines;
      y = drawWrappedParagraph(ctx, data.fullContent, PADDING_X, y, {
        lineHeight: 28,
        maxLines,
      });
    } else {
      y = drawWrappedParagraph(ctx, data.fullContent, PADDING_X, y, { lineHeight: 28 });
    }
    y += 8;
  }

  if (isTruncated && contentTruncated) {
    y = drawWrappedParagraph(ctx, TRUNCATION_NOTICE, PADDING_X, y, {
      lineHeight: 22,
      font: "13px sans-serif",
      color: "#b45309",
    });
    y += 8;
  }

  y = drawDivider(ctx, PADDING_X, y + 8);
  drawWrappedParagraph(
    ctx,
    "内容仅用于传统文化学习、自我反思和行动整理，不构成现实决策依据。",
    PADDING_X,
    y,
    { lineHeight: 20, font: "12px sans-serif", color: "#78716c" }
  );
}

Component({
  properties: {
    cardData: {
      type: Object,
      value: null,
    },
  },

  data: {
    previewVisible: false,
    imagePath: "",
    generating: false,
    saving: false,
    canvasHeight: MIN_POSTER_HEIGHT,
    truncatedHint: "",
  },

  lifetimes: {
    attached() {
      this.albumHelpers = getAlbumPermissionHelpers(this);
    },
    detached() {
      this.isDetached = true;
      this.canvasNode = null;
    },
  },

  methods: {
    async open(cardDataOverride) {
      if (this.data.generating) return false;

      const cardData = cardDataOverride || this.properties.cardData;
      if (!cardData?.id) {
        wx.showToast({ title: "长图数据暂不可用，请刷新后重试", icon: "none" });
        return false;
      }

      this.setData({ generating: true, truncatedHint: "" });
      try {
        const imagePath = await this.generateLongPoster(cardData);
        if (this.isDetached) return false;
        this.setData({ imagePath, previewVisible: true });
        return true;
      } catch (error) {
        if (!this.isDetached) {
          const message =
            error?.message === "Canvas 初始化失败"
              ? "长图画布初始化失败，请重新进入页面"
              : "长图生成失败，请稍后重试";
          wx.showToast({ title: message, icon: "none" });
        }
        return false;
      } finally {
        if (!this.isDetached) this.setData({ generating: false });
      }
    },

    getCanvasNode() {
      return new Promise((resolve, reject) => {
        wx.createSelectorQuery()
          .in(this)
          .select("#qimenShareCardCanvas")
          .fields({ node: true, size: true })
          .exec((result) => {
            const canvasInfo = result?.[0];
            if (!canvasInfo?.node) {
              reject(new Error("Canvas 初始化失败"));
              return;
            }
            resolve({
              node: canvasInfo.node,
              width: canvasInfo.width || POSTER_WIDTH,
              height: canvasInfo.height || this.data.canvasHeight || MIN_POSTER_HEIGHT,
            });
          });
      });
    },

    async generateLongPoster(cardData) {
      const rawHeight = estimateRawLongPosterHeight(cardData);
      const { canvasHeight, isTruncated, truncatedHint } = computePosterDimensions(rawHeight);

      await new Promise((resolve) => {
        this.setData({ canvasHeight, truncatedHint }, resolve);
      });

      const { node: canvas, width } = await this.getCanvasNode();
      const systemRatio = wx.getSystemInfoSync().pixelRatio || 2;
      const pixelRatio = resolveExportPixelRatio(canvasHeight, systemRatio);
      canvas.width = width * pixelRatio;
      canvas.height = canvasHeight * pixelRatio;
      const ctx = canvas.getContext("2d");
      ctx.scale(pixelRatio, pixelRatio);
      this.canvasNode = canvas;
      layoutLongPoster(ctx, cardData, canvasHeight, { isTruncated });

      await new Promise((resolve) => setTimeout(resolve, 60));
      return exportCanvasToTempFile(canvas, width, canvasHeight, this, pixelRatio);
    },

    closePreview() {
      this.setData({ previewVisible: false });
      this.triggerEvent("close");
    },

    preventTouchMove() {},

    async saveCard() {
      if (this.data.saving || !this.data.imagePath) return;
      this.setData({ saving: true });
      try {
        await this.albumHelpers.saveImageToAlbum(this.data.imagePath);
      } catch (_error) {
        wx.showToast({ title: "保存失败，请稍后重试", icon: "none" });
      } finally {
        if (!this.isDetached) this.setData({ saving: false });
      }
    },
  },
});
