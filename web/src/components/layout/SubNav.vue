<template>
  <nav class="bg-white border-b border-zinc-100 px-6">
    <div class="max-w-7xl mx-auto flex items-center gap-1 h-12">
      <router-link
        v-for="item in currentNavItems"
        :key="item.name"
        :to="item.href"
        class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all relative"
        :class="[
          isActive(item.href)
            ? activeClass
            : 'text-zinc-500 hover:text-zinc-700 hover:bg-zinc-50'
        ]"
      >
        <component :is="item.icon" class="h-4 w-4" />
        {{ item.name }}
        <div
          v-if="isActive(item.href)"
          class="absolute bottom-0 left-2 right-2 h-0.5 rounded-full"
          :class="activeBarClass"
        ></div>
      </router-link>
    </div>
  </nav>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import {
  LayoutDashboard, Video, BarChart3, Users,
  FileText, Building2, Target, BookOpen,
  GraduationCap, TrendingUp, UserCheck,
  Shield, Send, Database, Briefcase
} from 'lucide-vue-next'

const route = useRoute()

// Derive portal from route
const portal = computed(() => {
  const path = route.path
  if (path.startsWith('/enterprise')) return 'enterprise'
  if (path.startsWith('/university')) return 'university'
  return 'student'
})

const studentNav = [
  { name: '个人主页', href: '/student/dashboard', icon: LayoutDashboard },
  { name: '模拟面试', href: '/student/interview', icon: Video },
  { name: '复盘报告', href: '/student/history', icon: BarChart3 },
  { name: '成长中心', href: '/student/growth', icon: TrendingUp },
  { name: '简历匹配', href: '/student/resume', icon: FileText },
  { name: '校友社区', href: '/student/community', icon: Users },
]

const enterpriseNav = [
  { name: '企业总览', href: '/enterprise/dashboard', icon: Building2 },
  { name: '人才池', href: '/enterprise/talent', icon: UserCheck },
  { name: '岗位管理', href: '/enterprise/jobs', icon: Briefcase },
  { name: 'HR面试台', href: '/enterprise/hr-panel', icon: Video },
  { name: '数据分析', href: '/enterprise/analytics', icon: BarChart3 },
  { name: '标准共建', href: '/enterprise/standards', icon: Database },
]

const universityNav = [
  { name: '管理总览', href: '/university/dashboard', icon: GraduationCap },
  { name: '学生跟踪', href: '/university/tracking', icon: Target },
  { name: '帮扶体系', href: '/university/support', icon: Shield },
  { name: '课程资源', href: '/university/courses', icon: BookOpen },
  { name: '就业数据', href: '/university/employment', icon: BarChart3 },
  { name: '人才推送', href: '/university/talent-push', icon: Send },
]

const currentNavItems = computed(() => {
  if (portal.value === 'enterprise') return enterpriseNav
  if (portal.value === 'university') return universityNav
  return studentNav
})

const activeClass = computed(() => {
  if (portal.value === 'enterprise') return 'text-emerald-600'
  if (portal.value === 'university') return 'text-amber-600'
  return 'text-indigo-600'
})

const activeBarClass = computed(() => {
  if (portal.value === 'enterprise') return 'bg-emerald-600'
  if (portal.value === 'university') return 'bg-amber-600'
  return 'bg-indigo-600'
})

const isActive = (path) => {
  // Exact match for dashboard routes
  if (path.endsWith('/dashboard') && route.path === path) return true
  // Prefix match for non-dashboard routes
  if (!path.endsWith('/dashboard') && route.path.startsWith(path)) return true
  return false
}
</script>
