import { defineStore } from "pinia";
import { ref } from "vue";

const DEFAULT_TARGET_PARTICIPANTS = 4;
const DEFAULT_START_THRESHOLD = 2;
const DEFAULT_ROUND_DURATION = 90;

const normalizePositiveInt = (value, fallback) => {
  const parsed = Number(value);
  if (!Number.isFinite(parsed) || parsed <= 0) return fallback;
  return Math.floor(parsed);
};

export const useGroupInterviewStore = defineStore("groupInterviewStore", () => {
  const groupStarted = ref(false);
  const groupReadyCount = ref(0);
  const groupStartThreshold = ref(DEFAULT_START_THRESHOLD);
  const groupTargetParticipants = ref(DEFAULT_TARGET_PARTICIPANTS);

  const speakingQueue = ref([]);
  const currentSpeakerUserId = ref("");
  const roundDurationSec = ref(DEFAULT_ROUND_DURATION);
  const roundStartedAtMs = ref(0);
  const countdownSeconds = ref(0);

  let countdownTimer = null;

  const stopRoundCountdown = () => {
    if (!countdownTimer) return;
    clearInterval(countdownTimer);
    countdownTimer = null;
  };

  const calcCountdownByNow = () => {
    if (!roundStartedAtMs.value || !roundDurationSec.value) return 0;
    const elapsed = Math.floor((Date.now() - roundStartedAtMs.value) / 1000);
    const remaining = Math.max(0, roundDurationSec.value - elapsed);
    return remaining;
  };

  const startRoundCountdown = () => {
    stopRoundCountdown();
    countdownSeconds.value = calcCountdownByNow();
    if (countdownSeconds.value <= 0) return;

    countdownTimer = setInterval(() => {
      countdownSeconds.value = calcCountdownByNow();
      if (countdownSeconds.value <= 0) {
        stopRoundCountdown();
      }
    }, 1000);
  };

  const resetRoundRobin = () => {
    speakingQueue.value = [];
    currentSpeakerUserId.value = "";
    roundDurationSec.value = DEFAULT_ROUND_DURATION;
    roundStartedAtMs.value = 0;
    countdownSeconds.value = 0;
    stopRoundCountdown();
  };

  const resetSessionState = ({ targetParticipants, startThreshold } = {}) => {
    groupStarted.value = false;
    groupReadyCount.value = 0;
    groupTargetParticipants.value = normalizePositiveInt(
      targetParticipants,
      DEFAULT_TARGET_PARTICIPANTS,
    );
    groupStartThreshold.value = normalizePositiveInt(
      startThreshold,
      DEFAULT_START_THRESHOLD,
    );
    resetRoundRobin();
  };

  const syncGroupInvite = (payload = {}, defaults = {}) => {
    groupStartThreshold.value = normalizePositiveInt(
      payload.start_threshold,
      normalizePositiveInt(defaults.startThreshold, DEFAULT_START_THRESHOLD),
    );
    groupTargetParticipants.value = normalizePositiveInt(
      payload.target_participants,
      normalizePositiveInt(
        defaults.targetParticipants,
        DEFAULT_TARGET_PARTICIPANTS,
      ),
    );
  };

  const syncGroupStartStatus = (payload = {}, defaults = {}) => {
    groupReadyCount.value = normalizePositiveInt(payload.ready_count, 0);
    groupStartThreshold.value = normalizePositiveInt(
      payload.start_threshold,
      normalizePositiveInt(defaults.startThreshold, DEFAULT_START_THRESHOLD),
    );
    groupTargetParticipants.value = normalizePositiveInt(
      payload.target_participants,
      normalizePositiveInt(
        defaults.targetParticipants,
        DEFAULT_TARGET_PARTICIPANTS,
      ),
    );

    if (Boolean(payload.started)) {
      groupStarted.value = true;
    }
  };

  const markGroupStarted = () => {
    groupStarted.value = true;
  };

  const syncRoundRobinState = (payload = {}) => {
    const nextSpeaker = String(payload.current_speaker_user_id || "").trim();
    const queue = Array.isArray(payload.queue_user_ids)
      ? payload.queue_user_ids
          .map((item) => String(item || "").trim())
          .filter(Boolean)
      : [];

    speakingQueue.value = queue;
    currentSpeakerUserId.value = nextSpeaker;
    roundDurationSec.value = normalizePositiveInt(
      payload.round_duration_sec,
      DEFAULT_ROUND_DURATION,
    );

    const startedAtRaw = String(payload.round_started_at || "").trim();
    const startedAt = startedAtRaw ? new Date(startedAtRaw).getTime() : 0;
    roundStartedAtMs.value =
      Number.isFinite(startedAt) && startedAt > 0 ? startedAt : 0;

    countdownSeconds.value = normalizePositiveInt(
      payload.countdown_seconds,
      calcCountdownByNow(),
    );

    if (currentSpeakerUserId.value && roundStartedAtMs.value > 0) {
      startRoundCountdown();
    } else {
      stopRoundCountdown();
      countdownSeconds.value = 0;
    }
  };

  const cleanup = () => {
    stopRoundCountdown();
  };

  return {
    groupStarted,
    groupReadyCount,
    groupStartThreshold,
    groupTargetParticipants,
    speakingQueue,
    currentSpeakerUserId,
    roundDurationSec,
    countdownSeconds,
    resetSessionState,
    syncGroupInvite,
    syncGroupStartStatus,
    markGroupStarted,
    syncRoundRobinState,
    cleanup,
  };
});
