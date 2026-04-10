<script setup>
import { onMounted, onUnmounted } from "vue";

import PracticeModeBankView from "../../../components/PracticeMode/PracticeModeBankView.vue";
import PracticeModeDashboardView from "../../../components/PracticeMode/PracticeModeDashboardView.vue";
import PracticeModePracticeView from "../../../components/PracticeMode/PracticeModePracticeView.vue";
import PracticeModeResumeView from "../../../components/PracticeMode/PracticeModeResumeView.vue";
import PracticeModeSidebar from "../../../components/PracticeMode/PracticeModeSidebar.vue";
import PracticeModeTopbar from "../../../components/PracticeMode/PracticeModeTopbar.vue";
import PracticeModeWrongBookView from "../../../components/PracticeMode/PracticeModeWrongBookView.vue";
import { usePracticeStore } from "../../../stores/practice";

const practiceStore = usePracticeStore();

const handleShortcut = async (event) => {
  const tag = document.activeElement?.tagName?.toLowerCase();
  const isFormTag = tag === "input" || tag === "textarea" || tag === "select";
  if (isFormTag || document.activeElement?.isContentEditable) {
    return;
  }

  if (event.code === "Space" && practiceStore.currentQuestionHasOptions) {
    event.preventDefault();
    practiceStore.selectNextOption();
    return;
  }

  if (event.key.toLowerCase() === "n") {
    event.preventDefault();
    await practiceStore.loadQuestion();
    return;
  }

  if (event.key.toLowerCase() === "p") {
    event.preventDefault();
    practiceStore.moveHistory(-1);
    return;
  }

  if (event.key.toLowerCase() === "a") {
    event.preventDefault();
    await practiceStore.showCurrentSolution();
  }
};

onMounted(async () => {
  await practiceStore.initialize();
  document.addEventListener("keydown", handleShortcut);
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleShortcut);
  practiceStore.teardown();
});
</script>

<template>
  <div class="practice-mode-app">
    <div class="bg-decor bg-decor-one" aria-hidden="true"></div>
    <div class="bg-decor bg-decor-two" aria-hidden="true"></div>

    <main class="app-layout">
      <div class="sidebar-column">
        <PracticeModeSidebar />
      </div>

      <section class="content">
        <PracticeModeTopbar
          :title="practiceStore.viewTitle"
          :status="practiceStore.status.text"
          :alert="practiceStore.status.isAlert"
        />

        <PracticeModeBankView :active="practiceStore.activeView === 'bank'" />
        <PracticeModePracticeView :active="practiceStore.activeView === 'practice'" />
        <PracticeModeWrongBookView :active="practiceStore.activeView === 'wrong'" />
        <PracticeModeDashboardView :active="practiceStore.activeView === 'dashboard'" />
        <PracticeModeResumeView :active="practiceStore.activeView === 'resume'" />
      </section>
    </main>
  </div>
</template>

<style scoped>
:global(.practice-mode-app) {
  --bg-main: #f3f8fb;
  --panel: #ffffff;
  --panel-soft: #f7fbff;
  --ink-strong: #12263a;
  --ink: #3c546c;
  --brand: #1b6f95;
  --brand-soft: #dff5ff;
  --accent: #5c9fff;
  --danger: #cf4c4c;
  --ok: #1f8d67;
  --line: #d8e8f2;
  --shadow: 0 22px 50px rgba(17, 65, 94, 0.14);
  --radius: 16px;
  --page-padding: clamp(12px, 1.4vw, 20px);
  --layout-gap: clamp(14px, 1.2vw, 18px);
  --sidebar-width: clamp(280px, 20vw, 300px);
  position: relative;
  min-height: 100vh;
  font-family: "Noto Sans SC", "Source Han Sans SC", "Microsoft YaHei", sans-serif;
  background:
    radial-gradient(90rem 50rem at -10% -10%, #dff5ff 0%, rgba(223, 245, 255, 0) 52%),
    radial-gradient(64rem 38rem at 110% -12%, #dbeeff 0%, rgba(219, 238, 255, 0) 55%),
    linear-gradient(180deg, #f9fcff 0%, #edf5fb 100%);
  color: var(--ink);
  overflow-x: hidden;
}

:global(.practice-mode-app *) {
  box-sizing: border-box;
}

:global(.practice-mode-app select),
:global(.practice-mode-app input),
:global(.practice-mode-app textarea),
:global(.practice-mode-app button) {
  font: inherit;
}

:global(.practice-mode-app select),
:global(.practice-mode-app input),
:global(.practice-mode-app textarea) {
  width: 100%;
  border: 1px solid var(--line);
  border-radius: 10px;
  padding: 10px 12px;
  background: #fff;
  color: var(--ink-strong);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

:global(.practice-mode-app select:focus),
:global(.practice-mode-app input:focus),
:global(.practice-mode-app textarea:focus) {
  outline: none;
  border-color: #82c6e2;
  box-shadow: 0 0 0 3px rgba(130, 198, 226, 0.2);
}

:global(.practice-mode-app textarea) {
  min-height: 130px;
  resize: vertical;
}

:global(.practice-mode-app button) {
  border: none;
  border-radius: 12px;
  background: var(--brand);
  color: #fff;
  padding: 10px 14px;
  cursor: pointer;
  font-weight: 700;
  box-shadow: 0 8px 18px rgba(15, 110, 145, 0.24);
  transition: transform 0.16s ease, opacity 0.16s ease, box-shadow 0.16s ease;
}

:global(.practice-mode-app button:hover) {
  transform: translateY(-1px);
  opacity: 0.98;
  box-shadow: 0 10px 20px rgba(15, 110, 145, 0.3);
}

:global(.practice-mode-app button.secondary) {
  background: #2a7eb1;
  box-shadow: 0 8px 16px rgba(42, 126, 177, 0.24);
}

:global(.practice-mode-app button.danger) {
  background: var(--danger);
}

:global(.practice-mode-app .bg-decor) {
  position: fixed;
  border-radius: 50%;
  filter: blur(14px);
  z-index: 0;
}

:global(.practice-mode-app .bg-decor-one) {
  width: 280px;
  height: 280px;
  background: rgba(15, 110, 145, 0.22);
  top: -86px;
  right: -70px;
}

:global(.practice-mode-app .bg-decor-two) {
  width: 250px;
  height: 250px;
  background: rgba(83, 145, 214, 0.2);
  bottom: -90px;
  left: -80px;
}

:global(.practice-mode-app .app-layout) {
  position: relative;
  z-index: 2;
  display: grid;
  width: min(100%, 1920px);
  margin: 0 auto;
  grid-template-columns: minmax(280px, var(--sidebar-width)) minmax(0, 1fr);
  gap: var(--layout-gap);
  padding: var(--page-padding);
  min-height: 100dvh;
  align-items: stretch;
}

:global(.practice-mode-app .sidebar-column) {
  min-width: 0;
  display: flex;
  align-self: stretch;
}

:global(.practice-mode-app .content) {
  background: rgba(255, 255, 255, 0.85);
  border: 1px solid rgba(207, 227, 239, 0.9);
  border-radius: 20px;
  backdrop-filter: blur(4px);
  box-shadow: var(--shadow);
  padding: 18px;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: calc(100dvh - (var(--page-padding) * 2));
}

@media (max-width: 880px) {
  :global(.practice-mode-app .app-layout) {
    grid-template-columns: 1fr;
    min-height: auto;
  }

  :global(.practice-mode-app .content) {
    min-height: auto;
  }
}
</style>
