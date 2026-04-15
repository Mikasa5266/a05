const Layout = () => import("../../components/layout/Layout.vue");

const EnterpriseDashboard = () =>
  import("../../views/enterprise/EnterpriseDashboard.vue");
const TalentPool = () => import("../../views/enterprise/TalentPool.vue");
const JobManagement = () => import("../../views/enterprise/JobManagement.vue");
const HRPanel = () => import("../../views/enterprise/HRPanel.vue");
const InterviewWorkbench = () =>
  import("../../views/enterprise/InterviewWorkbench.vue");
const GroupInterviewWorkbench = () =>
  import("../../views/enterprise/GroupInterviewWorkbench.vue");
const Analytics = () => import("../../views/enterprise/Analytics.vue");
const Standards = () => import("../../views/enterprise/Standards.vue");
const LiveInterviewRoomOneOnOne = () =>
  import("../../views/LiveInterviewRoomOneOnOne.vue");
const LiveInterviewRoomGroup = () =>
  import("../../views/LiveInterviewRoomGroup.vue");
const Settings = () => import("../../views/Settings.vue");

const roleMeta = {
  requiresAuth: true,
  roles: ["enterprise"],
};

export const enterpriseRoutes = [
  {
    path: "/enterprise",
    component: Layout,
    meta: roleMeta,
    children: [
      {
        path: "",
        redirect: "/enterprise/dashboard",
        meta: roleMeta,
      },
      {
        path: "dashboard",
        name: "EnterpriseDashboard",
        component: EnterpriseDashboard,
        meta: roleMeta,
      },
      {
        path: "talent",
        name: "TalentPool",
        component: TalentPool,
        meta: roleMeta,
      },
      {
        path: "jobs",
        name: "JobManagement",
        component: JobManagement,
        meta: roleMeta,
      },
      {
        path: "hr-panel",
        name: "HRPanel",
        component: HRPanel,
        meta: roleMeta,
      },
      {
        path: "interview-workbench",
        name: "EnterpriseInterviewWorkbench",
        component: InterviewWorkbench,
        meta: roleMeta,
      },
      {
        path: "group-interview/workbench",
        name: "EnterpriseGroupInterviewWorkbench",
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
              ? "/enterprise/group-interview/workbench"
              : "/enterprise/interview-workbench";
          }
          const invitationCode = String(to.query?.invitation_code || "").trim();
          const targetPath = isGroup
            ? `/enterprise/live-interview/group/${encodeURIComponent(invitationId)}`
            : `/enterprise/live-interview/1v1/${encodeURIComponent(invitationId)}`;
          if (!invitationCode) {
            return targetPath;
          }
          return `${targetPath}?invitation_code=${encodeURIComponent(invitationCode)}`;
        },
        meta: roleMeta,
      },
      {
        path: "live-interview/1v1/:id",
        name: "EnterpriseLiveInterviewOneOnOne",
        component: LiveInterviewRoomOneOnOne,
        meta: roleMeta,
      },
      {
        path: "live-interview/group/:id",
        name: "EnterpriseLiveInterviewGroup",
        component: LiveInterviewRoomGroup,
        meta: roleMeta,
      },
      {
        path: "analytics",
        name: "Analytics",
        component: Analytics,
        meta: roleMeta,
      },
      {
        path: "standards",
        name: "Standards",
        component: Standards,
        meta: roleMeta,
      },
      {
        path: "settings",
        name: "EnterpriseSettings",
        component: Settings,
        meta: roleMeta,
      },
    ],
  },
];
