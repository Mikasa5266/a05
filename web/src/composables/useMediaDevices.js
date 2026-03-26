import { ref } from 'vue'

export function useMediaDevices(options = {}) {
  const isCameraOn = ref(true)
  const isMicOn = ref(true)
  const stream = ref(null)
  let cameraStartToken = 0

  const getAdditionalStreams = typeof options.getAdditionalStreams === 'function'
    ? options.getAdditionalStreams
    : () => []
  const onCameraStopped = typeof options.onCameraStopped === 'function'
    ? options.onCameraStopped
    : () => {}

  const stopMediaTracks = (targetStream) => {
    if (!targetStream) return
    targetStream.getTracks().forEach((track) => track.stop())
  }

  const applyMicEnabledState = (enabled) => {
    const streams = [stream.value, ...getAdditionalStreams()].filter(Boolean)
    streams.forEach((targetStream) => {
      targetStream.getAudioTracks().forEach((track) => {
        track.enabled = enabled
      })
    })
  }

  const toggleMic = () => {
    isMicOn.value = !isMicOn.value
    applyMicEnabledState(isMicOn.value)
  }

  const startCamera = async () => {
    const token = ++cameraStartToken
    try {
      const mediaStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
      if (token !== cameraStartToken) {
        stopMediaTracks(mediaStream)
        return null
      }
      stopMediaTracks(stream.value)
      stream.value = mediaStream
      applyMicEnabledState(isMicOn.value)
      isCameraOn.value = true
      return mediaStream
    } catch (err) {
      // Some browsers may throw transient errors when camera is reopened immediately after stop.
      if ((err?.name === 'NotReadableError' || err?.name === 'AbortError') && token === cameraStartToken) {
        try {
          await new Promise((resolve) => setTimeout(resolve, 220))
          const retryStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
          if (token !== cameraStartToken) {
            stopMediaTracks(retryStream)
            return null
          }
          stopMediaTracks(stream.value)
          stream.value = retryStream
          applyMicEnabledState(isMicOn.value)
          isCameraOn.value = true
          return retryStream
        } catch (retryErr) {
          console.error('Camera access denied:', retryErr)
          isCameraOn.value = false
          return null
        }
      }
      console.error('Camera access denied:', err)
      isCameraOn.value = false
      return null
    }
  }

  const stopCamera = () => {
    cameraStartToken += 1
    stopMediaTracks(stream.value)
    stream.value = null
    isCameraOn.value = false
    onCameraStopped()
  }

  const toggleCamera = async () => {
    if (isCameraOn.value) {
      stopCamera()
      return
    }
    await startCamera()
  }

  return {
    isCameraOn,
    isMicOn,
    stream,
    toggleCamera,
    toggleMic,
    startCamera,
    stopCamera,
    stopMediaTracks
  }
}
