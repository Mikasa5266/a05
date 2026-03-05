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
            ? 'text-indigo-600'
            : 'text-zinc-500 hover:text-zinc-700 hover:bg-zinc-50'
        ]"
      >
        <component :is="item.icon" class="h-4 w-4" />
        {{ item.name }}
        <div
          v-if="isActive(item.href)"
          class="absolute bottom-0 left-2 right-2 h-0.5 bg-indigo-600 rounded-full"
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
  GraduationCap, TrendingUp, Clock, UserCheck,
  Shield, Send, Database, Briefcase
} from 'lucide-vue-next'

const props = defineProps({
  portal: { type: String, default: 'student' }
})

const route = useRoute()

const studentNav = [
  { name: '个人主页', href: '/dashboard', icon: LayoutDashboard },
  { name: '模拟面试', href: '/interview', icon: Video },
  { name: '复盘报告', href: '/history', icon: BarChart3 },
  { name: '成长中心', href: '/growth', icon: TrendingUp },
  { name: '简历匹配', href: '/resume', icon: FileText },
  { name: '校友社区', href: '/community', icon: Users },
]

const enterpriseNav = [
  { name: '企业总览', href: '/enterprise', icon: Building2 },
  { name: '人才池', href: '/enterprise/talent', icon: UserCheck },
  { name: '岗位管理', href: '/enterprise/jobs', icon: Briefcase },
  { name: 'HR面试台', href: '/enterprise/hr-panel', icon: Video },
  { name: '数据分析', href: '/enterprise/analytics', icon: BarChart3 },
  { name: '标准共建', href: '/enterprise/standards', icon: Database },
]

const universityNav = [
  { name: '管理总览', href: '/university', icon: GraduationCap },
  { name: '学生跟踪', href: '/university/tracking', icon: Target },
  { name: '帮扶体系', href: '/university/support', icon: Shield },
  { name: '课程资源', href: '/university/courses', icon: BookOpen },
  { name: '就业数据', href: '/university/employment', icon: BarChart3 },
  { name: '人才推送', href: '/university/talent-push', icon: Send },
]

const currentNavItems = computed(() => {
  if (props.portal === 'enterprise') return enterpriseNav
  if (props.portal === 'university') return universityNav
  return studentNav
})

const isActive = (path) => {
  if (path === '/dashboard' && route.path === '/dashboard') return true
  if (path === '/enterprise' && route.path === '/enterprise') return true
  if (path === '/university' && route.path === '/university') return true
  return route.path.startsWith(path) && path !== '/dashboard' && path !== '/enterprise' && path !== '/university'
}
</script>
