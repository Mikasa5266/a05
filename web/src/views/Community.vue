<template>
  <div class="space-y-8">
    <header class="flex items-end justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-zinc-900">校友面试经验共享社区</h1>
        <p class="text-zinc-500 mt-2">搜索校友面经 · 预约1v1模拟面试 · 构建校友-学生-企业生态闭环</p>
      </div>
      <button @click="showShareModal = true" class="px-5 py-2.5 bg-indigo-600 text-white rounded-xl text-sm font-medium hover:bg-indigo-700 transition-colors shadow-lg shadow-indigo-200">
        + 分享面经
      </button>
    </header>

    <!-- Search & Filter -->
    <div class="flex items-center gap-3 flex-wrap">
      <div class="relative flex-1 max-w-md">
        <Search class="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-400" />
        <input v-model="searchQuery" placeholder="搜索目标岗位、公司、关键词..."
          class="w-full pl-11 pr-4 py-3 bg-white border border-zinc-200 rounded-2xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 shadow-sm" />
      </div>
      <select v-model="filterType" class="px-4 py-3 bg-white border border-zinc-200 rounded-2xl text-sm shadow-sm">
        <option value="">全部类型</option>
        <option value="experience">面经分享</option>
        <option value="tips">面试技巧</option>
        <option value="jobReq">岗位要求</option>
      </select>
      <select v-model="filterCompany" class="px-4 py-3 bg-white border border-zinc-200 rounded-2xl text-sm shadow-sm">
        <option value="">全部公司</option>
        <option value="baidu">百度</option>
        <option value="alibaba">阿里巴巴</option>
        <option value="tencent">腾讯</option>
        <option value="huawei">华为</option>
        <option value="bytedance">字节跳动</option>
      </select>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Posts List -->
      <div class="lg:col-span-2 space-y-6">
        <div v-for="(post, idx) in filteredPosts" :key="idx"
          class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm hover:shadow-md transition-shadow"
        >
          <!-- Author Info -->
          <div class="flex items-center gap-3 mb-4">
            <div class="h-10 w-10 rounded-full bg-indigo-100 text-indigo-600 flex items-center justify-center font-bold text-sm">
              {{ post.author.charAt(0) }}
            </div>
            <div>
              <div class="font-medium text-zinc-900 flex items-center gap-2">
                {{ post.author }}
                <span v-if="post.verified" class="px-1.5 py-0.5 bg-emerald-50 text-emerald-600 rounded text-[10px] font-bold">已就业认证</span>
              </div>
              <div class="text-xs text-zinc-400">{{ post.school }} · {{ post.year }}届 · {{ post.position }}</div>
            </div>
          </div>

          <!-- Post Content -->
          <h3 class="text-lg font-bold text-zinc-900 mb-2">{{ post.title }}</h3>
          <p class="text-sm text-zinc-600 leading-relaxed mb-4 line-clamp-3">{{ post.content }}</p>

          <!-- Tags -->
          <div class="flex items-center gap-2 flex-wrap mb-4">
            <span class="px-2 py-1 bg-indigo-50 text-indigo-600 rounded-full text-xs font-medium">{{ post.company }}</span>
            <span v-for="tag in post.tags" :key="tag" class="px-2 py-1 bg-zinc-100 text-zinc-600 rounded-full text-xs">{{ tag }}</span>
          </div>

          <!-- Actions -->
          <div class="flex items-center justify-between pt-4 border-t border-zinc-100">
            <div class="flex items-center gap-4 text-sm text-zinc-400">
              <button class="flex items-center gap-1 hover:text-indigo-600 transition-colors">
                <ThumbsUp class="h-4 w-4" /> {{ post.likes }}
              </button>
              <button class="flex items-center gap-1 hover:text-indigo-600 transition-colors">
                <MessageCircle class="h-4 w-4" /> {{ post.comments }}
              </button>
              <span class="flex items-center gap-1">
                <Eye class="h-4 w-4" /> {{ post.views }}
              </span>
            </div>
            <button class="px-3 py-1.5 bg-indigo-50 text-indigo-600 rounded-lg text-xs font-medium hover:bg-indigo-100 transition-colors">
              查看详情
            </button>
          </div>
        </div>
      </div>

      <!-- Right Sidebar -->
      <div class="space-y-6">
        <!-- Book Alumni -->
        <div class="bg-gradient-to-br from-indigo-600 to-violet-600 rounded-3xl p-6 text-white">
          <h3 class="font-bold mb-2">预约校友 1v1 模拟面试</h3>
          <p class="text-indigo-200 text-sm mb-4">与已就业校友进行模拟面试，获取一手行业经验</p>
          <button @click="showBookingModal = true" class="w-full py-3 bg-white text-indigo-600 rounded-xl text-sm font-bold hover:bg-indigo-50 transition-colors">
            立即预约
          </button>
        </div>

        <!-- Top Alumni -->
        <div class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm">
          <h3 class="font-bold text-zinc-900 mb-4 flex items-center gap-2">
            <Award class="h-4 w-4 text-amber-500" />
            活跃校友
          </h3>
          <div class="space-y-3">
            <div v-for="(alumni, idx) in topAlumni" :key="idx" class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="h-8 w-8 rounded-full bg-zinc-100 text-zinc-600 flex items-center justify-center font-bold text-xs">{{ alumni.name.charAt(0) }}</div>
                <div>
                  <div class="text-sm font-medium text-zinc-900">{{ alumni.name }}</div>
                  <div class="text-xs text-zinc-400">{{ alumni.company }}</div>
                </div>
              </div>
              <span class="text-xs text-zinc-400">{{ alumni.posts }}篇</span>
            </div>
          </div>
        </div>

        <!-- Hot Companies -->
        <div class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm">
          <h3 class="font-bold text-zinc-900 mb-4">热门公司面经</h3>
          <div class="flex flex-wrap gap-2">
            <span v-for="company in hotCompanies" :key="company"
              class="px-3 py-1.5 bg-zinc-100 text-zinc-600 rounded-full text-sm cursor-pointer hover:bg-indigo-50 hover:text-indigo-600 transition-colors"
              @click="filterCompany = company"
            >
              {{ company }}
            </span>
          </div>
        </div>

        <!-- AI Knowledge Base -->
        <div class="bg-zinc-900 text-white rounded-3xl p-6">
          <h3 class="font-bold mb-2">AI 知识库补充</h3>
          <p class="text-sm text-zinc-400 mb-4">校友经验自动融入 RAG 知识库，让 AI 面试反馈更具行业针对性</p>
          <div class="flex items-center gap-2 text-xs text-zinc-500">
            <div class="w-2 h-2 bg-emerald-500 rounded-full"></div>
            已收录 {{ totalExperiences }} 条面试经验
          </div>
        </div>
      </div>
    </div>

    <!-- Share Modal -->
    <div v-if="showShareModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50" @click.self="showShareModal = false">
      <div class="bg-white rounded-3xl p-8 w-full max-w-lg shadow-2xl">
        <h2 class="text-xl font-bold text-zinc-900 mb-6">分享面试经验</h2>
        <div class="space-y-4">
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase mb-1">标题</label>
            <input class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm" placeholder="例如：字节跳动前端面经分享" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="text-xs font-bold text-zinc-400 uppercase mb-1">公司</label>
              <input class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm" placeholder="公司名称" />
            </div>
            <div>
              <label class="text-xs font-bold text-zinc-400 uppercase mb-1">岗位</label>
              <input class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm" placeholder="岗位名称" />
            </div>
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase mb-1">面试经验内容</label>
            <textarea class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm h-32 resize-none" placeholder="分享你的面试过程、题目、心得..."></textarea>
          </div>
          <div class="flex items-center gap-3 justify-end">
            <button @click="showShareModal = false" class="px-4 py-2 text-zinc-500 hover:bg-zinc-100 rounded-lg transition-colors">取消</button>
            <button @click="showShareModal = false" class="px-6 py-2 bg-indigo-600 text-white rounded-xl font-medium hover:bg-indigo-700 transition-colors">发布</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Booking Modal -->
    <div v-if="showBookingModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50" @click.self="showBookingModal = false">
      <div class="bg-white rounded-3xl p-8 w-full max-w-lg shadow-2xl">
        <h2 class="text-xl font-bold text-zinc-900 mb-6">预约校友 1v1 模拟面试</h2>
        <div class="space-y-4">
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase mb-1">选择校友</label>
            <select class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm">
              <option v-for="a in topAlumni" :key="a.name" :value="a.name">{{ a.name }} ({{ a.company }})</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase mb-1">目标岗位</label>
            <input class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm" placeholder="输入目标岗位" />
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase mb-1">预约时间</label>
            <input type="datetime-local" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm" />
          </div>
          <div class="flex items-center gap-3 justify-end">
            <button @click="showBookingModal = false" class="px-4 py-2 text-zinc-500 hover:bg-zinc-100 rounded-lg transition-colors">取消</button>
            <button @click="showBookingModal = false" class="px-6 py-2 bg-indigo-600 text-white rounded-xl font-medium hover:bg-indigo-700 transition-colors">提交预约</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Search, ThumbsUp, MessageCircle, Eye, Award } from 'lucide-vue-next'

const searchQuery = ref('')
const filterType = ref('')
const filterCompany = ref('')
const showShareModal = ref(false)
const showBookingModal = ref(false)
const totalExperiences = ref(1247)

const posts = ref([
  {
    author: '张学长',
    school: '北京大学',
    year: '2025',
    position: 'Java后端工程师',
    company: '字节跳动',
    title: '字节跳动 Java后端三面面经 - 已拿Offer',
    content: '一面主要考察基础知识，包括 JVM 内存模型、多线程并发、Spring Boot 源码等。二面深入项目经验，问了很多分布式系统相关的问题。三面是 HR 面，主要聊职业规划和团队合作经验。整体难度中等偏上，建议重点准备分布式和微服务相关知识。',
    tags: ['Java', '分布式', '微服务'],
    likes: 234,
    comments: 45,
    views: 1890,
    verified: true,
  },
  {
    author: '李学姐',
    school: '清华大学',
    year: '2025',
    position: '前端工程师',
    company: '腾讯',
    title: '腾讯 WXG 前端面试全记录',
    content: '面试主要考察了 React 源码理解、性能优化方案、TypeScript 高级类型、以及现场手写了一个虚拟列表组件。注意准备好项目中的技术难点，面试官会深度追问。',
    tags: ['React', 'TypeScript', '性能优化'],
    likes: 189,
    comments: 32,
    views: 1456,
    verified: true,
  },
  {
    author: '王同学',
    school: '浙江大学',
    year: '2026',
    position: 'AI工程师',
    company: '百度',
    title: '百度 AI Lab 实习面试经验分享',
    content: '面试考了 Transformer 原理及变种、CNN 各种结构的优缺点、目标检测领域的经典论文。编程题是 LeetCode 中等难度。建议准备好论文阅读笔记，面试官会就某篇具体论文深入讨论。',
    tags: ['深度学习', 'Transformer', 'CV'],
    likes: 156,
    comments: 28,
    views: 1234,
    verified: false,
  },
  {
    author: '赵学长',
    school: '上海交大',
    year: '2024',
    position: '产品经理',
    company: '阿里巴巴',
    title: '阿里产品经理校招面经 + 群面技巧',
    content: '群面环节采用的是无领导小组讨论形式，题目是关于某个互联网产品的市场策略选择。建议在群面中找好自己的角色定位，既要有观点输出，又要善于总结和推进讨论。个面主要是产品分析和数据思维。',
    tags: ['产品设计', '群面', '数据分析'],
    likes: 98,
    comments: 19,
    views: 876,
    verified: true,
  },
])

const topAlumni = ref([
  { name: '张学长', company: '字节跳动', posts: 8 },
  { name: '李学姐', company: '腾讯', posts: 6 },
  { name: '赵学长', company: '阿里巴巴', posts: 5 },
  { name: '周学姐', company: '华为', posts: 4 },
])

const hotCompanies = ['字节跳动', '腾讯', '阿里巴巴', '华为', '百度', '美团', '京东', '网易']

const filteredPosts = computed(() => {
  return posts.value.filter(p => {
    if (searchQuery.value) {
      const q = searchQuery.value.toLowerCase()
      if (!p.title.toLowerCase().includes(q) && !p.company.toLowerCase().includes(q) && !p.tags.join(',').toLowerCase().includes(q)) return false
    }
    if (filterCompany.value && p.company !== filterCompany.value) return false
    return true
  })
})
</script>
