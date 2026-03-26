<template>
  <form class="space-y-4" @submit.prevent="handleSubmit" novalidate>
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <div class="space-y-1.5 sm:col-span-2">
        <label class="text-xs font-semibold tracking-wide text-slate-500">账号名称</label>
        <input
          v-model.trim="form.username"
          type="text"
          autocomplete="username"
          placeholder="请输入登录账号名"
          class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
          :class="inputClass(errors.username)"
        />
        <p v-if="errors.username" class="text-xs text-rose-500">{{ errors.username }}</p>
      </div>

      <div class="space-y-1.5 sm:col-span-2">
        <label class="text-xs font-semibold tracking-wide text-slate-500">邮箱</label>
        <input
          v-model.trim="form.email"
          type="email"
          autocomplete="email"
          placeholder="用于接收审核通知"
          class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
          :class="inputClass(errors.email)"
        />
        <p v-if="errors.email" class="text-xs text-rose-500">{{ errors.email }}</p>
      </div>

      <div class="space-y-1.5 sm:col-span-2">
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

      <template v-if="isEnterprise">
        <div class="space-y-1.5 sm:col-span-2">
          <label class="text-xs font-semibold tracking-wide text-slate-500">企业名称</label>
          <input
            v-model.trim="form.companyName"
            type="text"
            placeholder="请输入企业全称"
            class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
            :class="inputClass(errors.companyName)"
          />
          <p v-if="errors.companyName" class="text-xs text-rose-500">{{ errors.companyName }}</p>
        </div>

        <div class="space-y-1.5">
          <label class="text-xs font-semibold tracking-wide text-slate-500">统一社会信用代码</label>
          <input
            v-model.trim="form.creditCode"
            type="text"
            placeholder="18 位信用代码"
            class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
            :class="inputClass(errors.creditCode)"
          />
          <p v-if="errors.creditCode" class="text-xs text-rose-500">{{ errors.creditCode }}</p>
        </div>

        <div class="space-y-1.5">
          <label class="text-xs font-semibold tracking-wide text-slate-500">营业执照（可选）</label>
          <input
            type="file"
            accept="image/*,.pdf"
            class="w-full cursor-pointer rounded-xl border border-emerald-200 bg-white/80 px-3 py-2.5 text-sm text-slate-700 outline-none file:mr-3 file:cursor-pointer file:rounded-lg file:border-0 file:bg-emerald-50 file:px-3 file:py-1.5 file:text-emerald-700"
            @change="handleFileChange"
          />
          <p class="text-xs text-slate-400">支持图片/PDF，仅用于完善申请信息（当前版本仅记录文件名）</p>
        </div>

        <div class="space-y-1.5 sm:col-span-2">
          <label class="text-xs font-semibold tracking-wide text-slate-500">业务范围</label>
          <textarea
            v-model.trim="form.businessScope"
            rows="3"
            placeholder="请输入业务描述"
            class="w-full rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
            :class="inputClass(errors.businessScope)"
          ></textarea>
          <p v-if="errors.businessScope" class="text-xs text-rose-500">{{ errors.businessScope }}</p>
        </div>
      </template>

      <template v-else>
        <div class="space-y-1.5 sm:col-span-2">
          <label class="text-xs font-semibold tracking-wide text-slate-500">高校名称</label>
          <input
            v-model.trim="form.universityName"
            type="text"
            placeholder="请输入高校全称"
            class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
            :class="inputClass(errors.universityName)"
          />
          <p v-if="errors.universityName" class="text-xs text-rose-500">{{ errors.universityName }}</p>
        </div>

        <div class="space-y-1.5 sm:col-span-2">
          <label class="text-xs font-semibold tracking-wide text-slate-500">院系/部门</label>
          <input
            v-model.trim="form.department"
            type="text"
            placeholder="例如：就业指导中心"
            class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
            :class="inputClass(errors.department)"
          />
          <p v-if="errors.department" class="text-xs text-rose-500">{{ errors.department }}</p>
        </div>
      </template>

      <div class="space-y-1.5">
        <label class="text-xs font-semibold tracking-wide text-slate-500">联系人</label>
        <input
          v-model.trim="form.contactName"
          type="text"
          placeholder="请输入联系人姓名"
          class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
          :class="inputClass(errors.contactName)"
        />
        <p v-if="errors.contactName" class="text-xs text-rose-500">{{ errors.contactName }}</p>
      </div>

      <div class="space-y-1.5">
        <label class="text-xs font-semibold tracking-wide text-slate-500">联系电话</label>
        <input
          v-model.trim="form.contactPhone"
          type="text"
          placeholder="请输入手机号"
          class="w-full cursor-text rounded-xl border bg-white/80 px-4 py-3 text-sm text-slate-800 outline-none transition-all focus:shadow-sm"
          :class="inputClass(errors.contactPhone)"
        />
        <p v-if="errors.contactPhone" class="text-xs text-rose-500">{{ errors.contactPhone }}</p>
      </div>
    </div>

    <button
      type="submit"
      :disabled="loading"
      class="group relative mt-3 w-full cursor-pointer overflow-hidden rounded-xl px-4 py-3.5 text-sm font-semibold text-white transition disabled:cursor-not-allowed disabled:opacity-60"
      :class="buttonClass"
    >
      <span class="relative z-10">{{ loading ? '提交中...' : submitText }}</span>
      <span class="absolute inset-0 bg-white/0 transition group-hover:bg-white/10"></span>
    </button>
  </form>
</template>

<script setup>
import { computed, reactive } from 'vue'

const props = defineProps({
  role: {
    type: String,
    default: 'enterprise'
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['submit'])

const isEnterprise = computed(() => props.role === 'enterprise')

const form = reactive({
  username: '',
  email: '',
  password: '',
  companyName: '',
  creditCode: '',
  businessScope: '',
  businessLicenseFile: null,
  universityName: '',
  department: '',
  contactName: '',
  contactPhone: ''
})

const errors = reactive({
  username: '',
  email: '',
  password: '',
  companyName: '',
  creditCode: '',
  businessScope: '',
  universityName: '',
  department: '',
  contactName: '',
  contactPhone: ''
})

const submitText = computed(() => (isEnterprise.value ? '提交企业入驻申请' : '提交高校入驻申请'))

const buttonClass = computed(() => {
  if (isEnterprise.value) {
    return 'bg-gradient-to-r from-emerald-600 to-teal-600 shadow-lg shadow-emerald-500/25 hover:from-emerald-500 hover:to-teal-500'
  }
  return 'bg-gradient-to-r from-amber-600 to-orange-600 shadow-lg shadow-amber-500/25 hover:from-amber-500 hover:to-orange-500'
})

const inputClass = (hasError) => {
  if (hasError) return 'border-rose-300 focus:border-rose-400 focus:ring-4 focus:ring-rose-100'
  if (isEnterprise.value) return 'border-emerald-200 focus:border-emerald-400 focus:ring-4 focus:ring-emerald-100'
  return 'border-amber-200 focus:border-amber-400 focus:ring-4 focus:ring-amber-100'
}

const handleFileChange = (event) => {
  const file = event?.target?.files?.[0] || null
  form.businessLicenseFile = file
}

const validate = () => {
  Object.keys(errors).forEach((key) => {
    errors[key] = ''
  })

  const email = String(form.email || '').trim()
  const emailOk = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)

  if (!form.username) errors.username = '请输入账号名称'
  if (!email) errors.email = '请输入邮箱'
  else if (!emailOk) errors.email = '邮箱格式不正确'
  if (!form.password) errors.password = '请输入密码'
  else if (form.password.length < 6) errors.password = '密码至少 6 位'

  if (isEnterprise.value) {
    if (!form.companyName) errors.companyName = '请输入企业名称'
    if (!form.creditCode && !form.businessLicenseFile) {
      errors.creditCode = '请填写统一信用代码，或上传营业执照'
    }
  } else {
    if (!form.universityName) errors.universityName = '请输入高校名称'
  }

  if (!form.contactName) errors.contactName = '请输入联系人'
  if (!form.contactPhone) errors.contactPhone = '请输入联系电话'

  return !Object.values(errors).some(Boolean)
}

const handleSubmit = () => {
  if (!validate() || props.loading) return

  if (isEnterprise.value) {
    const scopeParts = [form.businessScope]
    if (form.creditCode) scopeParts.push(`统一社会信用代码:${form.creditCode}`)
    if (form.businessLicenseFile?.name) scopeParts.push(`营业执照文件:${form.businessLicenseFile.name}`)

    emit(
      'submit',
      {
        username: form.username,
        email: form.email,
        password: form.password,
        company_name: form.companyName,
        contact_name: form.contactName,
        contact_phone: form.contactPhone,
        business_scope: scopeParts.filter(Boolean).join(' | ')
      },
      form.email
    )
    return
  }

  emit(
    'submit',
    {
      username: form.username,
      email: form.email,
      password: form.password,
      university_name: form.universityName,
      contact_name: form.contactName,
      contact_phone: form.contactPhone,
      department: form.department
    },
    form.email
  )
}
</script>
