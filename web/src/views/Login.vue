<template>
  <div class="auth-page relative min-h-screen overflow-hidden bg-slate-50 text-slate-900">
    <div class="pointer-events-none absolute inset-0 opacity-80 auth-grid"></div>
    <div class="pointer-events-none absolute -left-24 -top-24 h-104 w-104 rounded-full bg-blue-400/20 blur-[110px]"></div>
    <div class="pointer-events-none absolute -right-24 -bottom-32 h-112 w-md rounded-full bg-teal-400/20 blur-[120px]"></div>

    <div class="relative z-10 mx-auto grid min-h-screen max-w-7xl grid-cols-1 px-4 py-6 lg:grid-cols-2 lg:gap-8 lg:px-8">
      <section class="hidden lg:flex lg:flex-col lg:justify-center lg:pr-4">
        <div class="space-y-6">
          <p class="inline-flex items-center gap-2 rounded-full border border-slate-300/70 bg-white/70 px-4 py-1.5 text-xs tracking-[0.18em] text-cyan-700">
            AI INTERVIEW OS
          </p>
          <h1 class="mt-8 max-w-xl text-5xl font-bold leading-tight text-slate-900">
            让招聘与求职
            <span class="bg-linear-to-r from-cyan-600 via-teal-500 to-emerald-600 bg-clip-text text-transparent">都更像一次智能协同</span>
          </h1>
          <p class="mt-6 max-w-lg text-base leading-7 text-slate-600">
            一体化覆盖 AI 模拟面试、简历评估、人才筛选与高校就业管理。用统一平台连接学生、企业与高校三端，让每一次决策有据可依。
          </p>

          <div class="overflow-hidden rounded-3xl border border-slate-200/70 bg-white/55 p-5 backdrop-blur">
            <div class="feature-track flex min-w-max gap-4">
              <article class="w-64 shrink-0 rounded-2xl border border-slate-200/60 bg-white/60 p-4">
                <p class="text-sm font-semibold text-slate-800">智能简历解析</p>
                <p class="mt-2 text-xs leading-6 text-slate-500">多维抽取候选人教育、项目与能力图谱，自动生成结构化人才档案。</p>
              </article>
              <article class="w-64 shrink-0 rounded-2xl border border-slate-200/60 bg-white/60 p-4">
                <p class="text-sm font-semibold text-slate-800">多模态面试评估</p>
                <p class="mt-2 text-xs leading-6 text-slate-500">结合语音、文本与行为指标，输出可解释的岗位匹配度和改进建议。</p>
              </article>
              <article class="w-64 shrink-0 rounded-2xl border border-slate-200/60 bg-white/60 p-4">
                <p class="text-sm font-semibold text-slate-800">企业级人才库</p>
                <p class="mt-2 text-xs leading-6 text-slate-500">构建可检索、可复用的候选池，支持岗位画像筛选和校企人才协同。</p>
              </article>
              <article class="w-64 shrink-0 rounded-2xl border border-slate-200/60 bg-white/60 p-4">
                <p class="text-sm font-semibold text-slate-800">智能简历解析</p>
                <p class="mt-2 text-xs leading-6 text-slate-500">多维抽取候选人教育、项目与能力图谱，自动生成结构化人才档案。</p>
              </article>
              <article class="w-64 shrink-0 rounded-2xl border border-slate-200/60 bg-white/60 p-4">
                <p class="text-sm font-semibold text-slate-800">多模态面试评估</p>
                <p class="mt-2 text-xs leading-6 text-slate-500">结合语音、文本与行为指标，输出可解释的岗位匹配度和改进建议。</p>
              </article>
              <article class="w-64 shrink-0 rounded-2xl border border-slate-200/60 bg-white/60 p-4">
                <p class="text-sm font-semibold text-slate-800">企业级人才库</p>
                <p class="mt-2 text-xs leading-6 text-slate-500">构建可检索、可复用的候选池，支持岗位画像筛选和校企人才协同。</p>
              </article>
            </div>
          </div>

          <div class="rounded-3xl border border-slate-200/70 bg-white/55 p-5 backdrop-blur">
            <svg viewBox="0 0 520 170" class="h-36 w-full">
              <defs>
                <linearGradient id="aiLine" x1="0" y1="0" x2="1" y2="0">
                  <stop offset="0%" stop-color="#0284c7" stop-opacity="0.35" />
                  <stop offset="50%" stop-color="#0d9488" stop-opacity="0.9" />
                  <stop offset="100%" stop-color="#059669" stop-opacity="0.35" />
                </linearGradient>
              </defs>
              <rect x="8" y="8" width="504" height="154" rx="18" fill="none" stroke="rgba(15,23,42,0.14)" />
              <path d="M24 130 Q88 96 142 120 T258 106 T372 116 T496 72" fill="none" stroke="url(#aiLine)" stroke-width="4" />
              <circle cx="142" cy="120" r="6" fill="#0369a1" />
              <circle cx="258" cy="106" r="6" fill="#0f766e" />
              <circle cx="372" cy="116" r="6" fill="#047857" />
              <text x="24" y="36" fill="rgba(15,23,42,0.88)" font-size="14">面试过程智能评估流</text>
              <text x="24" y="56" fill="rgba(71,85,105,0.9)" font-size="12">Resume Parsing → Dialogue Reasoning → Multimodal Scoring</text>
            </svg>
          </div>
        </div>

      </section>

      <section class="flex items-center justify-center py-6 lg:py-10">
        <div class="w-full max-w-xl rounded-3xl border border-slate-100 bg-white/95 p-6 text-slate-900 shadow-xl shadow-slate-200/50 backdrop-blur-xl sm:p-8">
          <div class="mb-6 flex items-start justify-between gap-4">
            <div>
              <p class="text-xs uppercase tracking-[0.18em] text-slate-500">Welcome back</p>
              <h2 class="mt-2 text-2xl font-bold">{{ currentTheme.title }}</h2>
              <p class="mt-1 text-sm text-slate-500">{{ currentTheme.subtitle }}</p>
            </div>
            <div class="flex h-12 w-12 items-center justify-center rounded-2xl text-white" :class="currentTheme.logoBg">
              <component :is="currentTheme.icon" class="h-6 w-6" />
            </div>
          </div>

          <div class="rounded-2xl border border-slate-200 bg-slate-100/80 p-1">
            <div class="grid grid-cols-3 gap-1">
              <button
                v-for="item in roleOptions"
                :key="item.role"
                type="button"
                class="cursor-pointer flex items-center justify-center gap-2 rounded-xl px-2 py-2 text-xs font-semibold transition sm:text-sm"
                :class="activeRole === item.role ? item.activeClass : 'text-slate-500 hover:bg-slate-200/50 hover:text-slate-700'"
                @click="switchRole(item.role)"
              >
                <component :is="item.icon" class="h-4 w-4" />
                {{ item.label }}
              </button>
            </div>
          </div>

          <div class="mt-5 rounded-2xl border border-slate-200 p-1">
            <div class="grid grid-cols-2 gap-1">
              <button
                v-for="item in modeOptions"
                :key="item.value"
                type="button"
                class="cursor-pointer rounded-xl px-4 py-2.5 text-sm font-semibold transition"
                :class="mode === item.value ? currentTheme.modeActive : 'text-slate-500 hover:bg-slate-100 hover:text-slate-800'"
                @click="mode = item.value"
              >
                {{ item.label }}
              </button>
            </div>
          </div>

          <div class="mt-6">
            <Transition name="auth-switch" mode="out-in">
              <LoginForm
                v-if="mode === 'login'"
                key="login"
                :role="activeRole"
                :loading="loading.login"
                :initial-email="prefillEmail"
                @submit="handleLogin"
              />

              <StudentRegisterForm
                v-else-if="activeRole === 'student'"
                key="student-register"
                :loading="loading.register"
                @submit="handleStudentRegister"
              />

              <EnterpriseApplyForm
                v-else
                key="org-apply"
                :role="activeRole"
                :loading="loading.register"
                @submit="handleOrganizationApply"
              />
            </Transition>
          </div>

          <div class="mt-6 flex items-center justify-end text-xs text-slate-500">
            <span>智聘 AI © 2026</span>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Building2, GraduationCap, User } from 'lucide-vue-next'

import { applyEnterprise, applyUniversity } from '../api/auth'
import LoginForm from '../components/auth/LoginForm.vue'
import StudentRegisterForm from '../components/auth/StudentRegisterForm.vue'
import EnterpriseApplyForm from '../components/auth/EnterpriseApplyForm.vue'
import { useUserStore } from '../stores/user'

const rolePathMap = {
  student: '/student/login',
  enterprise: '/enterprise/login',
  university: '/university/login'
}

const roleDashboard = {
  student: '/student/dashboard',
  enterprise: '/enterprise/dashboard',
  university: '/university/dashboard'
}

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const resolveRoleByPath = (path = '') => {
  if (path.startsWith('/enterprise')) return 'enterprise'
  if (path.startsWith('/university')) return 'university'
  return 'student'
}

const activeRole = ref(resolveRoleByPath(route.path))
const mode = ref('login')
const prefillEmail = ref('')

const loading = reactive({
  login: false,
  register: false
})

watch(
  () => route.path,
  (nextPath) => {
    activeRole.value = resolveRoleByPath(nextPath)
  }
)

const roleOptions = [
  {
    role: 'student',
    label: '学生',
    icon: User,
    activeClass: 'bg-cyan-600 text-white shadow-sm shadow-cyan-500/30'
  },
  {
    role: 'enterprise',
    label: '企业',
    icon: Building2,
    activeClass: 'bg-emerald-600 text-white shadow-sm shadow-emerald-500/30'
  },
  {
    role: 'university',
    label: '高校',
    icon: GraduationCap,
    activeClass: 'bg-amber-600 text-white shadow-sm shadow-amber-500/30'
  }
]

const currentTheme = computed(() => {
  const themes = {
    student: {
      title: '学生端认证',
      subtitle: '模拟面试、成长追踪与求职分析',
      icon: User,
      logoBg: 'bg-gradient-to-br from-cyan-600 to-blue-600',
      modeActive: 'bg-cyan-600 text-white'
    },
    enterprise: {
      title: '企业端认证',
      subtitle: '人才筛选、岗位管理与面试协同',
      icon: Building2,
      logoBg: 'bg-gradient-to-br from-emerald-600 to-teal-600',
      modeActive: 'bg-emerald-600 text-white'
    },
    university: {
      title: '高校端认证',
      subtitle: '就业数据、学生跟踪与人才推送',
      icon: GraduationCap,
      logoBg: 'bg-gradient-to-br from-amber-600 to-orange-600',
      modeActive: 'bg-amber-600 text-white'
    }
  }

  return themes[activeRole.value] || themes.student
})

const modeOptions = computed(() => {
  if (mode.value === 'login') {
    return [
      { value: 'login', label: '登录' },
      { value: 'register', label: activeRole.value === 'student' ? '注册' : '入驻申请' }
    ]
  }

  return [
    { value: 'login', label: '登录' },
    { value: 'register', label: activeRole.value === 'student' ? '注册' : '入驻申请' }
  ]
})

const switchRole = async (role) => {
  if (role === activeRole.value) return
  activeRole.value = role
  mode.value = 'login'
  await router.push({
    path: rolePathMap[role],
    query: route.query
  })
}

const handleLogin = async (payload) => {
  loading.login = true
  try {
    await userStore.login({ ...payload, role: activeRole.value })
    ElMessage.success('登录成功')

    const redirectTarget = String(route.query?.redirect || '').trim()
    if (redirectTarget) {
      router.push(redirectTarget)
      return
    }

    router.push(roleDashboard[activeRole.value])
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '登录失败，请检查账号信息')
  } finally {
    loading.login = false
  }
}

const handleStudentRegister = async (payload) => {
  loading.register = true
  try {
    await userStore.register(payload)
    ElMessage.success('注册成功，请登录')
    prefillEmail.value = payload.email
    mode.value = 'login'
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '注册失败，请稍后重试')
  } finally {
    loading.register = false
  }
}

const handleOrganizationApply = async (payload, email) => {
  loading.register = true
  try {
    if (activeRole.value === 'enterprise') {
      await applyEnterprise(payload)
      ElMessage.success('企业入驻申请已提交，等待审核')
    } else {
      await applyUniversity(payload)
      ElMessage.success('高校入驻申请已提交，等待审核')
    }
    prefillEmail.value = email || ''
    mode.value = 'login'
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '提交失败，请稍后重试')
  } finally {
    loading.register = false
  }
}
</script>

<style scoped>
.auth-grid {
  background-image:
    linear-gradient(rgba(0, 0, 0, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 0, 0, 0.04) 1px, transparent 1px);
  background-size: 42px 42px;
  mask-image: radial-gradient(circle at center, black 42%, transparent 100%);
}

.auth-switch-enter-active,
.auth-switch-leave-active {
  transition: all 0.24s ease;
}

.auth-switch-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.auth-switch-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.feature-track {
  animation: feature-slide 20s linear infinite;
}

@keyframes feature-slide {
  0% {
    transform: translateX(0);
  }
  100% {
    transform: translateX(-50%);
  }
}
</style>