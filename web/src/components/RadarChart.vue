<template>
  <div class="radar-chart">
    <Radar :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup>
import { Radar } from 'vue-chartjs'
import { Chart as ChartJS, RadialLinearScale, PointElement, LineElement, Filler, Tooltip, Legend } from 'chart.js'
import { computed } from 'vue'

ChartJS.register(RadialLinearScale, PointElement, LineElement, Filler, Tooltip, Legend)

const props = defineProps({
  data: {
    type: Object,
    required: true
  }
})

const chartData = computed(() => ({
  labels: Object.keys(props.data),
  datasets: [
    {
      label: '能力评估',
      backgroundColor: 'rgba(54, 162, 235, 0.2)',
      borderColor: 'rgba(54, 162, 235, 1)',
      pointBackgroundColor: 'rgba(54, 162, 235, 1)',
      pointBorderColor: '#fff',
      pointHoverBackgroundColor: '#fff',
      pointHoverBorderColor: 'rgba(54, 162, 235, 1)',
      data: Object.values(props.data)
    }
  ]
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    r: {
      angleLines: {
        display: false
      },
      suggestedMin: 0,
      suggestedMax: 100
    }
  }
}
</script>

<style scoped>
.radar-chart {
  width: 100%;
  height: 300px;
}
</style>