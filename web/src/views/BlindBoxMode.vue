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
  border: 1px solid rgba(158, 180, 205, 0.42);
  background:
    radial-gradient(
      circle at 18% 18%,
      rgba(121, 176, 239, 0.34),
      transparent 42%
    ),
    radial-gradient(
      circle at 84% 22%,
      rgba(156, 202, 255, 0.30),
      transparent 44%
    ),
    radial-gradient(
      circle at 56% 86%,
      rgba(147, 220, 195, 0.24),
      transparent 46%
    ),
    linear-gradient(
      150deg,
      #f4f9ff 0%,
      #e6f1fd 50%,
      #f2f8ff 100%
    );
  display: grid;
  place-items: center;
  padding: clamp(18px, 3.2vw, 36px);
}

.nebula {
  position: absolute;
  border-radius: 999px;
  filter: blur(26px);
  opacity: 0.24;
  animation: floatNebula 9s ease-in-out infinite;
}

.nebula-a {
  width: 220px;
  height: 220px;
  top: 12%;
  left: 10%;
  background: rgba(149, 219, 194, 0.48);
}

.nebula-b {
  width: 280px;
  height: 280px;
  right: 8%;
  top: 20%;
  background: rgba(146, 199, 255, 0.44);
  animation-delay: -2.5s;
}

.nebula-c {
  width: 220px;
  height: 220px;
  bottom: 8%;
  left: 35%;
  background: rgba(176, 213, 255, 0.42);
  animation-delay: -5s;
}

.blindbox-card {
  position: relative;
  z-index: 2;
  width: min(86%, 1240px);
  min-height: clamp(390px, 64vh, 700px);
  border-radius: 28px;
  border: 1px solid rgba(166, 188, 214, 0.45);
  background: linear-gradient(
    180deg,
    rgba(255, 255, 255, 0.92),
    rgba(245, 251, 255, 0.90)
  );
  backdrop-filter: blur(10px);
  padding: clamp(28px, 4vw, 54px);
  color: #1e293b;
  box-shadow: 0 24px 50px rgba(53, 96, 143, 0.16);
  animation: panelRise 0.45s ease;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.blindbox-eyebrow {
  margin: 0;
  font-size: 12px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: #3f78ba;
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
  color: #334155;
}

.blindbox-features {
  margin-top: 22px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.blindbox-features span {
  border-radius: 999px;
  border: 1px solid rgba(110, 165, 230, 0.35);
  background: rgba(230, 243, 255, 0.84);
  padding: 8px 12px;
  font-size: 12px;
  color: #2f5f96;
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
  background: linear-gradient(
    120deg,
    #3f8bdc 0%,
    #66a8ea 100%
  );
  box-shadow: 0 14px 24px rgba(63, 139, 220, 0.34);
}

.ghost-btn {
  color: #1f3652;
  background: rgba(245, 250, 255, 0.92);
  border: 1px solid rgba(166, 188, 214, 0.52);
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
    width: 100%;
    min-height: auto;
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
