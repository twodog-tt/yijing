const {
  ERROR_TYPES,
  LONG_REQUEST_TIMEOUT,
  RequestError,
  isBusinessError,
  request,
} = require("./request");
const { ensureSession } = require("./session");
const { getChinaTodayDate } = require("./date");

const API_PATHS = Object.freeze({
  health: "/health",
  sessions: "/sessions",
  categories: "/categories",
  divinations: "/divinations",
  dailyFortuneToday: "/daily-fortune/today",
});

function requireId(id) {
  const normalized = Number(id);
  if (!Number.isInteger(normalized) || normalized <= 0) {
    throw new RequestError({
      type: ERROR_TYPES.CONFIG,
      message: "记录 ID 无效。",
    });
  }
  return normalized;
}

function positiveInteger(value, fallback) {
  const normalized = Number(value);
  return Number.isInteger(normalized) && normalized > 0 ? normalized : fallback;
}

function detailPath(id) {
  return `/divinations/${requireId(id)}`;
}

function freePath(id) {
  return `${detailPath(id)}/interpretation/free`;
}

function fullPath(id) {
  return `${detailPath(id)}/interpretation/full`;
}

function unlockPath(id) {
  return `${detailPath(id)}/unlock`;
}

function health() {
  return request({
    path: API_PATHS.health,
    method: "GET",
    rawResponse: true,
    timeout: 10000,
  });
}

function createSession() {
  return ensureSession();
}

function getCategories() {
  return request({
    path: API_PATHS.categories,
    method: "GET",
  });
}

async function createDivination({ category_id, question } = {}) {
  const session = await ensureSession();
  return request({
    path: API_PATHS.divinations,
    method: "POST",
    data: {
      session_key: session.session_key,
      category_id,
      question: String(question || "").trim(),
      confirm_disclaimer: true,
    },
  });
}

function getDivination(id) {
  return request({
    path: detailPath(id),
    method: "GET",
  });
}

async function getDivinationHistory({ page = 1, page_size = 20 } = {}) {
  const session = await ensureSession();
  const query = [
    `session_key=${encodeURIComponent(session.session_key)}`,
    `page=${positiveInteger(page, 1)}`,
    `page_size=${positiveInteger(page_size, 20)}`,
  ].join("&");

  return request({
    path: `${API_PATHS.divinations}?${query}`,
    method: "GET",
  });
}

function getFreeInterpretation(id) {
  return request({
    path: freePath(id),
    method: "GET",
  });
}

async function getFullInterpretation(id) {
  const session = await ensureSession();
  const query = `session_key=${encodeURIComponent(session.session_key)}`;

  try {
    const result = await request({
      path: `${fullPath(id)}?${query}`,
      method: "GET",
    });
    return {
      unlocked: true,
      ...result,
    };
  } catch (error) {
    if (isBusinessError(error, 40301)) {
      return {
        unlocked: false,
        code: 40301,
        message: error.message,
        full_content: null,
      };
    }
    throw error;
  }
}

async function unlockDivination(id) {
  const session = await ensureSession();
  return request({
    path: unlockPath(id),
    method: "POST",
    data: {
      session_key: session.session_key,
      unlock_type: "mock_button",
    },
    timeout: LONG_REQUEST_TIMEOUT,
  });
}

async function getTodayFortune({ local_date } = {}) {
  const session = await ensureSession();
  return request({
    path: API_PATHS.dailyFortuneToday,
    method: "POST",
    data: {
      session_key: session.session_key,
      local_date: local_date || getChinaTodayDate(),
    },
  });
}

module.exports = {
  API_PATHS,
  createDivination,
  createSession,
  getCategories,
  getDivination,
  getDivinationHistory,
  getFreeInterpretation,
  getFullInterpretation,
  getTodayFortune,
  health,
  unlockDivination,
};
