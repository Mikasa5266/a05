<template>
  <form class="space-y-5" @submit.prevent="handleSubmit" novalidate>
    <div class="space-y-1.5">
      <label class="text-xs font-semibold tracking-wide text-slate-500">用户名</label>
      <input
        v-model.trim="form.username"
        type="text"
        autocomplete="username"
        placeholder="请输入用户名"
        class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
        :class="inputClass(errors.username)"
      />
      <p v-if="errors.username" class="text-xs text-rose-500">{{ errors.username }}</p>
    </div>

    <div class="space-y-1.5">
      <label class="text-xs font-semibold tracking-wide text-slate-500">真实姓名</label>
      <input
        v-model.trim="form.realName"
        type="text"
        autocomplete="name"
        placeholder="请输入真实姓名"
        class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
        :class="inputClass(errors.realName)"
      />
      <p v-if="errors.realName" class="text-xs text-rose-500">{{ errors.realName }}</p>
    </div>

    <div class="space-y-1.5">
      <label class="text-xs font-semibold tracking-wide text-slate-500">手机号</label>
      <input
        v-model.trim="form.phone"
        type="tel"
        autocomplete="tel"
        placeholder="请输入 11 位手机号"
        class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
        :class="inputClass(errors.phone)"
      />
      <p v-if="errors.phone" class="text-xs text-rose-500">{{ errors.phone }}</p>
    </div>

    <div class="space-y-1.5">
      <label class="text-xs font-semibold tracking-wide text-slate-500">身份证号</label>
      <input
        v-model.trim="form.idCardNo"
        type="text"
        autocomplete="off"
        placeholder="请输入 18 位身份证号"
        class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
        :class="inputClass(errors.idCardNo)"
      />
      <p v-if="errors.idCardNo" class="text-xs text-rose-500">{{ errors.idCardNo }}</p>
    </div>

    <div class="space-y-1.5">
      <label class="text-xs font-semibold tracking-wide text-slate-500">邮箱</label>
      <input
        v-model.trim="form.email"
        type="email"
        autocomplete="email"
        placeholder="请输入邮箱"
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
        autocomplete="new-password"
        placeholder="至少 6 位"
        class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
        :class="inputClass(errors.password)"
      />
      <p v-if="errors.password" class="text-xs text-rose-500">{{ errors.password }}</p>
    </div>

    <div class="space-y-1.5">
      <label class="text-xs font-semibold tracking-wide text-slate-500">确认密码</label>
      <input
        v-model="form.confirmPassword"
        type="password"
        autocomplete="new-password"
        placeholder="请再次输入密码"
        class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
        :class="inputClass(errors.confirmPassword)"
      />
      <p v-if="errors.confirmPassword" class="text-xs text-rose-500">{{ errors.confirmPassword }}</p>
    </div>

    <button
      type="submit"
      :disabled="loading"
      class="group relative mt-2 w-full cursor-pointer overflow-hidden rounded-xl bg-linear-to-r from-cyan-600 to-blue-600 px-4 py-3.5 text-sm font-semibold text-white shadow-lg shadow-cyan-500/25 transition hover:from-cyan-500 hover:to-blue-500 disabled:cursor-not-allowed disabled:opacity-60"
    >
      <span class="relative z-10">{{ loading ? '注册中...' : '创建学生账号' }}</span>
      <span class="absolute inset-0 bg-white/0 transition group-hover:bg-white/10"></span>
    </button>
  </form>
</template>

<script setup>
import { reactive } from 'vue'

defineProps({
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['submit'])

const form = reactive({
  username: '',
  realName: '',
  phone: '',
  idCardNo: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const errors = reactive({
  username: '',
  realName: '',
  phone: '',
  idCardNo: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const chinaPhonePattern = /^1[3-9]\d{9}$/
const chinaIDCardPattern = /^\d{17}[\dXx]$/
const idCardWeights = [7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2]
const idCardCodes = '10X98765432'

const isValidIDCardChecksum = (value = '') => {
  const normalized = String(value || '').trim().toUpperCase()
  if (!chinaIDCardPattern.test(normalized)) return false

  let sum = 0
  for (let i = 0; i < 17; i += 1) {
    sum += Number(normalized[i]) * idCardWeights[i]
  }

  return idCardCodes[sum % 11] === normalized[17]
}

const inputClass = (hasError) => {
  if (hasError) return 'border-rose-300 focus:border-rose-400 focus:ring-4 focus:ring-rose-100'
  return 'border-cyan-200 focus:border-cyan-400 focus:ring-4 focus:ring-cyan-100'
}

const validate = () => {
  errors.username = ''
  errors.realName = ''
  errors.phone = ''
  errors.idCardNo = ''
  errors.email = ''
  errors.password = ''
  errors.confirmPassword = ''

  const email = String(form.email || '').trim()
  const realName = String(form.realName || '').trim()
  const phone = String(form.phone || '').trim().replace(/[\s()-]/g, '')
  const idCardNo = String(form.idCardNo || '').trim().toUpperCase()
  const emailOk = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)

  if (!form.username) errors.username = '请输入用户名'
  if (!realName) errors.realName = '请输入真实姓名'
  else if (realName.length < 2) errors.realName = '真实姓名至少 2 个字符'
  if (!phone) errors.phone = '请输入手机号'
  else if (!chinaPhonePattern.test(phone)) errors.phone = '手机号格式不正确'
  if (!idCardNo) errors.idCardNo = '请输入身份证号'
  else if (!isValidIDCardChecksum(idCardNo)) errors.idCardNo = '身份证号格式或校验位不正确'
  if (!email) errors.email = '请输入邮箱'
  else if (!emailOk) errors.email = '邮箱格式不正确'
  if (!form.password) errors.password = '请输入密码'
  else if (form.password.length < 6) errors.password = '密码至少 6 位'
  if (!form.confirmPassword) errors.confirmPassword = '请确认密码'
  else if (form.password !== form.confirmPassword) errors.confirmPassword = '两次输入密码不一致'

  return !errors.username && !errors.realName && !errors.phone && !errors.idCardNo && !errors.email && !errors.password && !errors.confirmPassword
}

const handleSubmit = () => {
  if (!validate()) return
  emit('submit', {
    username: form.username,
    real_name: form.realName,
    phone: form.phone,
    id_card_no: form.idCardNo,
    email: form.email,
    password: form.password,
    role: 'student'
  })
}
</script>
