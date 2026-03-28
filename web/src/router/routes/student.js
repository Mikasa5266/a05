import Layout from '../../components/layout/Layout.vue'

const Home = () => import('../../views/Home.vue')
const ResumeMatching = () => import('../../views/ResumeMatching.vue')
const MockInterview = () => import('../../views/MockInterview.vue')
const InterviewModeSelect = () => import('../../views/InterviewModeSelect.vue')
const GrowthCenter = () => import('../../views/GrowthCenter.vue')
const History = () => import('../../views/History.vue')
const Report = () => import('../../views/Report.vue')
const Settings = () => import('../../views/Settings.vue')
const Community = () => import('../../views/Community.vue')
const CommunityPostDetail = () => import('../../views/CommunityPostDetail.vue')
const LiveInterviewRoom = () => import('../../views/LiveInterviewRoom.vue')

const roleMeta = {
  requiresAuth: true,
  roles: ['student']
}

export const studentRoutes = [
  {
    path: '/student',
    component: Layout,
    meta: roleMeta,
    children: [
      {
        path: '',
        redirect: '/student/dashboard',
        meta: roleMeta
      },
      {
        path: 'dashboard',
        name: 'StudentDashboard',
        component: Home,
        meta: roleMeta
      },
      {
        path: 'resume',
        name: 'ResumeMatching',
        component: ResumeMatching,
        meta: roleMeta
      },
      {
        path: 'interview',
        redirect: '/interview/mode-select',
        meta: roleMeta
      },
      {
        path: 'live-interview',
        redirect: (to) => {
          const invitationId = String(to.query?.invitation_id || '').trim() || '0'
          return `/interview/live/room?invitation_id=${invitationId}`
        },
        meta: roleMeta
      },
      {
        path: 'growth',
        name: 'GrowthCenter',
        component: GrowthCenter,
        meta: roleMeta
      },
      {
        path: 'history',
        name: 'History',
        component: History,
        meta: roleMeta
      },
      {
        path: 'report/:id',
        name: 'Report',
        component: Report,
        meta: roleMeta
      },
      {
        path: 'community',
        name: 'Community',
        component: Community,
        meta: roleMeta
      },
      {
        path: 'community/posts/:id',
        name: 'CommunityPostDetail',
        component: CommunityPostDetail,
        meta: roleMeta
      },
      {
        path: 'settings',
        name: 'StudentSettings',
        component: Settings,
        meta: roleMeta
      }
    ]
  },
  {
    path: '/interview',
    component: Layout,
    meta: roleMeta,
    children: [
      {
        path: 'mode-select',
        name: 'InterviewModeSelect',
        component: InterviewModeSelect,
        meta: roleMeta
      },
      {
        path: 'standard/setup',
        name: 'MockInterview',
        component: MockInterview,
        meta: roleMeta
      },
      {
        path: 'algorithm/setup',
        name: 'AlgorithmInterviewSetup',
        component: MockInterview,
        beforeEnter: (to) => {
          const mode = String(to.query?.mode || '').trim()
          const style = String(to.query?.style || '').trim()
          if (mode === 'technical' && style === 'algorithm') {
            return true
          }
          return {
            path: to.path,
            query: {
              ...to.query,
              mode: mode || 'technical',
              style: style || 'algorithm'
            }
          }
        },
        meta: roleMeta
      },
      {
        path: 'live/room/:id?',
        name: 'StudentLiveInterviewRoom',
        component: LiveInterviewRoom,
        meta: roleMeta
      }
    ]
  }
]
