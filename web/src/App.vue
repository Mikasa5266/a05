<script setup>
import { onMounted } from 'vue'
import { useUserStore } from './stores/user'
import CustomTitleBar from './components/desktop/CustomTitleBar.vue'
import { isElectronRuntime } from './utils/native'

const userStore = useUserStore()
const showCustomTitleBar = isElectronRuntime()

onMounted(() => {
  const role = userStore.currentRole
  if (!userStore.hasValidTokenByRole(role) || userStore.userInfoLoaded) return

  const loadUserProfile = () => {
    void userStore.getUserInfo(role).catch(() => {})
  }

  if (typeof window !== 'undefined' && 'requestIdleCallback' in window) {
    window.requestIdleCallback(loadUserProfile, { timeout: 1200 })
    return
  }

  window.setTimeout(loadUserProfile, 0)
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
