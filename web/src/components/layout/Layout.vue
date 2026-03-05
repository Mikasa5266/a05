<template>
  <div class="flex flex-col min-h-screen bg-zinc-50 dark:bg-zinc-900 font-sans transition-colors duration-200">
    <TopNav />
    <SubNav />
    <main class="flex-1 overflow-y-auto">
      <div class="max-w-7xl mx-auto px-6 py-6">
        <router-view v-slot="{ Component }">
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
        </router-view>
      </div>
    </main>
  </div>
</template>

<script setup>
import TopNav from './TopNav.vue'
import SubNav from './SubNav.vue'
</script>

<style>
/* 页面切换动画已经在 Tailwind 类中定义，这里不再需要额外的 CSS */
</style>