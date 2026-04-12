<script setup>
const props = defineProps({
  item: {
    type: Object,
    required: true
  },
  actionLoadingId: {
    type: [Number, String, null],
    default: null
  }
})

const emit = defineEmits(['accept', 'reject', 'enter-room'])
</script>

<template>
  <div class="mt-4 flex flex-wrap gap-2">
    <button
      v-if="props.item.status === 'pending'"
      class="px-3 py-1.5 rounded-lg bg-emerald-50 text-emerald-700 text-xs font-semibold hover:bg-emerald-100 disabled:opacity-60"
      :disabled="props.actionLoadingId === props.item.id"
      @click="emit('accept', props.item.id)"
    >
      接受邀请
    </button>
    <button
      v-if="props.item.status === 'pending'"
      class="px-3 py-1.5 rounded-lg bg-rose-50 text-rose-700 text-xs font-semibold hover:bg-rose-100 disabled:opacity-60"
      :disabled="props.actionLoadingId === props.item.id"
      @click="emit('reject', props.item.id)"
    >
      拒绝邀请
    </button>

    <button
      v-if="props.item.can_join"
      class="ml-auto px-3.5 py-1.5 rounded-lg border border-indigo-200 bg-indigo-50 text-indigo-700 text-xs font-semibold hover:bg-indigo-100"
      @click="emit('enter-room', props.item)"
    >
      进入面试房间
    </button>
  </div>
</template>
