import { contextBridge, ipcRenderer } from "electron";
import {
  IPC_CHANNELS,
  type NativeBridgeApi,
  type OpenExternalPayload,
  type UpdaterStatusPayload,
} from "./ipc";

const api: NativeBridgeApi = {
  minimizeWindow: () => ipcRenderer.invoke(IPC_CHANNELS.WINDOW_MINIMIZE),
  toggleMaximizeWindow: () =>
    ipcRenderer.invoke(IPC_CHANNELS.WINDOW_TOGGLE_MAXIMIZE),
  closeWindow: () => ipcRenderer.invoke(IPC_CHANNELS.WINDOW_CLOSE),
  openExternal: (url: string) => {
    const payload: OpenExternalPayload = { url };
    return ipcRenderer.invoke(IPC_CHANNELS.APP_OPEN_EXTERNAL, payload);
  },
  checkForUpdates: () => ipcRenderer.invoke(IPC_CHANNELS.UPDATER_CHECK),
  quitAndInstallUpdate: () =>
    ipcRenderer.invoke(IPC_CHANNELS.UPDATER_QUIT_AND_INSTALL),
  onUpdaterStatus: (listener) => {
    const wrapped = (
      _event: Electron.IpcRendererEvent,
      payload: UpdaterStatusPayload,
    ) => {
      listener(payload);
    };

    ipcRenderer.on(IPC_CHANNELS.UPDATER_STATUS, wrapped);
    return () => {
      ipcRenderer.removeListener(IPC_CHANNELS.UPDATER_STATUS, wrapped);
    };
  },
};

contextBridge.exposeInMainWorld("api", api);
