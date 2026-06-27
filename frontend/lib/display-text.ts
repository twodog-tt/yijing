const INTERNAL_TERM_REPLACEMENTS: Array<[RegExp, string]> = [
  [/bazi-simple-v1/gi, "简化学习口径"],
  [/bazi-v2-poc/gi, "节气观察口径"],
  [/qimen-simple-v1/gi, "简化学习口径"],
  [/qimen-v2-poc/gi, "九宫观察口径"],
  [/qimen-v2-professional/gi, "九宫结构观察口径"],
  [/algorithm_version/gi, "算法口径"],
  [/layout_version/gi, "布局口径"],
  [/layout_basis/gi, "布局口径"],
  [/calendar_basis/gi, "历法口径"],
  [/birth_hour_unknown/gi, "时辰未知"],
  [/birth_hour_branch/gi, "出生时辰"],
  [/birth_date/gi, "出生日期"],
  [/pillars_v2_summary/gi, "四柱结构摘要"],
  [/pillars_v2/gi, "四柱结构"],
  [/five_elements_summary/gi, "五行摘要"],
  [/five_elements/gi, "五行结构"],
  [/bazi_profile/gi, "八字观察视角"],
  [/interpretation_lens/gi, "解读视角"],
  [/question_profile/gi, "问事特征"],
  [/qimen_lens/gi, "奇门观察视角"],
  [/focus_palaces_summary/gi, "重点宫位摘要"],
  [/palaces_summary/gi, "宫位摘要"],
  [/input_payload/gi, "内部数据"],
  [/result_payload/gi, "内部数据"],
  [/payload/gi, "内部数据"],
  [/session_key/gi, "匿名会话标识"],
  [/method_note/gi, "方法说明"],
  [new RegExp("free_" + "unlock", "gi"), "查看完整报告"],
  [new RegExp("rewarded_" + "video_" + "mock", "gi"), "查看完整报告"],
  [/DeepSeek/gi, "智能生成"],
  [/\bmock\b/gi, "内测体验"],
  [/\bPOC\b/g, "试行口径"],
  [/\bprofessional\b/gi, "九宫结构观察"],
  [/\blayout\b/gi, "布局"],
  [/\bv2\b/gi, "新版口径"],
  [/\bv1\b/gi, "简化口径"],
];

export function sanitizeInternalTerms(value: unknown): string {
  if (value === null || value === undefined) return "";
  if (typeof value !== "string" && typeof value !== "number") return "";

  let text = String(value);
  if (!text) return "";

  INTERNAL_TERM_REPLACEMENTS.forEach(([pattern, replacement]) => {
    text = text.replace(pattern, replacement);
  });

  return text
    .replace(/\b[a-z]+(?:_[a-z0-9]+)+\s*=\s*[^；，。\s]+/gi, "结构化信息")
    .replace(/([一-龥]{2,})=结构化信息/g, "$1：已记录")
    .replace(/\b[a-z]+(?:_[a-z0-9]+)+\b/gi, "结构字段")
    .replace(/([·：:，,；;、\s])(?:true|false)(?=([·：:，,；;、\s]|$))/gi, "$1—")
    .replace(/\s{3,}/g, "  ")
    .trim();
}

export function sanitizeInternalTermList(list: unknown): string[] {
  if (!Array.isArray(list)) return [];
  return list.map((item) => sanitizeInternalTerms(item)).filter(Boolean);
}
