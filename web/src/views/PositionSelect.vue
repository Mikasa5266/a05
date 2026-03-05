<template>
  <div class="position-select">
    <el-card class="select-card">
      <template #header>
        <div class="card-header">
          <span>选择面试岗位</span>
        </div>
      </template>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="岗位名称" prop="position">
          <el-select v-model="form.position" placeholder="请选择岗位">
            <el-option label="Java工程师" value="Java" />
            <el-option label="Python工程师" value="Python" />
            <el-option label="前端工程师" value="Frontend" />
            <el-option label="Go工程师" value="Go" />
            <el-option label="测试工程师" value="QA" />
          </el-select>
        </el-form-item>
        <el-form-item label="难度级别" prop="difficulty">
          <el-radio-group v-model="form.difficulty">
            <el-radio label="Junior">初级</el-radio>
            <el-radio label="Middle">中级</el-radio>
            <el-radio label="Senior">高级</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleStart" :loading="loading">开始面试</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useInterviewStore } from '../stores/interview'
import { ElMessage } from 'element-plus'

const router = useRouter()
const interviewStore = useInterviewStore()

const form = reactive({
  position: '',
  difficulty: 'Junior'
})

const rules = {
  position: [{ required: true, message: '请选择岗位', trigger: 'change' }],
  difficulty: [{ required: true, message: '请选择难度', trigger: 'change' }]
}

const formRef = ref(null)
const loading = ref(false)

const handleStart = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        await interviewStore.start(form)
        router.push(`/student/interview/${interviewStore.interview.id}`)
      } catch (error) {
        ElMessage.error(error.response?.data?.error || '开始面试失败')
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
.position-select {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 80vh;
}

.select-card {
  width: 500px;
}

.card-header {
  text-align: center;
  font-size: 18px;
  font-weight: bold;
}
</style>