<template>
  <div class="interview-chat-container">
    <div class="chat-window" ref="chatWindowRef">
      <div v-for="(msg, index) in messages" :key="index" class="message" :class="msg.role">
        <div class="avatar">
          <el-avatar :icon="msg.role === 'ai' ? 'Service' : 'User'" :class="msg.role" />
        </div>
        <div class="content">
          <div class="bubble">
            <p v-if="msg.type === 'text'">{{ msg.content }}</p>
            <div v-else-if="msg.type === 'question'">
              <h3>{{ msg.content.title }}</h3>
              <p>{{ msg.content.content }}</p>
            </div>
          </div>
        </div>
      </div>
      <div v-if="loading" class="message ai">
        <div class="avatar"><el-avatar icon="Service" class="ai" /></div>
        <div class="content">
          <div class="bubble">
            <span class="typing-indicator"><span>.</span><span>.</span><span>.</span></span>
          </div>
        </div>
      </div>
    </div>

    <div class="input-area" v-if="!isCompleted">
      <el-input
        v-model="inputMessage"
        type="textarea"
        :rows="3"
        placeholder="请输入你的回答..."
        @keydown.enter.ctrl="sendMessage"
      />
      <div class="action-bar">
        <AudioRecorder @record-complete="handleAudioRecord" />
        <el-button type="primary" @click="sendMessage" :loading="loading" :disabled="!inputMessage.trim()">
          发送 (Ctrl+Enter)
        </el-button>
        <el-button @click="handleSkip" :disabled="loading">跳过</el-button>
      </div>
    </div>
    
    <div class="completed-area" v-else>
      <el-result icon="success" title="面试结束" sub-title="感谢你的参与，报告已生成">
        <template #extra>
          <el-button type="primary" @click="viewReport">查看报告</el-button>
        </template>
      </el-result>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, watch } from 'vue'
import { useInterviewStore } from '../stores/interview'
import { useRouter, useRoute } from 'vue-router'
import AudioRecorder from '../components/AudioRecorder.vue'
import { ElMessage } from 'element-plus'

const interviewStore = useInterviewStore()
const router = useRouter()
const route = useRoute()

const messages = ref([])
const inputMessage = ref('')
const loading = ref(false)
const chatWindowRef = ref(null)
const isCompleted = ref(false)

const scrollToBottom = () => {
  nextTick(() => {
    if (chatWindowRef.value) {
      chatWindowRef.value.scrollTop = chatWindowRef.value.scrollHeight
    }
  })
}

const addMessage = (role, content, type = 'text') => {
  messages.value.push({ role, content, type })
  scrollToBottom()
}

const initInterview = async () => {
  const id = route.params.id
  if (id) {
    try {
      loading.value = true
      await interviewStore.get(id)
      
      // 添加欢迎语
      addMessage('ai', '你好！我是你的 AI 面试官。我们将开始进行面试，请准备好回答接下来的问题。')
      
      // 添加第一题
      if (interviewStore.currentQuestion) {
        setTimeout(() => {
          addMessage('ai', interviewStore.currentQuestion, 'question')
        }, 1000)
      } else {
        isCompleted.value = true
      }
    } catch (error) {
      ElMessage.error('加载面试失败')
    } finally {
      loading.value = false
    }
  }
}

const sendMessage = async () => {
  if (!inputMessage.value.trim()) return
  
  const answer = inputMessage.value
  inputMessage.value = ''
  addMessage('user', answer)
  
  await submitAnswer(answer)
}

const handleAudioRecord = (base64Audio) => {
  addMessage('user', '【语音回答已发送】')
  submitAnswer('', base64Audio)
}

const submitAnswer = async (answerText, audioData = '') => {
  if (!interviewStore.currentQuestion) return

  loading.value = true
  try {
    await interviewStore.submit(interviewStore.interview.id, {
      question_id: interviewStore.currentQuestion.id,
      answer: answerText,
      audio_data: audioData
    })
    
    // 获取下一题或结束
    if (interviewStore.currentQuestion) {
      setTimeout(() => {
        addMessage('ai', interviewStore.currentQuestion, 'question')
      }, 1000)
    } else {
      setTimeout(() => {
        addMessage('ai', '面试已结束，正在为你生成评估报告...')
        isCompleted.value = true
        // 自动结束面试
        interviewStore.end(interviewStore.interview.id)
      }, 1000)
    }
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '提交失败')
  } finally {
    loading.value = false
  }
}

const handleSkip = async () => {
  addMessage('user', '跳过')
  await submitAnswer('跳过')
}

const viewReport = () => {
  router.push(`/report/${interviewStore.interview.id}`)
}

onMounted(() => {
  initInterview()
})
</script>

<style scoped>
.interview-chat-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 80px); /* 减去头部高度 */
  max-width: 800px;
  margin: 0 auto;
  background-color: #fff;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  border-radius: 8px;
}

.chat-window {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background-color: #f5f7fa;
}

.message {
  display: flex;
  margin-bottom: 20px;
}

.message.user {
  flex-direction: row-reverse;
}

.avatar {
  margin: 0 10px;
}

.content {
  max-width: 70%;
}

.bubble {
  padding: 12px 16px;
  border-radius: 8px;
  background-color: #fff;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  line-height: 1.5;
}

.message.user .bubble {
  background-color: #95ec69;
}

.message.ai .el-avatar {
  background-color: #409eff;
}

.input-area {
  padding: 20px;
  border-top: 1px solid #e6e6e6;
  background-color: #fff;
}

.action-bar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  margin-top: 10px;
  gap: 10px;
}

.typing-indicator span {
  display: inline-block;
  width: 4px;
  height: 4px;
  background-color: #909399;
  border-radius: 50%;
  animation: typing 1.4s infinite ease-in-out both;
  margin: 0 2px;
}

.typing-indicator span:nth-child(1) { animation-delay: -0.32s; }
.typing-indicator span:nth-child(2) { animation-delay: -0.16s; }

@keyframes typing {
  0%, 80%, 100% { transform: scale(0); }
  40% { transform: scale(1); }
}

.completed-area {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>