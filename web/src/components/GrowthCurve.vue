<template>
  <div class="growth-curve">
    <Line :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup>
import { Line } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend } from 'chart.js'
import { computed } from 'vue'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend)

const props = defineProps({
  data: {
    type: Array,
    required: true
  }
})

const normalizedData = computed(() => {
  const list = (props.data || [])
    .map((item) => {
      const rawScore = Number(item?.score)
      return {
        date: item?.date || '当前',
        score: Number.isFinite(rawScore) ? Math.max(0, Math.min(100, rawScore)) : null
      }
    })
    .filter((item) => item.score !== null)

  if (list.length === 0) {
    return [
      { date: '上次', score: 60 },
      { date: '当前', score: 68 }
    ]
  }

  if (list.length === 1) {
    return [
      { date: '上次', score: list[0].score },
      { date: list[0].date || '当前', score: list[0].score }
    ]
  }

  return list
})

const yRange = computed(() => {
  const scores = normalizedData.value.map((item) => Number(item.score))
  const minScore = Math.min(...scores)
  const maxScore = Math.max(...scores)

  let min = Math.max(0, Math.floor((minScore - 10) / 5) * 5)
  let max = Math.min(100, Math.ceil((maxScore + 10) / 5) * 5)

  if (max - min < 20) {
    min = Math.max(0, min - 10)
    max = Math.min(100, max + 10)
  }

  return { min, max }
})

const chartData = computed(() => ({
  labels: normalizedData.value.map(item => item.date),
  datasets: [
    {
      label: '综合得分',
      backgroundColor: 'rgba(93, 99, 242, 0.15)',
      borderColor: '#5d63f2',
      borderWidth: 3,
      tension: 0.28,
      pointRadius: 5,
      pointHoverRadius: 6,
      pointBackgroundColor: '#5d63f2',
      pointBorderColor: '#fff',
      pointBorderWidth: 2,
      fill: false,
      data: normalizedData.value.map(item => item.score)
    }
  ]
}))

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    x: {
      grid: {
        display: false
      },
      ticks: {
        display: false
      }
    },
    y: {
      min: yRange.value.min,
      max: yRange.value.max,
      ticks: {
        color: '#8f93a5',
        stepSize: 10
      },
      grid: {
        color: '#d9dce3',
        borderDash: [4, 4]
      }
    }
  },
  plugins: {
    legend: {
      display: false
    }
  }
}))
</script>

<style scoped>
.growth-curve {
  width: 100%;
  height: 300px;
}
</style>