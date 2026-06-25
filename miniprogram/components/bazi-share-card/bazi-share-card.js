const CARD_WIDTH = 600;
const CARD_HEIGHT = 960;

function normalizeText(value) {
  return String(value || "").replace(/\s+/g, " ").trim();
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
  },

  lifetimes: {
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
        wx.showToast({ title: "卡片数据暂不可用，请刷新后重试", icon: "none" });
        return false;
      }

      this.setData({ generating: true });
      try {
        const imagePath = await this.generateCard(cardData);
        if (this.isDetached) return false;
        this.setData({ imagePath, previewVisible: true });
        return true;
      } catch (error) {
        if (!this.isDetached) {
          const message =
            error?.message === "Canvas 初始化失败"
              ? "卡片画布初始化失败，请重新进入页面"
              : "卡片生成失败，请稍后重试";
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
              width: canvasInfo.width || CARD_WIDTH,
              height: canvasInfo.height || CARD_HEIGHT,
            });
          });
      });
    },

    async generateCard(cardData) {
      const { node: canvas, width, height } = await this.getCanvasNode();
      const pixelRatio = Math.min(wx.getSystemInfoSync().pixelRatio || 2, 3);
      canvas.width = width * pixelRatio;
      canvas.height = height * pixelRatio;
      const ctx = canvas.getContext("2d");
      ctx.scale(pixelRatio, pixelRatio);
      this.canvasNode = canvas;
      this.drawCard(ctx, cardData, width, height);

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

    drawCard(ctx, data, canvasWidth, canvasHeight) {
      const scaleX = canvasWidth / CARD_WIDTH;
      const scaleY = canvasHeight / CARD_HEIGHT;
      ctx.save();
      ctx.scale(scaleX, scaleY);

      ctx.fillStyle = "#f4efe5";
      ctx.fillRect(0, 0, CARD_WIDTH, CARD_HEIGHT);
      drawRoundedRect(ctx, 24, 24, 552, 912, 22, "#fffdf8");

      ctx.fillStyle = "#292524";
      ctx.font = "bold 30px sans-serif";
      ctx.fillText("文易传统文化", 48, 72);
      ctx.fillStyle = "#92400e";
      ctx.font = "14px sans-serif";
      ctx.fillText("传统文化学习 · 自我观察 · 行动整理", 48, 100);

      ctx.strokeStyle = "#e7e5e4";
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.moveTo(48, 122);
      ctx.lineTo(552, 122);
      ctx.stroke();

      ctx.fillStyle = "#292524";
      ctx.font = "bold 26px sans-serif";
      ctx.fillText("八字简析", 48, 162);
      drawWrappedText(
        ctx,
        "基于简化干支文化规则，仅供传统文化学习与自我反思。",
        {
          x: 48,
          y: 194,
          maxWidth: 504,
          lineHeight: 24,
          maxLines: 2,
          color: "#57534e",
          font: "14px sans-serif",
        }
      );

      drawRoundedRect(ctx, 48, 248, 504, 132, 16, "#faf8f3");
      ctx.fillStyle = "#292524";
      ctx.font = "bold 16px sans-serif";
      ctx.fillText("简化干支示意", 68, 278);
      ctx.fillStyle = "#57534e";
      ctx.font = "15px sans-serif";
      ctx.fillText(`年柱 · ${normalizeText(data.pillars?.year || "—")}`, 68, 308);
      ctx.fillText(`月柱 · ${normalizeText(data.pillars?.month || "—")}`, 68, 334);
      ctx.fillText(`日柱 · ${normalizeText(data.pillars?.day || "—")}`, 68, 360);
      if (data.hourUnknown) {
        ctx.fillText("时柱 · 时辰未知，本次不生成时柱", 68, 386);
      } else {
        ctx.fillText(`时柱 · ${normalizeText(data.pillars?.hour || "—")}`, 68, 386);
      }

      let cursorY = 408;
      ctx.fillStyle = "#292524";
      ctx.font = "bold 16px sans-serif";
      ctx.fillText("五行倾向", 48, cursorY);
      cursorY += 28;
      drawWrappedText(ctx, data.elementSummary || "木 0 · 火 0 · 土 0 · 金 0 · 水 0", {
        x: 48,
        y: cursorY,
        maxWidth: 504,
        lineHeight: 24,
        maxLines: 2,
        color: "#57534e",
        font: "15px sans-serif",
      });
      cursorY += 52;

      if (data.reflectionFocus) {
        ctx.fillStyle = "#292524";
        ctx.font = "bold 16px sans-serif";
        ctx.fillText("反思焦点", 48, cursorY);
        cursorY += 28;
        cursorY = drawWrappedText(ctx, data.reflectionFocus, {
          x: 48,
          y: cursorY,
          maxWidth: 504,
          lineHeight: 24,
          maxLines: 3,
          color: "#57534e",
          font: "15px sans-serif",
        }) + 12;
      }

      if (Array.isArray(data.actionSuggestions) && data.actionSuggestions.length) {
        ctx.fillStyle = "#292524";
        ctx.font = "bold 16px sans-serif";
        ctx.fillText("行动建议", 48, cursorY);
        cursorY += 28;
        data.actionSuggestions.forEach((item, index) => {
          cursorY = drawWrappedText(ctx, `${index + 1}. ${item}`, {
            x: 48,
            y: cursorY,
            maxWidth: 504,
            lineHeight: 24,
            maxLines: 2,
            color: "#57534e",
            font: "15px sans-serif",
          }) + 8;
        });
      }

      ctx.strokeStyle = "#e7e5e4";
      ctx.beginPath();
      ctx.moveTo(48, 820);
      ctx.lineTo(552, 820);
      ctx.stroke();
      drawWrappedText(
        ctx,
        "不等同于专业八字排盘，不构成现实决策依据。",
        {
          x: 48,
          y: 848,
          maxWidth: 504,
          lineHeight: 20,
          maxLines: 2,
          color: "#78716c",
          font: "12px sans-serif",
        }
      );

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
          content: "保存卡片需要相册权限。你可以前往设置开启，也可以取消后仅查看预览。",
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

    async saveCard() {
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
        wx.showToast({ title: "卡片已保存到相册", icon: "success" });
      } catch (_error) {
        wx.showToast({ title: "保存失败，请稍后重试", icon: "none" });
      } finally {
        if (!this.isDetached) this.setData({ saving: false });
      }
    },
  },
});
