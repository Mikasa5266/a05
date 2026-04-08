<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import {
  clearQuestionWrong,
  evaluateQuestion,
  generateResumeQuestionList,
  getPositionQuestionList,
  getQuestionBankPositions,
  listFavoriteQuestions,
  listWrongQuestions,
  markQuestionWrong,
  setQuestionFavorite,
} from "../../api/questionBank";
import { useCareerPrepStore } from "../../stores/useCareerPrepStore";

const prepStore = useCareerPrepStore();

const loading = ref(false);
const positions = ref([]);
const questions = ref([]);
const favorites = ref([]);
const wrongQuestions = ref([]);

const answerDraft = reactive({});
const evaluatingMap = reactive({});
const evaluationMap = reactive({});

const difficulty = computed({
  get: () => prepStore.selectedDifficulty,
  set: (val) => prepStore.setSelectedDifficulty(val),
});

const selectedPositionCode = computed({
  get: () => prepStore.selectedPositionCode,
  set: (val) => prepStore.setSelectedPositionCode(val),
});

const canUseResumePush = computed(() => Boolean(prepStore.resumeRecord?.id));

const refreshCollections = async () => {
  const [favRes, wrongRes] = await Promise.all([
    listFavoriteQuestions({ page: 1, page_size: 20 }),
    listWrongQuestions({ page: 1, page_size: 20 }),
  ]);
  favorites.value = Array.isArray(favRes?.questions) ? favRes.questions : [];
  wrongQuestions.value = Array.isArray(wrongRes?.questions)
    ? wrongRes.questions
    : [];
};

const loadPositions = async () => {
  const res = await getQuestionBankPositions();
  positions.value = Array.isArray(res?.positions) ? res.positions : [];

  if (!positions.value.some((x) => x.code === selectedPositionCode.value)) {
    selectedPositionCode.value = positions.value[0]?.code || "backend";
  }
};

const loadPositionQuestions = async () => {
  loading.value = true;
  try {
    const res = await getPositionQuestionList(selectedPositionCode.value, {
      difficulty: difficulty.value,
      limit: 12,
    });
    questions.value = Array.isArray(res?.questions) ? res.questions : [];
    prepStore.setLastQuestionList(questions.value);
  } finally {
    loading.value = false;
  }
};

const loadResumeQuestions = async () => {
  if (!prepStore.resumeRecord?.id) {
    ElMessage.warning("请先在简历解析中心上传简历");
    return;
  }

  loading.value = true;
  try {
    const res = await generateResumeQuestionList(prepStore.resumeRecord.id, {
      difficulty: difficulty.value,
      limit: 12,
    });
    questions.value = Array.isArray(res?.questions) ? res.questions : [];
    prepStore.setLastQuestionList(questions.value);
    ElMessage.success("已按简历完成精准推题");
  } catch (err) {
    ElMessage.error(err?.message || "简历推题失败");
  } finally {
    loading.value = false;
  }
};

const submitEvaluation = async (questionId) => {
  const answer = String(answerDraft[questionId] || "").trim();
  if (!answer) {
    ElMessage.warning("请先输入答案");
    return;
  }

  evaluatingMap[questionId] = true;
  try {
    const res = await evaluateQuestion(questionId, answer);
    evaluationMap[questionId] = res;
    if (Number(res?.score || 0) < 60) {
      await markQuestionWrong(questionId, "自动标记：评分较低");
      await refreshCollections();
    }
    ElMessage.success("评分完成");
  } catch (err) {
    ElMessage.error(err?.message || "评分失败");
  } finally {
    evaluatingMap[questionId] = false;
  }
};

const toggleFavorite = async (questionId, checked) => {
  try {
    await setQuestionFavorite(questionId, checked);
    await refreshCollections();
  } catch (err) {
    ElMessage.error(err?.message || "收藏操作失败");
  }
};

const removeWrong = async (questionId) => {
  try {
    await clearQuestionWrong(questionId);
    await refreshCollections();
  } catch (err) {
    ElMessage.error(err?.message || "移除错题失败");
  }
};

const markWrongManually = async (questionId) => {
  try {
    await markQuestionWrong(questionId, "手动标记");
    await refreshCollections();
    ElMessage.success("已加入错题本");
  } catch (err) {
    ElMessage.error(err?.message || "标记错题失败");
  }
};

const isFavorite = (qid) => favorites.value.some((q) => q.id === qid);

onMounted(async () => {
  loading.value = true;
  try {
    await loadPositions();
    await Promise.all([loadPositionQuestions(), refreshCollections()]);
  } catch (err) {
    ElMessage.error(err?.message || "题库加载失败");
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div class="space-y-6">
    <section class="bg-white border border-zinc-100 rounded-3xl p-6 shadow-sm">
      <h1 class="text-2xl font-bold text-zinc-900">岗位题库中心</h1>
      <p class="text-sm text-zinc-500 mt-1">支持岗位题单、简历精准推题、错题与收藏联动</p>

      <div class="mt-4 flex flex-wrap gap-3 items-center">
        <select
          v-model="selectedPositionCode"
          class="px-3 py-2 border border-zinc-200 rounded-xl text-sm"
          @change="loadPositionQuestions"
        >
          <option v-for="pos in positions" :key="pos.code" :value="pos.code">
            {{ pos.name }}
          </option>
        </select>

        <select
          v-model="difficulty"
          class="px-3 py-2 border border-zinc-200 rounded-xl text-sm"
          @change="loadPositionQuestions"
        >
          <option value="campus_intern">校招实习</option>
          <option value="campus_graduate">校招应届</option>
          <option value="social_junior">社招初级</option>
        </select>

        <button
          type="button"
          class="px-4 py-2 rounded-xl bg-zinc-900 text-white hover:bg-zinc-800 disabled:opacity-60"
          :disabled="loading"
          @click="loadPositionQuestions"
        >
          刷新岗位题单
        </button>

        <button
          type="button"
          class="px-4 py-2 rounded-xl bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-60"
          :disabled="loading || !canUseResumePush"
          @click="loadResumeQuestions"
        >
          按简历精准推题
        </button>
      </div>
    </section>

    <section class="grid grid-cols-1 xl:grid-cols-3 gap-6">
      <article class="xl:col-span-2 bg-white border border-zinc-100 rounded-3xl p-6 shadow-sm">
        <h2 class="text-lg font-semibold text-zinc-900 mb-4">当前题单</h2>
        <p v-if="loading" class="text-sm text-zinc-500">加载中...</p>

        <div v-else class="space-y-4">
          <div v-for="q in questions" :key="q.id" class="border border-zinc-100 rounded-2xl p-4 space-y-3">
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="font-medium text-zinc-900">{{ q.title || '未命名题目' }}</p>
                <p class="text-sm text-zinc-600 mt-1">{{ q.content }}</p>
              </div>
              <label class="inline-flex items-center gap-1 text-xs text-zinc-600 whitespace-nowrap">
                <input
                  type="checkbox"
                  :checked="isFavorite(q.id)"
                  @change="toggleFavorite(q.id, $event.target.checked)"
                />
                收藏
              </label>
            </div>

            <textarea
              v-model="answerDraft[q.id]"
              rows="3"
              class="w-full border border-zinc-200 rounded-xl p-2 text-sm"
              placeholder="输入你的答案后点击评分"
            />

            <div class="flex items-center gap-2">
              <button
                type="button"
                class="px-3 py-1.5 rounded-lg bg-indigo-600 text-white text-sm hover:bg-indigo-700 disabled:opacity-60"
                :disabled="Boolean(evaluatingMap[q.id])"
                @click="submitEvaluation(q.id)"
              >
                {{ evaluatingMap[q.id] ? '评分中...' : '提交评分' }}
              </button>
              <button
                type="button"
                class="px-3 py-1.5 rounded-lg border border-zinc-200 text-sm text-zinc-700 hover:bg-zinc-50"
                @click="markWrongManually(q.id)"
              >
                标记错题
              </button>
            </div>

            <div v-if="evaluationMap[q.id]" class="text-sm bg-zinc-50 border border-zinc-100 rounded-xl p-3">
              <p class="font-medium text-zinc-900">评分：{{ evaluationMap[q.id].score }}</p>
              <pre class="whitespace-pre-wrap text-xs text-zinc-600 mt-1">{{ JSON.stringify(evaluationMap[q.id].feedback, null, 2) }}</pre>
            </div>
          </div>

          <p v-if="!questions.length" class="text-sm text-zinc-500">暂无题目，请切换岗位或重新加载。</p>
        </div>
      </article>

      <aside class="space-y-6">
        <article class="bg-white border border-zinc-100 rounded-3xl p-5 shadow-sm">
          <h3 class="font-semibold text-zinc-900 mb-3">我的收藏</h3>
          <ul class="space-y-2">
            <li v-for="q in favorites" :key="`fav-${q.id}`" class="text-sm text-zinc-700">
              {{ q.title || q.content }}
            </li>
            <li v-if="!favorites.length" class="text-sm text-zinc-500">暂无收藏</li>
          </ul>
        </article>

        <article class="bg-white border border-zinc-100 rounded-3xl p-5 shadow-sm">
          <h3 class="font-semibold text-zinc-900 mb-3">我的错题</h3>
          <ul class="space-y-2">
            <li v-for="q in wrongQuestions" :key="`wrong-${q.id}`" class="text-sm text-zinc-700 flex items-center justify-between gap-2">
              <span class="truncate">{{ q.title || q.content }}</span>
              <button class="text-xs text-indigo-600" @click="removeWrong(q.id)">移除</button>
            </li>
            <li v-if="!wrongQuestions.length" class="text-sm text-zinc-500">暂无错题</li>
          </ul>
        </article>
      </aside>
    </section>
  </div>
</template>
