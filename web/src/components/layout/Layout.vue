<template>
  <div class="min-h-screen bg-zinc-50 font-sans pb-[env(safe-area-inset-bottom)]">
    <TopNav @toggle-mobile-menu="mobileDrawerOpen = true" />

    <div class="flex min-h-[calc(100vh-4rem)]">
      <Sidebar class="hidden md:flex md:shrink-0" />

      <div class="flex-1 min-w-0 flex flex-col">
        <SubNav class="hidden md:block" />

        <main class="flex-1 overflow-y-auto">
          <div class="max-w-7xl mx-auto px-4 md:px-6 py-5 md:py-6 pb-24 md:pb-6">
            <router-view v-slot="{ Component, route }">
              <ErrorBoundary :key="route.fullPath">
                <Suspense>
                  <template #default>
                    <transition
                      name="page"
                      mode="out-in"
                      enter-active-class="transition-all duration-200 ease-out"
                      leave-active-class="transition-all duration-200 ease-in"
                      enter-from-class="opacity-0 translate-y-4"
                      enter-to-class="opacity-100 translate-y-0"
                      leave-from-class="opacity-100 translate-y-0"
                      leave-to-class="opacity-0 -translate-y-4"
                    >
                      <component :is="Component" />
                    </transition>
                  </template>
                  <template #fallback>
                    <div class="h-64 flex items-center justify-center text-zinc-400">
                      页面加载中…
                    </div>
                  </template>
                </Suspense>
              </ErrorBoundary>
            </router-view>
          </div>
        </main>
      </div>
    </div>

    <el-drawer
      v-model="mobileDrawerOpen"
      direction="ltr"
      :size="'84%'"
      :with-header="false"
      :append-to-body="true"
      class="mobile-layout-drawer md:hidden"
    >
      <Sidebar mobile @navigate="closeMobileDrawer" />
    </el-drawer>

    <nav class="fixed bottom-0 left-0 right-0 z-40 border-t border-zinc-200 bg-white/95 backdrop-blur md:hidden pb-[env(safe-area-inset-bottom)]">
      <div class="grid grid-cols-5 gap-1 px-2 py-2">
        <router-link
          v-for="item in mobileNavItems"
          :key="item.href"
          :to="item.href"
          class="flex flex-col items-center justify-center gap-1 rounded-xl px-1 py-2 text-[11px] leading-tight text-zinc-500 transition-colors touch-manipulation active:bg-zinc-100"
          :class="isNavPathActive(route.path, item.href) ? mobileActiveClass : 'md:hover:bg-zinc-50'"
        >
          <component :is="item.icon" class="h-4 w-4" />
          <span class="truncate w-full text-center">{{ item.name }}</span>
        </router-link>
      </div>
    </nav>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import TopNav from './TopNav.vue'
import Sidebar from './Sidebar.vue'
import SubNav from './SubNav.vue'
import ErrorBoundary from '../shared/ErrorBoundary.vue'
import {
  getPortalFromPath,
  getPortalNavItems,
  isNavPathActive,
  portalBrandMap,
} from './navigation'

const route = useRoute()
const mobileDrawerOpen = ref(false)

const portal = computed(() => getPortalFromPath(route.path))
const portalConfig = computed(() => portalBrandMap[portal.value] || portalBrandMap.student)

const mobileNavItems = computed(() => getPortalNavItems(portal.value).slice(0, 5))

const mobileActiveClass = computed(() => portalConfig.value.activeBg + ' ' + portalConfig.value.activeText)

const closeMobileDrawer = () => {
  mobileDrawerOpen.value = false
}

watch(
  () => route.fullPath,
  () => {
    closeMobileDrawer()
  }
)
</script>

<style scoped>
:deep(.mobile-layout-drawer .el-drawer) {
  border-radius: 0 1rem 1rem 0;
}

:deep(.mobile-layout-drawer .el-drawer__body) {
  padding: 0;
}
</style>
