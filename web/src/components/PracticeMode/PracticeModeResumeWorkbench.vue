<template>
  <section class="resume-workbench-grid">
    <article class="module-card theme-structured">
      <header class="module-head">
        <h4>结构化提取</h4>
      </header>
      <div class="module-body"></div>
    </article>

    <article class="module-card theme-match">
      <header class="module-head">
        <h4>岗位匹配分析</h4>
      </header>
      <div class="module-body">
        <canvas width="420" height="240"></canvas>
      </div>
    </article>

    <article class="module-card theme-plan module-focus">
      <header class="module-head">
        <h4>📈 岗位定制刷题计划与优化建议</h4>
      </header>
      <div class="module-body"></div>
    </article>

    <article class="module-card theme-trace">
      <header class="module-head">
        <h4>Agentic RAG + MoE 决策轨迹</h4>
      </header>
      <div class="module-body"></div>
    </article>
  </section>
</template>

<style scoped>
.resume-workbench-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
}

.module-card {
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  min-height: 400px;
  aspect-ratio: 1 / 1;
  padding: 24px;
  border: 1px solid #cfe0ee;
  overflow: hidden;
  display: grid;
  grid-template-rows: auto 1fr;
  gap: 12px;
  background: #ffffff;
}

.module-focus {
  min-height: 440px;
  transform: translateY(-4px);
}

.module-head h4 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  line-height: 1.3;
  color: #1b6f95;
}

.module-focus .module-head h4 {
  font-size: 18px;
  font-weight: 700;
}

.module-body {
  overflow: auto;
  padding-right: 4px;
  line-height: 1.7;
  color: #1f2f3d;
}

.module-body::-webkit-scrollbar {
  width: 8px;
}

.module-body::-webkit-scrollbar-thumb {
  background: #bfd7e8;
  border-radius: 999px;
}

.theme-structured,
.theme-match,
.theme-plan,
.theme-trace {
  background: #ffffff;
  border-color: #cfe0ee;
}

canvas {
  margin-top: 8px;
  background: #fff;
  border-radius: 10px;
  border: 1px solid rgba(27, 111, 149, 0.22);
}

@media (max-width: 1180px) {
  .resume-workbench-grid {
    grid-template-columns: 1fr;
  }

  .module-card,
  .module-focus {
    aspect-ratio: auto;
    min-height: 400px;
    transform: none;
  }
}
</style>
