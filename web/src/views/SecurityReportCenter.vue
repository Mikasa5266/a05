<template>
  <div class="space-y-6">
    <header class="rounded-3xl border border-zinc-200 bg-white p-6 shadow-sm">
      <h1 class="text-2xl font-black text-zinc-900">安全举报中心</h1>
      <p class="mt-2 text-sm text-zinc-500">
        用于举报违法有害信息、骚扰行为和异常传播风险。我们会在 24 小时内处理并保留处置记录。
      </p>
      <div class="mt-4 flex flex-wrap gap-3 text-xs text-zinc-500">
        <span class="rounded-full bg-zinc-100 px-3 py-1">投诉电话：400-000-0000</span>
        <span class="rounded-full bg-zinc-100 px-3 py-1">举报邮箱：security@example.com</span>
      </div>
    </header>

    <section class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <div class="rounded-3xl border border-zinc-200 bg-white p-6 shadow-sm lg:col-span-2">
        <h2 class="mb-4 text-lg font-bold text-zinc-900">提交举报</h2>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <label class="block">
            <div class="mb-1 text-xs font-bold uppercase text-zinc-400">举报对象类型</div>
            <select v-model="form.targetType" class="w-full rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm">
              <option value="post">帖子</option>
              <option value="comment">评论</option>
              <option value="user">用户</option>
              <option value="other">其他</option>
            </select>
          </label>

          <label class="block">
            <div class="mb-1 text-xs font-bold uppercase text-zinc-400">对象 ID（可选）</div>
            <input
              v-model.number="form.targetId"
              type="number"
              min="0"
              class="w-full rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm"
              placeholder="例如 123"
            />
          </label>

          <label class="block md:col-span-2">
            <div class="mb-1 text-xs font-bold uppercase text-zinc-400">举报原因</div>
            <select v-model="form.reason" class="w-full rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm">
              <option value="违法有害信息">违法有害信息</option>
              <option value="侮辱谩骂或骚扰">侮辱谩骂或骚扰</option>
              <option value="虚假信息或冒充">虚假信息或冒充</option>
              <option value="疑似恶意传播扩散">疑似恶意传播扩散</option>
              <option value="其他">其他</option>
            </select>
          </label>

          <label class="block md:col-span-2">
            <div class="mb-1 text-xs font-bold uppercase text-zinc-400">补充说明</div>
            <textarea
              v-model.trim="form.description"
              rows="4"
              maxlength="1000"
              class="w-full resize-none rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm"
              placeholder="请描述问题上下文，便于平台快速核查"
            />
          </label>

          <label class="block md:col-span-2">
            <div class="mb-1 text-xs font-bold uppercase text-zinc-400">联系方式（可选）</div>
            <input
              v-model.trim="form.contact"
              maxlength="120"
              class="w-full rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm"
              placeholder="手机号或邮箱"
            />
          </label>
        </div>

        <div class="mt-5 flex items-center justify-end gap-3">
          <button class="rounded-lg px-4 py-2 text-sm text-zinc-500 hover:bg-zinc-100" @click="resetForm">重置</button>
          <button
            class="rounded-xl bg-rose-600 px-5 py-2 text-sm font-semibold text-white hover:bg-rose-700 disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="submitting || !form.reason"
            @click="submitReport"
          >
            {{ submitting ? '提交中...' : '提交举报' }}
          </button>
        </div>
      </div>

      <div class="rounded-3xl border border-zinc-200 bg-white p-6 shadow-sm">
        <h2 class="mb-3 text-base font-bold text-zinc-900">处置说明</h2>
        <ul class="space-y-2 text-sm text-zinc-600">
          <li>1. 受理后进入审核队列。</li>
          <li>2. 违规内容将被删除或限制展示。</li>
          <li>3. 处置结果会在“我的举报记录”展示。</li>
          <li>4. 严重违规将按要求协助监管取证。</li>
        </ul>
      </div>
    </section>

    <section class="rounded-3xl border border-zinc-200 bg-white p-6 shadow-sm">
      <div class="mb-4 flex items-center justify-between">
        <h2 class="text-lg font-bold text-zinc-900">我的举报记录</h2>
        <button class="rounded-lg px-3 py-1 text-xs text-zinc-500 hover:bg-zinc-100" @click="fetchMyReports">刷新</button>
      </div>

      <div v-if="loading" class="py-8 text-center text-sm text-zinc-400">加载中...</div>
      <div v-else-if="reports.length === 0" class="py-8 text-center text-sm text-zinc-400">暂无举报记录</div>
      <div v-else class="space-y-3">
        <article v-for="item in reports" :key="item.id" class="rounded-2xl border border-zinc-100 bg-zinc-50 p-4">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div class="text-sm font-semibold text-zinc-900">
              {{ item.reason }}
              <span class="ml-2 text-xs font-normal text-zinc-500">
                {{ item.target_type }}#{{ item.target_id || 0 }}
              </span>
            </div>
            <span class="rounded-full px-2.5 py-1 text-xs font-semibold" :class="statusClass(item.status)">
              {{ statusText(item.status) }}
            </span>
          </div>
          <p class="mt-2 whitespace-pre-wrap wrap-break-word text-sm text-zinc-600">{{ item.description || '（无补充说明）' }}</p>
          <div class="mt-2 text-xs text-zinc-400">
            提交时间：{{ formatTime(item.created_at) }}
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createSecurityReport, getMySecurityReports } from '../api/community'

const route = useRoute()

const submitting = ref(false)
const loading = ref(false)
const reports = ref([])

const form = reactive({
  targetType: 'post',
  targetId: 0,
  reason: '违法有害信息',
  description: '',
  contact: ''
})

const fillFromQuery = () => {
  const targetType = String(route.query?.target_type || '').trim().toLowerCase()
  if (['post', 'comment', 'user', 'other'].includes(targetType)) {
    form.targetType = targetType
  }

  const targetID = Number(route.query?.target_id || 0)
  form.targetId = Number.isFinite(targetID) && targetID > 0 ? targetID : 0

  const title = String(route.query?.title || '').trim()
  if (title && !form.description) {
    form.description = `来源内容：${title}`
  }
}

const resetForm = () => {
  form.targetType = 'post'
  form.targetId = 0
  form.reason = '违法有害信息'
  form.description = ''
  form.contact = ''
}

const submitReport = async () => {
  if (submitting.value) return
  submitting.value = true
  try {
    await createSecurityReport({
      target_type: form.targetType,
      target_id: Number(form.targetId) > 0 ? Number(form.targetId) : 0,
      reason: form.reason,
      description: form.description,
      contact: form.contact
    })
    ElMessage.success('举报已提交，平台将尽快处理')
    resetForm()
    await fetchMyReports()
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || error?.message || '提交失败')
  } finally {
    submitting.value = false
  }
}

const fetchMyReports = async () => {
  loading.value = true
  try {
    const res = await getMySecurityReports({ page: 1, page_size: 20 })
    reports.value = Array.isArray(res?.reports) ? res.reports : []
  } catch (error) {
    reports.value = []
    ElMessage.error(error?.response?.data?.error || error?.message || '加载举报记录失败')
  } finally {
    loading.value = false
  }
}

const statusText = (status) => {
  const normalized = String(status || '').toLowerCase()
  if (normalized === 'processing') return '处理中'
  if (normalized === 'resolved') return '已处置'
  if (normalized === 'rejected') return '已驳回'
  return '待处理'
}

const statusClass = (status) => {
  const normalized = String(status || '').toLowerCase()
  if (normalized === 'processing') return 'bg-amber-100 text-amber-700'
  if (normalized === 'resolved') return 'bg-emerald-100 text-emerald-700'
  if (normalized === 'rejected') return 'bg-zinc-200 text-zinc-700'
  return 'bg-rose-100 text-rose-700'
}

const formatTime = (input) => {
  if (!input) return '--'
  try {
    return new Date(input).toLocaleString('zh-CN')
  } catch (_) {
    return '--'
  }
}

onMounted(async () => {
  fillFromQuery()
  await fetchMyReports()
})
</script>
