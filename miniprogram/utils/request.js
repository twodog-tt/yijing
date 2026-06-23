const { getApiBaseUrl } = require("./config");

const DEFAULT_TIMEOUT = 20000;
const LONG_REQUEST_TIMEOUT = 90000;

const ERROR_TYPES = Object.freeze({
  CONFIG: "config",
  NETWORK: "network",
  TIMEOUT: "timeout",
  HTTP: "http",
  BUSINESS: "business",
  RESPONSE: "response",
});

const BUSINESS_ERROR_MESSAGES = Object.freeze({
  40002: "这个问题不适合使用卦象方式解读，请换成自我反思或行动选择类问题。",
  40301: "完整解读尚未解锁。",
  42901: "请求过于频繁，请稍后再试。",
});

class RequestError extends Error {
  constructor({ type, message, code = 0, httpStatus = 0, detail = "" }) {
    super(message);
    this.name = "RequestError";
    this.type = type;
    this.code = code;
    this.httpStatus = httpStatus;
    this.detail = detail;
  }
}

let loadingCount = 0;

function startLoading(title) {
  loadingCount += 1;
  if (loadingCount === 1) {
    wx.showLoading({
      title: title || "加载中…",
      mask: true,
    });
  }
}

function finishLoading() {
  if (loadingCount > 0) loadingCount -= 1;
  if (loadingCount === 0) wx.hideLoading();
}

function showErrorToast(error) {
  wx.showToast({
    title: error.message || "请求失败，请稍后重试。",
    icon: "none",
    duration: 2500,
  });
}

function normalizeMethod(method) {
  const normalized = String(method || "GET").toUpperCase();
  const supported = ["GET", "POST", "PUT", "DELETE"];
  if (!supported.includes(normalized)) {
    throw new RequestError({
      type: ERROR_TYPES.CONFIG,
      message: `不支持的请求方法：${normalized}`,
    });
  }
  return normalized;
}

function buildUrl(path) {
  if (typeof path !== "string" || !path.trim()) {
    throw new RequestError({
      type: ERROR_TYPES.CONFIG,
      message: "请求路径不能为空。",
    });
  }
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${getApiBaseUrl()}${normalizedPath}`;
}

function isEnvelope(body) {
  return Boolean(
    body &&
      typeof body === "object" &&
      Object.prototype.hasOwnProperty.call(body, "code") &&
      Object.prototype.hasOwnProperty.call(body, "message")
  );
}

function getBusinessMessage(code, serverMessage) {
  if (code === 40002 && serverMessage) return serverMessage;
  return BUSINESS_ERROR_MESSAGES[code] || serverMessage || "请求失败，请稍后重试。";
}

function getHttpMessage(statusCode) {
  if (statusCode === 404) return "请求的服务不存在。";
  if (statusCode === 401 || statusCode === 403) return "当前请求未获授权。";
  if (statusCode >= 500) return "服务器暂时不可用，请稍后重试。";
  return `请求失败（HTTP ${statusCode}）。`;
}

function parseResponse(response, rawResponse) {
  const statusCode = Number(response.statusCode || 0);
  const body = response.data;

  // 后端业务失败通常同时使用 HTTP 4xx，优先保留业务错误分类。
  if (!rawResponse && isEnvelope(body) && Number(body.code) !== 0) {
    const code = Number(body.code);
    throw new RequestError({
      type: ERROR_TYPES.BUSINESS,
      code,
      httpStatus: statusCode,
      message: getBusinessMessage(code, body.message),
    });
  }

  if (statusCode < 200 || statusCode >= 300) {
    throw new RequestError({
      type: ERROR_TYPES.HTTP,
      httpStatus: statusCode,
      message: getHttpMessage(statusCode),
    });
  }

  if (rawResponse) return body;

  if (!isEnvelope(body)) {
    throw new RequestError({
      type: ERROR_TYPES.RESPONSE,
      httpStatus: statusCode,
      message: "服务器响应格式异常。",
    });
  }

  return body.data;
}

function createNetworkError(wxError) {
  const detail = String(wxError?.errMsg || "");
  const isTimeout = /timeout/i.test(detail);
  return new RequestError({
    type: isTimeout ? ERROR_TYPES.TIMEOUT : ERROR_TYPES.NETWORK,
    message: isTimeout
      ? "请求超时，请检查网络后重试。"
      : "网络连接失败，请检查网络或后端服务。",
    detail,
  });
}

/**
 * @param {Object} options
 * @param {string} options.path 相对于 API Base URL 的路径
 * @param {"GET"|"POST"|"PUT"|"DELETE"} [options.method]
 * @param {Object} [options.data]
 * @param {Object} [options.header]
 * @param {number} [options.timeout]
 * @param {boolean} [options.rawResponse] 为 true 时返回原始响应体（如 /health）
 * @param {boolean} [options.toast] 失败时是否由 request 层提示，默认 false
 * @param {boolean} [options.loading] 是否显示统一 loading，默认 false
 * @param {string} [options.loadingText]
 */
function request(options = {}) {
  let method;
  let url;

  try {
    method = normalizeMethod(options.method);
    url = buildUrl(options.path);
  } catch (error) {
    if (options.toast) showErrorToast(error);
    return Promise.reject(error);
  }

  const showLoading = Boolean(options.loading);
  if (showLoading) startLoading(options.loadingText);

  return new Promise((resolve, reject) => {
    function rejectWith(error) {
      if (options.toast) showErrorToast(error);
      reject(error);
    }

    try {
      wx.request({
        url,
        method,
        data: options.data,
        header: {
          "content-type": "application/json",
          ...(options.header || {}),
        },
        timeout: Number(options.timeout) > 0 ? Number(options.timeout) : DEFAULT_TIMEOUT,
        success(response) {
          try {
            resolve(parseResponse(response, Boolean(options.rawResponse)));
          } catch (error) {
            rejectWith(error);
          }
        },
        fail(wxError) {
          rejectWith(createNetworkError(wxError));
        },
        complete() {
          if (showLoading) finishLoading();
        },
      });
    } catch (error) {
      if (showLoading) finishLoading();
      rejectWith(
        error instanceof RequestError
          ? error
          : new RequestError({
              type: ERROR_TYPES.NETWORK,
              message: "请求初始化失败，请稍后重试。",
            })
      );
    }
  });
}

function isRequestError(error) {
  return error instanceof RequestError;
}

function isBusinessError(error, code) {
  return (
    isRequestError(error) &&
    error.type === ERROR_TYPES.BUSINESS &&
    (code == null || error.code === code)
  );
}

module.exports = {
  BUSINESS_ERROR_MESSAGES,
  DEFAULT_TIMEOUT,
  ERROR_TYPES,
  LONG_REQUEST_TIMEOUT,
  RequestError,
  isBusinessError,
  isRequestError,
  request,
};
