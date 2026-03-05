<template>
  <div class="space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900">人才推送</h1>
      <p class="text-zinc-500 mt-2">向合作企业精准推荐优秀学生</p>
    </header>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Recommended Students -->
      <div class="lg:col-span-2 space-y-4">
        <div v-for="(s, idx) in recommendedStudents" :key="idx" class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm">
          <div class="flex items-start justify-between mb-4">
            <div class="flex items-center gap-4">
              <div class="h-12 w-12 rounded-full bg-indigo-100 text-indigo-600 flex items-center justify-center font-bold">{{ s.name.charAt(0) }}</div>
              <div>
                <div class="font-bold text-zinc-900">{{ s.name }}</div>
                <div class="text-xs text-zinc-400">{{ s.major }} · 综合评分 {{ s.score }}</div>
              </div>
            </div>
            <span class="px-3 py-1 rounded-full text-xs font-bold bg-emerald-50 text-emerald-600">{{ s.status }}</span>
          </div>
          <div class="flex items-center gap-2 flex-wrap mb-4">
            <span v-for="tag in s.tags" :key="tag" class="px-2 py-1 bg-zinc-100 text-zinc-600 rounded-full text-xs">{{ tag }}</span>
          </div>
          <div class="flex items-center gap-2">
            <button class="flex-1 py-2 bg-indigo-600 text-white rounded-xl text-sm font-medium hover:bg-indigo-700 transition-colors">推送给企业</button>
            <button class="px-4 py-2 border border-zinc-200 text-zinc-600 rounded-xl text-sm hover:bg-zinc-50 transition-colors">查看详情</button>
          </div>
        </div>
      </div>

      <!-- Push History -->
      <div class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm">
        <h3 class="font-bold text-zinc-900 mb-4">推送记录</h3>
        <div class="space-y-4">
          <div v-for="(record, idx) in pushHistory" :key="idx" class="p-3 rounded-xl bg-zinc-50">
            <div class="text-sm font-medium text-zinc-700">{{ record.student }} → {{ record.company }}</div>
            <div class="text-xs text-zinc-400 mt-1">{{ record.time }} · {{ record.result }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const recommendedStudents = ref([
  { name: '小陈', major: '人工智能', score: 92, status: '待推送', tags: ['Python', 'TensorFlow', '深度学习'] },
  { name: '小孙', major: '软件工程', score: 90, status: '已推送', tags: ['Java', 'Spring Boot', '分布式'] },
  { name: '小赵', major: '计算机科学', score: 85, status: '待推送', tags: ['React', 'TypeScript', '全栈'] },
])

const pushHistory = ref([
  { student: '小孙', company: '华为', time: '2天前', result: '面试中' },
  { student: '小周', company: '腾讯', time: '1周前', result: '已录用' },
  { student: '小吴', company: '阿里', time: '2周前', result: '已录用' },
  { student: '小林', company: '字节跳动', time: '3周前', result: '待反馈' },
])
</script>
