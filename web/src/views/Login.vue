<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-zinc-50 via-indigo-50/30 to-violet-50/30 relative overflow-hidden">
    <!-- Background decoration -->
    <div class="absolute top-0 left-0 w-[600px] h-[600px] bg-indigo-200/20 rounded-full -translate-x-1/2 -translate-y-1/2 blur-3xl"></div>
    <div class="absolute bottom-0 right-0 w-[500px] h-[500px] bg-violet-200/20 rounded-full translate-x-1/3 translate-y-1/3 blur-3xl"></div>

    <div class="w-full max-w-md mx-4 relative z-10">
      <!-- Logo & Title -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-indigo-600 to-violet-600 rounded-2xl shadow-xl shadow-indigo-200 mb-4">
          <span class="text-2xl font-black text-white">AI</span>
        </div>
        <h1 class="text-3xl font-bold text-zinc-900">智聘 AI</h1>
        <p class="text-zinc-500 mt-2 text-sm">AI 驱动的智能面试与人才评估平台</p>
      </div>

      <!-- Portal Selector -->
      <div class="flex items-center justify-center gap-1 mb-6 p-1 bg-zinc-100 rounded-2xl">
        <button
          v-for="portal in portals"
          :key="portal.key"
          @click="selectedPortal = portal.key"
          class="flex-1 py-2.5 px-4 rounded-xl text-sm font-medium transition-all duration-200"
          :class="selectedPortal === portal.key
            ? 'bg-white text-zinc-900 shadow-sm'
            : 'text-zinc-500 hover:text-zinc-700'"
        >
          {{ portal.label }}
        </button>
      </div>

      <!-- Card -->
      <div class="bg-white/80 backdrop-blur-xl rounded-3xl shadow-xl shadow-zinc-200/50 border border-white/50 p-8">
        <!-- Tabs -->
        <div class="flex items-center gap-6 mb-8 border-b border-zinc-100">
          <button
            @click="activeTab = 'login'"
            class="pb-3 text-sm font-semibold transition-colors border-b-2 -mb-px"
            :class="activeTab === 'login' ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-zinc-400 hover:text-zinc-600'"
          >
            登录
          </button>
          <button
            @click="activeTab = 'register'"
            class="pb-3 text-sm font-semibold transition-colors border-b-2 -mb-px"
            :class="activeTab === 'register' ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-zinc-400 hover:text-zinc-600'"
          >
            注册
          </button>
        </div>

        <!-- Login Form -->
        <form v-if="activeTab === 'login'" @submit.prevent="handleLogin" class="space-y-5">
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-1.5 block">邮箱</label>
            <input v-model="loginForm.email" type="email" required placeholder="请输入邮箱"
              class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-400 transition-colors" />
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-1.5 block">密码</label>
            <input v-model="loginForm.password" type="password" required placeholder="请输入密码" minlength="6"
              class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-400 transition-colors" />
          </div>
          <button type="submit" :disabled="loading"
            class="w-full py-3.5 bg-gradient-to-r from-indigo-600 to-violet-600 text-white rounded-xl text-sm font-bold hover:from-indigo-700 hover:to-violet-700 transition-all shadow-lg shadow-indigo-200 disabled:opacity-60">
            {{ loading ? '登录中...' : '登录' }}
          </button>
        </form>

        <!-- Register Form -->
        <form v-else @submit.prevent="handleRegister" class="space-y-5">
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-1.5 block">用户名</label>
            <input v-model="registerForm.username" type="text" required placeholder="请输入用户名"
              class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-400 transition-colors" />
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-1.5 block">邮箱</label>
            <input v-model="registerForm.email" type="email" required placeholder="请输入邮箱"
              class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-400 transition-colors" />
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-1.5 block">密码</label>
            <input v-model="registerForm.password" type="password" required placeholder="请输入密码 (至少6位)" minlength="6"
              class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-400 transition-colors" />
          </div>
          <button type="submit" :disabled="loading"
            class="w-full py-3.5 bg-gradient-to-r from-indigo-600 to-violet-600 text-white rounded-xl text-sm font-bold hover:from-indigo-700 hover:to-violet-700 transition-all shadow-lg shadow-indigo-200 disabled:opacity-60">
            {{ loading ? '注册中...' : '注册' }}
          </button>
        </form>
      </div>

      <!-- Footer -->
      <p class="text-center text-xs text-zinc-400 mt-6">智聘 AI &copy; 2025 · AI 驱动面试评估平台</p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { ElMessage } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()

const activeTab = ref('login')
const loading = ref(false)
const selectedPortal = ref('student')

const portals = [
  { key: 'student', label: '学生端' },
  { key: 'enterprise', label: '企业端' },
  { key: 'university', label: '高校端' },
]

const loginForm = reactive({
  email: '',
  password: ''
})

const registerForm = reactive({
  username: '',
  email: '',
  password: ''
})

const handleLogin = async () => {
  if (!loginForm.email || !loginForm.password) return
  loading.value = true
  try {
    await userStore.login(loginForm)
    ElMessage.success('登录成功')
    localStorage.setItem('portal', selectedPortal.value)
    const portalRoutes = { student: '/dashboard', enterprise: '/enterprise', university: '/university' }
    router.push(portalRoutes[selectedPortal.value] || '/dashboard')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '登录失败')
  } finally {
    loading.value = false
  }
}

const handleRegister = async () => {
  if (!registerForm.username || !registerForm.email || !registerForm.password) return
  loading.value = true
  try {
    await userStore.register(registerForm)
    ElMessage.success('注册成功，请登录')
    activeTab.value = 'login'
    loginForm.email = registerForm.email
    loginForm.password = registerForm.password
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '注册失败')
  } finally {
    loading.value = false
  }
}
</script>