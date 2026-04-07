<script setup>
import { onMounted } from 'vue'
import { useUserStore } from './stores/user'
import CustomTitleBar from './components/desktop/CustomTitleBar.vue'
import { isElectronRuntime } from './utils/native'

const userStore = useUserStore()
const showCustomTitleBar = isElectronRuntime()

onMounted(() => {
  if (userStore.token && !userStore.userInfo) {
    userStore.getUserInfo()
  }
})
</script>

<template>
  <div class="app-shell" :class="{ 'app-shell-electron': showCustomTitleBar }">
    <CustomTitleBar v-if="showCustomTitleBar" />
    <main class="app-content">
      <router-view />
    </main>
  </div>
</template>

<style>
.app-shell {
  min-height: 100vh;
}

.app-shell-electron {
  height: 100vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.app-content {
  flex: 1;
  min-height: 0;
}
</style>
