const { ensureSession } = require("../../utils/session");
const {
  HOME_BOUNDARY_ITEMS,
  HOME_BRAND,
  HOME_COMPLIANCE_NOTE,
  HOME_MODULES,
  HOME_PLANNED_SCENE_ITEMS,
  HOME_SCENE_ITEMS,
  HOME_TOOL_ITEMS,
} = require("../../utils/home");

Page({
  data: {
    brand: HOME_BRAND,
    modules: HOME_MODULES,
    plannedScenes: HOME_PLANNED_SCENE_ITEMS,
    sceneItems: HOME_SCENE_ITEMS,
    toolItems: HOME_TOOL_ITEMS,
    boundaryItems: HOME_BOUNDARY_ITEMS,
    complianceNote: HOME_COMPLIANCE_NOTE,
    sessionStatus: "preparing",
  },

  onLoad() {
    ensureSession().then(
      () => this.setData({ sessionStatus: "ready" }),
      () => this.setData({ sessionStatus: "retry_later" })
    );
  },

  onShareAppMessage() {
    return {
      title: "文易传统文化：易经问事、八字简析、奇门问事",
      path: "/pages/index/index",
    };
  },
});
