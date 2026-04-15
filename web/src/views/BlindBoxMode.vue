<template>
  <section class="blindbox-shell">
    <div class="nebula nebula-a" />
    <div class="nebula nebula-b" />
    <div class="nebula nebula-c" />

    <article class="blindbox-card">
      <p class="blindbox-eyebrow">Blind Box Interview</p>
      <h1>盲盒面试模式</h1>
      <p class="blindbox-desc">
        系统将随机抽取岗位与压力场景，进入未知任务链路。先准备心态，再开启本次沉浸式挑战。
      </p>

      <div class="blindbox-features">
        <span>随机岗位</span>
        <span>压力情境</span>
        <span>限时问答</span>
      </div>

      <div class="blindbox-actions">
        <button type="button" class="ghost-btn" @click="goHome">
          返回首页
        </button>
        <button type="button" class="primary-btn" @click="startBlindBox">
          开始抽取盲盒
        </button>
      </div>
    </article>
  </section>
</template>

<script setup>
import { useRouter } from "vue-router";

const router = useRouter();

const goHome = () => {
  router.push("/student/dashboard");
};

const startBlindBox = () => {
  router.push({
    path: "/interview/standard/setup",
    query: {
      mode: "blindbox",
      style: "gentle",
      interviewMode: "ai",
      presentationMode: "text_voice",
    },
  });
};
</script>

<style scoped>
.blindbox-shell {
  position: relative;
  min-height: calc(100vh - 220px);
  border-radius: 32px;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.26);
  background:
    radial-gradient(
      circle at 18% 18%,
      rgba(56, 189, 248, 0.2),
      transparent 33%
    ),
    radial-gradient(
      circle at 84% 22%,
      rgba(99, 102, 241, 0.14),
      transparent 36%
    ),
    radial-gradient(
      circle at 56% 86%,
      rgba(16, 185, 129, 0.16),
      transparent 38%
    ),
    linear-gradient(150deg, #f8fafc 0%, #f1f5f9 48%, #eef2ff 100%);
  display: grid;
  place-items: center;
  padding: 24px;
}

.nebula {
  position: absolute;
  border-radius: 999px;
  filter: blur(26px);
  opacity: 0.42;
  animation: floatNebula 9s ease-in-out infinite;
}

.nebula-a {
  width: 220px;
  height: 220px;
  top: 12%;
  left: 10%;
  background: rgba(16, 185, 129, 0.28);
}

.nebula-b {
  width: 280px;
  height: 280px;
  right: 8%;
  top: 20%;
  background: rgba(99, 102, 241, 0.26);
  animation-delay: -2.5s;
}

.nebula-c {
  width: 220px;
  height: 220px;
  bottom: 8%;
  left: 35%;
  background: rgba(56, 189, 248, 0.24);
  animation-delay: -5s;
}

.blindbox-card {
  position: relative;
  z-index: 2;
  width: min(760px, 100%);
  border-radius: 28px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  background: linear-gradient(
    180deg,
    rgba(255, 255, 255, 0.92),
    rgba(248, 250, 252, 0.88)
  );
  backdrop-filter: blur(10px);
  padding: 32px;
  color: #0f172a;
  box-shadow: 0 20px 46px rgba(71, 85, 105, 0.2);
  animation: panelRise 0.45s ease;
}

.blindbox-eyebrow {
  margin: 0;
  font-size: 12px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: #0284c7;
}

.blindbox-card h1 {
  margin: 14px 0 0;
  font-size: clamp(30px, 5vw, 46px);
  line-height: 1.08;
  color: #0f172a;
}

.blindbox-desc {
  margin: 18px 0 0;
  font-size: 15px;
  line-height: 1.85;
  color: #475569;
}

.blindbox-features {
  margin-top: 22px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.blindbox-features span {
  border-radius: 999px;
  border: 1px solid rgba(59, 130, 246, 0.2);
  background: rgba(239, 246, 255, 0.95);
  padding: 8px 12px;
  font-size: 12px;
  color: #0c4a6e;
}

.blindbox-actions {
  margin-top: 30px;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.primary-btn,
.ghost-btn {
  border: 0;
  border-radius: 14px;
  padding: 11px 18px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition:
    transform 180ms ease,
    box-shadow 180ms ease,
    opacity 180ms ease;
}

.primary-btn {
  color: #ffffff;
  background: linear-gradient(120deg, #0284c7 0%, #4f46e5 100%);
  box-shadow: 0 14px 24px rgba(79, 70, 229, 0.25);
}

.ghost-btn {
  color: #0f172a;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(148, 163, 184, 0.34);
}

.primary-btn:hover,
.ghost-btn:hover {
  transform: translateY(-1px);
}

@keyframes floatNebula {
  0%,
  100% {
    transform: translateY(0) scale(1);
  }
  50% {
    transform: translateY(-16px) scale(1.04);
  }
}

@keyframes panelRise {
  from {
    opacity: 0;
    transform: translateY(14px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 767px) {
  .blindbox-shell {
    min-height: calc(100vh - 170px);
    border-radius: 24px;
  }

  .blindbox-card {
    border-radius: 22px;
    padding: 24px;
  }

  .blindbox-actions {
    flex-direction: column;
  }

  .blindbox-actions button {
    width: 100%;
  }
}
</style>
