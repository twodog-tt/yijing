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

function estimateReportSectionsHeight(report) {
  if (!report) return 0;
  let height = 0;
  const sections = [
    report.summary,
    report.overall,
    report.current_state,
    report.opportunity,
    report.risk,
    report.emotion_reminder,
    report.disclaimer,
  ];
  sections.forEach((text) => {
    if (text) height += 40 + estimateParagraphHeight(text, 24, 28);
  });
  if (Array.isArray(report.action_steps) && report.action_steps.length) {
    height += 40;
    report.action_steps.forEach((item) => {
      height += estimateParagraphHeight(item, 24, 28);
    });
  }
  if (Array.isArray(report.reflection_questions) && report.reflection_questions.length) {
    height += 40;
    report.reflection_questions.forEach((item) => {
      height += estimateParagraphHeight(item, 24, 28);
    });
  }
  return height;
}

function estimateRawLongPosterHeight(data) {
  let height = 120;
  height += 80;
  height += 40;
  height += 220;
  height += 40;
  if (data.freeContent) {
    height += 40 + estimateParagraphHeight(data.freeContent, 24, 28);
  }
  if (data.fullReport) {
    height += estimateReportSectionsHeight(data.fullReport);
  } else if (data.fullFallbackText) {
    height += 40 + estimateParagraphHeight(data.fullFallbackText, 24, 28);
  }
  height += 80;
  return height;
}

function drawLimitedParagraph(ctx, text, x, y, contentStopY, options) {
  const lineHeight = options?.lineHeight || 24;
  const maxWidth = options?.maxWidth || CONTENT_WIDTH;
  const maxLines = remainingLines(contentStopY, y, lineHeight);
  if (maxLines <= 0) {
    return { y, truncated: true, drew: false };
  }
  const totalLines = wrapAllLines(ctx, text, maxWidth);
  const truncated = totalLines.length > maxLines;
  const nextY = drawWrappedParagraph(ctx, text, x, y, {
    ...options,
    lineHeight,
    maxLines,
  });
  return { y: nextY, truncated, drew: true };
}

function layoutLongPoster(ctx, data, canvasHeight, options) {
  const { isTruncated = false } = options || {};
  const contentStopY = canvasHeight - FOOTER_RESERVE;
  let contentTruncated = false;

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
  y = drawTitle(ctx, "卦象解析", PADDING_X, y, { font: "bold 26px sans-serif" });
  y = drawWrappedParagraph(
    ctx,
    "基于传统文化视角的卦象解读，仅供学习与自我反思。不等同于专业判断，不构成现实决策依据。",
    PADDING_X,
    y,
    { lineHeight: 22, font: "14px sans-serif", color: "#57534e" }
  );
  y += 12;

  y = drawSectionTitle(ctx, "问事分类", PADDING_X, y);
  y = drawWrappedParagraph(ctx, normalizeText(data.categoryName || "未分类"), PADDING_X, y);
  y += 4;
  y = drawWrappedParagraph(ctx, data.themeNote || "问事主题已用于本次解析", PADDING_X, y, {
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
  ctx.fillText(hexagramName(data.changedHexagram, "变卦"), 334, y + 58);
  drawHexagramLines(ctx, data.lines, 192, y + 72, false);
  drawHexagramLines(ctx, data.lines, 458, y + 72, true);
  y += 178;

  y = drawSectionTitle(ctx, "动爻", PADDING_X, y);
  y = drawWrappedParagraph(ctx, normalizeText(data.movingLinesDisplay || "无动爻"), PADDING_X, y);
  y += 8;

  if (data.freeContent) {
    y = drawSectionTitle(ctx, "免费解读", PADDING_X, y);
    if (isTruncated) {
      const result = drawLimitedParagraph(ctx, data.freeContent, PADDING_X, y, contentStopY);
      y = result.y + 8;
      contentTruncated = contentTruncated || result.truncated;
    } else {
      y = drawWrappedParagraph(ctx, data.freeContent, PADDING_X, y);
      y += 8;
    }
  }

  const report = data.fullReport;
  if (report) {
    y = drawSectionTitle(ctx, "完整解析", PADDING_X, y);

    const drawSection = (title, text) => {
      if (!text || y >= contentStopY) {
        contentTruncated = contentTruncated || Boolean(text);
        return;
      }
      if (title) y = drawSectionTitle(ctx, title, PADDING_X, y);
      if (isTruncated) {
        const result = drawLimitedParagraph(ctx, text, PADDING_X, y, contentStopY);
        y = result.y + 4;
        contentTruncated = contentTruncated || result.truncated;
      } else {
        y = drawWrappedParagraph(ctx, text, PADDING_X, y);
        y += 4;
      }
    };

    if (report.summary) drawSection(null, `一句话总结：${report.summary}`);
    drawSection("总体判断", report.overall);
    drawSection("当前处境", report.current_state);
    drawSection("机会点", report.opportunity);
    drawSection("风险点", report.risk);

    if (Array.isArray(report.action_steps) && report.action_steps.length) {
      if (y < contentStopY) y = drawSectionTitle(ctx, "行动建议", PADDING_X, y);
      report.action_steps.forEach((item, index) => {
        if (y >= contentStopY) {
          contentTruncated = true;
          return;
        }
        if (isTruncated) {
          const result = drawLimitedParagraph(
            ctx,
            `${index + 1}. ${item}`,
            PADDING_X,
            y,
            contentStopY
          );
          y = result.y + 4;
          contentTruncated = contentTruncated || result.truncated;
        } else {
          y = drawWrappedParagraph(ctx, `${index + 1}. ${item}`, PADDING_X, y);
          y += 4;
        }
      });
      y += 4;
    }

    drawSection("情绪提醒", report.emotion_reminder);

    if (Array.isArray(report.reflection_questions) && report.reflection_questions.length) {
      if (y < contentStopY) y = drawSectionTitle(ctx, "自我反思问题", PADDING_X, y);
      report.reflection_questions.forEach((item) => {
        if (y >= contentStopY) {
          contentTruncated = true;
          return;
        }
        if (isTruncated) {
          const result = drawLimitedParagraph(ctx, `· ${item}`, PADDING_X, y, contentStopY);
          y = result.y + 4;
          contentTruncated = contentTruncated || result.truncated;
        } else {
          y = drawWrappedParagraph(ctx, `· ${item}`, PADDING_X, y);
          y += 4;
        }
      });
      y += 4;
    }

    if (report.disclaimer) {
      drawSection(null, report.disclaimer);
    }
  } else if (data.fullFallbackText) {
    y = drawSectionTitle(ctx, "完整解析", PADDING_X, y);
    if (isTruncated) {
      const result = drawLimitedParagraph(
        ctx,
        data.fullFallbackText,
        PADDING_X,
        y,
        contentStopY
      );
      y = result.y + 8;
      contentTruncated = contentTruncated || result.truncated;
    } else {
      y = drawWrappedParagraph(ctx, data.fullFallbackText, PADDING_X, y);
      y += 8;
    }
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
      layoutLongPoster(ctx, posterData, canvasHeight, { isTruncated });

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
