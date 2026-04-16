<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import LiveInterviewRoomGroup from './LiveInterviewRoomGroup.vue'
import LiveInterviewRoomOneOnOne from './LiveInterviewRoomOneOnOne.vue'

const route = useRoute()

const isGroupMode = computed(() => {
  return String(route.path || '').includes('/group/')
})
</script>

<template>
  <LiveInterviewRoomGroup v-if="isGroupMode" />
  <LiveInterviewRoomOneOnOne v-else />
</template>
