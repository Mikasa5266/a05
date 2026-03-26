import { onUnmounted, ref } from 'vue'
import { uploadInterviewRecording as apiUploadInterviewRecording } from '../api/interview'

export function useInterviewRecording(options = {}) {
  const interviewId = options.interviewId
  const settings = options.settings
  const stream = options.stream
  const isMicOn = options.isMicOn

  const recordingStatus = ref('idle')
  const recordingUrl = ref('')

  let interviewMediaRecorder = null
  let interviewRecordedChunks = []
  let interviewRecordingStream = null

  const stopInterviewRecordingStream = () => {
    if (!interviewRecordingStream) return
    if (stream?.value && interviewRecordingStream === stream.value) return

    interviewRecordingStream.getTracks().forEach((track) => track.stop())
    interviewRecordingStream = null
  }

  const startInterviewRecording = async () => {
    if (!interviewId?.value || interviewMediaRecorder) return

    let targetStream = null
    if (settings?.value?.presentationMode === 'video_avatar' && stream?.value) {
      targetStream = stream.value
    } else {
      try {
        // Fallback to audio-only recording so replay can still be generated.
        targetStream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false })
        targetStream.getAudioTracks().forEach((track) => {
          track.enabled = !!isMicOn?.value
        })
      } catch (err) {
        console.warn('无法获取回放录制流:', err)
        recordingStatus.value = 'failed'
        return
      }
    }

    if (!targetStream) {
      recordingStatus.value = 'failed'
      return
    }

    interviewRecordingStream = targetStream

    try {
      interviewRecordedChunks = []
      recordingStatus.value = 'recording'
      interviewMediaRecorder = new MediaRecorder(targetStream, { mimeType: 'video/webm;codecs=vp8,opus' })
    } catch (_) {
      try {
        interviewMediaRecorder = new MediaRecorder(targetStream)
        recordingStatus.value = 'recording'
        interviewRecordedChunks = []
      } catch (err) {
        console.warn('无法创建视频录制器:', err)
        recordingStatus.value = 'failed'
        stopInterviewRecordingStream()
        return
      }
    }

    interviewMediaRecorder.ondataavailable = (event) => {
      if (event.data && event.data.size > 0) {
        interviewRecordedChunks.push(event.data)
      }
    }

    interviewMediaRecorder.onerror = (err) => {
      console.warn('回放录制异常:', err)
      recordingStatus.value = 'failed'
    }

    interviewMediaRecorder.start(1000)
  }

  const stopAndUploadInterviewRecording = async () => {
    if (!interviewId?.value) return false

    if (!interviewMediaRecorder) {
      recordingStatus.value = recordingStatus.value === 'uploaded' ? 'uploaded' : 'failed'
      return false
    }

    if (interviewMediaRecorder.state === 'recording') {
      await new Promise((resolve) => {
        interviewMediaRecorder.onstop = resolve
        try {
          interviewMediaRecorder.stop()
        } catch (_) {
          resolve()
        }
      })
    }

    if (!interviewRecordedChunks.length) {
      recordingStatus.value = 'failed'
      interviewMediaRecorder = null
      stopInterviewRecordingStream()
      return false
    }

    const blob = new Blob(interviewRecordedChunks, { type: 'video/webm' })
    const formData = new FormData()
    formData.append('recording', blob, `interview_${interviewId.value}.webm`)

    try {
      const res = await apiUploadInterviewRecording(interviewId.value, formData)
      recordingUrl.value = res.recording_url || ''
      recordingStatus.value = 'uploaded'
      return true
    } catch (err) {
      console.warn('视频上传失败:', err)
      recordingStatus.value = 'failed'
      return false
    } finally {
      interviewMediaRecorder = null
      interviewRecordedChunks = []
      stopInterviewRecordingStream()
    }
  }

  const cleanupInterviewRecording = () => {
    if (interviewMediaRecorder && interviewMediaRecorder.state === 'recording') {
      try {
        interviewMediaRecorder.stop()
      } catch {
        // Ignore stop errors during teardown.
      }
    }
    interviewMediaRecorder = null
    interviewRecordedChunks = []
    stopInterviewRecordingStream()
  }

  onUnmounted(() => {
    cleanupInterviewRecording()
  })

  return {
    recordingStatus,
    recordingUrl,
    startInterviewRecording,
    stopAndUploadInterviewRecording,
    stopInterviewRecordingStream,
    cleanupInterviewRecording
  }
}
