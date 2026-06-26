/**
 * 页面级激励视频广告 controller。
 * 每个页面独立实例，禁止跨页面共享 wx 广告对象。
 *
 * Phase AD0：生产 UI 不引用本模块；仅 dev 环境可用于开发测试。
 * 流量主开通后由 Phase AD1 接入真实激励视频广告。
 */

function createResult(mode, completed, reason) {
  return { mode, completed: Boolean(completed), reason };
}

function createPrepareResult(mode, ready, reason) {
  return { mode, ready: Boolean(ready), reason };
}

function normalizeEnv(env) {
  if (env === null || env === undefined) {
    return "";
  }
  return String(env).trim();
}

function isMockAllowed(config) {
  const enabled = config.enabled !== false;
  const mode = String(config.mode || "disabled");
  const env = normalizeEnv(config.env);
  return enabled && mode === "mock" && env === "dev";
}

function resolvePlaybackMode(config) {
  const enabled = config.enabled !== false;
  const mode = String(config.mode || "disabled");
  const adUnitId = String(config.rewardedVideoAdUnitId || "").trim();

  if (!enabled || mode === "disabled") {
    return { playback: "disabled", reportMode: "disabled" };
  }

  if (mode === "mock") {
    if (isMockAllowed(config)) {
      return { playback: "mock", reportMode: "mock" };
    }
    return { playback: "disabled", reportMode: "disabled" };
  }

  if (mode === "wechat") {
    if (!adUnitId) {
      return { playback: "invalid", reportMode: "wechat" };
    }
    return { playback: "wechat", reportMode: "wechat" };
  }

  return { playback: "disabled", reportMode: "disabled" };
}

function createRewardedAdController(config = {}) {
  let disposed = false;
  let busy = false;
  let pending = null;
  let timerIds = [];
  let adInstance = null;
  let adHandlers = null;

  function clearTimers() {
    timerIds.forEach((id) => clearTimeout(id));
    timerIds = [];
  }

  function schedule(fn, delayMs) {
    const id = setTimeout(fn, delayMs);
    timerIds.push(id);
    return id;
  }

  function settle(result) {
    if (!pending) return;
    const resolve = pending.resolve;
    pending = null;
    busy = false;
    resolve(result);
  }

  function detachAdHandlers() {
    if (adInstance && adHandlers) {
      if (typeof adInstance.offLoad === "function") {
        adInstance.offLoad(adHandlers.onLoad);
      }
      if (typeof adInstance.offError === "function") {
        adInstance.offError(adHandlers.onError);
      }
      if (typeof adInstance.offClose === "function") {
        adInstance.offClose(adHandlers.onClose);
      }
    }
    adHandlers = null;
  }

  function destroyAdInstance() {
    detachAdHandlers();
    if (adInstance && typeof adInstance.destroy === "function") {
      try {
        adInstance.destroy();
      } catch (_error) {
        // 忽略销毁异常，避免影响页面卸载。
      }
    }
    adInstance = null;
  }

  function getPlaybackConfig() {
    return resolvePlaybackMode(config);
  }

  function showMock() {
    const { reportMode } = getPlaybackConfig();
    const mockOutcome =
      String(config.mockOutcome || "completed").toLowerCase() === "cancelled"
        ? "cancelled"
        : "completed";

    if (mockOutcome === "cancelled") {
      settle(createResult(reportMode, false, "cancelled"));
      return;
    }

    const durationMs = Math.max(0, Number(config.mockDurationMs) || 5000);
    if (typeof wx !== "undefined" && wx.showLoading) {
      wx.showLoading({ title: "模拟播放中…", mask: true });
    }

    schedule(() => {
      if (typeof wx !== "undefined" && wx.hideLoading) {
        wx.hideLoading();
      }
      settle(createResult(reportMode, true, "completed"));
    }, durationMs);
  }

  function ensureWechatAdInstance() {
    if (adInstance) return adInstance;

    const adUnitId = String(config.rewardedVideoAdUnitId || "").trim();
    if (!adUnitId || typeof wx === "undefined" || !wx.createRewardedVideoAd) {
      return null;
    }

    adInstance = wx.createRewardedVideoAd({ adUnitId });
    return adInstance;
  }

  function showWechat() {
    const { reportMode } = getPlaybackConfig();
    const ad = ensureWechatAdInstance();
    if (!ad) {
      settle(createResult(reportMode, false, "unsupported"));
      return;
    }

    detachAdHandlers();

    adHandlers = {
      onLoad: () => {},
      onError: () => {
        settle(createResult(reportMode, false, "load_failed"));
      },
      onClose: (res) => {
        const isEnded = res && res.isEnded === true;
        if (isEnded) {
          settle(createResult(reportMode, true, "completed"));
        } else {
          settle(createResult(reportMode, false, "cancelled"));
        }
      },
    };

    ad.onLoad(adHandlers.onLoad);
    ad.onError(adHandlers.onError);
    ad.onClose(adHandlers.onClose);

    ad
      .load()
      .then(() => ad.show())
      .catch(() => {
        settle(createResult(reportMode, false, "show_failed"));
      });
  }

  return {
    prepare() {
      if (disposed || busy) {
        return Promise.resolve(createPrepareResult("disabled", false, "disabled"));
      }

      const { playback, reportMode } = getPlaybackConfig();
      if (playback === "disabled") {
        return Promise.resolve(createPrepareResult(reportMode, false, "disabled"));
      }
      if (playback === "invalid") {
        return Promise.resolve(createPrepareResult(reportMode, false, "invalid_config"));
      }
      if (playback === "mock") {
        return Promise.resolve(createPrepareResult(reportMode, true, "ready"));
      }

      const ad = ensureWechatAdInstance();
      if (!ad) {
        return Promise.resolve(createPrepareResult(reportMode, false, "unsupported"));
      }

      return ad
        .load()
        .then(() => createPrepareResult(reportMode, true, "ready"))
        .catch(() => createPrepareResult(reportMode, false, "load_failed"));
    },

    show(_options) {
      if (disposed) {
        return Promise.resolve(createResult("disabled", false, "page_unloaded"));
      }
      if (busy) {
        const { reportMode } = getPlaybackConfig();
        return Promise.resolve(createResult(reportMode, false, "busy"));
      }

      const { playback, reportMode } = getPlaybackConfig();

      if (playback === "disabled") {
        return Promise.resolve(createResult(reportMode, false, "disabled"));
      }
      if (playback === "invalid") {
        return Promise.resolve(createResult(reportMode, false, "invalid_config"));
      }

      busy = true;

      return new Promise((resolve) => {
        pending = { resolve };

        if (playback === "mock") {
          showMock();
          return;
        }

        if (playback === "wechat") {
          showWechat();
          return;
        }

        settle(createResult(reportMode, false, "unsupported"));
      });
    },

    dispose() {
      disposed = true;
      clearTimers();
      if (typeof wx !== "undefined" && wx.hideLoading) {
        wx.hideLoading();
      }

      if (pending) {
        const { reportMode } = getPlaybackConfig();
        settle(createResult(reportMode, false, "page_unloaded"));
      }

      destroyAdInstance();
      busy = false;
    },
  };
}

module.exports = {
  createRewardedAdController,
  resolvePlaybackMode,
};
