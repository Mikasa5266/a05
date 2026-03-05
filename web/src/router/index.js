import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'
import Login from '../views/Login.vue'
import Layout from '../components/layout/Layout.vue'
import Home from '../views/Home.vue'
import ResumeMatching from '../views/ResumeMatching.vue'
import MockInterview from '../views/MockInterview.vue'
import GrowthCenter from '../views/GrowthCenter.vue'
import History from '../views/History.vue'
import Report from '../views/Report.vue'
import Settings from '../views/Settings.vue'
import Community from '../views/Community.vue'

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
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: '/dashboard'
      },
      {
        path: 'dashboard',
        name: 'Dashboard',
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
        path: 'interview/select',
        redirect: '/interview'
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
        path: 'settings',
        name: 'Settings',
        component: Settings
      },
      // Enterprise Portal
      {
        path: 'enterprise',
        name: 'EnterpriseDashboard',
        component: EnterpriseDashboard
      },
      {
        path: 'enterprise/talent',
        name: 'TalentPool',
        component: TalentPool
      },
      {
        path: 'enterprise/jobs',
        name: 'JobManagement',
        component: JobManagement
      },
      {
        path: 'enterprise/hr-panel',
        name: 'HRPanel',
        component: HRPanel
      },
      {
        path: 'enterprise/analytics',
        name: 'Analytics',
        component: Analytics
      },
      {
        path: 'enterprise/standards',
        name: 'Standards',
        component: Standards
      },
      // University Portal
      {
        path: 'university',
        name: 'UniversityDashboard',
        component: UniversityDashboard
      },
      {
        path: 'university/tracking',
        name: 'StudentTracking',
        component: StudentTracking
      },
      {
        path: 'university/support',
        name: 'SupportSystem',
        component: SupportSystem
      },
      {
        path: 'university/courses',
        name: 'Courses',
        component: Courses
      },
      {
        path: 'university/employment',
        name: 'Employment',
        component: Employment
      },
      {
        path: 'university/talent-push',
        name: 'TalentPush',
        component: TalentPush
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  // Simple auth check
  if (to.meta.requiresAuth && !userStore.token && to.path !== '/login') {
    next('/login')
  } else {
    next()
  }
})

export default router
