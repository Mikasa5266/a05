<script setup>
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Upload, Loader2, Briefcase, Sparkles, ChevronRight, FileText, CheckCircle2, AlertCircle } from 'lucide-vue-next'
import { parseResume } from '../api/resume'

const props = defineProps({
  settings: {
    type: Object,
    required: true
  },
  isProcessing: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:settings', 'start-interview'])

const positionOptions = ['Java后端工程师', '前端工程师', '算法工程师', 'AI工程师']
const styleOptions = [
  { key: 'gentle', label: '引导型', description: '循序渐进，注重表达舒适度' },
  { key: 'stress', label: '压力型', description: '高压追问，模拟真实面试现场' },
  { key: 'deep', label: '深挖型', description: '聚焦原理与工程细节深度' },
  { key: 'practical', label: '实战型', description: '强调项目落地与问题解决' }
]

const fileInputRef = ref(null)
const parsingResume = ref(false)
const resumeParsed = ref(false)
const resumeError = ref('')
const resumeFileName = ref('')
const resumeData = ref(null)
const matchedRoles = ref([])

const settingsProxy = computed(() => props.settings || {})

const updateSettings = (patch) => {
  emit('update:settings', {
    ...(props.settings || {}),
    interviewMode: 'ai',
    mode: 'technical',
    ...patch
  })
}

const openFilePicker = () => {
  fileInputRef.value?.click()
}

const handleFileSelected = async (event) => {
  const file = event?.target?.files?.[0]
  if (!file) return
  const formData = new FormData()
  formData.append('file', file)
  parsingResume.value = true
  resumeParsed.value = false
  resumeError.value = ''
  resumeFileName.value = file.name
  try {
    const response = await parseResume(formData)
    resumeData.value = response?.resume || null
    matchedRoles.value = Array.isArray(response?.matches) ? response.matches : []
    resumeParsed.value = true
    const topMatch = matchedRoles.value[0]
    if (topMatch?.jobTitle) {
      updateSettings({ position: topMatch.jobTitle })
    } else if (resumeData.value?.intent) {
      updateSettings({ position: resumeData.value.intent })
    }
    ElMessage.success('简历解析完成，可直接开始常规面试')
  } catch (error) {
    const message = error?.response?.data?.error || error?.message || '简历解析失败'
    resumeError.value = String(message)
    resumeParsed.value = false
    resumeData.value = null
    matchedRoles.value = []
  } finally {
    parsingResume.value = false
    if (event?.target) {
      event.target.value = ''
    }
  }
}

const isValid = computed(() => {
  return Boolean(settingsProxy.value.position) && Boolean(settingsProxy.value.style)
})

const startInterview = () => {
  if (!isValid.value) {
    ElMessage.warning('请先完成岗位与面试官风格配置')
    return
  }
  emit('start-interview', {
    initType: 'standard',
    resumeData: resumeData.value,
    matchedRoles: matchedRoles.value
  })
}
</script>

<template>
  <div class="space-y-4 bg-white p-6 rounded-2xl border border-zinc-100 shadow-sm">
    <section class="rounded-xl border border-zinc-100 p-4 space-y-3">
      <p class="text-xs font-bold text-zinc-400 uppercase tracking-wider">Step 1 · 简历识别与岗位建议</p>
      <input ref="fileInputRef" type="file" accept=".pdf,.doc,.docx" class="hidden" @change="handleFileSelected" />
      <button
        type="button"
        class="w-full rounded-xl border border-dashed border-indigo-300 bg-indigo-50/50 px-4 py-4 flex items-center justify-between hover:bg-indigo-50 transition-colors"
        @click="openFilePicker"
      >
        <span class="flex items-center gap-2 text-sm font-medium text-indigo-700">
          <Upload class="w-4 h-4" />
          {{ resumeFileName || '上传简历（PDF / Word）' }}
        </span>
        <Loader2 v-if="parsingResume" class="w-4 h-4 text-indigo-600 animate-spin" />
        <span v-else class="text-xs text-indigo-500">点击上传</span>
      </button>

      <div v-if="resumeParsed" class="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-700 flex items-start gap-2">
        <CheckCircle2 class="w-4 h-4 mt-0.5 shrink-0" />
        <div>
          <p>简历识别通过</p>
          <p v-if="matchedRoles[0]?.jobTitle" class="text-xs mt-1">建议岗位：{{ matchedRoles[0].jobTitle }}</p>
        </div>
      </div>

      <div v-if="resumeError" class="rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700 flex items-start gap-2">
        <AlertCircle class="w-4 h-4 mt-0.5 shrink-0" />
        <span>{{ resumeError }}</span>
      </div>
    </section>

    <section class="rounded-xl border border-zinc-100 p-4 space-y-4">
      <p class="text-xs font-bold text-zinc-400 uppercase tracking-wider">Step 2 · 面试配置</p>

      <div class="space-y-2">
        <label class="text-xs font-semibold text-zinc-500 uppercase tracking-wider flex items-center gap-1.5">
          <Briefcase class="w-3.5 h-3.5" />
          目标岗位
        </label>
        <el-select
          :model-value="settingsProxy.position"
          placeholder="请选择目标岗位"
          class="w-full standard-position-select"
          popper-class="custom-light-select-popper"
          @update:model-value="(value) => updateSettings({ position: value })"
        >
          <el-option v-for="position in positionOptions" :key="position" :label="position" :value="position" />
        </el-select>
      </div>

      <div class="space-y-2">
        <label class="text-xs font-semibold text-zinc-500 uppercase tracking-wider flex items-center gap-1.5">
          <Sparkles class="w-3.5 h-3.5" />
          面试官风格
        </label>
        <div class="grid grid-cols-2 gap-2">
          <button
            v-for="style in styleOptions"
            :key="style.key"
            type="button"
            class="rounded-xl border px-3 py-2.5 text-left transition-colors"
            :class="settingsProxy.style === style.key ? 'border-indigo-300 bg-indigo-50 text-indigo-700' : 'border-zinc-200 bg-white text-zinc-600 hover:bg-zinc-50'"
            @click="updateSettings({ style: style.key })"
          >
            <p class="text-sm font-semibold">{{ style.label }}</p>
            <p class="text-[11px] mt-1 opacity-80">{{ style.description }}</p>
          </button>
        </div>
      </div>
    </section>

    <button
      type="button"
      class="w-full py-3 rounded-xl bg-indigo-600 text-white font-bold hover:bg-indigo-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center justify-center gap-2"
      :disabled="isProcessing || parsingResume || !isValid"
      @click="startInterview"
    >
      <Loader2 v-if="isProcessing" class="w-4 h-4 animate-spin" />
      <FileText v-else class="w-4 h-4" />
      <span>{{ isProcessing ? '启动中...' : '开始常规面试' }}</span>
      <ChevronRight v-if="!isProcessing" class="w-4 h-4" />
    </button>
  </div>
</template>
