<script setup>
const props = defineProps({
  item: {
    type: Object,
    required: true
  },
  statusLabel: {
    type: Function,
    required: true
  },
  statusClass: {
    type: Function,
    required: true
  },
  difficultyLabel: {
    type: Function,
    required: true
  },
  modeLabel: {
    type: Function,
    required: true
  },
  interviewStatusLabel: {
    type: Function,
    required: true
  },
  formatDateTime: {
    type: Function,
    required: true
  },
  formatCountdown: {
    type: Function,
    required: true
  }
})
</script>

<template>
  <div>
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <div class="flex items-center flex-wrap gap-2">
          <h3 class="text-base font-semibold text-zinc-900">
            {{ props.item.student_name || `学生#${props.item.student_id}` }}
          </h3>
          <span class="px-2 py-1 rounded-full text-xs font-medium" :class="props.statusClass(props.item.status)">
            {{ props.statusLabel(props.item.status) }}
          </span>
        </div>
        <p class="text-sm text-zinc-600 mt-1">
          {{ props.item.position || '未设置岗位' }} · {{ props.difficultyLabel(props.item.difficulty) }} · {{ props.modeLabel(props.item.mode) }}
        </p>
        <p class="text-xs text-zinc-500 mt-1">
          邀请码：{{ props.item.invitation_code || '-' }}
        </p>
        <p class="text-xs text-zinc-400 mt-1">
          创建时间：{{ props.formatDateTime(props.item.created_at) }}
        </p>
      </div>

      <div class="text-right min-w-[180px]">
        <p v-if="props.item.status === 'in_progress'" class="text-xs text-zinc-500">剩余倒计时</p>
        <p v-if="props.item.status === 'in_progress'" class="text-xl font-semibold text-indigo-700 mt-1">
          {{ props.formatCountdown(props.item) }}
        </p>
        <p v-else class="text-xs text-zinc-500">
          更新时间：{{ props.formatDateTime(props.item.updated_at) }}
        </p>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mt-4">
      <div class="rounded-xl bg-zinc-50 px-3 py-2">
        <p class="text-xs text-zinc-400">拟定时间</p>
        <p class="text-sm text-zinc-700 mt-1">{{ props.formatDateTime(props.item.scheduled_at) }}</p>
      </div>
      <div class="rounded-xl bg-zinc-50 px-3 py-2">
        <p class="text-xs text-zinc-400">面试状态</p>
        <p class="text-sm text-zinc-700 mt-1">{{ props.interviewStatusLabel(props.item.interview_status) }}</p>
      </div>
    </div>

    <div v-if="props.item.status === 'completed'" class="mt-4 rounded-xl border border-emerald-100 bg-emerald-50/50 p-3">
      <p class="text-xs text-emerald-700">评价记录</p>
      <p class="text-sm text-zinc-700 mt-1">评分：{{ props.item.human_score ?? '--' }}</p>
      <p class="text-sm text-zinc-600 mt-1">{{ props.item.human_feedback || '暂无评价内容' }}</p>
    </div>

    <p v-if="props.item.notes" class="text-xs text-zinc-500 mt-3">备注：{{ props.item.notes }}</p>
  </div>
</template>
