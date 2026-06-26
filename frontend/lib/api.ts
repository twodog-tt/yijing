import type {
  AIHealthInfo,
  AIStats,
  AnalysisModule,
  AnalysisRecord,
  AnalysisUnlockResult,
  Category,
  CreateBaziAnalysisPayload,
  CreateDivinationPayload,
  CreateQimenAnalysisPayload,
  DailyFortuneTodayResult,
  Divination,
  FreeInterpretation,
  FullInterpretationResponse,
  FullReport,
  PaginatedAILogs,
  PaginatedAnalysisList,
  PaginatedDivinations,
  SessionResult,
  UnlockResult,
} from "./types";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

const SESSION_HEADER = "X-Session-Key";
const ANALYSIS_UNLOCK_TIMEOUT_MS = 90_000;

function sessionHeaders(sessionKey: string): HeadersInit {
  return { [SESSION_HEADER]: sessionKey };
}

function requireAnalysisId(id: number): number {
  if (!Number.isInteger(id) || id <= 0) {
    throw new ApiError(40000, "记录 ID 无效");
  }
  return id;
}

export class ApiError extends Error {
  code: number;

  constructor(code: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.code = code;
  }
}

export function isSensitiveBlockError(err: unknown): boolean {
  return err instanceof ApiError && err.code === 40002;
}

export function isNotUnlockedError(err: unknown): boolean {
  return err instanceof ApiError && err.code === 40301;
}

export function isDebugDisabledError(err: unknown): boolean {
  return err instanceof ApiError && err.code === 40401;
}

export function isRateLimitError(err: unknown): boolean {
  return err instanceof ApiError && err.code === 42901;
}

async function requestDebug<T>(path: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE}${path}`;
  let res: Response;
  try {
    res = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
    });
  } catch (err) {
    if (err instanceof DOMException && err.name === "AbortError") {
      throw err;
    }
    throw new ApiError(50000, "网络连接失败，请确认后端服务已启动。");
  }

  if (res.status === 404) {
    throw new ApiError(40401, "调试接口未启用，仅本地开发环境可用。");
  }

  let body: { code: number; message: string; data?: T };
  try {
    body = await res.json();
  } catch {
    throw new ApiError(50000, "服务器响应异常，请稍后重试。");
  }

  if (body.code !== 0) {
    throw new ApiError(body.code, body.message || "请求失败");
  }

  return body.data as T;
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE}${path}`;
  let res: Response;
  try {
    res = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
    });
  } catch (err) {
    if (err instanceof DOMException && err.name === "AbortError") {
      throw err;
    }
    throw new ApiError(50000, "网络连接失败，请确认后端服务已启动。");
  }

  let body: { code: number; message: string; data?: T };
  try {
    body = await res.json();
  } catch {
    throw new ApiError(50000, "服务器响应异常，请稍后重试。");
  }

  if (body.code !== 0) {
    throw new ApiError(body.code, body.message || "请求失败");
  }

  return body.data as T;
}

export function createSession(sessionKey: string): Promise<SessionResult> {
  return request<SessionResult>("/sessions", {
    method: "POST",
    body: JSON.stringify({ session_key: sessionKey }),
  });
}

export function ensureSession(sessionKey: string): Promise<SessionResult> {
  return createSession(sessionKey);
}

export function getCategories(): Promise<Category[]> {
  return request<Category[]>("/categories");
}

export function createDivination(
  payload: CreateDivinationPayload
): Promise<Divination> {
  return request<Divination>("/divinations", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

const SHANGHAI_TIME_ZONE = "Asia/Shanghai";

export function getLocalDateString(): string {
  return new Intl.DateTimeFormat("en-CA", {
    timeZone: SHANGHAI_TIME_ZONE,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());
}

export function getDailyFortuneToday(
  sessionKey: string,
  localDate = getLocalDateString()
): Promise<DailyFortuneTodayResult> {
  return request<DailyFortuneTodayResult>("/daily-fortune/today", {
    method: "POST",
    body: JSON.stringify({ session_key: sessionKey, local_date: localDate }),
  });
}

export function getDivination(id: number): Promise<Divination> {
  return request<Divination>(`/divinations/${id}`);
}

export function getFreeInterpretation(
  id: number
): Promise<FreeInterpretation> {
  return request<FreeInterpretation>(`/divinations/${id}/interpretation/free`);
}

export function unlockDivination(
  id: number,
  sessionKey: string,
  unlockType: "mock_ad" | "mock_button" = "mock_ad"
): Promise<UnlockResult> {
  return request<UnlockResult>(`/divinations/${id}/unlock`, {
    method: "POST",
    body: JSON.stringify({ session_key: sessionKey, unlock_type: unlockType }),
  });
}

export function getFullInterpretation(
  id: number,
  sessionKey: string
): Promise<FullInterpretationResponse> {
  const qs = new URLSearchParams({ session_key: sessionKey });
  return request<FullInterpretationResponse>(
    `/divinations/${id}/interpretation/full?${qs}`
  );
}

export function getDivinationHistory(
  sessionKey: string,
  page = 1,
  pageSize = 20
): Promise<PaginatedDivinations> {
  const qs = new URLSearchParams({
    session_key: sessionKey,
    page: String(page),
    page_size: String(pageSize),
  });
  return request<PaginatedDivinations>(`/divinations?${qs}`);
}

export function parseFullReport(
  content: FullReport | string | null | undefined
): FullReport | null {
  if (!content) return null;
  if (typeof content === "object") return content;
  try {
    return JSON.parse(content) as FullReport;
  } catch {
    return null;
  }
}

export function getAILogs(
  page = 1,
  pageSize = 20
): Promise<PaginatedAILogs> {
  const qs = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  return requestDebug<PaginatedAILogs>(`/debug/ai-logs?${qs}`);
}

export function getAIHealth(): Promise<AIHealthInfo> {
  return requestDebug<AIHealthInfo>("/debug/ai-health");
}

export function getAIStats(): Promise<AIStats> {
  return requestDebug<AIStats>("/debug/ai-stats");
}

async function requestWithSession<T>(
  path: string,
  sessionKey: string,
  options?: RequestInit
): Promise<T> {
  return request<T>(path, {
    ...options,
    headers: {
      ...sessionHeaders(sessionKey),
      ...options?.headers,
    },
  });
}

export function createBaziAnalysis(
  payload: CreateBaziAnalysisPayload
): Promise<AnalysisRecord> {
  return request<AnalysisRecord>("/analysis/bazi", {
    method: "POST",
    body: JSON.stringify(payload),
    headers: sessionHeaders(payload.session_key),
  });
}

export function createQimenAnalysis(
  payload: CreateQimenAnalysisPayload
): Promise<AnalysisRecord> {
  return request<AnalysisRecord>("/analysis/qimen", {
    method: "POST",
    body: JSON.stringify(payload),
    headers: sessionHeaders(payload.session_key),
  });
}

export function getAnalysis(
  id: number,
  sessionKey: string
): Promise<AnalysisRecord> {
  const normalizedId = requireAnalysisId(id);
  return requestWithSession<AnalysisRecord>(
    `/analysis/${normalizedId}`,
    sessionKey
  );
}

export function getAnalysisList(
  sessionKey: string,
  module: AnalysisModule,
  page = 1,
  pageSize = 20
): Promise<PaginatedAnalysisList> {
  const qs = new URLSearchParams({
    module,
    page: String(page),
    page_size: String(pageSize),
  });
  return requestWithSession<PaginatedAnalysisList>(
    `/analysis?${qs}`,
    sessionKey
  );
}

export function deleteAnalysis(
  id: number,
  sessionKey: string
): Promise<void> {
  const normalizedId = requireAnalysisId(id);
  return requestWithSession<void>(
    `/analysis/${normalizedId}`,
    sessionKey,
    { method: "DELETE" }
  );
}

export function unlockAnalysis(
  id: number,
  sessionKey: string,
  unlockType: "free_unlock" | "rewarded_video_mock" = "free_unlock"
): Promise<AnalysisUnlockResult> {
  const normalizedId = requireAnalysisId(id);
  const controller = new AbortController();
  const timer = setTimeout(
    () => controller.abort(),
    ANALYSIS_UNLOCK_TIMEOUT_MS
  );

  return requestWithSession<AnalysisUnlockResult>(
    `/analysis/${normalizedId}/unlock`,
    sessionKey,
    {
      method: "POST",
      body: JSON.stringify({ unlock_type: unlockType }),
      signal: controller.signal,
    }
  )
    .catch((err) => {
      if (err instanceof DOMException && err.name === "AbortError") {
        throw new ApiError(50000, "解锁请求超时，请稍后重试。");
      }
      throw err;
    })
    .finally(() => clearTimeout(timer));
}
