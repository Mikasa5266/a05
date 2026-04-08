<script setup>
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { Upload, RefreshCw, ArrowRight } from "lucide-vue-next";
import { getLatestResumeAnalysis, parseResume } from "../../api/resume";
import { useCareerPrepStore } from "../../stores/useCareerPrepStore";

const router = useRouter();
const prepStore = useCareerPrepStore();

const loading = ref(false);
const loadingLatest = ref(false);
const parsing = ref(false);
const fileInputRef = ref(null);

const analysis = computed(() => prepStore.resumeAnalysis);
const record = computed(() => prepStore.resumeRecord);

const openFilePicker = () => {
  fileInputRef.value?.click();
};

const applyPayload = (payload) => {
  prepStore.setResumePayload({
    analysis: payload?.analysis || null,
    record: payload?.record || null,
  });
};

const loadLatest = async () => {
  loadingLatest.value = true;
  try {
    const res = await getLatestResumeAnalysis();
    applyPayload(res);
  } catch (err) {
    // Keep page quiet on first load when no resume exists.
  } finally {
    loadingLatest.value = false;
  }
};

const handleFileChange = async (event) => {
  const file = event?.target?.files?.[0];
  if (!file) return;

  const formData = new FormData();
  formData.append("file", file);

  parsing.value = true;
  try {
    const res = await parseResume(formData, "web_resume_center");
    applyPayload(res);
    ElMessage.success("简历解析完成");
  } catch (err) {
    ElMessage.error(err?.message || "简历解析失败");
  } finally {
    parsing.value = false;
    if (event?.target) {
      event.target.value = "";
    }
  }
};

const goQuestionBank = () => {
  const code =
    analysis.value?.suggested_positions?.[0]?.position_code ||
    record.value?.matched_position_code ||
    prepStore.selectedPositionCode;
  prepStore.setSelectedPositionCode(code || "backend");
  router.push("/student/question-bank");
};

onMounted(async () => {
  loading.value = true;
  await loadLatest();
  loading.value = false;
});
</script>

<template>
  <div class="space-y-6">
    <section class="bg-white border border-zinc-100 rounded-3xl p-6 shadow-sm">
      <div class="flex items-center justify-between gap-3 flex-wrap">
        <div>
          <h1 class="text-2xl font-bold text-zinc-900">简历解析中心</h1>
          <p class="text-sm text-zinc-500 mt-1">上传简历后自动生成结构化画像与岗位匹配建议</p>
        </div>
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="inline-flex items-center gap-2 px-3 py-2 rounded-xl border border-zinc-200 text-zinc-700 hover:bg-zinc-50"
            :disabled="loadingLatest || parsing"
            @click="loadLatest"
          >
            <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loadingLatest }" />
            拉取最新记录
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-60"
            :disabled="parsing"
            @click="openFilePicker"
          >
            <Upload class="w-4 h-4" />
            {{ parsing ? '解析中...' : '上传简历' }}
          </button>
          <input ref="fileInputRef" type="file" class="hidden" accept=".pdf,.docx,.txt,.md" @change="handleFileChange" />
        </div>
      </div>
    </section>

    <section v-if="loading" class="bg-white border border-zinc-100 rounded-3xl p-8 text-zinc-500">正在加载数据...</section>

    <section v-else-if="!analysis" class="bg-white border border-zinc-100 rounded-3xl p-8 text-zinc-500">
      暂无简历分析结果，请先上传简历。
    </section>

    <section v-else class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <article class="lg:col-span-2 bg-white border border-zinc-100 rounded-3xl p-6 shadow-sm space-y-4">
        <h2 class="text-lg font-semibold text-zinc-900">解析摘要</h2>
        <p class="text-sm text-zinc-600">{{ analysis.profile?.summary || '暂无摘要' }}</p>

        <div>
          <h3 class="text-sm font-semibold text-zinc-800 mb-2">核心技能</h3>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="skill in analysis.skills || []"
              :key="skill.name"
              class="px-2.5 py-1 rounded-full bg-indigo-50 text-indigo-700 text-xs"
            >
              {{ skill.name }}
            </span>
          </div>
        </div>

        <div>
          <h3 class="text-sm font-semibold text-zinc-800 mb-2">待提升能力</h3>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="gap in analysis.missing_skills || []"
              :key="gap"
              class="px-2.5 py-1 rounded-full bg-amber-50 text-amber-700 text-xs"
            >
              {{ gap }}
            </span>
          </div>
        </div>
      </article>

      <article class="bg-white border border-zinc-100 rounded-3xl p-6 shadow-sm space-y-4">
        <h2 class="text-lg font-semibold text-zinc-900">岗位推荐</h2>
        <ul class="space-y-2">
          <li
            v-for="item in analysis.suggested_positions || []"
            :key="item.position_code"
            class="border border-zinc-100 rounded-xl p-3"
          >
            <p class="text-sm font-medium text-zinc-900">{{ item.position_name }}</p>
            <p class="text-xs text-zinc-500 mt-1">匹配分: {{ item.score }}</p>
          </li>
        </ul>

        <button
          type="button"
          class="w-full inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl bg-zinc-900 text-white hover:bg-zinc-800"
          @click="goQuestionBank"
        >
          进入题库训练
          <ArrowRight class="w-4 h-4" />
        </button>
      </article>
    </section>
  </div>
</template>
