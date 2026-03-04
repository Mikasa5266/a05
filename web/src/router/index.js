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
        path: 'interview/select', // Keep for compatibility if needed, or redirect
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
        path: 'settings',
        name: 'Settings',
        component: Settings
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
