import { defineStore } from "pinia";
import { login, register, getUserProfile } from "../api/auth";

const USER_INFO_CACHE_TTL_MS = Number(
  import.meta.env.VITE_USER_INFO_CACHE_TTL || 90000,
);
const userInfoFetchInFlight = new Map();
const userInfoFetchedAt = new Map();

const ROLE_KEYS = {
  student: { token: "student_token", userInfo: "student_user_info" },
  enterprise: { token: "enterprise_token", userInfo: "enterprise_user_info" },
  university: { token: "university_token", userInfo: "university_user_info" },
};

const ROLE_NAMES = Object.keys(ROLE_KEYS);
const ROLE_LABEL_MAP = {
  student: "学生用户",
  enterprise: "企业用户",
  university: "高校用户",
};

const normalizeRole = (role = "") => {
  const value = String(role || "")
    .trim()
    .toLowerCase();
  return ROLE_NAMES.includes(value) ? value : "student";
};

const createDefaultRoleAuth = () => {
  return ROLE_NAMES.reduce((acc, role) => {
    acc[role] = { token: "", userInfo: null, profileLoaded: false };
    return acc;
  }, {});
};

const normalizeRoleAuthEntry = (entry = {}) => {
  const token = typeof entry?.token === "string" ? entry.token : "";
  const userInfo = entry?.userInfo || null;
  const profileLoaded = Boolean(entry?.profileLoaded || userInfo);
  return { token, userInfo, profileLoaded };
};

const getHashPath = (hash = "") => {
  const raw = String(hash || "")
    .replace(/^#/, "")
    .trim();
  if (!raw) return "";
  const noQuery = raw.split("?")[0] || "";
  if (!noQuery) return "";
  return noQuery.startsWith("/") ? noQuery : `/${noQuery}`;
};

const resolveRoleFromPath = (path = "") => {
  const pathname = String(path || "")
    .trim()
    .toLowerCase();
  if (pathname.startsWith("/enterprise")) return "enterprise";
  if (pathname.startsWith("/university")) return "university";
  if (pathname.startsWith("/student")) return "student";
  return "";
};

const resolveCurrentRole = () => {
  if (typeof window === "undefined") return "student";
  const roleFromPathname = resolveRoleFromPath(window.location.pathname);
  if (roleFromPathname) return roleFromPathname;
  const roleFromHash = resolveRoleFromPath(getHashPath(window.location.hash));
  if (roleFromHash) return roleFromHash;
  return "student";
};

const decodeJwtPayload = (token = "") => {
  const parts = String(token || "").split(".");
  if (parts.length < 2) return null;
  try {
    const normalized = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const decoded = atob(normalized);
    return JSON.parse(decoded);
  } catch {
    return null;
  }
};

const buildRoleDisplayName = (role, id) => {
  const prefix = ROLE_LABEL_MAP[normalizeRole(role)] || ROLE_LABEL_MAP.student;
  if (id !== undefined && id !== null && String(id).trim() !== "") {
    return `${prefix}#${id}`;
  }
  return prefix;
};

const isInvalidUsername = (username = "") => {
  const value = String(username || "")
    .trim()
    .toLowerCase();
  if (!value) return true;
  return value.includes("user.example");
};

const normalizeUserInfo = (userInfo, role) => {
  if (!userInfo || typeof userInfo !== "object") return null;
  const normalized = { ...userInfo };

  const username = String(normalized.username || "").trim();
  if (isInvalidUsername(username)) {
    normalized.username = buildRoleDisplayName(role, normalized.id);
  } else {
    normalized.username = username;
  }

  if (normalized.email) {
    normalized.email = String(normalized.email).trim();
  }

  return normalized;
};

export const useUserStore = defineStore("user", {
  state: () => ({
    roleAuth: createDefaultRoleAuth(),
  }),
  getters: {
    currentRole: () => resolveCurrentRole(),
    token(state) {
      return state.roleAuth[this.currentRole]?.token || "";
    },
    userInfo(state) {
      return state.roleAuth[this.currentRole]?.userInfo || null;
    },
    userInfoLoaded(state) {
      const auth = state.roleAuth[this.currentRole];
      return Boolean(auth?.profileLoaded || auth?.userInfo);
    },
  },
  actions: {
    getRoleAuth(role = resolveCurrentRole()) {
      const safeRole = normalizeRole(role);
      const found = this.roleAuth[safeRole];
      if (found) {
        const normalized = normalizeRoleAuthEntry(found);
        if (
          found.token !== normalized.token ||
          found.userInfo !== normalized.userInfo ||
          found.profileLoaded !== normalized.profileLoaded
        ) {
          this.roleAuth[safeRole] = normalized;
        }
        return this.roleAuth[safeRole];
      }
      this.roleAuth[safeRole] = {
        token: "",
        userInfo: null,
        profileLoaded: false,
      };
      return this.roleAuth[safeRole];
    },
    getTokenByRole(role = resolveCurrentRole()) {
      return this.getRoleAuth(role).token || "";
    },
    getUserInfoByRole(role = resolveCurrentRole()) {
      return this.getRoleAuth(role).userInfo || null;
    },
    isTokenExpired(token = "") {
      if (!token) return true;
      const payload = decodeJwtPayload(token);
      if (!payload?.exp) return false;
      const nowSeconds = Math.floor(Date.now() / 1000);
      return nowSeconds >= Number(payload.exp);
    },
    hasValidTokenByRole(role = resolveCurrentRole()) {
      const token = this.getTokenByRole(role);
      return !!token && !this.isTokenExpired(token);
    },
    setRoleAuth(role, payload = {}) {
      const safeRole = normalizeRole(role);
      const next = normalizeRoleAuthEntry({
        token: payload.token,
        userInfo: payload.userInfo,
        profileLoaded: payload.profileLoaded,
      });
      this.roleAuth[safeRole] = next;
    },
    setUserInfoByRole(role = resolveCurrentRole(), userInfo = null) {
      const safeRole = normalizeRole(role);
      const auth = this.getRoleAuth(safeRole);
      const normalizedUser = normalizeUserInfo(userInfo, safeRole);

      this.setRoleAuth(safeRole, {
        token: auth.token,
        userInfo: normalizedUser,
        profileLoaded: Boolean(normalizedUser),
      });

      if (normalizedUser) {
        userInfoFetchedAt.set(safeRole, Date.now());
      } else {
        userInfoFetchedAt.delete(safeRole);
      }
    },
    async login(data) {
      const res = await login(data);
      const role = normalizeRole(
        res?.user?.role || data?.role || resolveCurrentRole(),
      );
      this.setRoleAuth(role, {
        token: res.token,
        userInfo: normalizeUserInfo(res?.user, role),
        profileLoaded: false,
      });

      if (res?.token) {
        await this.getUserInfo(role);
      }

      return res;
    },
    async register(data) {
      return register(data);
    },
    async getUserInfo(role = resolveCurrentRole(), options = {}) {
      const safeRole = normalizeRole(role);
      const token = this.getTokenByRole(safeRole);
      if (!token || this.isTokenExpired(token)) {
        this.logout(safeRole);
        return null;
      }

      const forceRefresh = Boolean(options?.force);
      const roleAuth = this.getRoleAuth(safeRole);
      const lastFetchedAt = Number(userInfoFetchedAt.get(safeRole) || 0);
      const cacheAge = Date.now() - lastFetchedAt;

      if (
        !forceRefresh &&
        roleAuth.profileLoaded &&
        roleAuth.userInfo &&
        cacheAge >= 0 &&
        cacheAge < USER_INFO_CACHE_TTL_MS
      ) {
        return roleAuth.userInfo;
      }

      const inFlight = userInfoFetchInFlight.get(safeRole);
      if (inFlight) {
        return inFlight;
      }

      let requestPromise = null;
      requestPromise = getUserProfile()
        .then((res) => {
          this.setUserInfoByRole(safeRole, res?.user);
          return this.getUserInfoByRole(safeRole);
        })
        .catch((error) => {
          const status = Number(error?.response?.status || 0);
          if (status === 401 || status === 403) {
            this.logout(safeRole);
          }
          throw error;
        })
        .finally(() => {
          if (userInfoFetchInFlight.get(safeRole) === requestPromise) {
            userInfoFetchInFlight.delete(safeRole);
          }
        });

      userInfoFetchInFlight.set(safeRole, requestPromise);
      return requestPromise;
    },
    logout(role = resolveCurrentRole()) {
      const safeRole = normalizeRole(role);
      userInfoFetchInFlight.delete(safeRole);
      userInfoFetchedAt.delete(safeRole);
      this.setRoleAuth(safeRole, {
        token: "",
        userInfo: null,
        profileLoaded: false,
      });
    },
    logoutAll() {
      ROLE_NAMES.forEach((role) => {
        userInfoFetchInFlight.delete(role);
        userInfoFetchedAt.delete(role);
        this.setRoleAuth(role, {
          token: "",
          userInfo: null,
          profileLoaded: false,
        });
      });
    },
  },
  persist: true,
});

export { ROLE_KEYS, ROLE_NAMES, resolveRoleFromPath, resolveCurrentRole };
