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

function estimateRawLongPosterHeight(data) {
  let height = 120;
  height += 80;
  height += estimateParagraphHeight(data.methodNote, 22, 30);
  height += 140;
  height += 40;
  height += estimateParagraphHeight(data.elementSummary, 24, 30);
  if (data.reflectionFocus) {
    height += 40 + estimateParagraphHeight(data.reflectionFocus, 24, 30);
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
  y = drawTitle(ctx, "八字简析", PADDING_X, y, { font: "bold 26px sans-serif" });
  y = drawWrappedParagraph(
    ctx,
    "基于简化干支文化规则，仅供传统文化学习与自我反思。不等同于专业八字排盘，不构成现实决策依据。",
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
  y += 16;

  drawRoundedRect(ctx, PADDING_X, y, CONTENT_WIDTH, 132, 16, "#faf8f3");
  ctx.fillStyle = "#292524";
  ctx.font = "bold 16px sans-serif";
  ctx.fillText("简化干支示意", PADDING_X + 20, y + 30);
  ctx.fillStyle = "#57534e";
  ctx.font = "15px sans-serif";
  ctx.fillText(`年柱 · ${normalizeText(data.pillars?.year || "—")}`, PADDING_X + 20, y + 60);
  ctx.fillText(`月柱 · ${normalizeText(data.pillars?.month || "—")}`, PADDING_X + 20, y + 86);
  ctx.fillText(`日柱 · ${normalizeText(data.pillars?.day || "—")}`, PADDING_X + 20, y + 112);
  if (data.hourUnknown) {
    ctx.fillText("时柱 · 时辰未知，本次不生成时柱", PADDING_X + 20, y + 138);
  } else {
    ctx.fillText(`时柱 · ${normalizeText(data.pillars?.hour || "—")}`, PADDING_X + 20, y + 138);
  }
  y += 156;

  y = drawSectionTitle(ctx, "日主", PADDING_X, y);
  y = drawWrappedParagraph(ctx, normalizeText(data.dayMaster || "—"), PADDING_X, y);
  y += 8;

  y = drawSectionTitle(ctx, "五行倾向", PADDING_X, y);
  y = drawWrappedParagraph(ctx, data.elementSummary || "木 0 · 火 0 · 土 0 · 金 0 · 水 0", PADDING_X, y);
  y += 8;

  if (data.reflectionFocus) {
    y = drawSectionTitle(ctx, "反思焦点", PADDING_X, y);
    y = drawWrappedParagraph(ctx, data.reflectionFocus, PADDING_X, y);
    y += 8;
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
          .select("#baziShareCardCanvas")
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
