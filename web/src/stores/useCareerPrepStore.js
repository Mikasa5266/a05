import { defineStore } from "pinia";
import { ref } from "vue";

export const useCareerPrepStore = defineStore(
  "careerPrep",
  () => {
    const selectedPositionCode = ref("backend");
    const selectedDifficulty = ref("campus_intern");
    const resumeAnalysis = ref(null);
    const resumeRecord = ref(null);
    const lastQuestionList = ref([]);

    const setSelectedPositionCode = (code) => {
      const normalized = String(code || "")
        .trim()
        .toLowerCase();
      selectedPositionCode.value = normalized || "backend";
    };

    const setSelectedDifficulty = (difficulty) => {
      const normalized = String(difficulty || "").trim();
      selectedDifficulty.value = normalized || "campus_intern";
    };

    const setResumePayload = ({ analysis, record }) => {
      resumeAnalysis.value = analysis || null;
      resumeRecord.value = record || null;
      const inferredCode =
        analysis?.suggested_positions?.[0]?.position_code ||
        record?.matched_position_code ||
        "";
      if (inferredCode) {
        setSelectedPositionCode(inferredCode);
      }
    };

    const setLastQuestionList = (list) => {
      lastQuestionList.value = Array.isArray(list) ? list : [];
    };

    const clearResumePayload = () => {
      resumeAnalysis.value = null;
      resumeRecord.value = null;
    };

    return {
      selectedPositionCode,
      selectedDifficulty,
      resumeAnalysis,
      resumeRecord,
      lastQuestionList,
      setSelectedPositionCode,
      setSelectedDifficulty,
      setResumePayload,
      setLastQuestionList,
      clearResumePayload,
    };
  },
  {
    persist: true,
  },
);
