import {
  BarChart3,
  BookOpen,
  Briefcase,
  Building2,
  ClipboardList,
  Database,
  FileText,
  GraduationCap,
  LayoutDashboard,
  Send,
  Shield,
  Target,
  TrendingUp,
  User,
  UserCheck,
  Users,
  Video,
} from "lucide-vue-next";

export const getPortalFromPath = (path = "") => {
  if (path.startsWith("/enterprise")) return "enterprise";
  if (path.startsWith("/university")) return "university";
  return "student";
};

export const portalBrandMap = {
  student: {
    title: "\u667a\u8058 AI",
    label: "\u5b66\u751f\u7aef",
    icon: User,
    logoBg: "bg-indigo-600 shadow-indigo-200",
    logoText: "text-indigo-600",
    badgeClass: "bg-indigo-50 text-indigo-600 border border-indigo-100",
    activeText: "text-indigo-600",
    activeBg: "bg-indigo-50",
  },
  enterprise: {
    title: "\u667a\u8058 AI",
    label: "\u4f01\u4e1a\u7aef",
    icon: Building2,
    logoBg: "bg-emerald-600 shadow-emerald-200",
    logoText: "text-emerald-600",
    badgeClass: "bg-emerald-50 text-emerald-600 border border-emerald-100",
    activeText: "text-emerald-600",
    activeBg: "bg-emerald-50",
  },
  university: {
    title: "\u667a\u8058 AI",
    label: "\u9ad8\u6821\u7aef",
    icon: GraduationCap,
    logoBg: "bg-amber-600 shadow-amber-200",
    logoText: "text-amber-600",
    badgeClass: "bg-amber-50 text-amber-600 border border-amber-100",
    activeText: "text-amber-600",
    activeBg: "bg-amber-50",
  },
};

export const portalNavMap = {
  student: [
    {
      name: "\u4e2a\u4eba\u9996\u9875",
      href: "/student/dashboard",
      icon: LayoutDashboard,
    },
    {
      name: "\u6a21\u62df\u9762\u8bd5",
      href: "/student/interview",
      icon: Video,
    },
    {
      name: "\u771f\u4eba\u9762\u8bd5",
      href: "/interview/live/workbench",
      icon: ClipboardList,
    },
    {
      name: "\u5237\u9898\u4e2d\u5fc3",
      href: "/student/practice-mode",
      icon: BookOpen,
    },
    {
      name: "\u6210\u957f\u4e2d\u5fc3",
      href: "/student/growth",
      icon: TrendingUp,
    },
    {
      name: "\u7b80\u5386\u5206\u6790",
      href: "/student/resume",
      icon: FileText,
    },
    {
      name: "\u590d\u76d8\u62a5\u544a",
      href: "/student/history",
      icon: BarChart3,
    },
    {
      name: "\u6821\u53cb\u793e\u533a",
      href: "/student/community",
      icon: Users,
    },
  ],
  enterprise: [
    {
      name: "\u4f01\u4e1a\u603b\u89c8",
      href: "/enterprise/dashboard",
      icon: Building2,
    },
    {
      name: "\u9762\u8bd5\u5de5\u4f5c\u53f0",
      href: "/enterprise/interview-workbench",
      icon: ClipboardList,
    },
    { name: "\u4eba\u624d\u6c60", href: "/enterprise/talent", icon: UserCheck },
    {
      name: "\u5c97\u4f4d\u7ba1\u7406",
      href: "/enterprise/jobs",
      icon: Briefcase,
    },
    {
      name: "HR \u9762\u8bd5\u53f0",
      href: "/enterprise/hr-panel",
      icon: Video,
    },
    {
      name: "\u6570\u636e\u5206\u6790",
      href: "/enterprise/analytics",
      icon: BarChart3,
    },
    {
      name: "\u6807\u51c6\u5171\u5efa",
      href: "/enterprise/standards",
      icon: Database,
    },
  ],
  university: [
    {
      name: "\u7ba1\u7406\u603b\u89c8",
      href: "/university/dashboard",
      icon: GraduationCap,
    },
    {
      name: "\u9762\u8bd5\u5de5\u4f5c\u53f0",
      href: "/university/interview-workbench",
      icon: ClipboardList,
    },
    {
      name: "\u5b66\u751f\u8ddf\u8e2a",
      href: "/university/tracking",
      icon: Target,
    },
    {
      name: "\u5e2e\u6276\u4f53\u7cfb",
      href: "/university/support",
      icon: Shield,
    },
    {
      name: "\u8bfe\u7a0b\u8d44\u6e90",
      href: "/university/courses",
      icon: BookOpen,
    },
    {
      name: "\u5c31\u4e1a\u6570\u636e",
      href: "/university/employment",
      icon: BarChart3,
    },
    {
      name: "\u4eba\u624d\u63a8\u9001",
      href: "/university/talent-push",
      icon: Send,
    },
  ],
};

export const getPortalNavItems = (portal = "student") =>
  portalNavMap[portal] || portalNavMap.student;

export const isNavPathActive = (currentPath, navPath) => {
  if (navPath.endsWith("/dashboard")) return currentPath === navPath;
  return currentPath.startsWith(navPath);
};
