const POSTER_WIDTH = 600;
const POSTER_HEIGHT = 800;

function normalizeText(value) {
  return String(value || "").replace(/\s+/g, " ").trim();
}

function hexagramName(hexagram, fallback) {
  return normalizeText(hexagram?.full_name || hexagram?.name || fallback);
}

function fitLine(ctx, text, maxWidth, suffix = "…") {
  let output = normalizeText(text);
  while (output && ctx.measureText(`${output}${suffix}`).width > maxWidth) {
    output = output.slice(0, -1);
  }
  return `${output}${suffix}`;
}

function wrapLines(ctx, text, maxWidth, maxLines) {
  const normalized = normalizeText(text);
  if (!normalized) return [];

  const lines = [];
  let current = "";
  for (const character of normalized) {
    const candidate = `${current}${character}`;
    if (current && ctx.measureText(candidate).width > maxWidth) {
      lines.push(current);
      current = character;
    } else {
      current = candidate;
    }
  }
  if (current) lines.push(current);

  if (lines.length <= maxLines) return lines;
  const visible = lines.slice(0, maxLines);
  visible[maxLines - 1] = fitLine(ctx, visible[maxLines - 1], maxWidth);
  return visible;
}

function drawWrappedText(ctx, text, options) {
  const {
    x,
    y,
    maxWidth,
    lineHeight,
    maxLines,
    color = "#44403c",
    font = "16px sans-serif",
  } = options;
  ctx.fillStyle = color;
  ctx.font = font;
  const lines = wrapLines(ctx, text, maxWidth, maxLines);
  lines.forEach((line, index) => {
    ctx.fillText(line, x, y + index * lineHeight);
  });
  return y + lines.length * lineHeight;
}

function drawRoundedRect(ctx, x, y, width, height, radius, fillStyle) {
  const safeRadius = Math.min(radius, width / 2, height / 2);
  ctx.beginPath();
  ctx.moveTo(x + safeRadius, y);
  ctx.lineTo(x + width - safeRadius, y);
  ctx.quadraticCurveTo(x + width, y, x + width, y + safeRadius);
  ctx.lineTo(x + width, y + height - safeRadius);
  ctx.quadraticCurveTo(x + width, y + height, x + width - safeRadius, y + height);
  ctx.lineTo(x + safeRadius, y + height);
  ctx.quadraticCurveTo(x, y + height, x, y + height - safeRadius);
  ctx.lineTo(x, y + safeRadius);
  ctx.quadraticCurveTo(x, y, x + safeRadius, y);
  ctx.closePath();
  ctx.fillStyle = fillStyle;
  ctx.fill();
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
  },

  lifetimes: {
    detached() {
      this.isDetached = true;
      this.canvasNode = null;
    },
  },

  methods: {
    async open() {
      if (this.data.generating) return false;
      if (!this.properties.posterData?.id) {
        wx.showToast({ title: "结果尚未加载完成", icon: "none" });
        return false;
      }

      this.setData({ generating: true });
      try {
        const imagePath = await this.generatePoster();
        if (this.isDetached) return false;
        this.setData({ imagePath, previewVisible: true });
        return true;
      } catch (_error) {
        if (!this.isDetached) {
          wx.showToast({ title: "海报生成失败，请稍后重试", icon: "none" });
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
            if (!canvasInfo?.node || !canvasInfo.width || !canvasInfo.height) {
              reject(new Error("Canvas 初始化失败"));
              return;
            }
            resolve(canvasInfo);
          });
      });
    },

    async generatePoster() {
      const { node: canvas, width, height } = await this.getCanvasNode();
      const pixelRatio = Math.min(wx.getSystemInfoSync().pixelRatio || 2, 3);
      canvas.width = width * pixelRatio;
      canvas.height = height * pixelRatio;
      const ctx = canvas.getContext("2d");
      ctx.scale(pixelRatio, pixelRatio);
      this.canvasNode = canvas;
      this.drawPoster(ctx, this.properties.posterData, width, height);

      await new Promise((resolve) => setTimeout(resolve, 60));
      return new Promise((resolve, reject) => {
        wx.canvasToTempFilePath(
          {
            canvas,
            x: 0,
            y: 0,
            width,
            height,
            destWidth: width * pixelRatio,
            destHeight: height * pixelRatio,
            fileType: "png",
            quality: 1,
            success: (result) => resolve(result.tempFilePath),
            fail: reject,
          },
          this
        );
      });
    },

    drawPoster(ctx, data, canvasWidth, canvasHeight) {
      const scaleX = canvasWidth / POSTER_WIDTH;
      const scaleY = canvasHeight / POSTER_HEIGHT;
      ctx.save();
      ctx.scale(scaleX, scaleY);

      ctx.fillStyle = "#f4efe5";
      ctx.fillRect(0, 0, POSTER_WIDTH, POSTER_HEIGHT);
      drawRoundedRect(ctx, 24, 24, 552, 752, 22, "#fffdf8");

      ctx.fillStyle = "#292524";
      ctx.font = "bold 30px sans-serif";
      ctx.fillText("文易传统文化", 48, 72);
      ctx.fillStyle = "#92400e";
      ctx.font = "14px sans-serif";
      ctx.fillText("传统文化学习 · 趣味解读 · 自我反思", 48, 100);

      ctx.strokeStyle = "#e7e5e4";
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.moveTo(48, 122);
      ctx.lineTo(552, 122);
      ctx.stroke();

      ctx.fillStyle = "#a8a29e";
      ctx.font = "13px sans-serif";
      ctx.fillText(`事项类型 · ${normalizeText(data.categoryName || "未分类")}`, 48, 151);
      drawWrappedText(ctx, data.question || "一次卦象记录", {
        x: 48,
        y: 184,
        maxWidth: 504,
        lineHeight: 28,
        maxLines: 3,
        color: "#292524",
        font: "bold 21px sans-serif",
      });

      drawRoundedRect(ctx, 48, 267, 238, 158, 16, "#faf8f3");
      drawRoundedRect(ctx, 314, 267, 238, 158, 16, "#faf8f3");
      ctx.fillStyle = "#a8a29e";
      ctx.font = "13px sans-serif";
      ctx.fillText("本卦", 68, 295);
      ctx.fillText("变卦", 334, 295);
      ctx.fillStyle = "#292524";
      ctx.font = "bold 20px sans-serif";
      ctx.fillText(hexagramName(data.primaryHexagram, "本卦"), 68, 326, 126);
      ctx.fillText(hexagramName(data.changedHexagram, "变卦"), 334, 326, 126);
      drawHexagramLines(ctx, data.lines, 192, 300, false);
      drawHexagramLines(ctx, data.lines, 458, 300, true);

      ctx.fillStyle = "#92400e";
      ctx.font = "14px sans-serif";
      ctx.fillText(`动爻 · ${normalizeText(data.movingLinesDisplay || "无动爻")}`, 48, 458);

      ctx.fillStyle = "#292524";
      ctx.font = "bold 17px sans-serif";
      ctx.fillText("解读摘要", 48, 497);
      drawWrappedText(ctx, data.freeSummary || "可进入小程序查看本次趣味解读。", {
        x: 48,
        y: 526,
        maxWidth: 504,
        lineHeight: 24,
        maxLines: 5,
        color: "#57534e",
        font: "15px sans-serif",
      });

      ctx.strokeStyle = "#e7e5e4";
      ctx.beginPath();
      ctx.moveTo(48, 660);
      ctx.lineTo(552, 660);
      ctx.stroke();
      drawWrappedText(
        ctx,
        "内容仅用于传统文化学习和自我反思，不作为现实决策依据。",
        {
          x: 48,
          y: 687,
          maxWidth: 504,
          lineHeight: 18,
          maxLines: 2,
          color: "#78716c",
          font: "12px sans-serif",
        }
      );

      drawRoundedRect(ctx, 48, 728, 504, 28, 14, "#f5f5f4");
      ctx.fillStyle = "#57534e";
      ctx.font = "13px sans-serif";
      ctx.textAlign = "center";
      ctx.fillText("微信搜索：文易传统文化", 300, 747);
      ctx.textAlign = "left";
      ctx.restore();
    },

    closePreview() {
      this.setData({ previewVisible: false });
      this.triggerEvent("close");
    },

    preventTouchMove() {},

    getSetting() {
      return new Promise((resolve, reject) => {
        wx.getSetting({ success: resolve, fail: reject });
      });
    },

    authorizeAlbum() {
      return new Promise((resolve, reject) => {
        wx.authorize({ scope: "scope.writePhotosAlbum", success: resolve, fail: reject });
      });
    },

    async guideToSettings() {
      const modal = await new Promise((resolve) => {
        wx.showModal({
          title: "需要相册权限",
          content: "保存海报需要相册权限。你可以前往设置开启，也可以取消后仅查看预览。",
          confirmText: "打开设置",
          success: resolve,
          fail: () => resolve({ confirm: false }),
        });
      });
      if (!modal.confirm) return false;

      const setting = await new Promise((resolve) => {
        wx.openSetting({ success: resolve, fail: () => resolve({ authSetting: {} }) });
      });
      return setting.authSetting?.["scope.writePhotosAlbum"] === true;
    },

    async ensureAlbumPermission() {
      const setting = await this.getSetting();
      const permission = setting.authSetting?.["scope.writePhotosAlbum"];
      if (permission === true) return true;
      if (permission === false) return this.guideToSettings();

      try {
        await this.authorizeAlbum();
        return true;
      } catch (_error) {
        return this.guideToSettings();
      }
    },

    async savePoster() {
      if (this.data.saving || !this.data.imagePath) return;
      this.setData({ saving: true });
      try {
        const allowed = await this.ensureAlbumPermission();
        if (!allowed) {
          wx.showToast({ title: "未获得相册权限，可稍后在设置中开启", icon: "none" });
          return;
        }
        await new Promise((resolve, reject) => {
          wx.saveImageToPhotosAlbum({
            filePath: this.data.imagePath,
            success: resolve,
            fail: reject,
          });
        });
        wx.showToast({ title: "海报已保存到相册", icon: "success" });
      } catch (_error) {
        wx.showToast({ title: "保存失败，请稍后重试", icon: "none" });
      } finally {
        if (!this.isDetached) this.setData({ saving: false });
      }
    },
  },
});
