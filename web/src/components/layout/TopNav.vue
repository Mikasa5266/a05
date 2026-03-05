<template>
  <header class="h-16 bg-white border-b border-zinc-100 sticky top-0 z-50 flex items-center px-6 shadow-sm">
    <!-- Logo -->
    <router-link :to="portalHome" class="flex items-center gap-2.5 mr-8">
      <div class="h-9 w-9 rounded-xl flex items-center justify-center text-white shadow-lg" :class="portalConfig.logoBg">
        <component :is="portalConfig.icon" class="h-5 w-5" />
      </div>
      <span class="font-bold text-xl tracking-tight" :class="portalConfig.logoText">{{ portalConfig.title }}</span>
    </router-link>

    <!-- Portal Badge (Center) -->
    <div class="flex-1 flex justify-center">
      <div class="inline-flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium" :class="portalConfig.badgeClass">
        <component :is="portalConfig.icon" class="h-4 w-4" />
        {{ portalConfig.label }}
      </div>
    </div>

    <!-- Right Actions -->
    <div class="flex items-center gap-4">
      <!-- Notifications -->
      <button class="relative p-2 rounded-xl text-zinc-400 hover:text-zinc-600 hover:bg-zinc-50 transition-colors">
        <Bell class="h-5 w-5" />
        <span class="absolute top-1.5 right-1.5 w-2 h-2 bg-rose-500 rounded-full ring-2 ring-white"></span>
      </button>

      <!-- User Avatar -->
      <div class="relative" ref="dropdownRef">
        <button
          @click="showDropdown = !showDropdown"
          class="flex items-center gap-2 p-1.5 rounded-xl hover:bg-zinc-50 transition-colors"
        >
          <div class="h-8 w-8 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-600 font-bold text-sm overflow-hidden border-2 border-indigo-200">
            <img v-if="userStore.userInfo?.avatar" :src="avatarUrl" class="w-full h-full object-cover" />
            <span v-else>{{ userInitials }}</span>
          </div>
        </button>

        <!-- Dropdown -->
        <transition
          enter-active-class="transition-all duration-200 ease-out"
          leave-active-class="transition-all duration-150 ease-in"
          enter-from-class="opacity-0 scale-95 -translate-y-1"
          enter-to-class="opacity-100 scale-100 translate-y-0"
          leave-from-class="opacity-100 scale-100 translate-y-0"
          leave-to-class="opacity-0 scale-95 -translate-y-1"
        >
          <div v-if="showDropdown" class="absolute right-0 top-12 w-56 bg-white rounded-2xl shadow-xl border border-zinc-100 py-2 z-50">
            <div class="px-4 py-3 border-b border-zinc-100">
              <div class="font-medium text-zinc-900 text-sm">{{ userStore.userInfo?.username || 'Guest' }}</div>
              <div class="text-xs text-zinc-400">{{ userStore.userInfo?.email || '' }}</div>
            </div>
            <router-link :to="settingsPath" @click="showDropdown = false" class="flex items-center gap-3 px-4 py-2.5 text-sm text-zinc-600 hover:bg-zinc-50 transition-colors">
              <Settings class="h-4 w-4" /> 设置
            </router-link>
            <button @click="handleLogout" class="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-rose-600 hover:bg-rose-50 transition-colors">
              <LogOut class="h-4 w-4" /> 退出登录
            </button>
          </div>
        </transition>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '../../stores/user'
import {
  Bell, Settings, LogOut,
  User, Building2, GraduationCap
} from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const showDropdown = ref(false)
const dropdownRef = ref(null)

// Derive portal from current route
const currentPortal = computed(() => {
  const path = route.path
  if (path.startsWith('/enterprise')) return 'enterprise'
  if (path.startsWith('/university')) return 'university'
  return 'student'
})

const portalConfigs = {
  student: {
    title: '智聘AI',
    label: '学生端',
    icon: User,
    logoBg: 'bg-indigo-600 shadow-indigo-200',
    logoText: 'text-indigo-600',
    badgeClass: 'bg-indigo-50 text-indigo-600 border border-indigo-100',
  },
  enterprise: {
    title: '智聘AI',
    label: '企业端',
    icon: Building2,
    logoBg: 'bg-emerald-600 shadow-emerald-200',
    logoText: 'text-emerald-600',
    badgeClass: 'bg-emerald-50 text-emerald-600 border border-emerald-100',
  },
  university: {
    title: '智聘AI',
    label: '高校端',
    icon: GraduationCap,
    logoBg: 'bg-amber-600 shadow-amber-200',
    logoText: 'text-amber-600',
    badgeClass: 'bg-amber-50 text-amber-600 border border-amber-100',
  }
}

const portalConfig = computed(() => portalConfigs[currentPortal.value])

const portalHome = computed(() => `/${currentPortal.value}/dashboard`)

const settingsPath = computed(() => `/${currentPortal.value}/settings`)

const userInitials = computed(() => {
  const name = userStore.userInfo?.username || 'G'
  return name.substring(0, 2).toUpperCase()
})

const avatarUrl = computed(() => {
  if (!userStore.userInfo?.avatar) return ''
  if (userStore.userInfo.avatar.startsWith('http')) return userStore.userInfo.avatar
  return `http://localhost:8080${userStore.userInfo.avatar}`
})

const handleLogout = () => {
  showDropdown.value = false
  userStore.logout()
  router.push('/')
}

// Close dropdown on outside click
const handleClickOutside = (e) => {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target)) {
    showDropdown.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))
</script>
