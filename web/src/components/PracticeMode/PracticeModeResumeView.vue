<script setup>
import PracticeModeResumeTypeModal from "./PracticeModeResumeTypeModal.vue";
import PracticeModeResumeWorkbench from "./PracticeModeResumeWorkbench.vue";

defineProps({
  active: {
    type: Boolean,
    default: false,
  },
});
</script>

<template>
  <section class="view" :class="{ active }">
    <div class="resume-shell">
      <section class="resume-topbar-card">
        <div class="resume-topbar-left">
          <h3>多格式简历智能解析工作台</h3>
          <p>多格式解析、岗位匹配、定制刷题计划与闭环提升一体化。</p>
        </div>

        <div class="resume-upload-row">
          <input type="file" accept=".pdf,.docx,.txt,.png,.jpg,.jpeg" />
          <button type="button">开始解析与匹配</button>
        </div>

        <p class="resume-upload-hint">
          支持可复制文本PDF/Word、扫描件/图片，系统自动OCR识别
        </p>

        <div class="resume-action-row">
          <button type="button" class="secondary">一键OCR转换</button>
        </div>

        <div class="resume-kpi-strip">
          <div class="resume-kpi-card">
            <span>目标岗位</span>
            <strong>-</strong>
          </div>
          <div class="resume-kpi-card">
            <span>匹配度</span>
            <strong>0%</strong>
          </div>
          <div class="resume-kpi-card">
            <span>短板数</span>
            <strong>0</strong>
          </div>
        </div>

        <p class="resume-status">等待上传简历文件</p>
      </section>

      <PracticeModeResumeWorkbench />
    </div>

    <PracticeModeResumeTypeModal />
  </section>
</template>

<style scoped>
.view {
  display: none;
  animation: reveal 0.24s ease;
}

.view.active {
  display: block;
}

.resume-shell {
  display: grid;
  gap: 20px;
}

.resume-topbar-card {
  border: 1px solid #d8e6f2;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  padding: 24px;
  background: #fff;
  display: grid;
  gap: 14px;
}

.resume-topbar-left h3 {
  margin: 0;
  font-size: 24px;
  line-height: 1.2;
  color: #1b2f44;
}

.resume-topbar-left p {
  margin: 8px 0 0;
  font-size: 14px;
  line-height: 1.6;
  color: #4a6178;
}

.resume-upload-row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 10px;
}

.resume-upload-row button {
  background: #1677ff;
}

.resume-upload-hint {
  margin: -4px 0 0;
  font-size: 13px;
  line-height: 1.6;
  color: #4a6178;
}

.resume-action-row {
  display: flex;
  justify-content: flex-start;
}

.resume-kpi-strip {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.resume-kpi-card {
  border-radius: 12px;
  padding: 12px;
  border: 1px solid #d8e6f3;
  background: #f7fbff;
  display: grid;
  gap: 4px;
}

.resume-kpi-card span {
  font-size: 12px;
  font-weight: 700;
  color: #5d758b;
}

.resume-kpi-card strong {
  font-size: 20px;
  line-height: 1.2;
  color: #1c3f60;
}

.resume-status {
  margin: 0;
  font-size: 13px;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid #cfe0ee;
  background: #f2f8ff;
  color: #2a587a;
}

@keyframes reveal {
  from {
    opacity: 0;
    transform: translateY(4px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 1180px) {
  .resume-kpi-strip {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 880px) {
  .resume-upload-row {
    grid-template-columns: 1fr;
  }
}
</style>
