export const IPC_CHANNELS = {
  WINDOW_MINIMIZE: "window:minimize",
  WINDOW_TOGGLE_MAXIMIZE: "window:toggle-maximize",
  WINDOW_CLOSE: "window:close",
  APP_OPEN_EXTERNAL: "app:open-external",
  UPDATER_CHECK: "updater:check",
  UPDATER_QUIT_AND_INSTALL: "updater:quit-and-install",
  UPDATER_STATUS: "updater:status",
} as const;

export type OpenExternalPayload = {
  url: string;
};

export type UpdaterStatusPayload = {
  stage:
    | "idle"
    | "checking"
    | "available"
    | "not-available"
    | "downloading"
    | "downloaded"
    | "error";
  message?: string;
  version?: string;
  percent?: number;
  transferred?: number;
  total?: number;
  bytesPerSecond?: number;
};

export interface NativeBridgeApi {
  minimizeWindow: () => Promise<boolean>;
  toggleMaximizeWindow: () => Promise<boolean>;
  closeWindow: () => Promise<boolean>;
  openExternal: (url: string) => Promise<boolean>;
  checkForUpdates: () => Promise<boolean>;
  quitAndInstallUpdate: () => Promise<boolean>;
  onUpdaterStatus: (
    listener: (payload: UpdaterStatusPayload) => void,
  ) => () => void;
}
