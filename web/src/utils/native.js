/**
 * @typedef {Object} NativeApi
 * @property {() => Promise<boolean>} minimizeWindow
 * @property {() => Promise<boolean>} toggleMaximizeWindow
 * @property {() => Promise<boolean>} closeWindow
 * @property {(url: string) => Promise<boolean>} openExternal
 * @property {() => Promise<boolean>} checkForUpdates
 * @property {() => Promise<boolean>} quitAndInstallUpdate
 * @property {(listener: (payload: Record<string, unknown>) => void) => () => void} onUpdaterStatus
 */

const noopResult = Promise.resolve(false);

const browserFallback = {
  /** @returns {Promise<boolean>} */
  minimizeWindow() {
    return noopResult;
  },

  /** @returns {Promise<boolean>} */
  toggleMaximizeWindow() {
    return noopResult;
  },

  /** @returns {Promise<boolean>} */
  closeWindow() {
    return noopResult;
  },

  /** @param {string} url */
  async openExternal(url) {
    const target = String(url || "").trim();
    if (!target || typeof window === "undefined") {
      return false;
    }

    const opened = window.open(target, "_blank", "noopener,noreferrer");
    return Boolean(opened);
  },

  /** @returns {Promise<boolean>} */
  checkForUpdates() {
    return noopResult;
  },

  /** @returns {Promise<boolean>} */
  quitAndInstallUpdate() {
    return noopResult;
  },

  /** @returns {() => void} */
  onUpdaterStatus() {
    return () => {};
  },
};

const isValidApiObject = (candidate) => {
  if (!candidate || typeof candidate !== "object") return false;

  return (
    typeof candidate.minimizeWindow === "function" &&
    typeof candidate.toggleMaximizeWindow === "function" &&
    typeof candidate.closeWindow === "function" &&
    typeof candidate.openExternal === "function" &&
    typeof candidate.checkForUpdates === "function" &&
    typeof candidate.quitAndInstallUpdate === "function" &&
    typeof candidate.onUpdaterStatus === "function"
  );
};

/** @returns {NativeApi | null} */
const getElectronApi = () => {
  if (typeof window === "undefined") return null;
  const candidate = window.api;
  return isValidApiObject(candidate) ? candidate : null;
};

const withFallback = async (primary, fallback) => {
  try {
    return await primary();
  } catch (error) {
    console.warn("[native] call failed, using fallback:", error);
    return fallback();
  }
};

export const isElectronRuntime = () => Boolean(getElectronApi());

/** @type {NativeApi} */
export const native = {
  minimizeWindow() {
    const api = getElectronApi();
    if (!api) return browserFallback.minimizeWindow();
    return withFallback(
      () => api.minimizeWindow(),
      () => browserFallback.minimizeWindow(),
    );
  },

  toggleMaximizeWindow() {
    const api = getElectronApi();
    if (!api) return browserFallback.toggleMaximizeWindow();
    return withFallback(
      () => api.toggleMaximizeWindow(),
      () => browserFallback.toggleMaximizeWindow(),
    );
  },

  closeWindow() {
    const api = getElectronApi();
    if (!api) return browserFallback.closeWindow();
    return withFallback(
      () => api.closeWindow(),
      () => browserFallback.closeWindow(),
    );
  },

  openExternal(url) {
    const api = getElectronApi();
    if (!api) return browserFallback.openExternal(url);
    return withFallback(
      () => api.openExternal(url),
      () => browserFallback.openExternal(url),
    );
  },

  checkForUpdates() {
    const api = getElectronApi();
    if (!api) return browserFallback.checkForUpdates();
    return withFallback(
      () => api.checkForUpdates(),
      () => browserFallback.checkForUpdates(),
    );
  },

  quitAndInstallUpdate() {
    const api = getElectronApi();
    if (!api) return browserFallback.quitAndInstallUpdate();
    return withFallback(
      () => api.quitAndInstallUpdate(),
      () => browserFallback.quitAndInstallUpdate(),
    );
  },

  onUpdaterStatus(listener) {
    const api = getElectronApi();
    if (!api || typeof listener !== "function") {
      return browserFallback.onUpdaterStatus();
    }

    try {
      return api.onUpdaterStatus(listener);
    } catch (error) {
      console.warn("[native] subscribe updater status failed:", error);
      return browserFallback.onUpdaterStatus();
    }
  },
};

export default native;
