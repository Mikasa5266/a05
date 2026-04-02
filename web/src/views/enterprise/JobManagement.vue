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

    <!-- Desktop: Element Plus Table -->
    <div class="bg-white rounded-3xl border border-zinc-100 shadow-sm overflow-hidden max-md:hidden">
      <el-table :data="jobs" stripe class="w-full">
        <el-table-column label="岗位名称" min-width="220">
          <template #default="{ row }">
            <div class="py-1">
              <div class="font-medium text-zinc-900">{{ row.title }}</div>
              <div class="text-xs text-zinc-400">{{ row.department }}</div>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="能力图谱" min-width="220">
          <template #default="{ row }">
            <div class="flex flex-wrap items-center gap-1">
              <div v-for="d in row.dimensions" :key="d" class="px-2 py-0.5 bg-indigo-50 text-indigo-600 rounded text-xs">{{ d }}</div>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="candidates" label="候选人数" width="110" />

        <el-table-column label="平均匹配度" width="120">
          <template #default="{ row }">
            <span class="font-bold" :class="row.avgMatch >= 80 ? 'text-emerald-600' : 'text-amber-600'">{{ row.avgMatch }}%</span>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <span class="px-2 py-1 rounded-full text-xs font-medium"
              :class="row.status === 'active' ? 'bg-emerald-50 text-emerald-600' : row.status === 'paused' ? 'bg-amber-50 text-amber-600' : 'bg-zinc-100 text-zinc-500'">
              {{ row.status === 'active' ? '招聘中' : row.status === 'paused' ? '已暂停' : '已关闭' }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="140" align="right">
          <template #default>
            <button class="text-indigo-600 hover:text-indigo-700 font-medium text-sm mr-3">编辑</button>
            <button class="text-zinc-400 hover:text-zinc-600 text-sm">详情</button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- Mobile: Card List -->
    <div class="md:hidden space-y-3">
      <div
        v-for="(job, idx) in jobs"
        :key="idx"
        class="bg-white rounded-2xl border border-zinc-100 shadow-sm p-4 space-y-3"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <div class="font-semibold text-zinc-900">{{ job.title }}</div>
            <div class="text-xs text-zinc-400 mt-1">{{ job.department }}</div>
          </div>
          <span class="px-2 py-1 rounded-full text-xs font-medium whitespace-nowrap"
            :class="job.status === 'active' ? 'bg-emerald-50 text-emerald-600' : job.status === 'paused' ? 'bg-amber-50 text-amber-600' : 'bg-zinc-100 text-zinc-500'">
            {{ job.status === 'active' ? '招聘中' : job.status === 'paused' ? '已暂停' : '已关闭' }}
          </span>
        </div>

        <div class="flex flex-wrap items-center gap-1">
          <div v-for="d in job.dimensions" :key="d" class="px-2 py-0.5 bg-indigo-50 text-indigo-600 rounded text-xs">{{ d }}</div>
        </div>

        <div class="grid grid-cols-2 gap-3 text-sm">
          <div class="rounded-xl bg-zinc-50 px-3 py-2">
            <div class="text-zinc-500 text-xs">候选人数</div>
            <div class="font-semibold text-zinc-900 mt-1">{{ job.candidates }}</div>
          </div>
          <div class="rounded-xl bg-zinc-50 px-3 py-2">
            <div class="text-zinc-500 text-xs">平均匹配度</div>
            <div class="font-semibold mt-1" :class="job.avgMatch >= 80 ? 'text-emerald-600' : 'text-amber-600'">{{ job.avgMatch }}%</div>
          </div>
        </div>

        <div class="flex justify-end gap-3 pt-1">
          <button class="text-indigo-600 hover:text-indigo-700 font-medium text-sm">编辑</button>
          <button class="text-zinc-500 hover:text-zinc-700 text-sm">详情</button>
        </div>
      </div>
    </div>

    <!-- Desktop Dialog -->
    <el-dialog
      v-if="!isMobile"
      v-model="showCreateModal"
      title="新建岗位"
      :width="dialogWidth"
      destroy-on-close
    >
      <el-form
        :model="createForm"
        label-width="96px"
        :label-position="formLabelPosition"
        class="space-y-1"
      >
        <el-form-item label="岗位名称">
          <el-input v-model="createForm.title" placeholder="请输入岗位名称" style="--el-input-height: 44px" />
        </el-form-item>

        <el-form-item label="所属部门">
          <el-input v-model="createForm.department" placeholder="请输入所属部门" style="--el-input-height: 44px" />
        </el-form-item>

        <el-form-item label="状态">
          <el-select v-model="createForm.status" placeholder="请选择状态" class="w-full" style="--el-input-height: 44px">
            <el-option label="招聘中" value="active" />
            <el-option label="已暂停" value="paused" />
            <el-option label="已关闭" value="closed" />
          </el-select>
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="flex justify-end gap-2">
          <el-button @click="showCreateModal = false">取消</el-button>
          <el-button type="primary" @click="handleCreateSubmit">保存</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- Mobile Drawer For Complex Form -->
    <el-drawer
      v-else
      v-model="showCreateModal"
      direction="btt"
      size="88%"
      title="新建岗位"
      destroy-on-close
      :with-header="true"
    >
      <el-form
        :model="createForm"
        label-width="96px"
        :label-position="formLabelPosition"
        class="space-y-1"
      >
        <el-form-item label="岗位名称">
          <el-input v-model="createForm.title" placeholder="请输入岗位名称" style="--el-input-height: 44px" />
        </el-form-item>

        <el-form-item label="所属部门">
          <el-input v-model="createForm.department" placeholder="请输入所属部门" style="--el-input-height: 44px" />
        </el-form-item>

        <el-form-item label="状态">
          <el-select v-model="createForm.status" placeholder="请选择状态" class="w-full" style="--el-input-height: 44px">
            <el-option label="招聘中" value="active" />
            <el-option label="已暂停" value="paused" />
            <el-option label="已关闭" value="closed" />
          </el-select>
        </el-form-item>

        <div class="pt-2 flex gap-2">
          <el-button class="flex-1" @click="showCreateModal = false">取消</el-button>
          <el-button class="flex-1" type="primary" @click="handleCreateSubmit">保存</el-button>
        </div>
      </el-form>
    </el-drawer>

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
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'

const showCreateModal = ref(false)
const filterStatus = ref('')
const isMobile = ref(false)

let mobileMediaQuery = null

const createForm = reactive({
  title: '',
  department: '',
  status: 'active',
})

const dialogWidth = computed(() => (isMobile.value ? '95%' : '680px'))
const formLabelPosition = computed(() => (isMobile.value ? 'top' : 'right'))

const syncMobileState = () => {
  if (!mobileMediaQuery) return
  isMobile.value = mobileMediaQuery.matches
}

const handleCreateSubmit = () => {
  showCreateModal.value = false
}

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

onMounted(() => {
  mobileMediaQuery = window.matchMedia('(max-width: 767px)')
  syncMobileState()
  mobileMediaQuery.addEventListener('change', syncMobileState)
})

onBeforeUnmount(() => {
  if (mobileMediaQuery) {
    mobileMediaQuery.removeEventListener('change', syncMobileState)
  }
})
</script>
