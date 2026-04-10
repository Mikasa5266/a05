<script setup>
import { computed } from "vue";
import { useRouter } from "vue-router";

import { usePracticeStore } from "../../stores/practice";

defineProps({
  active: {
    type: Boolean,
    default: false,
  },
});

const practiceStore = usePracticeStore();
const router = useRouter();

const wrongItems = computed(() => practiceStore.wrongBook.items || []);

const normalizeLevel = (level) => {
  const value = String(level || "").toLowerCase();
  if (!value) return "未知难度";
  if (
    value.includes("hard") ||
    value.includes("困难") ||
    value.includes("高") ||
    value.includes("advanced")
  ) {
    return "高难";
  }
  if (
    value.includes("medium") ||
    value.includes("中") ||
    value.includes("normal")
  ) {
    return "中等";
  }
  if (
    value.includes("easy") ||
    value.includes("低") ||
    value.includes("基础")
  ) {
    return "基础";
  }
  return String(level);
};

const levelClass = (level) => {
  const normalized = normalizeLevel(level);
  if (normalized === "高难") return "hard";
  if (normalized === "中等") return "medium";
  return "easy";
};

const statusLabel = (status) => {
  if (status === "solved") return "已掌握";
  if (status === "wrong") return "待复习";
  return "待练习";
};

const statusClass = (status) => {
  if (status === "solved") return "solved";
  if (status === "wrong") return "wrong";
  return "pending";
};

const formatDate = (value) => {
  if (!value) return "未知";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "未知";
  return date.toLocaleString();
};

const summaryCards = computed(() => {
  const total = wrongItems.value.length;
  const favorite = wrongItems.value.filter((item) => item.is_favorite).length;
  const solved = wrongItems.value.filter((item) => item.status === "solved").length;
  const hard = wrongItems.value.filter(
    (item) => normalizeLevel(item.level) === "高难",
  ).length;

  return [
    {
      key: "total",
      title: "错题总数",
      value: total,
      tip: "当前筛选结果",
    },
    {
      key: "hard",
      title: "高难错题",
      value: hard,
      tip: "建议优先复习",
    },
    {
      key: "favorite",
      title: "重点收藏",
      value: favorite,
      tip: "用于二轮强化",
    },
    {
      key: "solved",
      title: "已掌握",
      value: solved,
      tip: "可降低复习频率",
    },
  ];
});

const refreshWrongBook = async () => {
  await practiceStore.loadWrongs();
};

const resetFilters = async () => {
  practiceStore.wrongFilters.role = "all";
  practiceStore.wrongFilters.questionType = "all";
  practiceStore.wrongFilters.point = "";
  practiceStore.wrongFilters.favoriteOnly = false;
  await refreshWrongBook();
};

const resolveQuestionId = (payload) => {
  if (typeof payload === "number" || typeof payload === "string") {
    return Number(payload);
  }
  return Number(payload?.id || payload?.question_id || 0);
};

const openQuestion = async (payload) => {
  const questionId = resolveQuestionId(payload);
  if (!questionId) return;

  if (router.currentRoute.value.name !== "StudentPracticeMode") {
    await router.push({ name: "StudentPracticeMode" });
  }

  await practiceStore.loadQuestionById(questionId, "已从错题本进入刷题模式");
};

const showRemedial = async (wrongItem) => {
  if (!wrongItem?.wrong_id) return;
  await practiceStore.loadWrongRemedial(wrongItem.wrong_id);
};

const removeWrong = async (wrongItem) => {
  if (!wrongItem?.wrong_id) return;
  await practiceStore.removeWrong(wrongItem.wrong_id);
};

const toggleFavorite = async (wrongItem) => {
  if (!wrongItem?.wrong_id) return;
  await practiceStore.toggleWrongFavorite(wrongItem.wrong_id);
};
</script>

<template>
  <section class="view" :class="{ active }">
    <div class="wrong-book-page">
      <header class="page-header">
        <div>
          <h3>错题本</h3>
          <p>围绕错因集中复习，优先处理高难与高频考点题。</p>
        </div>
        <button type="button" class="btn ghost" @click="refreshWrongBook">
          {{ practiceStore.wrongBook.loading ? "刷新中..." : "刷新数据" }}
        </button>
      </header>

      <section class="filter-panel">
        <div class="field">
          <label>岗位</label>
          <select v-model="practiceStore.wrongFilters.role">
            <option value="all">全部</option>
            <option
              v-for="role in practiceStore.wrongRoleOptions"
              :key="role.value"
              :value="role.value"
            >
              {{ role.label }}
            </option>
          </select>
        </div>

        <div class="field">
          <label>题型</label>
          <select v-model="practiceStore.wrongFilters.questionType">
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

        <div class="field keyword">
          <label>考点检索</label>
          <input
            v-model="practiceStore.wrongFilters.point"
            placeholder="输入关键词，例如：并发控制"
          />
        </div>

        <label class="favorite-toggle">
          <input v-model="practiceStore.wrongFilters.favoriteOnly" type="checkbox" />
          仅看收藏
        </label>

        <div class="filter-actions">
          <button type="button" class="btn secondary" @click="resetFilters">
            重置
          </button>
          <button type="button" class="btn primary" @click="refreshWrongBook">
            应用筛选
          </button>
        </div>
      </section>

      <section class="summary-grid">
        <article v-for="card in summaryCards" :key="card.key" class="summary-card">
          <p class="card-title">{{ card.title }}</p>
          <strong class="card-value">{{ card.value }}</strong>
          <span class="card-tip">{{ card.tip }}</span>
        </article>
      </section>

      <section class="wrong-list-wrap">
        <p v-if="practiceStore.wrongBook.loading" class="state-text">正在加载错题...</p>

        <template v-else-if="wrongItems.length">
          <article v-for="item in wrongItems" :key="item.wrong_id" class="wrong-card">
            <header class="card-header">
              <div class="title-block">
                <h4>{{ item.stem || "未命名题目" }}</h4>
                <div class="meta-tags">
                  <span class="tag">{{ item.question_type || "题型未知" }}</span>
                  <span class="tag">{{ item.specialty || "方向未标注" }}</span>
                  <span class="tag">{{ item.company_type || "场景未标注" }}</span>
                  <span class="tag">考点：{{ item.points || "未标注" }}</span>
                </div>
              </div>

              <div class="side-tags">
                <span class="level-tag" :class="levelClass(item.level)">
                  {{ normalizeLevel(item.level) }}
                </span>
                <span class="status-tag" :class="statusClass(item.status)">
                  {{ statusLabel(item.status) }}
                </span>
              </div>
            </header>

            <div class="card-body">
              <p v-if="item.error_reason" class="error-reason">
                <strong>错误原因：</strong>{{ item.error_reason }}
              </p>
              <p v-if="item.last_user_answer" class="answer-preview">
                <strong>你的答案：</strong>{{ item.last_user_answer }}
              </p>
            </div>

            <footer class="card-footer">
              <small>最近练习：{{ formatDate(item.updated_at) }}</small>

              <div class="card-actions">
                <button type="button" class="btn primary" @click="openQuestion(item)">
                  去复习
                </button>
                <button type="button" class="btn secondary" @click="showRemedial(item)">
                  补短板
                </button>
                <button type="button" class="btn ghost" @click="toggleFavorite(item)">
                  {{ item.is_favorite ? "取消收藏" : "收藏" }}
                </button>
                <button type="button" class="btn danger" @click="removeWrong(item)">
                  删除
                </button>
              </div>
            </footer>
          </article>
        </template>

        <p v-else class="state-text empty">暂无匹配的错题，换个筛选条件试试。</p>
      </section>

      <section v-if="practiceStore.wrongBook.remedial" class="remedial-card">
        <header>
          <h3>考点补短板</h3>
          <p>针对当前错题自动推荐基础题与进阶题，形成闭环训练。</p>
        </header>
        <p class="point-line">
          <strong>核心考点：</strong>{{ practiceStore.wrongBook.remedial.point || "未标注" }}
        </p>
        <p v-if="practiceStore.wrongBook.remedial.packet" class="memo-line">
          <strong>知识速记：</strong>{{ practiceStore.wrongBook.remedial.packet.memo }}
        </p>
        <div class="question-group">
          <div>
            <h4>基础题</h4>
            <ul>
              <li
                v-for="question in practiceStore.wrongBook.remedial.base_questions || []"
                :key="question.id"
              >
                <span>{{ question.stem }}</span>
                <button type="button" class="btn mini" @click="openQuestion(question)">
                  去做
                </button>
              </li>
            </ul>
          </div>
          <div>
            <h4>进阶题</h4>
            <ul>
              <li
                v-for="question in practiceStore.wrongBook.remedial.advanced_questions || []"
                :key="question.id"
              >
                <span>{{ question.stem }}</span>
                <button type="button" class="btn mini" @click="openQuestion(question)">
                  去做
                </button>
              </li>
            </ul>
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

.wrong-book-page {
  display: grid;
  gap: 14px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  align-items: center;
}

.page-header h3 {
  margin: 0;
  color: #16455f;
  font-size: 20px;
}

.page-header p {
  margin: 6px 0 0;
  color: #55718a;
  font-size: 13px;
}

.filter-panel {
  background: linear-gradient(145deg, #f9fcff 0%, #eef7ff 100%);
  border: 1px solid #d7e8f6;
  border-radius: 16px;
  padding: 14px;
  display: grid;
  grid-template-columns: repeat(2, minmax(180px, 1fr)) minmax(220px, 1.4fr) auto auto;
  gap: 10px;
  align-items: end;
}

.field {
  display: grid;
  gap: 6px;
}

.field label {
  font-size: 12px;
  color: #4a6781;
  font-weight: 700;
}

.favorite-toggle {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #fff;
  border: 1px solid #dbe9f5;
  border-radius: 10px;
  padding: 10px 12px;
  color: #365670;
  font-size: 13px;
}

.favorite-toggle input {
  width: 16px;
  height: 16px;
}

.filter-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.summary-card {
  background: #ffffff;
  border: 1px solid #deebf5;
  border-radius: 14px;
  padding: 12px 14px;
  display: grid;
  gap: 6px;
  box-shadow: 0 10px 24px rgba(18, 73, 103, 0.08);
}

.card-title {
  margin: 0;
  color: #537089;
  font-size: 12px;
}

.card-value {
  color: #123851;
  font-size: 30px;
  line-height: 1;
}

.card-tip {
  color: #6a8399;
  font-size: 12px;
}

.wrong-list-wrap {
  border: 1px solid #d8e8f2;
  border-radius: 16px;
  background: #ffffff;
  padding: 14px;
  display: grid;
  gap: 12px;
}

.wrong-card {
  border: 1px solid #d8e8f2;
  border-radius: 14px;
  padding: 14px;
  display: grid;
  gap: 12px;
  background: linear-gradient(180deg, #ffffff 0%, #fafdff 100%);
}

.card-header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.title-block {
  min-width: 0;
}

.title-block h4 {
  margin: 0;
  color: #173d57;
  line-height: 1.45;
}

.meta-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  margin-top: 10px;
}

.tag {
  border-radius: 999px;
  padding: 4px 10px;
  border: 1px solid #d7e7f4;
  background: #f2f8fd;
  color: #2a5877;
  font-size: 12px;
  font-weight: 600;
}

.side-tags {
  display: grid;
  gap: 7px;
  justify-items: end;
}

.level-tag,
.status-tag {
  border-radius: 999px;
  padding: 5px 12px;
  font-size: 12px;
  font-weight: 700;
}

.level-tag.easy {
  background: #e7f8f1;
  color: #1f7d58;
}

.level-tag.medium {
  background: #fff6e4;
  color: #936315;
}

.level-tag.hard {
  background: #ffeaea;
  color: #9d2d2d;
}

.status-tag.pending {
  background: #eef5ff;
  color: #1d5f95;
}

.status-tag.solved {
  background: #e6f7ef;
  color: #1f7d58;
}

.status-tag.wrong {
  background: #ffecef;
  color: #ab2f43;
}

.card-body {
  color: #35556e;
  font-size: 13px;
  line-height: 1.65;
}

.card-body p {
  margin: 0;
}

.card-body p + p {
  margin-top: 8px;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}

.card-footer small {
  color: #6f889d;
}

.card-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.btn {
  border-radius: 10px;
  padding: 8px 12px;
  min-width: 74px;
  font-size: 13px;
  font-weight: 700;
}

.btn.primary {
  background: linear-gradient(135deg, #1f7ead 0%, #2e95cb 100%);
}

.btn.secondary {
  background: linear-gradient(135deg, #3d8fbe 0%, #54a8d5 100%);
}

.btn.ghost {
  background: #e9f4fb;
  color: #285779;
  box-shadow: none;
}

.btn.danger {
  background: #cf4c4c;
}

.btn.mini {
  padding: 6px 10px;
  min-width: auto;
  font-size: 12px;
  background: linear-gradient(135deg, #1f7ead 0%, #2e95cb 100%);
}

.state-text {
  margin: 0;
  color: #5e7a90;
  padding: 20px;
  text-align: center;
}

.state-text.empty {
  border: 1px dashed #d5e5f2;
  border-radius: 12px;
  background: #fbfdff;
}

.remedial-card {
  border: 1px solid #d1e5f4;
  border-radius: 16px;
  background: linear-gradient(145deg, #fafdff 0%, #f1f8ff 100%);
  padding: 14px;
  display: grid;
  gap: 10px;
}

.remedial-card h3 {
  margin: 0;
  color: #174863;
}

.remedial-card header p,
.point-line,
.memo-line {
  margin: 0;
  color: #3e617b;
  font-size: 13px;
  line-height: 1.6;
}

.question-group {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.question-group h4 {
  margin: 0 0 8px;
  color: #174863;
}

.question-group ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 8px;
}

.question-group li {
  border: 1px solid #d8e8f3;
  background: #ffffff;
  border-radius: 10px;
  padding: 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  color: #2f546d;
  font-size: 13px;
}

@media (max-width: 1200px) {
  .filter-panel {
    grid-template-columns: 1fr 1fr;
  }

  .field.keyword {
    grid-column: 1 / -1;
  }

  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 880px) {
  .page-header,
  .card-header,
  .card-footer {
    flex-direction: column;
    align-items: flex-start;
  }

  .side-tags,
  .card-actions {
    justify-items: start;
    justify-content: flex-start;
  }

  .summary-grid,
  .question-group,
  .filter-panel {
    grid-template-columns: 1fr;
  }

  .filter-actions {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>
