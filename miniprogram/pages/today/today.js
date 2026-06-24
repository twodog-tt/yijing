const { getTodayFortune } = require("../../utils/api");
const { getChinaTodayDate } = require("../../utils/date");

function displayChinaDate(dateString) {
  const parts = String(dateString).split("-");
  if (parts.length !== 3) return dateString;
  return `${parts[0]}年${Number(parts[1])}月${Number(parts[2])}日`;
}

Page({
  data: {
    todayDate: "",
    todayDateDisplay: "",
    submitting: false,
    error: "",
    castingVisible: false,
    castingRecord: null,
  },

  onLoad() {
    const todayDate = getChinaTodayDate();
    this.setData({
      todayDate,
      todayDateDisplay: displayChinaDate(todayDate),
    });
  },

  onUnload() {
    this.pageUnloaded = true;
    this.flowInProgress = false;
    if (this.navigateTimer) clearTimeout(this.navigateTimer);
  },

  async handleStart() {
    if (this.flowInProgress || this.data.submitting) return;
    this.flowInProgress = true;
    this.setData({ submitting: true, error: "" });

    let flowStarted = false;
    try {
      const result = await getTodayFortune({
        local_date: this.data.todayDate,
      });
      const id = result?.divination?.id;
      if (!id) throw new Error("今日结果缺少记录 ID，请稍后重试。");

      if (result.daily_fortune?.is_existing) {
        wx.showToast({
          title: "今日一卦已生成，将为你打开今日结果",
          icon: "none",
          duration: 1600,
        });
        flowStarted = true;
        this.navigateTimer = setTimeout(() => {
          this.navigateTimer = null;
          this.openResult(id);
        }, 900);
      } else {
        flowStarted = true;
        this.navigationStarted = false;
        this.setData({
          castingRecord: result.divination,
          castingVisible: true,
        });
      }
    } catch (error) {
      this.setData({
        error: error?.message || "获取今日一卦失败，请稍后重试。",
      });
    } finally {
      if (!flowStarted) {
        this.flowInProgress = false;
        this.setData({ submitting: false });
      }
    }
  },

  handleCastingFinish(event) {
    this.openResult(event.detail?.recordId || this.data.castingRecord?.id);
  },

  handleCastingCancel(event) {
    wx.showToast({ title: "已跳过动画，正在打开结果", icon: "none" });
    this.openResult(event.detail?.recordId || this.data.castingRecord?.id);
  },

  openResult(recordId) {
    const id = Number(recordId);
    if (!id || this.navigationStarted || this.pageUnloaded) return;
    this.navigationStarted = true;
    this.setData({ castingVisible: false });
    wx.redirectTo({
      url: `/pages/result/result?id=${id}`,
      fail: () => {
        this.navigationStarted = false;
        this.flowInProgress = false;
        if (this.pageUnloaded) return;
        this.setData({
          submitting: false,
          error: "结果页打开失败，记录已保存，可从历史记录重新进入。",
        });
      },
    });
  },

  onShareAppMessage() {
    return {
      title: "今日一卦：一份传统文化趣味解读",
      path: "/pages/today/today",
    };
  },
});
