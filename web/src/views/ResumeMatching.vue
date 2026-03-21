<script setup>
import { ref } from 'vue'
import { Upload, Sparkles, BrainCircuit, Search, CheckCircle2 } from 'lucide-vue-next'
import {
  parseResume,
  analyzeResumeAuthenticity,
  getResumeOptimizationSuggestions,
  generateResumeTemplate,
  aiChatFallback
} from '../api/resume'
import { useRouter } from 'vue-router'

const router = useRouter()
const fileInput = ref(null)
const isUploading = ref(false)
const resumeData = ref(null)
const matches = ref([])
const targetRole = ref('')
const rawResumeText = ref('')
const templateSeniority = ref('3-5年')

const authenticityLoading = ref(false)
const optimizationLoading = ref(false)
const templateLoading = ref(false)

const authenticityReport = ref(null)
const optimizationReport = ref(null)
const resumeTemplate = ref(null)
const templateCopied = ref(false)
const pdfFontSize = ref(12)
const pdfLineHeight = ref(1.6)
const pdfMarginMm = ref(14)
const pdfKeepSections = ref(true)
const pdfIncludeGuides = ref(true)
const pdfIncludeMistakes = ref(true)
const activePdfPreset = ref('experienced')

const pdfPresets = {
  campus: {
    label: '校招简洁版',
    fontSize: 11.5,
    lineHeight: 1.55,
    marginMm: 13,
    keepSections: true,
    includeGuides: true,
    includeMistakes: false
  },
  experienced: {
    label: '社招平衡版',
    fontSize: 12,
    lineHeight: 1.6,
    marginMm: 14,
    keepSections: true,
    includeGuides: true,
    includeMistakes: true
  },
  technical: {
    label: '技术细节版',
    fontSize: 11,
    lineHeight: 1.5,
    marginMm: 12,
    keepSections: true,
    includeGuides: false,
    includeMistakes: false
  }
}

const applyPdfPreset = (presetKey) => {
  const preset = pdfPresets[presetKey]
  if (!preset) return
  activePdfPreset.value = presetKey
  pdfFontSize.value = preset.fontSize
  pdfLineHeight.value = preset.lineHeight
  pdfMarginMm.value = preset.marginMm
  pdfKeepSections.value = preset.keepSections
  pdfIncludeGuides.value = preset.includeGuides
  pdfIncludeMistakes.value = preset.includeMistakes
}

const triggerFileInput = () => {
  fileInput.value.click()
}

const handleFileChange = (event) => {
  const file = event.target.files[0]
  if (file) processFile(file)
}

const handleDrop = (event) => {
  const file = event.dataTransfer.files[0]
  if (file) processFile(file)
}

const processFile = async (file) => {
  isUploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    
    const res = await parseResume(formData)
    
    // Backend returns { resume: {...}, matches: [...] }
    resumeData.value = res.resume
    matches.value = res.matches || []
    targetRole.value = res.resume?.intent || ''
    
  } catch (error) {
    console.error('Failed to parse resume:', error)
    alert('简历解析失败: ' + (error.response?.data?.error || error.message))
  } finally {
    isUploading.value = false
  }
}

const resetUpload = () => {
  resumeData.value = null
  matches.value = []
  authenticityReport.value = null
  optimizationReport.value = null
  resumeTemplate.value = null
  rawResumeText.value = ''
  targetRole.value = ''
  if (fileInput.value) fileInput.value.value = ''
}

const startInterview = (match) => {
  router.push({
    name: 'MockInterview',
    query: { position: match.jobTitle }
  })
}

const runAuthenticity = async () => {
  if (!resumeData.value) return
  authenticityLoading.value = true
  try {
    const res = await analyzeResumeAuthenticity({
      resumeData: resumeData.value,
      rawText: rawResumeText.value,
      targetRole: targetRole.value
    })
    authenticityReport.value = res.report
  } catch (error) {
    try {
      if (error?.response?.status !== 404) throw error
      const fallbackPrompt = `请根据给定简历信息做真实性风险核验，仅返回JSON：
{
  "overallRiskScore": 0,
  "summary": "",
  "verifiableItems": [],
  "potentialRiskItems": [{"claim":"","riskLevel":"low|medium|high","reason":"","verificationTip":""}],
  "interviewChecks": [],
  "disclaimer": ""
}
要求：不能断言造假，只能给潜在风险与核验建议。
目标岗位：${targetRole.value || '未指定'}
简历数据：${JSON.stringify(resumeData.value)}
原文：${rawResumeText.value || '无'}
`
      const legacy = await aiChatFallback({ message: fallbackPrompt, context: 'resume-authenticity-fallback' })
      authenticityReport.value = parseJSONFromAIText(legacy.response)
      alert('已切换兼容模式：当前后端缺少 /resume/authenticity，使用 /ai/chat 完成分析。')
    } catch (fallbackError) {
      console.error('Failed to analyze authenticity:', fallbackError)
      alert('真实性分析失败: ' + (fallbackError.response?.data?.error || fallbackError.message))
    }
  } finally {
    authenticityLoading.value = false
  }
}

const runOptimization = async () => {
  if (!resumeData.value) return
  optimizationLoading.value = true
  try {
    const res = await getResumeOptimizationSuggestions({
      resumeData: resumeData.value,
      targetRole: targetRole.value
    })
    optimizationReport.value = res.report
  } catch (error) {
    try {
      if (error?.response?.status !== 404) throw error
      const fallbackPrompt = `请输出简历优化建议，仅返回JSON：
{
  "overallScore": 0,
  "strengths": [],
  "weaknesses": [],
  "suggestions": [],
  "rewriteDemo": [],
  "keywords": []
}
目标岗位：${targetRole.value || '未指定'}
简历数据：${JSON.stringify(resumeData.value)}
`
      const legacy = await aiChatFallback({ message: fallbackPrompt, context: 'resume-optimize-fallback' })
      optimizationReport.value = parseJSONFromAIText(legacy.response)
      alert('已切换兼容模式：当前后端缺少 /resume/optimize，使用 /ai/chat 完成分析。')
    } catch (fallbackError) {
      console.error('Failed to get optimization suggestions:', fallbackError)
      alert('简历优化失败: ' + (fallbackError.response?.data?.error || fallbackError.message))
    }
  } finally {
    optimizationLoading.value = false
  }
}

const runTemplate = async () => {
  templateLoading.value = true
  try {
    const res = await generateResumeTemplate({
      targetRole: targetRole.value || resumeData.value?.intent || '后端开发工程师',
      seniority: templateSeniority.value,
      language: 'zh-CN'
    })
    resumeTemplate.value = res.template
  } catch (error) {
    try {
      if (error?.response?.status !== 404) throw error
      const fallbackPrompt = `请生成简历范本，仅返回JSON：
{
  "targetRole": "",
  "templateMarkdown": "",
  "writingGuides": [],
  "commonMistakes": []
}
目标岗位：${targetRole.value || resumeData.value?.intent || '后端开发工程师'}
经验级别：${templateSeniority.value}
输出语言：zh-CN`
      const legacy = await aiChatFallback({ message: fallbackPrompt, context: 'resume-template-fallback' })
      resumeTemplate.value = parseJSONFromAIText(legacy.response)
      alert('已切换兼容模式：当前后端缺少 /resume/template，使用 /ai/chat 生成范本。')
    } catch (fallbackError) {
      console.error('Failed to generate resume template:', fallbackError)
      alert('简历范本生成失败: ' + (fallbackError.response?.data?.error || fallbackError.message))
    }
  } finally {
    templateLoading.value = false
  }
}

const parseJSONFromAIText = (text) => {
  const source = String(text || '').trim()
  const firstBrace = source.indexOf('{')
  const firstBracket = source.indexOf('[')
  let start = -1
  if (firstBrace === -1) start = firstBracket
  else if (firstBracket === -1) start = firstBrace
  else start = Math.min(firstBrace, firstBracket)

  const lastBrace = source.lastIndexOf('}')
  const lastBracket = source.lastIndexOf(']')
  const end = Math.max(lastBrace, lastBracket)
  if (start !== -1 && end > start) {
    return JSON.parse(source.slice(start, end + 1))
  }
  return JSON.parse(source)
}

const copyTemplate = async () => {
  if (!resumeTemplate.value?.templateMarkdown) return
  try {
    await navigator.clipboard.writeText(resumeTemplate.value.templateMarkdown)
    templateCopied.value = true
    setTimeout(() => {
      templateCopied.value = false
    }, 1500)
  } catch (error) {
    console.error('Failed to copy template:', error)
    alert('复制失败，请手动复制文本')
  }
}

const escapeHtml = (unsafe) => {
  return String(unsafe || '')
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;')
}

const markdownToHtml = (markdown) => {
  const lines = String(markdown || '').split('\n')
  const html = []
  let inList = false

  const closeListIfNeeded = () => {
    if (inList) {
      html.push('</ul>')
      inList = false
    }
  }

  for (const rawLine of lines) {
    const line = rawLine.trimRight()
    const safeLine = escapeHtml(line)

    if (line.trim() === '') {
      closeListIfNeeded()
      continue
    }

    if (line.startsWith('### ')) {
      closeListIfNeeded()
      html.push(`<h3>${escapeHtml(line.slice(4))}</h3>`)
      continue
    }

    if (line.startsWith('## ')) {
      closeListIfNeeded()
      html.push(`<h2>${escapeHtml(line.slice(3))}</h2>`)
      continue
    }

    if (line.startsWith('# ')) {
      closeListIfNeeded()
      html.push(`<h1>${escapeHtml(line.slice(2))}</h1>`)
      continue
    }

    if (line.startsWith('- ') || line.startsWith('* ')) {
      if (!inList) {
        html.push('<ul>')
        inList = true
      }
      html.push(`<li>${escapeHtml(line.slice(2))}</li>`)
      continue
    }

    closeListIfNeeded()
    html.push(`<p>${safeLine}</p>`)
  }

  closeListIfNeeded()
  return html.join('\n')
}

const openPdfWindow = (autoPrint) => {
  const markdown = resumeTemplate.value?.templateMarkdown
  if (!markdown) return

  const role = resumeTemplate.value?.targetRole || '通用岗位'
  const now = new Date().toLocaleString('zh-CN')
  const guides = pdfIncludeGuides.value ? (resumeTemplate.value?.writingGuides || []) : []
  const mistakes = pdfIncludeMistakes.value ? (resumeTemplate.value?.commonMistakes || []) : []

  const previewWindow = window.open('', '_blank', 'noopener,noreferrer,width=1024,height=1400')
  if (!previewWindow) {
    alert('浏览器拦截了导出窗口，请允许弹窗后重试。为保证质量，本功能仅使用打印引擎导出 PDF。')
    return
  }

  const guidesHtml = guides.map((item) => `<li>${escapeHtml(item)}</li>`).join('')
  const mistakesHtml = mistakes.map((item) => `<li>${escapeHtml(item)}</li>`).join('')
  const contentHtml = markdownToHtml(markdown)

  const doc = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Resume Template Export</title>
  <style>
    @page {
      size: A4;
      margin: ${pdfMarginMm.value}mm ${pdfMarginMm.value}mm ${pdfMarginMm.value + 2}mm ${pdfMarginMm.value}mm;
    }
    * {
      box-sizing: border-box;
    }
    body {
      margin: 0;
      color: #101828;
      font-family: "Segoe UI", "Microsoft YaHei", sans-serif;
      font-size: ${pdfFontSize.value}pt;
      line-height: ${pdfLineHeight.value};
      -webkit-print-color-adjust: exact;
      print-color-adjust: exact;
      text-rendering: geometricPrecision;
    }
    .sheet {
      max-width: 180mm;
      margin: 0 auto;
    }
    .meta {
      border: 1px solid #d0d5dd;
      border-radius: 10px;
      padding: 10px 12px;
      margin-bottom: 12px;
      background: #f8fafc;
    }
    .meta h1 {
      margin: 0 0 4px;
      font-size: 16pt;
      line-height: 1.3;
    }
    .meta p {
      margin: 0;
      color: #475467;
      font-size: 10.5pt;
    }
    .section {
      break-inside: ${pdfKeepSections.value ? 'avoid' : 'auto'};
      page-break-inside: ${pdfKeepSections.value ? 'avoid' : 'auto'};
      margin-bottom: 10px;
    }
    .section h2 {
      margin: 0 0 6px;
      font-size: 12pt;
      border-left: 3px solid #344054;
      padding-left: 8px;
      line-height: 1.35;
    }
    .section ul {
      margin: 0;
      padding-left: 18px;
    }
    .section li {
      margin-bottom: 3px;
    }
    .content {
      border: 1px solid #d0d5dd;
      border-radius: 10px;
      padding: 12px;
    }
    .content h1,
    .content h2,
    .content h3 {
      margin: 14px 0 6px;
      line-height: 1.35;
      break-after: ${pdfKeepSections.value ? 'avoid' : 'auto'};
      page-break-after: ${pdfKeepSections.value ? 'avoid' : 'auto'};
    }
    .content h1 { font-size: 15pt; }
    .content h2 { font-size: 13pt; }
    .content h3 { font-size: 12pt; }
    .content p {
      margin: 5px 0;
      white-space: pre-wrap;
      word-break: break-word;
      orphans: 3;
      widows: 3;
    }
    .content ul {
      margin: 4px 0 8px;
      padding-left: 20px;
    }
    .quality-note {
      margin-top: 10px;
      font-size: 10pt;
      color: #344054;
      border-top: 1px dashed #d0d5dd;
      padding-top: 8px;
    }
    @media print {
      .no-print {
        display: none !important;
      }
    }
  </style>
</head>
<body>
  <main class="sheet">
    <section class="meta">
      <h1>高质量简历范本导出</h1>
      <p>目标岗位：${escapeHtml(role)} | 级别：${escapeHtml(templateSeniority.value)} | 导出时间：${escapeHtml(now)}</p>
    </section>

    ${guides.length ? `<section class="section"><h2>撰写建议</h2><ul>${guidesHtml}</ul></section>` : ''}
    ${mistakes.length ? `<section class="section"><h2>常见错误</h2><ul>${mistakesHtml}</ul></section>` : ''}

    <section class="content">${contentHtml}</section>

    <p class="quality-note">质量建议：在打印窗口中选择“另存为 PDF”，纸张选 A4，缩放 100%，背景图形保持开启，可获得更稳定的矢量文本质量。</p>
    <div class="no-print" style="margin-top:12px; display:flex; gap:8px;">
      <button onclick="window.print()" style="border:1px solid #d0d5dd; background:#fff; padding:6px 10px; border-radius:8px; cursor:pointer;">打印/导出 PDF</button>
      <button onclick="window.close()" style="border:1px solid #d0d5dd; background:#fff; padding:6px 10px; border-radius:8px; cursor:pointer;">关闭</button>
    </div>
  </main>
</body>
</html>`

  previewWindow.document.open()
  previewWindow.document.write(doc)
  previewWindow.document.close()
  previewWindow.focus()
  if (autoPrint) {
    setTimeout(() => {
      previewWindow.print()
    }, 180)
  }
}

const previewTemplatePdf = () => {
  openPdfWindow(false)
}

const exportTemplatePdf = () => {
  openPdfWindow(true)
}

const downloadTemplate = (format) => {
  const markdown = resumeTemplate.value?.templateMarkdown
  if (!markdown) return

  const role = (resumeTemplate.value.targetRole || 'resume-template').replace(/\s+/g, '-')
  const ext = format === 'txt' ? 'txt' : 'md'
  const fileName = `${role}-${templateSeniority.value}.${ext}`
  const mimeType = format === 'txt' ? 'text/plain;charset=utf-8' : 'text/markdown;charset=utf-8'
  const blob = new Blob([markdown], { type: mimeType })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(link.href)
}
</script>

<template>
  <div class="space-y-8">
    <!-- Header -->
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900">简历智能匹配</h1>
      <p class="text-zinc-500 mt-2">上传您的简历，AI 将为您解析核心竞争力并推荐匹配岗位</p>
    </header>

    <!-- Upload Area (when no resume data) -->
    <div 
      v-if="!resumeData"
      class="border-2 border-dashed border-zinc-200 rounded-3xl p-12 flex flex-col items-center justify-center bg-white hover:border-indigo-300 transition-colors cursor-pointer relative"
      @click="triggerFileInput"
      @dragover.prevent
      @drop.prevent="handleDrop"
    >
      <input 
        type="file" 
        ref="fileInput" 
        class="hidden" 
        accept=".pdf,.doc,.docx,.txt"
        @change="handleFileChange"
      />
      
      <div class="h-16 w-16 bg-indigo-50 rounded-2xl text-indigo-600 mb-4 flex items-center justify-center">
        <Sparkles v-if="isUploading" class="h-8 w-8 animate-pulse" />
        <Upload v-else class="h-8 w-8" />
      </div>
      
      <h3 class="text-lg font-medium text-zinc-900">
        {{ isUploading ? '正在解析简历...' : '点击或拖拽文件上传简历' }}
      </h3>
      <p class="text-zinc-500 mt-1 text-sm">支持 PDF, Word 或 TXT 格式</p>
    </div>

    <!-- Results Area (when resume data exists) -->
    <div v-else class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Left Column: Parsing Result -->
      <div class="lg:col-span-1 space-y-6">
        <div class="bg-white rounded-3xl p-6 shadow-sm border border-zinc-100">
          <div class="flex items-center gap-3 mb-6">
            <div class="h-10 w-10 rounded-xl bg-indigo-50 flex items-center justify-center text-indigo-600">
              <BrainCircuit class="h-6 w-6" />
            </div>
            <h3 class="text-lg font-bold text-zinc-900">简历解析结果</h3>
          </div>

          <div class="space-y-6">
            <!-- Tech Stack -->
            <div>
              <h4 class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-3">技术栈</h4>
              <div class="flex flex-wrap gap-2">
                <span 
                  v-for="tech in resumeData.techStack" 
                  :key="tech"
                  class="px-3 py-1 bg-zinc-100 rounded-full text-sm font-medium text-zinc-700"
                >
                  {{ tech }}
                </span>
              </div>
            </div>

            <!-- Intent -->
            <div>
              <h4 class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-3">求职意向</h4>
              <p class="text-zinc-700 font-medium">{{ resumeData.intent }}</p>
            </div>

            <!-- Soft Skills -->
            <div>
              <h4 class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-3">软技能</h4>
              <div class="flex flex-wrap gap-2">
                <span 
                  v-for="skill in resumeData.softSkills" 
                  :key="skill"
                  class="px-3 py-1 bg-emerald-50 text-emerald-700 rounded-full text-sm font-medium"
                >
                  {{ skill }}
                </span>
              </div>
            </div>
          </div>

          <button 
            @click="resetUpload"
            class="w-full mt-8 py-3 border border-zinc-200 rounded-2xl text-sm font-medium text-zinc-600 hover:bg-zinc-50 transition-colors"
          >
            重新上传
          </button>
        </div>
      </div>

      <!-- Right Column: Job Matches -->
      <div class="lg:col-span-2 space-y-6">
        <div class="flex items-center gap-3 mb-2">
          <Search class="h-5 w-5 text-zinc-400" />
          <h3 class="text-lg font-bold text-zinc-900">推荐岗位匹配 (RAG)</h3>
        </div>

        <div class="space-y-4">
          <div 
            v-for="(match, index) in matches" 
            :key="index"
            class="bg-white rounded-3xl p-6 shadow-sm border border-zinc-100 hover:shadow-md transition-shadow"
          >
            <div class="flex items-start justify-between mb-4">
              <div>
                <h4 class="text-lg font-bold text-zinc-900 flex items-center gap-2">
                  {{ match.jobTitle }}
                  <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-600 text-xs font-bold">
                    <CheckCircle2 class="h-3 w-3" />
                    {{ match.matchScore }}% 匹配
                  </span>
                </h4>
              </div>
              <button 
                @click="startInterview(match)"
                class="bg-indigo-600 text-white rounded-xl px-4 py-2 text-sm font-medium hover:bg-indigo-700 transition-colors"
              >
                生成面试题
              </button>
            </div>

            <p class="text-zinc-600 text-sm leading-relaxed mb-4">
              {{ match.reason }}
            </p>

            <div class="flex flex-wrap gap-2">
              <span 
                v-for="req in match.requirements" 
                :key="req"
                class="border border-zinc-200 px-2 py-1 rounded-lg text-xs text-zinc-500"
              >
                {{ req }}
              </span>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-3xl p-6 shadow-sm border border-zinc-100">
          <h3 class="text-lg font-bold text-zinc-900 mb-4">简历增强分析</h3>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div>
              <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">目标岗位</label>
              <input
                v-model="targetRole"
                type="text"
                placeholder="例如：后端开发工程师"
                class="mt-2 w-full border border-zinc-200 rounded-xl px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
              />
            </div>

            <div>
              <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">范本级别</label>
              <select
                v-model="templateSeniority"
                class="mt-2 w-full border border-zinc-200 rounded-xl px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
              >
                <option value="0-1年">0-1年</option>
                <option value="1-3年">1-3年</option>
                <option value="3-5年">3-5年</option>
                <option value="5年以上">5年以上</option>
              </select>
            </div>
          </div>

          <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">可选：粘贴简历原文（提升真实性分析效果）</label>
          <textarea
            v-model="rawResumeText"
            rows="4"
            class="mt-2 w-full border border-zinc-200 rounded-xl px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
            placeholder="可粘贴 PDF 提取原文，帮助识别夸张表述与核验点"
          />

          <div class="grid grid-cols-1 md:grid-cols-3 gap-3 mt-4">
            <button
              @click="runAuthenticity"
              :disabled="authenticityLoading"
              class="py-2.5 rounded-xl text-sm font-medium border border-zinc-200 hover:bg-zinc-50 disabled:opacity-60"
            >
              {{ authenticityLoading ? '分析中...' : '真实性核验' }}
            </button>
            <button
              @click="runOptimization"
              :disabled="optimizationLoading"
              class="py-2.5 rounded-xl text-sm font-medium border border-zinc-200 hover:bg-zinc-50 disabled:opacity-60"
            >
              {{ optimizationLoading ? '生成中...' : '优化建议' }}
            </button>
            <button
              @click="runTemplate"
              :disabled="templateLoading"
              class="py-2.5 rounded-xl text-sm font-medium bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-60"
            >
              {{ templateLoading ? '生成中...' : '简历范本' }}
            </button>
          </div>
        </div>

        <div v-if="authenticityReport" class="bg-white rounded-3xl p-6 shadow-sm border border-zinc-100 space-y-4">
          <h3 class="text-lg font-bold text-zinc-900">真实性核验报告</h3>
          <p class="text-sm text-zinc-600">风险评分：<span class="font-semibold text-zinc-900">{{ authenticityReport.overallRiskScore }}</span></p>
          <p class="text-sm text-zinc-700">{{ authenticityReport.summary }}</p>
          <div>
            <h4 class="text-sm font-semibold text-zinc-900 mb-2">潜在风险项</h4>
            <div class="space-y-2">
              <div v-for="(item, i) in authenticityReport.potentialRiskItems || []" :key="i" class="p-3 rounded-xl bg-amber-50 border border-amber-100">
                <p class="text-sm font-medium text-zinc-900">{{ item.claim }} <span class="text-xs text-amber-700">({{ item.riskLevel }})</span></p>
                <p class="text-xs text-zinc-600 mt-1">{{ item.reason }}</p>
                <p class="text-xs text-zinc-500 mt-1">核验建议：{{ item.verificationTip }}</p>
              </div>
            </div>
          </div>
          <p class="text-xs text-zinc-400">{{ authenticityReport.disclaimer }}</p>
        </div>

        <div v-if="optimizationReport" class="bg-white rounded-3xl p-6 shadow-sm border border-zinc-100 space-y-4">
          <h3 class="text-lg font-bold text-zinc-900">简历优化建议</h3>
          <p class="text-sm text-zinc-600">综合评分：<span class="font-semibold text-zinc-900">{{ optimizationReport.overallScore }}</span></p>
          <div>
            <h4 class="text-sm font-semibold text-zinc-900 mb-2">建议清单</h4>
            <ul class="space-y-1 text-sm text-zinc-700">
              <li v-for="(s, i) in optimizationReport.suggestions || []" :key="i">- {{ s }}</li>
            </ul>
          </div>
          <div>
            <h4 class="text-sm font-semibold text-zinc-900 mb-2">改写示例</h4>
            <ul class="space-y-1 text-sm text-zinc-700">
              <li v-for="(d, i) in optimizationReport.rewriteDemo || []" :key="i">- {{ d }}</li>
            </ul>
          </div>
        </div>

        <div v-if="resumeTemplate" class="bg-white rounded-3xl p-6 shadow-sm border border-zinc-100 space-y-4">
          <div class="border border-zinc-200 rounded-2xl p-4 bg-zinc-50/70">
            <h4 class="text-sm font-semibold text-zinc-900 mb-3">PDF 导出参数（先预览再导出）</h4>
            <div class="mb-3 flex flex-wrap gap-2">
              <button
                @click="applyPdfPreset('campus')"
                :class="activePdfPreset === 'campus' ? 'bg-zinc-900 text-white border-zinc-900' : 'bg-white text-zinc-700 border-zinc-200'"
                class="px-3 py-1.5 rounded-lg border text-xs font-medium"
              >
                校招简洁版
              </button>
              <button
                @click="applyPdfPreset('experienced')"
                :class="activePdfPreset === 'experienced' ? 'bg-zinc-900 text-white border-zinc-900' : 'bg-white text-zinc-700 border-zinc-200'"
                class="px-3 py-1.5 rounded-lg border text-xs font-medium"
              >
                社招平衡版
              </button>
              <button
                @click="applyPdfPreset('technical')"
                :class="activePdfPreset === 'technical' ? 'bg-zinc-900 text-white border-zinc-900' : 'bg-white text-zinc-700 border-zinc-200'"
                class="px-3 py-1.5 rounded-lg border text-xs font-medium"
              >
                技术细节版
              </button>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
              <label class="text-xs text-zinc-600">
                字号（pt）
                <input v-model.number="pdfFontSize" type="number" min="10" max="14" step="0.5" class="mt-1 w-full border border-zinc-200 rounded-lg px-2 py-1.5 text-sm" />
              </label>
              <label class="text-xs text-zinc-600">
                行距
                <input v-model.number="pdfLineHeight" type="number" min="1.3" max="2" step="0.05" class="mt-1 w-full border border-zinc-200 rounded-lg px-2 py-1.5 text-sm" />
              </label>
              <label class="text-xs text-zinc-600">
                页边距（mm）
                <input v-model.number="pdfMarginMm" type="number" min="8" max="22" step="1" class="mt-1 w-full border border-zinc-200 rounded-lg px-2 py-1.5 text-sm" />
              </label>
            </div>
            <div class="mt-3 flex flex-wrap gap-4 text-xs text-zinc-600">
              <label class="inline-flex items-center gap-2"><input v-model="pdfKeepSections" type="checkbox" /> 强分页保护（减少段落断页）</label>
              <label class="inline-flex items-center gap-2"><input v-model="pdfIncludeGuides" type="checkbox" /> 包含撰写建议</label>
              <label class="inline-flex items-center gap-2"><input v-model="pdfIncludeMistakes" type="checkbox" /> 包含常见错误</label>
            </div>
          </div>

          <div class="flex flex-wrap items-center justify-between gap-3">
            <h3 class="text-lg font-bold text-zinc-900">简历范本：{{ resumeTemplate.targetRole }}</h3>
            <div class="flex flex-wrap gap-2">
              <button
                @click="copyTemplate"
                class="px-3 py-1.5 rounded-lg border border-zinc-200 text-xs font-medium text-zinc-700 hover:bg-zinc-50"
              >
                {{ templateCopied ? '已复制' : '复制范本' }}
              </button>
              <button
                @click="downloadTemplate('md')"
                class="px-3 py-1.5 rounded-lg border border-zinc-200 text-xs font-medium text-zinc-700 hover:bg-zinc-50"
              >
                下载 .md
              </button>
              <button
                @click="downloadTemplate('txt')"
                class="px-3 py-1.5 rounded-lg border border-zinc-200 text-xs font-medium text-zinc-700 hover:bg-zinc-50"
              >
                下载 .txt
              </button>
              <button
                @click="previewTemplatePdf"
                class="px-3 py-1.5 rounded-lg border border-zinc-200 text-xs font-medium text-zinc-700 hover:bg-zinc-50"
              >
                预览 PDF
              </button>
              <button
                @click="exportTemplatePdf"
                class="px-3 py-1.5 rounded-lg bg-indigo-600 text-xs font-medium text-white hover:bg-indigo-700"
              >
                导出高质量 PDF
              </button>
            </div>
          </div>

          <div v-if="(resumeTemplate.writingGuides || []).length">
            <h4 class="text-sm font-semibold text-zinc-900 mb-2">撰写建议</h4>
            <ul class="space-y-1 text-sm text-zinc-700">
              <li v-for="(guide, i) in resumeTemplate.writingGuides" :key="`guide-${i}`">- {{ guide }}</li>
            </ul>
          </div>

          <div v-if="(resumeTemplate.commonMistakes || []).length">
            <h4 class="text-sm font-semibold text-zinc-900 mb-2">常见错误</h4>
            <ul class="space-y-1 text-sm text-zinc-700">
              <li v-for="(mistake, i) in resumeTemplate.commonMistakes" :key="`mistake-${i}`">- {{ mistake }}</li>
            </ul>
          </div>

          <pre class="text-xs text-zinc-700 bg-zinc-50 border border-zinc-100 rounded-2xl p-4 whitespace-pre-wrap">{{ resumeTemplate.templateMarkdown }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>
