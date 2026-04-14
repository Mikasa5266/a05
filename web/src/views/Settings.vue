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
            <div class="h-16 w-16 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-600 font-bold text-xl overflow-hidden">
              <img v-if="portalUserInfo?.avatar" :src="avatarUrl" class="w-full h-full object-cover" />
              <span v-else>{{ userInitial }}</span>
            </div>
            <div>
              <div class="font-medium text-zinc-900">{{ displayName }}</div>
              <div class="text-sm text-zinc-500">{{ displayEmail }}</div>
            </div>
            <input type="file" ref="fileInput" class="hidden" accept="image/*" @change="handleFileChange" />
            <button
              @click="triggerFileInput"
              class="ml-auto text-sm text-indigo-600 font-medium hover:underline disabled:text-zinc-400 disabled:no-underline"
              :disabled="profileSyncing || avatarUploading"
            >
              {{ avatarUploading ? '上传中...' : '更换头像' }}
            </button>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-2">
            <div>
              <label class="block text-xs font-medium text-zinc-500 mb-1">昵称</label>
              <input
                v-model="profileForm.username"
                class="w-full px-4 py-2 border border-zinc-200 rounded-lg bg-zinc-50 text-zinc-900 focus:ring-2 focus:ring-indigo-500/20 outline-none transition-all"
                placeholder="请输入昵称"
                :disabled="profileSyncing || profileSubmitting"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-zinc-500 mb-1">邮箱</label>
              <input
                v-model="profileForm.email"
                class="w-full px-4 py-2 border border-zinc-200 rounded-lg bg-zinc-50 text-zinc-900 focus:ring-2 focus:ring-indigo-500/20 outline-none transition-all"
                placeholder="请输入邮箱"
                :disabled="profileSyncing || profileSubmitting"
              />
            </div>
          </div>

          <div class="flex justify-end">
            <button
              type="button"
              class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors shadow-lg shadow-indigo-500/20 disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="!canSubmitProfile || profileSyncing || profileSubmitting"
              @click="handleUpdateProfile"
            >
              {{ profileSubmitting ? '保存中...' : '保存资料' }}
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
          <div class="flex items-center justify-between opacity-50 cursor-not-allowed" title="暂未开放">
            <div>
              <div class="font-medium text-zinc-900">面试音效</div>
              <div class="text-sm text-zinc-500">播放 AI 语音反馈</div>
            </div>
            <button
              class="w-12 h-6 rounded-full bg-indigo-600 relative transition-colors"
              disabled
            >
              <div class="absolute top-1 left-1 bg-white w-4 h-4 rounded-full translate-x-6"></div>
            </button>
          </div>

          <p class="text-xs text-zinc-400">主题切换已移除，当前版本统一使用全局主题，避免页面间样式不一致。</p>
        </div>
      </div>

      <!-- Account Actions -->
      <div class="p-8 bg-zinc-50">
        <h2 class="text-lg font-bold text-zinc-900 mb-4 flex items-center gap-2">
          <Shield class="h-5 w-5 text-indigo-600" />
          账户安全
        </h2>
        <div class="space-y-4">
          <button
            @click="showPasswordModal = true"
            class="w-full text-left px-4 py-3 bg-white border border-zinc-200 rounded-xl text-sm font-medium text-zinc-700 hover:bg-zinc-50 transition-colors"
          >
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

    <!-- Password Modal -->
    <div v-if="showPasswordModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
      <div class="bg-white rounded-2xl p-6 w-96 shadow-xl animate-in fade-in zoom-in duration-200">
        <h3 class="text-lg font-bold mb-4 text-zinc-900">修改密码</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-xs font-medium text-zinc-500 mb-1">当前密码</label>
            <input v-model="passwordForm.oldPassword" type="password" class="w-full px-4 py-2 border border-zinc-200 rounded-lg bg-zinc-50 text-zinc-900 focus:ring-2 focus:ring-indigo-500/20 outline-none transition-all" />
          </div>
          <div>
            <label class="block text-xs font-medium text-zinc-500 mb-1">新密码</label>
            <input v-model="passwordForm.newPassword" type="password" class="w-full px-4 py-2 border border-zinc-200 rounded-lg bg-zinc-50 text-zinc-900 focus:ring-2 focus:ring-indigo-500/20 outline-none transition-all" />
          </div>
          <div>
            <label class="block text-xs font-medium text-zinc-500 mb-1">确认新密码</label>
            <input v-model="passwordForm.confirmPassword" type="password" class="w-full px-4 py-2 border border-zinc-200 rounded-lg bg-zinc-50 text-zinc-900 focus:ring-2 focus:ring-indigo-500/20 outline-none transition-all" />
          </div>
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button @click="showPasswordModal = false" class="px-4 py-2 text-zinc-500 hover:bg-zinc-100 rounded-lg transition-colors">取消</button>
          <button @click="handleUpdatePassword" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors shadow-lg shadow-indigo-500/20">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { resolveRoleFromPath, useUserStore } from '../stores/user'
import { getBackendAssetUrl } from '../utils/backend'
import { User, Settings as SettingsIcon, Shield, LogOut } from 'lucide-vue-next'
import { updateAvatar, updatePassword, updateUserProfile } from '../api/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const fileInput = ref(null)
const showPasswordModal = ref(false)
const profileSyncing = ref(false)
const profileSubmitting = ref(false)
const avatarUploading = ref(false)
const passwordForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })
const profileForm = reactive({
  username: '',
  email: ''
})

const currentRole = computed(() => {
  const role = resolveRoleFromPath(route.path)
  return role || 'student'
})

const roleLabelMap = {
  student: '学生用户',
  enterprise: '企业用户',
  university: '高校用户'
}

const portalUserInfo = computed(() => userStore.getUserInfoByRole(currentRole.value))

const canSubmitProfile = computed(() => {
  const username = String(profileForm.username || '').trim()
  const email = String(profileForm.email || '').trim()
  return Boolean(username) && Boolean(email)
})

const displayName = computed(() => {
  if (profileSyncing.value && !portalUserInfo.value) return '正在同步账户信息'
  const username = String(portalUserInfo.value?.username || '').trim()
  if (username) return username
  const email = String(portalUserInfo.value?.email || '').trim()
  if (email) return email.split('@')[0] || email
  const label = roleLabelMap[currentRole.value] || roleLabelMap.student
  return label
})

const displayEmail = computed(() => {
  if (profileSyncing.value && !portalUserInfo.value) return '正在同步账户信息'
  const email = String(portalUserInfo.value?.email || '').trim()
  return email || '未设置邮箱'
})

const userInitial = computed(() => {
  const text = displayName.value || 'U'
  return text.substring(0, 1).toUpperCase()
})

const avatarUrl = computed(() => {
  if (!portalUserInfo.value?.avatar) return ''
  return getBackendAssetUrl(portalUserInfo.value.avatar)
})

const resolveUserPayload = (payload) => {
  if (!payload || typeof payload !== 'object') return null
  const wrapped = payload?.user || payload?.data?.user
  if (wrapped && typeof wrapped === 'object') return wrapped
  if (Object.prototype.hasOwnProperty.call(payload, 'id') ||
      Object.prototype.hasOwnProperty.call(payload, 'username') ||
      Object.prototype.hasOwnProperty.call(payload, 'email')) {
    return payload
  }
  return null
}

const hydrateProfileForm = () => {
  profileForm.username = String(portalUserInfo.value?.username || '').trim()
  profileForm.email = String(portalUserInfo.value?.email || '').trim()
}

const syncUserProfile = async () => {
  if (!userStore.hasValidTokenByRole(currentRole.value)) return

  const roleAuth = userStore.getRoleAuth(currentRole.value)
  if (roleAuth.userInfo && roleAuth.profileLoaded && !roleAuth.profileError) {
    hydrateProfileForm()
    return
  }

  profileSyncing.value = true
  try {
    await userStore.getUserInfo(currentRole.value, {
      force: Boolean(roleAuth.profileError)
    })
    hydrateProfileForm()
  } finally {
    profileSyncing.value = false
  }
}

const handleUpdateProfile = async () => {
  const username = String(profileForm.username || '').trim()
  const email = String(profileForm.email || '').trim()

  if (!username || !email) {
    ElMessage.warning('昵称和邮箱不能为空')
    return
  }

  profileSubmitting.value = true
  try {
    const res = await updateUserProfile({ username, email })
    const nextUser = resolveUserPayload(res)
    if (nextUser) {
      userStore.setUserInfoByRole(currentRole.value, nextUser)
    } else {
      await userStore.getUserInfo(currentRole.value, { force: true })
    }
    hydrateProfileForm()
    ElMessage.success('个人资料更新成功')
  } catch (err) {
    ElMessage.error('个人资料更新失败: ' + (err.response?.data?.error || err.message))
  } finally {
    profileSubmitting.value = false
  }
}

const triggerFileInput = () => {
  if (profileSyncing.value || avatarUploading.value) return
  fileInput.value?.click()
}

const handleFileChange = async (e) => {
  const file = e.target.files[0]
  if (!file) return

  // Validate file type/size if needed
  if (file.size > 2 * 1024 * 1024) {
    ElMessage.error('图片大小不能超过 2MB')
    return
  }

  const formData = new FormData()
  formData.append('avatar', file)

  avatarUploading.value = true
  try {
    const res = await updateAvatar(formData)
    const nextUser = resolveUserPayload(res)
    if (nextUser) {
      userStore.setUserInfoByRole(currentRole.value, nextUser)
    } else {
      await userStore.getUserInfo(currentRole.value, { force: true })
    }
    hydrateProfileForm()
    ElMessage.success('头像更新成功')
  } catch (err) {
    ElMessage.error('头像更新失败: ' + (err.response?.data?.error || err.message))
  } finally {
    avatarUploading.value = false
    e.target.value = ''
  }
}

watch(
  () => [currentRole.value, portalUserInfo.value?.username, portalUserInfo.value?.email],
  () => {
    hydrateProfileForm()
  },
  { immediate: true }
)

const handleUpdatePassword = async () => {
  if (!passwordForm.oldPassword || !passwordForm.newPassword) {
    ElMessage.warning('请填写完整')
    return
  }
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }
  
  try {
    await updatePassword({
      old_password: passwordForm.oldPassword,
      new_password: passwordForm.newPassword
    })
    ElMessage.success('密码修改成功')
    showPasswordModal.value = false
    passwordForm.oldPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
  } catch (err) {
    ElMessage.error('密码修改失败: ' + (err.response?.data?.error || err.message))
  }
}

const handleLogout = () => {
  if (confirm('确定要退出登录吗？')) {
    const role = currentRole.value
    userStore.logout(role)
    router.push(`/${role}/login`)
  }
}

onMounted(() => {
  if (typeof window !== 'undefined') {
    document.documentElement.classList.remove('dark')
    window.localStorage.removeItem('theme')
  }
  void syncUserProfile()
})
</script>
