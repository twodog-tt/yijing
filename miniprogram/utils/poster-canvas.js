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

module.exports = {
  drawRoundedRect,
  drawWrappedText,
  normalizeText,
  wrapLines,
};
