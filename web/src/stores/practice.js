import { defineStore } from "pinia";
import { computed, reactive, ref } from "vue";

import {
  completePracticeAssessment,
  deletePracticeWrong,
  drawPracticeQuestion,
  exportPracticeRecords,
  getPracticeDashboard,
  getPracticeFilterOptions,
  getPracticeIntegrationSnapshot,
  getPracticeMeta,
  getPracticeQuestion,
  getPracticeQuestionLists,
  getPracticeSolution,
  getPracticeSpecialties,
  getPracticeWrongRemedial,
  getPracticeWrongs,
  importPracticeQuestionBank,
  listPracticeQuestions,
  startPracticeAssessment,
  submitPracticeAnswer,
  submitPracticeAssessmentAnswer,
  togglePracticeQuestionFavorite,
  togglePracticeWrongFavorite,
} from "../api/practice";

const DEFAULT_ROLE = "ai_engineer";
const DEFAULT_STATUS_TEXT = "准备就绪";
const DEFAULT_COACH_HINT =
  "先开始一题，我会在每个关键节点给你节奏和答题结构建议。";
const DEFAULT_COACH_CHECKLIST = ["审题", "结构", "量化"];
const DEFAULT_LOOP_HINTS = ["结构化回答", "先结论再展开", "举例验证方案"];
const VIEW_TITLE_MAP = {
  bank: "题库与题单",
  practice: "刷题模式",
  wrong: "错题本",
  dashboard: "数据统计",
  resume: "简历解析",
};

const buildEmptyPagination = () => ({
  page: 1,
  page_size: 15,
  total: 0,
  total_pages: 1,
});

const buildEmptyStatusStats = () => ({
  todo: 0,
  solved: 0,
  wrong: 0,
});

const buildDefaultPointLoop = () => ({
  type: "default",
  hints: [...DEFAULT_LOOP_HINTS],
  packet: null,
  pointState: null,
  assessment: null,
});

const clampPercent = (value) => {
  const numeric = Number(value || 0);
  if (!Number.isFinite(numeric)) return 0;
  return Math.max(0, Math.min(100, Math.round(numeric)));
};

const normalizeFilterValue = (value) => {
  const text = String(value ?? "").trim();
  if (!text || text === "all") return "";
  return text;
};

const normalizeQuestionOptions = (options) => {
  if (!Array.isArray(options)) return [];
  return options
    .map((item) => ({
      key: String(item?.key || "").trim(),
      text: String(item?.text || "").trim(),
    }))
    .filter((item) => item.key && item.text);
};

const normalizeQuestion = (question, fallbackRole = DEFAULT_ROLE) => {
  if (!question) return null;

  const options = normalizeQuestionOptions(question.options);
  const role = String(
    question.role || question.position_code || fallbackRole || DEFAULT_ROLE,
  ).trim();

  return {
    ...question,
    role,
    position_code: String(question.position_code || role).trim(),
    level: String(question.level || "").trim(),
    question_type: String(question.question_type || "").trim(),
    specialty: String(question.specialty || "").trim(),
    stem: String(question.stem || "").trim(),
    points: String(question.points || question.knowledge_point || "").trim(),
    company_type: String(question.company_type || "").trim(),
    difficulty_score: Number(question.difficulty_score || 0),
    has_options:
      typeof question.has_options === "boolean"
        ? question.has_options
        : options.length > 0,
    is_favorite: Boolean(question.is_favorite),
    options,
  };
};

const triggerBlobDownload = (blob, filename) => {
  if (!blob) return;
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
};

export const usePracticeStore = defineStore("practice", () => {
  const meta = ref(null);
  const booting = ref(false);
  const initialized = ref(false);

  const currentRole = ref(DEFAULT_ROLE);
  const activeView = ref("practice");
  const bankTab = ref("pool");
  const selectedListId = ref("");

  const currentQuestion = ref(null);
  const questionHistory = ref([]);
  const historyIndex = ref(-1);
  const answerInput = ref("");
  const selectedOption = ref("");
  const errorReason = ref("");
  const currentResult = ref(null);
  const pointLoop = ref(buildDefaultPointLoop());

  const status = reactive({
    text: DEFAULT_STATUS_TEXT,
    isAlert: false,
  });

  const coach = reactive({
    mood: "待命",
    hint: DEFAULT_COACH_HINT,
    checklist: [...DEFAULT_COACH_CHECKLIST],
  });

  const questionFilters = reactive({
    mode: "single",
    level: "all",
    questionType: "all",
    specialty: "all",
    companyType: "all",
    timedMode: false,
    perQuestionSeconds: 120,
  });

  const filterOptions = reactive({
    specialties: [],
    points: [],
    companyTypes: [],
    questionTypes: [],
    levels: {},
    statusOptions: [],
  });

  const bankFilters = reactive({
    keyword: "",
    level: "all",
    questionType: "all",
    specialty: "all",
    point: "all",
    companyType: "all",
    status: "all",
    favoriteOnly: false,
    page: 1,
    pageSize: 15,
  });

  const bank = reactive({
    items: [],
    pagination: buildEmptyPagination(),
    statusStats: buildEmptyStatusStats(),
    questionLists: [],
    loadingPool: false,
    loadingLists: false,
  });

  const wrongFilters = reactive({
    role: "all",
    questionType: "all",
    point: "",
    favoriteOnly: false,
  });

  const wrongBook = reactive({
    items: [],
    loading: false,
    remedial: null,
  });

  const dashboard = reactive({
    totalAttempts: 0,
    accuracy: 0,
    roleProgress: {},
    trend: [],
    radar: [],
    loading: false,
  });

  const assessment = reactive({
    active: false,
    id: null,
    questions: [],
    index: 0,
  });

  const timerState = reactive({
    total: 120,
    left: 0,
    lastElapsedSeconds: null,
    running: false,
  });

  const loading = reactive({
    question: false,
    answer: false,
    meta: false,
    assessment: false,
    import: false,
    export: false,
  });

  let timerHandle = null;

  const viewTitle = computed(
    () => VIEW_TITLE_MAP[activeView.value] || VIEW_TITLE_MAP.practice,
  );

  const roleOptions = computed(() => {
    const roles = meta.value?.roles || {};
    const counts = meta.value?.counts || {};
    return Object.entries(roles).map(([value, config]) => ({
      value,
      name: config?.name || value,
      label: `${config?.name || value}（${counts[value] || 0}题）`,
    }));
  });

  const wrongRoleOptions = computed(() => {
    return roleOptions.value.map((item) => ({
      value: item.value,
      label: item.name,
    }));
  });

  const levelEntries = computed(() => {
    const levels = meta.value?.levels || filterOptions.levels || {};
    return Object.entries(levels).map(([value, label]) => ({
      value,
      label,
    }));
  });

  const questionTypeOptions = computed(() => {
    const source = filterOptions.questionTypes?.length
      ? filterOptions.questionTypes
      : meta.value?.question_types || [];
    return source.map((item) => ({
      value: item,
      label: item,
    }));
  });

  const specialtyOptions = computed(() =>
    filterOptions.specialties.map((item) => ({
      value: item,
      label: item,
    })),
  );

  const pointOptions = computed(() =>
    filterOptions.points.map((item) => ({
      value: item,
      label: item,
    })),
  );

  const companyTypeOptions = computed(() =>
    filterOptions.companyTypes.map((item) => ({
      value: item,
      label: item,
    })),
  );

  const currentRoleProgress = computed(
    () => dashboard.roleProgress?.[currentRole.value] || null,
  );

  const currentQuestionHasOptions = computed(
    () => Boolean(currentQuestion.value?.has_options && currentQuestion.value?.options?.length),
  );

  const canGoPrev = computed(() => historyIndex.value > 0);

  const setStatus = (text, isAlert = false) => {
    status.text = String(text || DEFAULT_STATUS_TEXT);
    status.isAlert = Boolean(isAlert);
  };

  const setCoach = (mood, hint, checklist = []) => {
    coach.mood = String(mood || "待命");
    coach.hint = String(hint || DEFAULT_COACH_HINT);
    coach.checklist = Array.isArray(checklist) && checklist.length
      ? checklist.map((item) => String(item || "").trim()).filter(Boolean)
      : [...DEFAULT_COACH_CHECKLIST];
  };

  const setDefaultCoach = () => {
    setCoach("待命", DEFAULT_COACH_HINT, DEFAULT_COACH_CHECKLIST);
  };

  const setCoachForQuestion = (question) => {
    if (!question) {
      setDefaultCoach();
      return;
    }
    const point = question.points || "当前考点";
    setCoach(
      "准备中",
      `这题聚焦 ${point}，建议先用 20 秒列思路，再组织正式答案。`,
      [
        `识别考点：${point}`,
        currentQuestionHasOptions.value
          ? "先排除错误选项，再确认最优项"
          : "先结论，再按现状-方案-指标展开",
        "至少给出一条可量化结果",
      ],
    );
  };

  const resetTimer = () => {
    if (timerHandle) {
      clearInterval(timerHandle);
      timerHandle = null;
    }
    timerState.left = 0;
    timerState.lastElapsedSeconds = null;
    timerState.running = false;
  };

  const getUserAnswer = () => {
    if (!currentQuestion.value) return "";
    if (currentQuestionHasOptions.value) {
      return selectedOption.value.trim();
    }
    return answerInput.value.trim();
  };

  const syncHistoryQuestion = (questionId, updater) => {
    questionHistory.value = questionHistory.value.map((item) => {
      if (item.id !== questionId) return item;
      const next = updater(item);
      return normalizeQuestion(next, currentRole.value);
    });
  };

  const syncBankQuestion = (questionId, updater) => {
    bank.items = bank.items.map((item) => {
      if (item.id !== questionId) return item;
      return updater(item);
    });
  };

  const syncWrongQuestion = (questionId, updater) => {
    wrongBook.items = wrongBook.items.map((item) => {
      if (item.id !== questionId) return item;
      return updater(item);
    });
  };

  const applyQuestion = (rawQuestion, { fromHistory = false, statusText } = {}) => {
    const question = normalizeQuestion(rawQuestion, currentRole.value);
    if (!question) return;

    currentQuestion.value = question;
    currentResult.value = null;
    pointLoop.value = buildDefaultPointLoop();
    answerInput.value = "";
    selectedOption.value = "";
    errorReason.value = "";

    if (!fromHistory) {
      if (historyIndex.value < questionHistory.value.length - 1) {
        questionHistory.value = questionHistory.value.slice(
          0,
          historyIndex.value + 1,
        );
      }
      questionHistory.value.push(question);
      historyIndex.value = questionHistory.value.length - 1;
    }

    setCoachForQuestion(question);
    setStatus(statusText || "题目已加载");
    startTimer();
  };

  const startTimer = () => {
    resetTimer();
    if (!questionFilters.timedMode || !currentQuestion.value) {
      return;
    }

    timerState.total = Math.max(10, Number(questionFilters.perQuestionSeconds || 120));
    timerState.left = timerState.total;
    timerState.running = true;

    timerHandle = setInterval(async () => {
      timerState.left -= 1;
      timerState.lastElapsedSeconds = timerState.total - timerState.left;

      if (timerState.left === 20) {
        setCoach("节奏提醒", "剩余 20 秒，请立即给出结论并补 1 条关键理由。", [
          "先结论",
          "再证据",
          "最后风险",
        ]);
      }

      if (timerState.left === 8) {
        setCoach("冲刺", "最后 8 秒：优先交付核心答案，不要再展开新分支。", [
          "保主干",
          "少赘述",
          "及时提交",
        ]);
      }

      if (timerState.left <= 0) {
        resetTimer();
        await submitCurrentAnswer(true);
      }
    }, 1000);
  };

  const ensureCurrentRole = () => {
    if (meta.value?.roles?.[currentRole.value]) return;
    const firstRole = Object.keys(meta.value?.roles || {})[0];
    currentRole.value = firstRole || DEFAULT_ROLE;
  };

  const loadRoleOptions = async () => {
    const [specialtiesResponse, optionsResponse] = await Promise.all([
      getPracticeSpecialties(currentRole.value),
      getPracticeFilterOptions(currentRole.value),
    ]);

    filterOptions.specialties = specialtiesResponse?.specialties || [];
    filterOptions.points = optionsResponse?.points || [];
    filterOptions.companyTypes = optionsResponse?.company_types || [];
    filterOptions.questionTypes = optionsResponse?.question_types || [];
    filterOptions.levels = optionsResponse?.levels || {};
    filterOptions.statusOptions = optionsResponse?.status_options || [];

    if (!filterOptions.specialties.includes(questionFilters.specialty)) {
      questionFilters.specialty = "all";
    }
    if (!filterOptions.specialties.includes(bankFilters.specialty)) {
      bankFilters.specialty = "all";
    }
    if (!filterOptions.companyTypes.includes(questionFilters.companyType)) {
      questionFilters.companyType = "all";
    }
    if (!filterOptions.companyTypes.includes(bankFilters.companyType)) {
      bankFilters.companyType = "all";
    }
    if (!filterOptions.points.includes(bankFilters.point)) {
      bankFilters.point = "all";
    }
  };

  const loadMeta = async () => {
    loading.meta = true;
    try {
      meta.value = await getPracticeMeta();
      ensureCurrentRole();
      await loadRoleOptions();
    } finally {
      loading.meta = false;
    }
  };

  const buildRandomQuestionParams = () => {
    const params = {
      position_code: currentRole.value,
      level: normalizeFilterValue(questionFilters.level),
      question_type: normalizeFilterValue(questionFilters.questionType),
      specialty:
        questionFilters.mode === "special"
          ? normalizeFilterValue(questionFilters.specialty)
          : "",
      company_type: normalizeFilterValue(questionFilters.companyType),
    };

    if (selectedListId.value) {
      params.list_id = Number(selectedListId.value);
    }

    return params;
  };

  const loadQuestion = async () => {
    loading.question = true;
    try {
      if (assessment.active) {
        const nextQuestion = assessment.questions[assessment.index];
        if (nextQuestion) {
          applyQuestion(nextQuestion, {
            statusText: `能力测评进行中 ${assessment.index + 1}/${assessment.questions.length}`,
          });
        }
        return;
      }

      const data = await drawPracticeQuestion(buildRandomQuestionParams());
      applyQuestion(data?.question);
    } catch (error) {
      setStatus(error?.message || "题目加载失败", true);
    } finally {
      loading.question = false;
    }
  };

  const loadQuestionById = async (questionId, statusText = "已载入题目") => {
    if (!questionId) return;
    loading.question = true;
    try {
      const data = await getPracticeQuestion(questionId);
      activeView.value = "practice";
      applyQuestion(data?.question, { statusText });
    } catch (error) {
      setStatus(error?.message || "题目加载失败", true);
    } finally {
      loading.question = false;
    }
  };

  const moveHistory = (step) => {
    const nextIndex = historyIndex.value + step;
    if (nextIndex < 0 || nextIndex >= questionHistory.value.length) {
      setStatus("没有更多历史题目", true);
      return;
    }
    historyIndex.value = nextIndex;
    applyQuestion(questionHistory.value[nextIndex], {
      fromHistory: true,
      statusText: "已切换到历史题目",
    });
  };

  const showCurrentSolution = async () => {
    if (!currentQuestion.value?.id) return;
    try {
      const data = await getPracticeSolution(currentQuestion.value.id);
      currentResult.value = {
        kind: "solution",
        data,
      };
      setCoach("讲解中", "先对照参考答案补齐结构，再把这题抽象成可复用模板。", [
        "定位差异",
        "抽模板",
        "写复盘",
      ]);
    } catch (error) {
      setStatus(error?.message || "查看答案失败", true);
    }
  };

  const refreshAfterQuestionMutation = async () => {
    await Promise.allSettled([loadDashboard(), loadQuestionPool(), loadQuestionLists()]);
    if (activeView.value === "wrong") {
      await loadWrongs();
    }
  };

  const submitCurrentAnswer = async (autoTimeout = false) => {
    if (!currentQuestion.value?.id) {
      setStatus("请先加载题目", true);
      return;
    }

    const userAnswer = getUserAnswer();
    if (!userAnswer && !autoTimeout) {
      setStatus("请先作答", true);
      return;
    }

    loading.answer = true;
    const elapsed = timerState.lastElapsedSeconds;
    resetTimer();

    try {
      let data;
      if (assessment.active) {
        data = await submitPracticeAssessmentAnswer({
          assessment_id: assessment.id,
          question_id: currentQuestion.value.id,
          user_answer: userAnswer || "[超时未答]",
          elapsed_seconds: elapsed,
          is_timeout: autoTimeout,
        });
      } else {
        data = await submitPracticeAnswer({
          question_id: currentQuestion.value.id,
          user_answer: userAnswer || "[超时未答]",
          error_reason: errorReason.value.trim(),
          elapsed_seconds: elapsed,
          timed_mode: questionFilters.timedMode,
          is_timeout: autoTimeout,
        });
      }

      currentResult.value = {
        kind: "feedback",
        data,
        autoTimeout,
      };

      if (data?.point_packet) {
        pointLoop.value = {
          type: "packet",
          packet: data.point_packet,
          pointState: data.point_state,
          assessment: null,
          hints: [],
        };
      }

      if (!assessment.active && currentQuestion.value) {
        const nextStatus = data?.is_correct ? "solved" : "wrong";
        currentQuestion.value = {
          ...currentQuestion.value,
          status: nextStatus,
        };
        syncHistoryQuestion(currentQuestion.value.id, (item) => ({
          ...item,
          status: nextStatus,
        }));
      }

      if (data?.point_packet) {
        if (data.is_correct) {
          setCoach(
            "正反馈",
            `你在 ${data.point_packet.point} 的回答完成度不错，下一题继续保持结构化表达。`,
            ["保留结构", "补量化", "讲权衡"],
          );
        } else {
          setCoach(
            "纠偏",
            `这题暴露了 ${data.point_packet.point} 的薄弱点，先练基础再做进阶。`,
            ["回看解析", "补 2 题基础", "再做 1 题进阶"],
          );
        }
      }

      if (assessment.active) {
        assessment.index += 1;
        if (assessment.index < assessment.questions.length) {
          setStatus(
            `测评继续：${assessment.index + 1}/${assessment.questions.length}`,
          );
          await loadQuestion();
        } else {
          const report = await completePracticeAssessment(assessment.id);
          pointLoop.value = {
            type: "assessment",
            packet: null,
            pointState: null,
            assessment: report,
            hints: [],
          };
          assessment.active = false;
          assessment.id = null;
          assessment.questions = [];
          assessment.index = 0;
          setStatus("能力测评完成，已生成报告");
          setCoach(
            "测评总结",
            "优先攻克待补考点，建议今晚做一轮错题+补短板组合训练。",
            ["先弱项", "后综合", "再复盘"],
          );
        }
      } else {
        setStatus(
          data?.is_correct ? "本题已通过" : "本题已加入错题本",
          !data?.is_correct,
        );
      }

      await refreshAfterQuestionMutation();
    } catch (error) {
      setStatus(error?.message || "提交失败", true);
    } finally {
      loading.answer = false;
    }
  };

  const startAssessment = async () => {
    loading.assessment = true;
    try {
      const data = await startPracticeAssessment({
        role: currentRole.value,
        total_count: 12,
      });
      assessment.active = true;
      assessment.id = data?.assessment_id || null;
      assessment.questions = (data?.questions || []).map((item) =>
        normalizeQuestion(item, currentRole.value),
      );
      assessment.index = 0;
      questionFilters.timedMode = true;
      questionFilters.perQuestionSeconds = 90;
      activeView.value = "practice";
      await loadQuestion();
      setStatus("能力测评已开始：每题 90 秒");
    } catch (error) {
      setStatus(error?.message || "能力测评启动失败", true);
    } finally {
      loading.assessment = false;
    }
  };

  const toggleCurrentFavorite = async () => {
    if (!currentQuestion.value?.id) {
      setStatus("暂无可收藏题目", true);
      return;
    }

    try {
      const data = await togglePracticeQuestionFavorite(currentQuestion.value.id);
      currentQuestion.value = {
        ...currentQuestion.value,
        is_favorite: Boolean(data?.is_favorite),
      };

      syncHistoryQuestion(currentQuestion.value.id, (item) => ({
        ...item,
        is_favorite: Boolean(data?.is_favorite),
      }));
      syncBankQuestion(currentQuestion.value.id, (item) => ({
        ...item,
        is_favorite: Boolean(data?.is_favorite),
      }));
      syncWrongQuestion(currentQuestion.value.id, (item) => ({
        ...item,
        is_favorite: Boolean(data?.is_favorite),
      }));

      setStatus(data?.is_favorite ? "已收藏经典题目" : "已取消收藏");
      await loadQuestionPool();
    } catch (error) {
      setStatus(error?.message || "收藏失败", true);
    }
  };

  const loadWrongs = async () => {
    wrongBook.loading = true;
    try {
      const data = await getPracticeWrongs({
        position_code: normalizeFilterValue(wrongFilters.role),
        question_type: normalizeFilterValue(wrongFilters.questionType),
        point: wrongFilters.point.trim(),
        favorite: wrongFilters.favoriteOnly || undefined,
      });
      wrongBook.items = data?.items || [];
    } catch (error) {
      setStatus(error?.message || "错题加载失败", true);
    } finally {
      wrongBook.loading = false;
    }
  };

  const loadWrongRemedial = async (wrongId) => {
    try {
      wrongBook.remedial = await getPracticeWrongRemedial(wrongId);
      setStatus("已推送考点补题");
    } catch (error) {
      setStatus(error?.message || "补短板加载失败", true);
    }
  };

  const removeWrong = async (wrongId) => {
    try {
      await deletePracticeWrong(wrongId);
      await Promise.all([loadWrongs(), loadQuestionPool()]);
      setStatus("已删除错题记录");
    } catch (error) {
      setStatus(error?.message || "删除错题失败", true);
    }
  };

  const toggleWrongFavorite = async (wrongId) => {
    try {
      await togglePracticeWrongFavorite(wrongId);
      await loadWrongs();
      setStatus("错题收藏状态已更新");
    } catch (error) {
      setStatus(error?.message || "错题收藏失败", true);
    }
  };

  const loadQuestionPool = async () => {
    bank.loadingPool = true;
    try {
      const data = await listPracticeQuestions({
        position_code: currentRole.value,
        level: normalizeFilterValue(bankFilters.level),
        question_type: normalizeFilterValue(bankFilters.questionType),
        specialty: normalizeFilterValue(bankFilters.specialty),
        point: normalizeFilterValue(bankFilters.point),
        company_type: normalizeFilterValue(bankFilters.companyType),
        status: normalizeFilterValue(bankFilters.status),
        keyword: bankFilters.keyword.trim(),
        favorite: bankFilters.favoriteOnly || undefined,
        list_id: selectedListId.value ? Number(selectedListId.value) : undefined,
        page: bankFilters.page,
        page_size: bankFilters.pageSize,
      });

      bank.items = data?.items || [];
      bank.pagination = data?.pagination || buildEmptyPagination();
      bank.statusStats = data?.status_stats || buildEmptyStatusStats();
    } catch (error) {
      setStatus(error?.message || "题库加载失败", true);
    } finally {
      bank.loadingPool = false;
    }
  };

  const loadQuestionLists = async () => {
    bank.loadingLists = true;
    try {
      const data = await getPracticeQuestionLists(currentRole.value);
      bank.questionLists = data?.items || [];
    } catch (error) {
      setStatus(error?.message || "题单加载失败", true);
    } finally {
      bank.loadingLists = false;
    }
  };

  const loadDashboard = async () => {
    dashboard.loading = true;
    try {
      const data = await getPracticeDashboard(currentRole.value);
      dashboard.totalAttempts = Number(data?.total_attempts || 0);
      dashboard.accuracy = Number(data?.accuracy || 0);
      dashboard.roleProgress = data?.role_progress || {};
      dashboard.trend = data?.trend || [];
      dashboard.radar = data?.radar || [];
    } catch (error) {
      setStatus(error?.message || "统计加载失败", true);
    } finally {
      dashboard.loading = false;
    }
  };

  const switchView = async (view) => {
    activeView.value = view;
    if (view === "wrong") {
      await loadWrongs();
      return;
    }
    if (view === "dashboard") {
      await loadDashboard();
      return;
    }
    if (view === "bank") {
      await Promise.all([loadQuestionPool(), loadQuestionLists()]);
    }
  };

  const switchBankTab = (tab) => {
    bankTab.value = tab;
  };

  const useQuestionList = async (listId) => {
    selectedListId.value = String(listId || "");
    bankFilters.page = 1;
    bankTab.value = "pool";
    await loadQuestionPool();
    setStatus("已切换到题单题库，可直接筛选做题");
  };

  const selectNextOption = () => {
    if (!currentQuestionHasOptions.value) return;
    const options = currentQuestion.value?.options || [];
    if (!options.length) return;
    const currentIndex = options.findIndex(
      (item) => item.key === selectedOption.value,
    );
    const nextIndex = currentIndex < 0 ? 0 : (currentIndex + 1) % options.length;
    selectedOption.value = options[nextIndex]?.key || "";
  };

  const importQuestionBankFromFile = async (file) => {
    if (!file) return;
    loading.import = true;
    try {
      const text = await file.text();
      const json = JSON.parse(text);
      const items = Array.isArray(json) ? json : json?.items;
      const data = await importPracticeQuestionBank({ items });
      setStatus(`成功导入 ${data?.imported || 0} 道题目`);
      await Promise.all([loadMeta(), loadQuestionPool(), loadQuestionLists()]);
    } catch (error) {
      setStatus(`导入失败：${error?.message || "未知错误"}`, true);
    } finally {
      loading.import = false;
    }
  };

  const exportQuestionRecords = async () => {
    loading.export = true;
    try {
      const blob = await exportPracticeRecords();
      triggerBlobDownload(blob, `records_${Date.now()}.csv`);
      setStatus("答题记录已导出");
    } catch (error) {
      setStatus(error?.message || "导出失败", true);
    } finally {
      loading.export = false;
    }
  };

  const syncIntegrationSnapshot = async () => {
    try {
      await getPracticeIntegrationSnapshot(currentRole.value);
    } catch (_error) {
      // 同步失败不阻塞刷题模式。
    }
  };

  const manualCoachNudge = () => {
    if (!currentQuestion.value) {
      setCoach("待命", "先加载题目，我再给你针对这个考点的作答建议。", [
        "选岗位",
        "开始刷题",
        "再拿建议",
      ]);
      return;
    }

    setCoach(
      "即时建议",
      `当前题是 ${currentQuestion.value.points}，建议用“定义问题 -> 方案 -> 指标 -> 风险 -> 结果”五段答法。`,
      ["先定义", "再方案", "给指标"],
    );
  };

  const changeRole = async (role) => {
    try {
      currentRole.value = role || DEFAULT_ROLE;
      selectedListId.value = "";
      bankFilters.page = 1;
      assessment.active = false;
      assessment.id = null;
      assessment.questions = [];
      assessment.index = 0;
      resetTimer();
      await loadRoleOptions();
      await Promise.all([
        loadQuestionPool(),
        loadQuestionLists(),
        loadDashboard(),
        loadQuestion(),
      ]);
    } catch (error) {
      setStatus(error?.message || "岗位切换失败", true);
    }
  };

  const initialize = async () => {
    if (booting.value) return;
    booting.value = true;
    try {
      setStatus("正在初始化...");
      setDefaultCoach();
      await loadMeta();
      await Promise.all([
        loadQuestion(),
        loadQuestionPool(),
        loadQuestionLists(),
        loadDashboard(),
      ]);
      await syncIntegrationSnapshot();
      setStatus("系统已就绪（支持快捷键：Space/N/P/A）");
      initialized.value = true;
    } catch (error) {
      setStatus(error?.message || "初始化失败", true);
    } finally {
      booting.value = false;
    }
  };

  const teardown = () => {
    resetTimer();
  };

  return {
    meta,
    booting,
    initialized,
    currentRole,
    activeView,
    bankTab,
    selectedListId,
    currentQuestion,
    questionHistory,
    historyIndex,
    answerInput,
    selectedOption,
    errorReason,
    currentResult,
    pointLoop,
    status,
    coach,
    questionFilters,
    filterOptions,
    bankFilters,
    bank,
    wrongFilters,
    wrongBook,
    dashboard,
    assessment,
    timerState,
    loading,
    viewTitle,
    roleOptions,
    wrongRoleOptions,
    levelEntries,
    questionTypeOptions,
    specialtyOptions,
    pointOptions,
    companyTypeOptions,
    currentRoleProgress,
    currentQuestionHasOptions,
    canGoPrev,
    setStatus,
    setCoach,
    loadMeta,
    loadQuestion,
    loadQuestionById,
    moveHistory,
    showCurrentSolution,
    submitCurrentAnswer,
    startAssessment,
    toggleCurrentFavorite,
    loadWrongs,
    loadWrongRemedial,
    removeWrong,
    toggleWrongFavorite,
    loadQuestionPool,
    loadQuestionLists,
    loadDashboard,
    switchView,
    switchBankTab,
    useQuestionList,
    selectNextOption,
    importQuestionBankFromFile,
    exportQuestionRecords,
    manualCoachNudge,
    changeRole,
    initialize,
    teardown,
    clampPercent,
  };
});
