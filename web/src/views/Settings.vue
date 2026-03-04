<template>
  <div class="max-w-2xl mx-auto space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900">设置</h1>
      <p class="text-zinc-500 mt-2">管理您的账户偏好与应用设置</p>
    </header>

    <div class="bg-white rounded-3xl shadow-sm border border-zinc-100 overflow-hidden divide-y divide-zinc-100">
      <!-- Profile Section -->
      <div class="p-8">
        <h2 class="text-lg font-bold text-zinc-900 mb-4 flex items-center gap-2">
          <User class="h-5 w-5 text-indigo-600" />
          个人资料
        </h2>
        <div class="space-y-4">
          <div class="flex items-center gap-4">
            <div class="h-16 w-16 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-600 font-bold text-xl">
              {{ userStore.username ? userStore.username.charAt(0).toUpperCase() : 'U' }}
            </div>
            <div>
              <div class="font-medium text-zinc-900">{{ userStore.username || 'User' }}</div>
              <div class="text-sm text-zinc-500">Standard Plan</div>
            </div>
            <button class="ml-auto text-sm text-indigo-600 font-medium hover:underline">
              更换头像
            </button>
          </div>
        </div>
      </div>

      <!-- App Settings -->
      <div class="p-8">
        <h2 class="text-lg font-bold text-zinc-900 mb-4 flex items-center gap-2">
          <SettingsIcon class="h-5 w-5 text-indigo-600" />
          应用偏好
        </h2>
        <div class="space-y-6">
          <div class="flex items-center justify-between">
            <div>
              <div class="font-medium text-zinc-900">深色模式</div>
              <div class="text-sm text-zinc-500">启用暗黑主题界面</div>
            </div>
            <button 
              class="w-12 h-6 rounded-full bg-zinc-200 relative transition-colors"
              :class="{ 'bg-indigo-600': isDarkMode }"
              @click="toggleDarkMode"
            >
              <div 
                class="absolute top-1 left-1 bg-white w-4 h-4 rounded-full transition-transform"
                :class="{ 'translate-x-6': isDarkMode }"
              ></div>
            </button>
          </div>

          <div class="flex items-center justify-between">
            <div>
              <div class="font-medium text-zinc-900">面试音效</div>
              <div class="text-sm text-zinc-500">播放 AI 语音反馈</div>
            </div>
            <button 
              class="w-12 h-6 rounded-full bg-indigo-600 relative transition-colors"
            >
              <div class="absolute top-1 left-1 bg-white w-4 h-4 rounded-full translate-x-6"></div>
            </button>
          </div>
        </div>
      </div>

      <!-- Account Actions -->
      <div class="p-8 bg-zinc-50">
        <h2 class="text-lg font-bold text-zinc-900 mb-4 flex items-center gap-2">
          <Shield class="h-5 w-5 text-indigo-600" />
          账户安全
        </h2>
        <div class="space-y-4">
          <button class="w-full text-left px-4 py-3 bg-white border border-zinc-200 rounded-xl text-sm font-medium text-zinc-700 hover:bg-zinc-50 transition-colors">
            修改密码
          </button>
          
          <button 
            @click="handleLogout"
            class="w-full text-left px-4 py-3 bg-white border border-rose-200 rounded-xl text-sm font-medium text-rose-600 hover:bg-rose-50 transition-colors flex items-center justify-between group"
          >
            <span>退出登录</span>
            <LogOut class="h-4 w-4 group-hover:translate-x-1 transition-transform" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { User, Settings as SettingsIcon, Shield, LogOut } from 'lucide-vue-next'

const router = useRouter()
const userStore = useUserStore()
const isDarkMode = ref(false)

const toggleDarkMode = () => {
  isDarkMode.value = !isDarkMode.value
  // Implement actual dark mode toggle logic here if needed
}

const handleLogout = () => {
  if (confirm('确定要退出登录吗？')) {
    userStore.logout()
    router.push('/login')
  }
}
</script>
