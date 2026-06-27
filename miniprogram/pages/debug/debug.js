const {
  createSession,
  getCategories,
  getTodayFortune,
  health,
} = require("../../utils/api");
const { getCurrentEnvironment } = require("../../utils/config");
const { clearSession, getSessionKey } = require("../../utils/session");

function maskSessionKey(sessionKey) {
  if (!sessionKey) return "未创建";
  if (sessionKey.length < 10) return "已创建（已隐藏）";
  return `${sessionKey.slice(0, 4)}…${sessionKey.slice(-4)}`;
}

function errorMessage(error) {
  return error?.message || "检查失败，请稍后重试。";
}

Page({
  data: {
    isDev: true,
    running: "",
    error: "",
    healthResult: "未检查",
    sessionResult: "未检查",
    categoriesResult: "未检查",
    todayResult: "未检查",
  },

  onLoad() {
    this.setData({
      isDev: getCurrentEnvironment() === "dev",
    });
  },

  async testHealth() {
    this.setData({ running: "health", error: "" });
    try {
      const result = await health();
      this.setData({
        healthResult: `服务：${result.status || "未知"}；数据库：${result.db || "未知"}`,
      });
    } catch (error) {
      this.setData({ error: errorMessage(error) });
    } finally {
      this.setData({ running: "" });
    }
  },

  async testSession() {
    this.setData({ running: "session", error: "" });
    try {
      const result = await createSession();
      this.setData({
        sessionResult: `会话已就绪：${maskSessionKey(result.session_key)}`,
      });
    } catch (error) {
      this.setData({ error: errorMessage(error) });
    } finally {
      this.setData({ running: "" });
    }
  },

  async testCategories() {
    this.setData({ running: "categories", error: "" });
    try {
      const result = await getCategories();
      const names = Array.isArray(result) ? result.map((item) => item.name).join("、") : "";
      this.setData({
        categoriesResult: `${Array.isArray(result) ? result.length : 0} 项${names ? `：${names}` : ""}`,
      });
    } catch (error) {
      this.setData({ error: errorMessage(error) });
    } finally {
      this.setData({ running: "" });
    }
  },

  async testTodayFortune() {
    this.setData({ running: "today", error: "" });
    try {
      const result = await getTodayFortune();
      const metadata = result.daily_fortune || {};
      const divination = result.divination || {};
      this.setData({
        todayResult: `日期：${metadata.fortune_date || "未知"}；记录编号：${divination.id || "未知"}；${metadata.is_existing ? "已有记录" : "本次创建"}`,
      });
    } catch (error) {
      this.setData({ error: errorMessage(error) });
    } finally {
      this.setData({ running: "" });
    }
  },

  clearDebugSession() {
    wx.showModal({
      title: "清除本地调试会话",
      content: "清除后将无法通过当前设备会话继续查看原历史记录。仅建议调试时使用。",
      confirmText: "确认清除",
      success: (result) => {
        if (!result.confirm) return;
        clearSession();
        this.setData({
          sessionResult: `已清除；当前状态：${maskSessionKey(getSessionKey())}`,
          todayResult: "未检查",
        });
      },
    });
  },
});
