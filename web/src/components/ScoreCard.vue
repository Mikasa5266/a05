<template>
  <el-card class="score-card">
    <div class="score-header">
      <h3>{{ title }}</h3>
      <div class="score-value">{{ displayScore }}</div>
    </div>
    <div class="score-body">
      <el-progress :percentage="safeScore" :status="progressStatus" />
      <p class="feedback">{{ feedback }}</p>
    </div>
  </el-card>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  title: String,
  score: Number,
  feedback: String
})

const safeScore = computed(() => {
  const value = Number(props.score)
  if (!Number.isFinite(value)) return 0
  if (value < 0) return 0
  if (value > 100) return 100
  return value
})

const displayScore = computed(() => {
  return Number.isFinite(Number(props.score)) ? Math.round(Number(props.score)) : '--'
})

const progressStatus = computed(() => {
  if (safeScore.value >= 80) return 'success'
  if (safeScore.value >= 60) return 'warning'
  return 'exception'
})
</script>

<style scoped>
.score-card {
  margin-bottom: 20px;
}

.score-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.score-value {
  font-size: 24px;
  font-weight: bold;
}

.feedback {
  margin-top: 10px;
  color: #666;
  font-size: 14px;
}
</style>