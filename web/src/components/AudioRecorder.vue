<template>
  <div class="audio-recorder">
    <el-button 
      type="primary" 
      :icon="isRecording ? 'VideoPause' : 'Microphone'" 
      circle 
      size="large"
      @click="toggleRecording"
      :class="{ recording: isRecording }"
    />
    <div v-if="isRecording" class="recording-status">
      正在录音... {{ formatTime(duration) }}
    </div>
  </div>
</template>

<script setup>
import { ref, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'

const emit = defineEmits(['record-complete'])

const isRecording = ref(false)
const duration = ref(0)
const timer = ref(null)
let mediaRecorder = null
let audioChunks = []

const startRecording = async () => {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    mediaRecorder = new MediaRecorder(stream)
    audioChunks = []

    mediaRecorder.ondataavailable = (event) => {
      audioChunks.push(event.data)
    }

    mediaRecorder.onstop = () => {
      const audioBlob = new Blob(audioChunks, { type: 'audio/webm' })
      const reader = new FileReader()
      reader.readAsDataURL(audioBlob)
      reader.onloadend = () => {
        const base64Audio = reader.result.split(',')[1]
        emit('record-complete', base64Audio)
      }
    }

    mediaRecorder.start()
    isRecording.value = true
    duration.value = 0
    timer.value = setInterval(() => {
      duration.value++
    }, 1000)
  } catch (error) {
    ElMessage.error('无法访问麦克风')
    console.error(error)
  }
}

const stopRecording = () => {
  if (mediaRecorder && isRecording.value) {
    mediaRecorder.stop()
    isRecording.value = false
    clearInterval(timer.value)
    
    // 停止所有音频轨道
    mediaRecorder.stream.getTracks().forEach(track => track.stop())
  }
}

const toggleRecording = () => {
  if (isRecording.value) {
    stopRecording()
  } else {
    startRecording()
  }
}

const formatTime = (seconds) => {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

onUnmounted(() => {
  if (isRecording.value) {
    stopRecording()
  }
})
</script>

<style scoped>
.audio-recorder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.recording {
  animation: pulse 1.5s infinite;
  background-color: #f56c6c;
  border-color: #f56c6c;
}

.recording-status {
  font-size: 14px;
  color: #f56c6c;
}

@keyframes pulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.1); }
  100% { transform: scale(1); }
}
</style>