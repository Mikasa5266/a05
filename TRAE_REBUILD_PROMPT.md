# 🎯 AI Interview Pro — Trae 重建指南（完整设计 + 功能规格）

> **目标**：请严格按照本文档的设计系统、组件规格和功能说明，重建 AI Interview Pro 项目。
> 本项目已有可运行代码，你的任务是确保 UI **完全对齐**下方设计规范（原代码 UI 不够美观），同时功能保持完整。

---

## 📦 技术栈

```
React 19 + TypeScript
Vite 6
Tailwind CSS v4（@tailwindcss/vite 插件）
react-router-dom v7（BrowserRouter）
motion/react（Framer Motion）
lucide-react（图标库）
recharts（图表）
clsx + tailwind-merge（样式合并工具）
@google/genai（Gemini API）
```

字体引入（index.css）：
```css
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap');
```

---

## 🎨 设计系统（必须严格遵守）

### 色彩系统
| 用途 | Tailwind 类 | 说明 |
|------|------------|------|
| 页面背景 | `bg-zinc-50` | 整体底色 |
| 卡片背景 | `bg-white` | 所有卡片 |
| 深色卡片/侧边栏底部 | `bg-zinc-900` | 对比区块 |
| 主色调 | `indigo-600` | 按钮、高亮、图标 |
| 成功色 | `emerald-500/600` | 正向指标 |
| 警告色 | `amber-500/600` | 中等风险 |
| 危险色 | `rose-500/600` | 负向指标 |
| 文字主色 | `text-zinc-900` | 标题 |
| 文字次色 | `text-zinc-500` | 描述文字 |
| 文字弱色 | `text-zinc-400` | label、hint |

### 圆角规则
- **卡片**：`rounded-3xl`（所有卡片、面板、大区块）
- **按钮（主要）**：`rounded-2xl`
- **标签/徽章**：`rounded-full` 或 `rounded-lg`
- **小元素**（进度条等）：`rounded-full`
- **图标容器**：`rounded-2xl`

### 阴影规则
- 普通卡片：`shadow-sm border border-zinc-100`
- 强调卡片（悬浮中）：`shadow-md`
- 不使用 `shadow-lg` 或更强的阴影

### 间距规则
- 页面内边距：`p-8`
- 卡片内边距：`p-6` 或 `p-8`
- 区块间距：`space-y-8`
- 元素间距：`gap-4` / `gap-6`

### 字体规则
```
font-family: Inter（全局）
标题（页面级）: text-3xl font-bold / text-4xl font-bold
标题（卡片级）: text-lg font-bold / text-xl font-semibold
正文: text-sm / text-base，font-medium
标签/说明: text-xs font-bold text-zinc-400 uppercase tracking-wider
```

### 过渡动画
- 页面切换：`motion/react` AnimatePresence，`opacity: 0→1, y: 20→0`，duration 0.2s
- 卡片 hover：`hover:shadow-md transition-all`
- 图标 hover：`group-hover:scale-110 transition-transform`
- 按钮颜色：`transition-colors`

---

## 🗂️ 项目文件结构

```
/
├── index.html
├── index.css
├── main.tsx
├── App.tsx
├── types.ts
├── vite.config.ts
├── tsconfig.json
├── package.json
├── server.ts
├── metadata.json
├── components/
│   ├── ResumeMatching.tsx
│   ├── MockInterview.tsx
│   └── GrowthCenter.tsx
└── services/
    └── gemini.ts
```

---

## 🧭 App.tsx — 整体布局

### 布局结构
```
<Router>
  <div class="flex min-h-screen bg-zinc-50">
    <Sidebar />          ← 固定左侧，宽 w-64，h-screen sticky top-0
    <main class="flex-1 p-8 overflow-y-auto">
      <div class="max-w-5xl mx-auto">
        <AnimatePresence mode="wait">
          <Routes /> ← 带 motion.div 包裹的页面切换
        </AnimatePresence>
      </div>
    </main>
  </div>
</Router>
```

### Sidebar 规格
```
├── Logo 区（p-6）
│   ├── 8×8 indigo-600 rounded-lg 图标容器 + BrainCircuit 图标（白色）
│   └── "AI Interview" 文字，font-bold text-xl
│
├── 导航（flex-1 px-4 space-y-1）
│   导航项 5 个：首页(/)、简历匹配(/resume)、模拟面试(/interview)、成长中心(/growth)、面试记录(/history)
│   激活态：bg-indigo-50 text-indigo-600
│   未激活：text-zinc-500 hover:bg-zinc-50 hover:text-zinc-900
│   每项：flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium
│
└── 底部（p-4 border-t border-zinc-100）
    ├── 设置按钮（同导航项样式）
    └── 用户信息：头像（h-8 w-8 rounded-full bg-zinc-100）+ 姓名 + 职位
```

### 路由配置
```
/         → Home（首页仪表盘）
/resume   → ResumeMatching
/interview → MockInterview
/growth   → GrowthCenter
/history  → 占位页（"面试记录模块开发中..."）
```

---

## 🏠 Home 首页

### 区块 1：Header
```
<h1 class="text-4xl font-bold tracking-tight text-zinc-900">欢迎回来，面试官已就绪</h1>
<p class="text-zinc-500 text-lg">开启你的 AI 驱动面试成长之旅</p>
```

### 区块 2：3列功能入口卡片
`grid grid-cols-1 md:grid-cols-3 gap-6`

每张卡片规格（Link 包裹，to 对应路由）：
```
group relative overflow-hidden rounded-3xl bg-white p-8 shadow-sm border border-zinc-100
hover:shadow-md transition-all

内部：
- 图标容器：h-12 w-12 rounded-2xl 带色背景，group-hover:scale-110 transition-transform
- 标题：text-xl font-semibold text-zinc-900
- 描述：mt-2 text-zinc-500
- ChevronRight：absolute bottom-8 right-8，group-hover:translate-x-1

3张卡片：
1. 简历匹配 → indigo-50/indigo-600 + FileText 图标
2. 模拟面试 → emerald-50/emerald-600 + Video 图标
3. 成长中心 → amber-50/amber-600 + TrendingUp 图标
```

### 区块 3：面试表现看板（深色）
```
rounded-3xl bg-zinc-900 p-8 text-white

标题行：flex items-center justify-between
- h2 text-2xl font-bold "最近的面试表现"
- "查看全部" 按钮：text-sm text-zinc-400 hover:text-white

4 格统计（grid grid-cols-1 md:grid-cols-4 gap-4）：
每格：bg-zinc-800 rounded-2xl p-4
- label: text-zinc-400 text-sm mb-2
- 数值: text-2xl font-bold
- 进度条: h-1 w-16 bg-zinc-700 rounded-full overflow-hidden，内部对应色宽度

数据：
技术深度 85% indigo-500
表达能力 72% emerald-500
逻辑严谨 90% amber-500
岗位匹配 78% rose-500
```

---

## 📄 ResumeMatching.tsx

### 状态
```typescript
isUploading: boolean
resumeData: ResumeData | null
matches: JobMatch[]
```

### 未上传状态：上传区
```
border-2 border-dashed border-zinc-200 rounded-3xl p-12
flex flex-col items-center justify-center bg-white
hover:border-indigo-300 transition-colors cursor-pointer

内：
- 图标容器：h-16 w-16 bg-indigo-50 rounded-2xl text-indigo-600 mb-4
  - 上传中：<Sparkles animate-pulse />
  - 待上传：<Upload />
- 文字："点击或拖拽文件上传简历"（text-lg font-medium）
- 副文字："支持 PDF, Word 或 TXT 格式"（text-zinc-500 mt-1）
- <input type="file" class="absolute inset-0 opacity-0 cursor-pointer" />
```

### 已上传状态：3列网格
`grid grid-cols-1 lg:grid-cols-3 gap-8`

**左列（1/3）— 简历解析结果卡片：**
```
bg-white rounded-3xl p-6 shadow-sm border border-zinc-100

头部：h3 with Brain icon（indigo-600）"简历解析结果"

数据展示：
- 技术栈：flex flex-wrap gap-2，标签：px-3 py-1 bg-zinc-100 rounded-full text-sm
- 求职意向：直接文字
- 软技能：px-3 py-1 bg-emerald-50 text-emerald-700 rounded-full text-sm

底部：重新上传按钮（border rounded-2xl）
```

**右列（2/3）— 岗位匹配卡片列表：**
```
h3 with Search icon "推荐岗位匹配 (RAG)"

每个 JobMatch 卡片（motion.div，animate x方向 + 延迟）：
bg-white rounded-3xl p-6 shadow-sm border border-zinc-100

内：
- 左上：职位名（text-lg font-bold）+ 匹配度（CheckCircle2 + emerald-600 text-sm font-bold）
- 右上：生成面试题按钮（bg-indigo-600 text-white rounded-xl px-4 py-2 text-sm）
- 理由：text-zinc-600 text-sm leading-relaxed
- 技能标签：border border-zinc-100 px-2 py-1 rounded-lg text-xs text-zinc-400
```

### Gemini 集成（services/gemini.ts）
```typescript
// 调用 parseResume(text: string): Promise<ResumeData>
// 模型：gemini-2.0-flash（或最新可用模型）
// 使用结构化输出（responseSchema）
// 解析失败时 fallback 到 mock data
```

---

## 🎤 MockInterview.tsx

### 阶段状态
```
setup → interview → evaluating
```

### Setup 阶段（max-w-2xl mx-auto）

**预览区（aspect-video bg-zinc-900 rounded-2xl）：**
- 摄像头关闭时：VideoOff 图标 + "摄像头已关闭"
- 摄像头开启时：灰色背景 + "摄像头预览区域"
- 底部控制栏（absolute bottom-4 left-1/2 -translate-x-1/2）：
  - 麦克风按钮：h-12 w-12 rounded-full
    - 开启：bg-white/10 text-white
    - 关闭：bg-rose-500 text-white
  - 摄像头按钮：同上

**配置区（2列 grid）：**
- 面试模式 Select：技术面/HR面/综合面
- 面试官风格 Select：温和型/压力型/技术深挖型
- Select 样式：bg-zinc-50 border border-zinc-100 rounded-xl px-4 py-2 text-sm

**开始按钮：**
```
w-full py-4 bg-indigo-600 text-white rounded-2xl font-bold text-lg
hover:bg-indigo-700 transition-all
flex items-center justify-center gap-2
文字："进入面试间" + ChevronRight
```

### Interview 阶段（h-[calc(100vh-8rem)] flex flex-col gap-6）

**主视图（lg:col-span-3 bg-zinc-900 rounded-3xl relative overflow-hidden group）：**
- 中央：BrainCircuit 图标（h-32 w-32 bg-indigo-600/20 border-indigo-500/30 rounded-full animate-pulse）
- 用户视频（absolute bottom-6 right-6 w-64 aspect-video bg-zinc-800 rounded-2xl border border-white/10 shadow-2xl）
  - `<video ref={videoRef} autoPlay muted />`
  - 左上角：px-2 py-1 bg-black/50 rounded-lg text-[10px] text-white uppercase "YOU"
- 控制栏（opacity-0 group-hover:opacity-100 transition-opacity，absolute bottom-6 left-1/2 -translate-x-1/2）：
  - 设置按钮：bg-white/10 backdrop-blur-md
  - 退出按钮：bg-rose-500

**右侧栏（lg:col-span-1 flex flex-col gap-6）：**

问题卡片（bg-white rounded-3xl p-6 flex-1 flex flex-col）：
```
- 标题：text-sm font-bold text-zinc-400 uppercase tracking-wider "当前问题"
- 问题文本：text-xl font-bold text-zinc-900 leading-tight
- 录音按钮：
  录音中：bg-rose-50 text-rose-600 animate-pulse + Mic 图标 "正在录音..."
  待开始：bg-indigo-600 text-white + Play 图标 "开始回答"
- 下一题按钮：border border-zinc-200 text-zinc-600 hover:bg-zinc-50
```

实时分析卡片（bg-zinc-900 rounded-3xl p-6 text-white）：
```
- 头部：绿点 animate-pulse + "实时分析" uppercase 小标
- 两项指标（语速/情绪），各有：
  - 两端文字（左：指标名 text-zinc-400，右：状态 colored）
  - 进度条：h-1 bg-zinc-800 rounded-full，内部有色填充
```

---

## 📈 GrowthCenter.tsx

### 布局
```
space-y-8

Header（flex items-end justify-between）：
- 左：标题 + 描述
- 右：等级徽章（bg-indigo-50 text-indigo-600 px-4 py-2 rounded-2xl font-bold text-sm，Award 图标）
  文字："当前等级: L4 高级工程师"

主内容：grid grid-cols-1 lg:grid-cols-3 gap-8
```

**雷达图卡片（lg:col-span-1，bg-white rounded-3xl p-8）：**
```typescript
const radarData = [
  { subject: '技术深度', A: 85, fullMark: 100 },
  { subject: '表达能力', A: 72, fullMark: 100 },
  { subject: '逻辑严谨', A: 90, fullMark: 100 },
  { subject: '岗位匹配', A: 78, fullMark: 100 },
  { subject: '行为表现', A: 65, fullMark: 100 },
];
```
Recharts RadarChart 配置：
- PolarGrid stroke="#f4f4f5"
- PolarAngleAxis fill="#71717a" fontSize=12
- Radar stroke="#4f46e5" fill="#4f46e5" fillOpacity=0.1
- ResponsiveContainer width="100%" height="100%"，min-h-[300px]

卡片底部 AI 点评区：`bg-zinc-50 rounded-2xl p-4 mt-6`

**右侧区（lg:col-span-2 space-y-8）：**

成长曲线卡片（bg-white rounded-3xl p-8）：
```typescript
const growthData = [
  { name: 'Jan', score: 65 },
  { name: 'Feb', score: 68 },
  { name: 'Mar', score: 75 },
  { name: 'Apr', score: 82 },
  { name: 'May', score: 85 },
  { name: 'Jun', score: 90 },
];
```
LineChart 配置：
- CartesianGrid strokeDasharray="3 3" vertical=false stroke="#f4f4f5"
- XAxis axisLine=false tickLine=false fill="#a1a1aa" fontSize=12
- YAxis hide
- Tooltip：borderRadius=16 border=none boxShadow
- Line：stroke="#10b981" strokeWidth=4，dot fill="#10b981" r=4 stroke="#fff"

**下方 2 列 grid：**

技能缺口分析（bg-white rounded-3xl p-6）：
```
3 项，每项：flex items-center justify-between p-3 rounded-xl border border-zinc-50
左：skill 名（text-sm font-medium）
右：level 徽章（text-[10px] font-bold px-2 py-1 rounded-lg uppercase tracking-wider）

分布式事务 → 中等（amber）
React 性能优化 → 急需提升（rose）
系统架构设计 → 良好（emerald）
```

个性化学习地图（bg-zinc-900 text-white rounded-3xl p-6 relative overflow-hidden）：
```
- label: text-sm font-bold text-zinc-400 uppercase "个性化学习地图"
- 推荐文字: text-lg font-bold mb-4
- 开始学习按钮: text-sm font-bold text-indigo-400 + ChevronRight
- 背景装饰: BookOpen absolute -bottom-4 -right-4 h-24 w-24 text-white/5 rotate-12
```

**面试回放区（section）：**
```
标题：text-xl font-bold "面试回放与复盘"

2 列 grid，每项：
bg-white rounded-3xl p-4 shadow-sm border border-zinc-100
flex items-center gap-4 group cursor-pointer hover:border-indigo-200 transition-colors

- 缩略图：h-20 w-32 bg-zinc-100 rounded-2xl，内含 PlayCircle（group-hover:text-indigo-600）
- 信息区：
  - 标题: text-sm font-bold "2024-05-12 技术面模拟"
  - 副标: text-xs text-zinc-500 "岗位: 高级全栈工程师 | 得分: 82"
  - 两个徽章: AI 标注 12 处（bg-zinc-100 text-zinc-500）和 表现优异（bg-emerald-50 text-emerald-600）
- ChevronRight（group-hover:text-zinc-900 transition-colors）
```

---

## ⚙️ types.ts（必须完整保留）

```typescript
export interface ResumeData {
  techStack: string[];
  experience: { title: string; description: string; highlights: string[] }[];
  intent: string;
  softSkills: string[];
}

export interface JobMatch {
  jobTitle: string;
  matchScore: number;
  reason: string;
  requirements: string[];
}

export interface InterviewQuestion {
  id: string;
  text: string;
  type: 'technical' | 'behavioral' | 'follow-up';
  context?: string;
}

export interface EvaluationMetrics {
  technical: number;
  expression: number;
  logic: number;
  matching: number;
  behavior: number;
}

export interface InterviewSession {
  id: string;
  date: string;
  mode: string;
  style: string;
  questions: InterviewQuestion[];
  answers: { questionId: string; text: string; audioUrl?: string }[];
  evaluation?: {
    metrics: EvaluationMetrics;
    summary: string;
    strengths: string[];
    weaknesses: string[];
  };
}
```

---

## 🤖 services/gemini.ts

```typescript
import { GoogleGenAI, Type } from "@google/genai";
import { ResumeData, InterviewQuestion } from "../types";

const ai = new GoogleGenAI({ apiKey: process.env.GEMINI_API_KEY });

// parseResume: 使用结构化输出解析简历文本
// generateQuestions: 根据简历和面试模式生成5道问题
// evaluateAnswer: 评估单条回答，返回 feedback 和 score

// 模型建议使用：gemini-2.0-flash 或 gemini-1.5-flash（gemini-3 系列不存在）
```

---

## 📋 index.css

```css
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap');
@import "tailwindcss";

@theme {
  --font-sans: "Inter", ui-sans-serif, system-ui, sans-serif;
  --font-mono: "JetBrains Mono", ui-monospace, SFMono-Regular, monospace;
}

@layer base {
  body {
    @apply antialiased text-zinc-900 bg-zinc-50;
  }
}

/* Recharts 雷达图样式修复 */
.radar-chart-container .recharts-polar-grid-concentric-path {
  stroke: #f4f4f5;
}
.radar-chart-container .recharts-polar-grid-angle-line {
  stroke: #f4f4f5;
}
```

---

## ⚠️ 常见错误避免

1. **不要使用 `shadow-lg` 或更重的阴影**，统一用 `shadow-sm`
2. **卡片圆角必须是 `rounded-3xl`**，不是 `rounded-xl` 或 `rounded-2xl`
3. **所有主要按钮**颜色为 `bg-indigo-600`，hover 为 `bg-indigo-700`
4. **不要使用 `gemini-3-flash-preview`** 这个模型名不存在，用 `gemini-2.0-flash`
5. **`motion/react`** 是 `framer-motion` v12 的包名，不要写成 `framer-motion`
6. **页面 padding**：main 区域始终 `p-8`，内容限宽 `max-w-5xl mx-auto`
7. **所有 label 类文字**（区块标题）：`text-xs font-bold text-zinc-400 uppercase tracking-wider`
8. **深色卡片背景**：`bg-zinc-900`（不是 `bg-gray-900` 或 `bg-slate-900`）
9. **图标容器**统一：`inline-flex items-center justify-center rounded-2xl`，大小 `h-12 w-12`
10. **AnimatePresence** 包裹 Routes，每个路由页面用 `motion.div`（`initial opacity:0 y:20`，`animate opacity:1 y:0`，`exit opacity:0 y:-20`）

---

## ✅ 检查清单

完成重建后，请确认：
- [ ] 左侧 Sidebar 固定，宽度 w-64，包含 Logo + 导航 + 用户信息
- [ ] 路由切换有 fade + slide 动画
- [ ] 首页 3 列功能卡片，hover 有 shadow 变化和图标缩放
- [ ] 首页底部深色看板显示 4 项指标
- [ ] 简历上传页有拖拽区域，上传后显示解析结果和岗位匹配
- [ ] 模拟面试有 Setup → Interview 两个阶段
- [ ] Interview 阶段有主视图（AI 面试官）+ 用户小窗 + 控制栏（hover 显示）
- [ ] 成长中心有雷达图（Recharts）+ 折线图 + 技能缺口 + 面试回放
- [ ] 所有卡片圆角为 rounded-3xl
- [ ] 字体为 Inter

---

*本文档由原始项目代码逆向生成，完整还原了设计意图。如有疑问，以本文档为准。*
