<template>
  <div class="min-h-screen bg-zinc-50 font-sans">
    <TopNav />

    <main class="w-full overflow-y-auto">
      <div class="mx-auto w-full max-w-350 px-8 py-8">
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
</template>

<script setup>
import TopNav from './TopNav.vue'
import ErrorBoundary from '../shared/ErrorBoundary.vue'
</script>
