<template>
  <form class="space-y-5" @submit.prevent="handleSubmit" novalidate>
    <div class="space-y-1.5">
      <label class="text-xs font-semibold tracking-wide text-slate-500">邮箱</label>
      <input
        v-model.trim="form.email"
        type="email"
        autocomplete="email"
        placeholder="name@company.com"
        class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
        :class="inputClass(errors.email)"
      />
      <p v-if="errors.email" class="text-xs text-rose-500">{{ errors.email }}</p>
    </div>

    <div class="space-y-1.5">
      <label class="text-xs font-semibold tracking-wide text-slate-500">密码</label>
      <input
        v-model="form.password"
        type="password"
        autocomplete="current-password"
        placeholder="请输入密码"
        class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
        :class="inputClass(errors.password)"
      />
      <p v-if="errors.password" class="text-xs text-rose-500">{{ errors.password }}</p>
    </div>

    <p v-if="serverError" class="rounded-xl border border-rose-200 bg-rose-50 px-3 py-2 text-xs text-rose-600">
      {{ serverError }}
    </p>

    <button
      type="submit"
      :disabled="loading"
      class="group relative mt-2 w-full cursor-pointer overflow-hidden rounded-xl px-4 py-3.5 text-sm font-semibold text-white transition disabled:cursor-not-allowed disabled:opacity-60"
      :class="buttonClass"
    >
      <span class="relative z-10">{{ loading ? '登录中...' : '登录' }}</span>
      <span class="absolute inset-0 bg-white/0 transition group-hover:bg-white/10"></span>
    </button>
  </form>
</template>

<script setup>
import { computed, reactive, watch } from 'vue'

const props = defineProps({
  role: {
    type: String,
    default: 'student'
  },
  loading: {
    type: Boolean,
    default: false
  },
  initialEmail: {
    type: String,
    default: ''
  },
  serverError: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['submit', 'clear-error'])

const form = reactive({
  email: props.initialEmail || '',
  password: ''
})

const errors = reactive({
  email: '',
  password: ''
})

watch(
  () => props.initialEmail,
  (next) => {
    if (next) form.email = next
  }
)

watch(
  () => [form.email, form.password],
  () => {
    if (props.serverError) {
      emit('clear-error')
    }
  }
)

const roleStyles = {
  student: {
    inputFocus: 'border-cyan-200 focus:border-cyan-400 focus:ring-4 focus:ring-cyan-100',
    button: 'bg-gradient-to-r from-cyan-600 to-blue-600 shadow-lg shadow-cyan-500/25 hover:from-cyan-500 hover:to-blue-500'
  },
  enterprise: {
    inputFocus: 'border-emerald-200 focus:border-emerald-400 focus:ring-4 focus:ring-emerald-100',
    button: 'bg-gradient-to-r from-emerald-600 to-teal-600 shadow-lg shadow-emerald-500/25 hover:from-emerald-500 hover:to-teal-500'
  },
  university: {
    inputFocus: 'border-amber-200 focus:border-amber-400 focus:ring-4 focus:ring-amber-100',
    button: 'bg-gradient-to-r from-amber-600 to-orange-600 shadow-lg shadow-amber-500/25 hover:from-amber-500 hover:to-orange-500'
  }
}

const currentStyle = computed(() => roleStyles[props.role] || roleStyles.student)

const buttonClass = computed(() => currentStyle.value.button)

const inputClass = (hasError) => {
  if (hasError) {
    return 'border-rose-300 focus:border-rose-400 focus:ring-4 focus:ring-rose-100'
  }
  return currentStyle.value.inputFocus
}

const validate = () => {
  errors.email = ''
  errors.password = ''

  const email = String(form.email || '').trim()
  const password = String(form.password || '')
  const emailOk = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)

  if (!email) errors.email = '请输入邮箱'
  else if (!emailOk) errors.email = '邮箱格式不正确'

  if (!password) errors.password = '请输入密码'
  else if (password.length < 6) errors.password = '密码至少 6 位'

  return !errors.email && !errors.password
}

const handleSubmit = () => {
  if (!validate() || props.loading) return
  emit('submit', {
    email: form.email,
    password: form.password
  })
}
</script>
