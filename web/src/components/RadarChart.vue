<template>
  <div class="radar-chart">
    <Radar :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup>
import { Radar } from 'vue-chartjs'
import { Chart as ChartJS, RadialLinearScale, PointElement, LineElement, Filler, Tooltip, Legend } from 'chart.js'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

ChartJS.register(RadialLinearScale, PointElement, LineElement, Filler, Tooltip, Legend)

const props = defineProps({
  data: {
    type: Object,
    required: true
  }
})

const isMobile = ref(false)
let mobileMediaQuery = null

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

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: true,
      position: isMobile.value ? 'bottom' : 'top',
      labels: {
        boxWidth: isMobile.value ? 10 : 14,
        boxHeight: isMobile.value ? 10 : 14,
        padding: isMobile.value ? 10 : 16,
        font: {
          size: isMobile.value ? 10 : 12,
        },
      },
    },
  },
  scales: {
    r: {
      angleLines: {
        display: false
      },
      pointLabels: {
        color: '#71717a',
        font: {
          size: isMobile.value ? 10 : 12,
        },
      },
      ticks: {
        display: !isMobile.value,
        backdropColor: 'transparent',
      },
      suggestedMin: 0,
      suggestedMax: 100
    }
  }
}))

const syncMobileState = () => {
  if (!mobileMediaQuery) return
  isMobile.value = mobileMediaQuery.matches
}

onMounted(() => {
  mobileMediaQuery = window.matchMedia('(max-width: 767px)')
  syncMobileState()
  if (mobileMediaQuery.addEventListener) {
    mobileMediaQuery.addEventListener('change', syncMobileState)
  } else {
    mobileMediaQuery.addListener(syncMobileState)
  }
})

onBeforeUnmount(() => {
  if (!mobileMediaQuery) return
  if (mobileMediaQuery.removeEventListener) {
    mobileMediaQuery.removeEventListener('change', syncMobileState)
  } else {
    mobileMediaQuery.removeListener(syncMobileState)
  }
})
</script>

<style scoped>
.radar-chart {
  width: 100%;
  height: 300px;
}
</style>