<script setup>
import { usePracticeStore } from "../../stores/practice";

const practiceStore = usePracticeStore();

const navItems = [
  { key: "bank", label: "题库与题单" },
  { key: "practice", label: "刷题模式" },
  { key: "wrong", label: "错题本" },
  { key: "dashboard", label: "数据统计" },
  { key: "resume", label: "简历解析" },
];

const handleRoleChange = async (event) => {
  await practiceStore.changeRole(event.target.value);
};

const handleImport = async (event) => {
  const file = event.target.files?.[0];
  if (!file) return;
  await practiceStore.importQuestionBankFromFile(file);
  event.target.value = "";
};

const handleNavClick = async (view) => {
  await practiceStore.switchView(view);
};
</script>

<template>
  <aside class="sidebar">
    <div class="brand">
      <h1>Campus Tech Drill</h1>
      <p>四大岗位 · 三层难度 · 本地离线</p>
    </div>

    <section class="panel compact">
      <h3>岗位切换</h3>
      <select :value="practiceStore.currentRole" @change="handleRoleChange">
        <option
          v-for="role in practiceStore.roleOptions"
          :key="role.value"
          :value="role.value"
        >
          {{ role.label }}
        </option>
      </select>
    </section>

    <nav class="nav-group">
      <button
        v-for="item in navItems"
        :key="item.key"
        class="nav-btn"
        :class="{ active: practiceStore.activeView === item.key }"
        type="button"
        @click="handleNavClick(item.key)"
      >
        {{ item.label }}
      </button>
    </nav>

    <section class="panel compact">
      <h3>基础操作</h3>
      <label class="file-upload">
        本地导入题库
        <input type="file" accept="application/json" @change="handleImport" />
      </label>
      <button
        type="button"
        class="secondary"
        :disabled="practiceStore.loading.export"
        @click="practiceStore.exportQuestionRecords"
      >
        {{ practiceStore.loading.export ? "导出中..." : "导出答题记录" }}
      </button>
    </section>
  </aside>
</template>

<style scoped>
.sidebar {
  background: linear-gradient(160deg, #0f344b 0%, #15557a 56%, #1b6f95 100%);
  color: #ecf8ff;
  border-radius: 20px;
  padding: 18px;
  width: 100%;
  min-height: calc(100dvh - (var(--page-padding) * 2));
  display: flex;
  flex-direction: column;
  gap: 14px;
  box-shadow: var(--shadow);
  border: 1px solid rgba(199, 231, 247, 0.32);
}

.brand h1 {
  margin: 0;
  font-size: 22px;
  letter-spacing: 0.3px;
}

.brand p {
  margin: 8px 0 0;
  font-size: 13px;
  opacity: 0.88;
}

.panel {
  border: 1px solid rgba(209, 234, 248, 0.38);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.1);
  padding: 12px;
}

.panel.compact h3 {
  margin: 0 0 8px;
  font-size: 14px;
  font-weight: 700;
}

.nav-group {
  display: grid;
  gap: 9px;
}

.nav-btn {
  text-align: left;
  background: rgba(255, 255, 255, 0.16);
}

.nav-btn.active {
  background: linear-gradient(135deg, #2a7eb1 0%, #3f95ca 100%);
  color: #ffffff;
}

.file-upload {
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(236, 244, 255, 0.4);
  border-radius: 10px;
  padding: 11px;
  cursor: pointer;
  margin-bottom: 8px;
  background: linear-gradient(135deg, #2a7eb1 0%, #3f95ca 100%);
  color: #ffffff;
}

.file-upload input {
  display: none;
}

@media (max-width: 880px) {
  .sidebar {
    min-height: auto;
  }
}
</style>
