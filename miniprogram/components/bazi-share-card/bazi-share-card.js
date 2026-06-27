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
const { ALBUM_SAVE_ERROR_MESSAGE } = require("../../utils/ux-state");

const BAZI_METHOD_INTRO =
  "基于简化干支与五行观察，仅供传统文化学习参考，不等同于专业八字排盘，不构成现实决策依据。";

function estimateRawLongPosterHeight(data) {
  let height = 320;
  height += estimateParagraphHeight(data.methodNote, 22, 30);
  if (data.v2PosterNote) {
    height += estimateParagraphHeight(data.v2PosterNote, 22, 28);
  }
  height += estimateInfoSectionHeight(buildPillarLines(data));
  height += estimateInfoSectionHeight(buildProfileLines(data.baziProfile || {}));
  height += estimateInfoSectionHeight(buildLensLines(data.interpretationLens || {}));
  if (Array.isArray(data.actionPoints) && data.actionPoints.length) {
    height += 40;
    data.actionPoints.forEach((item) => {
      height += estimateParagraphHeight(item, 24, 30);
    });
  }
  height += 96;
  return height;
}

function estimateInfoCardHeight(lines) {
  const filtered = (Array.isArray(lines) ? lines : []).filter(Boolean);
  if (!filtered.length) return 0;
  const bodyHeight = filtered.reduce(
    (total, line) => total + estimateParagraphHeight(line, 24, 30) + 8,
    0
  );
  return 62 + bodyHeight;
}

function estimateInfoSectionHeight(lines) {
  const cardHeight = estimateInfoCardHeight(lines);
  return cardHeight ? cardHeight + 16 : 0;
}

function drawInfoCard(ctx, x, y, title, lines) {
  const filtered = (Array.isArray(lines) ? lines : []).filter(Boolean);
  if (!filtered.length) return y;
  const bodyWidth = CONTENT_WIDTH - 40;
  const cardHeight = estimateInfoCardHeight(filtered);
  drawRoundedRect(ctx, x, y, CONTENT_WIDTH, cardHeight, 16, "#faf8f3");
  ctx.fillStyle = "#292524";
  ctx.font = "bold 16px sans-serif";
  ctx.fillText(title, x + 20, y + 30);

  let cursorY = y + 60;
  filtered.forEach((line) => {
    cursorY = drawWrappedParagraph(ctx, line, x + 20, cursorY, {
      maxWidth: bodyWidth,
      lineHeight: 24,
      font: "15px sans-serif",
      color: "#57534e",
    });
    cursorY += 8;
  });
  return y + cardHeight + 16;
}

function buildPillarLines(data) {
  const pillarLines = [
    `年柱 · ${normalizeText(data.pillars?.year || "—")}`,
    `月柱 · ${normalizeText(data.pillars?.month || "—")}`,
    `日柱 · ${normalizeText(data.pillars?.day || "—")}`,
  ];
  if (!data.hourUnknown) {
    pillarLines.push(`时柱 · ${normalizeText(data.pillars?.hour || "—")}`);
  }
  pillarLines.push(`日主 · ${normalizeText(data.dayMaster || "—")}`);
  return pillarLines;
}

function buildProfileLines(profile) {
  return [
    profile.elementBalanceType ? `五行倾向 · ${profile.elementBalanceType}` : "",
    profile.actionStyle ? `行动风格 · ${profile.actionStyle}` : "",
    profile.reflectionTheme ? `反思主题 · ${profile.reflectionTheme}` : "",
    profile.seasonTendency ? `季节倾向 · ${profile.seasonTendency}` : "",
  ];
}

function buildLensLines(lens) {
  return [
    lens.strengthHint ? `可借助 · ${lens.strengthHint}` : "",
    lens.cautionHint ? `需留意 · ${lens.cautionHint}` : "",
    lens.pacingHint ? `节奏建议 · ${lens.pacingHint}` : "",
  ];
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
  y = drawTitle(ctx, "八字简析", PADDING_X, y, { font: "bold 26px sans-serif" });
  y = drawWrappedParagraph(ctx, BAZI_METHOD_INTRO, PADDING_X, y, {
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
  y += 12;
  if (data.v2PosterNote) {
    y = drawWrappedParagraph(ctx, normalizeText(data.v2PosterNote), PADDING_X, y, {
      lineHeight: 22,
      font: "13px sans-serif",
      color: "#92400e",
    });
    y += 8;
  }
  y += 4;

  y = drawInfoCard(ctx, PADDING_X, y, "四柱示意", buildPillarLines(data));

  const profile = data.baziProfile || {};
  y = drawInfoCard(ctx, PADDING_X, y, "解读视角", buildProfileLines(profile));

  const lens = data.interpretationLens || {};
  y = drawInfoCard(ctx, PADDING_X, y, "节奏与留意", buildLensLines(lens));

  if (Array.isArray(data.actionPoints) && data.actionPoints.length) {
    y = drawSectionTitle(ctx, "行动要点", PADDING_X, y);
    data.actionPoints.forEach((item, index) => {
      y = drawWrappedParagraph(ctx, `${index + 1}. ${item}`, PADDING_X, y);
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
      layoutLongPoster(ctx, cardData, canvasHeight);

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
        wx.showToast({ title: ALBUM_SAVE_ERROR_MESSAGE, icon: "none" });
      } finally {
        if (!this.isDetached) this.setData({ saving: false });
      }
    },
  },
});
