const { ERROR_TYPES, RequestError, request } = require("./request");

const SESSION_STORAGE_KEY = "yijing_session_key";
const SESSION_PATH = "/sessions";

let sessionPromise = null;
let readySession = null;

function getSessionKey() {
  const value = wx.getStorageSync(SESSION_STORAGE_KEY);
  return typeof value === "string" ? value.trim() : "";
}

function saveSessionKey(sessionKey) {
  const normalized = String(sessionKey || "").trim();
  if (!normalized) {
    throw new RequestError({
      type: ERROR_TYPES.RESPONSE,
      message: "服务器没有返回有效的 session_key。",
    });
  }
  wx.setStorageSync(SESSION_STORAGE_KEY, normalized);
  return normalized;
}

/**
 * 每次小程序进程只同步一次服务端会话；并发调用共享同一个 Promise。
 * 本地没有 key 时传空字符串，由后端生成 UUID。
 */
function ensureSession() {
  const localSessionKey = getSessionKey();

  if (
    readySession &&
    readySession.session_key &&
    readySession.session_key === localSessionKey
  ) {
    return Promise.resolve(readySession);
  }

  if (sessionPromise) return sessionPromise;

  sessionPromise = request({
    path: SESSION_PATH,
    method: "POST",
    data: {
      session_key: localSessionKey,
    },
  })
    .then((session) => {
      const sessionKey = saveSessionKey(session?.session_key);
      readySession = {
        ...session,
        session_key: sessionKey,
      };
      return readySession;
    })
    .then(
      (session) => {
        sessionPromise = null;
        return session;
      },
      (error) => {
        sessionPromise = null;
        throw error;
      }
    );

  return sessionPromise;
}

/** 仅用于开发调试；正常流程清除后将无法继续访问原匿名会话的历史记录。 */
function clearSession() {
  wx.removeStorageSync(SESSION_STORAGE_KEY);
  readySession = null;
  sessionPromise = null;
}

module.exports = {
  SESSION_STORAGE_KEY,
  clearSession,
  ensureSession,
  getSessionKey,
};
