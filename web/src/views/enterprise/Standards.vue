<template>
  <div class="space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900">岗位标准共建</h1>
      <p class="text-zinc-500 mt-2">企业与高校共同制定岗位能力标准，招聘数据实时更新</p>
    </header>

    <!-- Standard Builder -->
    <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
      <h2 class="text-lg font-bold text-zinc-900 mb-4">能力标准模板</h2>
      <p class="text-sm text-zinc-500 mb-6">基于企业真实岗位需求构建能力标准体系，与行业标杆对齐</p>
      
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div v-for="(template, idx) in templates" :key="idx" class="p-6 rounded-2xl border border-zinc-100 hover:border-indigo-200 transition-colors">
          <div class="flex items-center justify-between mb-4">
            <h3 class="font-bold text-zinc-900">{{ template.name }}</h3>
            <span class="px-2 py-1 bg-indigo-50 text-indigo-600 rounded-full text-xs font-medium">{{ template.version }}</span>
          </div>
          <div class="space-y-3 mb-4">
            <div v-for="dim in template.dimensions" :key="dim.name" class="flex items-center justify-between">
              <span class="text-sm text-zinc-600">{{ dim.name }}</span>
              <div class="flex items-center gap-2">
                <div class="w-24 h-1.5 bg-zinc-100 rounded-full overflow-hidden">
                  <div class="h-full bg-indigo-500 rounded-full" :style="{ width: dim.weight + '%' }"></div>
                </div>
                <span class="text-xs font-medium text-zinc-500 w-8">{{ dim.weight }}%</span>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <button class="flex-1 py-2 border border-indigo-200 text-indigo-600 rounded-xl text-sm font-medium hover:bg-indigo-50 transition-colors">编辑</button>
            <button class="px-3 py-2 border border-zinc-200 text-zinc-600 rounded-xl text-sm hover:bg-zinc-50 transition-colors">应用</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Data Sync -->
    <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
      <h2 class="text-lg font-bold text-zinc-900 mb-4">招聘数据更新体系</h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="p-5 rounded-2xl bg-zinc-50">
          <div class="text-2xl font-bold text-zinc-900 mb-1">147</div>
          <div class="text-sm text-zinc-500">本月新增面试数据</div>
        </div>
        <div class="p-5 rounded-2xl bg-zinc-50">
          <div class="text-2xl font-bold text-zinc-900 mb-1">23</div>
          <div class="text-sm text-zinc-500">标准更新次数</div>
        </div>
        <div class="p-5 rounded-2xl bg-zinc-50">
          <div class="text-2xl font-bold text-zinc-900 mb-1">8</div>
          <div class="text-sm text-zinc-500">合作高校数量</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const templates = ref([
  {
    name: 'Java后端工程师标准',
    version: 'v2.1',
    dimensions: [
      { name: '编程基础', weight: 30 },
      { name: '系统设计', weight: 25 },
      { name: '工程实践', weight: 20 },
      { name: '沟通协作', weight: 15 },
      { name: '学习能力', weight: 10 },
    ]
  },
  {
    name: '前端开发工程师标准',
    version: 'v1.8',
    dimensions: [
      { name: '前端技术', weight: 35 },
      { name: 'UI/UX感知', weight: 20 },
      { name: '工程化能力', weight: 20 },
      { name: '沟通表达', weight: 15 },
      { name: '创新思维', weight: 10 },
    ]
  },
])
</script>
