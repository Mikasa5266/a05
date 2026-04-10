<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";

const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  status: {
    type: String,
    default: "准备就绪",
  },
  alert: {
    type: Boolean,
    default: false,
  },
});

const route = useRoute();
const router = useRouter();

const backLabel = computed(() => {
  const from = String(route.query?.from || "");
  if (from.includes("/student/practice-mode")) {
    return "返回刷题模式";
  }
  if (from.includes("/student/resume")) {
    return "返回简历中心";
  }
  if (from.includes("/student/growth")) {
    return "返回成长中心";
  }
  return "返回学生端";
});

const goBackToStudent = () => {
  const from = String(route.query?.from || "").trim();
  if (from && from.startsWith("/student/")) {
    router.push(from);
    return;
  }
  router.push("/student/dashboard");
};
</script>

<template>
  <header class="topbar">
    <h2>{{ props.title }}</h2>
    <div class="topbar-actions">
      <button type="button" class="back-btn secondary" @click="goBackToStudent">
        {{ backLabel }}
      </button>
      <span class="status" :class="{ alert: props.alert }">{{ props.status }}</span>
    </div>
  </header>
</template>

<style scoped>
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

.topbar h2 {
  margin: 0;
  color: var(--ink-strong);
}

.topbar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.back-btn {
  min-height: 38px;
  padding-inline: 16px;
}

.status {
  font-size: 12px;
  padding: 6px 10px;
  border-radius: 999px;
  background: var(--brand-soft);
  color: var(--brand);
}

.status.alert {
  background: #ffe9e9;
  color: var(--danger);
}

@media (max-width: 880px) {
  .topbar {
    align-items: flex-start;
  }

  .topbar-actions {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
