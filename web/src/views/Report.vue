<template>
  <div class="report-container">
    <el-card class="report-header">
      <div class="header-info">
        <h2>面试报告 - {{ report.position }}</h2>
        <el-tag type="info">{{ report.difficulty }}</el-tag>
      </div>
      <div class="score-summary">
        <ScoreCard 
          title="综合得分" 
          :score="report.average_score" 
          :feedback="report.overall_analysis" 
        />
      </div>
    </el-card>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="12">
        <el-card header="能力雷达图">
          <RadarChart :data="abilityScores" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card header="成长曲线">
          <GrowthCurve :data="historyData" />
        </el-card>
      </el-col>
    </el-row>

    <el-card header="详细评估" class="detail-card">
      <el-collapse v-model="activeNames">
        <el-collapse-item title="优势分析" name="1">
          <ul>
            <li v-for="(strength, index) in report.strengths" :key="index">{{ strength }}</li>
          </ul>
        </el-collapse-item>
        <el-collapse-item title="待改进" name="2">
          <ul>
            <li v-for="(weakness, index) in report.weaknesses" :key="index">{{ weakness }}</li>
          </ul>
        </el-collapse-item>
        <el-collapse-item title="建议" name="3">
          <ul>
            <li v-for="(suggestion, index) in report.suggestions" :key="index">{{ suggestion }}</li>
          </ul>
        </el-collapse-item>
      </el-collapse>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { getReport, generateReport } from '../api/report'
import ScoreCard from '../components/ScoreCard.vue'
import RadarChart from '../components/RadarChart.vue'
import GrowthCurve from '../components/GrowthCurve.vue'

const route = useRoute()
const report = ref({})
const activeNames = ref(['1', '2', '3'])
const historyData = ref([]) // 这里可以从后端获取历史数据

const abilityScores = computed(() => {
  const safe = (value) => Number.isFinite(Number(value)) ? Number(value) : 0
  return {
    '技术能力': safe(report.value.technical_score),
    '表达沟通': safe(report.value.expression_score),
    '逻辑思维': safe(report.value.logic_score),
    '岗位匹配': safe(report.value.matching_score),
    '职业素养': safe(report.value.behavior_score)
  }
})

onMounted(async () => {
  const id = route.params.id
  if (id) {
    try {
    let res
    try {
      res = await getReport(id)
    } catch (_) {
      const generated = await generateReport({ interview_id: Number(id) })
      if (generated?.report?.id) {
        res = await getReport(generated.report.id)
      }
    }

    report.value = res?.report || {}
    const reportDate = (report.value.created_at || report.value.end_time || '').toString().slice(0, 10) || '当前'
    historyData.value = [{ date: reportDate, score: report.value.average_score || 0 }]
    } catch (error) {
      console.error('获取报告失败', error)
    }
  }
})
</script>

<style scoped>
.report-container {
  padding: 20px;
}

.report-header {
  margin-bottom: 20px;
}

.header-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.chart-row {
  margin-bottom: 20px;
}

.detail-card ul {
  padding-left: 20px;
}

.detail-card li {
  margin-bottom: 10px;
}
</style>