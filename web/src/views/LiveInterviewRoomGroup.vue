<script setup>
import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import LiveInterviewRoom from './LiveInterviewRoom.vue'

const route = useRoute()
const router = useRouter()

const invitationId = computed(() => String(route.params?.id || '').trim())

const fallbackPath = computed(() => {
  const currentPath = String(route.path || '')
  if (currentPath.startsWith('/enterprise/')) return '/enterprise/interview-workbench'
  if (currentPath.startsWith('/university/')) return '/university/interview-workbench'
  return '/interview/live/workbench'
})

watch(
  () => [route.params?.id, route.query?.invitation_id, route.query?.group_mode],
  () => {
    const id = invitationId.value
    if (!id) {
      router.replace(fallbackPath.value)
      return
    }

    const currentInvitationId = String(route.query?.invitation_id || '').trim()
    const currentGroupMode = String(route.query?.group_mode || '').trim()
    if (currentInvitationId === id && currentGroupMode === '1') {
      return
    }

    router.replace({
      path: route.path,
      query: {
        ...route.query,
        invitation_id: id,
        group_mode: '1'
      }
    })
  },
  { immediate: true }
)
</script>

<template>
  <LiveInterviewRoom />
</template>
