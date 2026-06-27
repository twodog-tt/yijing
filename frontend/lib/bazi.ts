import type { AnalysisRecord } from "./types";
import { sanitizeInternalTermList, sanitizeInternalTerms } from "./display-text";

export const HOUR_BRANCHES = [
  { value: "zi", label: "子时 23:00-00:59" },
  { value: "chou", label: "丑时 01:00-02:59" },
  { value: "yin", label: "寅时 03:00-04:59" },
  { value: "mao", label: "卯时 05:00-06:59" },
  { value: "chen", label: "辰时 07:00-08:59" },
  { value: "si", label: "巳时 09:00-10:59" },
  { value: "wu", label: "午时 11:00-12:59" },
  { value: "wei", label: "未时 13:00-14:59" },
  { value: "shen", label: "申时 15:00-16:59" },
  { value: "you", label: "酉时 17:00-18:59" },
  { value: "xu", label: "戌时 19:00-20:59" },
  { value: "hai", label: "亥时 21:00-22:59" },
] as const;

export const ELEMENT_LABELS: Record<string, string> = {
  wood: "木",
  fire: "火",
  earth: "土",
  metal: "金",
  water: "水",
};

export const MODULE_BAZI_LABEL = "八字简析";

export interface BaziProfileView {
  dayMasterObservation: string;
  seasonTendency: string;
  elementBalanceType: string;
  actionStyle: string;
  reflectionTheme: string;
}

export interface InterpretationLensView {
  strengthHint: string;
  cautionHint: string;
  pacingHint: string;
  relationshipWithElements: string;
}

export interface BaziAnalysisView {
  algorithmVersion: string;
  methodNote: string;
  pillars: {
    year: string;
    month: string;
    day: string;
    hour: string;
  };
  hourUnknown: boolean;
  dayMaster: string;
  elements: {
    wood: number;
    fire: number;
    earth: number;
    metal: number;
    water: number;
  };
  baziProfile: BaziProfileView;
  interpretationLens: InterpretationLensView;
  reflectionFocus: string;
  actionSuggestions: string[];
  freeContent: string;
}

function parseJSONField(raw: unknown): Record<string, unknown> {
  if (!raw) return {};
  if (typeof raw === "object") return raw as Record<string, unknown>;
  if (typeof raw !== "string") return {};
  try {
    return JSON.parse(raw) as Record<string, unknown>;
  } catch {
    return {};
  }
}

function displayTextField(value: unknown, fallback = ""): string {
  const text = sanitizeInternalTerms(value);
  return text || fallback;
}

export function buildAnalysisView(record: AnalysisRecord): BaziAnalysisView {
  const result = parseJSONField(record.result_payload);
  const pillars = (result.pillars as Record<string, string>) || {};
  const elements = (result.five_elements as Record<string, number>) || {};
  const hourUnknown = !pillars.hour;
  const suggestions = Array.isArray(result.action_suggestions)
    ? (result.action_suggestions as string[])
    : [];
  const profileRaw = (result.bazi_profile as Record<string, string>) || {};
  const lensRaw = (result.interpretation_lens as Record<string, string>) || {};

  return {
    algorithmVersion:
      (result.algorithm_version as string) ||
      record.algorithm_version ||
      "bazi-simple-v1",
    methodNote:
      displayTextField(result.method_note) ||
      "本功能采用简化干支文化规则，不等同于专业八字排盘。",
    pillars: {
      year: displayTextField(pillars.year, "—"),
      month: displayTextField(pillars.month, "—"),
      day: displayTextField(pillars.day, "—"),
      hour: displayTextField(pillars.hour),
    },
    hourUnknown,
    dayMaster: displayTextField(result.day_master, "—"),
    elements: {
      wood: Number(elements.wood) || 0,
      fire: Number(elements.fire) || 0,
      earth: Number(elements.earth) || 0,
      metal: Number(elements.metal) || 0,
      water: Number(elements.water) || 0,
    },
    baziProfile: {
      dayMasterObservation: displayTextField(profileRaw.day_master_observation),
      seasonTendency: displayTextField(profileRaw.season_tendency),
      elementBalanceType: displayTextField(profileRaw.element_balance_type),
      actionStyle: displayTextField(profileRaw.action_style),
      reflectionTheme: displayTextField(profileRaw.reflection_theme),
    },
    interpretationLens: {
      strengthHint: displayTextField(lensRaw.strength_hint),
      cautionHint: displayTextField(lensRaw.caution_hint),
      pacingHint: displayTextField(lensRaw.pacing_hint),
      relationshipWithElements: displayTextField(lensRaw.relationship_with_elements),
    },
    reflectionFocus: displayTextField(result.reflection_focus),
    actionSuggestions: sanitizeInternalTermList(suggestions),
    freeContent: displayTextField(record.free_content),
  };
}

export function buildElementRows(
  elements: BaziAnalysisView["elements"]
): { key: string; label: string; value: number }[] {
  return (["wood", "fire", "earth", "metal", "water"] as const).map((key) => ({
    key,
    label: ELEMENT_LABELS[key],
    value: elements[key] || 0,
  }));
}
