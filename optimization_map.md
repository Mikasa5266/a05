# Optimization Map

生成时间：2026-04-13

## 1) 标记扫描汇总

- LEGACY-TRASH：3 处
- FUTURE-LINK：1 处
- DEPRECATED：0 处（本轮扫描未发现）

## 2) 垃圾代码与兼容保留位

### LEGACY-TRASH 清单

1. 文件：server/internal/model/interview.go:21  
   标记：TODO: [LEGACY-TRASH] - Group Interview Refactor  
   说明：GroupInterviewRoom 里的 ParticipantIDs 为兼容阶段保留的 JSON 字符串方案。

2. 文件：server/internal/model/interview.go:86  
   标记：TODO: [LEGACY-TRASH] - Group Interview Refactor  
   说明：Interview 主模型群面兼容字段改造入口（后续与房间实体进一步解耦）。

3. 文件：server/internal/model/report.go:34  
   标记：TODO: [LEGACY-TRASH] - Group Interview Refactor  
   说明：Report 结构在单人/群面并存期的兼容保留点。

### DEPRECATED 清单

- 未扫描到 DEPRECATED 标记。
- 说明：第四阶段未对 ScoreCard.vue 做魔改，因此没有出现 DEPRECATED-STYLING-KEEP-FOR-STABILITY 标记。

## 3) FUTURE-LINK 与真实群面评分接入点

### FUTURE-LINK 清单

1. 文件：server/internal/service/ai/ai_evaluation_service.go:31  
   标记：FUTURE-LINK: RAG_EVAL_INTEGRATION

### 真实群面评分建议接入函数

- 首要接入函数：AIService.EvaluateAnswer  
  文件位置：server/internal/service/ai/ai_evaluation_service.go:43

- 当前群面分支：evaluateGroupAnswerWithMockTemplate  
  文件位置：server/internal/service/ai/ai_evaluation_service.go:106

- 预留接口：RealRAGAnalysis.AnalyzeGroupAssessment  
  文件位置：server/internal/service/ai/ai_evaluation_service.go:32

### 建议替换路径

1. 在 EvaluateAnswer 的群面判断分支中，优先调用 RealRAGAnalysis.AnalyzeGroupAssessment。  
2. 当 RAG 返回失败或超时时，再回退到 evaluateGroupAnswerWithMockTemplate。  
3. 产出的结构继续沿用 EvaluationResult，确保前端协议不破坏。

## 4) 图表组件完整性检查（ECharts / Chart.js）

### ECharts（本轮重构主力）

- 核心组件：web/src/components/report/GlassEchart.vue  
  说明：已使用 echarts/core 按需注册 Gauge/Bar/Line/Radar 与 Grid/Tooltip/Radar 组件及 CanvasRenderer。

- 报告页接入：web/src/views/Report.vue  
  使用项：综合得分仪表、维度柱图、能力雷达图、心理波动预留曲线、趋势曲线。

- 历史详情弹窗接入：web/src/views/History.vue  
  使用项：能力雷达图、心理波动预留曲线。

### Chart.js（保留且正常）

- 组件保留：web/src/components/RadarChart.vue  
- 组件保留：web/src/components/GrowthCurve.vue  
- 其它使用点：web/src/views/GrowthCenter.vue  
- 其它使用点：web/src/components/resume/ResumeAnalysisDashboard.vue

### 构建验证

- 已执行：web 目录 npm run build  
- 结果：构建通过，图表相关产物正常输出（含 vendor-charts 与 GlassEchart 产物）。
- 备注：GlassEchart 产物仍有大 chunk 警告（可运行，不阻塞功能）。

## 5) 一键清理检索建议

- 全量标记扫描：
  rg -n "LEGACY-TRASH|DEPRECATED|FUTURE-LINK" server web

- 仅查群面评分接入点：
  rg -n "RealRAGAnalysis|AnalyzeGroupAssessment|evaluateGroupAnswerWithMockTemplate|EvaluateAnswer" server/internal/service/ai

- 仅查图表接入点：
  rg -n "GlassEchart|vue-chartjs|chart.js|echarts/core" web/src
