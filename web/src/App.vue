<script setup>
import { onMounted } from 'vue'
import { useUserStore } from './stores/user'

const userStore = useUserStore()

onMounted(() => {
  if (userStore.token && !userStore.userInfo) {
    userStore.getUserInfo()
  }
})
</script>

<template>
  <router-view />
</template>

<style>
/* Global styles are imported in main.js or index.css */
</style>
