import { defineStore } from 'pinia'
import { startInterview, getInterview, submitAnswer, endInterview } from '../api/interview'

export const useInterviewStore = defineStore('interview', {
  state: () => ({
    interview: null,
    currentQuestion: null,
    answers: []
  }),
  actions: {
    async start(data) {
      const res = await startInterview(data)
      this.interview = res.interview
      
      // 后端返回的 interview.questions 是 InterviewQuestion 数组
      // 每个元素包含一个 question 对象
      // 我们需要正确提取 question 对象
      if (res.interview.questions && res.interview.questions.length > 0) {
        this.currentQuestion = res.interview.questions[0].question
      } else {
        this.currentQuestion = null
      }
    },
    async get(id) {
      const res = await getInterview(id)
      this.interview = res.interview
      
      if (res.interview.questions && res.interview.questions.length > 0) {
        const index = res.interview.current_index || 0
        if (index < res.interview.questions.length) {
          this.currentQuestion = res.interview.questions[index].question
        } else {
          this.currentQuestion = null
        }
      }
    },
    async submit(id, data) {
      const res = await submitAnswer(id, data)
      this.answers.push(res.result)
      
      this.interview.current_index++
      
      if (this.interview.questions && this.interview.current_index < this.interview.questions.length) {
        this.currentQuestion = this.interview.questions[this.interview.current_index].question
      } else {
        this.currentQuestion = null
      }
    },
    async end(id) {
      const res = await endInterview(id)
      this.interview = res.interview
    }
  }
})