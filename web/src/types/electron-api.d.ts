type NativeApi = {
  minimizeWindow: () => Promise<boolean>;
  toggleMaximizeWindow: () => Promise<boolean>;
  closeWindow: () => Promise<boolean>;
  openExternal: (url: string) => Promise<boolean>;
  checkForUpdates: () => Promise<boolean>;
  quitAndInstallUpdate: () => Promise<boolean>;
  onUpdaterStatus: (
    listener: (payload: {
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
    }) => void,
  ) => () => void;
};

declare global {
  interface Window {
    api?: NativeApi;
  }
}

export {};
