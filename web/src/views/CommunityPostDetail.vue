<template>
  <div class="space-y-6">
    <header class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <button @click="router.back()" class="px-4 py-2 rounded-xl border border-zinc-200 text-zinc-600 text-sm hover:bg-zinc-50 transition-colors">
          返回
        </button>
        <button v-if="post && userStore.userInfo && post.user_id === userStore.userInfo.id" 
          @click="handleDelete" 
          class="px-4 py-2 text-rose-600 hover:bg-rose-50 rounded-xl text-sm font-bold transition-all flex items-center gap-2"
        >
          <Trash2 class="h-4 w-4" />
          删除
        </button>
      </div>
      <div class="text-xs text-zinc-400">
        {{ post?.created_at ? new Date(post.created_at).toLocaleString('zh-CN') : '' }}
      </div>
    </header>

    <div v-if="loading" class="bg-white rounded-3xl p-8 border border-zinc-100 text-zinc-500">
      加载中...
    </div>

    <div v-else-if="error" class="bg-rose-50 rounded-3xl p-8 border border-rose-100 text-rose-700">
      {{ error }}
    </div>

    <div v-else class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-2 space-y-6">
        <div class="bg-white rounded-3xl p-7 border border-zinc-100 shadow-sm">
          <div class="flex items-center gap-3 mb-5">
            <img v-if="post.avatar" :src="post.avatar" class="h-10 w-10 rounded-full object-cover" />
            <div v-else class="h-10 w-10 rounded-full bg-indigo-100 text-indigo-600 flex items-center justify-center font-bold text-sm">
              {{ (post.author || '?').charAt(0) }}
            </div>
            <div class="min-w-0">
              <div class="font-medium text-zinc-900">{{ post.author || '匿名用户' }}</div>
              <div class="text-xs text-zinc-400">
                {{ post.company || '未填写公司' }} · {{ post.position || '未填写岗位' }}
              </div>
            </div>
          </div>

          <h1 class="text-2xl font-black text-zinc-900 leading-snug mb-4">{{ post.title }}</h1>

          <div class="flex flex-wrap gap-2 mb-6">
            <span v-if="post.is_indexed" class="px-2.5 py-1 bg-emerald-50 text-emerald-600 rounded-full text-xs font-bold border border-emerald-100 flex items-center gap-1.5">
              <BrainCircuit class="h-4 w-4" /> AI 知识库已收录
            </span>
            <span v-if="post.company" class="px-2.5 py-1 bg-indigo-50 text-indigo-700 rounded-full text-xs font-medium">
              {{ post.company }}
            </span>
            <span v-for="tag in (post.tags || [])" :key="tag" class="px-2.5 py-1 bg-zinc-100 text-zinc-600 rounded-full text-xs">
              {{ tag }}
            </span>
          </div>

          <!-- Structured Meta Info -->
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8 bg-zinc-50 rounded-2xl p-4 border border-zinc-100">
             <div>
               <div class="text-xs text-zinc-400 mb-1">面试难度</div>
               <div class="font-bold text-zinc-900">{{ post.difficulty ? post.difficulty + ' / 5' : '-' }}</div>
             </div>
             <div>
               <div class="text-xs text-zinc-400 mb-1">面试结果</div>
               <div class="font-bold text-zinc-900">
                 <span v-if="post.offer_status === 'Received'" class="text-emerald-600">已拿Offer</span>
                 <span v-else-if="post.offer_status === 'Rejected'" class="text-rose-600">未通过</span>
                 <span v-else-if="post.offer_status === 'Pending'" class="text-amber-600">等待中</span>
                 <span v-else>-</span>
               </div>
             </div>
             <div>
               <div class="text-xs text-zinc-400 mb-1">面试轮数</div>
               <div class="font-bold text-zinc-900">{{ post.rounds ? post.rounds + '轮' : '-' }}</div>
             </div>
             <div>
               <div class="text-xs text-zinc-400 mb-1">面试时间</div>
               <div class="font-bold text-zinc-900">{{ post.interview_date ? new Date(post.interview_date).toLocaleDateString() : '-' }}</div>
             </div>
          </div>

          <!-- Structured Content -->
          <div v-if="post.process || post.questions || post.review" class="space-y-8">
            <div v-if="post.process">
              <h3 class="text-lg font-bold text-zinc-900 mb-3 flex items-center gap-2">
                <div class="w-1 h-5 bg-indigo-600 rounded-full"></div>
                面试流程
              </h3>
              <div class="prose prose-zinc max-w-none">
                <pre class="whitespace-pre-wrap break-words text-sm text-zinc-700 leading-relaxed font-sans bg-transparent p-0 m-0">{{ post.process }}</pre>
              </div>
            </div>

            <div v-if="post.questions">
              <h3 class="text-lg font-bold text-zinc-900 mb-3 flex items-center gap-2">
                <div class="w-1 h-5 bg-indigo-600 rounded-full"></div>
                高频问题
              </h3>
              <div class="prose prose-zinc max-w-none">
                <pre class="whitespace-pre-wrap break-words text-sm text-zinc-700 leading-relaxed font-sans bg-transparent p-0 m-0">{{ post.questions }}</pre>
              </div>
            </div>

            <div v-if="post.review">
              <h3 class="text-lg font-bold text-zinc-900 mb-3 flex items-center gap-2">
                <div class="w-1 h-5 bg-indigo-600 rounded-full"></div>
                复盘与建议
              </h3>
              <div class="prose prose-zinc max-w-none">
                <pre class="whitespace-pre-wrap break-words text-sm text-zinc-700 leading-relaxed font-sans bg-transparent p-0 m-0">{{ post.review }}</pre>
              </div>
            </div>
          </div>

          <!-- Fallback Content -->
          <div v-else class="prose prose-zinc max-w-none">
            <pre class="whitespace-pre-wrap break-words text-sm text-zinc-700 leading-relaxed font-sans bg-transparent p-0 m-0">{{ post.content }}</pre>
          </div>

          <div class="flex items-center justify-between pt-6 mt-6 border-t border-zinc-100">
            <div class="flex items-center gap-4 text-sm text-zinc-400">
              <button @click="toggleLike" class="flex items-center gap-1 hover:text-indigo-600 transition-colors">
                <ThumbsUp class="h-4 w-4" /> {{ post.likes || 0 }}
              </button>
              <span class="flex items-center gap-1">
                <MessageCircle class="h-4 w-4" /> {{ post.comments || 0 }}
              </span>
              <span class="flex items-center gap-1">
                <Eye class="h-4 w-4" /> {{ post.views || 0 }}
              </span>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-3xl p-7 border border-zinc-100 shadow-sm">
          <div class="font-bold text-zinc-900 mb-4">评论</div>
          <div class="flex gap-3 mb-5">
            <textarea v-model="commentText" class="flex-1 px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-2xl text-sm h-24 resize-none" placeholder="写下你的看法或追问..." />
            <button @click="submitComment" :disabled="submittingComment || !commentText.trim()" class="px-5 py-3 bg-indigo-600 text-white rounded-2xl text-sm font-bold hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed">
              发布
            </button>
          </div>

          <div v-if="commentsLoading" class="text-sm text-zinc-400">评论加载中...</div>
          <div v-else-if="comments.length === 0" class="text-sm text-zinc-400">暂无评论</div>
          <div v-else class="space-y-3">
            <div v-for="c in comments" :key="c.id" class="p-4 rounded-2xl border border-zinc-100 bg-white">
              <div class="flex items-center justify-between mb-2">
                <div class="text-sm font-medium text-zinc-900">{{ c.author || '匿名用户' }}</div>
                <div class="text-xs text-zinc-400">{{ c.created_at ? new Date(c.created_at).toLocaleString('zh-CN') : '' }}</div>
              </div>
              <div class="text-sm text-zinc-700 whitespace-pre-wrap break-words leading-relaxed">{{ c.content }}</div>
            </div>
          </div>
        </div>
      </div>

      <div class="space-y-6">
        <div class="bg-zinc-900 text-white rounded-3xl p-6">
          <div class="text-sm font-bold mb-2">小提示</div>
          <p class="text-xs text-zinc-400 leading-relaxed">
            面经内容仅供参考。建议结合岗位 JD 与自身背景，选择性吸收，并在模拟面试里验证自己的表达与思路。
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ThumbsUp, MessageCircle, Eye, BrainCircuit, Trash2 } from 'lucide-vue-next'
import { getPost, likePost, getPostComments, commentOnPost, deletePost } from '../api/community'
import { useUserStore } from '../stores/user'

const userStore = useUserStore()
const route = useRoute()
const router = useRouter()

const loading = ref(true)
const error = ref('')
const post = ref(null)

const comments = ref([])
const commentsLoading = ref(false)
const commentText = ref('')
const submittingComment = ref(false)

const fetchPost = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await getPost(route.params.id)
    post.value = res.post
  } catch (e) {
    error.value = e?.response?.data?.error || e?.message || '加载失败'
  } finally {
    loading.value = false
  }
}

const fetchComments = async () => {
  commentsLoading.value = true
  try {
    const res = await getPostComments(route.params.id, { page: 1, page_size: 50 })
    comments.value = res.comments || []
  } catch (_) {
    comments.value = []
  } finally {
    commentsLoading.value = false
  }
}

const toggleLike = async () => {
  if (!post.value?.id) return
  try {
    const res = await likePost(post.value.id)
    if (typeof res.likes === 'number') post.value.likes = res.likes
  } catch (_) {}
}

const handleDelete = async () => {
  if (!confirm('确定要删除这篇面经吗？此操作不可撤销，知识库文件也会被永久移除。')) return
  try {
    await deletePost(post.value.id)
    router.push('/student/community')
  } catch (e) {
    alert(e?.response?.data?.error || '删除失败')
  }
}

const submitComment = async () => {
  if (!post.value?.id || !commentText.value.trim()) return
  submittingComment.value = true
  try {
    const res = await commentOnPost(post.value.id, { content: commentText.value.trim() })
    const c = res.comment
    if (c) {
      comments.value = [c, ...comments.value]
      post.value.comments = (post.value.comments || 0) + 1
    }
    commentText.value = ''
  } catch (_) {} finally {
    submittingComment.value = false
  }
}

onMounted(async () => {
  await fetchPost()
  await fetchComments()
})
</script>
