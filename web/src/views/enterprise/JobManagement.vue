<template>
  <div class="space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900">岗位管理</h1>
      <p class="text-zinc-500 mt-2">管理企业岗位、能力图谱与招聘标准</p>
    </header>

    <div class="flex items-center gap-3 mb-4">
      <button @click="showCreateModal = true" class="px-5 py-2.5 bg-indigo-600 text-white rounded-xl text-sm font-medium hover:bg-indigo-700 transition-colors">
        + 新建岗位
      </button>
      <select v-model="filterStatus" class="px-4 py-2.5 bg-zinc-50 border border-zinc-200 rounded-xl text-sm">
        <option value="">全部状态</option>
        <option value="active">招聘中</option>
        <option value="paused">已暂停</option>
        <option value="closed">已关闭</option>
      </select>
    </div>

    <!-- Jobs Table -->
    <div class="bg-white rounded-3xl border border-zinc-100 shadow-sm overflow-hidden">
      <table class="w-full text-left text-sm">
        <thead class="bg-zinc-50 border-b border-zinc-100">
          <tr>
            <th class="px-6 py-4 font-medium text-zinc-500">岗位名称</th>
            <th class="px-6 py-4 font-medium text-zinc-500">能力图谱</th>
            <th class="px-6 py-4 font-medium text-zinc-500">候选人数</th>
            <th class="px-6 py-4 font-medium text-zinc-500">平均匹配度</th>
            <th class="px-6 py-4 font-medium text-zinc-500">状态</th>
            <th class="px-6 py-4 font-medium text-zinc-500 text-right">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-zinc-100">
          <tr v-for="(job, idx) in jobs" :key="idx" class="hover:bg-zinc-50/50 transition-colors">
            <td class="px-6 py-4">
              <div class="font-medium text-zinc-900">{{ job.title }}</div>
              <div class="text-xs text-zinc-400">{{ job.department }}</div>
            </td>
            <td class="px-6 py-4">
              <div class="flex items-center gap-1">
                <div v-for="d in job.dimensions" :key="d" class="px-2 py-0.5 bg-indigo-50 text-indigo-600 rounded text-xs">{{ d }}</div>
              </div>
            </td>
            <td class="px-6 py-4 font-medium text-zinc-900">{{ job.candidates }}</td>
            <td class="px-6 py-4">
              <span class="font-bold" :class="job.avgMatch >= 80 ? 'text-emerald-600' : 'text-amber-600'">{{ job.avgMatch }}%</span>
            </td>
            <td class="px-6 py-4">
              <span class="px-2 py-1 rounded-full text-xs font-medium"
                :class="job.status === 'active' ? 'bg-emerald-50 text-emerald-600' : job.status === 'paused' ? 'bg-amber-50 text-amber-600' : 'bg-zinc-100 text-zinc-500'">
                {{ job.status === 'active' ? '招聘中' : job.status === 'paused' ? '已暂停' : '已关闭' }}
              </span>
            </td>
            <td class="px-6 py-4 text-right">
              <button class="text-indigo-600 hover:text-indigo-700 font-medium text-sm mr-3">编辑</button>
              <button class="text-zinc-400 hover:text-zinc-600 text-sm">详情</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Ability Atlas Section -->
    <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
      <h2 class="text-lg font-bold text-zinc-900 mb-6">岗位能力图谱 · 360° 全景</h2>
      <p class="text-sm text-zinc-500 mb-6">为每个技术岗位构建全方位能力图谱，根据不同岗位设置差异化权重，对齐行业标杆企业真实招聘需求。</p>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div v-for="dim in abilityDimensions" :key="dim.name" class="p-4 rounded-2xl border border-zinc-100 text-center">
          <div class="text-2xl font-bold text-indigo-600 mb-1">{{ dim.weight }}%</div>
          <div class="text-sm font-medium text-zinc-700">{{ dim.name }}</div>
          <div class="text-xs text-zinc-400 mt-1">权重占比</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const showCreateModal = ref(false)
const filterStatus = ref('')

const jobs = ref([
  { title: 'Java后端工程师', department: '技术部', dimensions: ['技术', '逻辑', '系统设计'], candidates: 45, avgMatch: 82, status: 'active' },
  { title: '前端开发工程师', department: '技术部', dimensions: ['技术', '表达', '协作'], candidates: 38, avgMatch: 78, status: 'active' },
  { title: '产品经理', department: '产品部', dimensions: ['逻辑', '表达', '商业'], candidates: 22, avgMatch: 75, status: 'active' },
  { title: '数据分析师', department: '数据部', dimensions: ['技术', '逻辑', '分析'], candidates: 15, avgMatch: 85, status: 'paused' },
])

const abilityDimensions = ref([
  { name: '技术深度', weight: 35 },
  { name: '表达沟通', weight: 20 },
  { name: '逻辑思维', weight: 25 },
  { name: '行为素养', weight: 20 },
])
</script>
