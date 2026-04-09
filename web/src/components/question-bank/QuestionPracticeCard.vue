<template>
  <section
    class="overflow-hidden rounded-[2rem] border border-slate-200/80 bg-white shadow-[0_24px_80px_-48px_rgba(15,23,42,0.35)]"
  >
    <div v-if="!question" class="space-y-6 p-7 md:p-8">
      <div
        class="inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-slate-100 text-slate-500"
      >
        <FileQuestion class="h-6 w-6" />
      </div>
      <div class="space-y-2">
        <p class="text-xs font-semibold uppercase tracking-[0.28em] text-slate-400">
          Practice Canvas
        </p>
        <h2 class="text-2xl font-semibold tracking-tight text-slate-950">
          先选择题单或随机抽一题
        </h2>
        <p class="max-w-2xl text-sm leading-7 text-slate-500">
          新工作台会用 Markdown 安全渲染题干、代码块和解析，避免再把原始标记直接暴露给页面。
        </p>
      </div>

      <button
        type="button"
        class="inline-flex items-center gap-2 rounded-2xl bg-slate-950 px-4 py-3 text-sm font-semibold text-white transition hover:bg-slate-800"
        @click="$emit('next')"
      >
        <Sparkles class="h-4 w-4" />
        抽取第一题
      </button>
    </div>

    <div v-else class="space-y-6 p-5 md:p-7">
      <header class="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
        <div class="space-y-3">
          <div class="flex flex-wrap items-center gap-2">
            <span class="rounded-full bg-sky-50 px-3 py-1 text-xs font-semibold text-sky-700">
              {{ question.position || positionLabel }}
            </span>
            <span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold text-slate-600">
              {{ question.level || "未分层" }}
            </span>
            <span
              v-if="question.question_type"
              class="rounded-full bg-violet-50 px-3 py-1 text-xs font-semibold text-violet-700"
            >
              {{ question.question_type }}
            </span>
            <span
              v-if="question.company_type"
              class="rounded-full bg-amber-50 px-3 py-1 text-xs font-semibold text-amber-700"
            >
              {{ question.company_type }}
            </span>
            <span
              v-if="collectionTitle"
              class="rounded-full bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-700"
            >
              {{ collectionTitle }}
            </span>
          </div>

          <div class="space-y-2">
            <p class="text-xs font-semibold uppercase tracking-[0.28em] text-slate-400">
              Question {{ question.id }}
            </p>
            <h2 class="text-2xl font-semibold tracking-tight text-slate-950 md:text-[2rem]">
              {{ question.title || "未命名题目" }}
            </h2>
          </div>
        </div>

        <div class="flex flex-col items-start gap-3 xl:items-end">
          <div class="flex flex-wrap items-center gap-2">
            <span
              v-if="question.specialty"
              class="rounded-full border border-slate-200 px-3 py-1 text-xs font-medium text-slate-600"
            >
              专项 {{ question.specialty }}
            </span>
            <span
              v-if="question.points"
              class="rounded-full border border-slate-200 px-3 py-1 text-xs font-medium text-slate-600"
            >
              考点 {{ question.points }}
            </span>
            <span
              class="rounded-full border border-slate-200 px-3 py-1 text-xs font-medium text-slate-600"
            >
              难度 {{ question.difficulty_score || 0 }}
            </span>
          </div>

          <div
            v-if="timedMode"
            class="rounded-2xl border px-4 py-3 text-right"
            :class="
              remainingSeconds !== null && remainingSeconds <= 20
                ? 'border-rose-200 bg-rose-50 text-rose-700'
                : 'border-sky-100 bg-sky-50 text-sky-700'
            "
          >
            <p class="text-[11px] font-semibold uppercase tracking-[0.24em] opacity-70">
              Timer
            </p>
            <p class="mt-1 text-xl font-semibold">
              {{ timerLabel }}
            </p>
          </div>
        </div>
      </header>

      <section class="rounded-[1.6rem] border border-slate-200 bg-slate-50/70 p-5 md:p-6">
        <p class="mb-3 text-xs font-semibold uppercase tracking-[0.26em] text-slate-400">
          题目正文
        </p>
        <QuestionMarkdown
          class="text-[15px] text-slate-700"
          :source="question.stem || question.title"
          empty-text="题干为空"
        />
      </section>

      <section
        v-if="hasOptions"
        class="space-y-4 rounded-[1.6rem] border border-slate-200 bg-white p-5 md:p-6"
      >
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p class="text-xs font-semibold uppercase tracking-[0.26em] text-slate-400">
              作答方式
            </p>
            <h3 class="mt-2 text-lg font-semibold text-slate-900">
              {{ normalizedAnswerMode === "multiple" ? "多选题" : "单选题" }}
            </h3>
          </div>

          <div class="inline-flex rounded-2xl border border-slate-200 bg-slate-50 p-1">
            <button
              type="button"
              class="rounded-2xl px-4 py-2 text-sm font-medium transition"
              :class="
                normalizedAnswerMode === 'single'
                  ? 'bg-white text-slate-950 shadow-sm'
                  : 'text-slate-500 hover:text-slate-900'
              "
              @click="updateAnswerMode('single')"
            >
              单选
            </button>
            <button
              type="button"
              class="rounded-2xl px-4 py-2 text-sm font-medium transition"
              :class="
                normalizedAnswerMode === 'multiple'
                  ? 'bg-white text-slate-950 shadow-sm'
                  : 'text-slate-500 hover:text-slate-900'
              "
              @click="updateAnswerMode('multiple')"
            >
              多选
            </button>
          </div>
        </div>

        <div class="grid gap-3">
          <button
            v-for="option in question.options"
            :key="option.key"
            type="button"
            class="group flex w-full items-start gap-4 rounded-[1.4rem] border px-4 py-4 text-left transition"
            :class="
              isChecked(option.key)
                ? 'border-sky-300 bg-sky-50 shadow-[0_16px_40px_-36px_rgba(14,165,233,0.9)]'
                : 'border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50'
            "
            @click="toggleOption(option.key)"
          >
            <span
              class="mt-0.5 inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full border text-sm font-semibold transition"
              :class="
                isChecked(option.key)
                  ? 'border-sky-400 bg-sky-500 text-white'
                  : 'border-slate-200 bg-slate-100 text-slate-600 group-hover:border-slate-300'
              "
            >
              {{ option.key }}
            </span>

            <QuestionMarkdown
              class="min-w-0 flex-1 text-sm text-slate-700"
              :source="option.text"
              empty-text="选项内容为空"
            />
          </button>
        </div>
      </section>

      <section v-else class="space-y-4 rounded-[1.6rem] border border-slate-200 bg-white p-5 md:p-6">
        <div class="space-y-1">
          <p class="text-xs font-semibold uppercase tracking-[0.26em] text-slate-400">
            简答作答区
          </p>
          <h3 class="text-lg font-semibold text-slate-900">写下你的回答</h3>
        </div>

        <textarea
          :value="textAnswer"
          rows="8"
          class="w-full rounded-[1.4rem] border border-slate-200 bg-slate-50 px-4 py-4 text-sm leading-7 text-slate-700 outline-none transition placeholder:text-slate-400 focus:border-sky-300 focus:bg-white focus:ring-4 focus:ring-sky-100"
          placeholder="支持粘贴 Markdown、代码片段和分点答案。"
          @input="handleTextAnswer"
        />
      </section>

      <div class="flex flex-wrap items-center gap-3">
        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-2xl bg-slate-950 px-5 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="submitting"
          @click="$emit('submit')"
        >
          <CheckCircle2 class="h-4 w-4" />
          {{ submitting ? "正在核对..." : "核对答案" }}
        </button>

        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-2xl border border-slate-200 bg-white px-5 py-3 text-sm font-semibold text-slate-700 transition hover:border-slate-300 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="revealing"
          @click="$emit('reveal')"
        >
          <BookOpenCheck class="h-4 w-4" />
          {{ revealing ? "展开中..." : "查看标准答案" }}
        </button>

        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-2xl border border-slate-200 bg-white px-5 py-3 text-sm font-semibold text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
          @click="$emit('next')"
        >
          <ArrowRight class="h-4 w-4" />
          {{ assessmentActive ? "下一道测评题" : "抽下一题" }}
        </button>
      </div>

      <div
        class="rounded-[1.5rem] border border-dashed border-slate-200 bg-slate-50/80 px-4 py-4 text-sm text-slate-500"
      >
        <p class="text-xs font-semibold uppercase tracking-[0.24em] text-slate-400">
          当前作答
        </p>
        <p class="mt-2 leading-7 text-slate-600">
          {{ answerPreview }}
        </p>
      </div>

      <section v-if="hasExplanation" class="space-y-5 border-t border-slate-200 pt-6">
        <div
          v-if="feedback"
          class="rounded-[1.6rem] border px-5 py-4"
          :class="
            feedback.is_correct
              ? 'border-emerald-200 bg-emerald-50 text-emerald-800'
              : 'border-rose-200 bg-rose-50 text-rose-800'
          "
        >
          <p class="text-xs font-semibold uppercase tracking-[0.24em] opacity-70">
            Answer Check
          </p>
          <p class="mt-2 text-base font-semibold">
            {{
              feedback.is_correct
                ? "这道题已经通过，可以继续推进到下一题。"
                : "答案未通过，建议先对照解析补齐关键点。"
            }}
          </p>
        </div>

        <div class="grid gap-5 xl:grid-cols-[minmax(0,0.92fr)_minmax(0,1.08fr)]">
          <article class="space-y-4 rounded-[1.6rem] border border-slate-200 bg-slate-50/70 p-5">
            <div class="space-y-1">
              <p class="text-xs font-semibold uppercase tracking-[0.24em] text-slate-400">
                Standard Answer
              </p>
              <h3 class="text-lg font-semibold text-slate-900">标准答案</h3>
            </div>
            <QuestionMarkdown
              class="text-sm text-slate-700"
              :source="resolvedStandardAnswer"
              empty-text="暂无标准答案"
            />
          </article>

          <article class="space-y-4 rounded-[1.6rem] border border-slate-200 bg-white p-5">
            <div class="space-y-1">
              <p class="text-xs font-semibold uppercase tracking-[0.24em] text-slate-400">
                Deep Analysis
              </p>
              <h3 class="text-lg font-semibold text-slate-900">详细解析</h3>
            </div>
            <QuestionMarkdown
              class="text-sm text-slate-700"
              :source="resolvedAnalysis"
              empty-text="暂无解析"
            />
          </article>
        </div>

        <div v-if="resolvedTips || resolvedExemplar" class="grid gap-5 xl:grid-cols-2">
          <article
            v-if="resolvedTips"
            class="rounded-[1.6rem] border border-amber-200 bg-amber-50/70 p-5"
          >
            <p class="text-xs font-semibold uppercase tracking-[0.24em] text-amber-500">
              Coach Tip
            </p>
            <QuestionMarkdown class="mt-3 text-sm text-amber-950" :source="resolvedTips" />
          </article>

          <article
            v-if="resolvedExemplar"
            class="rounded-[1.6rem] border border-violet-200 bg-violet-50/70 p-5"
          >
            <p class="text-xs font-semibold uppercase tracking-[0.24em] text-violet-500">
              Exemplar
            </p>
            <QuestionMarkdown class="mt-3 text-sm text-violet-950" :source="resolvedExemplar" />
          </article>
        </div>

        <div
          v-if="feedback?.matched_keywords?.length || feedback?.missing_keywords?.length"
          class="grid gap-5 xl:grid-cols-2"
        >
          <article
            v-if="feedback?.matched_keywords?.length"
            class="rounded-[1.6rem] border border-emerald-200 bg-emerald-50/80 p-5"
          >
            <p class="text-xs font-semibold uppercase tracking-[0.24em] text-emerald-600">
              Matched
            </p>
            <div class="mt-3 flex flex-wrap gap-2">
              <span
                v-for="item in feedback.matched_keywords"
                :key="`matched-${item}`"
                class="rounded-full bg-white px-3 py-1 text-xs font-semibold text-emerald-700"
              >
                {{ item }}
              </span>
            </div>
          </article>

          <article
            v-if="feedback?.missing_keywords?.length"
            class="rounded-[1.6rem] border border-rose-200 bg-rose-50/80 p-5"
          >
            <p class="text-xs font-semibold uppercase tracking-[0.24em] text-rose-600">
              Missing
            </p>
            <div class="mt-3 flex flex-wrap gap-2">
              <span
                v-for="item in feedback.missing_keywords"
                :key="`missing-${item}`"
                class="rounded-full bg-white px-3 py-1 text-xs font-semibold text-rose-700"
              >
                {{ item }}
              </span>
            </div>
          </article>
        </div>
      </section>
    </div>
  </section>
</template>

<script setup>
import { computed } from "vue";
import {
  ArrowRight,
  BookOpenCheck,
  CheckCircle2,
  FileQuestion,
  Sparkles,
} from "lucide-vue-next";
import QuestionMarkdown from "./QuestionMarkdown.vue";

const props = defineProps({
  question: {
    type: Object,
    default: null,
  },
  positionLabel: {
    type: String,
    default: "",
  },
  collectionTitle: {
    type: String,
    default: "",
  },
  selectedSingle: {
    type: String,
    default: "",
  },
  selectedMultiple: {
    type: Array,
    default: () => [],
  },
  textAnswer: {
    type: String,
    default: "",
  },
  answerMode: {
    type: String,
    default: "single",
  },
  feedback: {
    type: Object,
    default: null,
  },
  solution: {
    type: Object,
    default: null,
  },
  timedMode: {
    type: Boolean,
    default: false,
  },
  remainingSeconds: {
    type: Number,
    default: null,
  },
  submitting: {
    type: Boolean,
    default: false,
  },
  revealing: {
    type: Boolean,
    default: false,
  },
  assessmentActive: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits([
  "update:selectedSingle",
  "update:selectedMultiple",
  "update:textAnswer",
  "update:answerMode",
  "submit",
  "reveal",
  "next",
]);

const hasOptions = computed(() => Boolean(props.question?.options?.length));

const normalizedAnswerMode = computed(() => {
  if (!hasOptions.value) {
    return "text";
  }
  return props.answerMode === "multiple" ? "multiple" : "single";
});

const toggleOption = (key) => {
  if (!hasOptions.value) {
    return;
  }

  if (normalizedAnswerMode.value === "multiple") {
    const current = new Set(
      Array.isArray(props.selectedMultiple) ? props.selectedMultiple : [],
    );
    if (current.has(key)) {
      current.delete(key);
    } else {
      current.add(key);
    }
    emit("update:selectedMultiple", Array.from(current).sort());
    return;
  }

  emit("update:selectedSingle", key);
};

const isChecked = (key) => {
  if (normalizedAnswerMode.value === "multiple") {
    return Array.isArray(props.selectedMultiple)
      ? props.selectedMultiple.includes(key)
      : false;
  }
  return props.selectedSingle === key;
};

const updateAnswerMode = (mode) => {
  emit("update:answerMode", mode);
  emit("update:selectedSingle", "");
  emit("update:selectedMultiple", []);
};

const handleTextAnswer = (event) => {
  emit("update:textAnswer", event?.target?.value || "");
};

const answerPreview = computed(() => {
  if (!props.question) {
    return "未作答";
  }
  if (hasOptions.value) {
    if (normalizedAnswerMode.value === "multiple") {
      return props.selectedMultiple?.length
        ? props.selectedMultiple.join(", ")
        : "尚未选择选项";
    }
    return props.selectedSingle || "尚未选择选项";
  }
  return String(props.textAnswer || "").trim() || "尚未填写答案";
});

const resolvedStandardAnswer = computed(
  () =>
    props.feedback?.standard_answer ||
    props.solution?.standard_answer ||
    props.question?.standard_answer ||
    "",
);

const resolvedAnalysis = computed(
  () => props.feedback?.analysis || props.solution?.analysis || "",
);

const resolvedTips = computed(
  () => props.feedback?.tips || props.solution?.tips || "",
);

const resolvedExemplar = computed(
  () => props.feedback?.exemplar || props.solution?.exemplar || "",
);

const hasExplanation = computed(
  () =>
    Boolean(props.feedback) ||
    Boolean(props.solution) ||
    Boolean(resolvedStandardAnswer.value) ||
    Boolean(resolvedAnalysis.value) ||
    Boolean(resolvedTips.value) ||
    Boolean(resolvedExemplar.value),
);

const timerLabel = computed(() => {
  if (!props.timedMode || props.remainingSeconds === null) {
    return "不限时";
  }
  const safe = Math.max(0, props.remainingSeconds);
  const minutes = Math.floor(safe / 60)
    .toString()
    .padStart(2, "0");
  const seconds = (safe % 60).toString().padStart(2, "0");
  return `${minutes}:${seconds}`;
});
</script>
