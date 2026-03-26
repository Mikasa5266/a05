<script setup>
defineProps({
  position: {
    type: String,
    default: ''
  },
  currentQuestionIndex: {
    type: Number,
    default: 0
  }
})
</script>

<template>
  <div class="rounded-3xl p-5 border border-zinc-200 bg-gradient-to-br from-emerald-50 via-white to-cyan-50 shadow-xl flex flex-col gap-4 shrink-0 min-h-[150px]">
    <div>
      <p class="text-sm font-bold text-emerald-700">文字 + 语音模式进行中</p>
      <p class="text-xs text-zinc-500 mt-2">当前聚焦回答内容与逻辑质量，系统会持续展示实时语音转写与表达指标。</p>
    </div>
    <div class="grid grid-cols-2 gap-3 text-xs">
      <div class="rounded-xl bg-white border border-zinc-200 p-3">
        <p class="text-zinc-400">当前岗位</p>
        <p class="font-bold text-zinc-800 mt-1">{{ position }}</p>
      </div>
      <div class="rounded-xl bg-white border border-zinc-200 p-3">
        <p class="text-zinc-400">面试进度</p>
        <p class="font-bold text-zinc-800 mt-1">第 {{ currentQuestionIndex + 1 }} 题</p>
      </div>
    </div>
  </div>
</template>
