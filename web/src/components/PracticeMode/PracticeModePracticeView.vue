<script setup>
import PracticeModeCoachCard from "./PracticeModeCoachCard.vue";
import PracticeModeCodeChallengePanel from "./PracticeModeCodeChallengePanel.vue";
import PracticeModeControlPanel from "./PracticeModeControlPanel.vue";
import PracticeModeQuestionCard from "./PracticeModeQuestionCard.vue";

defineProps({
  active: {
    type: Boolean,
    default: false,
  },
});
</script>

<template>
  <section class="view" :class="{ active }">
    <PracticeModeControlPanel />

    <div class="practice-workspace">
      <PracticeModeQuestionCard />
      <div class="practice-side-stack">
        <PracticeModeCodeChallengePanel />
        <PracticeModeCoachCard />
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

.practice-workspace {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 12px;
  align-items: start;
}

.practice-side-stack {
  display: grid;
  gap: 12px;
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
  .practice-workspace {
    grid-template-columns: 1fr;
  }
}
</style>
