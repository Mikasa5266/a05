<script setup>
import { BrainCircuit, ChevronRight, History, User } from 'lucide-vue-next'

const props = defineProps({
  showHistory: {
    type: Boolean,
    default: false
  },
  messages: {
    type: Array,
    default: () => []
  },
  settings: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['update:showHistory'])

const closeDrawer = () => {
  emit('update:showHistory', false)
}

const buildFeedbackPlainText = (sections) => {
  const lines = []
  const evaluation = (sections?.evaluation || '').trim()
  if (evaluation) {
    lines.push(`评价：${evaluation}`)
  }

  const suggestions = Array.isArray(sections?.suggestions) ? sections.suggestions.filter(Boolean) : []
  if (suggestions.length > 0) {
    lines.push('建议：')
    suggestions.forEach((item) => lines.push(`- ${item}`))
  }

  const followUp = (sections?.followUp || '').trim()
  if (followUp) {
    lines.push(`追问方向：${followUp}`)
  }

  return lines.join('\n').trim() || '回答已提交，建议补充更具体的技术细节。'
}

const getHistoryMessageContent = (msg) => {
  if (!msg) return ''

  if (msg.type === 'feedback') {
    const sections = {
      evaluation: msg.feedbackEvaluation,
      suggestions: msg.feedbackSuggestions,
      followUp: msg.feedbackFollowUp
    }
    const hasStructured = sections.evaluation || (sections.suggestions && sections.suggestions.length) || sections.followUp
    if (hasStructured) {
      return buildFeedbackPlainText(sections)
    }
  }

  return typeof msg.content === 'string' ? msg.content : String(msg.content || '')
}
</script>

<template>
  <div v-if="showHistory" class="fixed inset-0 z-70 bg-black/20 backdrop-blur-sm flex justify-end" @click.self="closeDrawer">
    <div class="w-96 max-w-[92vw] bg-white h-dvh shadow-2xl animate-in slide-in-from-right duration-300 flex flex-col border-l border-zinc-100">
      <div class="p-5 border-b border-zinc-100 flex justify-between items-center bg-zinc-50/50">
        <h3 class="font-bold text-zinc-900 flex items-center gap-2">
          <History class="w-4 h-4 text-zinc-400" />
          对话历史
        </h3>
        <button @click="closeDrawer" class="p-2 hover:bg-zinc-200/50 rounded-full transition-colors text-zinc-400 hover:text-zinc-600">
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
            {{ msg.role === 'ai' ? (settings.interviewMode === 'human' ? '真人面试流程' : 'AI 面试官') : '你' }}
          </div>
          <div class="leading-relaxed whitespace-pre-wrap">{{ getHistoryMessageContent(msg) }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
