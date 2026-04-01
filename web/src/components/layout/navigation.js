import {
  User,
  Building2,
  GraduationCap,
  LayoutDashboard,
  Video,
  BarChart3,
  Users,
  FileText,
  Target,
  BookOpen,
  TrendingUp,
  UserCheck,
  Shield,
  Send,
  Database,
  Briefcase,
  ClipboardList,
} from 'lucide-vue-next'

export const getPortalFromPath = (path = '') => {
  if (path.startsWith('/enterprise')) return 'enterprise'
  if (path.startsWith('/university')) return 'university'
  return 'student'
}

export const portalBrandMap = {
  student: {
    title: '智聘AI',
    label: '学生端',
    icon: User,
    logoBg: 'bg-indigo-600 shadow-indigo-200',
    logoText: 'text-indigo-600',
    badgeClass: 'bg-indigo-50 text-indigo-600 border border-indigo-100',
    activeText: 'text-indigo-600',
    activeBg: 'bg-indigo-50',
  },
  enterprise: {
    title: '智聘AI',
    label: '企业端',
    icon: Building2,
    logoBg: 'bg-emerald-600 shadow-emerald-200',
    logoText: 'text-emerald-600',
    badgeClass: 'bg-emerald-50 text-emerald-600 border border-emerald-100',
    activeText: 'text-emerald-600',
    activeBg: 'bg-emerald-50',
  },
  university: {
    title: '智聘AI',
    label: '高校端',
    icon: GraduationCap,
    logoBg: 'bg-amber-600 shadow-amber-200',
    logoText: 'text-amber-600',
    badgeClass: 'bg-amber-50 text-amber-600 border border-amber-100',
    activeText: 'text-amber-600',
    activeBg: 'bg-amber-50',
  },
}

export const portalNavMap = {
  student: [
    { name: '个人主页', href: '/student/dashboard', icon: LayoutDashboard },
    { name: '简历匹配', href: '/student/resume', icon: FileText },
    { name: '模拟面试', href: '/student/interview', icon: Video },
    { name: '真人面试', href: '/interview/live/workbench', icon: ClipboardList },
    { name: '复盘报告', href: '/student/history', icon: BarChart3 },
    { name: '成长中心', href: '/student/growth', icon: TrendingUp },
    { name: '校友社区', href: '/student/community', icon: Users },
  ],
  enterprise: [
    { name: '企业总览', href: '/enterprise/dashboard', icon: Building2 },
    { name: '面试工作台', href: '/enterprise/interview-workbench', icon: ClipboardList },
    { name: '人才池', href: '/enterprise/talent', icon: UserCheck },
    { name: '岗位管理', href: '/enterprise/jobs', icon: Briefcase },
    { name: 'HR面试台', href: '/enterprise/hr-panel', icon: Video },
    { name: '数据分析', href: '/enterprise/analytics', icon: BarChart3 },
    { name: '标准共建', href: '/enterprise/standards', icon: Database },
  ],
  university: [
    { name: '管理总览', href: '/university/dashboard', icon: GraduationCap },
    { name: '面试工作台', href: '/university/interview-workbench', icon: ClipboardList },
    { name: '学生跟踪', href: '/university/tracking', icon: Target },
    { name: '帮扶体系', href: '/university/support', icon: Shield },
    { name: '课程资源', href: '/university/courses', icon: BookOpen },
    { name: '就业数据', href: '/university/employment', icon: BarChart3 },
    { name: '人才推送', href: '/university/talent-push', icon: Send },
  ],
}

export const getPortalNavItems = (portal = 'student') => portalNavMap[portal] || portalNavMap.student

export const isNavPathActive = (currentPath, navPath) => {
  if (navPath.endsWith('/dashboard')) return currentPath === navPath
  return currentPath.startsWith(navPath)
}
