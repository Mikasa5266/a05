<template>
  <header class="top-nav-surface sticky top-0 z-50 border-b">
    <div class="mx-auto flex h-16 w-full max-w-400 items-center px-8">
      <router-link :to="portalHome" class="brand-group mr-10 flex items-center gap-3">
        <div class="brand-logo flex h-9 w-9 items-center justify-center rounded-xl text-white" :class="portalConfig.logoBg">
          <component :is="portalConfig.icon" class="h-5 w-5" />
        </div>
        <div class="leading-tight">
          <div class="text-lg font-bold tracking-tight" :class="portalConfig.logoText">{{ portalConfig.title }}</div>
          <div class="text-[11px] font-medium text-zinc-400">{{ portalConfig.label }}</div>
        </div>
      </router-link>

      <nav class="menu-zone flex items-center gap-1">
        <router-link
          v-for="item in topNavItems"
          :key="item.href"
          :to="item.href"
          class="menu-item inline-flex items-center gap-2 rounded-xl px-4 py-2 text-sm font-semibold text-zinc-600"
          :class="isActive(item) ? menuActiveClass : ''"
        >
          <component :is="item.icon" class="h-4 w-4" />
          <span>{{ item.name }}</span>
        </router-link>
      </nav>

      <div class="ml-auto flex items-center gap-4">
        <div class="relative" ref="dropdownRef">
          <button
            type="button"
            class="avatar-trigger flex items-center gap-2 rounded-xl p-1.5"
            @click="showDropdown = !showDropdown"
          >
            <div class="h-9 w-9 overflow-hidden rounded-full border-2 border-indigo-200 bg-indigo-100 text-sm font-bold text-indigo-600 flex items-center justify-center">
              <img v-if="userStore.userInfo?.avatar" :src="avatarUrl" class="h-full w-full object-cover" />
              <span v-else>{{ userInitials }}</span>
            </div>
            <span class="max-w-35 truncate text-sm font-semibold text-zinc-700">{{ displayName }}</span>
            <ChevronDown class="h-4 w-4 text-zinc-400" />
          </button>

          <transition
            enter-active-class="transition-all duration-200 ease-out"
            leave-active-class="transition-all duration-150 ease-in"
            enter-from-class="opacity-0 scale-95 -translate-y-1"
            enter-to-class="opacity-100 scale-100 translate-y-0"
            leave-from-class="opacity-100 scale-100 translate-y-0"
            leave-to-class="opacity-0 scale-95 -translate-y-1"
          >
            <div v-if="showDropdown" class="absolute right-0 top-12 z-50 w-56 rounded-2xl border border-zinc-100 bg-white py-2 shadow-xl">
              <div class="border-b border-zinc-100 px-4 py-3">
                <div class="text-sm font-medium text-zinc-900">{{ displayName }}</div>
                <div class="text-xs text-zinc-400">{{ displayEmail }}</div>
              </div>
              <router-link
                :to="settingsPath"
                class="flex items-center gap-3 px-4 py-2.5 text-sm text-zinc-600 hover:bg-zinc-50"
                @click="showDropdown = false"
              >
                <Settings class="h-4 w-4" /> 设置
              </router-link>
              <button
                type="button"
                class="w-full flex items-center gap-3 px-4 py-2.5 text-left text-sm text-rose-600 hover:bg-rose-50"
                @click="handleLogout"
              >
                <LogOut class="h-4 w-4" /> 退出登录
              </button>
            </div>
          </transition>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '../../stores/user'
import { getBackendAssetUrl } from '../../utils/backend'
import { getPortalFromPath, getPortalTopNavItems, isNavPathActive, portalBrandMap } from './navigation'
import {
  ChevronDown,
  Settings,
  LogOut,
} from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const showDropdown = ref(false)
const dropdownRef = ref(null)

const currentPortal = computed(() => getPortalFromPath(route.path))
const portalConfig = computed(() => portalBrandMap[currentPortal.value] || portalBrandMap.student)
const topNavItems = computed(() => getPortalTopNavItems(currentPortal.value))

const portalHome = computed(() => '/' + currentPortal.value + '/dashboard')

const settingsPath = computed(() => '/' + currentPortal.value + '/settings')
const menuActiveClass = computed(() => portalConfig.value.activeBg + ' ' + portalConfig.value.activeText)

const displayName = computed(() => {
  if (!userStore.userInfoLoaded || !userStore.userInfo) {
    return '加载中...'
  }

  const username = String(userStore.userInfo.username || '').trim()
  if (username) return username

  const email = String(userStore.userInfo.email || '').trim()
  if (email) return email.split('@')[0] || email

  const id = userStore.userInfo.id
  if (id) return `用户#${id}`
  return '未命名用户'
})

const displayEmail = computed(() => {
  if (!userStore.userInfoLoaded || !userStore.userInfo) return '加载中...'
  return String(userStore.userInfo.email || '').trim()
})

const userInitials = computed(() => {
  const name = displayName.value || 'G'
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

const isActive = (item) => {
  if (item.exact) {
    return route.path === item.href
  }
  return isNavPathActive(route.path, item.href)
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
  background: color-mix(in srgb, var(--el-bg-color) 96%, transparent);
  border-color: var(--el-border-color-lighter);
  backdrop-filter: blur(8px);
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.06);
}

.brand-group {
  text-decoration: none;
}

.brand-logo {
  box-shadow: 0 10px 22px rgba(79, 70, 229, 0.25);
}

.menu-zone {
  min-width: 520px;
}

.menu-item {
  text-decoration: none;
  transition: all 0.2s ease;
}

.menu-item:hover {
  background: #f4f7fb;
  color: #334155;
}

.avatar-trigger {
  transition: background-color 0.2s ease;
}

.avatar-trigger:hover {
  background: #f4f7fb;
}
</style>
