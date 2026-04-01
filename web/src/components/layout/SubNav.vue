<template>
  <nav class="sub-nav-surface border-b px-6">
    <div class="max-w-7xl mx-auto flex items-center gap-1 h-12">
      <router-link
        v-for="item in currentNavItems"
        :key="item.name"
        :to="item.href"
        class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all relative touch-manipulation active:bg-zinc-100"
        :class="[
          isNavPathActive(route.path, item.href)
            ? activeClass
            : 'text-zinc-500 md:hover:text-zinc-700 md:hover:bg-zinc-50'
        ]"
      >
        <component :is="item.icon" class="h-4 w-4" />
        {{ item.name }}
        <div
          v-if="isNavPathActive(route.path, item.href)"
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
  getPortalFromPath,
  getPortalNavItems,
  isNavPathActive,
  portalBrandMap,
} from './navigation'

const route = useRoute()

const portal = computed(() => getPortalFromPath(route.path))

const currentNavItems = computed(() => getPortalNavItems(portal.value))

const activeClass = computed(() => (portalBrandMap[portal.value] || portalBrandMap.student).activeText)

const activeBarClass = computed(() => {
  if (portal.value === 'enterprise') return 'bg-emerald-600'
  if (portal.value === 'university') return 'bg-amber-600'
  return 'bg-indigo-600'
})
</script>

<style scoped>
.sub-nav-surface {
  background: color-mix(in srgb, var(--el-bg-color) 90%, transparent);
  border-color: var(--el-border-color-lighter);
}
</style>
