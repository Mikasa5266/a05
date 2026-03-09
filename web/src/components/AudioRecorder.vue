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
let recorderMimeType = ''

const pickSupportedAudioMime = () => {
  const candidates = [
    'audio/webm;codecs=opus',
    'audio/webm',
    'audio/mp4',
    'audio/ogg;codecs=opus',
    'audio/ogg'
  ]
  if (typeof MediaRecorder === 'undefined' || typeof MediaRecorder.isTypeSupported !== 'function') {
    return ''
  }
  for (const mime of candidates) {
    if (MediaRecorder.isTypeSupported(mime)) return mime
  }
  return ''
}

const normalizeAudioMime = (mime) => {
  const raw = String(mime || '').trim().toLowerCase()
  if (!raw) return ''
  const semi = raw.indexOf(';')
  return semi > 0 ? raw.slice(0, semi) : raw
}

const startRecording = async () => {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    const preferredMime = pickSupportedAudioMime()
    mediaRecorder = preferredMime ? new MediaRecorder(stream, { mimeType: preferredMime }) : new MediaRecorder(stream)
    recorderMimeType = normalizeAudioMime(mediaRecorder.mimeType || preferredMime)
    audioChunks = []

    mediaRecorder.ondataavailable = (event) => {
      audioChunks.push(event.data)
    }

    mediaRecorder.onstop = () => {
      const audioBlob = new Blob(audioChunks, { type: recorderMimeType || 'audio/webm' })
      const reader = new FileReader()
      reader.readAsDataURL(audioBlob)
      reader.onloadend = () => {
        const raw = String(reader.result || '')
        const matched = raw.match(/^data:([^;]+);base64,(.+)$/)
        if (matched && matched[2]) {
          emit('record-complete', { base64: matched[2], mime: normalizeAudioMime(matched[1]) })
          return
        }
        const parts = raw.split(',')
        if (parts.length < 2) return
        emit('record-complete', { base64: parts[1], mime: recorderMimeType || '' })
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
