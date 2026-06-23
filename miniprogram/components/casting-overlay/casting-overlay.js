const {
  CASTING_TIMING,
  prepareCastingLines,
} = require("../../utils/casting");

Component({
  properties: {
    visible: {
      type: Boolean,
      value: false,
    },
    record: {
      type: Object,
      value: null,
    },
    mode: {
      type: String,
      value: "ask",
    },
  },

  data: {
    phase: "prepare",
    phaseTitle: "正在整理你的问题",
    phaseSubtitle: "请稍候，卦象结果已由后端生成",
    progressText: "准备中",
    lineSlots: [],
    activeLine: null,
    primaryName: "",
    changedName: "",
    movingLinesText: "",
    hasMovingLines: false,
  },

  observers: {
    "visible, record": function onCastingInput(visible, record) {
      if (!visible) {
        this.stopAnimation();
        return;
      }
      if (!record || !record.id || this.animationRunning) return;
      this.startAnimation(record);
    },
  },

  lifetimes: {
    detached() {
      this.stopAnimation();
    },
  },

  methods: {
    resetState(record) {
      const lines = prepareCastingLines(record?.lines);
      const movingLines = Array.isArray(record?.moving_lines)
        ? record.moving_lines.map(Number)
        : [];
      this.setData({
        phase: "prepare",
        phaseTitle:
          this.properties.mode === "today"
            ? "正在整理今日状态"
            : "正在整理你的问题",
        phaseSubtitle: "卦象结果已返回，正在按顺序展示",
        progressText: "准备中",
        lineSlots: lines,
        activeLine: null,
        primaryName:
          record?.primary_hexagram?.full_name ||
          record?.primary_hexagram?.name ||
          "本卦",
        changedName:
          record?.changed_hexagram?.full_name ||
          record?.changed_hexagram?.name ||
          "变卦",
        movingLinesText: movingLines.length
          ? `第 ${movingLines.join("、")} 爻`
          : "无动爻",
        hasMovingLines: movingLines.length > 0,
      });
      return lines;
    },

    schedule(delay, token) {
      return new Promise((resolve) => {
        const entry = {
          timer: null,
          resolve,
        };
        entry.timer = setTimeout(() => {
          this.pendingTimers = (this.pendingTimers || []).filter(
            (item) => item !== entry
          );
          resolve(this.runToken === token && this.properties.visible);
        }, delay);
        this.pendingTimers = this.pendingTimers || [];
        this.pendingTimers.push(entry);
      });
    },

    clearTimers() {
      (this.pendingTimers || []).forEach((entry) => {
        clearTimeout(entry.timer);
        entry.resolve(false);
      });
      this.pendingTimers = [];
    },

    stopAnimation() {
      this.runToken = (this.runToken || 0) + 1;
      this.clearTimers();
      this.animationRunning = false;
    },

    async startAnimation(record) {
      this.stopAnimation();
      this.animationRunning = true;
      const token = this.runToken;
      const lines = this.resetState(record);

      if (!(await this.schedule(CASTING_TIMING.prepare, token))) return;

      if (!lines.length) {
        this.setData({
          phase: "interpret",
          phaseTitle: "卦象已生成",
          phaseSubtitle: "正在打开结果页",
          progressText: "结果已就绪",
        });
        if (!(await this.schedule(CASTING_TIMING.interpret, token))) return;
        this.finishAnimation(record);
        return;
      }

      for (let index = 0; index < lines.length; index += 1) {
        const line = { ...lines[index], active: true };
        this.setData({
          phase: "line",
          phaseTitle: `第 ${index + 1} 爻正在生成`,
          phaseSubtitle: `${line.position_label} · 爻值 ${line.value}`,
          progressText: `${index + 1} / ${lines.length}`,
          activeLine: line,
          [`lineSlots[${index}].active`]: true,
        });

        if (!(await this.schedule(CASTING_TIMING.line * 0.62, token))) return;
        this.setData({
          [`lineSlots[${index}].revealed`]: true,
          [`lineSlots[${index}].active`]: false,
          activeLine: { ...line, active: false, revealed: true },
        });
        if (!(await this.schedule(CASTING_TIMING.line * 0.38, token))) return;
      }

      this.setData({
        phase: "primary",
        phaseTitle: "本卦已形成",
        phaseSubtitle: this.data.primaryName,
        progressText: "本卦",
        activeLine: null,
      });
      if (!(await this.schedule(CASTING_TIMING.primary, token))) return;

      this.setData({
        phase: "changed",
        phaseTitle: this.data.hasMovingLines
          ? "动爻已标记，正在形成变卦"
          : "本次卦象无动爻，卦意保持稳定",
        phaseSubtitle: this.data.hasMovingLines
          ? `${this.data.movingLinesText} · ${this.data.changedName}`
          : this.data.primaryName,
        progressText: "变卦",
      });
      if (!(await this.schedule(CASTING_TIMING.changed, token))) return;

      this.setData({
        phase: "interpret",
        phaseTitle: "正在整理解读结构",
        phaseSubtitle: "即将打开卦象结果",
        progressText: "整理结果",
      });
      if (!(await this.schedule(CASTING_TIMING.interpret, token))) return;

      this.finishAnimation(record);
    },

    finishAnimation(record) {
      if (!this.animationRunning) return;
      this.animationRunning = false;
      this.clearTimers();
      this.triggerEvent("finish", {
        recordId: record?.id || this.properties.record?.id || 0,
      });
    },

    handleSkip() {
      const recordId = this.properties.record?.id || 0;
      this.stopAnimation();
      this.triggerEvent("cancel", { recordId });
    },
  },
});
