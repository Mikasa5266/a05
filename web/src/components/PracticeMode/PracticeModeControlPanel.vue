<script setup>
import { usePracticeStore } from "../../stores/practice";

const practiceStore = usePracticeStore();
</script>

<template>
  <div class="toolbar">
    <div class="row">
      <label>刷题模式</label>
      <select v-model="practiceStore.questionFilters.mode">
        <option value="single">单题训练</option>
        <option value="special">专项刷题</option>
      </select>
    </div>

    <div class="row">
      <label>难度</label>
      <select v-model="practiceStore.questionFilters.level">
        <option value="all">全部</option>
        <option
          v-for="item in practiceStore.levelEntries"
          :key="item.value"
          :value="item.value"
        >
          {{ item.label }}
        </option>
      </select>
    </div>

    <div class="row">
      <label>题型</label>
      <select v-model="practiceStore.questionFilters.questionType">
        <option value="all">全部</option>
        <option
          v-for="item in practiceStore.questionTypeOptions"
          :key="item.value"
          :value="item.value"
        >
          {{ item.label }}
        </option>
      </select>
    </div>

    <div class="row">
      <label>专项</label>
      <select
        v-model="practiceStore.questionFilters.specialty"
        :disabled="practiceStore.questionFilters.mode !== 'special'"
      >
        <option value="all">全部</option>
        <option
          v-for="item in practiceStore.specialtyOptions"
          :key="item.value"
          :value="item.value"
        >
          {{ item.label }}
        </option>
      </select>
    </div>

    <div class="row">
      <label>企业类型</label>
      <select v-model="practiceStore.questionFilters.companyType">
        <option value="all">全部</option>
        <option
          v-for="item in practiceStore.companyTypeOptions"
          :key="item.value"
          :value="item.value"
        >
          {{ item.label }}
        </option>
      </select>
    </div>

    <div class="row inline-row">
      <label><input v-model="practiceStore.questionFilters.timedMode" type="checkbox" /> 计时刷题</label>
      <input
        v-model.number="practiceStore.questionFilters.perQuestionSeconds"
        type="number"
        min="10"
        max="600"
        title="每题秒数"
      />
    </div>

    <button
      type="button"
      :disabled="practiceStore.loading.question || practiceStore.loading.answer"
      @click="practiceStore.loadQuestion"
    >
      {{ practiceStore.loading.question ? "加载中..." : "开始刷题" }}
    </button>
  </div>
</template>

<style scoped>
.toolbar {
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  padding: 14px;
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr)) auto;
  gap: 10px;
  margin-bottom: 12px;
  box-shadow: 0 8px 18px rgba(22, 77, 110, 0.06);
}

.row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.row label {
  font-size: 12px;
  color: #4d6783;
  font-weight: 700;
}

.inline-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.inline-row input[type="number"] {
  max-width: 90px;
}

@media (max-width: 1180px) {
  .toolbar {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
