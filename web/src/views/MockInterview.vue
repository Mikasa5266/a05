<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { 
  Video, VideoOff, Mic, MicOff, ChevronRight, 
  BrainCircuit, User, LogOut, Send, Loader2 
} from 'lucide-vue-next'
import { startInterview as apiStartInterview, submitAnswer as apiSubmitAnswer, endInterview as apiEndInterview } from '../api/interview'
import { generateReport as apiGenerateReport } from '../api/report'

const route = useRoute()
const router = useRouter()
const phase = ref('setup') // setup, interview, summary
const isCameraOn = ref(true)
const isMicOn = ref(true)
const previewVideo = ref(null)
const interviewVideo = ref(null)
const stream = ref(null)

// Interview State
const interviewId = ref(null)
const questions = ref([])
const currentQuestionIndex = ref(0)
const messages = ref([])
const userInput = ref('')
const isProcessing = ref(false)
const reportId = ref(null)
const isGeneratingReport = ref(false)

const settings = ref({
  position: route.query.position || 'Java后端开发',
  difficulty: 'Junior',
  mode: 'technical',
  style: 'gentle'
})

// Camera Logic
const toggleCamera = async () => {
  if (isCameraOn.value) {
    stopCamera()
  } else {
    await startCamera()
  }
}

const toggleMic = () => {
  isMicOn.value = !isMicOn.value
}

const startCamera = async () => {
  try {
    stream.value = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    isCameraOn.value = true
    nextTick(() => {
      if (previewVideo.value) previewVideo.value.srcObject = stream.value
      if (interviewVideo.value) interviewVideo.value.srcObject = stream.value
    })
  } catch (err) {
    console.error("Camera access denied:", err)
    isCameraOn.value = false
  }
}

const stopCamera = () => {
  if (stream.value) {
    stream.value.getTracks().forEach(track => track.stop())
    stream.value = null
  }
  isCameraOn.value = false
}

// Interview Logic
const startInterview = async () => {
  isProcessing.value = true
  try {
    const res = await apiStartInterview({
      position: settings.value.position,
      difficulty: settings.value.difficulty,
      mode: settings.value.mode,
      style: settings.value.style
    })
    
    // Backend returns { message: "...", interview: { ... } }
    // The interview object contains questions array if loaded correctly
    const interview = res.interview
    interviewId.value = interview.id
    const rawQuestions = interview.questions || []
    questions.value = rawQuestions
      .map((item) => {
        const nested = item.question || {}
        return {
          mapId: item.id,
          questionId: item.question_id || nested.id,
          title: nested.title || item.title || '',
          content: nested.content || item.content || '',
          expectedAnswer: nested.expected_answer || item.expected_answer || ''
        }
      })
      .filter((q) => q.questionId && (q.content || q.title))
    
    if (questions.value.length === 0) {
      alert("未获取到面试题目，请重试")
      isProcessing.value = false
      return
    }

    // Switch to interview phase
    phase.value = 'interview'
    currentQuestionIndex.value = 0
    
    // Initialize Chat
    messages.value = [
      {
        role: 'ai',
        content: `你好！我是你的 AI 面试官。我们将进行一场关于 ${settings.value.position} 的 ${settings.value.mode === 'technical' ? '技术' : '综合'} 面试。准备好了吗？让我们开始吧。`
      }
    ]
    
    // Push first question after a short delay
    setTimeout(() => {
      pushAIQuestion(questions.value[0])
      scrollToBottom()
    }, 1000)

    // Handle video transition
    if (isCameraOn.value) {
      // Small delay to ensure DOM is ready
      setTimeout(async () => {
        if (!stream.value) await startCamera()
        else if (interviewVideo.value) interviewVideo.value.srcObject = stream.value
      }, 500)
    }

  } catch (error) {
    console.error('Failed to start interview:', error)
    alert('启动面试失败: ' + (error.response?.data?.error || error.message))
  } finally {
    isProcessing.value = false
  }
}

const pushAIQuestion = (question) => {
  const text = (question?.content || question?.title || '').trim()
  if (!text) return
  messages.value.push({
    role: 'ai',
    content: text,
    type: 'question'
  })
}

const formatFeedback = (feedback) => {
  if (feedback == null) return '回答已提交，建议补充更具体的技术细节。'

  const extractText = (val) => {
    if (!val) return []
    if (typeof val === 'string') {
      const text = val.trim()
      if (!text) return []
      if (text.startsWith('{') || text.startsWith('[')) {
        try {
          return extractText(JSON.parse(text))
        } catch (_) {
          return [text]
        }
      }
      return [text]
    }
    if (Array.isArray(val)) {
      return val.flatMap((item) => extractText(item))
    }
    if (typeof val === 'object') {
      const blocks = []
      if (typeof val.content === 'string' && val.content.trim()) blocks.push(val.content.trim())
      if (Array.isArray(val.suggestions)) {
        val.suggestions.forEach((s) => {
          if (typeof s === 'string' && s.trim()) blocks.push(`建议：${s.trim()}`)
        })
      }
      const keys = ['feedback', 'analysis', 'comment', 'summary', 'advice', 'suggestion', 'message']
      keys.forEach((k) => {
        if (val[k] !== undefined) blocks.push(...extractText(val[k]))
      })
      return blocks
    }
    return []
  }

  const texts = extractText(feedback).filter(Boolean)
  return texts.length > 0 ? texts.join('\n') : '回答已提交，建议补充更具体的技术细节。'
}

const splitFeedbackSections = (text) => {
  const source = (text || '').trim()
  if (!source) {
    return {
      evaluation: '回答已提交，建议补充更具体的技术细节。',
      suggestions: []
    }
  }

  const evalMatch = source.match(/【评价】([\s\S]*?)(?:【建议】|$)/)
  const suggestBlockMatch = source.match(/【建议】([\s\S]*)$/)
  if (evalMatch || suggestBlockMatch) {
    const evaluationText = (evalMatch?.[1] || '').trim() || source
    const suggestionLines = (suggestBlockMatch?.[1] || '')
      .split('\n')
      .map((line) => line.replace(/^[-•\d.)、\s]+/, '').trim())
      .filter(Boolean)
    return {
      evaluation: evaluationText,
      suggestions: suggestionLines
    }
  }

  const lines = source
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
  const evaluationParts = []
  const suggestions = []
  lines.forEach((line) => {
    const normalized = line.replace(/^[-•\d.)\s]+/, '').trim()
    if (/^(建议|改进建议|可优化|下一步|你可以)/.test(normalized)) {
      suggestions.push(normalized.replace(/^建议[:：]?\s*/, ''))
      return
    }
    if (/^(1|2|3|4|5)[.)、]\s*/.test(line) && /建议|改进|优化/.test(normalized)) {
      suggestions.push(normalized)
      return
    }
    if (/^(建议：|建议:)/.test(line)) {
      suggestions.push(line.replace(/^(建议：|建议:)\s*/, '').trim())
      return
    }
    evaluationParts.push(line)
  })
  return {
    evaluation: evaluationParts.join('\n') || source,
    suggestions
  }
}

const sendMessage = async () => {
  if (!userInput.value.trim() || isProcessing.value) return
  
  const answer = userInput.value
  userInput.value = ''
  
  // 1. Add User Message
  messages.value.push({
    role: 'user',
    content: answer
  })
  
  isProcessing.value = true
  
  try {
    const currentQ = questions.value[currentQuestionIndex.value]
    if (!currentQ || !currentQ.questionId) {
      throw new Error('当前题目ID无效，请重新开始面试')
    }
    
    // 2. Submit to Backend
    const res = await apiSubmitAnswer(interviewId.value, {
      question_id: currentQ.questionId,
      answer: answer
    })
    
    const result = res.result
    
    // 3. Add AI Feedback
    const formatted = formatFeedback(result.feedback)
    const feedbackSections = splitFeedbackSections(formatted)
    messages.value.push({
      role: 'ai',
      content: formatted,
      type: 'feedback',
      score: result.score,
      feedbackEvaluation: feedbackSections.evaluation,
      feedbackSuggestions: feedbackSections.suggestions
    })
    
    // 4. Move to Next Question
    if (currentQuestionIndex.value < questions.value.length - 1) {
      currentQuestionIndex.value++
      setTimeout(() => {
        pushAIQuestion(questions.value[currentQuestionIndex.value])
      }, 1500)
    } else {
      setTimeout(() => {
        messages.value.push({
          role: 'ai',
          content: "面试结束！辛苦了。您可以点击下方按钮查看详细报告。",
          type: 'system'
        })
        completeInterview()
      }, 1500)
    }
    
  } catch (error) {
    console.error('Failed to submit answer:', error)
    const errMsg = error?.response?.data?.error || error?.message || '未知错误'
    messages.value.push({
      role: 'system',
      content: `提交答案失败：${errMsg}`,
      type: 'system'
    })
  } finally {
    isProcessing.value = false
    scrollToBottom()
  }
}

const completeInterview = async () => {
  if (isGeneratingReport.value || !interviewId.value) return
  isGeneratingReport.value = true
  try {
    await apiEndInterview(interviewId.value)
    const reportRes = await apiGenerateReport({
      interview_id: interviewId.value
    })
    if (reportRes?.report?.id) {
      reportId.value = reportRes.report.id
    }
    if (!reportId.value) {
      messages.value.push({
        role: 'system',
        content: '报告生成中，请稍后点击“查看面试报告”。',
        type: 'system'
      })
    }
  } catch (error) {
    console.error('Failed to end interview:', error)
    const errMsg = error?.response?.data?.error || error?.message || '未知错误'
    messages.value.push({
      role: 'system',
      content: `报告生成失败：${errMsg}`,
      type: 'system'
    })
  } finally {
    isGeneratingReport.value = false
    scrollToBottom()
  }
}

const viewReport = async () => {
  if (!reportId.value && interviewId.value) {
    try {
      const reportRes = await apiGenerateReport({
        interview_id: interviewId.value
      })
      if (reportRes?.report?.id) {
        reportId.value = reportRes.report.id
      }
    } catch (error) {
      const errMsg = error?.response?.data?.error || error?.message || '未知错误'
      messages.value.push({
        role: 'system',
        content: `获取报告失败：${errMsg}`,
        type: 'system'
      })
      scrollToBottom()
      return
    }
  }
  if (reportId.value) {
    router.push(`/report/${reportId.value}`)
    return
  }
  messages.value.push({
    role: 'system',
    content: '报告暂未生成完成，请稍后再试。',
    type: 'system'
  })
  scrollToBottom()
}

const scrollToBottom = () => {
  nextTick(() => {
    const container = document.getElementById('chat-container')
    if (container) container.scrollTop = container.scrollHeight
  })
}

const endInterviewEarly = async () => {
  if (confirm('确定要结束面试吗？进度将不会保存。')) {
    stopCamera()
    phase.value = 'setup'
    currentQuestionIndex.value = 0
    messages.value = []
    if (interviewId.value) {
        try { await apiEndInterview(interviewId.value) } catch(e){}
    }
  }
}

onMounted(() => {
  startCamera()
})

onUnmounted(() => {
  stopCamera()
})
</script>

<template>
  <div class="h-[calc(100vh-8rem)] flex flex-col">
    <!-- Setup Phase -->
    <div v-if="phase === 'setup'" class="flex-1 flex flex-col items-center justify-center max-w-4xl mx-auto w-full space-y-8 animate-in fade-in duration-500">
      <header class="text-center">
        <h1 class="text-3xl font-bold text-zinc-900">AI 模拟面试</h1>
        <p class="text-zinc-500 mt-2">配置您的面试环境与偏好，开启真实对话体验</p>
      </header>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-8 w-full">
        <!-- Preview Area -->
        <div class="aspect-video bg-zinc-900 rounded-2xl relative overflow-hidden flex items-center justify-center group shadow-xl">
          <video ref="previewVideo" class="w-full h-full object-cover" autoplay muted v-if="isCameraOn"></video>
          <div v-else class="flex flex-col items-center text-zinc-500">
            <VideoOff class="h-12 w-12 mb-2" />
            <span>摄像头已关闭</span>
          </div>
          
          <div class="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-3">
            <button 
              @click="toggleMic"
              class="h-10 w-10 rounded-full flex items-center justify-center transition-all hover:scale-110 active:scale-95"
              :class="isMicOn ? 'bg-white/10 text-white backdrop-blur-md hover:bg-white/20' : 'bg-rose-500 text-white'"
            >
              <Mic v-if="isMicOn" class="h-4 w-4" />
              <MicOff v-else class="h-4 w-4" />
            </button>
            <button 
              @click="toggleCamera"
              class="h-10 w-10 rounded-full flex items-center justify-center transition-all hover:scale-110 active:scale-95"
              :class="isCameraOn ? 'bg-white/10 text-white backdrop-blur-md hover:bg-white/20' : 'bg-rose-500 text-white'"
            >
              <Video v-if="isCameraOn" class="h-4 w-4" />
              <VideoOff v-else class="h-4 w-4" />
            </button>
          </div>
        </div>

        <!-- Settings Area -->
        <div class="space-y-5 bg-white p-6 rounded-2xl border border-zinc-100 shadow-sm">
          <div class="space-y-2">
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">目标岗位</label>
            <input 
              v-model="settings.position" 
              class="w-full bg-zinc-50 border border-zinc-200 rounded-xl px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all"
              placeholder="例如: Java 开发工程师"
            />
          </div>
          
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">难度</label>
              <select v-model="settings.difficulty" class="w-full bg-zinc-50 border border-zinc-200 rounded-xl px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all">
                <option value="Junior">初级 (Junior)</option>
                <option value="Mid">中级 (Mid)</option>
                <option value="Senior">高级 (Senior)</option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">模式</label>
              <select v-model="settings.mode" class="w-full bg-zinc-50 border border-zinc-200 rounded-xl px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all">
                <option value="technical">技术深挖</option>
                <option value="hr">HR 行为</option>
                <option value="comprehensive">综合评估</option>
              </select>
            </div>
          </div>
          
          <div class="space-y-2">
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">面试官风格</label>
            <div class="grid grid-cols-3 gap-2">
              <button 
                v-for="style in ['gentle', 'stress', 'deep']" 
                :key="style"
                @click="settings.style = style"
                class="px-3 py-2 rounded-lg text-sm font-medium border transition-all"
                :class="settings.style === style ? 'bg-indigo-50 border-indigo-200 text-indigo-600' : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
              >
                {{ style === 'gentle' ? '温和' : style === 'stress' ? '压力' : '深度' }}
              </button>
            </div>
          </div>

          <button 
            @click="startInterview"
            :disabled="isProcessing"
            class="w-full mt-4 py-4 bg-indigo-600 text-white rounded-xl font-bold text-lg hover:bg-indigo-700 transition-all flex items-center justify-center gap-2 shadow-lg shadow-indigo-200 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Loader2 v-if="isProcessing" class="h-5 w-5 animate-spin" />
            <span v-else>开始面试</span>
            <ChevronRight v-if="!isProcessing" class="h-5 w-5" />
          </button>
        </div>
      </div>
    </div>

    <!-- Interview Phase (Chat Interface) -->
    <div v-else-if="phase === 'interview'" class="h-full flex gap-6 overflow-hidden">
      
      <!-- Left Sidebar: Visuals -->
      <div class="w-80 flex flex-col gap-4 shrink-0">
        <!-- AI Avatar -->
        <div class="flex-1 bg-zinc-900 rounded-3xl flex flex-col items-center justify-center relative overflow-hidden shadow-lg">
          <div class="h-24 w-24 rounded-full bg-indigo-500/20 border border-indigo-500/30 flex items-center justify-center animate-pulse">
            <BrainCircuit class="h-12 w-12 text-indigo-400" />
          </div>
          <p class="mt-6 text-zinc-400 font-medium text-sm">AI 面试官在线</p>
          <div v-if="isProcessing" class="absolute bottom-8 flex gap-1">
            <div class="w-2 h-2 bg-indigo-500 rounded-full animate-bounce" style="animation-delay: 0s"></div>
            <div class="w-2 h-2 bg-indigo-500 rounded-full animate-bounce" style="animation-delay: 0.2s"></div>
            <div class="w-2 h-2 bg-indigo-500 rounded-full animate-bounce" style="animation-delay: 0.4s"></div>
          </div>
        </div>

        <!-- User Video (Small) -->
        <div class="h-48 bg-zinc-800 rounded-3xl overflow-hidden relative shadow-lg">
          <video ref="interviewVideo" class="w-full h-full object-cover" autoplay muted v-if="isCameraOn"></video>
          <div v-else class="w-full h-full flex items-center justify-center text-zinc-500 bg-zinc-900">
            <User class="h-8 w-8" />
          </div>
          <div class="absolute top-3 left-3 px-2 py-1 bg-black/50 rounded-lg text-[10px] text-white uppercase font-bold backdrop-blur-sm">YOU</div>
        </div>
        
        <button 
          @click="endInterviewEarly"
          class="py-3 bg-rose-50 text-rose-600 rounded-2xl font-medium hover:bg-rose-100 transition-colors flex items-center justify-center gap-2"
        >
          <LogOut class="h-4 w-4" />
          结束面试
        </button>
      </div>

      <!-- Right Area: Chat -->
      <div class="flex-1 bg-white rounded-3xl border border-zinc-100 shadow-sm flex flex-col overflow-hidden">
        <!-- Chat Header -->
        <div class="px-6 py-4 border-b border-zinc-100 flex justify-between items-center bg-zinc-50/50">
          <div>
            <h2 class="font-bold text-zinc-900">面试进行中</h2>
            <p class="text-xs text-zinc-500">进度: {{ currentQuestionIndex + 1 }} / {{ questions.length }}</p>
          </div>
          <div class="flex items-center gap-2">
            <div class="h-2 w-2 rounded-full bg-emerald-500 animate-pulse"></div>
            <span class="text-xs font-medium text-emerald-600">实时连接</span>
          </div>
        </div>

        <!-- Messages Area -->
        <div id="chat-container" class="flex-1 overflow-y-auto p-6 space-y-6 bg-zinc-50/30">
          <div 
            v-for="(msg, index) in messages" 
            :key="index"
            class="flex gap-4 max-w-3xl"
            :class="msg.role === 'user' ? 'ml-auto flex-row-reverse' : ''"
          >
            <!-- Avatar -->
            <div 
              class="h-10 w-10 rounded-full flex shrink-0 items-center justify-center shadow-sm"
              :class="msg.role === 'ai' ? 'bg-indigo-100 text-indigo-600' : 'bg-zinc-200 text-zinc-600'"
            >
              <BrainCircuit v-if="msg.role === 'ai'" class="h-5 w-5" />
              <User v-else class="h-5 w-5" />
            </div>

            <!-- Bubble -->
            <div class="flex flex-col gap-1 max-w-[80%]">
              <div 
                class="p-4 rounded-2xl shadow-sm text-sm leading-relaxed whitespace-pre-wrap"
                :class="[
                  msg.role === 'user' ? 'bg-indigo-600 text-white rounded-tr-none' : 'bg-white border border-zinc-100 text-zinc-800 rounded-tl-none',
                  msg.type === 'feedback' ? 'border-l-4 border-l-amber-400 bg-amber-50/50' : ''
                ]"
              >
                <div v-if="msg.type === 'feedback'" class="font-bold text-amber-600 mb-2 text-xs uppercase tracking-wider flex justify-between">
                  <span>面试官反馈</span>
                  <span v-if="msg.score !== undefined && msg.score !== null" class="px-2 py-0.5 rounded-md bg-indigo-100 text-indigo-700">
                    评分: {{ msg.score }}
                  </span>
                </div>
                <template v-if="msg.type === 'feedback'">
                  <div class="rounded-lg bg-amber-50 border border-amber-200 p-2">
                    <div class="text-amber-700 font-semibold text-xs mb-1">评价</div>
                    <div class="text-zinc-800">{{ msg.feedbackEvaluation || msg.content }}</div>
                  </div>
                  <div v-if="msg.feedbackSuggestions && msg.feedbackSuggestions.length > 0" class="mt-3 border-t border-zinc-200 pt-2">
                    <div class="text-emerald-700 font-semibold text-xs mb-1">改进建议</div>
                    <div
                      v-for="(suggestion, sIndex) in msg.feedbackSuggestions"
                      :key="`${index}-${sIndex}`"
                      class="text-emerald-700"
                    >
                      {{ sIndex + 1 }}. {{ suggestion }}
                    </div>
                  </div>
                </template>
                <template v-else>
                  {{ msg.content }}
                </template>
              </div>
              <span class="text-[10px] text-zinc-400 px-1">
                {{ msg.role === 'ai' ? 'AI Interviewer' : 'You' }}
              </span>
            </div>
          </div>
          
          <div v-if="isProcessing" class="flex gap-4 max-w-3xl">
            <div class="h-10 w-10 rounded-full bg-indigo-100 text-indigo-600 flex shrink-0 items-center justify-center">
              <BrainCircuit class="h-5 w-5" />
            </div>
            <div class="bg-white border border-zinc-100 p-4 rounded-2xl rounded-tl-none flex items-center gap-2">
              <span class="text-xs text-zinc-500">正在评估您的回答...</span>
              <Loader2 class="h-3 w-3 animate-spin text-indigo-500" />
            </div>
          </div>
        </div>

        <!-- Input Area -->
        <div class="p-4 bg-white border-t border-zinc-100">
          <div class="relative flex items-end gap-2 bg-zinc-50 p-2 rounded-2xl border border-zinc-200 focus-within:ring-2 focus-within:ring-indigo-500/20 focus-within:border-indigo-500 transition-all">
            <textarea 
              v-model="userInput"
              @keydown.ctrl.enter="sendMessage"
              placeholder="输入您的回答... (Ctrl + Enter 发送)"
              class="w-full bg-transparent border-none focus:ring-0 resize-none max-h-32 min-h-[50px] py-3 px-2 text-sm text-zinc-900 placeholder:text-zinc-400"
              rows="1"
            ></textarea>
            
            <div class="flex gap-2 pb-2 pr-2">
              <button 
                class="p-2 rounded-xl text-zinc-400 hover:text-zinc-600 hover:bg-zinc-200/50 transition-colors"
                title="语音输入 (暂不可用)"
              >
                <Mic class="h-5 w-5" />
              </button>
              <button 
                @click="sendMessage"
                :disabled="!userInput.trim() || isProcessing"
                class="p-2 rounded-xl bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-sm"
              >
                <Send class="h-5 w-5" />
              </button>
            </div>
          </div>
          <div v-if="messages.length > 0 && messages[messages.length-1].type === 'system'" class="mt-2 text-center">
             <button @click="viewReport" class="text-indigo-600 text-sm font-bold hover:underline">查看面试报告</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Custom scrollbar for chat */
#chat-container::-webkit-scrollbar {
  width: 6px;
}
#chat-container::-webkit-scrollbar-track {
  background: transparent;
}
#chat-container::-webkit-scrollbar-thumb {
  background-color: #e4e4e7;
  border-radius: 20px;
}
</style>
