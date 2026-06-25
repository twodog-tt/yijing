const POSTER_WIDTH = 600;
const CONTENT_WIDTH = 504;
const PADDING_X = 48;
const MIN_POSTER_HEIGHT = 960;
const MAX_CANVAS_HEIGHT = 12000;
const FOOTER_RESERVE = 132;
const TRUNCATION_NOTICE =
  "内容较长，长图仅展示前半部分内容；完整内容请在小程序结果页查看。";

function normalizeText(value) {
  return String(value || "").replace(/\s+/g, " ").trim();
}

function preserveNewlines(text) {
  return String(text || "")
    .replace(/\r\n/g, "\n")
    .split("\n")
    .map((line) => line.replace(/\s+/g, " ").trim())
    .join("\n");
}

function wrapAllLines(ctx, text, maxWidth) {
  const normalized = preserveNewlines(text);
  if (!normalized) return [];

  const lines = [];
  normalized.split("\n").forEach((paragraph) => {
    if (!paragraph) {
      lines.push("");
      return;
    }
    let current = "";
    for (const character of paragraph) {
      const candidate = `${current}${character}`;
      if (current && ctx.measureText(candidate).width > maxWidth) {
        lines.push(current);
        current = character;
      } else {
        current = candidate;
      }
    }
    if (current) lines.push(current);
  });
  return lines;
}

function measureWrappedText(ctx, text, maxWidth, lineHeight, font) {
  ctx.font = font;
  const lines = wrapAllLines(ctx, text, maxWidth);
  return lines.length * lineHeight;
}

function drawWrappedParagraph(ctx, text, x, y, options) {
  const {
    maxWidth = CONTENT_WIDTH,
    lineHeight = 24,
    color = "#57534e",
    font = "15px sans-serif",
    maxLines = 0,
  } = options || {};

  ctx.fillStyle = color;
  ctx.font = font;
  let lines = wrapAllLines(ctx, text, maxWidth);
  if (maxLines > 0 && lines.length > maxLines) {
    lines = lines.slice(0, maxLines);
    const last = lines[maxLines - 1];
    let output = last;
    while (output && ctx.measureText(`${output}…`).width > maxWidth) {
      output = output.slice(0, -1);
    }
    lines[maxLines - 1] = `${output}…`;
  }

  lines.forEach((line, index) => {
    ctx.fillText(line, x, y + index * lineHeight);
  });
  return y + lines.length * lineHeight;
}

function drawTitle(ctx, text, x, y, options) {
  const { font = "bold 30px sans-serif", color = "#292524" } = options || {};
  ctx.fillStyle = color;
  ctx.font = font;
  ctx.fillText(text, x, y);
  return y + 36;
}

function drawSectionTitle(ctx, text, x, y, options) {
  const { font = "bold 16px sans-serif", color = "#292524" } = options || {};
  ctx.fillStyle = color;
  ctx.font = font;
  ctx.fillText(text, x, y);
  return y + 28;
}

function drawDivider(ctx, x, y, width = CONTENT_WIDTH) {
  ctx.strokeStyle = "#e7e5e4";
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.moveTo(x, y);
  ctx.lineTo(x + width, y);
  ctx.stroke();
  return y + 20;
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

function estimateParagraphHeight(text, lineHeight, charsPerLine = 28) {
  const normalized = preserveNewlines(text);
  if (!normalized) return 0;
  let lines = 0;
  normalized.split("\n").forEach((paragraph) => {
    if (!paragraph) {
      lines += 1;
      return;
    }
    lines += Math.max(1, Math.ceil(paragraph.length / charsPerLine));
  });
  return lines * lineHeight;
}

function clampHeight(height) {
  return Math.min(Math.max(height, MIN_POSTER_HEIGHT), MAX_CANVAS_HEIGHT);
}

function computePosterDimensions(rawHeight) {
  const normalizedRaw = Math.max(Number(rawHeight) || 0, MIN_POSTER_HEIGHT);
  const isTruncated = normalizedRaw > MAX_CANVAS_HEIGHT;
  return {
    rawHeight: normalizedRaw,
    canvasHeight: clampHeight(normalizedRaw),
    isTruncated,
    truncatedHint: isTruncated
      ? "内容较长，长图仅展示前半部分；完整内容请在小程序内查看"
      : "",
  };
}

function remainingLines(maxY, currentY, lineHeight) {
  if (lineHeight <= 0) return 0;
  return Math.max(0, Math.floor((maxY - currentY) / lineHeight));
}

function resolveExportPixelRatio(canvasHeight, systemPixelRatio) {
  const ratio = Number(systemPixelRatio) || 2;
  if (canvasHeight >= 10000) return 1;
  if (canvasHeight >= 8000) return Math.min(ratio, 1.5);
  if (canvasHeight >= 5000) return Math.min(ratio, 2);
  return Math.min(ratio, 3);
}

function exportCanvasToTempFile(canvas, width, height, component, pixelRatio) {
  const ratio =
    Number(pixelRatio) ||
    resolveExportPixelRatio(height, wx.getSystemInfoSync().pixelRatio || 2);
  return new Promise((resolve, reject) => {
    wx.canvasToTempFilePath(
      {
        canvas,
        x: 0,
        y: 0,
        width,
        height,
        destWidth: width * ratio,
        destHeight: height * ratio,
        fileType: "png",
        quality: 1,
        success: (result) => resolve(result.tempFilePath),
        fail: reject,
      },
      component
    );
  });
}

function getAlbumPermissionHelpers(component) {
  return {
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
          content: "保存长图需要相册权限。你可以前往设置开启，也可以取消后仅查看预览。",
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
    async saveImageToAlbum(filePath) {
      const allowed = await this.ensureAlbumPermission();
      if (!allowed) {
        wx.showToast({ title: "未获得相册权限，可稍后在设置中开启", icon: "none" });
        return false;
      }
      await new Promise((resolve, reject) => {
        wx.saveImageToPhotosAlbum({ filePath, success: resolve, fail: reject });
      });
      wx.showToast({ title: "长图已保存到相册", icon: "success" });
      return true;
    },
  };
}

module.exports = {
  POSTER_WIDTH,
  CONTENT_WIDTH,
  PADDING_X,
  MIN_POSTER_HEIGHT,
  MAX_CANVAS_HEIGHT,
  FOOTER_RESERVE,
  TRUNCATION_NOTICE,
  normalizeText,
  preserveNewlines,
  wrapAllLines,
  measureWrappedText,
  drawWrappedParagraph,
  drawTitle,
  drawSectionTitle,
  drawDivider,
  drawRoundedRect,
  estimateParagraphHeight,
  clampHeight,
  computePosterDimensions,
  remainingLines,
  resolveExportPixelRatio,
  exportCanvasToTempFile,
  getAlbumPermissionHelpers,
};
