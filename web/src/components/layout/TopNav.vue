<template>
  <header class="top-nav-surface h-16 border-b sticky top-0 z-50 flex items-center px-4 md:px-6">
    <!-- Mobile Menu -->
    <button
      type="button"
      class="mr-3 inline-flex md:hidden h-10 w-10 items-center justify-center rounded-xl text-zinc-500 active:bg-zinc-100 transition-colors touch-manipulation"
      aria-label="打开导航菜单"
      @click="emit('toggle-mobile-menu')"
    >
      <Menu class="h-5 w-5" />
    </button>

    <!-- Logo -->
    <router-link :to="portalHome" class="flex items-center gap-2.5 mr-3 md:mr-8 touch-manipulation">
      <div class="h-9 w-9 rounded-xl flex items-center justify-center text-white shadow-lg" :class="portalConfig.logoBg">
        <component :is="portalConfig.icon" class="h-5 w-5" />
      </div>
      <span class="font-bold text-lg md:text-xl tracking-tight" :class="portalConfig.logoText">{{ portalConfig.title }}</span>
    </router-link>

    <!-- Portal Badge (Center) -->
    <div class="hidden md:flex flex-1 justify-center">
      <div class="inline-flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium" :class="portalConfig.badgeClass">
        <component :is="portalConfig.icon" class="h-4 w-4" />
        {{ portalConfig.label }}
      </div>
    </div>

    <!-- Right Actions -->
    <div class="ml-auto flex items-center gap-2 md:gap-4">
      <router-link
        v-if="showPracticeModeEntry"
        :to="practiceModeTarget"
        class="top-nav-practice-entry inline-flex items-center gap-2 rounded-full px-3 py-2 text-sm font-semibold text-white shadow-sm transition-all touch-manipulation"
      >
        <BookOpen class="h-4 w-4" />
        <span class="hidden sm:inline">进入刷题模式</span>
        <span class="sm:hidden">刷题</span>
      </router-link>

      <!-- Notifications -->
      <button class="relative p-2 rounded-xl text-zinc-400 md:hover:text-zinc-600 md:hover:bg-zinc-50 active:text-zinc-600 active:bg-zinc-100 transition-colors touch-manipulation">
        <Bell class="h-5 w-5" />
        <span class="absolute top-1.5 right-1.5 w-2 h-2 bg-rose-500 rounded-full ring-2 ring-white"></span>
      </button>

      <!-- User Avatar -->
      <div class="relative" ref="dropdownRef">
        <button
          @click="showDropdown = !showDropdown"
          class="flex items-center gap-2 p-1.5 rounded-xl md:hover:bg-zinc-50 active:bg-zinc-100 transition-colors touch-manipulation"
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
            <router-link :to="settingsPath" @click="showDropdown = false" class="flex items-center gap-3 px-4 py-2.5 text-sm text-zinc-600 md:hover:bg-zinc-50 active:bg-zinc-100 transition-colors touch-manipulation">
              <Settings class="h-4 w-4" /> 设置
            </router-link>
            <button @click="handleLogout" class="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-rose-600 md:hover:bg-rose-50 active:bg-rose-100 transition-colors touch-manipulation">
              <LogOut class="h-4 w-4" /> 退出登录
            </button>
          </div>
        </transition>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '../../stores/user'
import { getBackendAssetUrl } from '../../utils/backend'
import { getPortalFromPath, portalBrandMap } from './navigation'
import {
  BookOpen,
  Menu,
  Bell, Settings, LogOut,
} from 'lucide-vue-next'

const emit = defineEmits(['toggle-mobile-menu'])

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const showDropdown = ref(false)
const dropdownRef = ref(null)

// Derive portal from current route
const currentPortal = computed(() => getPortalFromPath(route.path))

const portalConfig = computed(() => portalBrandMap[currentPortal.value] || portalBrandMap.student)

const portalHome = computed(() => '/' + currentPortal.value + '/dashboard')

const practiceModePath = '/student/practice-mode'
const practiceModeTarget = computed(() => ({
  path: practiceModePath,
  query: {
    from: route.fullPath,
  },
}))

const settingsPath = computed(() => '/' + currentPortal.value + '/settings')

const showPracticeModeEntry = computed(() => currentPortal.value === 'student')

const userInitials = computed(() => {
  const name = userStore.userInfo?.username || 'G'
  return name.substring(0, 2).toUpperCase()
})

const avatarUrl = computed(() => {
  if (!userStore.userInfo?.avatar) return ''
  return getBackendAssetUrl(userStore.userInfo.avatar)
})

const handleLogout = () => {
  showDropdown.value = false
  const role = currentPortal.value
  userStore.logout(role)
  router.push('/' + role + '/login')
}

// Close dropdown on outside click
const handleClickOutside = (e) => {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target)) {
    showDropdown.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
.top-nav-surface {
  background: color-mix(in srgb, var(--el-bg-color) 94%, transparent);
  border-color: var(--el-border-color-lighter);
  backdrop-filter: blur(6px);
  box-shadow: 0 4px 16px rgba(15, 23, 42, 0.05);
}

.top-nav-practice-entry {
  background: linear-gradient(135deg, #165d86 0%, #2f8fc5 100%);
}

.top-nav-practice-entry:hover {
  transform: translateY(-1px);
  box-shadow: 0 14px 30px rgba(22, 93, 134, 0.22);
}
</style>
