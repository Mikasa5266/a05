<script setup>
defineProps({
  active: {
    type: Boolean,
    default: false,
  },
});
</script>

<template>
  <section class="view" :class="{ active }">
    <div class="stat-cards">
      <div class="stat-card">
        <span>总刷题量</span>
        <strong>0</strong>
      </div>
      <div class="stat-card">
        <span>整体正确率</span>
        <strong>0%</strong>
      </div>
      <div class="stat-card">
        <span>当前岗位进度</span>
        <strong>0%</strong>
      </div>
    </div>

    <div class="charts">
      <div class="chart-box">
        <h3>近 14 天刷题趋势</h3>
        <canvas width="740" height="260"></canvas>
      </div>
      <div class="chart-box">
        <h3>岗位能力雷达图</h3>
        <canvas width="740" height="420"></canvas>
      </div>
    </div>
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

.stat-cards {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  margin-bottom: 10px;
}

.stat-card {
  border: 1px solid var(--line);
  border-radius: 14px;
  background: #fff;
  padding: 14px;
}

.stat-card span {
  color: #57728f;
  font-size: 13px;
}

.stat-card strong {
  display: block;
  margin-top: 8px;
  color: #173c60;
  font-size: 28px;
}

.charts {
  display: grid;
  gap: 10px;
}

.chart-box {
  border: 1px solid var(--line);
  border-radius: var(--radius);
  background: #fff;
  padding: 12px;
  box-shadow: 0 10px 22px rgba(23, 75, 106, 0.06);
}

.chart-box h3 {
  margin: 0 0 8px;
  color: #173c60;
  font-size: 15px;
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

@media (max-width: 880px) {
  .stat-cards {
    grid-template-columns: 1fr;
  }
}
</style>
