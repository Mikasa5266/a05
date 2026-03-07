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
        <option value="百度">百度</option>
        <option value="阿里巴巴">阿里巴巴</option>
        <option value="腾讯">腾讯</option>
        <option value="华为">华为</option>
        <option value="字节跳动">字节跳动</option>
      </select>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Posts List -->
      <div class="lg:col-span-2 space-y-6">
        <div v-if="postsLoading" class="bg-white rounded-3xl p-8 border border-zinc-100 text-zinc-500">
          加载中...
        </div>
        <div v-else-if="postsError" class="bg-rose-50 rounded-3xl p-8 border border-rose-100 text-rose-700">
          {{ postsError }}
        </div>
        <div v-else-if="filteredPosts.length === 0" class="bg-white rounded-3xl p-8 border border-zinc-100 text-zinc-400">
          暂无帖子，快来发布第一篇面经吧
        </div>

        <div v-for="post in filteredPosts" :key="post.id"
          class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm hover:shadow-md transition-shadow"
        >
          <!-- Author Info -->
          <div class="flex items-center gap-3 mb-4">
            <img v-if="post.avatar" :src="post.avatar" class="h-10 w-10 rounded-full object-cover" />
            <div v-else class="h-10 w-10 rounded-full bg-indigo-100 text-indigo-600 flex items-center justify-center font-bold text-sm">
              {{ (post.author || '?').charAt(0) }}
            </div>
            <div>
              <div class="font-medium text-zinc-900 flex items-center gap-2">
                {{ post.author || '匿名用户' }}
              </div>
              <div class="text-xs text-zinc-400">{{ post.position || '未填写岗位' }}</div>
            </div>
          </div>

          <!-- Post Content -->
          <h3 class="text-lg font-bold text-zinc-900 mb-2">{{ post.title }}</h3>
          <p class="text-sm text-zinc-600 leading-relaxed mb-4 line-clamp-3">{{ post.content }}</p>

          <!-- Tags -->
          <div class="flex items-center gap-2 flex-wrap mb-4">
            <span v-if="post.is_indexed" class="px-2 py-1 bg-emerald-50 text-emerald-600 rounded-full text-[10px] font-bold border border-emerald-100 flex items-center gap-1">
              <BrainCircuit class="h-3 w-3" /> 已入库
            </span>
            <span v-if="post.company" class="px-2 py-1 bg-indigo-50 text-indigo-600 rounded-full text-xs font-medium">{{ post.company }}</span>
            <span v-for="tag in (post.tags || [])" :key="tag" class="px-2 py-1 bg-zinc-100 text-zinc-600 rounded-full text-xs">{{ tag }}</span>
          </div>

          <!-- Actions -->
          <div class="flex items-center justify-between pt-4 border-t border-zinc-100">
            <div class="flex items-center gap-4 text-sm text-zinc-400">
              <button @click="onLike(post)" class="flex items-center gap-1 hover:text-indigo-600 transition-colors">
                <ThumbsUp class="h-4 w-4" /> {{ post.likes }}
              </button>
              <button class="flex items-center gap-1 hover:text-indigo-600 transition-colors">
                <MessageCircle class="h-4 w-4" /> {{ post.comments }}
              </button>
              <span class="flex items-center gap-1">
                <Eye class="h-4 w-4" /> {{ post.views }}
              </span>
            </div>
            <div class="flex items-center gap-2">
              <button v-if="userStore.userInfo && post.user_id === userStore.userInfo.id" 
                @click.stop="handleDelete(post.id)" 
                class="px-3 py-1.5 text-rose-600 hover:bg-rose-50 rounded-lg text-xs font-bold transition-all flex items-center gap-1"
              >
                <Trash2 class="h-3.5 w-3.5" />
                删除
              </button>
              <button @click="goDetail(post.id)" class="px-3 py-1.5 bg-indigo-50 text-indigo-600 rounded-lg text-xs font-medium hover:bg-indigo-100 transition-colors">
                查看详情
              </button>
            </div>
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
                <img v-if="alumni.avatar" :src="alumni.avatar" class="h-8 w-8 rounded-full object-cover" />
                <div v-else class="h-8 w-8 rounded-full bg-zinc-100 text-zinc-600 flex items-center justify-center font-bold text-xs">{{ (alumni.name || '?').charAt(0) }}</div>
                <div>
                  <div class="text-sm font-medium text-zinc-900">{{ alumni.name || '匿名' }}</div>
                  <div class="text-xs text-zinc-400">{{ alumni.company || '未填写' }}</div>
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
            <span v-for="company in hotCompanies" :key="company.name"
              class="px-3 py-1.5 bg-zinc-100 text-zinc-600 rounded-full text-sm cursor-pointer hover:bg-indigo-50 hover:text-indigo-600 transition-colors"
              @click="filterCompany = company.name"
            >
              {{ company.name }}
              <span class="text-xs text-zinc-400 ml-1">{{ company.posts }}</span>
            </span>
          </div>
        </div>

        <!-- AI Knowledge Base -->
        <div class="bg-zinc-900 text-white rounded-3xl p-6">
          <h3 class="font-bold mb-2">AI 知识库查询</h3>
          <p class="text-sm text-zinc-400 mb-4">校友经验已自动融入 RAG 知识库，您可以在此查询收录情况</p>
          
          <div class="relative mb-4">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500" />
            <input 
              v-model="ragSearchQuery" 
              @keyup.enter="handleRagSearch"
              placeholder="搜索知识库中的面经..."
              class="w-full pl-9 pr-3 py-2 bg-white/10 border border-white/10 rounded-xl text-xs focus:outline-none focus:ring-1 focus:ring-indigo-500 transition-all" 
            />
          </div>

          <div v-if="ragSearching" class="text-xs text-zinc-500 animate-pulse">搜索中...</div>
          <div v-else-if="ragResults.length > 0" class="space-y-3 mb-4 max-h-60 overflow-y-auto pr-1 custom-scrollbar">
            <div v-for="(res, idx) in ragResults" :key="idx" class="p-3 bg-white/5 rounded-xl border border-white/5 hover:bg-white/10 transition-colors">
              <div class="text-[10px] text-indigo-400 font-bold uppercase mb-1">{{ res.category === 'community' ? '校友贡献' : '官方知识' }}</div>
              <div class="text-xs font-medium text-zinc-200 line-clamp-2 mb-1">{{ res.content.substring(0, 100) }}...</div>
              <div class="text-[10px] text-zinc-500">来源: {{ formatSource(res.title) }}</div>
            </div>
          </div>
          <div v-else-if="ragHasSearched" class="text-xs text-zinc-500 mb-4 text-center py-2 bg-white/5 rounded-xl">未找到相关收录内容</div>

          <div class="flex items-center gap-2 text-xs text-zinc-500 border-t border-white/5 pt-4">
            <div class="w-2 h-2 bg-emerald-500 rounded-full"></div>
            已收录 {{ totalExperiences }} 条优质面试经验
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
            <input v-model="shareForm.title" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm" placeholder="例如：字节跳动前端三面面经（含高频题+复盘）" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="text-xs font-bold text-zinc-400 uppercase mb-1">公司</label>
              <input v-model="shareForm.company" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm" placeholder="例如：字节跳动" />
            </div>
            <div>
              <label class="text-xs font-bold text-zinc-400 uppercase mb-1">岗位</label>
              <input v-model="shareForm.position" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm" placeholder="例如：前端工程师 / Java后端" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="text-xs font-bold text-zinc-400 uppercase mb-1">面试难度 (1-5)</label>
              <select v-model="shareForm.difficulty" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm">
                <option :value="1">1 - 很简单</option>
                <option :value="2">2 - 简单</option>
                <option :value="3">3 - 一般</option>
                <option :value="4">4 - 困难</option>
                <option :value="5">5 - 很难</option>
              </select>
            </div>
            <div>
              <label class="text-xs font-bold text-zinc-400 uppercase mb-1">面试结果</label>
              <select v-model="shareForm.offerStatus" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm">
                <option value="Pending">等待中</option>
                <option value="Received">已拿Offer</option>
                <option value="Rejected">未通过</option>
              </select>
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
             <div>
              <label class="text-xs font-bold text-zinc-400 uppercase mb-1">面试轮数</label>
              <input v-model.number="shareForm.rounds" type="number" min="1" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm" />
            </div>
            <div>
              <label class="text-xs font-bold text-zinc-400 uppercase mb-1">面试时间</label>
              <input v-model="shareForm.interviewDate" type="date" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm" />
            </div>
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase mb-1">标签（可选）</label>
            <input v-model="shareForm.tagsInput" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm" placeholder="例如：React,TypeScript,性能优化（逗号分隔）" />
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase mb-1">面试流程（可选）</label>
            <textarea v-model="shareForm.process" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm h-24 resize-none" placeholder="例如：一面基础+手撕；二面项目深挖；三面系统设计/HR..." />
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase mb-1">高频问题（可选）</label>
            <textarea v-model="shareForm.questions" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm h-24 resize-none" placeholder="例如：1）React diff；2）浏览器渲染；3）缓存策略..." />
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase mb-1">复盘与建议（可选）</label>
            <textarea v-model="shareForm.review" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm h-24 resize-none" placeholder="例如：哪些地方回答不足、如何改进、推荐资料..." />
          </div>
          <div class="flex items-center gap-3 justify-end">
            <button @click="showShareModal = false" class="px-4 py-2 text-zinc-500 hover:bg-zinc-100 rounded-lg transition-colors">取消</button>
            <button @click="submitShare" :disabled="shareSubmitting || !shareForm.title.trim()" class="px-6 py-2 bg-indigo-600 text-white rounded-xl font-medium hover:bg-indigo-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
              {{ shareSubmitting ? '发布中...' : '发布' }}
            </button>
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
import { ref, computed, onMounted, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Search, ThumbsUp, MessageCircle, Eye, Award, BrainCircuit, Trash2 } from 'lucide-vue-next'
import { getPosts, createPost, likePost, getTopAlumni, getHotCompanies, queryKnowledgeBase, deletePost } from '../api/community'
import { useUserStore } from '../stores/user'

const userStore = useUserStore()
const router = useRouter()
const searchQuery = ref('')
const filterType = ref('')
const filterCompany = ref('')
const showShareModal = ref(false)
const showBookingModal = ref(false)
const totalExperiences = ref(1247)

const ragSearchQuery = ref('')
const ragSearching = ref(false)
const ragResults = ref([])
const ragHasSearched = ref(false)

const posts = ref([])
const postsLoading = ref(false)
const postsError = ref('')

const shareForm = reactive({
  title: '',
  company: '',
  position: '',
  tagsInput: '',
  difficulty: 3,
  offerStatus: 'Pending',
  rounds: 1,
  interviewDate: '',
  process: '',
  questions: '',
  review: ''
})
const shareSubmitting = ref(false)

const topAlumni = ref([])
const hotCompanies = ref([])

const normalizeTags = (val) => {
  if (!val) return []
  if (Array.isArray(val)) return val.map(s => String(s).trim()).filter(Boolean)
  return String(val).split(',').map(s => s.trim()).filter(Boolean)
}

const formatSource = (title) => {
  if (!title) return '未知来源'
  if (title.includes('post_')) return '社区贡献'
  return title.replace('.md', '')
}

const handleRagSearch = async () => {
  if (!ragSearchQuery.value.trim()) return
  ragSearching.value = true
  ragHasSearched.value = true
  try {
    const res = await queryKnowledgeBase({ query: ragSearchQuery.value })
    ragResults.value = res.sources || []
  } catch (e) {
    console.error('RAG Search failed:', e)
  } finally {
    ragSearching.value = false
  }
}

const fetchPosts = async () => {
  postsLoading.value = true
  postsError.value = ''
  try {
    const res = await getPosts({
      search: searchQuery.value || undefined,
      company: filterCompany.value || undefined,
      page: 1,
      page_size: 20
    })
    const list = res.posts || []
    posts.value = list.map((p) => ({
      ...p,
      tags: normalizeTags(p.tags)
    }))
    totalExperiences.value = typeof res.total === 'number' ? res.total : totalExperiences.value
  } catch (e) {
    posts.value = []
    postsError.value = e?.response?.data?.error || e?.message || '加载失败'
  } finally {
    postsLoading.value = false
  }
}

const filteredPosts = computed(() => {
  return posts.value
})

const buildContentFromForm = () => {
  const lines = []
  lines.push('【背景】')
  if (shareForm.company.trim()) lines.push(`公司：${shareForm.company.trim()}`)
  if (shareForm.position.trim()) lines.push(`岗位：${shareForm.position.trim()}`)
  lines.push('')
  if (shareForm.process.trim()) {
    lines.push('【流程】')
    lines.push(shareForm.process.trim())
    lines.push('')
  }
  if (shareForm.questions.trim()) {
    lines.push('【高频问题】')
    lines.push(shareForm.questions.trim())
    lines.push('')
  }
  if (shareForm.review.trim()) {
    lines.push('【复盘与建议】')
    lines.push(shareForm.review.trim())
    lines.push('')
  }
  return lines.join('\n')
}

const submitShare = async () => {
  if (shareSubmitting.value) return
  const title = shareForm.title.trim()
  if (!title) return
  const tags = normalizeTags(shareForm.tagsInput)
  const content = buildContentFromForm()
  shareSubmitting.value = true
  try {
    await createPost({
      title,
      company: shareForm.company.trim(),
      position: shareForm.position.trim(),
      tags,
      content,
      process: shareForm.process.trim(),
      questions: shareForm.questions.trim(),
      review: shareForm.review.trim(),
      difficulty: Number(shareForm.difficulty),
      offer_status: shareForm.offerStatus,
      rounds: Number(shareForm.rounds),
      interview_date: shareForm.interviewDate ? new Date(shareForm.interviewDate).toISOString() : null
    })
    showShareModal.value = false
    shareForm.title = ''
    shareForm.company = ''
    shareForm.position = ''
    shareForm.tagsInput = ''
    shareForm.difficulty = 3
    shareForm.offerStatus = 'Pending'
    shareForm.rounds = 1
    shareForm.interviewDate = ''
    shareForm.process = ''
    shareForm.questions = ''
    shareForm.review = ''
    await fetchPosts()
  } catch (e) {
    alert('发布失败：' + (e?.response?.data?.error || e?.message || '未知错误'))
  } finally {
    shareSubmitting.value = false
  }
}

const onLike = async (post) => {
  if (!post?.id) return
  try {
    const res = await likePost(post.id)
    if (typeof res.likes === 'number') post.likes = res.likes
  } catch (_) {}
}

const goDetail = (id) => {
  router.push(`/student/community/posts/${id}`)
}

const handleDelete = async (id) => {
  if (!confirm('确定要删除这篇面经吗？知识库中的相关内容也将被移除。')) return
  try {
    await deletePost(id)
    posts.value = posts.value.filter(p => p.id !== id)
    fetchTopAlumni()
    fetchHotCompanies()
  } catch (e) {
    alert(e?.response?.data?.error || '删除失败')
  }
}

const fetchTopAlumni = async () => {
  try {
    const res = await getTopAlumni()
    topAlumni.value = res.alumni || []
  } catch (_) {
    topAlumni.value = []
  }
}

const fetchHotCompanies = async () => {
  try {
    const res = await getHotCompanies()
    hotCompanies.value = res.companies || []
  } catch (_) {
    hotCompanies.value = []
  }
}

let debounceTimer = null
watch([searchQuery, filterCompany], () => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    fetchPosts()
  }, 250)
})

onMounted(() => {
  fetchPosts()
  fetchTopAlumni()
  fetchHotCompanies()
})
</script>
