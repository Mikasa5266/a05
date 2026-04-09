<template>
  <div class="qb-page">
    <section class="qb-hero">
      <div class="qb-hero-main">
        <p class="qb-kicker">Question Bank Studio</p>
        <h1>结构化刷题工作台</h1>
        <p class="qb-copy">题干、代码块、标准答案和详细解析统一走新的 Markdown 安全渲染链路。</p>
        <div class="qb-actions">
          <button class="qb-btn qb-btn-primary" :disabled="drawing || !positionCode" @click="drawRandomQuestion">随机抽题</button>
          <button class="qb-btn" :disabled="startingAssessment || !positionCode" @click="startAssessment">开始岗位测评</button>
          <button class="qb-btn" :disabled="refreshing || !positionCode" @click="loadWorkspace(false)">刷新题库</button>
        </div>
        <div class="qb-stats">
          <article class="qb-stat">
            <span>当前岗位</span>
            <strong>{{ activePositionName }}</strong>
            <small>{{ questionCount }} 道题</small>
          </article>
          <article class="qb-stat">
            <span>当前题单</span>
            <strong>{{ activeCollection?.title || "全部题目" }}</strong>
            <small>{{ activeCollection ? formatPercent(activeCollection.progress) : "未筛选" }}</small>
          </article>
          <article class="qb-stat">
            <span>模式</span>
            <strong>{{ assessment.active ? "岗位测评中" : "自由练习" }}</strong>
            <small>{{ filters.timedMode ? `${filters.secondsPerQuestion}s 计时` : "不限时" }}</small>
          </article>
        </div>
      </div>
      <aside class="qb-context">
        <div class="qb-context-card">
          <span>当前考点</span>
          <strong>{{ activePointInfo?.point || currentPoint || "未定位" }}</strong>
          <small>{{ activePointInfo?.progress ? `完成度 ${formatPercent(activePointInfo.progress.completion)}` : "提交答案后回填掌握度" }}</small>
        </div>
        <div class="qb-context-card">
          <span>练习进度</span>
          <strong>{{ assessment.active ? `${assessment.answeredIds.length}/${assessment.questions.length}` : `${statusStats.solved} 通过` }}</strong>
          <small>{{ assessment.active ? formatPercent(assessmentPercent) : `待复习 ${statusStats.wrong}` }}</small>
        </div>
        <div v-if="filterChips.length" class="qb-chip-row">
          <span v-for="chip in filterChips" :key="chip">{{ chip }}</span>
        </div>
      </aside>
    </section>

    <section v-if="loading" class="qb-skeleton-grid">
      <div v-for="i in 3" :key="i" class="qb-skeleton" />
    </section>

    <section v-else class="qb-grid">
      <aside class="qb-rail">
        <article class="qb-card">
          <div class="qb-card-head">
            <div>
              <p class="qb-kicker">Position</p>
              <h2>岗位切换</h2>
            </div>
            <Target class="h-5 w-5 text-sky-600" />
          </div>
          <div class="qb-stack">
            <button v-for="item in meta.positions" :key="item.code" class="qb-pick" :class="{ 'is-active': item.code === positionCode }" @click="changePosition(item.code)">
              <div>
                <strong>{{ item.name }}</strong>
                <small>{{ formatCount(meta.question_counts?.[item.code]) }} 道题</small>
              </div>
              <ChevronRight class="h-4 w-4 text-slate-400" />
            </button>
          </div>
        </article>

        <article class="qb-card">
          <div class="qb-card-head">
            <div>
              <p class="qb-kicker">Filters</p>
              <h2>练习筛选</h2>
            </div>
            <Filter class="h-5 w-5 text-sky-600" />
          </div>
          <div class="qb-stack">
            <input v-model.trim="filters.keyword" class="qb-control" placeholder="搜索题干关键词" @keydown.enter.prevent="applyFilters" />
            <select v-model="filters.level" class="qb-control">
              <option value="">全部层级</option>
              <option v-for="item in filterOptions.levels" :key="item.key" :value="item.key">{{ item.label }}</option>
            </select>
            <select v-model="filters.questionType" class="qb-control">
              <option value="">全部题型</option>
              <option v-for="item in filterOptions.question_types" :key="item" :value="item">{{ item }}</option>
            </select>
            <select v-model="filters.specialty" class="qb-control">
              <option value="">全部专项</option>
              <option v-for="item in filterOptions.specialties" :key="item" :value="item">{{ item }}</option>
            </select>
            <select v-model="filters.point" class="qb-control">
              <option value="">全部考点</option>
              <option v-for="item in filterOptions.points" :key="item" :value="item">{{ item }}</option>
            </select>
          </div>
        </article>
      </aside>

      <div class="qb-main">
        <article class="qb-card">
          <div class="qb-stack">
            <select v-model="filters.companyType" class="qb-control">
              <option value="">全部企业类型</option>
              <option v-for="item in filterOptions.company_types" :key="item" :value="item">{{ item }}</option>
            </select>
            <label class="qb-check">
              <input v-model="filters.timedMode" type="checkbox" />
              <span>计时模式</span>
            </label>
            <input v-model.number="filters.secondsPerQuestion" class="qb-control" type="number" min="30" max="600" placeholder="每题秒数" />
            <div class="qb-actions">
              <button class="qb-btn qb-btn-primary qb-btn-grow" :disabled="bankLoading" @click="applyFilters">{{ bankLoading ? "更新中..." : "应用筛选" }}</button>
              <button class="qb-btn" @click="resetFilters">重置</button>
            </div>
          </div>
        </article>

        <QuestionPracticeCard
          :question="currentQuestion"
          :position-label="activePositionName"
          :collection-title="activeCollection?.title || ''"
          :selected-single="selectedSingle"
          :selected-multiple="selectedMultiple"
          :text-answer="textAnswer"
          :answer-mode="answerMode"
          :feedback="feedback"
          :solution="solution"
          :timed-mode="filters.timedMode"
          :remaining-seconds="remainingSeconds"
          :submitting="submitting"
          :revealing="revealing"
          :assessment-active="assessment.active"
          @update:selected-single="selectedSingle = $event"
          @update:selected-multiple="selectedMultiple = $event"
          @update:text-answer="textAnswer = $event"
          @update:answer-mode="answerMode = $event"
          @submit="submitAnswer"
          @reveal="showSolution"
          @next="nextQuestion"
        />

        <div v-if="workspaceError" class="qb-error">{{ workspaceError }}</div>
      </div>

      <aside class="qb-rail">
        <article class="qb-card">
          <div class="qb-card-head">
            <div>
              <p class="qb-kicker">Collections</p>
              <h2>岗位题单</h2>
            </div>
            <ListChecks class="h-5 w-5 text-sky-600" />
          </div>
          <div class="qb-stack">
            <button v-for="item in questionLists" :key="item.key" class="qb-list" :class="{ 'is-active': item.key === filters.listKey }" @click="toggleCollection(item.key)">
              <strong>{{ item.title }}</strong>
              <small>{{ item.description }}</small>
              <div class="qb-progress"><span :style="{ width: `${Math.max(6, item.progress)}%` }" /></div>
            </button>
            <div v-if="!questionLists.length" class="qb-empty">当前岗位暂无题单。</div>
          </div>
        </article>

        <article class="qb-card">
          <div class="qb-card-head">
            <div>
              <p class="qb-kicker">Progress</p>
              <h2>做题进度</h2>
            </div>
            <Brain class="h-5 w-5 text-sky-600" />
          </div>
          <div class="qb-stack">
            <div class="qb-progress-card">
              <strong>{{ assessment.active ? "岗位测评进行中" : assessment.summary ? "最近一次测评" : "自由练习" }}</strong>
              <small>{{ assessment.active ? `已完成 ${assessment.answeredIds.length}/${assessment.questions.length}` : assessment.summary ? `得分 ${formatPercent(assessment.summary.score)}` : `已通过 ${statusStats.solved} / 待复习 ${statusStats.wrong}` }}</small>
              <div class="qb-progress"><span :style="{ width: `${Math.max(6, assessment.active ? assessmentPercent : activeCollection?.progress || solvedPercent)}%` }" /></div>
            </div>
            <div v-if="assessment.active && assessment.questions.length" class="qb-step-grid">
              <button v-for="(item, index) in assessment.questions" :key="`${item.id}-${index}`" class="qb-step" :class="stepClass(index, item.id)" @click="jumpAssessment(index)">
                {{ index + 1 }}
              </button>
            </div>
            <div v-if="assessment.summary" class="qb-report">
              <strong>推荐目标企业：{{ assessment.summary.target_company_type || "综合岗位" }}</strong>
              <div v-if="assessment.summary.need_improve_points.length" class="qb-chip-row">
                <span v-for="item in assessment.summary.need_improve_points" :key="item">{{ item }}</span>
              </div>
            </div>
          </div>
        </article>

        <article class="qb-card">
          <div class="qb-card-head">
            <div>
              <p class="qb-kicker">Point Insight</p>
              <h2>考点回路</h2>
            </div>
            <BookOpen class="h-5 w-5 text-sky-600" />
          </div>
          <div v-if="activePointInfo" class="qb-stack">
            <div class="qb-progress-card">
              <strong>{{ activePointInfo.point }}</strong>
              <small>{{ activePointInfo.progress ? `已完成 ${activePointInfo.progress.solved}/${activePointInfo.progress.total}` : "提交答案后会回填掌握进度" }}</small>
              <div v-if="activePointInfo.progress" class="qb-progress"><span :style="{ width: `${Math.max(6, activePointInfo.progress.completion)}%` }" /></div>
            </div>
            <QuestionMarkdown :source="activePointInfo.memo" class="text-sm text-slate-600" empty-text="暂无考点说明" />
            <ul v-if="activePointInfo.interview_extensions.length" class="qb-extension-list">
              <li v-for="item in activePointInfo.interview_extensions" :key="item">{{ item }}</li>
            </ul>
          </div>
          <div v-else class="qb-empty">选中题目后会在这里展示当前考点的说明和掌握进度。</div>
        </article>

        <article class="qb-card">
          <div class="qb-card-head">
            <div>
              <p class="qb-kicker">Question Pool</p>
              <h2>题目预览</h2>
            </div>
            <div class="qb-page-switch">
              <button :disabled="filters.page <= 1" @click="changePage(-1)">上一页</button>
              <span>{{ pagination.page }} / {{ pagination.total_pages }}</span>
              <button :disabled="filters.page >= pagination.total_pages" @click="changePage(1)">下一页</button>
            </div>
          </div>
          <div class="qb-chip-row">
            <span>未做 {{ statusStats.todo }}</span>
            <span>已通过 {{ statusStats.solved }}</span>
            <span>待复习 {{ statusStats.wrong }}</span>
          </div>
          <div class="qb-stack">
            <button v-for="item in questionPool" :key="item.id" class="qb-pool-item" :class="{ 'is-active': item.id === currentQuestion?.id }" :disabled="assessment.active" @click="selectQuestion(item)">
              <strong>{{ item.title || item.stem }}</strong>
              <small>{{ item.points || item.specialty || item.question_type || "综合题目" }}</small>
              <span class="qb-status" :class="statusClass(item.status)">{{ statusLabel(item.status) }}</span>
            </button>
            <div v-if="!questionPool.length" class="qb-empty">当前筛选条件下暂无题目。</div>
          </div>
        </article>
      </aside>
    </section>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import {
  BookOpen,
  Brain,
  ChevronRight,
  Filter,
  ListChecks,
  Play,
  RefreshCcw,
  Sparkles,
  Target,
} from "lucide-vue-next";
import QuestionMarkdown from "../../components/question-bank/QuestionMarkdown.vue";
import QuestionPracticeCard from "../../components/question-bank/QuestionPracticeCard.vue";
import {
  completeQuestionBankAssessment,
  drawQuestionBankQuestion,
  getQuestionBankFilterOptions,
  getQuestionBankLists,
  getQuestionBankMeta,
  getQuestionBankPointSummary,
  getQuestionBankSolution,
  listQuestionBankQuestions,
  startQuestionBankAssessment,
  submitQuestionBankAnswer,
  submitQuestionBankAssessmentAnswer,
} from "../../api/questionBank";

const clean = (value = "") =>
  String(value ?? "")
    .replaceAll("\uFEFF", "")
    .replaceAll("\uFFFD", "")
    .replace(/\r/g, "")
    .trim();

const list = (value) =>
  Array.isArray(value) ? value.map((item) => clean(item)).filter(Boolean) : [];

const normalizeQuestion = (item = {}) => ({
  ...item,
  id: Number(item.id || 0),
  title: clean(item.title),
  stem: clean(item.stem),
  position: clean(item.position),
  level: clean(item.level),
  specialty: clean(item.specialty),
  points: clean(item.points),
  question_type: clean(item.question_type),
  company_type: clean(item.company_type),
  difficulty_score: Number(item.difficulty_score || 0),
  status: clean(item.status || "todo"),
  has_options: Boolean(item.has_options || item.options?.length),
  options: Array.isArray(item.options)
    ? item.options.map((opt) => ({ key: clean(opt?.key), text: clean(opt?.text) }))
    : [],
});

const inferMode = (question) =>
  !question?.has_options
    ? "text"
    : /多选|multiple|多个|多项/i.test(
        `${question.question_type} ${question.title} ${question.stem}`,
      )
      ? "multiple"
      : "single";

const makeAssessment = () => ({
  active: false,
  id: 0,
  questions: [],
  index: 0,
  answeredIds: [],
  summary: null,
});

const loading = ref(true);
const refreshing = ref(false);
const bankLoading = ref(false);
const drawing = ref(false);
const submitting = ref(false);
const revealing = ref(false);
const startingAssessment = ref(false);
const workspaceError = ref("");

const meta = ref({ positions: [], question_counts: {} });
const filterOptions = ref({
  points: [],
  specialties: [],
  levels: [],
  question_types: [],
  company_types: [],
});
const questionLists = ref([]);
const questionPool = ref([]);
const pagination = ref({ page: 1, total_pages: 1 });
const statusStats = ref({ todo: 0, solved: 0, wrong: 0 });
const positionCode = ref("");
const currentQuestion = ref(null);
const feedback = ref(null);
const solution = ref(null);
const pointSummary = ref(null);
const assessment = ref(makeAssessment());

const filters = reactive({
  keyword: "",
  level: "",
  questionType: "",
  specialty: "",
  point: "",
  companyType: "",
  listKey: "",
  page: 1,
  pageSize: 10,
  timedMode: false,
  secondsPerQuestion: 120,
});

const selectedSingle = ref("");
const selectedMultiple = ref([]);
const textAnswer = ref("");
const answerMode = ref("single");
const startedAt = ref(0);
const nowTick = ref(Date.now());
let ticker = null;

const activePositionName = computed(
  () =>
    meta.value.positions.find((item) => item.code === positionCode.value)?.name ||
    "岗位未选择",
);

const activeCollection = computed(
  () => questionLists.value.find((item) => item.key === filters.listKey) || null,
);

const questionCount = computed(
  () => Number(meta.value.question_counts?.[positionCode.value] || 0),
);

const solvedPercent = computed(() => {
  const total =
    statusStats.value.todo + statusStats.value.solved + statusStats.value.wrong;
  return total
    ? Number(((statusStats.value.solved / total) * 100).toFixed(2))
    : 0;
});

const assessmentPercent = computed(() =>
  assessment.value.questions.length
    ? Number(
        (
          (assessment.value.answeredIds.length / assessment.value.questions.length) *
          100
        ).toFixed(2),
      )
    : 0,
);

const currentPoint = computed(() => clean(currentQuestion.value?.points));
const elapsedSeconds = computed(() =>
  startedAt.value ? Math.floor((nowTick.value - startedAt.value) / 1000) : 0,
);

const remainingSeconds = computed(() =>
  filters.timedMode && startedAt.value
    ? Math.max(0, Number(filters.secondsPerQuestion || 0) - elapsedSeconds.value)
    : null,
);

const activePointInfo = computed(() =>
  feedback.value?.point_packet
    ? {
        point: feedback.value.point_packet.point,
        memo: feedback.value.point_packet.memo,
        interview_extensions: feedback.value.point_packet.interview_extensions || [],
        progress: feedback.value.point_state || null,
      }
    : pointSummary.value,
);

const filterChips = computed(() =>
  [
    activeCollection.value?.title,
    filters.level,
    filters.questionType,
    filters.specialty,
    filters.point,
    filters.companyType,
  ].filter(Boolean),
);

const answerValue = computed(() =>
  !currentQuestion.value
    ? ""
    : currentQuestion.value.has_options
      ? answerMode.value === "multiple"
        ? selectedMultiple.value.slice().sort().join(", ")
        : clean(selectedSingle.value)
      : clean(textAnswer.value),
);

const formatPercent = (value) => `${Number(value || 0).toFixed(0)}%`;
const formatCount = (value) => String(Number(value || 0));
const statusLabel = (value) =>
  clean(value) === "solved" ? "已通过" : clean(value) === "wrong" ? "待复习" : "未做";
const statusClass = (value) =>
  clean(value) === "solved"
    ? "qb-status-good"
    : clean(value) === "wrong"
      ? "qb-status-bad"
      : "qb-status-neutral";
const stepClass = (index, id) => ({
  "qb-step-active": id === currentQuestion.value?.id,
  "qb-step-done":
    assessment.value.answeredIds.includes(id) && id !== currentQuestion.value?.id,
  "qb-step-pending":
    !assessment.value.answeredIds.includes(id) &&
    id !== currentQuestion.value?.id,
});

const query = () => ({
  position_code: positionCode.value,
  level: filters.level || undefined,
  question_type: filters.questionType || undefined,
  specialty: filters.specialty || undefined,
  point: filters.point || undefined,
  company_type: filters.companyType || undefined,
  keyword: clean(filters.keyword) || undefined,
  list_key: filters.listKey || undefined,
});

const syncTicker = () => {
  if (!filters.timedMode) {
    if (ticker) {
      window.clearInterval(ticker);
      ticker = null;
    }
    return;
  }
  if (!ticker) {
    ticker = window.setInterval(() => {
      nowTick.value = Date.now();
    }, 1000);
  }
};

const resetComposer = (question) => {
  selectedSingle.value = "";
  selectedMultiple.value = [];
  textAnswer.value = "";
  answerMode.value = inferMode(question);
  feedback.value = null;
  solution.value = null;
  startedAt.value = Date.now();
  nowTick.value = Date.now();
};

const loadMeta = async () => {
  const response = await getQuestionBankMeta();
  meta.value = {
    positions: Array.isArray(response?.positions)
      ? response.positions.map((item) => ({
          code: clean(item.code),
          name: clean(item.name),
        }))
      : [],
    question_counts: response?.question_counts || {},
  };
  if (!positionCode.value) {
    positionCode.value = meta.value.positions[0]?.code || "";
  }
};

const loadOptions = async () => {
  const response = await getQuestionBankFilterOptions(positionCode.value);
  filterOptions.value = {
    points: list(response?.points),
    specialties: list(response?.specialties),
    levels: Array.isArray(response?.levels)
      ? response.levels.map((item) => ({
          key: clean(item.key),
          label: clean(item.label || item.key),
        }))
      : [],
    question_types: list(response?.question_types),
    company_types: list(response?.company_types),
  };
};

const loadLists = async () => {
  const response = await getQuestionBankLists(positionCode.value);
  questionLists.value = Array.isArray(response?.items)
    ? response.items.map((item) => ({
        key: clean(item.key),
        title: clean(item.title),
        description: clean(item.description),
        progress: Number(item.progress || 0),
      }))
    : [];
};

const loadQuestions = async (prime = false) => {
  bankLoading.value = true;
  try {
    const response = await listQuestionBankQuestions({
      ...query(),
      page: filters.page,
      page_size: filters.pageSize,
    });
    questionPool.value = Array.isArray(response?.items)
      ? response.items.map(normalizeQuestion)
      : [];
    pagination.value = {
      page: Number(response?.pagination?.page || 1),
      total_pages: Number(response?.pagination?.total_pages || 1),
    };
    statusStats.value = {
      todo: Number(response?.status_stats?.todo || 0),
      solved: Number(response?.status_stats?.solved || 0),
      wrong: Number(response?.status_stats?.wrong || 0),
    };
    if (prime) {
      await selectQuestion(questionPool.value[0] || null);
    }
  } finally {
    bankLoading.value = false;
  }
};

const loadPoint = async (question) => {
  const point = clean(question?.points);
  if (!point) {
    pointSummary.value = null;
    return;
  }
  try {
    const response = await getQuestionBankPointSummary(point, positionCode.value);
    pointSummary.value = {
      point: clean(response?.point),
      memo: clean(response?.memo),
      interview_extensions: list(response?.interview_extensions),
      progress: response?.progress
        ? {
            total: Number(response.progress.total || 0),
            solved: Number(response.progress.solved || 0),
            completion: Number(response.progress.completion || 0),
          }
        : null,
    };
  } catch {
    pointSummary.value = null;
  }
};

const selectQuestion = async (question) => {
  currentQuestion.value = question ? normalizeQuestion(question) : null;
  if (!currentQuestion.value) {
    feedback.value = null;
    solution.value = null;
    pointSummary.value = null;
    return;
  }
  resetComposer(currentQuestion.value);
  await loadPoint(currentQuestion.value);
};

const loadWorkspace = async (prime = true) => {
  if (!positionCode.value) {
    return;
  }
  refreshing.value = true;
  workspaceError.value = "";
  try {
    await Promise.all([loadOptions(), loadLists()]);
    await loadQuestions(prime);
    if (!prime && currentQuestion.value) {
      await loadPoint(currentQuestion.value);
    }
  } catch (error) {
    workspaceError.value = error?.message || "加载题库失败";
    ElMessage.error(workspaceError.value);
  } finally {
    refreshing.value = false;
  }
};

const resetFilters = async () => {
  if (assessment.value.active) {
    ElMessage.warning("岗位测评进行中，请先完成当前测评。");
    return;
  }
  filters.keyword = "";
  filters.level = "";
  filters.questionType = "";
  filters.specialty = "";
  filters.point = "";
  filters.companyType = "";
  filters.listKey = "";
  filters.page = 1;
  await loadWorkspace(true);
};

const applyFilters = async () => {
  if (assessment.value.active) {
    ElMessage.warning("岗位测评进行中，请先完成当前测评。");
    return;
  }
  filters.page = 1;
  await loadWorkspace(true);
};

const changePosition = async (code) => {
  if (code === positionCode.value) {
    return;
  }
  positionCode.value = clean(code);
  assessment.value = makeAssessment();
  await resetFilters();
};

const toggleCollection = async (key) => {
  if (assessment.value.active) {
    ElMessage.warning("岗位测评进行中，请先完成当前测评。");
    return;
  }
  filters.listKey = filters.listKey === key ? "" : key;
  filters.page = 1;
  await loadWorkspace(true);
};

const changePage = async (offset) => {
  const next = Math.max(1, filters.page + Number(offset || 0));
  if (next === filters.page || next > pagination.value.total_pages) {
    return;
  }
  filters.page = next;
  await loadQuestions(false);
};

const drawRandomQuestion = async () => {
  drawing.value = true;
  try {
    const response = await drawQuestionBankQuestion(query());
    await selectQuestion(response?.question);
    ElMessage.success("已切换到新的题目");
  } catch (error) {
    ElMessage.error(error?.message || "抽题失败");
  } finally {
    drawing.value = false;
  }
};

const submitAnswer = async () => {
  if (!currentQuestion.value) {
    ElMessage.warning("请先选择一道题目。");
    return;
  }
  if (!answerValue.value) {
    ElMessage.warning("请先完成作答。");
    return;
  }
  submitting.value = true;
  try {
    const payload = {
      question_id: currentQuestion.value.id,
      user_answer: answerValue.value,
      elapsed_seconds: filters.timedMode ? elapsedSeconds.value : undefined,
      is_timeout: false,
    };
    const response = assessment.value.active
      ? await submitQuestionBankAssessmentAnswer({
          assessment_id: assessment.value.id,
          ...payload,
        })
      : await submitQuestionBankAnswer({
          ...payload,
          timed_mode: filters.timedMode,
        });

    feedback.value = {
      is_correct: Boolean(response?.is_correct),
      standard_answer: clean(response?.standard_answer),
      analysis: clean(response?.analysis),
      tips: clean(response?.tips),
      exemplar: clean(response?.exemplar),
      matched_keywords: list(response?.matched_keywords),
      missing_keywords: list(response?.missing_keywords),
      point_packet: response?.point_packet
        ? {
            point: clean(response.point_packet.point),
            memo: clean(response.point_packet.memo),
            interview_extensions: list(response.point_packet.interview_extensions),
          }
        : null,
      point_state: response?.point_state
        ? {
            total: Number(response.point_state.total || 0),
            solved: Number(response.point_state.solved || 0),
            completion: Number(response.point_state.completion || 0),
          }
        : null,
    };

    if (
      assessment.value.active &&
      !assessment.value.answeredIds.includes(currentQuestion.value.id)
    ) {
      assessment.value = {
        ...assessment.value,
        answeredIds: [...assessment.value.answeredIds, currentQuestion.value.id],
      };
    }

    await loadQuestions(false);
    if (
      assessment.value.active &&
      assessment.value.answeredIds.length >= assessment.value.questions.length
    ) {
      await completeAssessment();
    }
    ElMessage.success(feedback.value.is_correct ? "回答已通过" : "解析已生成，可继续复盘");
  } catch (error) {
    ElMessage.error(error?.message || "提交答案失败");
  } finally {
    submitting.value = false;
  }
};

const showSolution = async () => {
  if (!currentQuestion.value) {
    ElMessage.warning("请先选择一道题目。");
    return;
  }
  revealing.value = true;
  try {
    const response = await getQuestionBankSolution(currentQuestion.value.id);
    solution.value = {
      standard_answer: clean(response?.standard_answer),
      analysis: clean(response?.analysis),
      tips: clean(response?.tips),
      exemplar: clean(response?.exemplar),
    };
  } catch (error) {
    ElMessage.error(error?.message || "加载解析失败");
  } finally {
    revealing.value = false;
  }
};

const startAssessment = async () => {
  startingAssessment.value = true;
  try {
    const response = await startQuestionBankAssessment({
      position_code: positionCode.value,
      total_count: 12,
    });
    assessment.value = {
      active: true,
      id: Number(response?.assessment_id || 0),
      questions: Array.isArray(response?.questions)
        ? response.questions.map(normalizeQuestion)
        : [],
      index: 0,
      answeredIds: [],
      summary: null,
    };
    await selectQuestion(assessment.value.questions[0] || null);
    ElMessage.success("岗位测评已开始");
  } catch (error) {
    ElMessage.error(error?.message || "启动测评失败");
  } finally {
    startingAssessment.value = false;
  }
};

const completeAssessment = async () => {
  const response = await completeQuestionBankAssessment(assessment.value.id);
  assessment.value = {
    ...assessment.value,
    active: false,
    summary: {
      score: Number(response?.score || 0),
      target_company_type: clean(response?.target_company_type),
      need_improve_points: list(response?.need_improve_points),
    },
  };
};

const jumpAssessment = async (index) => {
  const next = assessment.value.questions[index];
  if (!next) {
    return;
  }
  assessment.value = { ...assessment.value, index };
  await selectQuestion(next);
};

const nextQuestion = async () => {
  if (assessment.value.active) {
    const next = assessment.value.questions[assessment.value.index + 1];
    if (!next) {
      return completeAssessment();
    }
    assessment.value = {
      ...assessment.value,
      index: assessment.value.index + 1,
    };
    return selectQuestion(next);
  }
  return drawRandomQuestion();
};

watch(() => filters.timedMode, () => syncTicker());
watch(() => filters.secondsPerQuestion, () => {
  if (filters.timedMode && currentQuestion.value) {
    startedAt.value = Date.now();
  }
});

onMounted(async () => {
  try {
    await loadMeta();
    await loadWorkspace(true);
  } finally {
    loading.value = false;
  }
});

onBeforeUnmount(() => {
  if (ticker) {
    window.clearInterval(ticker);
  }
});
</script>

<style scoped>
.qb-page { display: grid; gap: 1.5rem; }
.qb-hero, .qb-card { border: 1px solid #e2e8f0; border-radius: 2rem; background: #fff; }
.qb-hero { display: grid; gap: 1.5rem; padding: 1.5rem; background: radial-gradient(circle at top left, rgba(14, 165, 233, 0.16), transparent 28%), linear-gradient(180deg, #f8fbff 0%, #eef5ff 100%); }
.qb-hero-main, .qb-rail, .qb-main, .qb-stack { display: grid; gap: 1rem; }
.qb-kicker { margin: 0; font-size: 0.72rem; font-weight: 700; letter-spacing: 0.28em; text-transform: uppercase; color: #94a3b8; }
.qb-hero h1, .qb-card h2 { margin: 0; color: #0f172a; }
.qb-hero h1 { font-size: clamp(2rem, 4vw, 3.4rem); line-height: 1.06; }
.qb-copy, .qb-empty, .qb-error { color: #475569; font-size: 0.95rem; line-height: 1.8; }
.qb-actions, .qb-inline, .qb-chip-row, .qb-page-switch { display: flex; flex-wrap: wrap; gap: 0.75rem; align-items: center; }
.qb-btn { display: inline-flex; align-items: center; justify-content: center; gap: 0.5rem; border: 1px solid #e2e8f0; border-radius: 1rem; background: #fff; padding: 0.85rem 1rem; font-size: 0.9rem; font-weight: 600; color: #334155; transition: 0.2s ease; }
.qb-btn:hover, .qb-page-switch button:hover { background: #f8fafc; border-color: #cbd5e1; }
.qb-btn:disabled, .qb-page-switch button:disabled, .qb-pool-item:disabled { opacity: 0.55; cursor: not-allowed; }
.qb-btn-primary { background: #0f172a; border-color: #0f172a; color: #fff; }
.qb-btn-primary:hover { background: #1e293b; }
.qb-btn-grow { flex: 1 1 auto; }
.qb-stats, .qb-grid, .qb-skeleton-grid, .qb-step-grid { display: grid; gap: 1rem; }
.qb-stats { grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr)); }
.qb-stat, .qb-context-card, .qb-progress-card, .qb-report { border-radius: 1.5rem; border: 1px solid rgba(226, 232, 240, 0.9); background: rgba(255, 255, 255, 0.86); padding: 1rem; }
.qb-stat span, .qb-context-card span { font-size: 0.72rem; font-weight: 700; letter-spacing: 0.18em; text-transform: uppercase; color: #94a3b8; }
.qb-stat strong, .qb-context-card strong { display: block; margin-top: 0.5rem; color: #0f172a; font-size: 1.12rem; }
.qb-stat small, .qb-context-card small, .qb-pick small, .qb-list small, .qb-pool-item small, .qb-progress-card small { display: block; margin-top: 0.25rem; color: #64748b; font-size: 0.82rem; line-height: 1.6; }
.qb-context, .qb-skeleton-grid { display: grid; gap: 0.85rem; }
.qb-chip-row span, .qb-status { border-radius: 999px; padding: 0.3rem 0.7rem; font-size: 0.72rem; font-weight: 700; }
.qb-chip-row span { background: #fff; border: 1px solid #e2e8f0; color: #475569; }
.qb-card { padding: 1.15rem; }
.qb-card-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 0.75rem; margin-bottom: 1rem; }
.qb-pick, .qb-list, .qb-pool-item { width: 100%; border: 1px solid #e2e8f0; border-radius: 1.35rem; background: #fff; padding: 0.95rem 1rem; text-align: left; transition: 0.2s ease; position: relative; }
.qb-pick:hover, .qb-list:hover, .qb-pool-item:hover { border-color: #cbd5e1; background: #f8fafc; }
.qb-pick.is-active, .qb-list.is-active, .qb-pool-item.is-active { border-color: #7dd3fc; background: #f0f9ff; box-shadow: 0 22px 60px -48px rgba(14, 165, 233, 0.95); }
.qb-control { width: 100%; border: 1px solid #e2e8f0; border-radius: 1rem; background: #f8fafc; padding: 0.9rem 1rem; font-size: 0.9rem; color: #334155; outline: none; transition: 0.2s ease; }
.qb-control:focus { border-color: #7dd3fc; background: #fff; box-shadow: 0 0 0 4px rgba(125, 211, 252, 0.24); }
.qb-check { display: flex; align-items: center; gap: 0.7rem; border: 1px solid #e2e8f0; border-radius: 1rem; background: #f8fafc; padding: 0.9rem 1rem; color: #334155; font-size: 0.9rem; }
.qb-progress { height: 0.5rem; overflow: hidden; border-radius: 999px; background: #e2e8f0; margin-top: 0.75rem; }
.qb-progress span { display: block; height: 100%; border-radius: 999px; background: linear-gradient(90deg, #0f172a 0%, #0284c7 55%, #22d3ee 100%); }
.qb-step-grid { grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 0.5rem; }
.qb-step { height: 2.5rem; border: 1px solid #e2e8f0; border-radius: 1rem; background: #fff; color: #475569; font-size: 0.78rem; font-weight: 700; }
.qb-step-active { border-color: #7dd3fc; background: #f0f9ff; color: #0369a1; }
.qb-step-done { border-color: #bbf7d0; background: #f0fdf4; color: #15803d; }
.qb-step-pending { background: #fff; }
.qb-report { background: #f0fdf4; border-color: #bbf7d0; }
.qb-page-switch button { border: 1px solid #e2e8f0; border-radius: 999px; background: #fff; padding: 0.2rem 0.6rem; font-size: 0.72rem; font-weight: 700; color: #475569; }
.qb-status { position: absolute; top: 0.9rem; right: 0.9rem; }
.qb-status-neutral { background: #f1f5f9; color: #475569; }
.qb-status-good { background: #dcfce7; color: #15803d; }
.qb-status-bad { background: #ffe4e6; color: #be123c; }
.qb-extension-list { display: grid; gap: 0.6rem; margin: 0; padding: 0; list-style: none; }
.qb-extension-list li { border-radius: 1rem; background: #f8fafc; padding: 0.8rem 0.9rem; color: #475569; font-size: 0.85rem; line-height: 1.7; }
.qb-empty { border: 1px dashed #cbd5e1; border-radius: 1.25rem; background: #f8fafc; padding: 1rem; }
.qb-error { border: 1px solid #fecdd3; border-radius: 1.5rem; background: #fff1f2; padding: 1rem 1.1rem; color: #be123c; }
.qb-skeleton { height: 26rem; border-radius: 2rem; background: linear-gradient(90deg, #f8fafc 25%, #eef2f7 37%, #f8fafc 63%); background-size: 400% 100%; animation: pulse 1.4s ease infinite; }
@keyframes pulse { 0% { background-position: 100% 50%; } 100% { background-position: 0 50%; } }
@media (min-width: 1280px) { .qb-grid { grid-template-columns: 320px minmax(0, 1fr) 320px; } .qb-hero { grid-template-columns: minmax(0, 1.15fr) 340px; } }
</style>
