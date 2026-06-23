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
