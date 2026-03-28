import { defineAsyncComponent } from 'vue'

const Layout = defineAsyncComponent(() => import('../../components/layout/Layout.vue'))

const UniversityDashboard = defineAsyncComponent(() => import('../../views/university/UniversityDashboard.vue'))
const StudentTracking = defineAsyncComponent(() => import('../../views/university/StudentTracking.vue'))
const SupportSystem = defineAsyncComponent(() => import('../../views/university/SupportSystem.vue'))
const Courses = defineAsyncComponent(() => import('../../views/university/Courses.vue'))
const InterviewWorkbench = defineAsyncComponent(() => import('../../views/university/InterviewWorkbench.vue'))
const Employment = defineAsyncComponent(() => import('../../views/university/Employment.vue'))
const TalentPush = defineAsyncComponent(() => import('../../views/university/TalentPush.vue'))
const LiveInterviewRoom = defineAsyncComponent(() => import('../../views/LiveInterviewRoom.vue'))
const Settings = defineAsyncComponent(() => import('../../views/Settings.vue'))

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
        path: 'interview-workbench',
        name: 'UniversityInterviewWorkbench',
        component: InterviewWorkbench,
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
