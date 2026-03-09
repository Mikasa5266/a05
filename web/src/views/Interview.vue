<template>
  <div class="interview-chat-container">
    <div class="header-bar" v-if="interviewStore.interview">
      <div class="topic-info">
        <span class="topic-label">当前话题:</span>
        <span class="topic-value">{{ currentTopic || '基础面试' }}</span>
      </div>
      <div class="progress-info">
        <!-- Hide exact count to keep it dynamic -->
        <span>面试进行中</span>
      </div>
    </div>

    <div class="chat-window" ref="chatWindowRef">
      <div v-for="(msg, index) in messages" :key="index" class="message" :class="msg.role">
        <div class="avatar">
          <el-avatar :icon="msg.role === 'ai' ? 'Service' : 'User'" :class="msg.role" />
        </div>
        <div class="content">
          <div class="bubble">
            <p v-if="msg.type === 'text'">{{ msg.content }}</p>
            <div v-else-if="msg.type === 'question'">
              <h3 class="font-bold text-lg mb-2">{{ msg.content.title }}</h3>
              <p class="whitespace-pre-wrap">{{ msg.content.content }}</p>
            </div>
          </div>
        </div>
      </div>
      <div v-if="loading" class="message ai">
        <div class="avatar"><el-avatar icon="Service" class="ai" /></div>
        <div class="content">
          <div class="bubble">
            <p class="loading-text">{{ loadingText || '面试官正在思考中...' }}</p>
            <span class="typing-indicator"><span>.</span><span>.</span><span>.</span></span>
          </div>
        </div>
      </div>
    </div>

    <div class="input-area" v-if="!isCompleted">
      <div v-if="loading" class="loading-hint">
        {{ loadingText || '面试官正在评估你的回答，请稍等...' }}
      </div>
      <el-input
        v-model="inputMessage"
        type="textarea"
        :rows="3"
        placeholder="请输入你的回答..."
        @keydown.enter.ctrl="sendMessage"
      />
      <div class="action-bar">
        <AudioRecorder @record-complete="handleAudioRecord" />
        <el-button type="primary" @click="sendMessage" :loading="loading" :disabled="loading || !inputMessage.trim()">
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
import { ref, onMounted, nextTick } from 'vue'
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
const loadingText = ref('')
const chatWindowRef = ref(null)
const isCompleted = ref(false)
const currentTopic = ref('')

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
        currentTopic.value = interviewStore.interview.current_topic || '基础面试'
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

const handleAudioRecord = async (audioPayload) => {
  const base64Audio = typeof audioPayload === 'string'
    ? audioPayload
    : String(audioPayload?.base64 || '')
  const audioMime = typeof audioPayload === 'string'
    ? ''
    : String(audioPayload?.mime || '').trim()

  if (!base64Audio.trim()) {
    ElMessage.error('语音数据为空')
    return
  }

  addMessage('user', '【语音回答已发送】')
  await submitAnswer('', base64Audio, audioMime)
}

const submitAnswer = async (answerText, audioData = '', audioMime = '') => {
  if (!interviewStore.currentQuestion) return

  loading.value = true
  loadingText.value = '面试官正在评估你的回答...'
  try {
    const payload = {
      question_id: interviewStore.currentQuestion.id,
      answer: answerText,
      audio_data: audioData
    }
    if (audioMime) {
      payload.audio_mime = audioMime
    }

    await interviewStore.submit(interviewStore.interview.id, payload)
    loadingText.value = '面试官正在组织下一轮问题...'
    
    // 获取下一题或结束
    if (interviewStore.currentQuestion) {
      // Update topic
      if (interviewStore.interview.current_topic) {
        currentTopic.value = interviewStore.interview.current_topic
      }
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
    loadingText.value = ''
  }
}

const handleSkip = async () => {
  addMessage('user', '跳过')
  await submitAnswer('跳过')
}

const viewReport = () => {
  router.push(`/student/report/${interviewStore.interview.id}`)
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

.header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  border-bottom: 1px solid #eee;
  background-color: #fff;
  border-radius: 8px 8px 0 0;
}

.topic-info {
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.topic-label {
  color: #909399;
  font-weight: bold;
}

.topic-value {
  color: #409eff;
  font-weight: 600;
  background: #ecf5ff;
  padding: 2px 8px;
  border-radius: 4px;
}

.progress-info {
  font-size: 12px;
  color: #909399;
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

.loading-hint {
  margin-bottom: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  background: #ecf5ff;
  color: #409eff;
  font-size: 13px;
}

.loading-text {
  margin-bottom: 8px;
  color: #606266;
  font-size: 13px;
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
