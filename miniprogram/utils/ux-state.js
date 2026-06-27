const { ERROR_TYPES } = require("./request");

const NETWORK_ERROR_MESSAGE = "网络似乎不太稳定，请稍后再试";
const CONTENT_LOAD_ERROR_MESSAGE = "内容暂时加载失败，请返回后重试";
const RECORD_OPEN_ERROR_MESSAGE = "记录暂时无法打开，请返回历史记录重新进入";
const REPORT_GENERATE_ERROR_MESSAGE = "报告生成失败，请稍后再试";
const ALBUM_SAVE_ERROR_MESSAGE = "保存失败，请检查相册权限";

function isNetworkLikeError(error) {
  return error?.type === ERROR_TYPES.NETWORK || error?.type === ERROR_TYPES.TIMEOUT;
}

function networkOr(error, fallback) {
  return isNetworkLikeError(error) ? NETWORK_ERROR_MESSAGE : fallback;
}

module.exports = {
  ALBUM_SAVE_ERROR_MESSAGE,
  CONTENT_LOAD_ERROR_MESSAGE,
  NETWORK_ERROR_MESSAGE,
  RECORD_OPEN_ERROR_MESSAGE,
  REPORT_GENERATE_ERROR_MESSAGE,
  isNetworkLikeError,
  networkOr,
};
