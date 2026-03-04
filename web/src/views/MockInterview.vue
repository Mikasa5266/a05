<script setup>
import { ref, onMounted, onUnmounted, nextTick, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { 
  Video, VideoOff, Mic, MicOff, ChevronRight, 
  BrainCircuit, User, LogOut, Send, Loader2,
  History, MessageSquare, Lightbulb
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
const showHistory = ref(false)

const latestAIMessage = computed(() => {
  const aiMsgs = messages.value.filter(m => m.role === 'ai' || m.type === 'system')
  return aiMsgs.length > 0 ? aiMsgs[aiMsgs.length - 1] : null
})

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

    <!-- Interview Phase (New Layout) -->
    <div v-else-if="phase === 'interview'" class="h-full flex flex-col lg:flex-row gap-6 p-6 bg-zinc-50 overflow-hidden">
      
      <!-- Left Main Column (Video + Input) -->
      <div class="flex-1 flex flex-col gap-6 min-w-0 h-full">
        <!-- Video Section (Top) -->
        <div class="flex-1 bg-black rounded-3xl relative overflow-hidden shadow-2xl group ring-1 ring-zinc-900/5">
          <!-- Status Badge -->
          <div class="absolute top-6 left-6 flex items-center gap-3 z-10 pointer-events-none">
            <div class="bg-rose-600 text-white text-xs font-bold px-3 py-1.5 rounded-full flex items-center gap-2 shadow-lg shadow-rose-900/20">
              <div class="w-2 h-2 bg-white rounded-full animate-pulse"></div>
              REC
            </div>
            <div class="bg-black/40 text-white/90 text-xs px-3 py-1.5 rounded-full backdrop-blur-md border border-white/10 shadow-sm">
              多模态情绪监测中...
            </div>
          </div>
          
          <!-- Video Element -->
          <video ref="interviewVideo" class="w-full h-full object-cover transform scale-x-[-1]" autoplay muted v-if="isCameraOn"></video>
          <div v-else class="w-full h-full flex flex-col items-center justify-center text-zinc-600 bg-zinc-900/50">
             <User class="h-24 w-24 mb-6 opacity-20" />
             <p class="font-medium tracking-wide opacity-50">摄像头已关闭</p>
          </div>

          <!-- Controls (Bottom Center - Auto hide) -->
          <div class="absolute bottom-8 left-1/2 -translate-x-1/2 flex gap-4 transition-all duration-500 translate-y-4 opacity-0 group-hover:translate-y-0 group-hover:opacity-100">
            <button 
              @click="toggleMic"
              class="h-12 w-12 rounded-full flex items-center justify-center backdrop-blur-md transition-all hover:scale-110 shadow-lg border border-white/10"
              :class="isMicOn ? 'bg-white/20 text-white hover:bg-white/30' : 'bg-rose-500 text-white'"
            >
              <Mic v-if="isMicOn" class="h-5 w-5" />
              <MicOff v-else class="h-5 w-5" />
            </button>
            <button 
              @click="toggleCamera"
              class="h-12 w-12 rounded-full flex items-center justify-center backdrop-blur-md transition-all hover:scale-110 shadow-lg border border-white/10"
              :class="isCameraOn ? 'bg-white/20 text-white hover:bg-white/30' : 'bg-rose-500 text-white'"
            >
              <Video v-if="isCameraOn" class="h-5 w-5" />
              <VideoOff v-else class="h-5 w-5" />
            </button>
          </div>
        </div>

        <!-- Transcript / Input Section (Bottom) -->
        <div class="h-1/3 min-h-[200px] bg-white rounded-3xl p-6 shadow-xl shadow-zinc-200/50 border border-white flex flex-col relative transition-all duration-300 focus-within:ring-2 focus-within:ring-indigo-500/20 focus-within:shadow-indigo-500/10 group">
           <div class="flex justify-between items-center mb-4">
             <h3 class="font-bold text-zinc-900 flex items-center gap-2 group-focus-within:text-indigo-600 transition-colors">
               <div class="w-1.5 h-4 bg-zinc-300 rounded-full group-focus-within:bg-indigo-600 transition-colors"></div>
               实时回答转写
               <span v-if="userInput.length > 0" class="text-xs font-normal text-emerald-600 flex items-center gap-1 bg-emerald-50 px-2 py-0.5 rounded-full animate-in fade-in zoom-in duration-300">
                 <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                 正在输入...
               </span>
             </h3>
             <button 
                @click.stop="showHistory = true" 
                class="text-xs text-zinc-400 hover:text-indigo-600 transition-colors flex items-center gap-1 px-2 py-1 hover:bg-zinc-50 rounded-lg"
             >
               <History class="w-3 h-3" /> 历史记录
             </button>
           </div>
           
           <textarea 
              v-model="userInput" 
              @keydown.ctrl.enter="sendMessage"
              placeholder="在此处输入您的回答..."
              class="flex-1 w-full resize-none border-none focus:ring-0 p-4 text-lg text-zinc-700 placeholder:text-zinc-300 bg-zinc-50/50 rounded-xl leading-relaxed transition-all focus:bg-white focus:shadow-inner custom-scrollbar"
           ></textarea>

           <div class="absolute bottom-8 right-8 text-[10px] font-medium text-zinc-300 pointer-events-none bg-white/80 backdrop-blur px-2 py-1 rounded-md border border-zinc-100">
             Ctrl + Enter 发送
           </div>
        </div>
      </div>

      <!-- Right Sidebar -->
      <div class="w-full lg:w-[400px] flex flex-col gap-6 shrink-0 h-full">
        <!-- AI Profile Card -->
        <div class="bg-white p-5 rounded-3xl border border-white shadow-lg shadow-zinc-200/50 flex items-center gap-4 hover:shadow-xl transition-shadow duration-300">
          <div class="h-14 w-14 rounded-2xl bg-gradient-to-br from-indigo-600 to-violet-600 flex items-center justify-center text-white shadow-lg shadow-indigo-500/30 ring-4 ring-indigo-50">
            <BrainCircuit class="h-7 w-7" />
          </div>
          <div>
            <h3 class="font-bold text-zinc-900 text-lg">智聘智能引擎</h3>
            <p class="text-xs text-zinc-500 font-medium flex items-center gap-1">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"></span>
              AI 面试官在线
            </p>
          </div>
        </div>

        <!-- Question / Feedback Card -->
        <div class="bg-white rounded-3xl border border-white shadow-xl shadow-zinc-200/50 flex-1 flex flex-col relative overflow-hidden group transition-all duration-300 hover:shadow-2xl hover:shadow-zinc-200/60">
           <!-- Card Header -->
           <div class="px-6 py-5 border-b border-zinc-50 flex justify-between items-center bg-zinc-50/50 backdrop-blur-sm z-10">
             <div class="inline-flex items-center gap-1.5 px-3 py-1 bg-indigo-50 text-indigo-700 text-xs font-bold rounded-full border border-indigo-100/50">
               <span class="w-1.5 h-1.5 rounded-full bg-indigo-600"></span>
               当前提问 ({{ currentQuestionIndex + 1 }}/{{ questions.length }})
             </div>
             <div v-if="latestAIMessage?.type === 'feedback'" class="flex items-center gap-2 animate-in fade-in slide-in-from-right duration-500">
                <span class="text-xs text-zinc-400 font-medium">评分</span>
                <span class="text-xl font-black text-indigo-600 tracking-tight">{{ latestAIMessage.score }}</span>
             </div>
           </div>
           
           <!-- Content Area -->
           <div class="flex-1 overflow-y-auto p-6 custom-scrollbar relative">
             <!-- Loading State -->
             <div v-if="isProcessing && !latestAIMessage" class="absolute inset-0 flex flex-col items-center justify-center text-zinc-400 gap-3 bg-white/80 backdrop-blur-sm z-20">
                <div class="relative">
                  <div class="absolute inset-0 bg-indigo-500/20 blur-xl rounded-full"></div>
                  <Loader2 class="h-10 w-10 animate-spin text-indigo-600 relative z-10" />
                </div>
                <p class="text-sm font-medium animate-pulse">正在生成评估...</p>
             </div>

             <!-- Content -->
             <div v-else class="space-y-4 animate-in fade-in slide-in-from-bottom-4 duration-500">
               <!-- If it's a Question -->
               <template v-if="latestAIMessage?.type === 'question' || (latestAIMessage?.role === 'ai' && !latestAIMessage?.type)">
                 <h2 class="text-xl font-bold text-zinc-900 leading-relaxed tracking-wide">
                   {{ latestAIMessage?.content }}
                 </h2>
               </template>

               <!-- If it's Feedback -->
               <template v-else-if="latestAIMessage?.type === 'feedback'">
                 <div class="space-y-4">
                   <div class="p-5 bg-gradient-to-br from-amber-50 to-orange-50/30 rounded-2xl border border-amber-100/60 shadow-sm">
                      <h4 class="text-xs font-bold text-amber-600 uppercase mb-3 flex items-center gap-2">
                        <div class="p-1 bg-amber-100 rounded-md">
                          <MessageSquare class="w-3.5 h-3.5" />
                        </div>
                        评价
                      </h4>
                      <p class="text-sm text-zinc-800 leading-relaxed text-justify">{{ latestAIMessage.feedbackEvaluation }}</p>
                    </div>
                    
                    <div class="p-5 bg-gradient-to-br from-emerald-50 to-teal-50/30 rounded-2xl border border-emerald-100/60 shadow-sm">
                      <h4 class="text-xs font-bold text-emerald-600 uppercase mb-3 flex items-center gap-2">
                        <div class="p-1 bg-emerald-100 rounded-md">
                          <Lightbulb class="w-3.5 h-3.5" />
                        </div>
                        改进建议
                      </h4>
                     <ul class="space-y-3">
                       <li v-for="(s, i) in latestAIMessage.feedbackSuggestions" :key="i" class="text-sm text-emerald-900 flex gap-3 leading-relaxed group/item">
                         <span class="font-bold text-emerald-600/40 font-mono text-xs mt-0.5 group-hover/item:text-emerald-600 transition-colors">0{{ i + 1 }}</span>
                         {{ s }}
                       </li>
                     </ul>
                   </div>
                 </div>
               </template>
               
               <!-- System Message -->
               <template v-else-if="latestAIMessage?.type === 'system'">
                  <div class="p-6 bg-zinc-50 rounded-2xl text-center text-zinc-600 text-sm border border-zinc-100">
                    <p class="mb-4">{{ latestAIMessage.content }}</p>
                    <div v-if="messages[messages.length-1].content.includes('面试结束')" class="flex justify-center">
                       <button @click="viewReport" class="px-8 py-3 bg-indigo-600 text-white rounded-xl text-sm font-bold hover:bg-indigo-700 transition-all shadow-lg shadow-indigo-200 hover:shadow-indigo-300 hover:-translate-y-0.5 active:translate-y-0">
                         查看详细报告
                       </button>
                    </div>
                  </div>
               </template>
             </div>
           </div>
        </div>

        <!-- Hint Card -->
        <div v-if="latestAIMessage?.type !== 'feedback'" class="bg-gradient-to-br from-indigo-50/80 to-white p-6 rounded-3xl border border-white shadow-lg shadow-zinc-200/30 backdrop-blur-sm">
          <h4 class="text-xs font-bold text-indigo-400 uppercase mb-3 flex items-center gap-2">
            <span class="w-1.5 h-1.5 bg-indigo-400 rounded-full animate-pulse"></span>
            面试提示
          </h4>
          <p class="text-sm text-zinc-600 italic leading-relaxed opacity-80">
            "建议从 STAR 原则出发，重点描述你在项目中遇到的挑战以及你是如何克服它的。"
          </p>
        </div>

        <!-- Action Button -->
        <button 
          @click="sendMessage"
          :disabled="isProcessing || (!userInput.trim() && latestAIMessage?.type !== 'feedback')"
          class="w-full py-4 bg-zinc-900 text-white rounded-2xl font-bold text-lg hover:bg-zinc-800 hover:shadow-xl hover:shadow-zinc-900/20 active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-3 group relative overflow-hidden"
        >
          <div class="absolute inset-0 bg-white/10 translate-y-full group-hover:translate-y-0 transition-transform duration-300"></div>
          <span v-if="isProcessing" class="flex items-center gap-2 relative z-10">
            <Loader2 class="w-5 h-5 animate-spin" />
            正在思考...
          </span>
          <span v-else-if="latestAIMessage?.type === 'feedback'" class="flex items-center gap-2 relative z-10">
            下一题 <ChevronRight class="w-5 h-5 group-hover:translate-x-1 transition-transform" />
          </span>
          <span v-else class="flex items-center gap-2 relative z-10">
            发送回答 
            <Send class="h-4 w-4 group-hover:translate-x-1 transition-transform" />
          </span>
        </button>
      </div>

      <!-- History Drawer (Overlay) -->
      <div v-if="showHistory" class="absolute inset-0 z-50 bg-black/20 backdrop-blur-sm flex justify-end" @click.self="showHistory = false">
        <div class="w-96 bg-white h-full shadow-2xl animate-in slide-in-from-right duration-300 flex flex-col border-l border-zinc-100">
          <div class="p-5 border-b border-zinc-100 flex justify-between items-center bg-zinc-50/50">
            <h3 class="font-bold text-zinc-900 flex items-center gap-2">
              <History class="w-4 h-4 text-zinc-400" />
              对话历史
            </h3>
            <button @click="showHistory = false" class="p-2 hover:bg-zinc-200/50 rounded-full transition-colors text-zinc-400 hover:text-zinc-600">
              <ChevronRight class="h-5 w-5" />
            </button>
          </div>
          <div class="flex-1 overflow-y-auto p-4 space-y-4 custom-scrollbar bg-zinc-50/30">
            <div v-for="(msg, i) in messages" :key="i" class="text-sm p-4 rounded-2xl border shadow-sm transition-all hover:shadow-md" 
              :class="msg.role === 'user' ? 'bg-white border-zinc-100 text-zinc-800 ml-4' : 'bg-indigo-50/50 border-indigo-100 text-zinc-800 mr-4'">
              <div class="text-[10px] uppercase tracking-wider font-bold mb-2 flex items-center gap-1" 
                :class="msg.role === 'user' ? 'text-zinc-400 justify-end' : 'text-indigo-400'">
                <User v-if="msg.role === 'user'" class="w-3 h-3" />
                <BrainCircuit v-else class="w-3 h-3" />
                {{ msg.role === 'ai' ? 'AI 面试官' : '你' }}
              </div>
              <div class="leading-relaxed whitespace-pre-wrap">{{ msg.content }}</div>
            </div>
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
