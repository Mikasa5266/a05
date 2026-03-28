<script setup>
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import MockInterview from './MockInterview.vue'

const route = useRoute()
const router = useRouter()

const normalizedVideoQuery = computed(() => ({
  ...route.query,
  mode: String(route.query?.mode || 'technical'),
  style: String(route.query?.style || 'gentle'),
  interviewMode: String(route.query?.interviewMode || 'ai'),
  presentationMode: String(route.query?.presentationMode || 'video_avatar')
}))

const syncVideoModeQuery = async () => {
  const currentMode = String(route.query?.mode || '')
  const currentStyle = String(route.query?.style || '')
  const currentInterviewMode = String(route.query?.interviewMode || '')
  const currentPresentationMode = String(route.query?.presentationMode || '')

  const nextQuery = normalizedVideoQuery.value

  const isSynced =
    currentMode === nextQuery.mode &&
    currentStyle === nextQuery.style &&
    currentInterviewMode === nextQuery.interviewMode &&
    currentPresentationMode === nextQuery.presentationMode

  if (isSynced) return

  await router.replace({
    path: route.path,
    query: nextQuery
  })
}

onMounted(() => {
  syncVideoModeQuery()
})
</script>

<template>
  <MockInterview />
</template>
