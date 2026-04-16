<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import LiveInterviewRoomGroup from './LiveInterviewRoomGroup.vue'
import LiveInterviewRoomOneOnOne from './LiveInterviewRoomOneOnOne.vue'

const route = useRoute()

const isGroupMode = computed(() => {
  if (String(route.query?.group_mode || '').trim() === '1') return true
  return String(route.path || '').includes('/group/')
})
</script>

<template>
  <LiveInterviewRoomGroup v-if="isGroupMode" />
  <LiveInterviewRoomOneOnOne v-else />
</template>
