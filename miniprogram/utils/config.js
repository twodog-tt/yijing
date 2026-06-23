// 所有 API 地址必须集中维护在这里，不要散落到页面代码中。
const ENVIRONMENTS = Object.freeze({
  dev: Object.freeze({
    apiBaseUrl: "http://123.57.48.214/api/v1",
  }),
  prod: Object.freeze({
    // 仅预留：ICP 备案、HTTPS 证书和微信 request 合法域名配置完成后才能用于正式版。
    apiBaseUrl: "https://api.wenyiapp.cn/api/v1",
  }),
});

function readEnvVersion() {
  try {
    if (typeof __wxConfig !== "undefined" && __wxConfig.envVersion) {
      return __wxConfig.envVersion;
    }
  } catch (_error) {
    // 部分开发者工具环境没有暴露 __wxConfig，继续使用官方同步接口。
  }

  try {
    if (typeof wx !== "undefined" && wx.getAccountInfoSync) {
      const accountInfo = wx.getAccountInfoSync();
      return accountInfo.miniprogram?.envVersion || "";
    }
  } catch (_error) {
    // touristappid 或基础库不支持时默认走 dev。
  }

  return "";
}

function getCurrentEnvironment() {
  // 开发版 develop、体验版 trial 和无法识别的环境都使用 dev。
  // 只有微信正式版 release 才映射到预留的 prod。
  return readEnvVersion() === "release" ? "prod" : "dev";
}

function getEnvironmentConfig() {
  return ENVIRONMENTS[getCurrentEnvironment()];
}

function getApiBaseUrl() {
  return getEnvironmentConfig().apiBaseUrl;
}

module.exports = {
  ENVIRONMENTS,
  getApiBaseUrl,
  getCurrentEnvironment,
  getEnvironmentConfig,
};
