import { defineStore } from 'pinia'

export const useResumeStore = defineStore('resume', {
  state: () => ({
    pendingFile: null
  }),
  actions: {
    setPendingFile(file) {
      this.pendingFile = file
    },
    clearPendingFile() {
      this.pendingFile = null
    }
  }
})
