<script setup>
import { ref } from 'vue'
import { Upload, Sparkles, BrainCircuit, Search, CheckCircle2 } from 'lucide-vue-next'
import { parseResume } from '../api/resume'
import { useRouter } from 'vue-router'

const router = useRouter()
const fileInput = ref(null)
const isUploading = ref(false)
const resumeData = ref(null)
const matches = ref([])

const triggerFileInput = () => {
  fileInput.value.click()
}

const handleFileChange = (event) => {
  const file = event.target.files[0]
  if (file) processFile(file)
}

const handleDrop = (event) => {
  const file = event.dataTransfer.files[0]
  if (file) processFile(file)
}

const processFile = async (file) => {
  isUploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    
    const res = await parseResume(formData)
    
    // Backend returns { resume: {...}, matches: [...] }
    resumeData.value = res.resume
    matches.value = res.matches || []
    
  } catch (error) {
    console.error('Failed to parse resume:', error)
    alert('简历解析失败: ' + (error.response?.data?.error || error.message))
  } finally {
    isUploading.value = false
  }
}

const resetUpload = () => {
  resumeData.value = null
  matches.value = []
  if (fileInput.value) fileInput.value.value = ''
}

const startInterview = (match) => {
  router.push({
    name: 'MockInterview',
    query: { position: match.jobTitle }
  })
}
</script>

<template>
  <div class="space-y-8">
    <!-- Header -->
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900">简历智能匹配</h1>
      <p class="text-zinc-500 mt-2">上传您的简历，AI 将为您解析核心竞争力并推荐匹配岗位</p>
    </header>

    <!-- Upload Area (when no resume data) -->
    <div 
      v-if="!resumeData"
      class="border-2 border-dashed border-zinc-200 rounded-3xl p-12 flex flex-col items-center justify-center bg-white hover:border-indigo-300 transition-colors cursor-pointer relative"
      @click="triggerFileInput"
      @dragover.prevent
      @drop.prevent="handleDrop"
    >
      <input 
        type="file" 
        ref="fileInput" 
        class="hidden" 
        accept=".pdf,.doc,.docx,.txt"
        @change="handleFileChange"
      />
      
      <div class="h-16 w-16 bg-indigo-50 rounded-2xl text-indigo-600 mb-4 flex items-center justify-center">
        <Sparkles v-if="isUploading" class="h-8 w-8 animate-pulse" />
        <Upload v-else class="h-8 w-8" />
      </div>
      
      <h3 class="text-lg font-medium text-zinc-900">
        {{ isUploading ? '正在解析简历...' : '点击或拖拽文件上传简历' }}
      </h3>
      <p class="text-zinc-500 mt-1 text-sm">支持 PDF, Word 或 TXT 格式</p>
    </div>

    <!-- Results Area (when resume data exists) -->
    <div v-else class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Left Column: Parsing Result -->
      <div class="lg:col-span-1 space-y-6">
        <div class="bg-white rounded-3xl p-6 shadow-sm border border-zinc-100">
          <div class="flex items-center gap-3 mb-6">
            <div class="h-10 w-10 rounded-xl bg-indigo-50 flex items-center justify-center text-indigo-600">
              <BrainCircuit class="h-6 w-6" />
            </div>
            <h3 class="text-lg font-bold text-zinc-900">简历解析结果</h3>
          </div>

          <div class="space-y-6">
            <!-- Tech Stack -->
            <div>
              <h4 class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-3">技术栈</h4>
              <div class="flex flex-wrap gap-2">
                <span 
                  v-for="tech in resumeData.techStack" 
                  :key="tech"
                  class="px-3 py-1 bg-zinc-100 rounded-full text-sm font-medium text-zinc-700"
                >
                  {{ tech }}
                </span>
              </div>
            </div>

            <!-- Intent -->
            <div>
              <h4 class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-3">求职意向</h4>
              <p class="text-zinc-700 font-medium">{{ resumeData.intent }}</p>
            </div>

            <!-- Soft Skills -->
            <div>
              <h4 class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-3">软技能</h4>
              <div class="flex flex-wrap gap-2">
                <span 
                  v-for="skill in resumeData.softSkills" 
                  :key="skill"
                  class="px-3 py-1 bg-emerald-50 text-emerald-700 rounded-full text-sm font-medium"
                >
                  {{ skill }}
                </span>
              </div>
            </div>
          </div>

          <button 
            @click="resetUpload"
            class="w-full mt-8 py-3 border border-zinc-200 rounded-2xl text-sm font-medium text-zinc-600 hover:bg-zinc-50 transition-colors"
          >
            重新上传
          </button>
        </div>
      </div>

      <!-- Right Column: Job Matches -->
      <div class="lg:col-span-2 space-y-6">
        <div class="flex items-center gap-3 mb-2">
          <Search class="h-5 w-5 text-zinc-400" />
          <h3 class="text-lg font-bold text-zinc-900">推荐岗位匹配 (RAG)</h3>
        </div>

        <div class="space-y-4">
          <div 
            v-for="(match, index) in matches" 
            :key="index"
            class="bg-white rounded-3xl p-6 shadow-sm border border-zinc-100 hover:shadow-md transition-shadow"
          >
            <div class="flex items-start justify-between mb-4">
              <div>
                <h4 class="text-lg font-bold text-zinc-900 flex items-center gap-2">
                  {{ match.jobTitle }}
                  <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-600 text-xs font-bold">
                    <CheckCircle2 class="h-3 w-3" />
                    {{ match.matchScore }}% 匹配
                  </span>
                </h4>
              </div>
              <button 
                @click="startInterview(match)"
                class="bg-indigo-600 text-white rounded-xl px-4 py-2 text-sm font-medium hover:bg-indigo-700 transition-colors"
              >
                生成面试题
              </button>
            </div>

            <p class="text-zinc-600 text-sm leading-relaxed mb-4">
              {{ match.reason }}
            </p>

            <div class="flex flex-wrap gap-2">
              <span 
                v-for="req in match.requirements" 
                :key="req"
                class="border border-zinc-200 px-2 py-1 rounded-lg text-xs text-zinc-500"
              >
                {{ req }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
