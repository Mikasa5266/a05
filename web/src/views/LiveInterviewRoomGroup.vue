<script setup>
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
</script>

<template>
  <div class="min-h-screen bg-slate-950 text-slate-100 px-4 py-5 md:px-8 md:py-6">
    <div class="max-w-425 mx-auto">
      <header class="flex items-start justify-between gap-4 mb-5">
        <div>
          <h1 class="text-2xl md:text-3xl font-semibold tracking-wide">群面实战房间</h1>
          <p class="text-sm text-slate-300 mt-1">多人 AI / 真人混合群面</p>
        </div>
        <button class="px-3 py-2 rounded-lg border border-slate-700 bg-slate-900 hover:bg-slate-800 inline-flex items-center gap-2" @click="goBack">
          <ArrowLeft class="w-4 h-4" />
          返回
        </button>
      </header>

      <div v-if="!hasRoom && !loading" class="h-[calc(100vh-12rem)] flex items-center justify-center">
        <div class="w-full max-w-xl rounded-2xl border border-slate-700 bg-slate-900 p-8">
          <h2 class="text-xl font-semibold">等待房间授权</h2>
          <p class="text-sm text-slate-300 mt-3">群面房间仅允许受邀用户进入。</p>
          <button class="mt-6 w-full rounded-lg bg-sky-600 hover:bg-sky-500 py-2" @click="goBack">返回工作台</button>
        </div>
      </div>

      <div v-else class="rounded-2xl border border-slate-700 bg-slate-900 p-4 md:p-5 h-[calc(100vh-12rem)] flex flex-col gap-4">
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

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 flex-1 min-h-0">
            <article class="rounded-xl border border-slate-700 bg-slate-800/60 p-3 flex flex-col min-h-0">
              <div class="text-xs text-slate-300 mb-2">我的画面 · {{ getSelfDisplayName() }}</div>
              <div class="relative rounded-lg overflow-hidden border border-slate-700 bg-slate-950 aspect-video">
                <video ref="localVideoRef" autoplay playsinline muted class="w-full h-full object-cover"></video>
              </div>
            </article>

            <article class="rounded-xl border border-slate-700 bg-slate-800/60 p-3 flex flex-col min-h-0">
              <div class="text-xs text-slate-300 mb-2">{{ remoteMembers[0]?.displayName || '席位 2（待接入）' }}</div>
              <div class="relative rounded-lg overflow-hidden border border-slate-700 bg-slate-950 aspect-video">
                <video ref="remoteVideoRefA" autoplay playsinline class="w-full h-full object-cover"></video>
              </div>
            </article>

            <article class="rounded-xl border border-slate-700 bg-slate-800/60 p-3 flex flex-col min-h-0">
              <div class="text-xs text-slate-300 mb-2">{{ remoteMembers[1]?.displayName || '席位 3（待接入）' }}</div>
              <div class="relative rounded-lg overflow-hidden border border-slate-700 bg-slate-950 aspect-video">
                <video ref="remoteVideoRefB" autoplay playsinline class="w-full h-full object-cover"></video>
              </div>
            </article>

            <article class="rounded-xl border border-slate-700 bg-slate-800/60 p-3 flex flex-col min-h-0">
              <div class="text-xs text-slate-300 mb-2">{{ remoteMembers[2]?.displayName || '席位 4（待接入）' }}</div>
              <div class="relative rounded-lg overflow-hidden border border-slate-700 bg-slate-950 aspect-video">
                <video ref="remoteVideoRefC" autoplay playsinline class="w-full h-full object-cover"></video>
              </div>
            </article>
          </div>

          <div class="flex flex-wrap gap-2">
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

          <div class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_340px] gap-3 min-h-0">
            <div class="rounded-xl border border-slate-700 bg-slate-800/40 p-3 flex flex-wrap gap-2">
              <div v-for="member in roomMembers" :key="member.userId" class="inline-flex items-center gap-1 rounded-full border border-slate-600 px-2 py-1 text-xs text-slate-200">
                <Users class="w-3 h-3" />
                {{ member.displayName }}
              </div>
            </div>

            <div class="rounded-xl border border-slate-700 bg-slate-800/40 p-3 flex flex-col min-h-0">
              <div class="flex-1 min-h-0 overflow-y-auto space-y-2">
                <p v-if="messages.length === 0" class="text-xs text-slate-400">等待消息中...</p>
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
              <div class="mt-2 flex gap-2">
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
          </div>
        </template>
      </div>
    </div>
  </div>
</template>
