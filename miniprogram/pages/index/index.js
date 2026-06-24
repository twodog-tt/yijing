const { ensureSession } = require("../../utils/session");

Page({
  data: {
    sessionStatus: "preparing",
  },

  onLoad() {
    // 轻量预热匿名会话；失败不阻塞首页，进入业务页时会自动重试。
    ensureSession().then(
      () => this.setData({ sessionStatus: "ready" }),
      () => this.setData({ sessionStatus: "retry_later" })
    );
  },

  onShareAppMessage() {
    return {
      title: "文易传统文化：学习、趣味解读与自我反思",
      path: "/pages/index/index",
    };
  },
});
