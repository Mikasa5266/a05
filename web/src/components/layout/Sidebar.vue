<template>
  <aside
    :class="[
      'flex flex-col border-zinc-100 bg-white/95 backdrop-blur-sm shadow-sm',
      mobile
        ? 'w-full h-full border-0 shadow-none'
        : 'w-64 h-[calc(100vh-4rem)] sticky top-16 border-r'
    ]"
  >
    <!-- Logo 区 -->
    <div class="p-6 flex items-center gap-3">
      <div class="h-8 w-8 rounded-lg flex items-center justify-center text-white" :class="portalConfig.logoBg">
        <BrainCircuit class="h-5 w-5" />
      </div>
      <span class="font-bold text-xl" :class="portalConfig.logoText">{{ portalConfig.title }}</span>
    </div>

    <!-- 导航区 -->
    <nav class="flex-1 px-4 space-y-1 overflow-y-auto">
      <router-link
        v-for="item in currentNavItems"
        :key="item.name"
        :to="item.href"
        @click="handleNavigate"
        class="flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium transition-colors touch-manipulation active:scale-[0.99]"
        :class="[
          isNavPathActive(route.path, item.href)
            ? portalConfig.activeBg + ' ' + portalConfig.activeText
            : 'text-zinc-500 md:hover:bg-zinc-50 md:hover:text-zinc-900 active:bg-zinc-100'
        ]"
      >
        <component :is="item.icon" class="h-5 w-5" />
        {{ item.name }}
      </router-link>
    </nav>

    <!-- 底部区 -->
    <div class="p-4 border-t border-zinc-100 space-y-4">
      <router-link 
        :to="settingsPath"
        @click="handleNavigate"
        class="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium text-zinc-500 md:hover:bg-zinc-50 md:hover:text-zinc-900 active:bg-zinc-100 transition-colors touch-manipulation"
      >
        <Settings class="h-5 w-5" />
        设置
      </router-link>
      
      <router-link :to="settingsPath" @click="handleNavigate" class="flex items-center gap-3 px-3 py-2 rounded-xl md:hover:bg-zinc-50 active:bg-zinc-100 transition-colors cursor-pointer group touch-manipulation">
        <div class="h-8 w-8 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-600 font-bold overflow-hidden border border-indigo-200 md:group-hover:border-indigo-300 transition-colors">
          <img v-if="userStore.userInfo?.avatar" :src="avatarUrl" class="w-full h-full object-cover" />
          <span v-else>{{ userInitials }}</span>
        </div>
        <div class="flex flex-col">
          <span class="text-sm font-medium text-zinc-900 md:group-hover:text-indigo-600 transition-colors">{{ userStore.userInfo?.username || 'Guest' }}</span>
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
import { getBackendAssetUrl } from '../../utils/backend'
import { BrainCircuit, Settings } from 'lucide-vue-next'
import {
  getPortalFromPath,
  getPortalNavItems,
  isNavPathActive,
  portalBrandMap,
} from './navigation'

const props = defineProps({
  mobile: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['navigate'])

const route = useRoute()
const userStore = useUserStore()

const portal = computed(() => getPortalFromPath(route.path))
const portalConfig = computed(() => portalBrandMap[portal.value] || portalBrandMap.student)
const currentNavItems = computed(() => getPortalNavItems(portal.value))

const settingsPath = computed(() => '/' + portal.value + '/settings')

const handleNavigate = () => {
  if (props.mobile) emit('navigate')
}

const userInitials = computed(() => {
  const name = userStore.userInfo?.username || 'G'
  return name.substring(0, 2).toUpperCase()
})

const avatarUrl = computed(() => {
  if (!userStore.userInfo?.avatar) return ''
  return getBackendAssetUrl(userStore.userInfo.avatar)
})
</script>
