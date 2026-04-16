import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
} from "vue";
import { storeToRefs } from "pinia";
import { onBeforeRouteLeave, useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { useUserStore } from "../stores/user";
import { useGroupInterviewStore } from "../stores/useGroupInterviewStore";
import { useGroupWebRTCStore } from "../stores/useGroupWebRTCStore";
import { startLiveInterview } from "../api/interview";
import { API_BASE_URL, WEBRTC_ICE_SERVERS } from "../utils/backend";
import {
  MAX_REMOTE_SLOTS,
  TARGET_PARTICIPANTS,
  TEST_START_THRESHOLD,
  authorizeJoin as helperAuthorizeJoin,
  buildChatItem,
  bindRemoteStreamToSlots,
  buildBackPath,
  ensureInterviewSession as helperEnsureInterviewSession,
  loadInvitationByID as helperLoadInvitationByID,
  removeRoomMember,
  resolveInvitationCode,
  resolveInvitationId,
  upsertRoomMember,
} from "./liveInterviewGroupRoomHelpers";

export function useLiveInterviewGroupRoom() {
  const route = useRoute();
  const router = useRouter();
  const userStore = useUserStore();

  const localVideoRef = ref(null);
  const remoteVideoRefA = ref(null);
  const remoteVideoRefB = ref(null);
  const remoteVideoRefC = ref(null);

  const loading = ref(false);
  const joining = ref(false);
  const finishing = ref(false);
  const startingInterview = ref(false);
  const isRouteLeaving = ref(false);
  const statusText = ref("待进入房间");

  const roomId = ref("");
  const invitationCode = ref("");
  const interviewId = ref(0);
  const invitation = ref(null);

  const messageInput = ref("");
  const messages = ref([]);
  const members = ref([]);

  let signalSocket = null;

  const groupInterviewStore = useGroupInterviewStore();
  const {
    groupStarted,
    groupReadyCount,
    groupStartThreshold,
    groupTargetParticipants,
    currentSpeakerUserId,
    countdownSeconds,
  } = storeToRefs(groupInterviewStore);

  const groupWebRTCStore = useGroupWebRTCStore();
  const { micOn, peers } = storeToRefs(groupWebRTCStore);

  const role = computed(() =>
    String(userStore.userInfo?.role || "")
      .trim()
      .toLowerCase(),
  );
  const selfUserId = computed(() =>
    String(userStore.userInfo?.id || "").trim(),
  );
  const isStudent = computed(() => role.value === "student");
  const hasRoom = computed(() => Boolean(roomId.value));

  const remoteVideoRefs = computed(() => [
    remoteVideoRefA.value,
    remoteVideoRefB.value,
    remoteVideoRefC.value,
  ]);

  const backPath = computed(() => buildBackPath(route.path));

  const roomMembers = computed(() => {
    return [...members.value].sort((a, b) =>
      a.isSelf === b.isSelf ? 0 : a.isSelf ? -1 : 1,
    );
  });

  const remoteMembers = computed(() => {
    return roomMembers.value
      .filter((item) => !item.isSelf)
      .slice(0, MAX_REMOTE_SLOTS);
  });

  const currentSpeakerName = computed(() => {
    const speakerID = String(currentSpeakerUserId.value || "").trim();
    if (!speakerID) return "暂无";
    const member = members.value.find((item) => item.userId === speakerID);
    return member?.displayName || `成员 ${speakerID}`;
  });

  const canVoteStart = computed(() => {
    if (
      !hasRoom.value ||
      !signalSocket ||
      signalSocket.readyState !== WebSocket.OPEN
    )
      return false;
    if (groupStarted.value) return false;
    return true;
  });

  const isInvitationInitiator = computed(() => {
    const invitationId = Number(
      invitation.value?.initiator_user_id || invitation.value?.student_id || 0,
    );
    const currentUserId = Number(selfUserId.value || 0);
    return (
      invitationId > 0 && currentUserId > 0 && invitationId === currentUserId
    );
  });

  const normalizedInvitationStatus = computed(() => {
    return String(invitation.value?.status || "")
      .trim()
      .toLowerCase();
  });

  const canStartInterview = computed(() => {
    if (
      !hasRoom.value ||
      !isInvitationInitiator.value ||
      startingInterview.value
    )
      return false;
    if (groupStarted.value) return false;
    return (
      normalizedInvitationStatus.value === "pending" ||
      normalizedInvitationStatus.value === "accepted"
    );
  });

  function goBack() {
    router.push(backPath.value);
  }

  function getSelfDisplayName() {
    return (
      userStore.userInfo?.username || (isStudent.value ? "候选人" : "面试官")
    );
  }

  function getWsSignalUrl() {
    const url = new URL(API_BASE_URL, window.location.origin);
    url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
    url.pathname = `${url.pathname.replace(/\/$/, "")}/ws/interview/group`;
    url.searchParams.set("room_id", roomId.value);
    url.searchParams.set(
      "invitation_code",
      invitationCode.value || resolveInvitationCode(route),
    );
    url.searchParams.set("token", userStore.token || "");
    return url.toString();
  }

  function upsertMember(payload = {}) {
    members.value = upsertRoomMember({
      members: members.value,
      payload,
      selfUserId: selfUserId.value,
    });
  }

  function removeMember(userId) {
    members.value = removeRoomMember(members.value, userId);
  }

  function appendMessage(kind, text, fromSelf, senderName) {
    const item = buildChatItem({ kind, text, fromSelf, senderName });
    if (!item) return;
    messages.value.push(item);
  }

  function sendSignal(type, data = {}, targetUserId = "") {
    if (!signalSocket || signalSocket.readyState !== WebSocket.OPEN) return;
    const target = String(targetUserId || "").trim();
    const payload = {
      type,
      interview_id: roomId.value,
      data,
    };
    if (target) {
      payload.target_user_id = target;
    }
    signalSocket.send(JSON.stringify(payload));
  }

  function syncRemoteSlots() {
    bindRemoteStreamToSlots({
      refs: remoteVideoRefs.value,
      peers: peers.value,
      remoteMembers: remoteMembers.value,
      selfUserId: selfUserId.value,
      maxRemoteSlots: MAX_REMOTE_SLOTS,
    });
  }

  async function initLocalMedia() {
    await groupWebRTCStore.initLocalMedia({ audio: true, video: true });
    groupWebRTCStore.setLocalVideoElement(localVideoRef.value);
  }

  function sendChatMessage() {
    const content = messageInput.value.trim();
    if (!content) return;
    sendSignal("chat", {
      text: content,
      sender_name: getSelfDisplayName(),
      role: role.value,
      interview_id: interviewId.value,
    });
    messageInput.value = "";
  }

  function sendGroupInvite() {
    sendSignal("group_invite", {
      sender_name: getSelfDisplayName(),
      role: role.value,
      target_participants: TARGET_PARTICIPANTS,
      start_threshold: TEST_START_THRESHOLD,
      interview_id: interviewId.value,
    });
  }

  function voteGroupStart() {
    if (!canVoteStart.value) return;
    sendSignal("group_start_vote", {
      sender_name: getSelfDisplayName(),
      role: role.value,
      target_participants: groupTargetParticipants.value,
      start_threshold: groupStartThreshold.value,
      interview_id: interviewId.value,
    });
  }

  function claimMicRound() {
    sendSignal("group_claim_mic", {
      sender_name: getSelfDisplayName(),
      role: role.value,
      round_duration_sec: 90,
      interview_id: interviewId.value,
    });
  }

  function passToNextSpeaker() {
    sendSignal("group_round_next", {
      sender_name: getSelfDisplayName(),
      role: role.value,
      interview_id: interviewId.value,
    });
  }

  async function triggerStartInterview() {
    if (!invitation.value?.id || !canStartInterview.value) return;
    startingInterview.value = true;
    try {
      const res = await startLiveInterview({
        invitation_id: Number(invitation.value.id),
      });
      const session = res?.session || {};
      const nextStatus = String(session?.status || "")
        .trim()
        .toLowerCase();
      if (nextStatus) {
        invitation.value = { ...invitation.value, status: nextStatus };
      }
      const syncedInterviewId = Number(session?.interview_id || 0);
      if (syncedInterviewId > 0) {
        interviewId.value = syncedInterviewId;
        invitation.value = {
          ...invitation.value,
          interview_id: syncedInterviewId,
        };
      }

      groupInterviewStore.markGroupStarted();
      statusText.value = "主控已开始面试";
      sendSignal("group_start", {
        message: "主控方已开始面试",
        interview_id: interviewId.value,
        started_at: session?.started_at || undefined,
      });
      ElMessage.success("已开始面试");
    } catch (err) {
      ElMessage.error(
        err?.response?.data?.error || err?.message || "开始面试失败",
      );
    } finally {
      startingInterview.value = false;
    }
  }

  async function handleSignalMessage(raw) {
    const msg = JSON.parse(raw);
    if (!msg?.type) return;

    const senderID = String(msg.sender_user_id || msg.user_id || "").trim();
    const data = msg?.data && typeof msg.data === "object" ? msg.data : {};
    const isSelfSignal = senderID && senderID === selfUserId.value;

    if (senderID) {
      upsertMember({
        userId: senderID,
        senderName: data.sender_name,
        role: data.role,
      });
    }

    if (msg.type === "chat") {
      appendMessage(
        "chat",
        data.display_text || data.text,
        isSelfSignal,
        data.sender_name || "成员",
      );
      return;
    }

    if (msg.type === "group_invite") {
      groupInterviewStore.syncGroupInvite(data, {
        startThreshold: TEST_START_THRESHOLD,
        targetParticipants: TARGET_PARTICIPANTS,
      });
      return;
    }

    if (msg.type === "group_start_status") {
      groupInterviewStore.syncGroupStartStatus(data, {
        startThreshold: TEST_START_THRESHOLD,
        targetParticipants: TARGET_PARTICIPANTS,
      });
      if (Boolean(data.started)) {
        statusText.value = "群面已开始";
      }
      return;
    }

    if (msg.type === "group_start") {
      groupInterviewStore.markGroupStarted();
      statusText.value = String(data.message || "群面流程已开始");
      return;
    }

    if (msg.type === "group_round_sync") {
      groupInterviewStore.syncRoundRobinState(data);
      return;
    }

    if (msg.type === "join") {
      statusText.value = "成员已进入房间";
      if (!isSelfSignal && senderID && senderID !== selfUserId.value) {
        await groupWebRTCStore.createAndSendOfferToUser(senderID, {
          iceServers: WEBRTC_ICE_SERVERS,
          selfUserID: selfUserId.value,
          sendSignal,
          onStatusChange: (nextStatus) => {
            statusText.value = nextStatus;
          },
        });
      }
      return;
    }

    if (msg.type === "leave") {
      if (senderID) {
        removeMember(senderID);
        groupWebRTCStore.removePeer(senderID);
      }
      return;
    }

    if (isSelfSignal) return;

    if (
      msg.type === "offer" ||
      msg.type === "answer" ||
      msg.type === "candidate"
    ) {
      await groupWebRTCStore.handleSignalMessage(msg, {
        iceServers: WEBRTC_ICE_SERVERS,
        selfUserID: selfUserId.value,
        sendSignal,
        onStatusChange: (nextStatus) => {
          statusText.value = nextStatus;
        },
      });
      return;
    }

    if (msg.type === "session_sync") {
      const syncedInterviewId = Number(data?.interview_id || 0);
      if (syncedInterviewId > 0 && interviewId.value === 0) {
        interviewId.value = syncedInterviewId;
      }
    }
  }

  function connectSignalSocket() {
    signalSocket = new WebSocket(getWsSignalUrl());

    signalSocket.onopen = () => {
      statusText.value = "已进入房间，等待群面开始";
      upsertMember({
        userId: selfUserId.value,
        senderName: getSelfDisplayName(),
        role: role.value,
      });

      sendSignal("join", {
        role: role.value,
        sender_name: getSelfDisplayName(),
      });

      if (interviewId.value > 0) {
        sendSignal("session_sync", { interview_id: interviewId.value });
      }

      sendGroupInvite();
      sendSignal("group_round_sync_request", {
        interview_id: interviewId.value,
      });
    };

    signalSocket.onmessage = async (event) => {
      try {
        await handleSignalMessage(event.data);
        syncRemoteSlots();
      } catch (err) {
        console.error("signal message handling failed", err);
      }
    };

    signalSocket.onerror = () => {
      statusText.value = "信令连接异常";
    };

    signalSocket.onclose = () => {
      statusText.value = "信令已断开";
    };
  }

  function cleanup() {
    groupInterviewStore.cleanup();
    if (signalSocket) {
      if (signalSocket.readyState === WebSocket.OPEN) {
        sendSignal("leave", {
          user_id: selfUserId.value,
          sender_name: getSelfDisplayName(),
          role: role.value,
        });
      }
      signalSocket.onopen = null;
      signalSocket.onmessage = null;
      signalSocket.onerror = null;
      signalSocket.onclose = null;
      if (
        signalSocket.readyState === WebSocket.OPEN ||
        signalSocket.readyState === WebSocket.CONNECTING
      ) {
        signalSocket.close();
      }
    }
    signalSocket = null;

    groupWebRTCStore.cleanup();
    if (remoteVideoRefA.value) remoteVideoRefA.value.srcObject = null;
    if (remoteVideoRefB.value) remoteVideoRefB.value.srcObject = null;
    if (remoteVideoRefC.value) remoteVideoRefC.value.srcObject = null;
  }

  async function leaveRoom() {
    if (finishing.value) return;
    finishing.value = true;
    cleanup();
    goBack();
    finishing.value = false;
  }

  async function initAndJoinRoom(invitationID) {
    if (joining.value || isRouteLeaving.value) return;

    joining.value = true;
    loading.value = true;
    statusText.value = "正在准备房间...";

    try {
      cleanup();
      members.value = [];
      messages.value = [];
      groupInterviewStore.resetSessionState({
        startThreshold: TEST_START_THRESHOLD,
        targetParticipants: TARGET_PARTICIPANTS,
      });

      if (invitationID <= 0) {
        throw new Error("缺少 invitation_id，无法进入群面房间");
      }

      const invitationData = await helperLoadInvitationByID({
        invitationID,
        isStudent: isStudent.value,
        fallbackInvitationCode: resolveInvitationCode(route),
      });
      invitation.value = invitationData.invitation;
      invitationCode.value = invitationData.invitationCode;
      interviewId.value = invitationData.interviewId;

      const nextInterviewId = await helperEnsureInterviewSession({
        invitation: invitation.value,
        isStudent: isStudent.value,
        selfUserId: selfUserId.value,
        interviewId: interviewId.value,
      });
      if (nextInterviewId > 0) {
        interviewId.value = nextInterviewId;
        invitation.value = invitation.value
          ? { ...invitation.value, interview_id: nextInterviewId }
          : invitation.value;
      }

      const joinData = await helperAuthorizeJoin({
        invitationID,
        invitationCode: invitationCode.value,
        fallbackInvitationCode: resolveInvitationCode(route),
      });
      roomId.value = joinData.roomId;
      invitationCode.value = joinData.invitationCode;
      if (joinData.interviewId > 0) {
        interviewId.value = joinData.interviewId;
      }

      await initLocalMedia();
      connectSignalSocket();

      await nextTick();
      groupWebRTCStore.setLocalVideoElement(localVideoRef.value);
      syncRemoteSlots();
    } catch (err) {
      ElMessage.error(
        err?.response?.data?.error || err?.message || "进入房间失败",
      );
      cleanup();
      roomId.value = "";
    } finally {
      loading.value = false;
      joining.value = false;
    }
  }

  function toggleMic() {
    groupWebRTCStore.toggleMic();
  }

  onMounted(async () => {
    const invitationID = resolveInvitationId(route);
    if (invitationID <= 0) {
      ElMessage.warning("请从面试管理工作台进入指定房间");
      goBack();
      return;
    }
    await initAndJoinRoom(invitationID);
  });

  watch(
    () => [route.params?.id, route.query?.invitation_id],
    async () => {
      const invitationID = resolveInvitationId(route);
      if (invitationID <= 0 || joining.value || isRouteLeaving.value) return;
      await initAndJoinRoom(invitationID);
    },
  );

  watch(localVideoRef, () =>
    groupWebRTCStore.setLocalVideoElement(localVideoRef.value),
  );

  watch(
    [remoteVideoRefA, remoteVideoRefB, remoteVideoRefC, peers, members],
    () => {
      syncRemoteSlots();
    },
    { deep: true },
  );

  watch(micOn, (enabled) => groupWebRTCStore.setMicEnabled(enabled));

  onBeforeRouteLeave(() => {
    isRouteLeaving.value = true;
    cleanup();
    return true;
  });

  onBeforeUnmount(() => {
    isRouteLeaving.value = true;
    cleanup();
  });
  return {
    localVideoRef,
    remoteVideoRefA,
    remoteVideoRefB,
    remoteVideoRefC,
    loading,
    finishing,
    startingInterview,
    statusText,
    roomId,
    messageInput,
    messages,
    roomMembers,
    remoteMembers,
    hasRoom,
    micOn,
    groupStarted,
    groupReadyCount,
    groupStartThreshold,
    groupTargetParticipants,
    currentSpeakerName,
    countdownSeconds,
    canVoteStart,
    isInvitationInitiator,
    canStartInterview,
    goBack,
    getSelfDisplayName,
    sendChatMessage,
    sendGroupInvite,
    voteGroupStart,
    claimMicRound,
    passToNextSpeaker,
    toggleMic,
    triggerStartInterview,
    leaveRoom,
  };
}
