import { defineAsyncComponent } from "vue";

const Layout = defineAsyncComponent(
  () => import("../../components/layout/Layout.vue"),
);

const EnterpriseDashboard = defineAsyncComponent(
  () => import("../../views/enterprise/EnterpriseDashboard.vue"),
);
const TalentPool = defineAsyncComponent(
  () => import("../../views/enterprise/TalentPool.vue"),
);
const JobManagement = defineAsyncComponent(
  () => import("../../views/enterprise/JobManagement.vue"),
);
const HRPanel = defineAsyncComponent(
  () => import("../../views/enterprise/HRPanel.vue"),
);
const InterviewWorkbench = defineAsyncComponent(
  () => import("../../views/enterprise/InterviewWorkbench.vue"),
);
const Analytics = defineAsyncComponent(
  () => import("../../views/enterprise/Analytics.vue"),
);
const Standards = defineAsyncComponent(
  () => import("../../views/enterprise/Standards.vue"),
);
const LiveInterviewRoomOneOnOne = defineAsyncComponent(
  () => import("../../views/LiveInterviewRoomOneOnOne.vue"),
);
const LiveInterviewRoomGroup = defineAsyncComponent(
  () => import("../../views/LiveInterviewRoomGroup.vue"),
);
const Settings = defineAsyncComponent(() => import("../../views/Settings.vue"));

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
        path: "live-interview",
        redirect: (to) => {
          const invitationId = String(to.query?.invitation_id || "").trim();
          if (!invitationId) {
            return "/enterprise/interview-workbench";
          }
          const invitationCode = String(to.query?.invitation_code || "").trim();
          const isGroup = String(to.query?.group_mode || "").trim() === "1";
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
