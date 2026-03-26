import Layout from '../../components/layout/Layout.vue'

const UniversityDashboard = () => import('../../views/university/UniversityDashboard.vue')
const StudentTracking = () => import('../../views/university/StudentTracking.vue')
const SupportSystem = () => import('../../views/university/SupportSystem.vue')
const Courses = () => import('../../views/university/Courses.vue')
const Employment = () => import('../../views/university/Employment.vue')
const TalentPush = () => import('../../views/university/TalentPush.vue')
const LiveInterviewRoom = () => import('../../views/LiveInterviewRoom.vue')
const Settings = () => import('../../views/Settings.vue')

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
