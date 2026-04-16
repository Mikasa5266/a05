import { defineStore } from "pinia";
import {
  deleteHumanInvitation,
  getLiveInterviewWorkbench,
  respondHumanInvitation,
} from "../api/interview";

const TAB_KEYS = [
  "invite_list",
  "pending",
  "processed",
  "in_progress",
  "history",
];

export const useInterviewWorkbenchStore = defineStore("interviewWorkbench", {
  state: () => ({
    activeTab: "invite_list",
    loading: false,
    actionLoadingId: 0,
    fetchedAt: 0,
    serverTime: "",
    inviteList: [],
    pending: [],
    processed: [],
    inProgress: [],
    history: [],
  }),
  getters: {
    summary(state) {
      return {
        invite_count: state.inviteList.length,
        pending_count: state.pending.length,
        processed_count: state.processed.length,
        in_progress_count: state.inProgress.length,
        history_count: state.history.length,
      };
    },
    currentList(state) {
      if (state.activeTab === "pending") return state.pending;
      if (state.activeTab === "processed") return state.processed;
      if (state.activeTab === "in_progress") return state.inProgress;
      if (state.activeTab === "history") return state.history;
      return state.inviteList;
    },
  },
  actions: {
    setActiveTab(tab) {
      if (TAB_KEYS.includes(tab)) {
        this.activeTab = tab;
      }
    },
    async fetchWorkbench() {
      this.loading = this.fetchedAt === 0;
      try {
        const res = await getLiveInterviewWorkbench();
        const payload = res?.workbench || {};
        this.inviteList = Array.isArray(payload.invite_list)
          ? payload.invite_list
          : [];
        this.pending = Array.isArray(payload.pending) ? payload.pending : [];
        this.processed = Array.isArray(payload.processed)
          ? payload.processed
          : [];
        this.inProgress = Array.isArray(payload.in_progress)
          ? payload.in_progress
          : [];
        this.history = Array.isArray(payload.history) ? payload.history : [];
        this.serverTime = payload.server_time || "";
        this.fetchedAt = Date.now();
      } finally {
        this.loading = false;
      }
    },
    async respondInvitation(invitationId, action) {
      this.actionLoadingId = Number(invitationId || 0);
      try {
        await respondHumanInvitation(invitationId, action);
        await this.fetchWorkbench();
      } finally {
        this.actionLoadingId = 0;
      }
    },
    async deleteInvitation(invitationId) {
      this.actionLoadingId = Number(invitationId || 0);
      try {
        await deleteHumanInvitation(invitationId);
        await this.fetchWorkbench();
      } finally {
        this.actionLoadingId = 0;
      }
    },
  },
});
