const Layout = () => import("../../components/layout/Layout.vue");

const UniversityDashboard = () =>
  import("../../views/university/UniversityDashboard.vue");
const StudentTracking = () =>
  import("../../views/university/StudentTracking.vue");
const SupportSystem = () => import("../../views/university/SupportSystem.vue");
const Courses = () => import("../../views/university/Courses.vue");
const InterviewWorkbench = () =>
  import("../../views/university/InterviewWorkbench.vue");
const GroupInterviewWorkbench = () =>
  import("../../views/university/GroupInterviewWorkbench.vue");
const Employment = () => import("../../views/university/Employment.vue");
const TalentPush = () => import("../../views/university/TalentPush.vue");
const LiveInterviewRoomOneOnOne = () =>
  import("../../views/LiveInterviewRoomOneOnOne.vue");
const LiveInterviewRoomGroup = () =>
  import("../../views/LiveInterviewRoomGroup.vue");
const Settings = () => import("../../views/Settings.vue");

const roleMeta = {
  requiresAuth: true,
  roles: ["university"],
};

export const universityRoutes = [
  {
    path: "/university",
    component: Layout,
    meta: roleMeta,
    children: [
      {
        path: "",
        redirect: "/university/dashboard",
        meta: roleMeta,
      },
      {
        path: "dashboard",
        name: "UniversityDashboard",
        component: UniversityDashboard,
        meta: roleMeta,
      },
      {
        path: "tracking",
        name: "StudentTracking",
        component: StudentTracking,
        meta: roleMeta,
      },
      {
        path: "support",
        name: "SupportSystem",
        component: SupportSystem,
        meta: roleMeta,
      },
      {
        path: "courses",
        name: "Courses",
        component: Courses,
        meta: roleMeta,
      },
      {
        path: "employment",
        name: "Employment",
        component: Employment,
        meta: roleMeta,
      },
      {
        path: "talent-push",
        name: "TalentPush",
        component: TalentPush,
        meta: roleMeta,
      },
      {
        path: "interview-workbench",
        name: "UniversityInterviewWorkbench",
        component: InterviewWorkbench,
        meta: roleMeta,
      },
      {
        path: "group-interview/workbench",
        name: "UniversityGroupInterviewWorkbench",
        component: GroupInterviewWorkbench,
        meta: roleMeta,
      },
      {
        path: "live-interview",
        redirect: (to) => {
          const invitationId = String(to.query?.invitation_id || "").trim();
          const isGroup = String(to.query?.group_mode || "").trim() === "1";
          if (!invitationId) {
            return isGroup
              ? "/university/group-interview/workbench"
              : "/university/interview-workbench";
          }
          const invitationCode = String(to.query?.invitation_code || "").trim();
          const targetPath = isGroup
            ? `/university/live-interview/group/${encodeURIComponent(invitationId)}`
            : `/university/live-interview/1v1/${encodeURIComponent(invitationId)}`;
          if (!invitationCode) {
            return targetPath;
          }
          return `${targetPath}?invitation_code=${encodeURIComponent(invitationCode)}`;
        },
        meta: roleMeta,
      },
      {
        path: "live-interview/1v1/:id",
        name: "UniversityLiveInterviewOneOnOne",
        component: LiveInterviewRoomOneOnOne,
        meta: roleMeta,
      },
      {
        path: "live-interview/group/:id",
        name: "UniversityLiveInterviewGroup",
        component: LiveInterviewRoomGroup,
        meta: roleMeta,
      },
      {
        path: "settings",
        name: "UniversitySettings",
        component: Settings,
        meta: roleMeta,
      },
    ],
  },
];
