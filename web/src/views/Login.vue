<template>
  <div class="min-h-screen flex items-center justify-center relative overflow-hidden" :class="portalTheme.bgGradient">
    <!-- Background decoration -->
    <div class="absolute top-0 left-0 w-[600px] h-[600px] rounded-full -translate-x-1/2 -translate-y-1/2 blur-3xl" :class="portalTheme.blob1"></div>
    <div class="absolute bottom-0 right-0 w-[500px] h-[500px] rounded-full translate-x-1/3 translate-y-1/3 blur-3xl" :class="portalTheme.blob2"></div>

    <div class="w-full max-w-md mx-4 relative z-10">
      <!-- Logo & Title -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl shadow-xl mb-4" :class="portalTheme.logoBg">
          <component :is="portalTheme.icon" class="h-8 w-8 text-white" />
        </div>
        <h1 class="text-3xl font-bold text-zinc-900">{{ portalTheme.title }}</h1>
        <p class="text-zinc-500 mt-2 text-sm">{{ portalTheme.subtitle }}</p>
      </div>

      <!-- Card -->
      <div class="bg-white/80 backdrop-blur-xl rounded-3xl shadow-xl shadow-zinc-200/50 border border-white/50 p-8">
        <!-- Tabs -->
        <div class="flex items-center gap-6 mb-8 border-b border-zinc-100">
          <button
            @click="activeTab = 'login'"
            class="pb-3 text-sm font-semibold transition-colors border-b-2 -mb-px"
            :class="activeTab === 'login' ? portalTheme.tabActive : 'border-transparent text-zinc-400 hover:text-zinc-600'"
          >
            登录
          </button>
          <button
            @click="activeTab = 'register'"
            class="pb-3 text-sm font-semibold transition-colors border-b-2 -mb-px"
            :class="activeTab === 'register' ? portalTheme.tabActive : 'border-transparent text-zinc-400 hover:text-zinc-600'"
          >
            注册
          </button>
        </div>

        <!-- Login Form -->
        <form v-if="activeTab === 'login'" @submit.prevent="handleLogin" class="space-y-5">
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-1.5 block">邮箱</label>
            <input v-model="loginForm.email" type="email" required placeholder="请输入邮箱"
              class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:border-indigo-400 transition-colors" />
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-1.5 block">密码</label>
            <input v-model="loginForm.password" type="password" required placeholder="请输入密码" minlength="6"
              class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:border-indigo-400 transition-colors" />
          </div>
          <button type="submit" :disabled="loading"
            class="w-full py-3.5 text-white rounded-xl text-sm font-bold transition-all shadow-lg disabled:opacity-60"
            :class="portalTheme.btnClass"
          >
            {{ loading ? '登录中...' : '登录' }}
          </button>
        </form>

        <!-- Register Form -->
        <form v-else @submit.prevent="handleRegister" class="space-y-5">
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-1.5 block">用户名</label>
            <input v-model="registerForm.username" type="text" required placeholder="请输入用户名"
              class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:border-indigo-400 transition-colors" />
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-1.5 block">邮箱</label>
            <input v-model="registerForm.email" type="email" required placeholder="请输入邮箱"
              class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:border-indigo-400 transition-colors" />
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-1.5 block">密码</label>
            <input v-model="registerForm.password" type="password" required placeholder="请输入密码 (至少6位)" minlength="6"
              class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:border-indigo-400 transition-colors" />
          </div>
          <button type="submit" :disabled="loading"
            class="w-full py-3.5 text-white rounded-xl text-sm font-bold transition-all shadow-lg disabled:opacity-60"
            :class="portalTheme.btnClass"
          >
            {{ loading ? '注册中...' : '注册' }}
          </button>
        </form>
      </div>

      <!-- Back to portal select -->
      <div class="text-center mt-6">
        <router-link to="/" class="text-sm text-zinc-400 hover:text-zinc-600 transition-colors">← 返回入口选择</router-link>
      </div>

      <!-- Footer -->
      <p class="text-center text-xs text-zinc-400 mt-4">智聘 AI &copy; 2025 · AI 驱动面试评估平台</p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '../stores/user'
import { ElMessage } from 'element-plus'
import { User, Building2, GraduationCap } from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const activeTab = ref('login')
const loading = ref(false)

// Determine portal from route path
const portal = computed(() => {
  const path = route.path
  if (path.startsWith('/enterprise')) return 'enterprise'
  if (path.startsWith('/university')) return 'university'
  return 'student'
})

const themes = {
  student: {
    title: '智聘 AI · 学生端',
    subtitle: '模拟面试 · 简历匹配 · 成长追踪',
    icon: User,
    bgGradient: 'bg-gradient-to-br from-zinc-50 via-indigo-50/30 to-violet-50/30',
    blob1: 'bg-indigo-200/20',
    blob2: 'bg-violet-200/20',
    logoBg: 'bg-gradient-to-br from-indigo-600 to-violet-600 shadow-indigo-200',
    tabActive: 'border-indigo-600 text-indigo-600',
    btnClass: 'bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-700 hover:to-violet-700 shadow-indigo-200',
  },
  enterprise: {
    title: '智聘 AI · 企业端',
    subtitle: '人才筛选 · 岗位管理 · 数据分析',
    icon: Building2,
    bgGradient: 'bg-gradient-to-br from-zinc-50 via-emerald-50/30 to-teal-50/30',
    blob1: 'bg-emerald-200/20',
    blob2: 'bg-teal-200/20',
    logoBg: 'bg-gradient-to-br from-emerald-600 to-teal-600 shadow-emerald-200',
    tabActive: 'border-emerald-600 text-emerald-600',
    btnClass: 'bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-700 hover:to-teal-700 shadow-emerald-200',
  },
  university: {
    title: '智聘 AI · 高校端',
    subtitle: '学生跟踪 · 就业统计 · 人才推送',
    icon: GraduationCap,
    bgGradient: 'bg-gradient-to-br from-zinc-50 via-amber-50/30 to-orange-50/30',
    blob1: 'bg-amber-200/20',
    blob2: 'bg-orange-200/20',
    logoBg: 'bg-gradient-to-br from-amber-600 to-orange-600 shadow-amber-200',
    tabActive: 'border-amber-600 text-amber-600',
    btnClass: 'bg-gradient-to-r from-amber-600 to-orange-600 hover:from-amber-700 hover:to-orange-700 shadow-amber-200',
  }
}

const portalTheme = computed(() => themes[portal.value])

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
    await userStore.login({ ...loginForm, role: portal.value })
    ElMessage.success('登录成功')
    const portalRoutes = {
      student: '/student/dashboard',
      enterprise: '/enterprise/dashboard',
      university: '/university/dashboard'
    }
    router.push(portalRoutes[portal.value])
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
    await userStore.register({ ...registerForm, role: portal.value })
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