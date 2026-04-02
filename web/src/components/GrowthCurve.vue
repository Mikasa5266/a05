<template>
  <div class="growth-curve">
    <Line :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup>
import { Line } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend } from 'chart.js'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend)

const props = defineProps({
  data: {
    type: Array,
    required: true
  }
})

const isMobile = ref(false)
let mobileMediaQuery = null

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

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: true,
      position: isMobile.value ? 'bottom' : 'top',
      labels: {
        usePointStyle: true,
        pointStyle: 'circle',
        boxWidth: isMobile.value ? 8 : 12,
        boxHeight: isMobile.value ? 8 : 12,
        padding: isMobile.value ? 10 : 16,
        font: {
          size: isMobile.value ? 10 : 12,
        },
      },
    },
  },
  scales: {
    x: {
      ticks: {
        autoSkip: true,
        maxTicksLimit: isMobile.value ? 4 : 8,
        maxRotation: 0,
        minRotation: 0,
        color: '#71717a',
        font: {
          size: isMobile.value ? 10 : 12,
        },
      },
      grid: {
        display: !isMobile.value,
      },
    },
    y: {
      beginAtZero: true,
      max: 100,
      ticks: {
        color: '#71717a',
        font: {
          size: isMobile.value ? 10 : 12,
        },
      },
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
.growth-curve {
  width: 100%;
  height: 300px;
}
</style>