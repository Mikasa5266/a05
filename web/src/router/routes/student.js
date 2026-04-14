import { defineAsyncComponent } from "vue";

const Layout = defineAsyncComponent(
  () => import("../../components/layout/Layout.vue"),
);
const PracticeModeLayout = defineAsyncComponent(
  () => import("../../components/layout/PracticeModeLayout.vue"),
);

const Home = defineAsyncComponent(() => import("../../views/Home.vue"));
const Interview = defineAsyncComponent(
  () => import("../../views/Interview.vue"),
);
const MockInterview = defineAsyncComponent(
  () => import("../../views/MockInterview.vue"),
);
const BlindBoxMode = defineAsyncComponent(
  () => import("../../views/BlindBoxMode.vue"),
);
const InterviewModeSelect = defineAsyncComponent(
  () => import("../../views/InterviewModeSelect.vue"),
);
const StudentLiveInterviewWorkbench = defineAsyncComponent(
  () => import("../../views/student/LiveInterviewWorkbench.vue"),
);
const LiveInterviewRoomOneOnOne = defineAsyncComponent(
  () => import("../../views/LiveInterviewRoomOneOnOne.vue"),
);
const LiveInterviewRoomGroup = defineAsyncComponent(
  () => import("../../views/LiveInterviewRoomGroup.vue"),
);
const ResumeCenter = defineAsyncComponent(
  () => import("../../views/student/ResumeCenter.vue"),
);
const History = defineAsyncComponent(() => import("../../views/History.vue"));
const Report = defineAsyncComponent(() => import("../../views/Report.vue"));
const Settings = defineAsyncComponent(() => import("../../views/Settings.vue"));
const Community = defineAsyncComponent(
  () => import("../../views/Community.vue"),
);
const CommunityPostDetail = defineAsyncComponent(
  () => import("../../views/CommunityPostDetail.vue"),
);
const PracticeModeIndex = defineAsyncComponent(
  () => import("../../views/student/PracticeMode/Index.vue"),
);

const roleMeta = {
  requiresAuth: true,
  roles: ["student"],
};

export const studentRoutes = [
  {
    path: "/student/practice-mode",
    component: PracticeModeLayout,
    meta: roleMeta,
    children: [
      {
        path: "",
        name: "StudentPracticeMode",
        component: PracticeModeIndex,
        meta: roleMeta,
      },
    ],
  },
  {
    path: "/student",
    component: Layout,
    meta: roleMeta,
    children: [
      {
        path: "",
        redirect: "/student/dashboard",
        meta: roleMeta,
      },
      {
        path: "dashboard",
        name: "StudentDashboard",
        component: Home,
        meta: roleMeta,
      },
      {
        path: "interview",
        redirect: "/interview/mode-select",
        meta: roleMeta,
      },
      {
        path: "live-interview",
        redirect: (to) => {
          const invitationId = String(to.query?.invitation_id || "").trim();
          if (!invitationId) {
            return "/interview/live/workbench";
          }
          const invitationCode = String(to.query?.invitation_code || "").trim();
          const isGroup = String(to.query?.group_mode || "").trim() === "1";
          const targetPath = isGroup
            ? `/interview/live/group/${encodeURIComponent(invitationId)}`
            : `/interview/live/1v1/${encodeURIComponent(invitationId)}`;
          if (!invitationCode) {
            return targetPath;
          }
          return `${targetPath}?invitation_code=${encodeURIComponent(invitationCode)}`;
        },
        meta: roleMeta,
      },
      // {
      //   path: "growth",
      //   name: "GrowthCenter",
      //   component: GrowthCenter,
      //   meta: roleMeta,
      // },
      {
        path: "resume",
        name: "ResumeCenter",
        component: ResumeCenter,
        meta: roleMeta,
      },
      {
        path: "history",
        name: "History",
        component: History,
        meta: roleMeta,
      },
      {
        path: "report/:id",
        name: "Report",
        component: Report,
        meta: roleMeta,
      },
      {
        path: "community",
        name: "Community",
        component: Community,
        meta: roleMeta,
      },
      {
        path: "community/posts/:id",
        name: "CommunityPostDetail",
        component: CommunityPostDetail,
        meta: roleMeta,
      },
      {
        path: "settings",
        name: "StudentSettings",
        component: Settings,
        meta: roleMeta,
      },
    ],
  },
  {
    path: "/interview",
    component: Layout,
    meta: roleMeta,
    children: [
      {
        path: "mode-select",
        name: "InterviewModeSelect",
        component: InterviewModeSelect,
        meta: roleMeta,
      },
      {
        path: "blindbox",
        name: "BlindBoxMode",
        component: BlindBoxMode,
        meta: roleMeta,
      },
      {
        path: "standard/setup",
        name: "MockInterview",
        component: MockInterview,
        meta: roleMeta,
      },
      {
        path: "video",
        name: "InterviewVideoMode",
        component: Interview,
        beforeEnter: (to) => {
          const normalized = {
            ...to.query,
            mode: String(to.query?.mode || "technical"),
            style: String(to.query?.style || "gentle"),
            interviewMode: String(to.query?.interviewMode || "ai"),
            presentationMode: String(
              to.query?.presentationMode || "video_avatar",
            ),
          };

          const sameMode = String(to.query?.mode || "") === normalized.mode;
          const sameStyle = String(to.query?.style || "") === normalized.style;
          const sameInterviewMode =
            String(to.query?.interviewMode || "") === normalized.interviewMode;
          const samePresentation =
            String(to.query?.presentationMode || "") ===
            normalized.presentationMode;

          if (sameMode && sameStyle && sameInterviewMode && samePresentation) {
            return true;
          }

          return {
            path: to.path,
            query: normalized,
          };
        },
        meta: roleMeta,
      },
      {
        path: "algorithm/setup",
        name: "AlgorithmInterviewSetup",
        component: MockInterview,
        beforeEnter: (to) => {
          const mode = String(to.query?.mode || "").trim();
          const style = String(to.query?.style || "").trim();
          if (mode === "technical" && style === "algorithm") {
            return true;
          }
          return {
            path: to.path,
            query: {
              ...to.query,
              mode: mode || "technical",
              style: style || "algorithm",
            },
          };
        },
        meta: roleMeta,
      },
      {
        path: "live/workbench",
        name: "StudentLiveInterviewWorkbench",
        component: StudentLiveInterviewWorkbench,
        meta: roleMeta,
      },
      {
        path: "live/1v1/:id",
        name: "StudentLiveInterviewRoomOneOnOne",
        component: LiveInterviewRoomOneOnOne,
        meta: roleMeta,
      },
      {
        path: "live/group/:id",
        name: "StudentLiveInterviewRoomGroup",
        component: LiveInterviewRoomGroup,
        meta: roleMeta,
      },
      {
        path: "live/room",
        redirect: (to) => {
          const invitationId = String(to.query?.invitation_id || "").trim();
          if (!invitationId) {
            return "/interview/live/workbench";
          }
          const invitationCode = String(to.query?.invitation_code || "").trim();
          const isGroup = String(to.query?.group_mode || "").trim() === "1";
          const targetPath = isGroup
            ? `/interview/live/group/${encodeURIComponent(invitationId)}`
            : `/interview/live/1v1/${encodeURIComponent(invitationId)}`;
          if (!invitationCode) {
            return targetPath;
          }
          return `${targetPath}?invitation_code=${encodeURIComponent(invitationCode)}`;
        },
        meta: roleMeta,
      },
    ],
  },
];
