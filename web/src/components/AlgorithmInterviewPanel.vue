<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Loader2, PlayCircle, SkipForward, CheckCircle2, XCircle, TerminalSquare, Code2 } from 'lucide-vue-next'
import { getAlgorithmSession, runAlgorithmCode, skipAlgorithmProblem } from '../api/interview'

const props = defineProps({
  interviewId: {
    type: Number,
    required: true
  },
  difficulty: {
    type: String,
    default: 'campus_intern'
  }
})

const emit = defineEmits(['brief-updated', 'progress-updated', 'finished'])

const loading = ref(false)
const running = ref(false)
const skipping = ref(false)
const session = ref(null)
const currentIndex = ref(0)
const language = ref('java')
const code = ref('')
const runResult = ref(null)
const questionStates = ref([])

const languageTemplates = {
  java: `import java.util.*;

class Solution {
    public int solve(int[] nums) {
        // TODO: implement your algorithm here
        return 0;
    }
}`,
  cpp: `#include <bits/stdc++.h>
using namespace std;

class Solution {
public:
    int solve(vector<int>& nums) {
        // TODO: implement your algorithm here
        return 0;
    }
};`,
  python: `class Solution:
    def solve(self, nums):
        # TODO: implement your algorithm here
        return 0`,
  typescript: `function solve(nums: number[]): number {
  // TODO: implement your algorithm here
  return 0
}`
}

const languageLabels = {
  java: 'Java',
  cpp: 'C++',
  python: 'Python',
  typescript: 'TypeScript'
}

const languageOrder = ['java', 'cpp', 'python', 'typescript']

const currentProblem = computed(() => {
  const list = session.value?.problems || []
  return list[currentIndex.value] || null
})

const totalProblems = computed(() => (session.value?.problems || []).length)

const finishedCount = computed(() => {
  return questionStates.value.filter((item) => item.status === 'passed' || item.status === 'skipped').length
})

const summary = computed(() => {
  const passed = questionStates.value.filter((item) => item.status === 'passed').length
  const skipped = questionStates.value.filter((item) => item.status === 'skipped').length
  const failed = questionStates.value.filter((item) => item.status === 'failed').length
  return { passed, skipped, failed }
})

const isAllFinished = computed(() => {
  return totalProblems.value > 0 && finishedCount.value >= totalProblems.value
})

const updateParentBrief = () => {
  if (!currentProblem.value) {
    emit('brief-updated', '请保持算法思维，围绕边界条件与复杂度组织表达。')
    return
  }

  const p = currentProblem.value
  emit(
    'brief-updated',
    `请用${p.method_hint || '算法思维'}，在满足${p.requirement_hint || '时间与空间约束'}的情况下，实现如下算法题。`
  )
}

const updateParentProgress = () => {
  emit('progress-updated', {
    current: currentIndex.value + 1,
    total: totalProblems.value,
    finished: finishedCount.value,
    passed: summary.value.passed,
    skipped: summary.value.skipped,
    failed: summary.value.failed
  })
}

const applyTemplate = () => {
  code.value = languageTemplates[language.value] || languageTemplates.java
}

watch(language, () => {
  if (!code.value.trim()) {
    applyTemplate()
  }
})

watch(currentIndex, () => {
  runResult.value = null
  if (!code.value.trim()) {
    applyTemplate()
  }
  updateParentBrief()
  updateParentProgress()
})

const moveToNext = () => {
  if (currentIndex.value + 1 < totalProblems.value) {
    currentIndex.value += 1
    return
  }

  emit('finished', {
    total: totalProblems.value,
    passed: summary.value.passed,
    skipped: summary.value.skipped,
    failed: summary.value.failed
  })
}

const ensureStateSize = () => {
  const next = []
  const total = totalProblems.value
  for (let i = 0; i < total; i += 1) {
    next.push(questionStates.value[i] || { status: 'pending', failedCase: '' })
  }
  questionStates.value = next
}

const loadSession = async () => {
  loading.value = true
  try {
    const res = await getAlgorithmSession(props.interviewId, { difficulty: props.difficulty })
    session.value = {
      sessionId: res.session_id,
      problems: Array.isArray(res.problems) ? res.problems : []
    }
    ensureStateSize()
    applyTemplate()
    updateParentBrief()
    updateParentProgress()
  } catch (err) {
    ElMessage.error(`加载算法题失败: ${err?.response?.data?.error || err.message}`)
    session.value = { sessionId: '', problems: [] }
    updateParentBrief()
    updateParentProgress()
  } finally {
    loading.value = false
  }
}

const runCode = async () => {
  if (!currentProblem.value || running.value) return
  if (!code.value.trim()) {
    ElMessage.warning('请先输入代码')
    return
  }

  running.value = true
  runResult.value = null
  try {
    const res = await runAlgorithmCode(props.interviewId, {
      session_id: session.value?.sessionId || '',
      problem_id: currentProblem.value.id,
      language: language.value,
      code: code.value
    })

    const passed = !!res.passed
    runResult.value = {
      passed,
      failedCase: res.failed_case || '',
      message: res.message || (passed ? '测试通过' : '测试未通过')
    }

    questionStates.value[currentIndex.value] = {
      status: passed ? 'passed' : 'failed',
      failedCase: passed ? '' : (res.failed_case || '未知测试用例')
    }

    updateParentProgress()
  } catch (err) {
    runResult.value = {
      passed: false,
      failedCase: '',
      message: err?.response?.data?.error || err.message || '运行失败'
    }
    questionStates.value[currentIndex.value] = {
      status: 'failed',
      failedCase: '服务异常'
    }
    updateParentProgress()
  } finally {
    running.value = false
  }
}

const skipProblem = async () => {
  if (!currentProblem.value || skipping.value) return
  skipping.value = true
  try {
    await skipAlgorithmProblem(props.interviewId, {
      session_id: session.value?.sessionId || '',
      problem_id: currentProblem.value.id
    })
    questionStates.value[currentIndex.value] = { status: 'skipped', failedCase: '' }
    runResult.value = {
      passed: false,
      failedCase: '',
      message: '已跳过本题，按 0 分处理。'
    }
    updateParentProgress()
    moveToNext()
  } catch (err) {
    ElMessage.error(`跳过失败: ${err?.response?.data?.error || err.message}`)
  } finally {
    skipping.value = false
  }
}

const goNextAfterPass = () => {
  if (!runResult.value?.passed) return
  moveToNext()
}

onMounted(() => {
  loadSession()
})
</script>

<template>
  <div class="bg-white rounded-3xl border border-zinc-100 shadow-xl shadow-zinc-200/50 p-5 lg:p-6 min-h-[420px] flex flex-col gap-4">
    <div class="flex items-center justify-between gap-3">
      <div>
        <p class="text-xs font-semibold uppercase tracking-wider text-emerald-600">算法考察工作台</p>
        <h3 class="text-lg font-bold text-zinc-900 mt-1">仿真编码区域</h3>
      </div>
      <div class="text-right text-xs text-zinc-500">
        <p>进度 {{ finishedCount }}/{{ totalProblems }}</p>
        <p>通过 {{ summary.passed }} · 跳过 {{ summary.skipped }}</p>
      </div>
    </div>

    <div v-if="loading" class="flex-1 rounded-2xl border border-zinc-200 bg-zinc-50 flex items-center justify-center gap-2 text-zinc-500">
      <Loader2 class="w-4 h-4 animate-spin" />
      加载题目中...
    </div>

    <template v-else-if="!currentProblem">
      <div class="flex-1 rounded-2xl border border-zinc-200 bg-zinc-50 p-6 text-sm text-zinc-500 flex items-center justify-center text-center">
        暂无算法题可展示，请稍后重试。
      </div>
    </template>

    <template v-else>
      <div class="rounded-2xl border border-cyan-100 bg-gradient-to-br from-cyan-50 to-white p-4">
        <div class="flex items-center gap-2 text-cyan-700 text-xs font-semibold uppercase tracking-wider mb-2">
          <TerminalSquare class="w-3.5 h-3.5" />
          出题区（具体题面）
        </div>
        <h4 class="text-base font-bold text-zinc-900">{{ currentProblem.title }}</h4>
        <p class="text-sm text-zinc-700 mt-2 leading-relaxed whitespace-pre-wrap">{{ currentProblem.prompt }}</p>
        <div class="mt-3 flex flex-wrap gap-2">
          <span class="px-2 py-1 rounded-full text-[11px] bg-zinc-100 text-zinc-600">难度: {{ currentProblem.level || '中等' }}</span>
          <span class="px-2 py-1 rounded-full text-[11px] bg-zinc-100 text-zinc-600">{{ currentProblem.test_case_count || 0 }} 个测试案例</span>
          <span v-for="tag in currentProblem.tags || []" :key="tag" class="px-2 py-1 rounded-full text-[11px] bg-emerald-100 text-emerald-700">{{ tag }}</span>
        </div>
      </div>

      <div class="flex-1 grid grid-cols-1 gap-3 min-h-0">
        <div class="rounded-2xl border border-zinc-200 bg-zinc-50 p-3 flex items-center justify-between gap-3">
          <div class="flex items-center gap-2 text-sm text-zinc-700">
            <Code2 class="w-4 h-4 text-indigo-500" />
            代码区
          </div>
          <div class="algorithm-language-selector">
            <button
              v-for="key in languageOrder"
              :key="key"
              @click="language = key"
              class="algorithm-lang-btn px-2.5 py-1.5 text-xs rounded-lg border transition-colors"
              :class="language === key ? 'bg-indigo-600 text-white border-indigo-600' : 'bg-white text-zinc-600 border-zinc-200 hover:bg-zinc-100'"
            >
              <span class="algorithm-lang-label">{{ languageLabels[key] }}</span>
            </button>
          </div>
        </div>

        <textarea
          v-model="code"
          class="flex-1 min-h-[220px] max-h-[46vh] w-full rounded-2xl border border-zinc-200 bg-zinc-950 text-zinc-100 p-4 font-mono text-sm leading-6 focus:outline-none focus:ring-2 focus:ring-cyan-500/60 resize-y"
          placeholder="请在这里输入你的算法实现代码..."
        ></textarea>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
        <button
          @click="runCode"
          :disabled="running"
          class="sm:col-span-2 py-2.5 rounded-xl bg-emerald-600 text-white font-semibold text-sm hover:bg-emerald-700 transition-colors disabled:opacity-60 flex items-center justify-center gap-2"
        >
          <Loader2 v-if="running" class="w-4 h-4 animate-spin" />
          <PlayCircle v-else class="w-4 h-4" />
          运行测试
        </button>

        <button
          @click="skipProblem"
          :disabled="skipping"
          class="py-2.5 rounded-xl bg-white border border-zinc-200 text-zinc-700 font-semibold text-sm hover:bg-zinc-50 transition-colors disabled:opacity-60 flex items-center justify-center gap-2"
        >
          <Loader2 v-if="skipping" class="w-4 h-4 animate-spin" />
          <SkipForward v-else class="w-4 h-4" />
          跳过本题
        </button>
      </div>

      <div v-if="runResult" class="rounded-2xl border p-3 text-sm"
        :class="runResult.passed ? 'border-emerald-200 bg-emerald-50 text-emerald-700' : 'border-rose-200 bg-rose-50 text-rose-700'">
        <div class="flex items-start gap-2">
          <CheckCircle2 v-if="runResult.passed" class="w-4 h-4 mt-0.5" />
          <XCircle v-else class="w-4 h-4 mt-0.5" />
          <div>
            <p class="font-semibold">{{ runResult.message }}</p>
            <p v-if="!runResult.passed && runResult.failedCase" class="text-xs mt-1">未通过测试案例: {{ runResult.failedCase }}</p>
          </div>
        </div>
      </div>

      <button
        v-if="runResult?.passed"
        @click="goNextAfterPass"
        class="py-2.5 rounded-xl bg-zinc-900 text-white font-semibold text-sm hover:bg-zinc-800 transition-colors"
      >
        {{ currentIndex + 1 < totalProblems ? '进入下一题' : '结束算法考察' }}
      </button>

      <div v-if="isAllFinished" class="rounded-2xl border border-indigo-200 bg-indigo-50 p-3 text-sm text-indigo-700">
        算法考察已完成: 通过 {{ summary.passed }} 题，跳过 {{ summary.skipped }} 题。
      </div>
    </template>
  </div>
</template>

<style scoped>
.algorithm-language-selector {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
  max-width: 280px;
}

.algorithm-lang-btn {
  min-width: 0;
  max-width: 112px;
  flex: 1 1 calc(50% - 4px);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.algorithm-lang-label {
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 640px) {
  .algorithm-language-selector {
    max-width: 100%;
    width: 100%;
    justify-content: stretch;
  }

  .algorithm-lang-btn {
    max-width: none;
  }
}
</style>
