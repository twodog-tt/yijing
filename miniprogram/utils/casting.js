const CASTING_TIMING = Object.freeze({
  prepare: 400,
  line: 400,
  primary: 500,
  changed: 500,
  interpret: 400,
});

const POSITION_LABELS = ["", "初爻", "二爻", "三爻", "四爻", "五爻", "上爻"];

const VALUE_META = Object.freeze({
  6: Object.freeze({ typeLabel: "老阴", natureLabel: "阴爻", movingLabel: "动爻" }),
  7: Object.freeze({ typeLabel: "少阳", natureLabel: "阳爻", movingLabel: "静爻" }),
  8: Object.freeze({ typeLabel: "少阴", natureLabel: "阴爻", movingLabel: "静爻" }),
  9: Object.freeze({ typeLabel: "老阳", natureLabel: "阳爻", movingLabel: "动爻" }),
});

function normalizeLine(line) {
  const value = Number(line?.value);
  const position = Number(line?.position);
  const meta = VALUE_META[value] || {
    typeLabel: "爻象",
    natureLabel: Number(line?.is_yang) === 1 ? "阳爻" : "阴爻",
    movingLabel: Number(line?.is_moving) === 1 ? "动爻" : "静爻",
  };

  return {
    ...line,
    value,
    position,
    position_label: POSITION_LABELS[position] || `第${position}爻`,
    is_yang: Number(line?.is_yang) === 1,
    is_moving: Number(line?.is_moving) === 1,
    type_label: meta.typeLabel,
    nature_label: meta.natureLabel,
    moving_label: meta.movingLabel,
    revealed: false,
    active: false,
  };
}

/** 只排序和补充展示字段，不生成、修改或重新计算任何爻。 */
function prepareCastingLines(lines = []) {
  return (Array.isArray(lines) ? lines : [])
    .map(normalizeLine)
    .filter((line) => Number.isInteger(line.position) && line.position > 0)
    .sort((a, b) => a.position - b.position);
}

function getCastingDurationMs(lineCount = 6) {
  const count = Math.max(0, Math.min(Number(lineCount) || 0, 6));
  return (
    CASTING_TIMING.prepare +
    count * CASTING_TIMING.line +
    CASTING_TIMING.primary +
    CASTING_TIMING.changed +
    CASTING_TIMING.interpret
  );
}

module.exports = {
  CASTING_TIMING,
  VALUE_META,
  getCastingDurationMs,
  prepareCastingLines,
};
