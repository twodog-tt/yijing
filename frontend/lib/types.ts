export interface ApiResponse<T> {
  code: number;
  message: string;
  data?: T;
}

export interface Category {
  id: number;
  name: string;
  code?: string;
}

export interface Hexagram {
  id: number;
  number: number;
  name: string;
  full_name: string;
  binary_code: string;
  summary?: string;
}

export interface Line {
  position: number;
  value: number;
  is_yang: number;
  is_moving: number;
}

export interface Divination {
  id: number;
  question: string;
  category: Category;
  primary_hexagram: Hexagram;
  changed_hexagram: Hexagram;
  lines: Line[];
  moving_lines: number[];
  unlock_status: number;
  created_at: string;
  free_interpretation?: string;
}

export interface FreeInterpretation {
  divination_id: number;
  free_content: string;
  ai_provider: string;
  generation_status: number;
  generated_at?: string;
}

export interface FullReport {
  summary: string;
  overall: string;
  current_state: string;
  opportunity: string;
  risk: string;
  action_steps: string[];
  emotion_reminder: string;
  reflection_questions: string[];
  disclaimer: string;
}

export interface UnlockResult {
  divination_id: number;
  unlock_status: number;
  mock_transaction_id: string;
  full_interpretation: FullReport | string;
}

export interface FullInterpretationResponse {
  divination_id: number;
  full_content: FullReport | string;
  ai_provider: string;
}

export interface DivinationListItem {
  id: number;
  question: string;
  category: Category;
  primary_hexagram: Hexagram;
  changed_hexagram: Hexagram;
  moving_lines: number[];
  unlock_status: number;
  created_at: string;
}

export interface PaginatedDivinations {
  items: DivinationListItem[];
  page: number;
  page_size: number;
  total: number;
}

export interface SessionResult {
  session_id: number;
  session_key: string;
}

export interface CreateDivinationPayload {
  session_key: string;
  category_id: number;
  question: string;
  confirm_disclaimer: boolean;
}

export interface DailyFortuneMeta {
  fortune_date: string;
  is_existing: boolean;
}

export interface DailyFortuneTodayResult {
  daily_fortune: DailyFortuneMeta;
  divination: Divination;
}

export interface AIGenerationLog {
  id: number;
  divination_id: number;
  ai_provider: string;
  model_name: string;
  status: number;
  duration_ms: number;
  fallback_used: number;
  error_message?: string;
  created_at: string;
}

export interface PaginatedAILogs {
  items: AIGenerationLog[];
  page: number;
  page_size: number;
  total: number;
}

export interface AIHealthInfo {
  provider: string;
  api_key_configured?: boolean;
  model: string;
  base_url?: string;
  timeout_seconds?: number;
}

export interface AIStats {
  total_count: number;
  success_count: number;
  fail_count: number;
  fallback_count: number;
  avg_duration_ms: number;
  latest_created_at?: string;
}

export const ModuleTypeBazi = 1;
export const ModuleTypeQimen = 2;

export type AnalysisModule = "bazi" | "qimen";

export interface AnalysisRecord {
  id: number;
  module_type: number;
  algorithm_version: string;
  category_id?: number | null;
  question?: string | null;
  input_payload?: unknown;
  result_payload?: unknown;
  free_content?: string | null;
  full_content?: string | null;
  unlock_status: number;
  unlock_type?: string | null;
  ai_provider?: string | null;
  generation_status: number;
  created_at: string;
  updated_at: string;
}

export interface AnalysisListItem {
  id: number;
  module_type: number;
  algorithm_version: string;
  category_id?: number | null;
  question?: string | null;
  unlock_status: number;
  generation_status: number;
  created_at: string;
}

export interface PaginatedAnalysisList {
  items: AnalysisListItem[];
  page: number;
  page_size: number;
  total: number;
}

export interface CreateBaziAnalysisPayload {
  session_key: string;
  birth_date: string;
  birth_hour_unknown: boolean;
  birth_hour_branch?: string;
  confirm_disclaimer: boolean;
}

export interface CreateQimenAnalysisPayload {
  session_key: string;
  question: string;
  category: string;
  confirm_disclaimer: boolean;
}

export interface AnalysisUnlockResult {
  id: number;
  unlock_status: number;
  unlock_type: string;
  full_content: string;
  generation_status: number;
  ai_provider: string;
}
