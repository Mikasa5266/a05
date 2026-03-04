<template>
  <div class="max-w-2xl mx-auto space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900 dark:text-white">设置</h1>
      <p class="text-zinc-500 mt-2">管理您的账户偏好与应用设置</p>
    </header>

    <div class="bg-white dark:bg-zinc-800 rounded-3xl shadow-sm border border-zinc-100 dark:border-zinc-700 overflow-hidden divide-y divide-zinc-100 dark:divide-zinc-700">
      <!-- Profile Section -->
      <div class="p-8">
        <h2 class="text-lg font-bold text-zinc-900 dark:text-white mb-4 flex items-center gap-2">
          <User class="h-5 w-5 text-indigo-600" />
          个人资料
        </h2>
        <div class="space-y-4">
          <div class="flex items-center gap-4">
            <div class="h-16 w-16 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-600 font-bold text-xl overflow-hidden">
              <img v-if="userStore.userInfo?.avatar" :src="avatarUrl" class="w-full h-full object-cover" />
              <span v-else>{{ userStore.userInfo?.username ? userStore.userInfo.username.charAt(0).toUpperCase() : 'U' }}</span>
            </div>
            <div>
              <div class="font-medium text-zinc-900 dark:text-white">{{ userStore.userInfo?.username || 'User' }}</div>
              <div class="text-sm text-zinc-500">{{ userStore.userInfo?.email || 'user@example.com' }}</div>
            </div>
            <input type="file" ref="fileInput" class="hidden" accept="image/*" @change="handleFileChange" />
            <button @click="triggerFileInput" class="ml-auto text-sm text-indigo-600 font-medium hover:underline">
              更换头像
            </button>
          </div>
        </div>
      </div>

      <!-- App Settings -->
      <div class="p-8">
        <h2 class="text-lg font-bold text-zinc-900 dark:text-white mb-4 flex items-center gap-2">
          <SettingsIcon class="h-5 w-5 text-indigo-600" />
          应用偏好
        </h2>
        <div class="space-y-6">
          <div class="flex items-center justify-between">
            <div>
              <div class="font-medium text-zinc-900 dark:text-white">深色模式</div>
              <div class="text-sm text-zinc-500">启用暗黑主题界面</div>
            </div>
            <button 
              class="w-12 h-6 rounded-full bg-zinc-200 dark:bg-zinc-600 relative transition-colors"
              :class="{ 'bg-indigo-600': isDarkMode }"
              @click="toggleDarkMode"
            >
              <div 
                class="absolute top-1 left-1 bg-white w-4 h-4 rounded-full transition-transform"
                :class="{ 'translate-x-6': isDarkMode }"
              ></div>
            </button>
          </div>

          <div class="flex items-center justify-between opacity-50 cursor-not-allowed" title="暂未开放">
            <div>
              <div class="font-medium text-zinc-900 dark:text-white">面试音效</div>
              <div class="text-sm text-zinc-500">播放 AI 语音反馈</div>
            </div>
            <button 
              class="w-12 h-6 rounded-full bg-indigo-600 relative transition-colors"
              disabled
            >
              <div class="absolute top-1 left-1 bg-white w-4 h-4 rounded-full translate-x-6"></div>
            </button>
          </div>
        </div>
      </div>

      <!-- Account Actions -->
      <div class="p-8 bg-zinc-50 dark:bg-zinc-900/50">
        <h2 class="text-lg font-bold text-zinc-900 dark:text-white mb-4 flex items-center gap-2">
          <Shield class="h-5 w-5 text-indigo-600" />
          账户安全
        </h2>
        <div class="space-y-4">
          <button 
            @click="showPasswordModal = true"
            class="w-full text-left px-4 py-3 bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-xl text-sm font-medium text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50 dark:hover:bg-zinc-700 transition-colors"
          >
            修改密码
          </button>
          
          <button 
            @click="handleLogout"
            class="w-full text-left px-4 py-3 bg-white dark:bg-zinc-800 border border-rose-200 dark:border-rose-900/30 rounded-xl text-sm font-medium text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-900/10 transition-colors flex items-center justify-between group"
          >
            <span>退出登录</span>
            <LogOut class="h-4 w-4 group-hover:translate-x-1 transition-transform" />
          </button>
        </div>
      </div>
    </div>

    <!-- Password Modal -->
    <div v-if="showPasswordModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
      <div class="bg-white dark:bg-zinc-800 rounded-2xl p-6 w-96 shadow-xl animate-in fade-in zoom-in duration-200">
        <h3 class="text-lg font-bold mb-4 text-zinc-900 dark:text-white">修改密码</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-xs font-medium text-zinc-500 mb-1">当前密码</label>
            <input v-model="passwordForm.oldPassword" type="password" class="w-full px-4 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-900 text-zinc-900 dark:text-white focus:ring-2 focus:ring-indigo-500/20 outline-none transition-all" />
          </div>
          <div>
            <label class="block text-xs font-medium text-zinc-500 mb-1">新密码</label>
            <input v-model="passwordForm.newPassword" type="password" class="w-full px-4 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-900 text-zinc-900 dark:text-white focus:ring-2 focus:ring-indigo-500/20 outline-none transition-all" />
          </div>
          <div>
            <label class="block text-xs font-medium text-zinc-500 mb-1">确认新密码</label>
            <input v-model="passwordForm.confirmPassword" type="password" class="w-full px-4 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-900 text-zinc-900 dark:text-white focus:ring-2 focus:ring-indigo-500/20 outline-none transition-all" />
          </div>
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button @click="showPasswordModal = false" class="px-4 py-2 text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-700 rounded-lg transition-colors">取消</button>
          <button @click="handleUpdatePassword" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors shadow-lg shadow-indigo-500/20">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { User, Settings as SettingsIcon, Shield, LogOut } from 'lucide-vue-next'
import { updateAvatar, updatePassword } from '../api/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()
const isDarkMode = ref(document.documentElement.classList.contains('dark'))

const fileInput = ref(null)
const showPasswordModal = ref(false)
const passwordForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

// Construct full avatar URL
const avatarUrl = computed(() => {
  if (!userStore.userInfo?.avatar) return ''
  if (userStore.userInfo.avatar.startsWith('http')) return userStore.userInfo.avatar
  // Assuming backend runs on same host/port or configured base URL
  // If backend is on 8080 and frontend on 5173, we need base URL
  // For now, assume relative path works if proxied, or add base URL
  return `http://localhost:8080${userStore.userInfo.avatar}`
})

const toggleDarkMode = () => {
  isDarkMode.value = !isDarkMode.value
  if (isDarkMode.value) {
    document.documentElement.classList.add('dark')
    localStorage.setItem('theme', 'dark')
  } else {
    document.documentElement.classList.remove('dark')
    localStorage.setItem('theme', 'light')
  }
}

const triggerFileInput = () => {
  fileInput.value.click()
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

  try {
    const res = await updateAvatar(formData)
    userStore.userInfo = res.user
    ElMessage.success('头像更新成功')
  } catch (err) {
    ElMessage.error('头像更新失败: ' + (err.response?.data?.error || err.message))
  }
}

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
    userStore.logout()
    router.push('/login')
  }
}
</script>
