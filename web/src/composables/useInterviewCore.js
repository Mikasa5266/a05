import axios from "axios";
import { ElMessage, ElMessageBox } from "element-plus";
import { computed, onBeforeUnmount, ref } from "vue";
import { synthesizeInterviewSpeech as apiSynthesizeInterviewSpeech } from "../api/interview";
import { useBlindBox } from "./useBlindBox";
import { useInterviewRecording } from "./useInterviewRecording";
import { useMediaDevices } from "./useMediaDevices";

export function useInterviewCore(options = {}) {
  const settings = options.settings;
  const interviewId = options.interviewId;
  let isProcessingRef = options.isProcessing || null;
  let processingHintRef = options.processingHint || null;

  const getAdditionalStreams =
    typeof options.getAdditionalStreams === "function"
      ? options.getAdditionalStreams
      : () => [];
  const onCameraStopped =
    typeof options.onCameraStopped === "function"
      ? options.onCameraStopped
      : () => {};

  const {
    isCameraOn,
    isMicOn,
    stream,
    toggleCamera,
    toggleMic,
    startCamera,
    stopCamera,
  } = useMediaDevices({
    getAdditionalStreams,
    onCameraStopped,
  });

  const {
    blindBoxScenario,
    blindBoxRevealing,
    blindBoxRevealed,
    questionTimeLimit,
    questionTimer,
    pressureLevel,
    pressureColors,
    pressureLabels,
    drawBlindBox,
    reDrawBlindBox,
    startQuestionTimer,
    stopQuestionTimer,
  } = useBlindBox();

  const {
    recordingStatus,
    startInterviewRecording,
    stopAndUploadInterviewRecording,
    cleanupInterviewRecording,
  } = useInterviewRecording({
    interviewId,
    settings,
    stream,
    isMicOn,
  });

  const isAvatarSpeaking = ref(false);
  let currentSpeechAudio = null;

  const managedTimeouts = new Set();
  const interviewInitStageTexts = [
    "正在分析您的岗位需求...",
    "正在从题库匹配核心考察点...",
    "AI 面试官正在生成专属考题...",
  ];
  const interviewInitReadyText = "准备就绪！";
  const initLoadingStageText = ref("");
  const initLoadingStageIndex = ref(0);
  const initLoadingElapsedSeconds = ref(0);

  let initLoadingStageTimer = null;
  let initLoadingElapsedTimer = null;

  const initLoadingStageTotal = computed(
    () => interviewInitStageTexts.length + 1,
  );

  const registerManagedTimeout = (callback, delay) => {
    const timer = window.setTimeout(() => {
      managedTimeouts.delete(timer);
      try {
        const maybePromise = callback();
        if (maybePromise && typeof maybePromise.then === "function") {
          maybePromise.catch((err) => {
            console.error("Delayed task failed:", err);
          });
        }
      } catch (err) {
        console.error("Delayed task failed:", err);
      }
    }, delay);
    managedTimeouts.add(timer);
    return timer;
  };

  const clearManagedTimeouts = () => {
    managedTimeouts.forEach((timer) => window.clearTimeout(timer));
    managedTimeouts.clear();
  };

  const clearInitLoadingTimers = () => {
    if (initLoadingStageTimer) {
      window.clearInterval(initLoadingStageTimer);
      initLoadingStageTimer = null;
    }
    if (initLoadingElapsedTimer) {
      window.clearInterval(initLoadingElapsedTimer);
      initLoadingElapsedTimer = null;
    }
  };

  const startInterviewInitLoadingFlow = () => {
    clearInitLoadingTimers();
    initLoadingStageIndex.value = 0;
    initLoadingElapsedSeconds.value = 0;
    initLoadingStageText.value = interviewInitStageTexts[0];
    if (processingHintRef?.value !== undefined) {
      processingHintRef.value = interviewInitStageTexts[0];
    }

    initLoadingElapsedTimer = window.setInterval(() => {
      initLoadingElapsedSeconds.value += 1;
    }, 1000);

    initLoadingStageTimer = window.setInterval(() => {
      if (!isProcessingRef?.value) return;
      initLoadingStageIndex.value =
        (initLoadingStageIndex.value + 1) % interviewInitStageTexts.length;
      const nextStage = interviewInitStageTexts[initLoadingStageIndex.value];
      initLoadingStageText.value = nextStage;
      if (processingHintRef?.value !== undefined) {
        processingHintRef.value = nextStage;
      }
    }, 4000);
  };

  const stopInterviewInitLoadingFlow = () => {
    clearInitLoadingTimers();
    initLoadingStageText.value = "";
    initLoadingStageIndex.value = 0;
    initLoadingElapsedSeconds.value = 0;
  };

  const markInterviewInitReady = async () => {
    clearInitLoadingTimers();
    initLoadingStageIndex.value = interviewInitStageTexts.length;
    initLoadingStageText.value = interviewInitReadyText;
    if (processingHintRef?.value !== undefined) {
      processingHintRef.value = interviewInitReadyText;
    }
    await new Promise((resolve) => {
      registerManagedTimeout(resolve, 520);
    });
  };

  const isInterviewInitTimeoutError = (error) => {
    if (!axios.isAxiosError(error)) {
      return false;
    }
    const status = Number(error?.response?.status || 0);
    const code = String(error?.code || "").toUpperCase();
    const message = String(error?.message || "").toLowerCase();
    return (
      status === 504 ||
      code === "ECONNABORTED" ||
      code === "ETIMEDOUT" ||
      message.includes("timeout") ||
      message.includes("超时")
    );
  };

  const showInterviewInitTimeoutDialog = async () => {
    try {
      await ElMessageBox.confirm(
        "当前面试初始化耗时较长，可能是 AI 服务排队或网络波动。你可以立即重试，系统会保留刚才的配置。",
        "启动超时",
        {
          type: "warning",
          confirmButtonText: "立即重试",
          cancelButtonText: "稍后再试",
          closeOnClickModal: false,
          distinguishCancelAndClose: true,
        },
      );
      return true;
    } catch {
      return false;
    }
  };

  const stopAISpeech = () => {
    if (currentSpeechAudio) {
      currentSpeechAudio.pause();
      currentSpeechAudio = null;
    }
    isAvatarSpeaking.value = false;
  };

  const speakAIText = async (text) => {
    if (settings?.value?.interviewMode === "human") return;
    if (settings?.value?.presentationMode !== "video_avatar") return;
    if (!interviewId?.value || !text) return;

    try {
      const res = await apiSynthesizeInterviewSpeech(interviewId.value, {
        text,
      });
      if (!res?.audio_base64) return;
      stopAISpeech();
      const audio = new Audio(`data:audio/mpeg;base64,${res.audio_base64}`);
      currentSpeechAudio = audio;
      isAvatarSpeaking.value = true;
      audio.onended = () => {
        isAvatarSpeaking.value = false;
        currentSpeechAudio = null;
      };
      audio.onerror = () => {
        isAvatarSpeaking.value = false;
        currentSpeechAudio = null;
      };
      await audio.play();
    } catch (err) {
      isAvatarSpeaking.value = false;
      console.warn("TTS playback failed:", err);
    }
  };

  const setPresentationMode = async (mode) => {
    if (mode === "video_avatar" && settings?.value?.interviewMode === "human") {
      settings.value.presentationMode = "text_voice";
      stopAISpeech();
      stopCamera();
      ElMessage.warning(
        "真人面试模式不展示 AI 虚拟面试官，请使用文字语音模式或进入实时面试间",
      );
      return;
    }

    settings.value.presentationMode = mode;
    if (mode === "video_avatar") {
      await startCamera();
      return;
    }

    stopAISpeech();
    stopCamera();
  };

  const setInterviewMode = (mode) => {
    settings.value.interviewMode = mode;
    if (
      mode === "human" &&
      settings?.value?.presentationMode === "video_avatar"
    ) {
      settings.value.presentationMode = "text_voice";
      stopAISpeech();
      stopCamera();
    }
  };

  const bindProcessingState = ({ isProcessing, processingHint } = {}) => {
    isProcessingRef = isProcessing || null;
    processingHintRef = processingHint || null;
  };

  const cleanupInterviewCore = () => {
    clearManagedTimeouts();
    clearInitLoadingTimers();
    cleanupInterviewRecording();
    stopQuestionTimer();
    stopAISpeech();
    stopCamera();
  };

  onBeforeUnmount(() => {
    cleanupInterviewCore();
  });

  return {
    isCameraOn,
    isMicOn,
    stream,
    toggleCamera,
    toggleMic,
    startCamera,
    stopCamera,

    blindBoxScenario,
    blindBoxRevealing,
    blindBoxRevealed,
    questionTimeLimit,
    questionTimer,
    pressureLevel,
    pressureColors,
    pressureLabels,
    drawBlindBox,
    reDrawBlindBox,
    startQuestionTimer,
    stopQuestionTimer,

    recordingStatus,
    startInterviewRecording,
    stopAndUploadInterviewRecording,
    cleanupInterviewRecording,

    isAvatarSpeaking,
    stopAISpeech,
    speakAIText,
    setPresentationMode,
    setInterviewMode,
    bindProcessingState,

    registerManagedTimeout,
    clearManagedTimeouts,
    initLoadingStageText,
    initLoadingStageIndex,
    initLoadingElapsedSeconds,
    initLoadingStageTotal,
    startInterviewInitLoadingFlow,
    stopInterviewInitLoadingFlow,
    markInterviewInitReady,
    isInterviewInitTimeoutError,
    showInterviewInitTimeoutDialog,
    cleanupInterviewCore,
  };
}
