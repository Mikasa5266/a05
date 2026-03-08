<template>
  <div class="radar-chart">
    <Radar :data="computedChartData" :options="computedChartOptions" />
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

const computedChartData = computed(() => ({
  labels: Object.keys(props.data),
  datasets: [
    {
      label: '能力评估',
      backgroundColor: 'rgba(99, 102, 241, 0.2)',
      borderColor: '#6366f1',
      borderWidth: 2,
      pointBackgroundColor: '#6366f1',
      pointBorderColor: '#fff',
      pointHoverBackgroundColor: '#fff',
      pointHoverBorderColor: '#6366f1',
      data: Object.values(props.data)
    }
  ]
}))

const computedChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    r: {
      angleLines: {
        color: '#e5e7eb'
      },
      grid: {
        color: '#e5e7eb'
      },
      pointLabels: {
        color: '#4b5563',
        font: {
          size: 12,
          family: "'Inter', sans-serif"
        }
      },
      ticks: {
        display: false,
        stepSize: 20
      },
      suggestedMin: 0,
      suggestedMax: 100
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
.radar-chart {
  width: 100%;
  height: 300px;
}
</style>