import Layout from '../../components/layout/Layout.vue'

import UniversityDashboard from '../../views/university/UniversityDashboard.vue'
import StudentTracking from '../../views/university/StudentTracking.vue'
import SupportSystem from '../../views/university/SupportSystem.vue'
import Courses from '../../views/university/Courses.vue'
import Employment from '../../views/university/Employment.vue'
import TalentPush from '../../views/university/TalentPush.vue'
import LiveInterviewRoom from '../../views/LiveInterviewRoom.vue'
import Settings from '../../views/Settings.vue'

const roleMeta = {
  requiresAuth: true,
  roles: ['university']
}

export const universityRoutes = [
  {
    path: '/university',
    component: Layout,
    meta: roleMeta,
    children: [
      {
        path: '',
        redirect: '/university/dashboard',
        meta: roleMeta
      },
      {
        path: 'dashboard',
        name: 'UniversityDashboard',
        component: UniversityDashboard,
        meta: roleMeta
      },
      {
        path: 'tracking',
        name: 'StudentTracking',
        component: StudentTracking,
        meta: roleMeta
      },
      {
        path: 'support',
        name: 'SupportSystem',
        component: SupportSystem,
        meta: roleMeta
      },
      {
        path: 'courses',
        name: 'Courses',
        component: Courses,
        meta: roleMeta
      },
      {
        path: 'employment',
        name: 'Employment',
        component: Employment,
        meta: roleMeta
      },
      {
        path: 'talent-push',
        name: 'TalentPush',
        component: TalentPush,
        meta: roleMeta
      },
      {
        path: 'live-interview',
        name: 'UniversityLiveInterview',
        component: LiveInterviewRoom,
        meta: roleMeta
      },
      {
        path: 'settings',
        name: 'UniversitySettings',
        component: Settings,
        meta: roleMeta
      }
    ]
  }
]
