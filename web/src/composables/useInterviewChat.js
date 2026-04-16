import { computed, ref } from "vue";
import {
  submitAnswer as apiSubmitAnswer,
  analyzeSpeechChunk as apiAnalyzeSpeechChunk,
} from "../api/interview";

const classifySpeechRateLevelClient = (rate) => {
  if (rate < 120) return "slow";
  if (rate <= 240) return "normal";
  return "fast";
};

const normalizeChunkTranscript = (text = "") => {
  return String(text || "")
    .replace(/\s+/g, " ")
    .trim();
};

const mergeChunkTranscript = (existing, incoming) => {
  const base = normalizeChunkTranscript(existing);
  const next = normalizeChunkTranscript(incoming);
  if (!next) return base;
  if (!base) return next;
  if (base.includes(next)) return base;
  if (next.includes(base)) return next;

  const maxOverlap = Math.min(base.length, next.length, 24);
  for (let overlap = maxOverlap; overlap >= 4; overlap -= 1) {
    if (base.slice(-overlap) === next.slice(0, overlap)) {
      return `${base}${next.slice(overlap)}`.trim();
    }
  }

  return `${base} ${next}`.trim();
};

const pickSupportedAudioMime = () => {
  const candidates = [
    "audio/webm;codecs=opus",
    "audio/webm",
    "audio/mp4",
    "audio/ogg;codecs=opus",
    "audio/ogg",
  ];
  if (
    typeof MediaRecorder === "undefined" ||
    typeof MediaRecorder.isTypeSupported !== "function"
  ) {
    return "";
  }
  for (const mime of candidates) {
    if (MediaRecorder.isTypeSupported(mime)) return mime;
  }
  return "";
};

const normalizeAudioMime = (mime) => {
  const raw = String(mime || "")
    .trim()
    .toLowerCase();
  if (!raw) return "";
  const semi = raw.indexOf(";");
  return semi > 0 ? raw.slice(0, semi) : raw;
};

const formatFeedback = (feedback) => {
  if (feedback == null) return "回答已提交，建议补充更具体的技术细节。";

  if (typeof feedback === "string") {
    const trimmed = feedback.trim();
    if (trimmed.startsWith("{")) {
      try {
        const parsed = JSON.parse(trimmed);
        if (parsed.evaluation) {
          return trimmed;
        }
      } catch {
        // noop
      }
    }
  }

  const extractText = (val) => {
    if (!val) return [];
    if (typeof val === "string") {
      const text = val.trim();
      if (!text) return [];
      if (text.startsWith("{") || text.startsWith("[")) {
        try {
          return extractText(JSON.parse(text));
        } catch {
          return [text];
        }
      }
      return [text];
    }
    if (Array.isArray(val)) {
      return val.flatMap((item) => extractText(item));
    }
    if (typeof val === "object") {
      const blocks = [];
      if (typeof val.content === "string" && val.content.trim())
        blocks.push(val.content.trim());
      if (Array.isArray(val.suggestions)) {
        val.suggestions.forEach((s) => {
          if (typeof s === "string" && s.trim())
            blocks.push(`建议：${s.trim()}`);
        });
      }
      const keys = [
        "feedback",
        "analysis",
        "comment",
        "summary",
        "advice",
        "suggestion",
        "message",
      ];
      keys.forEach((k) => {
        if (val[k] !== undefined) blocks.push(...extractText(val[k]));
      });
      return blocks;
    }
    return [];
  };

  const texts = extractText(feedback).filter(Boolean);
  return texts.length > 0
    ? texts.join("\n")
    : "回答已提交，建议补充更具体的技术细节。";
};

const splitFeedbackSections = (text) => {
  const source = (text || "").trim();
  if (!source) {
    return {
      evaluation: "回答已提交，建议补充更具体的技术细节。",
      suggestions: [],
      dimensions: null,
      highlights: [],
      gaps: [],
      modelAnswerOutline: "",
      followUp: "",
    };
  }

  if (source.startsWith("{")) {
    try {
      const parsed = JSON.parse(source);
      if (parsed.evaluation) {
        return {
          evaluation: parsed.evaluation || "",
          suggestions: Array.isArray(parsed.suggestions)
            ? parsed.suggestions
            : parsed.suggestions
              ? [parsed.suggestions]
              : [],
          dimensions: parsed.dimensions || null,
          highlights: Array.isArray(parsed.highlights)
            ? parsed.highlights.filter(Boolean)
            : [],
          gaps: Array.isArray(parsed.gaps) ? parsed.gaps.filter(Boolean) : [],
          modelAnswerOutline: parsed.model_answer_outline || "",
          followUp: parsed.follow_up || "",
        };
      }
    } catch {
      // noop
    }
  }

  const evalMatch = source.match(/【评价】([\s\S]*?)(?:【建议】|$)/);
  const suggestBlockMatch = source.match(/【建议】([\s\S]*)$/);
  if (evalMatch || suggestBlockMatch) {
    const evaluationText = (evalMatch?.[1] || "").trim() || source;
    const suggestionLines = (suggestBlockMatch?.[1] || "")
      .split("\n")
      .map((line) => line.replace(/^[-•\d.)、\s]+/, "").trim())
      .filter(Boolean);
    return {
      evaluation: evaluationText,
      suggestions: suggestionLines,
      dimensions: null,
      highlights: [],
      gaps: [],
      modelAnswerOutline: "",
      followUp: "",
    };
  }

  const lines = source
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);
  const evaluationParts = [];
  const suggestions = [];
  lines.forEach((line) => {
    const normalized = line.replace(/^[-•\d.)\s]+/, "").trim();
    if (/^(建议|改进建议|可优化|下一步|你可以)/.test(normalized)) {
      suggestions.push(normalized.replace(/^建议[:：]?\s*/, ""));
      return;
    }
    if (
      /^(1|2|3|4|5)[.)、]\s*/.test(line) &&
      /建议|改进|优化/.test(normalized)
    ) {
      suggestions.push(normalized);
      return;
    }
    if (/^(建议：|建议:)/.test(line)) {
      suggestions.push(line.replace(/^(建议：|建议:)\s*/, "").trim());
      return;
    }
    evaluationParts.push(line);
  });
  return {
    evaluation: evaluationParts.join("\n") || source,
    suggestions,
    dimensions: null,
    highlights: [],
    gaps: [],
    modelAnswerOutline: "",
    followUp: "",
  };
};

const normalizeAnswerSubmitError = (msg = "") => {
  const text = String(msg || "");
  if (!text) return "未知错误";
  if (/err_ssl_protocol_error|ssl\s*protocol/i.test(text)) {
    return "请求协议异常（ERR_SSL_PROTOCOL_ERROR）：通常是 HTTPS 请求打到了仅支持 HTTP 的后端。请检查 VITE_PROXY_TARGET 是否为 http://127.0.0.1:8082，并确保 ngrok 转发到前端端口 3001（无需本地 HTTPS）";
  }
  if (
    /network\s*error|err_network|econnreset|wsarecv|forcibly\s+closed/i.test(
      text,
    )
  ) {
    return "网络连接中断（请求链路异常）。请重试；若使用 ngrok，请确认隧道、前端 3001 与后端 8082 均在线，并检查协议是否为“前端 HTTPS(ngrok) -> 本地 HTTP(8082)”";
  }
  if (
    /field\s+validation.*answer.*required/i.test(text) ||
    /key:\s*'answer'/i.test(text)
  ) {
    return "您似乎没有做出任何回答";
  }
  if (/audio\s+too\s+large|413/i.test(text)) {
    return "语音文件过大，请缩短录音后重试";
  }
  if (
    /status:\s*401|invalid\s+api\s*key|unauthorized|authentication/i.test(text)
  ) {
    return "语音服务鉴权失败，请检查 ASR 的 API Key 是否有效";
  }
  if (/status:\s*429|quota|rate\s*limit|too\s+many\s+requests/i.test(text)) {
    return "语音服务额度或频率受限，请稍后重试";
  }
  if (
    /instruction\s+text|prompt\s+echo|possible\s+model\/provider\s+mismatch/i.test(
      text,
    )
  ) {
    return "语音转写服务返回了提示词回显，当前模型可能不兼容音频转写";
  }
  if (
    /model|unsupported\s+asr\s+provider|not\s+found/i.test(text) &&
    /transcrib|audio/i.test(text)
  ) {
    return "当前语音模型不可用，请检查 asr.model 和服务商兼容性";
  }
  if (
    /failed\s+to\s+transcribe\s+audio/i.test(text) ||
    /empty\s+transcription\s+result/i.test(text)
  ) {
    return "未识别到有效语音，请靠近麦克风并清晰作答后重试";
  }
  return text;
};

const formatVoiceSeconds = (seconds) => {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${mins}:${String(secs).padStart(2, "0")}`;
};

export function useInterviewChat(options = {}) {
  const phase = options.phase;
  const settings = options.settings;
  const interviewId = options.interviewId;
  const currentQuestion = options.currentQuestion;
  const isAvatarSpeaking = options.isAvatarSpeaking;
  const isMicOn = options.isMicOn;
  const stream = options.stream;

  const onAdvanceToNextQuestion =
    typeof options.onAdvanceToNextQuestion === "function"
      ? options.onAdvanceToNextQuestion
      : () => {};
  const onCompleteInterview =
    typeof options.onCompleteInterview === "function"
      ? options.onCompleteInterview
      : async () => {};
  const onScrollToBottom =
    typeof options.onScrollToBottom === "function"
      ? options.onScrollToBottom
      : () => {};
  const onResetQuietSeconds =
    typeof options.onResetQuietSeconds === "function"
      ? options.onResetQuietSeconds
      : () => {};

  const messages = ref([]);
  const userInput = ref("");
  const isProcessing = ref(false);
  const processingHint = ref("");
  const currentQuestionIndex = ref(0);
  const pendingNextQuestion = ref(null);
  const pendingEnd = ref(false);
  const interactionAbortVersion = ref(0);

  const answerVoiceStatus = ref("idle");
  const answerVoiceSeconds = ref(0);
  const answerVoiceError = ref("");

  const speechMetrics = ref({
    speechRate: 0,
    speechRateLevel: "normal",
    fillerWordCount: 0,
    fluencyAlert: false,
    totalFillerWords: 0,
    transcribedText: "",
  });
  const energyLevel = ref(0);
  const speechAnalysisActive = ref(false);
  const speechRateSmoother = ref(0);
  const answerRecordingPeakEnergy = ref(0);
  const chunkTranscriptHistory = ref([]);

  const latestAIMessage = computed(() => {
    const aiMsgs = messages.value.filter(
      (m) => m.role === "ai" || m.type === "system",
    );
    return aiMsgs.length > 0 ? aiMsgs[aiMsgs.length - 1] : null;
  });

  const latestUserTranscript = computed(() => {
    const userMsgs = messages.value.filter(
      (m) =>
        m.role === "user" &&
        typeof m.rawTranscript === "string" &&
        m.rawTranscript.trim(),
    );
    return userMsgs.length > 0
      ? userMsgs[userMsgs.length - 1].rawTranscript
      : "";
  });

  const isHumanInterviewMode = computed(
    () => settings.value.interviewMode === "human",
  );
  const isVideoInterviewMode = computed(
    () =>
      settings.value.presentationMode === "video_avatar" &&
      !isHumanInterviewMode.value,
  );

  const canAnswerCurrentQuestion = computed(() => {
    if (phase.value !== "interview") return false;
    if (!currentQuestion.value?.questionId) return false;
    if (isProcessing.value) return false;
    if (isAvatarSpeaking.value) return false;
    if (latestAIMessage.value?.type === "feedback") return false;
    return latestAIMessage.value?.type === "question";
  });

  let answerMediaRecorder = null;
  let answerAudioChunks = [];
  let answerVoiceTimer = null;
  let answerStatusResetTimer = null;
  let answerRecorderStream = null;
  let answerRecorderMimeType = "";
  let skipNextAnswerSubmit = false;
  let analysisSourceStream = null;

  let audioContext = null;
  let analyserNode = null;
  let chunkMediaRecorder = null;
  let chunkRecordingStream = null;
  let chunkInterval = null;
  let energyAnimFrame = null;
  const speechChunkSeconds = 6;
  let chunkRecorderMimeType = "";

  const appendMessage = (message) => {
    messages.value.push(message);
  };

  const replaceMessages = (list = []) => {
    messages.value = Array.isArray(list) ? list : [];
  };

  const setCurrentQuestionIndex = (nextIndex) => {
    currentQuestionIndex.value = Number(nextIndex) || 0;
  };

  const incrementCurrentQuestionIndex = () => {
    currentQuestionIndex.value += 1;
  };

  const setPendingState = ({ nextQuestion = null, isEnd = false } = {}) => {
    pendingNextQuestion.value = nextQuestion;
    pendingEnd.value = !!isEnd;
  };

  const resetConversationState = () => {
    currentQuestionIndex.value = 0;
    messages.value = [];
    pendingNextQuestion.value = null;
    pendingEnd.value = false;
    userInput.value = "";
  };

  const isInteractionAborted = (version) => {
    return version !== interactionAbortVersion.value;
  };

  const sendSpeechChunk = async (
    audioBase64,
    duration,
    audioMime = "",
    chunkEnergy = 0,
  ) => {
    if (!interviewId.value) return;
    try {
      const res = await apiAnalyzeSpeechChunk(interviewId.value, {
        audio_data: audioBase64,
        audio_mime: audioMime || undefined,
        duration,
        energy_level: chunkEnergy,
      });
      if (!res.metrics) return;

      const m = res.metrics;
      const transcribed = String(m.transcribed_text || "").trim();
      const charCount = Number(m.char_count) || 0;
      const rawRate = Number(m.speech_rate) || 0;
      const audioDetected =
        typeof m.audio_detected === "boolean"
          ? m.audio_detected
          : (Number(chunkEnergy) || 0) >= 0.02;
      let boundedRate = Math.max(0, Math.min(rawRate, 280));

      if (!audioDetected || !transcribed || charCount <= 1) {
        boundedRate = 0;
      }

      const alpha = audioDetected && transcribed ? 0.35 : 0.2;
      if (
        !speechRateSmoother.value ||
        !Number.isFinite(speechRateSmoother.value)
      ) {
        speechRateSmoother.value = boundedRate;
      } else {
        speechRateSmoother.value =
          speechRateSmoother.value * (1 - alpha) + boundedRate * alpha;
      }

      speechMetrics.value.speechRate =
        Math.round(speechRateSmoother.value * 10) / 10;
      speechMetrics.value.speechRateLevel = classifySpeechRateLevelClient(
        speechMetrics.value.speechRate,
      );
      speechMetrics.value.fillerWordCount = m.filler_word_count;
      speechMetrics.value.fluencyAlert = m.fluency_alert;
      speechMetrics.value.totalFillerWords += m.filler_word_count;

      if (audioDetected && transcribed) {
        const merged = mergeChunkTranscript(
          speechMetrics.value.transcribedText,
          transcribed,
        );
        speechMetrics.value.transcribedText = merged;
        chunkTranscriptHistory.value.push(transcribed);
        if (chunkTranscriptHistory.value.length > 120) {
          chunkTranscriptHistory.value.shift();
        }
      }
    } catch (err) {
      console.warn("Speech analysis chunk failed:", err);
    }
  };

  const startChunkRecording = (sourceStream) => {
    if (!sourceStream) return;

    const startNewChunk = () => {
      if (!speechAnalysisActive.value || !sourceStream) return;

      const audioTracks = sourceStream.getAudioTracks();
      if (audioTracks.length === 0) return;
      const clonedTracks = audioTracks.map((track) => track.clone());
      chunkRecordingStream = new MediaStream(clonedTracks);
      const preferredMime = pickSupportedAudioMime();

      try {
        chunkMediaRecorder = preferredMime
          ? new MediaRecorder(chunkRecordingStream, { mimeType: preferredMime })
          : new MediaRecorder(chunkRecordingStream);
      } catch {
        chunkMediaRecorder = new MediaRecorder(chunkRecordingStream);
      }
      chunkRecorderMimeType = normalizeAudioMime(
        chunkMediaRecorder.mimeType || preferredMime,
      );

      const chunks = [];
      let chunkPeakEnergy = 0;
      const chunkEnergySampler = setInterval(() => {
        chunkPeakEnergy = Math.max(
          chunkPeakEnergy,
          Number(energyLevel.value) || 0,
        );
      }, 120);

      chunkMediaRecorder.ondataavailable = (e) => {
        if (e.data.size > 0) chunks.push(e.data);
      };

      chunkMediaRecorder.onstop = () => {
        clearInterval(chunkEnergySampler);
        if (chunkRecordingStream) {
          chunkRecordingStream.getTracks().forEach((track) => track.stop());
          chunkRecordingStream = null;
        }
        if (chunks.length === 0 || !interviewId.value) return;
        const blob = new Blob(chunks, {
          type: chunkRecorderMimeType || "audio/webm",
        });
        const reader = new FileReader();
        reader.onloadend = () => {
          const raw = String(reader.result || "");
          const parts = raw.split(",");
          if (parts.length < 2 || !parts[1]) return;
          sendSpeechChunk(
            parts[1],
            speechChunkSeconds,
            chunkRecorderMimeType || "",
            chunkPeakEnergy,
          );
        };
        reader.readAsDataURL(blob);
      };

      chunkMediaRecorder.start();

      chunkInterval = setTimeout(() => {
        if (chunkMediaRecorder && chunkMediaRecorder.state === "recording") {
          chunkMediaRecorder.stop();
        }
        if (speechAnalysisActive.value) startNewChunk();
      }, speechChunkSeconds * 1000);
    };

    startNewChunk();
  };

  const startSpeechAnalysis = (sourceStream = null) => {
    const activeStream = sourceStream || answerRecorderStream || stream.value;
    if (speechAnalysisActive.value || !activeStream) return;
    const activeAudioTracks = activeStream.getAudioTracks();
    if (!activeAudioTracks.length) return;

    analysisSourceStream = activeStream;
    speechAnalysisActive.value = true;

    audioContext = new (window.AudioContext || window.webkitAudioContext)();
    if (audioContext.state === "suspended") {
      audioContext.resume().catch((err) => {
        console.warn("AudioContext resume failed:", err);
      });
    }

    const source = audioContext.createMediaStreamSource(activeStream);
    analyserNode = audioContext.createAnalyser();
    analyserNode.fftSize = 1024;
    analyserNode.smoothingTimeConstant = 0.82;
    source.connect(analyserNode);

    const dataArray = new Uint8Array(analyserNode.fftSize);
    let smoothedEnergy = 0;
    let noiseFloor = 0.003;

    const updateEnergy = () => {
      if (!speechAnalysisActive.value) return;
      analyserNode.getByteTimeDomainData(dataArray);
      let sumSquares = 0;
      for (let i = 0; i < dataArray.length; i += 1) {
        const sample = (dataArray[i] - 128) / 128;
        sumSquares += sample * sample;
      }
      const rms = Math.sqrt(sumSquares / dataArray.length);
      noiseFloor = noiseFloor * 0.992 + rms * 0.008;
      const gated = Math.max(0, rms - noiseFloor * 1.15);
      const normalized = Math.min(1, gated * 28);
      smoothedEnergy = smoothedEnergy * 0.68 + normalized * 0.32;
      energyLevel.value = Math.min(1, smoothedEnergy);
      if (answerVoiceStatus.value === "recording") {
        answerRecordingPeakEnergy.value = Math.max(
          answerRecordingPeakEnergy.value,
          energyLevel.value,
        );
      }
      energyAnimFrame = requestAnimationFrame(updateEnergy);
    };

    updateEnergy();
    startChunkRecording(activeStream);
  };

  const stopSpeechAnalysis = () => {
    speechAnalysisActive.value = false;
    if (chunkInterval) {
      clearTimeout(chunkInterval);
      chunkInterval = null;
    }
    if (chunkMediaRecorder && chunkMediaRecorder.state === "recording") {
      chunkMediaRecorder.stop();
    }
    if (chunkRecordingStream) {
      chunkRecordingStream.getTracks().forEach((track) => track.stop());
      chunkRecordingStream = null;
    }
    if (energyAnimFrame) {
      cancelAnimationFrame(energyAnimFrame);
      energyAnimFrame = null;
    }
    if (audioContext) {
      audioContext.close().catch(() => {});
      audioContext = null;
    }
    energyLevel.value = 0;
    analyserNode = null;
    chunkMediaRecorder = null;
    analysisSourceStream = null;
  };

  const submitCurrentAnswer = async (
    answerText = "",
    audioData = "",
    audioMime = "",
    requestVersion = interactionAbortVersion.value,
  ) => {
    const currentQ = currentQuestion.value;
    if (!currentQ || !currentQ.questionId) {
      throw new Error("当前题目ID无效，请重新开始面试");
    }

    const payload = {
      question_id: currentQ.questionId,
      question_title: currentQ.title || "",
      question_content: currentQ.content || "",
      answer: answerText,
      audio_data: audioData,
    };
    if (audioMime) {
      payload.audio_mime = audioMime;
    }

    const res = await apiSubmitAnswer(interviewId.value, payload);
    if (isInteractionAborted(requestVersion)) {
      return { aborted: true };
    }
    const result = res.result;
    const formatted = formatFeedback(result.feedback);
    const feedbackSections = splitFeedbackSections(formatted);

    appendMessage({
      role: "ai",
      content: formatted,
      type: "feedback",
      score: result.score,
      feedbackEvaluation: feedbackSections.evaluation,
      feedbackSuggestions: feedbackSections.suggestions,
      feedbackDimensions: feedbackSections.dimensions,
      feedbackHighlights: feedbackSections.highlights,
      feedbackGaps: feedbackSections.gaps,
      feedbackModelAnswer: feedbackSections.modelAnswerOutline,
      feedbackFollowUp: feedbackSections.followUp,
    });

    if (result.next_question) {
      pendingNextQuestion.value = {
        mapId: null,
        questionId: result.next_question.id || currentQ.questionId,
        title: result.next_question.title || "",
        content: result.next_question.content || "",
        expectedAnswer: result.next_question.expected_answer || "",
        source: result.next_question.source || "standard",
      };
      pendingEnd.value = false;
    } else {
      pendingNextQuestion.value = null;
      pendingEnd.value = !!result.interview_completed;
    }

    return result;
  };

  const submitAudioAnswer = async (audioData, audioMime = "") => {
    if (!audioData) return;
    if (isProcessing.value) return;
    const requestVersion = interactionAbortVersion.value;

    const userMsg = {
      role: "user",
      content: "【语音回答转写中...】",
      rawTranscript: "",
    };
    const userMsgIndex = messages.value.length;
    appendMessage({ ...userMsg });

    isProcessing.value = true;
    processingHint.value = "面试官正在转写并评估你的语音回答...";
    answerVoiceStatus.value = "submitting";
    answerVoiceError.value = "";

    try {
      const result = await submitCurrentAnswer(
        "",
        audioData,
        audioMime,
        requestVersion,
      );
      if (isInteractionAborted(requestVersion) || result?.aborted) {
        answerVoiceStatus.value = "idle";
        answerVoiceError.value = "";
        return;
      }
      const transcript = String(result?.answer || "").trim();
      const plainText = transcript || "（未识别到有效语音文本）";
      const rendered = `【语音回答】\n${plainText}`;
      userMsg.content = rendered;
      userMsg.rawTranscript = plainText;
      if (messages.value[userMsgIndex]) {
        messages.value[userMsgIndex] = { ...userMsg };
      }
      answerVoiceStatus.value = "success";
      if (answerStatusResetTimer) {
        clearTimeout(answerStatusResetTimer);
      }
      answerStatusResetTimer = setTimeout(() => {
        answerStatusResetTimer = null;
        if (answerVoiceStatus.value === "success") {
          answerVoiceStatus.value = "idle";
        }
      }, 1600);
    } catch (error) {
      if (isInteractionAborted(requestVersion)) {
        return;
      }
      const rawErrMsg =
        error?.response?.data?.error || error?.message || "未知错误";
      const errMsg = normalizeAnswerSubmitError(rawErrMsg);
      answerVoiceError.value = errMsg;
      answerVoiceStatus.value = "error";

      if (errMsg.includes("not in progress") || errMsg.includes("已结束")) {
        appendMessage({
          role: "ai",
          content: "面试结束！辛苦了。您可以点击下方按钮查看详细报告。",
          type: "system",
        });
        await onCompleteInterview();
      } else {
        appendMessage({
          role: "system",
          content: `提交语音答案失败：${errMsg}`,
          type: "system",
        });
      }
    } finally {
      isProcessing.value = false;
      processingHint.value = "";
      onScrollToBottom();
    }
  };

  const startAnswerRecording = async () => {
    if (isProcessing.value || !interviewId.value) return;
    if (!canAnswerCurrentQuestion.value) {
      answerVoiceError.value = "请等待题目描述完成后再开始语音回答";
      answerVoiceStatus.value = "error";
      return;
    }
    if (!isMicOn.value) {
      answerVoiceError.value = "麦克风已关闭，请先开启麦克风";
      answerVoiceStatus.value = "error";
      return;
    }

    answerVoiceStatus.value = "requesting";
    answerVoiceError.value = "";
    answerVoiceSeconds.value = 0;
    onResetQuietSeconds();
    answerAudioChunks = [];
    answerRecorderMimeType = "";
    skipNextAnswerSubmit = false;
    answerRecordingPeakEnergy.value = 0;
    speechRateSmoother.value = 0;
    speechMetrics.value.transcribedText = "";
    chunkTranscriptHistory.value = [];

    try {
      answerRecorderStream = await navigator.mediaDevices.getUserMedia({
        audio: true,
      });
      const preferredMime = pickSupportedAudioMime();

      try {
        answerMediaRecorder = preferredMime
          ? new MediaRecorder(answerRecorderStream, { mimeType: preferredMime })
          : new MediaRecorder(answerRecorderStream);
      } catch {
        answerMediaRecorder = new MediaRecorder(answerRecorderStream);
      }
      answerRecorderMimeType = normalizeAudioMime(
        answerMediaRecorder.mimeType || preferredMime,
      );

      answerMediaRecorder.ondataavailable = (event) => {
        if (event.data && event.data.size > 0) {
          answerAudioChunks.push(event.data);
        }
      };

      answerMediaRecorder.onstop = async () => {
        if (answerVoiceTimer) {
          clearInterval(answerVoiceTimer);
          answerVoiceTimer = null;
        }

        if (skipNextAnswerSubmit) {
          skipNextAnswerSubmit = false;
          answerAudioChunks = [];
          answerVoiceStatus.value = "idle";
          answerVoiceError.value = "";
          answerVoiceSeconds.value = 0;
          answerMediaRecorder = null;
          return;
        }

        if (!answerAudioChunks.length) {
          answerVoiceError.value = "未检测到有效语音，请重试";
          answerVoiceStatus.value = "error";
          answerMediaRecorder = null;
          return;
        }
        if (
          isVideoInterviewMode.value &&
          answerRecordingPeakEnergy.value < 0.06
        ) {
          answerVoiceError.value =
            "未检测到有效语音输入，请检查麦克风并靠近后重试";
          answerVoiceStatus.value = "error";
          answerMediaRecorder = null;
          return;
        }

        answerVoiceStatus.value = "transcribing";
        const audioBlob = new Blob(answerAudioChunks, {
          type: answerRecorderMimeType || "audio/webm",
        });
        const reader = new FileReader();
        reader.onloadend = async () => {
          const raw = String(reader.result || "");
          const parts = raw.split(",");
          if (parts.length < 2 || !parts[1]) {
            answerVoiceError.value = "音频编码失败，请重试";
            answerVoiceStatus.value = "error";
            answerMediaRecorder = null;
            return;
          }
          await submitAudioAnswer(parts[1], answerRecorderMimeType || "");
          answerMediaRecorder = null;
        };
        reader.readAsDataURL(audioBlob);
      };

      answerMediaRecorder.start();
      if (isVideoInterviewMode.value) {
        startSpeechAnalysis(answerRecorderStream);
      }
      answerVoiceStatus.value = "recording";
      answerVoiceTimer = setInterval(() => {
        answerVoiceSeconds.value += 1;
      }, 1000);
    } catch (err) {
      console.warn("startAnswerRecording failed:", err);
      answerVoiceError.value = "无法访问麦克风权限";
      answerVoiceStatus.value = "error";
    }
  };

  const stopAnswerRecording = () => {
    if (!answerMediaRecorder || answerVoiceStatus.value !== "recording") return;
    skipNextAnswerSubmit = false;
    stopSpeechAnalysis();
    onResetQuietSeconds();
    answerMediaRecorder.stop();
    if (answerRecorderStream) {
      answerRecorderStream.getTracks().forEach((track) => track.stop());
      answerRecorderStream = null;
    }
  };

  const toggleAnswerRecording = async () => {
    if (answerVoiceStatus.value === "recording") {
      stopAnswerRecording();
      return;
    }
    await startAnswerRecording();
  };

  const sendMessage = async () => {
    if (isProcessing.value) return;
    if (latestAIMessage.value?.type === "feedback") {
      onAdvanceToNextQuestion();
      return;
    }
    if (isVideoInterviewMode.value) return;
    if (!userInput.value.trim()) return;

    const answer = userInput.value;
    userInput.value = "";

    appendMessage({
      role: "user",
      content: answer,
    });

    isProcessing.value = true;
    processingHint.value = "面试官正在评估你的回答...";
    const requestVersion = interactionAbortVersion.value;

    try {
      const result = await submitCurrentAnswer(answer, "", "", requestVersion);
      if (isInteractionAborted(requestVersion) || result?.aborted) {
        return;
      }
      processingHint.value = "面试官正在生成下一轮追问...";
    } catch (error) {
      if (isInteractionAborted(requestVersion)) {
        return;
      }
      console.error("Failed to submit answer:", error);
      const rawErrMsg =
        error?.response?.data?.error || error?.message || "未知错误";
      const errMsg = normalizeAnswerSubmitError(rawErrMsg);

      if (errMsg.includes("not in progress") || errMsg.includes("已结束")) {
        appendMessage({
          role: "ai",
          content: "面试结束！辛苦了。您可以点击下方按钮查看详细报告。",
          type: "system",
        });
        await onCompleteInterview();
      } else {
        appendMessage({
          role: "system",
          content: `提交答案失败：${errMsg}`,
          type: "system",
        });
      }
    } finally {
      isProcessing.value = false;
      processingHint.value = "";
      onScrollToBottom();
    }
  };

  const forceInterruptCurrentAnswerFlow = () => {
    interactionAbortVersion.value += 1;
    isProcessing.value = false;
    processingHint.value = "";
    onResetQuietSeconds();

    if (answerStatusResetTimer) {
      clearTimeout(answerStatusResetTimer);
      answerStatusResetTimer = null;
    }

    if (answerMediaRecorder && answerMediaRecorder.state === "recording") {
      skipNextAnswerSubmit = true;
      stopSpeechAnalysis();
      answerMediaRecorder.stop();
    } else {
      stopSpeechAnalysis();
    }

    if (answerRecorderStream) {
      answerRecorderStream.getTracks().forEach((track) => track.stop());
      answerRecorderStream = null;
    }

    if (answerVoiceTimer) {
      clearInterval(answerVoiceTimer);
      answerVoiceTimer = null;
    }

    answerAudioChunks = [];
    answerVoiceStatus.value = "idle";
    answerVoiceError.value = "";
    answerVoiceSeconds.value = 0;
  };

  const getVoiceStatusLabel = () => {
    const labels = {
      idle: "待命",
      requesting: "请求麦克风权限",
      recording: `录音中 ${formatVoiceSeconds(answerVoiceSeconds.value)}`,
      transcribing: "语音转写中",
      submitting: "提交语音答案中",
      success: "语音答案已提交",
      error: answerVoiceError.value || "语音失败",
    };
    return labels[answerVoiceStatus.value] || "待命";
  };

  const getAdditionalStreams = () =>
    [answerRecorderStream, analysisSourceStream].filter(Boolean);

  const cleanupInterviewChat = () => {
    forceInterruptCurrentAnswerFlow();
    answerMediaRecorder = null;
    answerAudioChunks = [];
    answerVoiceStatus.value = "idle";
  };

  return {
    messages,
    userInput,
    isProcessing,
    processingHint,
    currentQuestionIndex,
    pendingNextQuestion,
    pendingEnd,
    latestAIMessage,
    latestUserTranscript,
    canAnswerCurrentQuestion,
    answerVoiceStatus,
    answerVoiceSeconds,
    answerVoiceError,
    speechMetrics,
    energyLevel,
    speechAnalysisActive,
    appendMessage,
    replaceMessages,
    setCurrentQuestionIndex,
    incrementCurrentQuestionIndex,
    setPendingState,
    resetConversationState,
    submitCurrentAnswer,
    sendMessage,
    startAnswerRecording,
    stopAnswerRecording,
    toggleAnswerRecording,
    getVoiceStatusLabel,
    startSpeechAnalysis,
    stopSpeechAnalysis,
    getAdditionalStreams,
    normalizeAnswerSubmitError,
    forceInterruptCurrentAnswerFlow,
    cleanupInterviewChat,
  };
}
