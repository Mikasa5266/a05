<template>
  <aside class="w-64 h-screen sticky top-0 flex flex-col border-r border-zinc-100 bg-white">
    <!-- Logo 区 -->
    <div class="p-6 flex items-center gap-3">
      <div class="h-8 w-8 bg-indigo-600 rounded-lg flex items-center justify-center text-white">
        <BrainCircuit class="h-5 w-5" />
      </div>
      <span class="font-bold text-xl text-zinc-900">AI Interview</span>
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
            ? 'bg-indigo-50 text-indigo-600'
            : 'text-zinc-500 hover:bg-zinc-50 hover:text-zinc-900'
        ]"
      >
        <component :is="item.icon" class="h-5 w-5" />
        {{ item.name }}
      </router-link>
    </nav>

    <!-- 底部区 -->
    <div class="p-4 border-t border-zinc-100 space-y-4">
      <router-link 
        to="/settings"
        class="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium text-zinc-500 hover:bg-zinc-50 hover:text-zinc-900 transition-colors"
      >
        <Settings class="h-5 w-5" />
        设置
      </router-link>
      
      <div class="flex items-center gap-3 px-3 py-2">
        <div class="h-8 w-8 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-600 font-bold">
          {{ userInitials }}
        </div>
        <div class="flex flex-col">
          <span class="text-sm font-medium text-zinc-900">{{ userStore.userInfo?.username || 'Guest' }}</span>
          <span class="text-xs text-zinc-400">求职者</span>
        </div>
      </div>
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
  { name: '首页', href: '/dashboard', icon: LayoutDashboard },
  { name: '简历匹配', href: '/resume', icon: FileText },
  { name: '模拟面试', href: '/interview/select', icon: Video },
  { name: '成长中心', href: '/growth', icon: TrendingUp },
  { name: '面试记录', href: '/history', icon: Clock },
]

const isActive = (path) => {
  if (path === '/dashboard' && route.path === '/dashboard') return true
  return route.path.startsWith(path) && path !== '/dashboard'
}

const userInitials = computed(() => {
  const name = userStore.userInfo?.username || 'G'
  return name.substring(0, 2).toUpperCase()
})
</script>