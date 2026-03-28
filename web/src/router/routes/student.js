import { defineAsyncComponent } from 'vue'

const Layout = defineAsyncComponent(() => import('../../components/layout/Layout.vue'))

const Home = defineAsyncComponent(() => import('../../views/Home.vue'))
const ResumeMatching = defineAsyncComponent(() => import('../../views/ResumeMatching.vue'))
const Interview = defineAsyncComponent(() => import('../../views/Interview.vue'))
const MockInterview = defineAsyncComponent(() => import('../../views/MockInterview.vue'))
const InterviewModeSelect = defineAsyncComponent(() => import('../../views/InterviewModeSelect.vue'))
const GrowthCenter = defineAsyncComponent(() => import('../../views/GrowthCenter.vue'))
const History = defineAsyncComponent(() => import('../../views/History.vue'))
const Report = defineAsyncComponent(() => import('../../views/Report.vue'))
const Settings = defineAsyncComponent(() => import('../../views/Settings.vue'))
const Community = defineAsyncComponent(() => import('../../views/Community.vue'))
const CommunityPostDetail = defineAsyncComponent(() => import('../../views/CommunityPostDetail.vue'))
const StudentLiveInterviewWorkbench = defineAsyncComponent(() => import('../../views/student/LiveInterviewWorkbench.vue'))
const LiveInterviewRoom = defineAsyncComponent(() => import('../../views/LiveInterviewRoom.vue'))

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
          const invitationId = String(to.query?.invitation_id || '').trim()
          if (!invitationId) {
            return '/interview/live/workbench'
          }
          const invitationCode = String(to.query?.invitation_code || '').trim()
          if (!invitationCode) {
            return `/interview/live/room?invitation_id=${invitationId}`
          }
          return `/interview/live/room?invitation_id=${invitationId}&invitation_code=${encodeURIComponent(invitationCode)}`
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
        path: 'video',
        name: 'InterviewVideoMode',
        component: Interview,
        beforeEnter: (to) => {
          const normalized = {
            ...to.query,
            mode: String(to.query?.mode || 'technical'),
            style: String(to.query?.style || 'gentle'),
            interviewMode: String(to.query?.interviewMode || 'ai'),
            presentationMode: String(to.query?.presentationMode || 'video_avatar')
          }

          const sameMode = String(to.query?.mode || '') === normalized.mode
          const sameStyle = String(to.query?.style || '') === normalized.style
          const sameInterviewMode = String(to.query?.interviewMode || '') === normalized.interviewMode
          const samePresentation = String(to.query?.presentationMode || '') === normalized.presentationMode

          if (sameMode && sameStyle && sameInterviewMode && samePresentation) {
            return true
          }

          return {
            path: to.path,
            query: normalized
          }
        },
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
        path: 'live/workbench',
        name: 'StudentLiveInterviewWorkbench',
        component: StudentLiveInterviewWorkbench,
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
