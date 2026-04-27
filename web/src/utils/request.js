import axios from "axios";
import { ElMessage } from "element-plus";
import router from "../router/index.js";
import { useUserStore, resolveRoleFromPath } from "../stores/user";

const REQUEST_TIMEOUT = Number(import.meta.env.VITE_API_TIMEOUT || 60000);
const DEFAULT_API_BASE_URL = "/api/v1";

const resolveApiBaseURL = () => {
  const configured = String(
    import.meta.env.VITE_API_BASE_URL ||
      import.meta.env.VITE_API_URL ||
      DEFAULT_API_BASE_URL,
  ).trim();
  return configured || DEFAULT_API_BASE_URL;
};

const normalizeBackendErrorMessage = (msg = "") => {
  const text = String(msg || "");
  if (!text) return text;
  if (
    /invalid credentials|incorrect password|wrong password|authentication failed/i.test(
      text,
    )
  ) {
    return "账号或密码错误，请重新输入";
  }
  if (/user not found|account not found/i.test(text)) {
    return "账号不存在，请检查后重试";
  }
  if (/network error/i.test(text)) {
    return "网络异常，请检查连接后重试";
  }
  if (
    /field\s+validation.*answer.*required/i.test(text) ||
    /key:\s*'answer'/i.test(text)
  ) {
    return "您似乎没有做出任何回答";
  }
  if (
    /failed\s+to\s+transcribe\s+audio/i.test(text) ||
    /empty\s+transcription\s+result/i.test(text)
  ) {
    return "未识别到有效语音，请靠近麦克风并清晰作答后重试";
  }
  return text;
};

const unwrapBackendErrorMessage = (payload) => {
  if (!payload) return "";
  if (typeof payload === "string") return payload;
  if (typeof payload?.message === "string") return payload.message;
  if (typeof payload?.error === "string") return payload.error;
  if (typeof payload?.error?.message === "string") return payload.error.message;
  return "";
};

const extractNormalizedErrorMessage = (error) => {
  const message =
    unwrapBackendErrorMessage(error?.response?.data) ||
    unwrapBackendErrorMessage(error?.response?.data?.error) ||
    error?.response?.data?.message ||
    error?.message ||
    "";
  return normalizeBackendErrorMessage(message);
};

const service = axios.create({
  timeout: REQUEST_TIMEOUT,
});

const ROLE_AUTH_STRATEGY = new Map([
  ["enterprise", { loginPath: "/enterprise/login" }],
  ["university", { loginPath: "/university/login" }],
  ["student", { loginPath: "/student/login" }],
]);

const resolveRoleContext = (path = "") => {
  const role = resolveRoleFromPath(path);
  const strategy =
    ROLE_AUTH_STRATEGY.get(role) || ROLE_AUTH_STRATEGY.get("student");
  return {
    role,
    loginPath: strategy.loginPath,
  };
};

const getActiveRoute = () => {
  const currentRoute = router?.currentRoute?.value;
  if (currentRoute) return currentRoute;

  if (typeof window === "undefined") {
    return { path: "/", fullPath: "/" };
  }

  const rawHash = String(window.location.hash || "");
  if (rawHash.startsWith("#/")) {
    const hashPath = rawHash.slice(1);
    return { path: hashPath.split("?")[0] || "/", fullPath: hashPath || "/" };
  }

  const path = window.location.pathname || "/";
  const fullPath = `${path}${window.location.search || ""}`;
  return { path, fullPath: fullPath || "/" };
};

const getCurrentPathWithQuery = () => {
  return getActiveRoute().fullPath || "/";
};

const normalizeRequestUrl = (url = "") => {
  const raw = String(url || "").trim();
  if (!raw) return "";
  return raw.startsWith("/") ? raw : `/${raw}`;
};

const PUBLIC_ENDPOINTS = new Set(["/login", "/register", "/security/contact"]);

const isPublicEndpoint = (url = "") => {
  const path = normalizeRequestUrl(url);
  return PUBLIC_ENDPOINTS.has(path);
};

const shouldBypassAuthEnforcement = (config = {}) => {
  const currentPath = getActiveRoute().path;
  if (currentPath.endsWith("/login") || currentPath === "/") return true;
  if (
    String(readHeaderValue(config?.headers, "X-Skip-Auth")).toLowerCase() ===
    "true"
  )
    return true;
  return isPublicEndpoint(config?.url);
};

const readHeaderValue = (headers, key) => {
  if (!headers) return "";
  if (typeof headers.get === "function") {
    return headers.get(key) || headers.get(String(key).toLowerCase()) || "";
  }
  return headers[key] || headers[String(key).toLowerCase()] || "";
};

const shouldSkipErrorToast = (error) => {
  const headerValue = readHeaderValue(
    error?.config?.headers,
    "X-Skip-Error-Toast",
  );
  return String(headerValue || "").toLowerCase() === "true";
};

const redirectToRoleLogin = (loginPath, fromPath) => {
  const redirect = encodeURIComponent(fromPath || "/");
  router.replace(`${loginPath}?redirect=${redirect}`).catch(() => {});
};

const enforceRoleLogoutAndRedirect = ({ role, loginPath }) => {
  const userStore = useUserStore();
  userStore.logout(role);
  redirectToRoleLogin(loginPath, getCurrentPathWithQuery());
};

service.interceptors.request.use(
  (config) => {
    config.baseURL = config.baseURL || resolveApiBaseURL();

    const userStore = useUserStore();
    const currentPath = getActiveRoute().path;
    const { role, loginPath } = resolveRoleContext(currentPath);
    const token = userStore.getTokenByRole(role);

    if (token && !userStore.isTokenExpired(token)) {
      config.headers = config.headers || {};
      config.headers["Authorization"] = `Bearer ${token}`;
      return config;
    }

    if (!shouldBypassAuthEnforcement(config)) {
      enforceRoleLogoutAndRedirect({ role, loginPath });
      return Promise.reject(new Error("登录状态已失效，请重新登录"));
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  },
);

service.interceptors.response.use(
  (response) => {
    return response.data;
  },
  (error) => {
    const res = error?.response;
    const normalizedMessage = extractNormalizedErrorMessage(error);
    const skipErrorToast = shouldSkipErrorToast(error);

    if (res?.data?.error) {
      const normalizedError = normalizeBackendErrorMessage(
        unwrapBackendErrorMessage(res.data.error) || res.data.error,
      );
      if (typeof res.data.error === "object" && res.data.error !== null) {
        res.data.error = {
          ...res.data.error,
          message: normalizedError,
        };
      } else {
        res.data.error = normalizedError;
      }
    }
    if (normalizedMessage) {
      error.message = normalizedMessage;
    }

    if (res?.status === 401) {
      const msg = (res.data && res.data.error) || normalizedMessage || "";
      if (/invalid token/i.test(msg) || /authorization/i.test(msg)) {
        const currentPath = getActiveRoute().path;
        const context = resolveRoleContext(currentPath);
        enforceRoleLogoutAndRedirect(context);
      }
    } else if (error?.code === "ECONNABORTED") {
      error.message = "请求超时：服务响应时间较长，请稍后重试。";
    }

    const isCanceled = axios.isCancel(error) || error?.code === "ERR_CANCELED";
    if (res?.status !== 401 && !isCanceled && !skipErrorToast) {
      ElMessage.error(error?.message || "请求失败，请稍后重试");
    }

    return Promise.reject(error);
  },
);

export default service;
