<script setup>
import { computed } from 'vue';
import { ArrowLeft, Loader2, Mic, MicOff, PhoneOff, SendHorizontal, Users } from 'lucide-vue-next';
import { useLiveInterviewGroupRoom } from '../composables/useLiveInterviewGroupRoom';

const {
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
} = useLiveInterviewGroupRoom();

const genericMemberPattern = /^成员\s*\d+$/;

const resolveMemberDisplayName = (member) => {
  const normalized = String(member?.displayName || '').trim();
  if (!normalized) return '';
  if (!genericMemberPattern.test(normalized)) return normalized;

  const idx = roomMembers.value.findIndex((item) => item.userId === member.userId);
  if (idx >= 0) {
    return `成员 ${idx + 1}`;
  }
  return normalized;
};

const resolveRemoteSeatLabel = (seatIndex) => {
  const member = remoteMembers.value?.[seatIndex];
  if (!member) return `席位 ${seatIndex + 2}（待接入）`;
  return resolveMemberDisplayName(member);
};

const totalMemberLabel = computed(() => `${roomMembers.value.length} 人`);
</script>

<template>
  <div class="min-h-screen bg-slate-950 text-slate-100 px-4 py-5 md:px-8 md:py-6">
    <div class="max-w-425 mx-auto">
      <header class="flex items-start justify-between gap-4 mb-5">
        <div>
          <h1 class="text-2xl md:text-3xl font-semibold tracking-wide">群面实战房间</h1>
          <p class="text-sm text-slate-300 mt-1">多位面试者 + 1 位 AI 面试官协作群面</p>
        </div>
        <button class="px-3 py-2 rounded-lg border border-slate-700 bg-slate-900 hover:bg-slate-800 inline-flex items-center gap-2" @click="goBack">
          <ArrowLeft class="w-4 h-4" />
          返回
        </button>
      </header>

      <div v-if="!hasRoom && !loading" class="h-[calc(100vh-9rem)] flex items-center justify-center">
        <div class="w-full max-w-xl rounded-2xl border border-slate-700 bg-slate-900 p-8">
          <h2 class="text-xl font-semibold">等待房间授权</h2>
          <p class="text-sm text-slate-300 mt-3">群面房间仅允许受邀用户进入。</p>
          <button class="mt-6 w-full rounded-lg bg-sky-600 hover:bg-sky-500 py-2" @click="goBack">返回工作台</button>
        </div>
      </div>

      <div v-else class="rounded-2xl border border-slate-700 bg-slate-900 p-4 md:p-5 h-[calc(100vh-9rem)] min-h-155 flex flex-col gap-4">
        <div v-if="loading" class="h-full flex items-center justify-center gap-2 text-slate-300">
          <Loader2 class="w-5 h-5 animate-spin" />
          <span>正在初始化设备与连接...</span>
        </div>

        <template v-else>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div class="rounded-xl border border-slate-700 bg-slate-800/60 p-3">
              <p class="text-xs uppercase tracking-wide text-slate-400">房间号</p>
              <p class="mt-2 text-base font-semibold break-all">{{ roomId || '--' }}</p>
            </div>
            <div class="rounded-xl border border-slate-700 bg-slate-800/60 p-3">
              <p class="text-xs uppercase tracking-wide text-slate-400">开考状态</p>
              <p class="mt-2 text-base font-semibold">{{ groupStarted ? '进行中' : '等待开考' }}</p>
              <p class="text-sm text-slate-300 mt-1">{{ statusText }}</p>
              <p class="text-xs text-slate-400 mt-1">投票：{{ groupReadyCount }} / {{ groupStartThreshold }} · 目标 {{ groupTargetParticipants }} 人</p>
              <p class="text-xs text-slate-400 mt-1">当前发言：{{ currentSpeakerName }}<span v-if="countdownSeconds > 0">（{{ countdownSeconds }}s）</span></p>
            </div>
          </div>

          <div class="flex flex-wrap gap-2 shrink-0">
            <button class="px-3 py-2 rounded-lg border border-slate-700 bg-slate-800 hover:bg-slate-700 inline-flex items-center gap-2" @click="sendGroupInvite">
              <Users class="w-4 h-4" />
              同步群面配置
            </button>
            <button
              v-if="isInvitationInitiator"
              class="px-3 py-2 rounded-lg border border-sky-500/40 bg-sky-500/20 hover:bg-sky-500/30 disabled:opacity-60"
              :disabled="!canStartInterview"
              @click="triggerStartInterview"
            >
              {{ startingInterview ? '开始中...' : '主控开始面试' }}
            </button>
            <button class="px-3 py-2 rounded-lg border border-slate-700 bg-slate-800 hover:bg-slate-700 disabled:opacity-60" :disabled="!canVoteStart" @click="voteGroupStart">
              发送准备信号
            </button>
            <button class="px-3 py-2 rounded-lg border border-slate-700 bg-slate-800 hover:bg-slate-700 disabled:opacity-60" :disabled="!hasRoom || !groupStarted" @click="claimMicRound">
              抢麦发言
            </button>
            <button class="px-3 py-2 rounded-lg border border-slate-700 bg-slate-800 hover:bg-slate-700 disabled:opacity-60" :disabled="!hasRoom || !groupStarted" @click="passToNextSpeaker">
              下一位
            </button>
            <button class="px-3 py-2 rounded-lg border border-slate-700 bg-slate-800 hover:bg-slate-700 inline-flex items-center gap-2" @click="toggleMic">
              <Mic v-if="micOn" class="w-4 h-4" />
              <MicOff v-else class="w-4 h-4" />
              {{ micOn ? '语音开启中' : '语音关闭' }}
            </button>
            <button class="ml-auto px-4 py-2 rounded-lg border border-rose-400/40 bg-rose-500/20 hover:bg-rose-500/30 inline-flex items-center gap-2 disabled:opacity-60" :disabled="finishing" @click="leaveRoom">
              <PhoneOff class="w-4 h-4" />
              {{ finishing ? '正在结束...' : '结束并离开' }}
            </button>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-[minmax(0,1.75fr)_minmax(320px,1fr)] gap-3 flex-1 min-h-0">
            <div class="grid grid-cols-1 md:grid-cols-2 auto-rows-fr gap-3 min-h-0">
              <article class="rounded-xl border border-slate-700 bg-slate-800/60 p-3 flex flex-col min-h-0">
                <div class="text-xs text-slate-300 mb-2">我的画面 · {{ getSelfDisplayName() }}</div>
                <div class="relative rounded-lg overflow-hidden border border-slate-700 bg-slate-950 h-full min-h-42.5">
                  <video ref="localVideoRef" autoplay playsinline muted class="w-full h-full object-cover"></video>
                </div>
              </article>

              <article class="rounded-xl border border-slate-700 bg-slate-800/60 p-3 flex flex-col min-h-0">
                <div class="text-xs text-slate-300 mb-2">{{ resolveRemoteSeatLabel(0) }}</div>
                <div class="relative rounded-lg overflow-hidden border border-slate-700 bg-slate-950 h-full min-h-42.5">
                  <video ref="remoteVideoRefA" autoplay playsinline class="w-full h-full object-cover"></video>
                </div>
              </article>

              <article class="rounded-xl border border-slate-700 bg-slate-800/60 p-3 flex flex-col min-h-0">
                <div class="text-xs text-slate-300 mb-2">{{ resolveRemoteSeatLabel(1) }}</div>
                <div class="relative rounded-lg overflow-hidden border border-slate-700 bg-slate-950 h-full min-h-42.5">
                  <video ref="remoteVideoRefB" autoplay playsinline class="w-full h-full object-cover"></video>
                </div>
              </article>

              <article class="rounded-xl border border-slate-700 bg-slate-800/60 p-3 flex flex-col min-h-0">
                <div class="text-xs text-slate-300 mb-2">{{ resolveRemoteSeatLabel(2) }}</div>
                <div class="relative rounded-lg overflow-hidden border border-slate-700 bg-slate-950 h-full min-h-42.5">
                  <video ref="remoteVideoRefC" autoplay playsinline class="w-full h-full object-cover"></video>
                </div>
              </article>
            </div>

            <div class="grid grid-rows-[minmax(0,1fr)_minmax(0,210px)] gap-3 min-h-90 lg:min-h-0">
              <div class="rounded-xl border border-slate-700 bg-slate-800/40 p-3 flex flex-col min-h-0 overflow-hidden">
                <div class="flex items-center justify-between gap-2 mb-2 shrink-0">
                  <p class="text-xs text-slate-300">群面公屏</p>
                  <span class="text-xs text-slate-400">{{ messages.length }} 条</span>
                </div>
                <div class="flex-1 min-h-0 overflow-y-auto pr-1 space-y-2">
                  <p v-if="messages.length === 0" class="text-xs text-slate-400 rounded-lg border border-slate-700/70 bg-slate-900/35 px-3 py-3 leading-5">
                    公屏已就绪，消息会在此区域滚动展示。建议先发送准备信号，再同步关键观点与结论。
                  </p>
                  <div
                    v-for="item in messages"
                    :key="item.id"
                    class="rounded-lg border px-3 py-2 text-sm"
                    :class="item.fromSelf ? 'border-sky-400/40 bg-sky-500/10' : 'border-slate-600 bg-slate-800/70'"
                  >
                    <div class="flex items-center justify-between text-xs text-slate-400 mb-1">
                      <span>{{ item.senderName }}</span>
                      <span>{{ item.createdAt }}</span>
                    </div>
                    <p class="wrap-break-word whitespace-pre-wrap">{{ item.text }}</p>
                  </div>
                </div>
                <div class="mt-2 flex gap-2 shrink-0">
                  <input
                    v-model="messageInput"
                    type="text"
                    class="flex-1 rounded-lg border border-slate-600 bg-slate-900 px-3 py-2 text-sm"
                    placeholder="输入群面公屏消息"
                    @keyup.enter="sendChatMessage"
                  />
                  <button class="px-3 rounded-lg bg-sky-600 hover:bg-sky-500" @click="sendChatMessage">
                    <SendHorizontal class="w-4 h-4" />
                  </button>
                </div>
              </div>

              <div class="rounded-xl border border-slate-700 bg-slate-800/40 p-3 flex flex-col min-h-0 overflow-hidden">
                <div class="flex items-center justify-between gap-2 mb-2 shrink-0">
                  <p class="text-xs text-slate-300">在场成员</p>
                  <span class="text-xs rounded-full border border-slate-600 px-2 py-0.5 text-slate-200">{{ totalMemberLabel }}</span>
                </div>
                <div class="flex-1 min-h-0 overflow-y-auto pr-1 space-y-2">
                  <p v-if="roomMembers.length === 0" class="text-xs text-slate-400 py-2">等待成员进入房间...</p>
                  <div
                    v-for="member in roomMembers"
                    :key="member.userId"
                    class="rounded-lg border border-slate-700/80 bg-slate-900/40 px-3 py-2 flex items-center justify-between gap-2"
                  >
                    <div class="inline-flex items-center gap-2 text-sm text-slate-100">
                      <Users class="w-4 h-4 text-slate-300" />
                      <span>{{ resolveMemberDisplayName(member) || '成员' }}</span>
                    </div>
                    <span class="text-xs text-slate-400">{{ member.isSelf ? '我' : '在线' }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>
