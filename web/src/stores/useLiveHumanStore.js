import { defineStore } from "pinia";
import { ref } from "vue";

export const useLiveHumanStore = defineStore("liveHumanStore", () => {
  const micOn = ref(true);
  const localStream = ref(null);
  const remoteStream = ref(null);

  const localVideoElement = ref(null);
  const remoteVideoElements = ref([]);

  const peer = ref(null);
  const pendingCandidates = ref([]);
  const isMakingOffer = ref(false);

  const bindLocalVideo = () => {
    if (!localVideoElement.value) return;
    localVideoElement.value.srcObject = localStream.value || null;
    localVideoElement.value.play?.().catch(() => {});
  };

  const bindRemoteVideo = () => {
    const targets = Array.isArray(remoteVideoElements.value)
      ? remoteVideoElements.value
      : [];
    targets.forEach((target) => {
      if (!target) return;
      target.srcObject = remoteStream.value || null;
      target.play?.().catch(() => {});
    });
  };

  const setLocalVideoElement = (el) => {
    localVideoElement.value = el || null;
    bindLocalVideo();
  };

  const setRemoteVideoElements = (elements = []) => {
    remoteVideoElements.value = Array.isArray(elements)
      ? elements.filter(Boolean)
      : [];
    bindRemoteVideo();
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

    bindLocalVideo();
  };

  const ensurePeer = ({ iceServers, sendSignal, onStatusChange } = {}) => {
    if (peer.value) return peer.value;

    peer.value = new RTCPeerConnection({ iceServers });

    localStream.value?.getTracks?.().forEach((track) => {
      peer.value.addTrack(track, localStream.value);
    });

    peer.value.onicecandidate = (event) => {
      if (!event.candidate || typeof sendSignal !== "function") return;
      sendSignal("candidate", event.candidate);
    };

    peer.value.ontrack = (event) => {
      const [stream] = event.streams;
      if (!stream) return;
      remoteStream.value = stream;
      bindRemoteVideo();
      onStatusChange?.("音视频已连通");
    };

    peer.value.onconnectionstatechange = () => {
      if (!peer.value) return;
      if (peer.value.connectionState === "connected") {
        onStatusChange?.("连接稳定");
      }
      if (
        peer.value.connectionState === "disconnected" ||
        peer.value.connectionState === "failed"
      ) {
        onStatusChange?.("连接中断，等待重连");
      }
    };

    return peer.value;
  };

  const flushPendingCandidates = async (pc) => {
    while (pendingCandidates.value.length > 0) {
      const candidate = pendingCandidates.value.shift();
      await pc.addIceCandidate(new RTCIceCandidate(candidate));
    }
  };

  const createAndSendOffer = async ({
    iceServers,
    sendSignal,
    targetUserId,
    onStatusChange,
  } = {}) => {
    if (isMakingOffer.value) return;

    isMakingOffer.value = true;
    try {
      const pc = ensurePeer({ iceServers, sendSignal, onStatusChange });
      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);
      sendSignal?.("offer", offer, targetUserId);
      onStatusChange?.("已发起通话邀请，等待接听");
    } finally {
      isMakingOffer.value = false;
    }
  };

  const handleSignalMessage = async (
    msg,
    { iceServers, sendSignal, onStatusChange } = {},
  ) => {
    if (!msg?.type) return;

    const pc = ensurePeer({ iceServers, sendSignal, onStatusChange });

    const senderUserId = String(msg.sender_user_id || msg.user_id || "").trim();

    if (msg.type === "offer") {
      await pc.setRemoteDescription(new RTCSessionDescription(msg.data));
      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);
      sendSignal?.("answer", answer, senderUserId);
      await flushPendingCandidates(pc);
      onStatusChange?.("正在建立连接");
      return;
    }

    if (msg.type === "answer") {
      await pc.setRemoteDescription(new RTCSessionDescription(msg.data));
      await flushPendingCandidates(pc);
      onStatusChange?.("连接协商完成");
      return;
    }

    if (msg.type === "candidate") {
      if (pc.remoteDescription) {
        await pc.addIceCandidate(new RTCIceCandidate(msg.data));
      } else {
        pendingCandidates.value.push(msg.data);
      }
      return;
    }

    if (msg.type === "leave") {
      onStatusChange?.("对端已离开房间");
    }
  };

  const cleanup = () => {
    if (peer.value) {
      peer.value.onicecandidate = null;
      peer.value.ontrack = null;
      peer.value.onconnectionstatechange = null;
      peer.value.close();
      peer.value = null;
    }

    if (localStream.value) {
      localStream.value.getTracks().forEach((track) => track.stop());
      localStream.value = null;
    }

    remoteStream.value = null;
    pendingCandidates.value = [];
    isMakingOffer.value = false;

    if (localVideoElement.value) {
      localVideoElement.value.srcObject = null;
    }

    const targets = Array.isArray(remoteVideoElements.value)
      ? remoteVideoElements.value
      : [];
    targets.forEach((target) => {
      if (!target) return;
      target.srcObject = null;
    });
  };

  return {
    micOn,
    localStream,
    remoteStream,
    isMakingOffer,
    setLocalVideoElement,
    setRemoteVideoElements,
    setMicEnabled,
    toggleMic,
    initLocalMedia,
    ensurePeer,
    createAndSendOffer,
    handleSignalMessage,
    cleanup,
  };
});
