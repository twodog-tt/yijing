const {
  POSTER_WIDTH,
  CONTENT_WIDTH,
  PADDING_X,
  MIN_POSTER_HEIGHT,
  POSTER_DISCLAIMER,
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
  resolveExportPixelRatio,
} = require("../../utils/long-poster-canvas");

const DIVINATION_METHOD_INTRO =
  "卦象解读仅供传统文化学习与自我观察参考，不等同于专业判断，不构成现实决策依据。";

function hexagramName(hexagram, fallback) {
  return normalizeText(hexagram?.full_name || hexagram?.name || fallback);
}

function drawHexagramLines(ctx, lines, x, y, changed = false) {
  const ordered = (Array.isArray(lines) ? lines : [])
    .slice()
    .sort((a, b) => Number(b.position) - Number(a.position));
  const fallback = Array.from({ length: 6 }, (_, index) => ({
    position: 6 - index,
    is_yang: index % 2 === 0,
    is_moving: false,
  }));
  const visibleLines = ordered.length === 6 ? ordered : fallback;

  visibleLines.forEach((line, index) => {
    const originalYang = Boolean(line.is_yang);
    const isYang = changed && line.is_moving ? !originalYang : originalYang;
    const lineY = y + index * 14;
    ctx.fillStyle = "#292524";
    if (isYang) {
      ctx.fillRect(x, lineY, 66, 5);
    } else {
      ctx.fillRect(x, lineY, 27, 5);
      ctx.fillRect(x + 39, lineY, 27, 5);
    }
    if (line.is_moving) {
      ctx.beginPath();
      ctx.arc(x + 78, lineY + 2.5, 3.5, 0, Math.PI * 2);
      ctx.fillStyle = "#b45309";
      ctx.fill();
    }
  });
}

function drawInsightCard(ctx, x, y, title, lines) {
  const filtered = (Array.isArray(lines) ? lines : []).filter(Boolean);
  if (!filtered.length) return y;
  const cardHeight = 28 + filtered.length * 26;
  drawRoundedRect(ctx, x, y, CONTENT_WIDTH, cardHeight, 16, "#faf8f3");
  ctx.fillStyle = "#292524";
  ctx.font = "bold 16px sans-serif";
  ctx.fillText(title, x + 20, y + 30);
  ctx.fillStyle = "#57534e";
  ctx.font = "15px sans-serif";
  filtered.forEach((line, index) => {
    ctx.fillText(line, x + 20, y + 58 + index * 26);
  });
  return y + cardHeight + 16;
}

function estimateRawLongPosterHeight(data) {
  let height = 320;
  height += estimateParagraphHeight(data.methodNote || DIVINATION_METHOD_INTRO, 22, 30);
  height += 178;
  height += 72;
  height += 40 + estimateParagraphHeight(data.situationSummary || "", 24, 28);
  height += 120;
  if (Array.isArray(data.actionPoints) && data.actionPoints.length) {
    height += 40;
    data.actionPoints.forEach((item) => {
      height += estimateParagraphHeight(item, 24, 30);
    });
  }
  if (Array.isArray(data.reflectionQuestions) && data.reflectionQuestions.length) {
    height += 40;
    data.reflectionQuestions.forEach((item) => {
      height += estimateParagraphHeight(item, 24, 30);
    });
  }
  height += 96;
  return height;
}

function layoutLongPoster(ctx, data, canvasHeight) {
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
  y = drawTitle(ctx, normalizeText(data.moduleTitle || "问事起卦"), PADDING_X, y, {
    font: "bold 26px sans-serif",
  });
  y = drawWrappedParagraph(ctx, DIVINATION_METHOD_INTRO, PADDING_X, y, {
    lineHeight: 22,
    font: "14px sans-serif",
    color: "#57534e",
  });
  y += 12;
  y = drawWrappedParagraph(ctx, normalizeText(data.methodNote), PADDING_X, y, {
    lineHeight: 22,
    font: "14px sans-serif",
    color: "#78716c",
  });
  y += 16;

  y = drawSectionTitle(ctx, "卦象概览", PADDING_X, y);
  y = drawWrappedParagraph(
    ctx,
    `事项类型 · ${normalizeText(data.categoryName || "未分类")}`,
    PADDING_X,
    y
  );
  y += 4;
  y = drawWrappedParagraph(ctx, data.questionSummary || "用户问题已用于本次卦象梳理", PADDING_X, y, {
    lineHeight: 22,
    font: "14px sans-serif",
    color: "#78716c",
  });
  y += 12;

  drawRoundedRect(ctx, PADDING_X, y, 238, 158, 16, "#faf8f3");
  drawRoundedRect(ctx, 314, y, 238, 158, 16, "#faf8f3");
  ctx.fillStyle = "#a8a29e";
  ctx.font = "13px sans-serif";
  ctx.fillText("本卦", PADDING_X + 20, y + 28);
  ctx.fillText("变卦", 334, y + 28);
  ctx.fillStyle = "#292524";
  ctx.font = "bold 18px sans-serif";
  ctx.fillText(hexagramName(data.primaryHexagram, "本卦"), PADDING_X + 20, y + 58);
  ctx.fillText(
    data.hasHexagramChange
      ? hexagramName(data.changedHexagram, "变卦")
      : normalizeText(data.changedHexagramLabel || "无明显变卦"),
    334,
    y + 58
  );
  drawHexagramLines(ctx, data.lines, 192, y + 72, false);
  drawHexagramLines(ctx, data.lines, 458, y + 72, true);
  y += 178;

  y = drawWrappedParagraph(ctx, normalizeText(data.movingHint || data.movingLinesDisplay || "无明显动爻"), PADDING_X, y, {
    lineHeight: 22,
    font: "14px sans-serif",
    color: "#57534e",
  });
  y += 12;

  y = drawSectionTitle(ctx, "局势摘要", PADDING_X, y);
  y = drawWrappedParagraph(ctx, normalizeText(data.situationSummary), PADDING_X, y);
  y += 8;

  y = drawInsightCard(ctx, PADDING_X, y, "变化观察", data.changeObservations);

  if (Array.isArray(data.actionPoints) && data.actionPoints.length) {
    y = drawSectionTitle(ctx, "行动提醒", PADDING_X, y);
    data.actionPoints.forEach((item, index) => {
      y = drawWrappedParagraph(ctx, `${index + 1}. ${item}`, PADDING_X, y);
      y += 4;
    });
    y += 4;
  }

  if (Array.isArray(data.reflectionQuestions) && data.reflectionQuestions.length) {
    y = drawSectionTitle(ctx, "自我反思", PADDING_X, y);
    data.reflectionQuestions.forEach((item) => {
      y = drawWrappedParagraph(ctx, `· ${item}`, PADDING_X, y);
      y += 4;
    });
    y += 4;
  }

  y = drawDivider(ctx, PADDING_X, y + 8);
  drawWrappedParagraph(ctx, POSTER_DISCLAIMER, PADDING_X, y, {
    lineHeight: 20,
    font: "12px sans-serif",
    color: "#78716c",
  });
}

Component({
  properties: {
    posterData: {
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
    async open(posterDataOverride) {
      if (this.data.generating) return false;

      const posterData = posterDataOverride || this.properties.posterData;
      if (!posterData?.id) {
        wx.showToast({ title: "长图数据暂不可用，请刷新后重试", icon: "none" });
        return false;
      }

      this.setData({ generating: true, truncatedHint: "" });
      try {
        const imagePath = await this.generateLongPoster(posterData);
        if (this.isDetached) return false;
        this.setData({ imagePath, previewVisible: true });
        return true;
      } catch (_error) {
        if (!this.isDetached) {
          wx.showToast({ title: "长图生成失败，请稍后重试", icon: "none" });
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
          .select("#sharePosterCanvas")
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

    async generateLongPoster(posterData) {
      const rawHeight = estimateRawLongPosterHeight(posterData);
      const { canvasHeight, truncatedHint } = computePosterDimensions(rawHeight);

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
      layoutLongPoster(ctx, posterData, canvasHeight);

      await new Promise((resolve) => setTimeout(resolve, 60));
      return exportCanvasToTempFile(canvas, width, canvasHeight, this, pixelRatio);
    },

    closePreview() {
      this.setData({ previewVisible: false });
      this.triggerEvent("close");
    },

    preventTouchMove() {},

    async savePoster() {
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
