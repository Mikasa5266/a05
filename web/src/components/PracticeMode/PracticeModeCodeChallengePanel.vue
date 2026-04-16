<script setup>
import { computed, onMounted, ref, watch } from "vue";

import LiveCodeEditor from "../interview/LiveCodeEditor.vue";
import { usePracticeStore } from "../../stores/practice";

const practiceStore = usePracticeStore();

const showCodeEditor = ref(false);
const codeValue = ref("");
const codeLanguage = ref("javascript");
const codeSubmitting = ref(false);
const lastSubmittedSummary = ref("");

const DRAFT_STORAGE_KEY = "practice_code_challenge_drafts_v1";
const MAX_DRAFT_COUNT = 40;
const draftMap = ref({});

const currentDraftKey = computed(() => {
  const questionId = Number(practiceStore.currentQuestion?.id || 0);
  if (!questionId) return "";
  return `${String(practiceStore.currentRole || "unknown")}:${questionId}`;
});

const canUseCodeMode = computed(
  () => Boolean(practiceStore.currentQuestion && !practiceStore.currentQuestionHasOptions),
);

const currentQuestionForEditor = computed(() => {
  const question = practiceStore.currentQuestion;
  if (!question) {
    return {
      title: "代码实战（刷题模式）",
      content: "请先在左侧加载一道题目，再开启代码实战模式。",
    };
  }

  const point = String(question.points || "综合能力").trim();
  const stem = String(question.stem || "").trim();
  return {
    title: `代码实战 · 题目 #${question.id || "--"}`,
    content: [
      stem || "请根据题目要求完成代码实现。",
      "",
      `考点：${point}`,
      "要求：请提交可运行的核心逻辑，并在关键分支添加必要注释。",
    ].join("\n"),
  };
});

const loadDraftMap = () => {
  if (typeof window === "undefined" || !window.localStorage) return;
  try {
    const raw = window.localStorage.getItem(DRAFT_STORAGE_KEY);
    if (!raw) {
      draftMap.value = {};
      return;
    }
    const parsed = JSON.parse(raw);
    draftMap.value = parsed && typeof parsed === "object" ? parsed : {};
  } catch {
    draftMap.value = {};
  }
};

const persistDraftMap = () => {
  if (typeof window === "undefined" || !window.localStorage) return;
  try {
    window.localStorage.setItem(DRAFT_STORAGE_KEY, JSON.stringify(draftMap.value));
  } catch {
    // ignore storage failures
  }
};

const normalizeLanguage = (lang) => String(lang || "javascript").trim().toLowerCase() || "javascript";

const restoreDraftForCurrentQuestion = () => {
  const key = currentDraftKey.value;
  if (!key) {
    codeValue.value = "";
    codeLanguage.value = "javascript";
    return;
  }

  const draft = draftMap.value[key];
  if (!draft) {
    codeValue.value = "";
    codeLanguage.value = "javascript";
    return;
  }

  codeValue.value = String(draft.code || "");
  codeLanguage.value = normalizeLanguage(draft.language);
};

const saveDraftForCurrentQuestion = () => {
  const key = currentDraftKey.value;
  if (!key) return;

  const code = String(codeValue.value || "");
  const next = { ...draftMap.value };
  if (!code.trim()) {
    if (Object.prototype.hasOwnProperty.call(next, key)) {
      delete next[key];
      draftMap.value = next;
      persistDraftMap();
    }
    return;
  }

  next[key] = {
    code,
    language: normalizeLanguage(codeLanguage.value),
    updatedAt: Date.now(),
  };

  const trimmed = Object.entries(next)
    .sort((a, b) => Number(b[1]?.updatedAt || 0) - Number(a[1]?.updatedAt || 0))
    .slice(0, MAX_DRAFT_COUNT)
    .reduce((acc, [entryKey, value]) => {
      acc[entryKey] = value;
      return acc;
    }, {});

  draftMap.value = trimmed;
  persistDraftMap();
};

const clearDraftForCurrentQuestion = () => {
  const key = currentDraftKey.value;
  if (!key) return;

  if (!Object.prototype.hasOwnProperty.call(draftMap.value, key)) return;
  const next = { ...draftMap.value };
  delete next[key];
  draftMap.value = next;
  persistDraftMap();
};

const openCodeEditor = () => {
  if (!practiceStore.currentQuestion) {
    practiceStore.setStatus("请先加载题目，再开启代码实战", true);
    return;
  }
  if (practiceStore.currentQuestionHasOptions) {
    practiceStore.setStatus("当前为选择题，代码实战仅在简答/项目题开放", true);
    return;
  }
  showCodeEditor.value = true;
};

const closeCodeEditor = () => {
  if (codeSubmitting.value) return;
  showCodeEditor.value = false;
};

const submitCodeAnswer = async () => {
  if (codeSubmitting.value) return;

  const code = String(codeValue.value || "").trim();
  if (!code) {
    practiceStore.setStatus("请先编写代码后再提交", true);
    return;
  }

  codeSubmitting.value = true;
  try {
    const stem = String(practiceStore.currentQuestion?.stem || "").trim();
    const payload = [
      "[代码作答模式]",
      stem ? `题目：${stem}` : "",
      `语言：${codeLanguage.value}`,
      "",
      `\`\`\`${codeLanguage.value}`,
      code,
      "\`\`\`",
    ]
      .filter(Boolean)
      .join("\n");

    practiceStore.answerInput = payload;
    await practiceStore.submitCurrentAnswer(false);

    clearDraftForCurrentQuestion();
    lastSubmittedSummary.value = `${codeLanguage.value.toUpperCase()} · ${new Date().toLocaleString("zh-CN")}`;
    showCodeEditor.value = false;
  } finally {
    codeSubmitting.value = false;
  }
};

watch(
  () => currentDraftKey.value,
  () => {
    showCodeEditor.value = false;
    restoreDraftForCurrentQuestion();
    codeSubmitting.value = false;
  },
  { immediate: true },
);

watch(
  () => [codeValue.value, codeLanguage.value, currentDraftKey.value],
  () => {
    saveDraftForCurrentQuestion();
  },
);

onMounted(() => {
  loadDraftMap();
  restoreDraftForCurrentQuestion();
});
</script>

<template>
  <article class="code-panel">
    <div class="code-panel-head">
      <div>
        <h3>代码实战模式</h3>
        <p>将原代码面试能力完整迁移到刷题页，独立于普通文本作答流程。</p>
      </div>
      <button type="button" class="secondary" :disabled="!practiceStore.currentQuestion" @click="openCodeEditor">
        打开代码编辑器
      </button>
    </div>

    <p v-if="!practiceStore.currentQuestion" class="code-note">当前未加载题目</p>
    <p v-else-if="practiceStore.currentQuestionHasOptions" class="code-note warning">当前为选择题，请切换到简答/项目题后使用代码模式</p>
    <p v-else class="code-note ok">当前题目支持代码作答，可使用提交评分链路同步评估</p>

    <p v-if="lastSubmittedSummary" class="code-last">最近一次提交：{{ lastSubmittedSummary }}</p>

    <div
      v-if="showCodeEditor"
      class="fixed inset-0 z-220 bg-slate-950/70 backdrop-blur-sm flex items-center justify-center p-4 sm:p-6"
      @click.self="closeCodeEditor"
    >
      <div class="w-full max-w-5xl max-h-[90vh] overflow-hidden rounded-3xl border border-white/15 bg-slate-950/90 shadow-2xl">
        <div class="max-h-[90vh] overflow-y-auto custom-scrollbar p-4 sm:p-6">
          <LiveCodeEditor
            :visible="showCodeEditor"
            :model-value="codeValue"
            :question="currentQuestionForEditor"
            :language="codeLanguage"
            :submitting="codeSubmitting"
            :read-only="!canUseCodeMode"
            @update:modelValue="codeValue = $event"
            @language-change="codeLanguage = String($event || 'javascript')"
            @close="closeCodeEditor"
            @submit="submitCodeAnswer"
          />
        </div>
      </div>
    </div>
  </article>
</template>

<style scoped>
.code-panel {
  border: 1px solid var(--line);
  border-radius: var(--radius);
  background: var(--panel);
  padding: 14px;
  box-shadow: 0 10px 22px rgba(22, 77, 110, 0.08);
}

.code-panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.code-panel-head h3 {
  margin: 0;
  color: var(--ink-strong);
}

.code-panel-head p {
  margin: 6px 0 0;
  color: #567189;
  font-size: 12px;
}

.code-note {
  margin: 12px 0 0;
  font-size: 12px;
  color: #49647d;
}

.code-note.warning {
  color: #9a2d2d;
}

.code-note.ok {
  color: #1f8d67;
}

.code-last {
  margin: 8px 0 0;
  font-size: 12px;
  color: #6a8094;
}
</style>
