<script setup>
import { Calendar, Clock, X, Briefcase, MessageSquare } from 'lucide-vue-next'

const props = defineProps({
  showBookingDialog: {
    type: Boolean,
    default: false
  },
  showBookingsPanel: {
    type: Boolean,
    default: false
  },
  selectedInvitee: {
    type: Object,
    default: null
  },
  userInvitations: {
    type: Array,
    default: () => []
  },
  bookingForm: {
    type: Object,
    default: () => ({ scheduledAt: '', notes: '' })
  },
  normalizeCandidateRole: {
    type: Function,
    required: true
  }
})

const emit = defineEmits([
  'update:showBookingDialog',
  'update:showBookingsPanel',
  'update:bookingForm',
  'submit-booking',
  'go-live-room'
])

const closeBookingDialog = () => {
  emit('update:showBookingDialog', false)
}

const closeBookingsPanel = () => {
  emit('update:showBookingsPanel', false)
}

const updateBookingField = (field, value) => {
  emit('update:bookingForm', {
    ...(props.bookingForm || {}),
    [field]: value
  })
}

const formatInviteStatus = (status) => {
  if (status === 'pending') return '待对方确认'
  if (status === 'accepted') return '已接受，可开始'
  if (status === 'in_progress') return '进行中'
  if (status === 'completed') return '已完成'
  if (status === 'rejected') return '已拒绝'
  return '已取消'
}

const inviteStatusClass = (status) => {
  if (status === 'pending') return 'bg-amber-100 text-amber-700'
  if (status === 'accepted') return 'bg-sky-100 text-sky-700'
  if (status === 'in_progress') return 'bg-emerald-100 text-emerald-700'
  if (status === 'completed') return 'bg-blue-100 text-blue-700'
  if (status === 'rejected') return 'bg-rose-100 text-rose-700'
  return 'bg-zinc-100 text-zinc-500'
}

const onGoLiveRoom = (booking) => {
  emit('go-live-room', booking)
  closeBookingsPanel()
}
</script>

<template>
  <div v-if="showBookingDialog && selectedInvitee" class="fixed inset-0 z-[60] bg-black/30 backdrop-blur-sm flex items-center justify-center" @click.self="closeBookingDialog">
    <div class="bg-white rounded-2xl shadow-2xl border border-zinc-100 p-6 w-[420px] max-w-[90vw] animate-in fade-in zoom-in-95 duration-300">
      <div class="flex items-center justify-between mb-5">
        <h3 class="font-bold text-lg text-zinc-900">发送真人面试邀请</h3>
        <button @click="closeBookingDialog" class="p-2 hover:bg-zinc-100 rounded-lg transition-colors">
          <X class="w-4 h-4 text-zinc-400" />
        </button>
      </div>

      <div class="flex items-center gap-3 p-3 bg-zinc-50 rounded-xl mb-4">
        <div class="h-12 w-12 rounded-full bg-gradient-to-br from-indigo-100 to-purple-100 flex items-center justify-center text-indigo-700 font-bold shrink-0">
          {{ selectedInvitee.username?.[0] || '?' }}
        </div>
        <div>
          <p class="font-bold text-zinc-800">{{ selectedInvitee.username }}</p>
          <p class="text-xs text-zinc-500">{{ selectedInvitee.email }} · {{ normalizeCandidateRole(selectedInvitee.role) }}</p>
        </div>
      </div>

      <div class="space-y-3">
        <div>
          <label class="text-xs font-bold text-zinc-500 mb-1 block">计划开始时间（可选）</label>
          <input
            type="datetime-local"
            :value="bookingForm?.scheduledAt || ''"
            @input="updateBookingField('scheduledAt', $event.target.value)"
            class="w-full bg-zinc-50 border border-zinc-200 rounded-xl px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>
        <div>
          <label class="text-xs font-bold text-zinc-500 mb-1 block">备注（可选）</label>
          <textarea
            :value="bookingForm?.notes || ''"
            @input="updateBookingField('notes', $event.target.value)"
            placeholder="如：希望重点考察微服务架构设计能力，并增加2轮追问"
            class="w-full bg-zinc-50 border border-zinc-200 rounded-xl px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none h-20"
          ></textarea>
        </div>
      </div>

      <button
        @click="emit('submit-booking')"
        class="w-full mt-4 py-3 bg-indigo-600 text-white rounded-xl font-bold text-sm hover:bg-indigo-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
      >
        发送邀请
      </button>
    </div>
  </div>

  <div v-if="showBookingsPanel" class="fixed inset-0 z-[60] bg-black/30 backdrop-blur-sm flex items-center justify-center" @click.self="closeBookingsPanel">
    <div class="bg-white rounded-2xl shadow-2xl border border-zinc-100 p-6 w-[480px] max-w-[90vw] max-h-[70vh] flex flex-col animate-in fade-in zoom-in-95 duration-300">
      <div class="flex items-center justify-between mb-4">
        <h3 class="font-bold text-lg text-zinc-900 flex items-center gap-2">
          <Calendar class="w-5 h-5 text-indigo-600" />
          我的真人面试邀请
        </h3>
        <button @click="closeBookingsPanel" class="p-2 hover:bg-zinc-100 rounded-lg transition-colors">
          <X class="w-4 h-4 text-zinc-400" />
        </button>
      </div>

      <div class="flex-1 overflow-y-auto space-y-3 custom-scrollbar">
        <div v-if="userInvitations.length === 0" class="text-center py-8">
          <Calendar class="w-10 h-10 text-zinc-200 mx-auto mb-3" />
          <p class="text-sm text-zinc-400">暂无邀请记录</p>
        </div>
        <div v-for="booking in userInvitations" :key="booking.id" class="p-4 rounded-xl border border-zinc-100 hover:border-zinc-200 transition-all">
          <div class="flex items-center justify-between mb-2">
            <span class="text-sm font-bold text-zinc-800">{{ booking.invitee?.username || `用户#${booking.invitee_user_id}` }}</span>
            <span class="text-[10px] px-2 py-0.5 rounded-full font-bold" :class="inviteStatusClass(booking.status)">
              {{ formatInviteStatus(booking.status) }}
            </span>
          </div>
          <div class="text-xs text-zinc-500 space-y-1">
            <p class="flex items-center gap-1.5" v-if="booking.scheduled_at"><Clock class="w-3 h-3" /> {{ new Date(booking.scheduled_at).toLocaleString('zh-CN') }}</p>
            <p class="flex items-center gap-1.5"><Briefcase class="w-3 h-3" /> {{ booking.position }} · {{ booking.difficulty }} · {{ normalizeCandidateRole(booking.invitee_role) }}</p>
            <p v-if="booking.notes" class="flex items-start gap-1.5"><MessageSquare class="w-3 h-3 mt-0.5 shrink-0" /> {{ booking.notes }}</p>
          </div>
          <button
            v-if="booking.status === 'accepted' || booking.status === 'in_progress'"
            @click="onGoLiveRoom(booking)"
            class="mt-3 w-full py-2 rounded-xl text-xs font-semibold border border-indigo-200 bg-indigo-50 text-indigo-700 hover:bg-indigo-100 transition-all"
          >
            进入真人视频面试房间
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
