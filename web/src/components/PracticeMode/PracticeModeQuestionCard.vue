<script setup>
import { computed } from "vue";

import { usePracticeStore } from "../../stores/practice";

const practiceStore = usePracticeStore();

const questionTitle = computed(() => {
  if (!practiceStore.currentQuestion?.id) {
    return "请先选择参数并开始刷题";
  }
  return `题目 #${practiceStore.currentQuestion.id}`;
});

const questionMetaItems = computed(() => {
  const question = practiceStore.currentQuestion;
  if (!question) return [];

  const levels = Object.fromEntries(
    practiceStore.levelEntries.map((item) => [item.value, item.label]),
  );
  const roleName =
    practiceStore.meta?.roles?.[question.role]?.name ||
    practiceStore.meta?.roles?.[question.position_code]?.name ||
    question.position_code ||
    question.role;

  return [
    roleName,
    levels[question.level] || question.level,
    question.question_type,
    question.points ? `考点：${question.points}` : "",
    question.company_type ? `适配：${question.company_type}` : "",
    question.difficulty_score ? `难度系数：${question.difficulty_score}` : "",
  ].filter(Boolean);
});

const timerText = computed(() => {
  if (!practiceStore.questionFilters.timedMode || !practiceStore.timerState.running) {
    return "计时：--";
  }
  return `计时：${Math.max(0, practiceStore.timerState.left)}s`;
});

const isTimerWarning = computed(
    () =>
      practiceStore.questionFilters.timedMode &&
      practiceStore.timerState.running &&
      practiceStore.timerState.left <= 15,
);

const favoriteLabel = computed(() =>
  practiceStore.currentQuestion?.is_favorite ? "取消收藏" : "收藏题目",
);

const defaultLoopHints = computed(() => practiceStore.pointLoop?.hints || []);
</script>

<template>
  <article class="question-card">
    <h3>{{ questionTitle }}</h3>

    <div class="question-meta">
      <span v-for="item in questionMetaItems" :key="item">{{ item }}</span>
    </div>

    <p class="stem">{{ practiceStore.currentQuestion?.stem || "" }}</p>

    <div v-if="practiceStore.currentQuestionHasOptions" class="options">
      <label
        v-for="option in practiceStore.currentQuestion?.options || []"
        :key="option.key"
        class="option-item"
      >
        <span class="option-radio">
          <input
            v-model="practiceStore.selectedOption"
            type="radio"
            name="optionAnswer"
            :value="option.key"
          />
        </span>
        <span class="option-content">
          <strong>{{ option.key }}.</strong>
          <span>{{ option.text }}</span>
        </span>
      </label>
    </div>

    <textarea
      v-else
      v-model="practiceStore.answerInput"
      placeholder="请输入你的答案（简答题/项目题）"
    ></textarea>

    <div class="actions">
      <button
        type="button"
        :disabled="practiceStore.loading.answer || !practiceStore.currentQuestion"
        @click="practiceStore.submitCurrentAnswer(false)"
      >
        {{ practiceStore.loading.answer ? "提交中..." : "提交答案" }}
      </button>
      <button
        type="button"
        class="secondary"
        :disabled="!practiceStore.canGoPrev"
        @click="practiceStore.moveHistory(-1)"
      >
        上一题
      </button>
      <button
        type="button"
        class="secondary"
        :disabled="practiceStore.loading.question"
        @click="practiceStore.loadQuestion"
      >
        下一题
      </button>
      <button
        type="button"
        class="secondary"
        :disabled="!practiceStore.currentQuestion"
        @click="practiceStore.showCurrentSolution"
      >
        查看答案(A)
      </button>
      <button
        type="button"
        class="secondary"
        :disabled="!practiceStore.currentQuestion"
        @click="practiceStore.toggleCurrentFavorite"
      >
        {{ favoriteLabel }}
      </button>
    </div>

    <div class="timer-bar" :class="{ warning: isTimerWarning }">{{ timerText }}</div>

    <label class="reason-label" for="practice-error-reason">
      若答错，可填写错误原因（可选）
    </label>
    <input
      id="practice-error-reason"
      v-model="practiceStore.errorReason"
      type="text"
      placeholder="例如：知识点遗忘 / 审题不清 / 公式混淆"
    />

    <div class="result-box">
      <div
        v-if="practiceStore.currentResult?.kind === 'feedback'"
        class="feedback-card"
        :class="practiceStore.currentResult.data?.is_correct ? 'success' : 'danger'"
      >
        <h4>
          {{
            practiceStore.currentResult.data?.is_correct
              ? "回答正确"
              : practiceStore.currentResult.autoTimeout
                ? "超时自动提交"
                : "回答待优化"
          }}
        </h4>
        <p><strong>参考答案：</strong>{{ practiceStore.currentResult.data?.standard_answer }}</p>
        <p><strong>解析：</strong>{{ practiceStore.currentResult.data?.analysis }}</p>
        <p v-if="practiceStore.currentResult.data?.tips">
          <strong>答题技巧：</strong>{{ practiceStore.currentResult.data?.tips }}
        </p>
        <p v-if="practiceStore.currentResult.data?.exemplar">
          <strong>高分范例：</strong>{{ practiceStore.currentResult.data?.exemplar }}
        </p>
      </div>

      <div
        v-else-if="practiceStore.currentResult?.kind === 'solution'"
        class="feedback-card success"
      >
        <h4>答案速览</h4>
        <p><strong>参考答案：</strong>{{ practiceStore.currentResult.data?.standard_answer }}</p>
        <p><strong>解析：</strong>{{ practiceStore.currentResult.data?.analysis }}</p>
        <p v-if="practiceStore.currentResult.data?.tips">
          <strong>答题技巧：</strong>{{ practiceStore.currentResult.data?.tips }}
        </p>
        <p v-if="practiceStore.currentResult.data?.exemplar">
          <strong>高分范例：</strong>{{ practiceStore.currentResult.data?.exemplar }}
        </p>
      </div>
    </div>

    <div class="point-loop">
      <div v-if="practiceStore.pointLoop?.type === 'packet'" class="loop-card">
        <h4>考点闭环：{{ practiceStore.pointLoop.packet?.point }}</h4>
        <p><strong>知识速记：</strong>{{ practiceStore.pointLoop.packet?.memo }}</p>
        <p>
          <strong>高频面试延伸：</strong>
          {{ practiceStore.pointLoop.packet?.interview_extensions?.join("；") }}
        </p>
        <p>
          <strong>当前考点进度：</strong>
          {{ practiceStore.pointLoop.pointState?.completion || 0 }}%
          （{{ practiceStore.pointLoop.pointState?.solved || 0 }}/{{ practiceStore.pointLoop.pointState?.total || 0 }}）
        </p>
      </div>

      <div v-else-if="practiceStore.pointLoop?.type === 'assessment'" class="loop-card">
        <h4>岗位能力测评报告</h4>
        <p><strong>综合得分：</strong>{{ practiceStore.pointLoop.assessment?.score }}</p>
        <p>
          <strong>建议求职企业：</strong>{{ practiceStore.pointLoop.assessment?.target_company_type }}
        </p>
        <p>
          <strong>待补考点：</strong>
          {{ practiceStore.pointLoop.assessment?.need_improve_points?.join("、") || "暂无" }}
        </p>
      </div>

      <div v-else class="loop-card">
        <h4>答题闭环提醒</h4>
        <p>进入作答后，这里会给出解析、学习包和同考点强化建议。</p>
        <ul>
          <li v-for="item in defaultLoopHints" :key="item">{{ item }}</li>
        </ul>
      </div>
    </div>
  </article>
</template>

<style scoped>
.question-card {
  border: 1px solid var(--line);
  border-radius: var(--radius);
  background: var(--panel);
  padding: 18px;
  box-shadow: 0 14px 24px rgba(22, 77, 110, 0.08);
}

.question-card h3 {
  margin-top: 0;
  color: var(--ink-strong);
}

.question-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 10px;
}

.question-meta span {
  background: var(--panel-soft);
  border: 1px solid #d3e7f4;
  border-radius: 999px;
  padding: 4px 10px;
  font-size: 12px;
  color: #245775;
}

.stem {
  min-height: 1.75rem;
  line-height: 1.75;
  color: #233f5a;
}

.options {
  min-height: 0.5rem;
  display: grid;
  gap: 8px;
  margin: 12px 0;
}

.option-item {
  border: 1px solid var(--line);
  border-radius: 10px;
  padding: 12px;
  display: grid;
  grid-template-columns: 28px 1fr;
  gap: 10px;
  align-items: center;
  background: #fbfeff;
  transition: border-color 0.16s ease, transform 0.16s ease;
  cursor: pointer;
}

.option-item:hover {
  border-color: #a7d0e4;
  transform: translateY(-1px);
}

.option-radio {
  display: grid;
  place-items: center;
}

.option-item input[type="radio"] {
  width: 18px;
  height: 18px;
  margin: 0;
  accent-color: var(--brand);
}

.option-content {
  display: inline-flex;
  gap: 8px;
  align-items: flex-start;
  color: #21465f;
  line-height: 1.55;
  text-align: left;
}

.option-content strong {
  min-width: 22px;
  color: #2b5876;
}

.option-item:has(input:checked) {
  border-color: #70b7d8;
  background: #eef9ff;
  box-shadow: inset 0 0 0 1px rgba(112, 183, 216, 0.35);
}

.actions {
  margin-top: 12px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.timer-bar {
  margin-top: 10px;
  padding: 8px 10px;
  border-radius: 8px;
  background: #e9f7ff;
  color: #205877;
  font-weight: 700;
  font-size: 13px;
}

.timer-bar.warning {
  background: #fff0f0;
  color: #9a2d2d;
}

.reason-label {
  margin-top: 12px;
  display: block;
  font-size: 12px;
  color: #4d6783;
  font-weight: 700;
}

.result-box {
  margin-top: 14px;
}

.feedback-card {
  border-radius: 12px;
  padding: 12px;
  border: 1px solid;
}

.feedback-card.success {
  background: #ecfbf4;
  border-color: #bfead3;
  color: #1f5b42;
}

.feedback-card.danger {
  background: #fff0f0;
  border-color: #f3c7c7;
  color: #6f2828;
}

.feedback-card h4 {
  margin: 0 0 10px;
}

.feedback-card p {
  margin: 6px 0;
}

.point-loop {
  margin-top: 12px;
}

.loop-card {
  border: 1px solid #d9eaf5;
  background: #f7fcff;
  border-radius: 12px;
  padding: 12px;
}

.loop-card h4 {
  margin: 0 0 8px;
  color: #194b69;
}

.loop-card p {
  margin: 6px 0;
  font-size: 13px;
  color: #40617a;
}

.loop-card ul {
  margin: 6px 0;
  padding-left: 18px;
}

.loop-card li {
  margin: 4px 0;
  cursor: pointer;
}
</style>
