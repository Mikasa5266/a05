import Layout from '../../components/layout/Layout.vue'

import EnterpriseDashboard from '../../views/enterprise/EnterpriseDashboard.vue'
import TalentPool from '../../views/enterprise/TalentPool.vue'
import JobManagement from '../../views/enterprise/JobManagement.vue'
import HRPanel from '../../views/enterprise/HRPanel.vue'
import Analytics from '../../views/enterprise/Analytics.vue'
import Standards from '../../views/enterprise/Standards.vue'
import LiveInterviewRoom from '../../views/LiveInterviewRoom.vue'
import Settings from '../../views/Settings.vue'

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
