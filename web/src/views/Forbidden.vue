<template>
  <div class="min-h-screen bg-slate-50 flex items-center justify-center px-6">
    <div class="max-w-xl w-full bg-white rounded-2xl shadow-xl border border-slate-100 p-10 text-center">
      <p class="text-sm font-semibold tracking-wider text-rose-500">403</p>
      <h1 class="mt-2 text-3xl font-bold text-slate-900">无权限访问该页面</h1>
      <p class="mt-4 text-slate-600">
        你当前账号角色与目标页面不匹配，请返回对应角色控制台后重试。
      </p>
      <div class="mt-8 flex items-center justify-center gap-3">
        <button
          type="button"
          class="px-4 py-2 rounded-lg bg-slate-900 text-white hover:bg-slate-800 transition-colors"
          @click="goBack"
        >
          返回上一页
        </button>
        <button
          type="button"
          class="px-4 py-2 rounded-lg border border-slate-200 text-slate-700 hover:bg-slate-50 transition-colors"
          @click="goDashboard"
        >
          前往我的控制台
        </button>
      </div>
      <p v-if="expectedText" class="mt-5 text-xs text-slate-400">
        页面要求角色：{{ expectedText }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const roleDashboard = {
  student: '/student/dashboard',
  enterprise: '/enterprise/dashboard',
  university: '/university/dashboard'
}

const expectedText = computed(() => {
  const expected = String(route.query?.expected || '').trim()
  return expected ? expected : ''
})

const normalizeRole = (role) => {
  if (role === 'enterprise' || role === 'university') return role
  return 'student'
}

const goBack = () => {
  router.back()
}

const goDashboard = () => {
  const actualRole = normalizeRole(String(route.query?.actual || 'student'))
  router.push(roleDashboard[actualRole])
}
</script>
