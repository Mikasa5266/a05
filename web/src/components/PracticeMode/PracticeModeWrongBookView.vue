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

const statusLabel = (status) => {
  if (status === "wrong") return "待复习";
  if (status === "solved") return "已通过";
  return "未做";
};

const refreshWrongBook = async () => {
  await practiceStore.loadWrongs();
};

const openQuestion = async (questionId) => {
  if (!questionId) return;
  await practiceStore.loadQuestionById(questionId, "已从错题本进入做题模式");
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
    <div class="wrong-layout">
      <aside class="wrong-filter">
        <h3>错题筛选</h3>

        <div class="row">
          <label>岗位</label>
          <select v-model="practiceStore.wrongFilters.role">
            <option value="all">全部</option>
            <option
              v-for="role in practiceStore.roleOptions"
              :key="role.value"
              :value="role.value"
            >
              {{ role.label }}
            </option>
          </select>
        </div>

        <div class="row">
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

        <div class="row">
          <label>考点</label>
          <input
            v-model="practiceStore.wrongFilters.point"
            placeholder="输入考点关键词"
          />
        </div>

        <label class="checkbox-line">
          <input v-model="practiceStore.wrongFilters.favoriteOnly" type="checkbox" />
          仅看收藏
        </label>

        <button type="button" @click="refreshWrongBook">
          刷新错题
        </button>
      </aside>

      <section class="wrong-main">
        <div class="wrong-list-header">
          <h3>错题本</h3>
          <p>你当前的错题记录会根据筛选条件刷新。</p>
        </div>

        <div v-if="practiceStore.wrongBook.items.length" class="wrong-items">
          <article
            v-for="item in practiceStore.wrongBook.items"
            :key="item.wrong_id"
            class="wrong-card"
          >
            <header class="card-head">
              <div>
                <h4>{{ item.stem }}</h4>
                <div class="badges">
                  <span>{{ item.level }}</span>
                  <span>{{ item.question_type }}</span>
                  <span>{{ item.specialty }}</span>
                  <span>{{ item.company_type }}</span>
                  <span>考点：{{ item.points }}</span>
                </div>
              </div>
              <span class="status-tag wrong">错题</span>
            </header>

            <div class="card-body">
              <p v-if="item.last_user_answer"><strong>你的答案：</strong>{{ item.last_user_answer }}</p>
              <p v-if="item.error_reason"><strong>错误原因：</strong>{{ item.error_reason }}</p>
              <small>最后更新：{{ item.updated_at ? new Date(item.updated_at).toLocaleString() : "未知" }}</small>
            </div>

            <div class="card-actions">
              <button type="button" class="secondary" @click="openQuestion(item.id)">
                去做题
              </button>
              <button type="button" class="secondary" @click="showRemedial(item)">
                补短板
              </button>
              <button type="button" class="secondary" @click="toggleFavorite(item)">
                {{ item.is_favorite ? "取消收藏" : "收藏" }}
              </button>
              <button type="button" class="danger" @click="removeWrong(item)">
                删除
              </button>
            </div>
          </article>
        </div>

        <p v-else class="empty">暂无匹配的错题，换个筛选条件试试。</p>

        <div v-if="practiceStore.wrongBook.remedial" class="remedial-card">
          <h3>考点补短板</h3>
          <p><strong>考点：</strong>{{ practiceStore.wrongBook.remedial.point }}</p>
          <p v-if="practiceStore.wrongBook.remedial.packet">
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
                  {{ question.stem }}
                  <button type="button" class="mini secondary" @click="openQuestion(question.id)">
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
                  {{ question.stem }}
                  <button type="button" class="mini secondary" @click="openQuestion(question.id)">
                    去做
                  </button>
                </li>
              </ul>
            </div>
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

.wrong-layout {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 12px;
}

.wrong-filter,
.wrong-main {
  background: #fff;
  border: 1px solid var(--line);
  border-radius: var(--radius);
  padding: 14px;
}

.wrong-filter h3,
.wrong-list-header h3 {
  margin: 0 0 10px;
  color: #184762;
}

.row {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 10px;
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
  margin-bottom: 12px;
}

button {
  border: none;
  border-radius: 10px;
  background: var(--brand);
  color: #fff;
  padding: 10px 14px;
  cursor: pointer;
  transition: transform 0.15s ease;
}

button:hover {
  transform: translateY(-1px);
}

.wrong-list-header p {
  margin: 4px 0 0;
  color: #4b6885;
  font-size: 13px;
}

.wrong-items {
  display: grid;
  gap: 12px;
}

.wrong-card {
  border: 1px solid var(--line);
  border-radius: 14px;
  padding: 14px;
  display: grid;
  gap: 12px;
  background: #fbfeff;
}

.card-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
}

.card-head h4 {
  margin: 0;
  color: #1f4863;
}

.badges {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.badges span {
  background: #eef7ff;
  border: 1px solid #d7eaf6;
  border-radius: 999px;
  padding: 4px 8px;
  font-size: 12px;
  color: #245975;
}

.status-tag.wrong {
  background: #fff1f1;
  color: #9a2d2d;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
}

.card-body p {
  margin: 0 0 8px;
  font-size: 13px;
  color: #405d72;
}

.card-body small {
  color: #7a8997;
}

.card-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.card-actions .danger {
  background: #cf4c4c;
}

.remedial-card {
  margin-top: 18px;
  border: 1px solid #d1e8f5;
  border-radius: 14px;
  background: #f7fcff;
  padding: 14px;
}

.remedial-card h3 {
  margin: 0 0 10px;
  color: #194b69;
}

.question-group {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.question-group h4 {
  margin: 0 0 8px;
  color: #1d4f6d;
}

.question-group ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 8px;
}

.question-group li {
  background: #fff;
  border: 1px solid #d5e7f2;
  border-radius: 10px;
  padding: 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  color: #2b4f68;
}

.question-group .mini {
  padding: 6px 10px;
  font-size: 12px;
}

.empty {
  color: #66788f;
  padding: 30px;
  text-align: center;
}

@media (max-width: 1040px) {
  .wrong-layout {
    grid-template-columns: 1fr;
  }

  .question-group {
    grid-template-columns: 1fr;
  }
}
</style>
