<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import {
  ChevronRight,
  ChevronDown,
  BrainCircuit,
  Monitor,
  Shuffle,
  UserCheck,
  Headphones,
  Heart,
  Building2,
  Package,
  Timer,
  Calendar,
  Clock,
  Loader2,
  Flame,
  Search,
  Code,
  Briefcase,
  GraduationCap
} from 'lucide-vue-next'

const props = defineProps({
  settings: {
    type: Object,
    required: true
  },
  isProcessing: {
    type: Boolean,
    default: false
  },
  activeInvitationId: {
    type: [Number, String],
    default: null
  },
  inviteCandidates: {
    type: Array,
    default: () => []
  },
  inviteCandidatesLoading: {
    type: Boolean,
    default: false
  },
  activeInvitation: {
    type: Object,
    default: null
  },
  blindBoxRevealing: {
    type: Boolean,
    default: false
  },
  blindBoxRevealed: {
    type: Boolean,
    default: false
  },
  blindBoxScenario: {
    type: Object,
    default: null
  },
  pressureColors: {
    type: Object,
    default: () => ({})
  },
  pressureLevel: {
    type: String,
    default: 'low'
  },
  pressureLabels: {
    type: Object,
    default: () => ({})
  },
  shadowCoachEnabled: {
    type: Boolean,
    default: true
  },
  normalizeCandidateRole: {
    type: Function,
    required: true
  }
})

const emit = defineEmits([
  'update:settings',
  'update:shadow-coach-enabled',
  'change-presentation-mode',
  'change-interview-mode',
  'load-invite-candidates',
  'select-invite-candidate',
  'open-bookings',
  'draw-blind-box',
  'redraw-blind-box',
  'start-interview'
])

const positionDropdownRef = ref(null)
const showPositionDropdown = ref(false)

const positionOptions = [
  'Java后端工程师',
  '前端工程师',
  '算法工程师',
  'AI工程师'
]

const settingsProxy = computed(() => props.settings || {})

const updateSettings = (patch) => {
  emit('update:settings', {
    ...(props.settings || {}),
    ...patch
  })
}

const togglePositionDropdown = () => {
  showPositionDropdown.value = !showPositionDropdown.value
}

const selectPosition = (position) => {
  updateSettings({ position })
  showPositionDropdown.value = false
}

const closePositionDropdownOnOutsideClick = (event) => {
  if (!positionDropdownRef.value) return
  if (!positionDropdownRef.value.contains(event.target)) {
    showPositionDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', closePositionDropdownOnOutsideClick)
})

onUnmounted(() => {
  document.removeEventListener('click', closePositionDropdownOnOutsideClick)
})
</script>

<template>
  <div class="space-y-5 bg-white p-6 rounded-2xl border border-zinc-100 shadow-sm overflow-y-auto max-h-[480px]">
    <div class="space-y-2">
      <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">目标岗位</label>
      <div ref="positionDropdownRef" class="relative" @click.stop>
        <button
          type="button"
          @click="togglePositionDropdown"
          class="interview-position-select w-full bg-gradient-to-b from-white to-zinc-50 border border-zinc-200 rounded-xl px-4 py-3 text-sm font-medium text-zinc-700 transition-all flex items-center justify-between hover:border-indigo-300"
          :class="showPositionDropdown ? 'ring-2 ring-indigo-500 border-indigo-300' : ''"
        >
          <span>{{ settingsProxy.position }}</span>
          <ChevronDown class="w-4 h-4 text-zinc-400 transition-transform" :class="showPositionDropdown ? 'rotate-180' : ''" />
        </button>

        <transition name="dropdown-fade">
          <div
            v-if="showPositionDropdown"
            class="absolute z-30 mt-2 w-full rounded-2xl border border-zinc-200 bg-white shadow-[0_14px_30px_rgba(15,23,42,0.12)] p-2 max-h-64 overflow-y-auto custom-scrollbar"
          >
            <button
              v-for="position in positionOptions"
              :key="position"
              type="button"
              @click="selectPosition(position)"
              class="w-full text-left px-3 py-2.5 rounded-xl text-sm transition-all"
              :class="settingsProxy.position === position ? 'bg-indigo-50 text-indigo-700 font-semibold' : 'text-zinc-700 hover:bg-zinc-50'"
            >
              {{ position }}
            </button>
          </div>
        </transition>
      </div>
    </div>

    <div class="space-y-2">
      <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">对话呈现模式</label>
      <div class="grid grid-cols-2 gap-2">
        <button
          @click="emit('change-presentation-mode', 'video_avatar')"
          class="flex flex-col items-center gap-1 px-3 py-3 rounded-xl text-xs font-medium border transition-all text-center"
          :class="settingsProxy.presentationMode === 'video_avatar' ? 'bg-sky-50 border-sky-300 text-sky-700 ring-1 ring-sky-200' : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
        >
          <span class="text-lg">🎥</span>
          <span class="font-bold">视频模式</span>
          <span class="text-[10px] text-zinc-400">3D 面试官 + 语音播报</span>
        </button>
        <button
          @click="emit('change-presentation-mode', 'text_voice')"
          class="flex flex-col items-center gap-1 px-3 py-3 rounded-xl text-xs font-medium border transition-all text-center"
          :class="settingsProxy.presentationMode === 'text_voice' ? 'bg-emerald-50 border-emerald-300 text-emerald-700 ring-1 ring-emerald-200' : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
        >
          <span class="text-lg">🎙️</span>
          <span class="font-bold">文字语音模式</span>
          <span class="text-[10px] text-zinc-400">低成本高效率训练</span>
        </button>
      </div>
    </div>

    <div class="space-y-2">
      <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">面试类型</label>
      <div class="grid grid-cols-3 gap-2">
        <button
          v-for="m in [
            { key: 'technical', label: '技术面', icon: Monitor, desc: '编程/算法/系统设计' },
            { key: 'hr', label: 'HR面', icon: UserCheck, desc: '沟通/职业规划/软技能' },
            { key: 'comprehensive', label: '综合面', icon: BrainCircuit, desc: '技术+HR联合面试' }
          ]"
          :key="m.key"
          @click="updateSettings({ mode: m.key })"
          class="flex flex-col items-center gap-1 px-3 py-3 rounded-xl text-sm font-medium border transition-all text-center relative group"
          :class="settingsProxy.mode === m.key
            ? 'bg-indigo-50 border-indigo-200 text-indigo-600 ring-1 ring-indigo-200'
            : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
        >
          <component :is="m.icon" class="h-5 w-5 shrink-0" />
          <span class="font-bold text-xs">{{ m.label }}</span>
          <span class="text-[10px] text-zinc-400 leading-tight">{{ m.desc }}</span>
        </button>
      </div>
    </div>

    <div class="space-y-2">
      <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">面试官风格</label>
      <div class="grid grid-cols-3 gap-2">
        <button
          v-for="s in [
            { key: 'gentle', label: '温和型', icon: Heart },
            { key: 'stress', label: '压力型', icon: Flame },
            { key: 'deep', label: '技术深挖', icon: Search },
            { key: 'practical', label: '项目实战', icon: Briefcase },
            { key: 'algorithm', label: '算法考察', icon: Code }
          ]"
          :key="s.key"
          @click="updateSettings({ style: s.key })"
          class="flex items-center gap-2 px-3 py-2.5 rounded-xl text-xs font-medium border transition-all"
          :class="settingsProxy.style === s.key ? 'bg-indigo-50 border-indigo-200 text-indigo-600 ring-1 ring-indigo-200' : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
        >
          <component :is="s.icon" class="h-3.5 w-3.5 shrink-0" />
          {{ s.label }}
        </button>
      </div>
    </div>

    <div class="space-y-2">
      <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider flex items-center gap-2">
        <Building2 class="w-3.5 h-3.5" /> 大厂面试风格（可选）
      </label>
      <div class="grid grid-cols-3 gap-2">
        <button
          v-for="c in [
            { key: '', label: '不限', emoji: '🌐' },
            { key: 'ali', label: '阿里', emoji: '🟠' },
            { key: 'bytedance', label: '字节', emoji: '⚡' },
            { key: 'tencent', label: '腾讯', emoji: '🐧' },
            { key: 'meituan', label: '美团', emoji: '🟡' },
            { key: 'baidu', label: '百度', emoji: '🔵' }
          ]"
          :key="c.key"
          @click="updateSettings({ company: c.key })"
          class="flex items-center gap-1.5 px-3 py-2 rounded-xl text-xs font-medium border transition-all"
          :class="settingsProxy.company === c.key ? 'bg-orange-50 border-orange-200 text-orange-700 ring-1 ring-orange-200' : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
        >
          <span>{{ c.emoji }}</span>
          {{ c.label }}
        </button>
      </div>
    </div>

    <div class="space-y-2">
      <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider flex items-center gap-2">
        <GraduationCap class="w-3.5 h-3.5" /> 难度等级
      </label>
      <div class="grid grid-cols-3 gap-2">
        <button
          v-for="d in [
            { key: 'campus_intern', label: '校招实习', desc: '在校实习生' },
            { key: 'campus_graduate', label: '校招应届', desc: '应届毕业生' },
            { key: 'social_junior', label: '社招初级', desc: '1-3年经验' }
          ]"
          :key="d.key"
          @click="updateSettings({ difficulty: d.key })"
          class="flex flex-col items-center gap-0.5 px-3 py-2.5 rounded-xl text-xs font-medium border transition-all"
          :class="settingsProxy.difficulty === d.key ? 'bg-indigo-50 border-indigo-200 text-indigo-600 ring-1 ring-indigo-200' : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
        >
          <span class="font-bold">{{ d.label }}</span>
          <span class="text-[10px] text-zinc-400">{{ d.desc }}</span>
        </button>
      </div>
    </div>

    <div class="space-y-2">
      <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">面试模式</label>
      <div class="grid grid-cols-3 gap-2">
        <button
          v-for="im in [
            { key: 'ai', label: 'AI仿真面试官', icon: '🤖', desc: 'AI模拟真实面试' },
            { key: 'human', label: '真人面试', icon: '👤', desc: '邀请高校端/企业端账号' },
            { key: 'random', label: '随机模式', icon: '🎲', desc: '风格随机不提前告知' }
          ]"
          :key="im.key"
          @click="emit('change-interview-mode', im.key)"
          class="flex flex-col items-center gap-1 px-3 py-3 rounded-xl text-xs font-medium border transition-all text-center"
          :class="settingsProxy.interviewMode === im.key
            ? (im.key === 'random' ? 'bg-violet-50 border-violet-300 text-violet-700 ring-1 ring-violet-200' : im.key === 'human' ? 'bg-emerald-50 border-emerald-200 text-emerald-700 ring-1 ring-emerald-200' : 'bg-indigo-50 border-indigo-200 text-indigo-600 ring-1 ring-indigo-200')
            : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
        >
          <span class="text-xl">{{ im.icon }}</span>
          <span class="font-bold">{{ im.label }}</span>
          <span class="text-[10px] text-zinc-400 leading-tight">{{ im.desc }}</span>
        </button>
      </div>
    </div>

    <div v-if="settingsProxy.interviewMode === 'random'" class="p-3 bg-violet-50 rounded-xl border border-violet-200 animate-in fade-in duration-300">
      <div class="flex items-start gap-2">
        <Shuffle class="w-4 h-4 text-violet-600 mt-0.5 shrink-0" />
        <div>
          <p class="text-xs font-bold text-violet-700">随机模式说明</p>
          <p class="text-[11px] text-violet-600 leading-relaxed mt-1">
            系统将随机分配面试官风格（温和/压力/深挖等），可能随机匹配大厂面试风格。
            面试过程中不会提前告知风格类型，模拟真实企业的"突然压力追问"场景。
            面试结束后将揭晓本次的面试官风格。
          </p>
        </div>
      </div>
    </div>

    <div v-if="settingsProxy.interviewMode === 'human'" class="space-y-3 animate-in fade-in slide-in-from-top-2 duration-300">
      <div class="flex gap-2">
        <button
          v-for="t in [{key: '', label: '全部'}, {key: 'campus', label: '🏫 校内老师'}, {key: 'enterprise', label: '🏢 企业专家'}]"
          :key="t.key"
          @click="emit('load-invite-candidates', t.key)"
          class="px-3 py-1.5 rounded-lg text-xs font-medium border transition-all"
          :class="'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
        >
          {{ t.label }}
        </button>
      </div>

      <div v-if="inviteCandidatesLoading" class="text-center py-4">
        <Loader2 class="w-6 h-6 text-zinc-400 animate-spin mx-auto" />
        <p class="text-xs text-zinc-400 mt-2">加载可邀请用户...</p>
      </div>
      <div v-else-if="inviteCandidates.length > 0" class="space-y-2 max-h-[180px] overflow-y-auto custom-scrollbar">
        <div
          v-for="candidate in inviteCandidates"
          :key="candidate.id"
          @click="emit('select-invite-candidate', candidate)"
          class="flex items-center gap-3 p-3 rounded-xl border border-zinc-100 hover:border-indigo-200 hover:bg-indigo-50/30 cursor-pointer transition-all group"
        >
          <div class="h-10 w-10 rounded-full bg-gradient-to-br from-indigo-100 to-purple-100 flex items-center justify-center text-indigo-700 font-bold text-sm shrink-0">
            {{ candidate.username?.[0] || '?' }}
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="text-sm font-bold text-zinc-800">{{ candidate.username }}</span>
              <span class="text-[10px] px-1.5 py-0.5 rounded-full"
                :class="candidate.role === 'university' ? 'bg-blue-100 text-blue-700' : 'bg-orange-100 text-orange-700'">
                {{ candidate.role === 'university' ? '高校端' : '企业端' }}
              </span>
            </div>
            <p class="text-[11px] text-zinc-500 truncate">{{ candidate.email }}</p>
            <div class="flex items-center gap-2 mt-0.5">
              <span class="text-[10px] text-zinc-400">{{ normalizeCandidateRole(candidate.role) }}</span>
            </div>
          </div>
          <Calendar class="w-4 h-4 text-zinc-300 group-hover:text-indigo-500 transition-colors shrink-0" />
        </div>
      </div>
      <div v-else class="p-4 bg-zinc-50 rounded-xl text-center">
        <p class="text-xs text-zinc-400">暂无可邀请用户，请先在高校端/企业端注册账号。</p>
      </div>

      <button @click="emit('open-bookings')" class="w-full py-2 rounded-xl text-xs font-medium border border-zinc-200 bg-white hover:bg-zinc-50 text-zinc-600 transition-all flex items-center justify-center gap-1.5">
        <Clock class="w-3 h-3" /> 查看我的邀请
      </button>

      <div v-if="activeInvitation" class="rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-xs text-emerald-700">
        已选择邀请：{{ activeInvitation.invitee?.username || activeInvitation.invitee_user_id }}（{{ normalizeCandidateRole(activeInvitation.invitee_role) }}）
      </div>
    </div>

    <div v-if="settingsProxy.mode === 'blindbox'" class="space-y-3 animate-in fade-in slide-in-from-top-2 duration-300">
      <div v-if="blindBoxRevealing" class="p-6 bg-gradient-to-br from-violet-100 to-purple-50 rounded-2xl border border-violet-200 flex flex-col items-center justify-center gap-3">
        <div class="relative">
          <Package class="h-12 w-12 text-violet-600 animate-bounce" />
          <div class="absolute -top-1 -right-1 w-4 h-4 bg-yellow-400 rounded-full animate-ping"></div>
        </div>
        <p class="text-sm font-bold text-violet-700 animate-pulse">正在抽取面试场景...</p>
      </div>
      <div v-else-if="blindBoxRevealed && blindBoxScenario"
        class="p-4 rounded-2xl border-2 shadow-md animate-in fade-in zoom-in-95 duration-500 relative overflow-hidden"
        :class="[pressureColors[pressureLevel]?.bg, pressureColors[pressureLevel]?.border]"
      >
        <div class="absolute top-0 right-0 w-20 h-20 bg-gradient-to-bl from-white/40 to-transparent rounded-bl-full pointer-events-none"></div>
        <div class="flex items-start gap-3">
          <span class="text-3xl">{{ blindBoxScenario.icon }}</span>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 mb-1">
              <h4 class="font-bold text-base" :class="pressureColors[pressureLevel]?.text">{{ blindBoxScenario.name }}</h4>
              <span class="text-[10px] font-bold px-2 py-0.5 rounded-full" :class="pressureColors[pressureLevel]?.badge">
                {{ pressureLabels[pressureLevel] }}
              </span>
            </div>
            <p class="text-xs text-zinc-600 leading-relaxed mb-2">{{ blindBoxScenario.description }}</p>
            <div class="flex flex-wrap gap-1.5">
              <span v-for="tag in blindBoxScenario.tags" :key="tag" class="text-[10px] px-2 py-0.5 rounded-full bg-white/60 text-zinc-500 border border-zinc-200">{{ tag }}</span>
              <span v-if="blindBoxScenario.time_limit" class="text-[10px] px-2 py-0.5 rounded-full bg-white/60 text-zinc-500 border border-zinc-200 flex items-center gap-1">
                <Timer class="w-2.5 h-2.5" /> {{ blindBoxScenario.time_limit }}s/题
              </span>
            </div>
          </div>
        </div>
        <button @click="emit('redraw-blind-box')" class="mt-3 w-full py-2 rounded-xl text-xs font-medium border border-zinc-200 bg-white/80 hover:bg-white text-zinc-600 transition-all flex items-center justify-center gap-1.5">
          <Shuffle class="w-3 h-3" /> 换一个场景
        </button>
      </div>
      <div v-else class="p-4 bg-zinc-50 rounded-2xl border border-dashed border-zinc-300 text-center">
        <button @click="emit('draw-blind-box')" class="px-4 py-2 rounded-xl bg-violet-600 text-white text-sm font-medium hover:bg-violet-700 transition-all flex items-center gap-2 mx-auto">
          <Package class="w-4 h-4" /> 抽取盲盒场景
        </button>
      </div>
    </div>

    <div class="flex items-center justify-between p-3 bg-zinc-50 rounded-xl">
      <div class="flex items-center gap-2">
        <Headphones class="h-4 w-4 text-indigo-600" />
        <span class="text-sm font-medium text-zinc-700">AI 影子教练 (实时耳返)</span>
      </div>
      <button
        @click="emit('update:shadow-coach-enabled', !shadowCoachEnabled)"
        class="w-10 h-5 rounded-full transition-colors relative"
        :class="shadowCoachEnabled ? 'bg-indigo-600' : 'bg-zinc-300'"
      >
        <div class="absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full transition-transform" :class="shadowCoachEnabled ? 'translate-x-5' : ''"></div>
      </button>
    </div>

    <button
      @click="emit('start-interview')"
      :disabled="isProcessing || (settingsProxy.interviewMode === 'human' && !activeInvitationId)"
      class="w-full mt-2 py-4 bg-indigo-600 text-white rounded-xl font-bold text-lg hover:bg-indigo-700 transition-all flex items-center justify-center gap-2 shadow-lg shadow-indigo-200 disabled:opacity-50 disabled:cursor-not-allowed"
    >
      <Loader2 v-if="isProcessing" class="h-5 w-5 animate-spin" />
      <span v-else-if="settingsProxy.interviewMode === 'human' && !activeInvitationId">请先选择一个邀请</span>
      <span v-else>开始面试</span>
      <ChevronRight v-if="!isProcessing && (settingsProxy.interviewMode !== 'human' || activeInvitationId)" class="h-5 w-5" />
    </button>
  </div>
</template>
