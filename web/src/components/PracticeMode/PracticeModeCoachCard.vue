<script setup>
import { usePracticeStore } from "../../stores/practice";

const practiceStore = usePracticeStore();
</script>

<template>
  <aside class="coach-card">
    <div class="coach-head">
      <div class="coach-avatar">AI</div>
      <div>
        <h4>影子教练</h4>
        <p>状态：{{ practiceStore.coach.mood }}</p>
      </div>
    </div>
    <p class="coach-hint">{{ practiceStore.coach.hint }}</p>
    <div class="coach-checklist">
      <span
        v-for="(item, index) in practiceStore.coach.checklist"
        :key="`${index}-${item}`"
      >
        {{ index + 1 }}. {{ item }}
      </span>
    </div>
    <button type="button" class="secondary" @click="practiceStore.manualCoachNudge">
      给我一条答题建议
    </button>
  </aside>
</template>

<style scoped>
.coach-card {
  position: sticky;
  top: 12px;
  border: 1px solid #cce4f2;
  border-radius: var(--radius);
  background: linear-gradient(145deg, #eef9ff 0%, #ffffff 52%, #f8f3ff 100%);
  padding: 14px;
  box-shadow: 0 16px 30px rgba(20, 74, 106, 0.12);
  animation: coachReveal 0.35s ease;
}

@keyframes coachReveal {
  from {
    opacity: 0;
    transform: translateY(6px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.coach-head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.coach-avatar {
  width: 40px;
  height: 40px;
  border-radius: 999px;
  display: grid;
  place-items: center;
  font-weight: 800;
  color: #fff;
  background: linear-gradient(135deg, #0f6e91 0%, #f08a4b 100%);
}

.coach-head h4 {
  margin: 0;
  color: #184662;
}

.coach-head p {
  margin: 2px 0 0;
  font-size: 12px;
  color: #5c7387;
}

.coach-hint {
  margin: 0;
  line-height: 1.6;
  font-size: 13px;
  color: #2e4c63;
  background: rgba(255, 255, 255, 0.72);
  border: 1px solid #d7e9f4;
  border-radius: 10px;
  padding: 10px;
}

.coach-checklist {
  margin: 10px 0;
  display: grid;
  gap: 6px;
}

.coach-checklist span {
  font-size: 12px;
  color: #47657b;
  border-left: 3px solid #1b789e;
  padding-left: 8px;
}

@media (max-width: 1180px) {
  .coach-card {
    position: static;
  }
}
</style>
