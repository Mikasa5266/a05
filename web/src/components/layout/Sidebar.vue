<template>
  <aside class="w-64 h-screen sticky top-0 flex flex-col border-r border-zinc-100 dark:border-zinc-800 bg-white dark:bg-zinc-900 transition-colors duration-200">
    <!-- Logo 区 -->
    <div class="p-6 flex items-center gap-3">
      <div class="h-8 w-8 bg-indigo-600 rounded-lg flex items-center justify-center text-white">
        <BrainCircuit class="h-5 w-5" />
      </div>
      <span class="font-bold text-xl text-zinc-900 dark:text-white">AI Interview</span>
    </div>

    <!-- 导航区 -->
    <nav class="flex-1 px-4 space-y-1 overflow-y-auto">
      <router-link
        v-for="item in navigation"
        :key="item.name"
        :to="item.href"
        class="flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium transition-colors"
        :class="[
          isActive(item.href)
            ? 'bg-indigo-50 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400'
            : 'text-zinc-500 hover:bg-zinc-50 hover:text-zinc-900 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-100'
        ]"
      >
        <component :is="item.icon" class="h-5 w-5" />
        {{ item.name }}
      </router-link>
    </nav>

    <!-- 底部区 -->
    <div class="p-4 border-t border-zinc-100 dark:border-zinc-800 space-y-4">
      <router-link 
        to="/student/settings"
        class="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium text-zinc-500 hover:bg-zinc-50 hover:text-zinc-900 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-100 transition-colors"
      >
        <Settings class="h-5 w-5" />
        设置
      </router-link>
      
      <router-link to="/student/settings" class="flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-colors cursor-pointer group">
        <div class="h-8 w-8 rounded-full bg-indigo-100 dark:bg-indigo-900/50 flex items-center justify-center text-indigo-600 dark:text-indigo-400 font-bold overflow-hidden border border-indigo-200 dark:border-indigo-800 group-hover:border-indigo-300 dark:group-hover:border-indigo-700 transition-colors">
          <img v-if="userStore.userInfo?.avatar" :src="avatarUrl" class="w-full h-full object-cover" />
          <span v-else>{{ userInitials }}</span>
        </div>
        <div class="flex flex-col">
          <span class="text-sm font-medium text-zinc-900 dark:text-white group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">{{ userStore.userInfo?.username || 'Guest' }}</span>
          <span class="text-xs text-zinc-400">求职者</span>
        </div>
      </router-link>
    </div>
  </aside>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '../../stores/user'
import { 
  BrainCircuit, 
  LayoutDashboard, 
  FileText, 
  Video, 
  TrendingUp, 
  Clock, 
  Settings 
} from 'lucide-vue-next'

const route = useRoute()
const userStore = useUserStore()

const navigation = [
  { name: '首页', href: '/student/dashboard', icon: LayoutDashboard },
  { name: '简历匹配', href: '/student/resume', icon: FileText },
  { name: '模拟面试', href: '/student/interview', icon: Video },
  { name: '成长中心', href: '/student/growth', icon: TrendingUp },
  { name: '面试记录', href: '/student/history', icon: Clock },
]

const isActive = (path) => {
  if (path === '/student/dashboard' && route.path === '/student/dashboard') return true
  return route.path.startsWith(path) && path !== '/student/dashboard'
}

const userInitials = computed(() => {
  const name = userStore.userInfo?.username || 'G'
  return name.substring(0, 2).toUpperCase()
})

const avatarUrl = computed(() => {
  if (!userStore.userInfo?.avatar) return ''
  if (userStore.userInfo.avatar.startsWith('http')) return userStore.userInfo.avatar
  return `http://localhost:8080${userStore.userInfo.avatar}`
})
</script>