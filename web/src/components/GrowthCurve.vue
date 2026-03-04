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

const chartData = computed(() => ({
  labels: props.data.map(item => item.date),
  datasets: [
    {
      label: '综合得分',
      backgroundColor: '#f87979',
      borderColor: '#f87979',
      data: props.data.map(item => item.score)
    }
  ]
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    y: {
      beginAtZero: true,
      max: 100
    }
  }
}
</script>

<style scoped>
.growth-curve {
  width: 100%;
  height: 300px;
}
</style>