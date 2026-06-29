const HOME_BRAND = Object.freeze({
  eyebrow: "传统文化观察工具",
  title: "把问题整理清楚",
  subtitle: "用传统文化模型，帮你观察结构与行动节奏。",
});

const HOME_COMPLIANCE_NOTE =
  "内容仅供传统文化学习与自我观察参考，请结合实际情况判断。";

const HOME_MODULES = Object.freeze([
  {
    key: "ask",
    title: "问事起卦",
    subtitle: "适合梳理一个具体问题的当前状态与行动提醒",
    tags: ["具体问题", "行动提醒"],
    buttonText: "开始问事",
    url: "/pages/ask/ask",
    accent: "ask",
  },
  {
    key: "bazi",
    title: "八字简析",
    subtitle: "适合从传统文化视角观察个人结构与长期节奏",
    tags: ["个人结构", "长期节奏"],
    buttonText: "查看八字",
    url: "/pages/bazi/bazi",
    accent: "bazi",
  },
  {
    key: "qimen",
    title: "奇门问事",
    subtitle: "适合观察当前局势、资源关系与推进节奏",
    tags: ["局势梳理", "推进节奏"],
    buttonText: "进入奇门",
    url: "/pages/qimen/qimen",
    accent: "qimen",
  },
]);

const HOME_SCENE_ITEMS = Object.freeze([
  {
    key: "relationship",
    title: "感情关系观察",
    subtitle: "梳理关系状态、沟通节奏与边界提醒",
    buttonText: "进入",
    url: "/pages/ask/ask?scene=relationship",
  },
]);

const HOME_PLANNED_SCENE_ITEMS = Object.freeze([
  { key: "career", title: "事业选择" },
  { key: "study", title: "学习规划" },
  { key: "communication", title: "人际沟通" },
]);

const HOME_TOOL_ITEMS = Object.freeze([
  { key: "dream", title: "梦境意象解析", hint: "意象整理" },
  { key: "name-stroke", title: "姓名笔画观察", hint: "笔画参考" },
  { key: "naming", title: "起名灵感助手", hint: "灵感收集" },
  { key: "relationship-note", title: "感情签", hint: "关系提醒" },
]);

const HOME_BOUNDARY_ITEMS = Object.freeze([
  "不承诺具体结果",
  "不替代现实决策",
  "不提供投资、医疗、法律等专业建议",
]);

module.exports = {
  HOME_BOUNDARY_ITEMS,
  HOME_BRAND,
  HOME_COMPLIANCE_NOTE,
  HOME_MODULES,
  HOME_PLANNED_SCENE_ITEMS,
  HOME_SCENE_ITEMS,
  HOME_TOOL_ITEMS,
};
