<template>
  <div class="space-y-8">
    <header class="flex items-end justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-zinc-900">人才池</h1>
        <p class="text-zinc-500 mt-2">精准人才推送与管理，AI 认证评估结果一览</p>
      </div>
      <div class="flex items-center gap-3">
        <input v-model="searchQuery" placeholder="搜索候选人..." class="px-4 py-2.5 bg-zinc-50 border border-zinc-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 w-64" />
        <select v-model="filterPosition" class="px-4 py-2.5 bg-zinc-50 border border-zinc-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
          <option value="">全部岗位</option>
          <option value="frontend">前端开发</option>
          <option value="backend">后端开发</option>
          <option value="ai">AI工程师</option>
          <option value="product">产品经理</option>
        </select>
      </div>
    </header>

    <!-- Talent Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div v-for="(talent, idx) in filteredTalents" :key="idx"
        class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm hover:shadow-md transition-shadow"
      >
        <div class="flex items-start justify-between mb-4">
          <div class="flex items-center gap-3">
            <div class="h-12 w-12 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-600 font-bold">
              {{ talent.name.charAt(0) }}
            </div>
            <div>
              <div class="font-bold text-zinc-900">{{ talent.name }}</div>
              <div class="text-xs text-zinc-400">{{ talent.school }}</div>
            </div>
          </div>
          <span class="px-2 py-1 rounded-full text-xs font-bold"
            :class="talent.match >= 85 ? 'bg-emerald-50 text-emerald-600' : talent.match >= 70 ? 'bg-amber-50 text-amber-600' : 'bg-zinc-100 text-zinc-500'">
            {{ talent.match }}% 匹配
          </span>
        </div>

        <!-- Ability Bars -->
        <div class="space-y-2 mb-4">
          <div v-for="skill in talent.skills" :key="skill.name">
            <div class="flex items-center justify-between text-xs mb-1">
              <span class="text-zinc-500">{{ skill.name }}</span>
              <span class="font-medium text-zinc-700">{{ skill.value }}</span>
            </div>
            <div class="h-1.5 bg-zinc-100 rounded-full overflow-hidden">
              <div class="h-full bg-indigo-500 rounded-full" :style="{ width: skill.value + '%' }"></div>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2 flex-wrap mb-4">
          <span v-for="tag in talent.tags" :key="tag" class="px-2 py-1 bg-zinc-100 text-zinc-600 rounded-full text-xs">{{ tag }}</span>
        </div>

        <div class="flex items-center gap-2">
          <button class="flex-1 py-2 bg-indigo-600 text-white rounded-xl text-sm font-medium hover:bg-indigo-700 transition-colors">
            邀请面试
          </button>
          <button class="px-3 py-2 border border-zinc-200 text-zinc-600 rounded-xl text-sm hover:bg-zinc-50 transition-colors">
            收藏
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const searchQuery = ref('')
const filterPosition = ref('')

const talents = ref([
  { name: '张明', school: '北京大学', match: 95, tags: ['Java', 'Spring', '微服务'], skills: [{ name: '技术能力', value: 92 }, { name: '表达能力', value: 85 }, { name: '逻辑思维', value: 90 }] },
  { name: '李华', school: '清华大学', match: 88, tags: ['React', 'TypeScript', '前端'], skills: [{ name: '技术能力', value: 88 }, { name: '表达能力', value: 90 }, { name: '逻辑思维', value: 85 }] },
  { name: '王芳', school: '浙江大学', match: 82, tags: ['Python', 'TensorFlow', 'AI'], skills: [{ name: '技术能力', value: 85 }, { name: '表达能力', value: 78 }, { name: '逻辑思维', value: 88 }] },
  { name: '刘洋', school: '复旦大学', match: 79, tags: ['Go', '分布式', '后端'], skills: [{ name: '技术能力', value: 80 }, { name: '表达能力', value: 82 }, { name: '逻辑思维', value: 78 }] },
  { name: '陈浩', school: '上海交大', match: 91, tags: ['产品设计', '数据分析'], skills: [{ name: '技术能力', value: 75 }, { name: '表达能力', value: 95 }, { name: '逻辑思维', value: 92 }] },
  { name: '赵丽', school: '南京大学', match: 85, tags: ['测试', '自动化', 'CI/CD'], skills: [{ name: '技术能力', value: 82 }, { name: '表达能力', value: 88 }, { name: '逻辑思维', value: 80 }] },
])

const filteredTalents = computed(() => {
  return talents.value.filter(t => {
    if (searchQuery.value && !t.name.includes(searchQuery.value) && !t.tags.join(',').includes(searchQuery.value)) return false
    return true
  })
})
</script>
