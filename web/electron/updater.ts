import {
  app,
  BrowserWindow,
  dialog,
  ipcMain,
  type MessageBoxOptions,
} from "electron";
import { autoUpdater } from "electron-updater";
import type { ProgressInfo, UpdateInfo } from "electron-updater";
import { IPC_CHANNELS, type UpdaterStatusPayload } from "./ipc";

let updaterInitialized = false;
let hasDownloadedUpdate = false;
let isChecking = false;

const broadcastUpdaterStatus = (payload: UpdaterStatusPayload) => {
  for (const win of BrowserWindow.getAllWindows()) {
    if (win.isDestroyed()) continue;
    win.webContents.send(IPC_CHANNELS.UPDATER_STATUS, payload);
  }
};

const registerUpdaterEvents = () => {
  autoUpdater.on("checking-for-update", () => {
    broadcastUpdaterStatus({ stage: "checking", message: "正在检查更新..." });
  });

  autoUpdater.on("update-available", (info: UpdateInfo) => {
    broadcastUpdaterStatus({
      stage: "available",
      version: info.version,
      message: `发现新版本 ${info.version}，开始下载...`,
    });

    autoUpdater.downloadUpdate().catch((error: Error) => {
      broadcastUpdaterStatus({
        stage: "error",
        message: `下载更新失败：${error.message}`,
      });
    });
  });

  autoUpdater.on("update-not-available", () => {
    broadcastUpdaterStatus({
      stage: "not-available",
      message: "当前已是最新版本",
    });
  });

  autoUpdater.on("download-progress", (progress: ProgressInfo) => {
    broadcastUpdaterStatus({
      stage: "downloading",
      percent: progress.percent,
      transferred: progress.transferred,
      total: progress.total,
      bytesPerSecond: progress.bytesPerSecond,
      message: `正在下载更新 ${progress.percent.toFixed(1)}%`,
    });
  });

  autoUpdater.on("update-downloaded", async (info: UpdateInfo) => {
    hasDownloadedUpdate = true;

    broadcastUpdaterStatus({
      stage: "downloaded",
      version: info.version,
      message: "更新已下载完成，重启后安装",
    });

    const focusedWindow = BrowserWindow.getFocusedWindow();
    const messageBoxOptions: MessageBoxOptions = {
      type: "info",
      title: "更新已就绪",
      message: "新版本已下载完成，是否立即重启安装？",
      buttons: ["立即重启", "稍后"],
      defaultId: 0,
      cancelId: 1,
      noLink: true,
      detail: `版本：${info.version}`,
    };

    const { response } = focusedWindow
      ? await dialog.showMessageBox(focusedWindow, messageBoxOptions)
      : await dialog.showMessageBox(messageBoxOptions);

    if (response === 0) {
      autoUpdater.quitAndInstall(false, true);
    }
  });

  autoUpdater.on("error", (error: Error) => {
    broadcastUpdaterStatus({
      stage: "error",
      message: error?.message || "更新服务异常",
    });
  });
};

export const initUpdater = async () => {
  if (updaterInitialized) return;
  updaterInitialized = true;

  autoUpdater.autoDownload = false;
  autoUpdater.autoInstallOnAppQuit = true;

  registerUpdaterEvents();

  if (!app.isPackaged) {
    broadcastUpdaterStatus({
      stage: "idle",
      message: "开发环境不执行自动更新检查",
    });
    return;
  }

  setTimeout(() => {
    void checkForUpdates();
  }, 3000);
};

export const checkForUpdates = async (): Promise<boolean> => {
  if (!app.isPackaged) {
    broadcastUpdaterStatus({
      stage: "error",
      message: "仅打包后的应用支持自动更新",
    });
    return false;
  }

  if (isChecking) return false;

  try {
    isChecking = true;
    await autoUpdater.checkForUpdates();
    return true;
  } catch (error) {
    broadcastUpdaterStatus({
      stage: "error",
      message: `检查更新失败：${(error as Error).message}`,
    });
    return false;
  } finally {
    isChecking = false;
  }
};

export const quitAndInstallUpdate = async (): Promise<boolean> => {
  if (!hasDownloadedUpdate) {
    broadcastUpdaterStatus({
      stage: "error",
      message: "尚未下载可安装的更新包",
    });
    return false;
  }

  autoUpdater.quitAndInstall(false, true);
  return true;
};

export const registerUpdaterIpc = () => {
  ipcMain.removeHandler(IPC_CHANNELS.UPDATER_CHECK);
  ipcMain.removeHandler(IPC_CHANNELS.UPDATER_QUIT_AND_INSTALL);

  ipcMain.handle(IPC_CHANNELS.UPDATER_CHECK, async () => {
    return checkForUpdates();
  });

  ipcMain.handle(IPC_CHANNELS.UPDATER_QUIT_AND_INSTALL, async () => {
    return quitAndInstallUpdate();
  });
};
