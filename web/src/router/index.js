import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'
import Login from '../views/Login.vue'
import PortalSelect from '../views/PortalSelect.vue'
import Layout from '../components/layout/Layout.vue'

// Student Views
import Home from '../views/Home.vue'
import ResumeMatching from '../views/ResumeMatching.vue'
import MockInterview from '../views/MockInterview.vue'
import GrowthCenter from '../views/GrowthCenter.vue'
import History from '../views/History.vue'
import Report from '../views/Report.vue'
import Settings from '../views/Settings.vue'
import Community from '../views/Community.vue'
import CommunityPostDetail from '../views/CommunityPostDetail.vue'

// Enterprise Views
import EnterpriseDashboard from '../views/enterprise/EnterpriseDashboard.vue'
import TalentPool from '../views/enterprise/TalentPool.vue'
import JobManagement from '../views/enterprise/JobManagement.vue'
import HRPanel from '../views/enterprise/HRPanel.vue'
import Analytics from '../views/enterprise/Analytics.vue'
import Standards from '../views/enterprise/Standards.vue'

// University Views
import UniversityDashboard from '../views/university/UniversityDashboard.vue'
import StudentTracking from '../views/university/StudentTracking.vue'
import SupportSystem from '../views/university/SupportSystem.vue'
import Courses from '../views/university/Courses.vue'
import Employment from '../views/university/Employment.vue'
import TalentPush from '../views/university/TalentPush.vue'

const routes = [
  {
    path: '/',
    name: 'PortalSelect',
    component: PortalSelect
  },
  // ====== Student Login ======
  {
    path: '/student/login',
    name: 'StudentLogin',
    component: Login
  },
  // ====== Enterprise Login ======
  {
    path: '/enterprise/login',
    name: 'EnterpriseLogin',
    component: Login
  },
  // ====== University Login ======
  {
    path: '/university/login',
    name: 'UniversityLogin',
    component: Login
  },
  // ====== Student Portal ======
  {
    path: '/student',
    component: Layout,
    meta: { requiresAuth: true, role: 'student' },
    children: [
      {
        path: '',
        redirect: '/student/dashboard'
      },
      {
        path: 'dashboard',
        name: 'StudentDashboard',
        component: Home
      },
      {
        path: 'resume',
        name: 'ResumeMatching',
        component: ResumeMatching
      },
      {
        path: 'interview',
        name: 'MockInterview',
        component: MockInterview
      },
      {
        path: 'growth',
        name: 'GrowthCenter',
        component: GrowthCenter
      },
      {
        path: 'history',
        name: 'History',
        component: History
      },
      {
        path: 'report/:id',
        name: 'Report',
        component: Report
      },
      {
        path: 'community',
        name: 'Community',
        component: Community
      },
      {
        path: 'community/posts/:id',
        name: 'CommunityPostDetail',
        component: CommunityPostDetail
      },
      {
        path: 'settings',
        name: 'StudentSettings',
        component: Settings
      }
    ]
  },
  // ====== Enterprise Portal ======
  {
    path: '/enterprise',
    component: Layout,
    meta: { requiresAuth: true, role: 'enterprise' },
    children: [
      {
        path: '',
        redirect: '/enterprise/dashboard'
      },
      {
        path: 'dashboard',
        name: 'EnterpriseDashboard',
        component: EnterpriseDashboard
      },
      {
        path: 'talent',
        name: 'TalentPool',
        component: TalentPool
      },
      {
        path: 'jobs',
        name: 'JobManagement',
        component: JobManagement
      },
      {
        path: 'hr-panel',
        name: 'HRPanel',
        component: HRPanel
      },
      {
        path: 'analytics',
        name: 'Analytics',
        component: Analytics
      },
      {
        path: 'standards',
        name: 'Standards',
        component: Standards
      },
      {
        path: 'settings',
        name: 'EnterpriseSettings',
        component: Settings
      }
    ]
  },
  // ====== University Portal ======
  {
    path: '/university',
    component: Layout,
    meta: { requiresAuth: true, role: 'university' },
    children: [
      {
        path: '',
        redirect: '/university/dashboard'
      },
      {
        path: 'dashboard',
        name: 'UniversityDashboard',
        component: UniversityDashboard
      },
      {
        path: 'tracking',
        name: 'StudentTracking',
        component: StudentTracking
      },
      {
        path: 'support',
        name: 'SupportSystem',
        component: SupportSystem
      },
      {
        path: 'courses',
        name: 'Courses',
        component: Courses
      },
      {
        path: 'employment',
        name: 'Employment',
        component: Employment
      },
      {
        path: 'talent-push',
        name: 'TalentPush',
        component: TalentPush
      },
      {
        path: 'settings',
        name: 'UniversitySettings',
        component: Settings
      }
    ]
  },
  // Legacy redirects for old paths
  { path: '/login', redirect: '/' },
  { path: '/dashboard', redirect: '/student/dashboard' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()

  // If route requires auth
  if (to.matched.some(record => record.meta.requiresAuth)) {
    if (!userStore.token) {
      // Redirect to the appropriate login page based on route
      const portal = to.path.startsWith('/enterprise') ? 'enterprise'
                   : to.path.startsWith('/university') ? 'university'
                   : 'student'
      next(`/${portal}/login`)
      return
    }

    // Check role if specified
    const requiredRole = to.matched.find(record => record.meta.role)?.meta.role
    if (requiredRole && userStore.userInfo?.role && userStore.userInfo.role !== requiredRole) {
      // Add a final level redirect for unknown user roles
      const userRole = userStore.userInfo.role
      const targetPath = ['student', 'enterprise', 'university'].includes(userRole)
        ? `/${userRole}/dashboard`
        : '/student/dashboard'
      
      if (to.path !== targetPath) {
        next(targetPath)
      } else {
        // If we are already at the target path but role mismatch persists,
        // it means the user's role is invalid for this route.
        // To avoid loop, we just let it pass (or logout).
        // Here we pass, assuming the component might handle it or show 403.
        next()
      }
      return
    }
  }

  // If already logged in and visiting a login page, redirect to their portal
  if (to.path.endsWith('/login') || (to.path === '/' && userStore.token && userStore.userInfo?.role)) {
    // Safety check: ensure userInfo exists
    if (!userStore.userInfo) {
      userStore.logout()
      next()
      return
    }
    const role = userStore.userInfo.role
    const targetPath = ['student', 'enterprise', 'university'].includes(role)
      ? `/${role}/dashboard`
      : '/student/dashboard'
      
    if (to.path !== targetPath) {
      next(targetPath)
    } else {
      next()
    }
    return
  }

  next()
})

export default router
