// 所有 API 地址必须集中维护在这里，不要散落到页面代码中。
const ENVIRONMENTS = Object.freeze({
  dev: Object.freeze({
    apiBaseUrl: "http://123.57.48.214/api/v1",
    ad: Object.freeze({
      enabled: true,
      mode: "mock",
      // 未来开通流量主后，才允许配置真实公开 adUnitId；不要写入 AppSecret 或广告后台密钥。
      rewardedVideoAdUnitId: "",
      mockDurationMs: 5000,
      // dev 专用：completed | cancelled，用于测试 mock 激励视频完成/中途退出
      mockOutcome: "completed",
    }),
  }),
  prod: Object.freeze({
    // 仅预留：ICP 备案、HTTPS 证书和微信 request 合法域名配置完成后才能用于正式版。
    apiBaseUrl: "https://api.wenyiapp.cn/api/v1",
    ad: Object.freeze({
      enabled: false,
      mode: "disabled",
      rewardedVideoAdUnitId: "",
      mockDurationMs: 0,
      mockOutcome: "completed",
    }),
  }),
});

function isWechatDevtools() {
  try {
    if (typeof wx !== "undefined" && wx.getSystemInfoSync) {
      return wx.getSystemInfoSync().platform === "devtools";
    }
  } catch (_error) {
    // 部分基础库不支持时按 fail closed 处理。
  }
  return false;
}

function readEnvVersion() {
  try {
    if (typeof __wxConfig !== "undefined" && __wxConfig.envVersion) {
      return String(__wxConfig.envVersion).trim();
    }
  } catch (_error) {
    // 部分开发者工具环境没有暴露 __wxConfig，继续使用官方同步接口。
  }

  try {
    if (typeof wx !== "undefined" && wx.getAccountInfoSync) {
      const accountInfo = wx.getAccountInfoSync();
      return String(accountInfo.miniprogram?.envVersion || "").trim();
    }
  } catch (_error) {
    // touristappid 或基础库不支持时不默认 dev。
  }

  return "";
}

function getCurrentEnvironment() {
  const envVersion = readEnvVersion();

  if (envVersion === "release") {
    return "prod";
  }

  if (envVersion === "develop" || envVersion === "trial") {
    return "dev";
  }

  // 未知 envVersion：仅微信开发者工具允许 dev/mock；真机未知环境 fail closed 到 prod。
  if (isWechatDevtools()) {
    return "dev";
  }

  return "prod";
}

function getEnvironmentConfig() {
  return ENVIRONMENTS[getCurrentEnvironment()];
}

function getApiBaseUrl() {
  return getEnvironmentConfig().apiBaseUrl;
}

function getAdConfig() {
  return getEnvironmentConfig().ad;
}

module.exports = {
  ENVIRONMENTS,
  getAdConfig,
  getApiBaseUrl,
  getCurrentEnvironment,
  getEnvironmentConfig,
  isWechatDevtools,
  readEnvVersion,
};
