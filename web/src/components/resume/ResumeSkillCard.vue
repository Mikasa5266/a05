<template>
  <article class="skill-card" :class="toneClass">
    <div class="skill-card-head">
      <div>
        <p class="skill-card-title">{{ title }}</p>
        <p v-if="subtitle" class="skill-card-subtitle">{{ subtitle }}</p>
      </div>
      <span v-if="badge" class="skill-card-badge">{{ badge }}</span>
    </div>

    <div v-if="normalizedItems.length" class="skill-card-body">
      <div v-for="item in normalizedItems" :key="`${title}-${item.name}-${item.meta}`" class="skill-row">
        <div class="min-w-0">
          <p class="skill-row-name" :title="item.name">{{ item.name }}</p>
          <p v-if="item.description" class="skill-row-desc" :title="item.description">{{ item.description }}</p>
        </div>
        <span v-if="item.meta" class="skill-row-meta" :title="item.meta">{{ item.meta }}</span>
      </div>
    </div>

    <p v-else class="skill-empty">{{ emptyText }}</p>
  </article>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  subtitle: {
    type: String,
    default: '',
  },
  badge: {
    type: String,
    default: '',
  },
  tone: {
    type: String,
    default: 'default',
  },
  items: {
    type: Array,
    default: () => [],
  },
  emptyText: {
    type: String,
    default: '暂无数据',
  },
})

const normalizedItems = computed(() =>
  props.items
    .map((item) => ({
      name: String(item?.name || '').trim(),
      meta: String(item?.meta || '').trim(),
      description: String(item?.description || '').trim(),
    }))
    .filter((item) => item.name)
)

const toneClass = computed(() => `skill-card--${String(props.tone || 'default').trim()}`)
</script>

<style scoped>
.skill-card {
  --skill-card-accent: #2563eb;
  --skill-row-bg: #f8fafc;
  --skill-title-color: #0f172a;
  --skill-desc-color: #64748b;
  --skill-meta-bg: #e2e8f0;
  --skill-meta-color: #334155;
  border-radius: 22px;
  border: 1px solid rgba(226, 232, 240, 0.95);
  background: #ffffff;
  border-top: 3px solid var(--skill-card-accent);
  padding: 24px;
  box-shadow: 0 14px 36px rgba(15, 23, 42, 0.05);
  display: grid;
  gap: 16px;
}

.skill-card--profile {
  --skill-card-accent: #0284c7;
  --skill-row-bg: #f0f9ff;
  --skill-title-color: #0c4a6e;
  --skill-desc-color: #0369a1;
  --skill-meta-bg: #e0f2fe;
  --skill-meta-color: #0c4a6e;
}

.skill-card--language {
  --skill-card-accent: #7c3aed;
  --skill-row-bg: #f5f3ff;
  --skill-title-color: #5b21b6;
  --skill-desc-color: #6d28d9;
  --skill-meta-bg: #ede9fe;
  --skill-meta-color: #5b21b6;
}

.skill-card--framework {
  --skill-card-accent: #0891b2;
  --skill-row-bg: #ecfeff;
  --skill-title-color: #155e75;
  --skill-desc-color: #0e7490;
  --skill-meta-bg: #cffafe;
  --skill-meta-color: #155e75;
}

.skill-card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.skill-card-title {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: var(--skill-title-color);
}

.skill-card-subtitle {
  margin: 6px 0 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--skill-desc-color);
}

.skill-card-badge {
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--skill-card-accent) 28%, white);
  background: color-mix(in srgb, var(--skill-card-accent) 12%, white);
  color: var(--skill-title-color);
  font-size: 11px;
  font-weight: 600;
  padding: 5px 10px;
}

.skill-card-body {
  display: grid;
  gap: 10px;
}

.skill-row {
  box-sizing: border-box;
  display: flex;
  width: 100%;
  min-width: 0;
  justify-content: space-between;
  align-items: flex-start;
  gap: 10px;
  border-radius: 12px;
  background: var(--skill-row-bg);
  padding: 10px 12px;
}

.skill-row .min-w-0 {
  flex: 1 1 auto;
  min-width: 0;
}

.skill-row-name {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--skill-title-color);
}

.skill-row-desc {
  margin: 4px 0 0;
  font-size: 12px;
  line-height: 1.5;
  color: var(--skill-desc-color);
}

.skill-row-meta {
  border-radius: 999px;
  background: var(--skill-meta-bg);
  color: var(--skill-meta-color);
  font-size: 11px;
  font-weight: 600;
  padding: 3px 8px;
  white-space: nowrap;
}

.skill-empty {
  margin: 0;
  font-size: 13px;
  color: #94a3b8;
}
</style>
