import {
  createRouter,
  createWebHashHistory,
  createWebHistory,
} from "vue-router";
import { resolveRoleFromPath, useUserStore } from "../stores/user";
import { commonRoutes } from "./routes/common";
import { studentRoutes } from "./routes/student";
import { enterpriseRoutes } from "./routes/enterprise";
import { universityRoutes } from "./routes/university";

const ROLE_NAMES = ["student", "enterprise", "university"];
const PROFILE_LOAD_BUDGET_MS = Number(
  import.meta.env.VITE_PROFILE_LOAD_BUDGET_MS || 1200,
);
const ROLE_DASHBOARD = {
  student: "/student/dashboard",
  enterprise: "/enterprise/dashboard",
  university: "/university/dashboard",
};

const routes = [
  ...commonRoutes,
  ...studentRoutes,
  ...enterpriseRoutes,
  ...universityRoutes,
];

const resolveRouterHistory = () => {
  const routerMode = String(
    import.meta.env.VITE_ROUTER_MODE || "",
  ).toLowerCase();
  const buildTarget = String(
    import.meta.env.VITE_BUILD_TARGET || "",
  ).toLowerCase();
  const useHash = routerMode === "hash" || buildTarget === "electron";
  return useHash ? createWebHashHistory() : createWebHistory();
};

const router = createRouter({
  history: resolveRouterHistory(),
  routes,
});

const isKnownRole = (role) => ROLE_NAMES.includes(role);

const normalizeRole = (role) => {
  const value = String(role || "")
    .trim()
    .toLowerCase();
  if (value === "teacher" || value === "mentor" || value === "faculty") {
    return "university";
  }
  if (value === "hr" || value === "interviewer" || value === "recruiter") {
    return "enterprise";
  }
  if (isKnownRole(value)) return value;
  return "student";
};

const readTokenByRole = (userStore, role) => {
  const safeRole = normalizeRole(role);
  return userStore.getTokenByRole(safeRole);
};

const findLoggedRole = (userStore) => {
  for (const role of ROLE_NAMES) {
    const token = readTokenByRole(userStore, role);
    if (token && !userStore.isTokenExpired(token)) {
      return role;
    }
  }
  return "";
};

const buildLoginPath = (role, redirectPath) => {
  const safeRole = normalizeRole(role);
  const target = String(redirectPath || "/").trim() || "/";
  return `/${safeRole}/login?redirect=${encodeURIComponent(target)}`;
};

const buildForbiddenLocation = (to, expectedRoles, actualRole) => {
  return {
    path: "/403",
    query: {
      from: to.fullPath,
      expected: expectedRoles.join(","),
      actual: String(actualRole || ""),
    },
  };
};

const extractRequiredRoles = (to, fallbackRole) => {
  const matchedRoles = to.matched
    .flatMap((record) => {
      const roles = record.meta?.roles;
      if (!Array.isArray(roles)) return [];
      return roles;
    })
    .filter(isKnownRole);

  const uniqueRoles = [...new Set(matchedRoles)];
  if (uniqueRoles.length > 0) return uniqueRoles;

  return [normalizeRole(fallbackRole)];
};

router.beforeEach(async (to) => {
  const userStore = useUserStore();
  const pathRole = normalizeRole(resolveRoleFromPath(to.path));
  const requiresAuth = to.matched.some(
    (record) => record.meta?.requiresAuth === true,
  );

  if (to.path === "/") {
    const loggedRole = findLoggedRole(userStore);
    if (loggedRole) {
      return ROLE_DASHBOARD[loggedRole];
    }
  }

  if (to.path.endsWith("/login")) {
    const token = readTokenByRole(userStore, pathRole);
    if (token && !userStore.isTokenExpired(token)) {
      try {
        await userStore.getUserInfo(pathRole);
      } catch {
        // token may still be unexpired locally but invalid on server (e.g. seeded user replaced)
        return true;
      }

      const redirectTarget = String(to.query?.redirect || "").trim();
      if (redirectTarget) {
        return redirectTarget;
      }
      return ROLE_DASHBOARD[pathRole];
    }
    return true;
  }

  if (!requiresAuth) {
    return true;
  }

  const token = readTokenByRole(userStore, pathRole);
  if (!token || userStore.isTokenExpired(token)) {
    userStore.logout(pathRole);

    const loggedRole = findLoggedRole(userStore);
    if (loggedRole && loggedRole !== pathRole) {
      return ROLE_DASHBOARD[loggedRole];
    }

    return buildLoginPath(pathRole, to.fullPath);
  }

  const roleAuth = userStore.getRoleAuth(pathRole);
  if (!roleAuth.userInfo || !roleAuth.profileLoaded) {
    try {
      await Promise.race([
        userStore.getUserInfo(pathRole),
        new Promise((resolve) => {
          setTimeout(resolve, PROFILE_LOAD_BUDGET_MS);
        }),
      ]);
    } catch {
      // Profile request can fail transiently; keep token-based access path.
    }
  }

  if (!userStore.hasValidTokenByRole(pathRole)) {
    return buildLoginPath(pathRole, to.fullPath);
  }

  const requiredRoles = extractRequiredRoles(to, pathRole);
  const userInfo = userStore.getUserInfoByRole(pathRole);
  const actualRole = normalizeRole(userInfo?.role || pathRole);

  if (!requiredRoles.includes(actualRole)) {
    const fallback = ROLE_DASHBOARD[actualRole];
    if (fallback && fallback !== to.path) {
      return fallback;
    }
    return buildForbiddenLocation(to, requiredRoles, actualRole);
  }

  return true;
});

export default router;
