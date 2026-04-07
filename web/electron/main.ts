import {
  app,
  BrowserWindow,
  ipcMain,
  Menu,
  nativeImage,
  shell,
  Tray,
} from "electron";
import { existsSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { IPC_CHANNELS, type OpenExternalPayload } from "./ipc";
import { initUpdater, registerUpdaterIpc } from "./updater";

const __dirname = dirname(fileURLToPath(import.meta.url));
let mainWindow: BrowserWindow | null = null;
let tray: Tray | null = null;
let isQuitting = false;
const isAutoStartLaunch = process.argv.includes("--autostart");

const resolvePreloadPath = (): string => {
  const candidates = ["preload.mjs", "preload.js", "preload.cjs"];
  for (const fileName of candidates) {
    const filePath = join(__dirname, fileName);
    if (existsSync(filePath)) return filePath;
  }
  return join(__dirname, "preload.mjs");
};

const isAllowedExternalUrl = (rawUrl: string): boolean => {
  try {
    const parsed = new URL(String(rawUrl || "").trim());
    return ["http:", "https:", "mailto:"].includes(parsed.protocol);
  } catch {
    return false;
  }
};

const buildFallbackTrayIcon = () => {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><rect width="64" height="64" rx="14" fill="#2563eb"/><path d="M19 22h26v6H35v14h-6V28H19z" fill="#fff"/></svg>`;
  const dataUrl = `data:image/svg+xml;base64,${Buffer.from(svg).toString("base64")}`;
  return nativeImage.createFromDataURL(dataUrl);
};

const resolveTrayIcon = () => {
  const candidates = [
    join(process.cwd(), "public", "vite.svg"),
    join(__dirname, "../dist/vite.svg"),
  ];

  for (const iconPath of candidates) {
    if (!existsSync(iconPath)) continue;
    const icon = nativeImage.createFromPath(iconPath);
    if (!icon.isEmpty()) return icon.resize({ width: 18, height: 18 });
  }

  return buildFallbackTrayIcon().resize({ width: 18, height: 18 });
};

const showMainWindow = () => {
  if (!mainWindow || mainWindow.isDestroyed()) {
    createWindow();
    return;
  }

  if (mainWindow.isMinimized()) {
    mainWindow.restore();
  }
  mainWindow.show();
  mainWindow.focus();
};

const createTray = () => {
  if (tray) return;

  tray = new Tray(resolveTrayIcon());
  tray.setToolTip("Interview AI");

  tray.setContextMenu(
    Menu.buildFromTemplate([
      {
        label: "显示主界面",
        click: () => showMainWindow(),
      },
      {
        type: "separator",
      },
      {
        label: "退出应用",
        click: () => {
          isQuitting = true;
          app.quit();
        },
      },
    ]),
  );

  tray.on("click", () => {
    showMainWindow();
  });
};

const withSenderWindow = <T>(
  event: Electron.IpcMainInvokeEvent,
  callback: (win: BrowserWindow) => T,
  fallback: T,
): T => {
  const win = BrowserWindow.fromWebContents(event.sender);
  if (!win) return fallback;
  return callback(win);
};

const registerIpcHandlers = () => {
  ipcMain.removeHandler(IPC_CHANNELS.WINDOW_MINIMIZE);
  ipcMain.removeHandler(IPC_CHANNELS.WINDOW_TOGGLE_MAXIMIZE);
  ipcMain.removeHandler(IPC_CHANNELS.WINDOW_CLOSE);
  ipcMain.removeHandler(IPC_CHANNELS.APP_OPEN_EXTERNAL);

  ipcMain.handle(IPC_CHANNELS.WINDOW_MINIMIZE, (event) => {
    return withSenderWindow(
      event,
      (win) => {
        win.minimize();
        return true;
      },
      false,
    );
  });

  ipcMain.handle(IPC_CHANNELS.WINDOW_TOGGLE_MAXIMIZE, (event) => {
    return withSenderWindow(
      event,
      (win) => {
        if (win.isMaximized()) {
          win.unmaximize();
          return false;
        }
        win.maximize();
        return true;
      },
      false,
    );
  });

  ipcMain.handle(IPC_CHANNELS.WINDOW_CLOSE, (event) => {
    return withSenderWindow(
      event,
      (win) => {
        win.close();
        return true;
      },
      false,
    );
  });

  ipcMain.handle(
    IPC_CHANNELS.APP_OPEN_EXTERNAL,
    async (_event, payload: OpenExternalPayload) => {
      const targetUrl = String(payload?.url || "").trim();
      if (!isAllowedExternalUrl(targetUrl)) {
        throw new Error("Only http(s) and mailto links are allowed.");
      }
      await shell.openExternal(targetUrl);
      return true;
    },
  );
};

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 800,
    minWidth: 1024,
    minHeight: 640,
    frame: false,
    show: false,
    webPreferences: {
      preload: resolvePreloadPath(),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });

  mainWindow.once("ready-to-show", () => {
    if (isAutoStartLaunch) return;
    mainWindow?.show();
  });

  mainWindow.on("close", (event) => {
    if (isQuitting || process.platform === "darwin") return;
    event.preventDefault();
    mainWindow?.hide();
  });

  mainWindow.on("closed", () => {
    mainWindow = null;
  });

  if (process.env.VITE_DEV_SERVER_URL) {
    mainWindow.loadURL(process.env.VITE_DEV_SERVER_URL);
  } else {
    mainWindow.loadFile(join(__dirname, "../dist/index.html"));
  }
}

app.whenReady().then(() => {
  registerIpcHandlers();
  registerUpdaterIpc();
  createTray();
  createWindow();
  void initUpdater();

  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin" && isQuitting) {
    app.quit();
  }
});

app.on("before-quit", () => {
  isQuitting = true;
});
