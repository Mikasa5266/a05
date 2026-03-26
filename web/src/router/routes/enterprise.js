import Layout from '../../components/layout/Layout.vue'

const EnterpriseDashboard = () => import('../../views/enterprise/EnterpriseDashboard.vue')
const TalentPool = () => import('../../views/enterprise/TalentPool.vue')
const JobManagement = () => import('../../views/enterprise/JobManagement.vue')
const HRPanel = () => import('../../views/enterprise/HRPanel.vue')
const Analytics = () => import('../../views/enterprise/Analytics.vue')
const Standards = () => import('../../views/enterprise/Standards.vue')
const LiveInterviewRoom = () => import('../../views/LiveInterviewRoom.vue')
const Settings = () => import('../../views/Settings.vue')

const roleMeta = {
  requiresAuth: true,
  roles: ['enterprise']
}

export const enterpriseRoutes = [
  {
    path: '/enterprise',
    component: Layout,
    meta: roleMeta,
    children: [
      {
        path: '',
        redirect: '/enterprise/dashboard',
        meta: roleMeta
      },
      {
        path: 'dashboard',
        name: 'EnterpriseDashboard',
        component: EnterpriseDashboard,
        meta: roleMeta
      },
      {
        path: 'talent',
        name: 'TalentPool',
        component: TalentPool,
        meta: roleMeta
      },
      {
        path: 'jobs',
        name: 'JobManagement',
        component: JobManagement,
        meta: roleMeta
      },
      {
        path: 'hr-panel',
        name: 'HRPanel',
        component: HRPanel,
        meta: roleMeta
      },
      {
        path: 'live-interview',
        name: 'EnterpriseLiveInterview',
        component: LiveInterviewRoom,
        meta: roleMeta
      },
      {
        path: 'analytics',
        name: 'Analytics',
        component: Analytics,
        meta: roleMeta
      },
      {
        path: 'standards',
        name: 'Standards',
        component: Standards,
        meta: roleMeta
      },
      {
        path: 'settings',
        name: 'EnterpriseSettings',
        component: Settings,
        meta: roleMeta
      }
    ]
  }
]
