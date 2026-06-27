const HOME_BRAND = Object.freeze({
  eyebrow: "文易传统文化",
  title: "易经问事、八字简析、奇门问事",
  subtitle: "传统文化学习参考，不构成现实决策依据",
});

const HOME_COMPLIANCE_NOTE =
  "本小程序内容基于传统文化模型生成，仅供学习参考、自我观察与行动节奏整理。结果不构成现实决策依据，请结合实际情况判断。";

const HOME_MODULES = Object.freeze([
  {
    key: "ask",
    title: "问事起卦",
    subtitle: "适合梳理一个具体问题的当前状态与行动提醒",
    tags: ["具体问题", "局势观察", "行动提醒"],
    buttonText: "开始问事",
    url: "/pages/ask/ask",
    accent: "ask",
  },
  {
    key: "bazi",
    title: "八字简析",
    subtitle: "适合从传统文化视角观察个人结构与长期节奏",
    tags: ["自我观察", "五行结构", "长期节奏"],
    buttonText: "查看八字",
    url: "/pages/bazi/bazi",
    accent: "bazi",
  },
  {
    key: "qimen",
    title: "奇门问事",
    subtitle: "适合观察当前局势、资源关系与推进节奏",
    tags: ["局势梳理", "资源关系", "推进节奏"],
    buttonText: "进入奇门",
    url: "/pages/qimen/qimen",
    accent: "qimen",
  },
]);

const HOME_GUIDE_ITEMS = Object.freeze([
  { scenario: "临时问题 / 具体事情", module: "问事起卦" },
  { scenario: "自我结构 / 长期节奏", module: "八字简析" },
  { scenario: "当前局势 / 行动节奏", module: "奇门问事" },
]);

const HOME_BOUNDARY_ITEMS = Object.freeze([
  "不做精准预测",
  "不替代现实决策",
  "不提供投资、医疗、法律等建议",
]);

module.exports = {
  HOME_BOUNDARY_ITEMS,
  HOME_BRAND,
  HOME_COMPLIANCE_NOTE,
  HOME_GUIDE_ITEMS,
  HOME_MODULES,
};
