<template>
  <div ref="chartRef" class="glass-echart" :style="{ height }"></div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { use, init } from 'echarts/core'
import { GaugeChart, BarChart, LineChart, RadarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, RadarComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([GaugeChart, BarChart, LineChart, RadarChart, GridComponent, TooltipComponent, RadarComponent, CanvasRenderer])

const props = defineProps({
  option: {
    type: Object,
    default: () => ({})
  },
  height: {
    type: String,
    default: '280px'
  }
})

const chartRef = ref(null)
let chartInstance = null

const ensureChart = () => {
  if (!chartRef.value) return
  if (!chartInstance) {
    chartInstance = init(chartRef.value)
  }
  chartInstance.setOption(props.option || {}, true)
}

const resizeChart = () => {
  chartInstance?.resize()
}

watch(
  () => props.option,
  async () => {
    await nextTick()
    ensureChart()
  },
  { deep: true, immediate: true }
)

onMounted(() => {
  window.addEventListener('resize', resizeChart)
  nextTick(() => ensureChart())
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeChart)
  if (chartInstance) {
    chartInstance.dispose()
    chartInstance = null
  }
})
</script>

<style scoped>
.glass-echart {
  width: 100%;
}
</style>
