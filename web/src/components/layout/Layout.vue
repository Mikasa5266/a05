<template>
  <div class="flex min-h-screen bg-zinc-50 font-sans">
    <Sidebar />
    <main class="flex-1 p-8 overflow-y-auto h-screen">
      <div class="max-w-5xl mx-auto">
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
import Sidebar from './Sidebar.vue'
</script>

<style>
/* 页面切换动画已经在 Tailwind 类中定义，这里不再需要额外的 CSS */
</style>