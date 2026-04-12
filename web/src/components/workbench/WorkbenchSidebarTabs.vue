<script setup>
const props = defineProps({
  tabs: {
    type: Array,
    default: () => []
  },
  activeTab: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['select'])
</script>

<template>
  <section class="bg-white border border-zinc-100 rounded-2xl p-2 shadow-sm">
    <div class="flex flex-wrap gap-2">
      <button
        v-for="tab in props.tabs"
        :key="tab.key"
        class="px-3.5 py-2 rounded-xl text-sm font-medium transition-colors"
        :class="props.activeTab === tab.key
          ? 'bg-indigo-50 text-indigo-700 border border-indigo-100'
          : 'text-zinc-600 hover:bg-zinc-50 border border-transparent'"
        @click="emit('select', tab.key)"
      >
        {{ tab.label }}
        <span class="ml-1.5 text-xs rounded-full px-1.5 py-0.5"
          :class="props.activeTab === tab.key ? 'bg-indigo-100 text-indigo-700' : 'bg-zinc-100 text-zinc-600'"
        >
          {{ tab.count }}
        </span>
      </button>
    </div>
  </section>
</template>
