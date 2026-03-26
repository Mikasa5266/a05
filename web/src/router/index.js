import { createRouter, createWebHistory } from 'vue-router'
import { ROLE_KEYS, resolveRoleFromPath, useUserStore } from '../stores/user'
import { commonRoutes } from './routes/common'
import { studentRoutes } from './routes/student'
import { enterpriseRoutes } from './routes/enterprise'
import { universityRoutes } from './routes/university'

const ROLE_NAMES = ['student', 'enterprise', 'university']
const ROLE_DASHBOARD = {
  student: '/student/dashboard',
  enterprise: '/enterprise/dashboard',
  university: '/university/dashboard'
}

const routes = [
  ...commonRoutes,
  ...studentRoutes,
  ...enterpriseRoutes,
  ...universityRoutes
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

const isKnownRole = (role) => ROLE_NAMES.includes(role)

const normalizeRole = (role) => {
  if (isKnownRole(role)) return role
  return 'student'
}

const readTokenFromStorageByRole = (role) => {
  const safeRole = normalizeRole(role)
  const key = ROLE_KEYS[safeRole]?.token
  if (!key) return ''
  return localStorage.getItem(key) || ''
}

const findLoggedRole = (userStore) => {
  for (const role of ROLE_NAMES) {
    const token = readTokenFromStorageByRole(role)
    if (token && !userStore.isTokenExpired(token)) {
      return role
    }
  }
  return ''
}

const buildLoginPath = (role, redirectPath) => {
  const safeRole = normalizeRole(role)
  const target = String(redirectPath || '/').trim() || '/'
  return `/${safeRole}/login?redirect=${encodeURIComponent(target)}`
}

const buildForbiddenLocation = (to, expectedRoles, actualRole) => {
  return {
    path: '/403',
    query: {
      from: to.fullPath,
      expected: expectedRoles.join(','),
      actual: String(actualRole || '')
    }
  }
}

const extractRequiredRoles = (to, fallbackRole) => {
  const matchedRoles = to.matched
    .flatMap((record) => {
      const roles = record.meta?.roles
      if (!Array.isArray(roles)) return []
      return roles
    })
    .filter(isKnownRole)

  const uniqueRoles = [...new Set(matchedRoles)]
  if (uniqueRoles.length > 0) return uniqueRoles

  return [normalizeRole(fallbackRole)]
}

router.beforeEach((to) => {
  const userStore = useUserStore()
  const pathRole = normalizeRole(resolveRoleFromPath(to.path))
  const requiresAuth = to.matched.some((record) => record.meta?.requiresAuth === true)

  if (to.path === '/') {
    const loggedRole = findLoggedRole(userStore)
    if (loggedRole) {
      return ROLE_DASHBOARD[loggedRole]
    }
  }

  if (to.path.endsWith('/login')) {
    const token = readTokenFromStorageByRole(pathRole)
    if (token && !userStore.isTokenExpired(token)) {
      const redirectTarget = String(to.query?.redirect || '').trim()
      if (redirectTarget) {
        return redirectTarget
      }
      return ROLE_DASHBOARD[pathRole]
    }
    return true
  }

  if (!requiresAuth) {
    return true
  }

  const token = readTokenFromStorageByRole(pathRole)
  if (!token || userStore.isTokenExpired(token)) {
    userStore.logout(pathRole)

    const loggedRole = findLoggedRole(userStore)
    if (loggedRole && loggedRole !== pathRole) {
      return ROLE_DASHBOARD[loggedRole]
    }

    return buildLoginPath(pathRole, to.fullPath)
  }

  const requiredRoles = extractRequiredRoles(to, pathRole)
  const userInfo = userStore.getUserInfoByRole(pathRole)
  const actualRole = normalizeRole(userInfo?.role || pathRole)

  if (!requiredRoles.includes(actualRole)) {
    const fallback = ROLE_DASHBOARD[actualRole]
    if (fallback && fallback !== to.path) {
      return fallback
    }
    return buildForbiddenLocation(to, requiredRoles, actualRole)
  }

  return true
})

export default router
