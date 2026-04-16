import { defineStore } from "pinia";
import { ref } from "vue";

const normalizeUserID = (value) => String(value || "").trim();

export const useGroupWebRTCStore = defineStore("groupWebRTCStore", () => {
  const micOn = ref(true);
  const localStream = ref(null);
  const localVideoElement = ref(null);
  const peers = ref(new Map());

  const clonePeers = () => new Map(peers.value);

  const commitPeer = (userID, nextState) => {
    const key = normalizeUserID(userID);
    if (!key) return;
    const nextMap = clonePeers();
    nextMap.set(key, nextState);
    peers.value = nextMap;
  };

  const removePeer = (userID) => {
    const key = normalizeUserID(userID);
    if (!key) return;

    const state = peers.value.get(key);
    if (state?.pc) {
      state.pc.onicecandidate = null;
      state.pc.ontrack = null;
      state.pc.onconnectionstatechange = null;
      state.pc.close();
    }

    const nextMap = clonePeers();
    nextMap.delete(key);
    peers.value = nextMap;
  };

  const setLocalVideoElement = (el) => {
    localVideoElement.value = el || null;
    if (!localVideoElement.value) return;
    localVideoElement.value.srcObject = localStream.value || null;
    localVideoElement.value.play?.().catch(() => {});
  };

  const setMicEnabled = (enabled) => {
    const next = Boolean(enabled);
    micOn.value = next;
    if (!localStream.value) return;
    localStream.value.getAudioTracks().forEach((track) => {
      track.enabled = next;
    });
  };

  const toggleMic = () => {
    setMicEnabled(!micOn.value);
  };

  const initLocalMedia = async ({ audio = true, video = true } = {}) => {
    if (!navigator.mediaDevices?.getUserMedia) {
      throw new Error("当前浏览器无法访问摄像头/麦克风");
    }

    localStream.value = await navigator.mediaDevices.getUserMedia({
      audio,
      video,
    });

    localStream.value.getAudioTracks().forEach((track) => {
      track.enabled = micOn.value;
    });
    localStream.value.getVideoTracks().forEach((track) => {
      track.enabled = true;
    });

    if (localVideoElement.value) {
      localVideoElement.value.srcObject = localStream.value;
      localVideoElement.value.play?.().catch(() => {});
    }
  };

  const flushPendingCandidates = async (state) => {
    if (!state?.pc) return;
    while (state.pendingCandidates.length > 0) {
      const candidate = state.pendingCandidates.shift();
      await state.pc.addIceCandidate(new RTCIceCandidate(candidate));
    }
  };

  const ensurePeer = (
    userID,
    { iceServers, sendSignal, onStatusChange } = {},
  ) => {
    const key = normalizeUserID(userID);
    if (!key) {
      throw new Error("peer user id is required");
    }

    const existing = peers.value.get(key);
    if (existing?.pc) return existing;

    const pc = new RTCPeerConnection({ iceServers });
    const state = {
      userID: key,
      pc,
      remoteStream: null,
      pendingCandidates: [],
      isMakingOffer: false,
    };

    localStream.value?.getTracks?.().forEach((track) => {
      pc.addTrack(track, localStream.value);
    });

    pc.onicecandidate = (event) => {
      if (!event.candidate || typeof sendSignal !== "function") return;
      sendSignal("candidate", event.candidate, key);
    };

    pc.ontrack = (event) => {
      const [stream] = event.streams;
      if (!stream) return;
      state.remoteStream = stream;
      commitPeer(key, state);
      onStatusChange?.("音视频已连通");
    };

    pc.onconnectionstatechange = () => {
      if (!state.pc) return;
      if (state.pc.connectionState === "connected") {
        onStatusChange?.("连接稳定");
      }
      if (
        state.pc.connectionState === "failed" ||
        state.pc.connectionState === "closed"
      ) {
        removePeer(key);
      }
    };

    commitPeer(key, state);
    return state;
  };

  const createAndSendOfferToUser = async (
    userID,
    { iceServers, sendSignal, onStatusChange, selfUserID } = {},
  ) => {
    const targetUserID = normalizeUserID(userID);
    const selfID = normalizeUserID(selfUserID);
    if (!targetUserID || targetUserID === selfID) return;

    const state = ensurePeer(targetUserID, {
      iceServers,
      sendSignal,
      onStatusChange,
    });
    if (state.isMakingOffer) return;

    state.isMakingOffer = true;
    commitPeer(targetUserID, state);

    try {
      const offer = await state.pc.createOffer();
      await state.pc.setLocalDescription(offer);
      sendSignal?.("offer", offer, targetUserID);
      onStatusChange?.("已发起通话邀请，等待接听");
    } finally {
      state.isMakingOffer = false;
      commitPeer(targetUserID, state);
    }
  };

  const handleSignalMessage = async (
    msg,
    { iceServers, sendSignal, onStatusChange, selfUserID } = {},
  ) => {
    if (!msg?.type) return;

    const senderUserID = normalizeUserID(msg.sender_user_id || msg.user_id);
    const selfID = normalizeUserID(selfUserID);
    if (!senderUserID || senderUserID === selfID) return;

    const state = ensurePeer(senderUserID, {
      iceServers,
      sendSignal,
      onStatusChange,
    });

    if (msg.type === "offer") {
      await state.pc.setRemoteDescription(new RTCSessionDescription(msg.data));
      const answer = await state.pc.createAnswer();
      await state.pc.setLocalDescription(answer);
      sendSignal?.("answer", answer, senderUserID);
      await flushPendingCandidates(state);
      onStatusChange?.("正在建立连接");
      commitPeer(senderUserID, state);
      return;
    }

    if (msg.type === "answer") {
      await state.pc.setRemoteDescription(new RTCSessionDescription(msg.data));
      await flushPendingCandidates(state);
      onStatusChange?.("连接协商完成");
      commitPeer(senderUserID, state);
      return;
    }

    if (msg.type === "candidate") {
      if (state.pc.remoteDescription) {
        await state.pc.addIceCandidate(new RTCIceCandidate(msg.data));
      } else {
        state.pendingCandidates.push(msg.data);
      }
      commitPeer(senderUserID, state);
      return;
    }

    if (msg.type === "leave") {
      removePeer(senderUserID);
    }
  };

  const cleanup = () => {
    Array.from(peers.value.keys()).forEach((userID) => removePeer(userID));
    peers.value = new Map();

    if (localStream.value) {
      localStream.value.getTracks().forEach((track) => track.stop());
      localStream.value = null;
    }

    if (localVideoElement.value) {
      localVideoElement.value.srcObject = null;
    }
  };

  return {
    micOn,
    localStream,
    peers,
    setLocalVideoElement,
    setMicEnabled,
    toggleMic,
    initLocalMedia,
    ensurePeer,
    createAndSendOfferToUser,
    handleSignalMessage,
    removePeer,
    cleanup,
  };
});
