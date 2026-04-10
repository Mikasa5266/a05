<script setup>
import { computed } from "vue";

import { usePracticeStore } from "../../stores/practice";

defineProps({
  active: {
    type: Boolean,
    default: false,
  },
});

const practiceStore = usePracticeStore();

const pageInfo = computed(() => {
  const pagination = practiceStore.bank.pagination || {};
  const page = pagination.page || 1;
  const totalPages = pagination.total_pages || 1;
  return `${page} / ${totalPages}`;
});

const statusPills = computed(() => [
  { key: "todo", label: `未做 ${practiceStore.bank.statusStats?.todo || 0}` },
  { key: "solved", label: `已通过 ${practiceStore.bank.statusStats?.solved || 0}` },
  { key: "wrong", label: `待复习 ${practiceStore.bank.statusStats?.wrong || 0}` },
]);

const statusLabel = (status) => {
  if (status === "solved") return "已通过";
  if (status === "wrong") return "待复习";
  return "未做";
};

const applySearch = async () => {
  practiceStore.bankFilters.page = 1;
  await practiceStore.loadQuestionPool();
};

const goPrevPage = async () => {
  if ((practiceStore.bank.pagination?.page || 1) <= 1) return;
  practiceStore.bankFilters.page -= 1;
  await practiceStore.loadQuestionPool();
};

const goNextPage = async () => {
  const page = practiceStore.bank.pagination?.page || 1;
  const totalPages = practiceStore.bank.pagination?.total_pages || 1;
  if (page >= totalPages) return;
  practiceStore.bankFilters.page += 1;
  await practiceStore.loadQuestionPool();
};

const openQuestionFromBank = async (questionId) => {
  await practiceStore.loadQuestionById(questionId, "已从题库进入做题模式");
};
</script>

<template>
  <section class="view" :class="{ active }">
    <div class="bank-layout">
      <aside class="bank-filter">
        <h3>分类筛选</h3>

        <div class="row">
          <label>关键词</label>
          <input
            v-model="practiceStore.bankFilters.keyword"
            placeholder="搜索题干关键词"
          />
        </div>

        <div class="row">
          <label>难度层</label>
          <select v-model="practiceStore.bankFilters.level">
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
          <select v-model="practiceStore.bankFilters.questionType">
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
          <select v-model="practiceStore.bankFilters.specialty">
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
          <label>考点</label>
          <select v-model="practiceStore.bankFilters.point">
            <option value="all">全部</option>
            <option
              v-for="item in practiceStore.pointOptions"
              :key="item.value"
              :value="item.value"
            >
              {{ item.label }}
            </option>
          </select>
        </div>

        <div class="row">
          <label>企业类型</label>
          <select v-model="practiceStore.bankFilters.companyType">
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

        <div class="row">
          <label>状态</label>
          <select v-model="practiceStore.bankFilters.status">
            <option value="all">全部</option>
            <option value="todo">未做</option>
            <option value="solved">已通过</option>
            <option value="wrong">待复习</option>
          </select>
        </div>

        <label class="checkbox-line">
          <input v-model="practiceStore.bankFilters.favoriteOnly" type="checkbox" />
          仅看收藏
        </label>

        <button
          type="button"
          :disabled="practiceStore.bank.loadingPool"
          @click="applySearch"
        >
          {{ practiceStore.bank.loadingPool ? "筛选中..." : "筛选题目" }}
        </button>

        <div class="assessment-box">
          <h4>岗位能力测评</h4>
          <p>自动抽取各考点题目，完成后生成岗位能力报告。</p>
          <button
            type="button"
            class="secondary"
            :disabled="practiceStore.loading.assessment"
            @click="practiceStore.startAssessment"
          >
            {{ practiceStore.loading.assessment ? "启动中..." : "开始能力测评（12题）" }}
          </button>
        </div>
      </aside>

      <section class="bank-main">
        <div class="bank-tabs">
          <button
            type="button"
            class="tab-btn"
            :class="{ active: practiceStore.bankTab === 'pool' }"
            @click="practiceStore.switchBankTab('pool')"
          >
            分类题库
          </button>
          <button
            type="button"
            class="tab-btn"
            :class="{ active: practiceStore.bankTab === 'lists' }"
            @click="practiceStore.switchBankTab('lists')"
          >
            岗位题单
          </button>
        </div>

        <div class="tab-panel" :class="{ active: practiceStore.bankTab === 'pool' }">
          <div class="pool-head">
            <div class="status-pills">
              <span v-for="item in statusPills" :key="item.key" class="pill">
                {{ item.label }}
              </span>
            </div>

            <div class="pager">
              <button type="button" class="secondary" @click="goPrevPage">上一页</button>
              <span>{{ pageInfo }}</span>
              <button type="button" class="secondary" @click="goNextPage">下一页</button>
            </div>
          </div>

          <div class="question-list">
            <p v-if="!practiceStore.bank.items.length" class="empty">
              当前筛选条件下暂无题目。
            </p>

            <article
              v-for="item in practiceStore.bank.items"
              :key="item.id"
              class="question-row"
            >
              <div class="row-main">
                <h4>{{ item.stem }}</h4>
                <div class="badges">
                  <span>{{ item.level }}</span>
                  <span>{{ item.question_type }}</span>
                  <span>{{ item.specialty }}</span>
                  <span>{{ item.company_type }}</span>
                  <span>考点：{{ item.points }}</span>
                  <span>难度 {{ item.difficulty_score }}</span>
                </div>
              </div>

              <div class="row-actions">
                <span class="status-tag" :class="item.status">
                  {{ statusLabel(item.status) }}
                </span>
                <button
                  type="button"
                  class="mini secondary practice-from-bank"
                  @click="openQuestionFromBank(item.id)"
                >
                  去做题
                </button>
              </div>
            </article>
          </div>
        </div>

        <div class="tab-panel" :class="{ active: practiceStore.bankTab === 'lists' }">
          <div class="list-cards">
            <p v-if="!practiceStore.bank.questionLists.length" class="empty">
              当前岗位暂无题单。
            </p>

            <article
              v-for="item in practiceStore.bank.questionLists"
              :key="item.id"
              class="list-card"
            >
              <h4>{{ item.title }}</h4>
              <p>{{ item.description }}</p>
              <div class="badges">
                <span v-for="tag in item.tags || []" :key="tag">{{ tag }}</span>
              </div>
              <div class="list-progress">
                <div class="list-progress-bar">
                  <span :style="{ width: `${practiceStore.clampPercent(item.progress)}%` }"></span>
                </div>
                <small>
                  完成 {{ item.solved_count }}/{{ item.total_count }}
                  ({{ practiceStore.clampPercent(item.progress) }}%)
                </small>
              </div>
              <button type="button" class="secondary use-list" @click="practiceStore.useQuestionList(item.id)">
                进入题单
              </button>
            </article>
          </div>
        </div>
      </section>
    </div>
  </section>
</template>

<style scoped>
.view {
  display: none;
  animation: reveal 0.24s ease;
}

.view.active {
  display: block;
}

.bank-layout {
  display: grid;
  grid-template-columns: 290px 1fr;
  gap: 12px;
}

.bank-filter {
  background: #fff;
  border: 1px solid var(--line);
  border-radius: var(--radius);
  padding: 12px;
  display: grid;
  gap: 10px;
  height: fit-content;
}

.bank-filter h3 {
  margin: 0;
  color: #173c60;
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

.checkbox-line {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #3e5d7d;
}

.checkbox-line input {
  width: auto;
}

.assessment-box {
  border: 1px dashed #c0d6f2;
  border-radius: 10px;
  padding: 10px;
  background: #f6faff;
}

.assessment-box h4 {
  margin: 0 0 6px;
  color: #1d4367;
}

.assessment-box p {
  margin: 0 0 8px;
  font-size: 12px;
  color: #4b6885;
}

.bank-main {
  background: #fff;
  border: 1px solid var(--line);
  border-radius: var(--radius);
  padding: 12px;
}

.bank-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 10px;
}

.tab-btn {
  background: #eaf7ff;
  color: #145072;
}

.tab-btn.active {
  background: var(--brand);
  color: #fff;
}

.tab-panel {
  display: none;
}

.tab-panel.active {
  display: block;
}

.pool-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.status-pills {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.pill {
  background: #f1f7ff;
  border: 1px solid #d4e5fb;
  border-radius: 999px;
  padding: 4px 10px;
  font-size: 12px;
  color: #294f73;
}

.pager {
  display: flex;
  align-items: center;
  gap: 8px;
}

.question-list {
  display: grid;
  gap: 8px;
}

.question-row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 10px;
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 10px;
  background: #fcfdff;
}

.question-row h4 {
  margin: 0 0 8px;
  font-size: 14px;
  color: #173b5f;
}

.row-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-tag {
  padding: 4px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
}

.status-tag.todo {
  background: #edf2f8;
  color: #50677f;
}

.status-tag.solved {
  background: #e8fbf1;
  color: #1a7750;
}

.status-tag.wrong {
  background: #fff1f1;
  color: #9a3a3a;
}

.badges {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.badges span {
  background: #f1faff;
  border: 1px solid #d7eaf6;
  border-radius: 999px;
  padding: 3px 8px;
  font-size: 12px;
}

.mini {
  padding: 7px 12px;
  font-size: 12px;
}

.list-cards {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.list-card {
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 12px;
  background: #fcfdff;
  display: grid;
  gap: 8px;
}

.list-card h4 {
  margin: 0;
  color: #173b5f;
}

.list-card p {
  margin: 0;
  font-size: 13px;
  color: #4a6683;
}

.list-progress {
  display: grid;
  gap: 6px;
}

.list-progress-bar {
  height: 8px;
  background: #edf3fc;
  border-radius: 999px;
  overflow: hidden;
}

.list-progress-bar span {
  display: block;
  height: 100%;
  background: linear-gradient(90deg, #2f75b4 0%, #1f507f 100%);
}

.list-progress small {
  color: #4a6683;
}

.empty {
  text-align: center;
  color: #607c9b;
  border: 1px dashed #bfd3ec;
  border-radius: 10px;
  padding: 26px;
}

@keyframes reveal {
  from {
    opacity: 0;
    transform: translateY(4px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 1180px) {
  .bank-layout {
    grid-template-columns: 1fr;
  }

  .list-cards {
    grid-template-columns: 1fr;
  }
}
</style>
