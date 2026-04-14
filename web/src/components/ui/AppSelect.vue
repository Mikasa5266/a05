<template>
  <div
    ref="rootRef"
    class="app-select"
    :class="[
      fullWidth ? 'w-full' : '',
      disabled ? 'app-select-disabled' : '',
      open ? 'app-select-open' : ''
    ]"
    @keydown="handleRootKeydown"
  >
    <button
      ref="triggerRef"
      type="button"
      class="app-select-trigger"
      :class="sizeClass"
      :disabled="disabled"
      :aria-expanded="open"
      aria-haspopup="listbox"
      @click="toggleOpen"
    >
      <span class="app-select-value" :class="hasValue ? 'text-zinc-800' : 'text-zinc-400'">
        {{ selectedLabel || placeholder }}
      </span>
      <span class="app-select-actions">
        <span
          v-if="clearable && hasValue && !disabled"
          class="app-select-clear"
          role="button"
          tabindex="0"
          @click.stop="clearValue"
          @keydown.enter.prevent.stop="clearValue"
          @keydown.space.prevent.stop="clearValue"
          aria-label="清空"
        >
          <X class="h-3.5 w-3.5" />
        </span>
        <ChevronDown class="h-4 w-4 transition-transform" :class="open ? 'rotate-180' : ''" />
      </span>
    </button>

    <transition
      enter-active-class="transition-all duration-150 ease-out"
      leave-active-class="transition-all duration-120 ease-in"
      enter-from-class="opacity-0 translate-y-1"
      enter-to-class="opacity-100 translate-y-0"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 translate-y-1"
    >
      <div
        v-if="open"
        class="app-select-menu"
        role="listbox"
        :style="{ maxHeight: computedMenuMaxHeight }"
      >
        <div v-if="searchable" class="app-select-search-wrap">
          <input
            ref="searchRef"
            v-model="searchQuery"
            type="text"
            class="app-select-search"
            placeholder="搜索选项..."
            @keydown.down.prevent="focusNext"
            @keydown.up.prevent="focusPrev"
            @keydown.enter.prevent="selectHighlighted"
          />
        </div>

        <div v-if="loading" class="app-select-empty">加载中...</div>

        <template v-else>
          <button
            v-for="(option, index) in filteredOptions"
            :key="option.key"
            type="button"
            class="app-select-option"
            :class="[
              option.disabled ? 'app-select-option-disabled' : '',
              isSelected(option) ? 'app-select-option-selected' : '',
              highlightedIndex === index ? 'app-select-option-highlighted' : ''
            ]"
            :disabled="option.disabled"
            @mouseenter="highlightedIndex = index"
            @click="selectOption(option)"
          >
            <span class="truncate">{{ option.label }}</span>
            <Check v-if="isSelected(option)" class="h-4 w-4 text-indigo-600" />
          </button>

          <div v-if="!filteredOptions.length" class="app-select-empty">
            {{ emptyText }}
          </div>
        </template>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { Check, ChevronDown, X } from 'lucide-vue-next'

const props = defineProps({
  modelValue: {
    type: [String, Number, Boolean, null],
    default: ''
  },
  options: {
    type: Array,
    default: () => []
  },
  placeholder: {
    type: String,
    default: '请选择'
  },
  disabled: {
    type: Boolean,
    default: false
  },
  clearable: {
    type: Boolean,
    default: false
  },
  searchable: {
    type: Boolean,
    default: false
  },
  loading: {
    type: Boolean,
    default: false
  },
  fullWidth: {
    type: Boolean,
    default: false
  },
  size: {
    type: String,
    default: 'md'
  },
  menuMaxHeight: {
    type: [Number, String],
    default: 260
  },
  emptyText: {
    type: String,
    default: '暂无选项'
  },
  optionLabelKey: {
    type: String,
    default: 'label'
  },
  optionValueKey: {
    type: String,
    default: 'value'
  }
})

const emit = defineEmits(['update:modelValue', 'change', 'open', 'close'])

const rootRef = ref(null)
const triggerRef = ref(null)
const searchRef = ref(null)
const open = ref(false)
const searchQuery = ref('')
const highlightedIndex = ref(-1)

const normalizedOptions = computed(() => {
  return props.options.map((item, index) => {
    if (item && typeof item === 'object' && !Array.isArray(item)) {
      const label = String(item[props.optionLabelKey] ?? item.label ?? '')
      const value = item[props.optionValueKey] ?? item.value ?? ''
      return {
        key: `opt-${index}-${String(value)}`,
        label: label || String(value),
        value,
        disabled: Boolean(item.disabled)
      }
    }

    return {
      key: `opt-${index}-${String(item)}`,
      label: String(item ?? ''),
      value: item,
      disabled: false
    }
  })
})

const filteredOptions = computed(() => {
  const query = String(searchQuery.value || '').trim().toLowerCase()
  if (!query) return normalizedOptions.value
  return normalizedOptions.value.filter((option) => option.label.toLowerCase().includes(query))
})

const hasValue = computed(() => props.modelValue !== '' && props.modelValue !== null && props.modelValue !== undefined)

const selectedOption = computed(() => {
  return normalizedOptions.value.find((option) => option.value === props.modelValue) || null
})

const selectedLabel = computed(() => selectedOption.value?.label || '')

const sizeClass = computed(() => {
  if (props.size === 'lg') return 'app-select-size-lg'
  if (props.size === 'sm') return 'app-select-size-sm'
  return 'app-select-size-md'
})

const computedMenuMaxHeight = computed(() => {
  if (typeof props.menuMaxHeight === 'number') {
    return `${props.menuMaxHeight}px`
  }
  return String(props.menuMaxHeight || '260px')
})

const isSelected = (option) => option.value === props.modelValue

const openMenu = async () => {
  if (props.disabled || open.value) return
  open.value = true
  emit('open')
  const selectedIndex = filteredOptions.value.findIndex((option) => isSelected(option) && !option.disabled)
  highlightedIndex.value = selectedIndex >= 0 ? selectedIndex : filteredOptions.value.findIndex((option) => !option.disabled)
  await nextTick()
  if (props.searchable) {
    searchRef.value?.focus?.()
  }
}

const closeMenu = () => {
  if (!open.value) return
  open.value = false
  searchQuery.value = ''
  highlightedIndex.value = -1
  emit('close')
}

const toggleOpen = () => {
  if (open.value) {
    closeMenu()
    return
  }
  void openMenu()
}

const clearValue = () => {
  emit('update:modelValue', '')
  emit('change', '')
}

const selectOption = (option) => {
  if (!option || option.disabled) return
  emit('update:modelValue', option.value)
  emit('change', option.value)
  closeMenu()
}

const focusNext = () => {
  if (!filteredOptions.value.length) return
  let cursor = highlightedIndex.value
  for (let i = 0; i < filteredOptions.value.length; i += 1) {
    cursor = (cursor + 1) % filteredOptions.value.length
    if (!filteredOptions.value[cursor].disabled) {
      highlightedIndex.value = cursor
      return
    }
  }
}

const focusPrev = () => {
  if (!filteredOptions.value.length) return
  let cursor = highlightedIndex.value < 0 ? 0 : highlightedIndex.value
  for (let i = 0; i < filteredOptions.value.length; i += 1) {
    cursor = (cursor - 1 + filteredOptions.value.length) % filteredOptions.value.length
    if (!filteredOptions.value[cursor].disabled) {
      highlightedIndex.value = cursor
      return
    }
  }
}

const selectHighlighted = () => {
  if (highlightedIndex.value < 0) return
  const option = filteredOptions.value[highlightedIndex.value]
  if (!option || option.disabled) return
  selectOption(option)
}

const handleRootKeydown = (event) => {
  if (props.disabled) return

  if (!open.value) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      void openMenu()
    }
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      void openMenu()
    }
    return
  }

  if (event.key === 'Escape') {
    event.preventDefault()
    closeMenu()
    triggerRef.value?.focus?.()
    return
  }

  if (event.key === 'ArrowDown') {
    event.preventDefault()
    focusNext()
    return
  }

  if (event.key === 'ArrowUp') {
    event.preventDefault()
    focusPrev()
    return
  }

  if (event.key === 'Enter') {
    event.preventDefault()
    selectHighlighted()
  }
}

const handleClickOutside = (event) => {
  if (!rootRef.value || rootRef.value.contains(event.target)) return
  closeMenu()
}

watch(
  () => open.value,
  (isOpen) => {
    if (isOpen) {
      document.addEventListener('click', handleClickOutside)
      return
    }
    document.removeEventListener('click', handleClickOutside)
  }
)

watch(
  () => props.options,
  () => {
    if (!open.value) return
    const selectedIndex = filteredOptions.value.findIndex((option) => isSelected(option) && !option.disabled)
    highlightedIndex.value = selectedIndex >= 0 ? selectedIndex : filteredOptions.value.findIndex((option) => !option.disabled)
  }
)

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.app-select {
  position: relative;
  min-width: 170px;
}

.app-select-disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.app-select-trigger {
  width: 100%;
  border-radius: 14px;
  border: 1px solid #d4dbe7;
  background: linear-gradient(180deg, #ffffff 0%, #f9fbff 100%);
  box-shadow: 0 6px 16px rgba(15, 23, 42, 0.06);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 0 14px;
  font-size: 14px;
  transition: border-color 0.18s ease, box-shadow 0.18s ease, transform 0.12s ease;
}

.app-select-trigger:hover:not(:disabled) {
  border-color: #b8c3d6;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.08);
}

.app-select-trigger:focus-visible {
  outline: none;
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.16);
}

.app-select-size-sm {
  height: 38px;
}

.app-select-size-md {
  height: 46px;
}

.app-select-size-lg {
  height: 52px;
}

.app-select-value {
  min-width: 0;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-align: left;
}

.app-select-actions {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: #64748b;
}

.app-select-clear {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 999px;
  color: #94a3b8;
}

.app-select-clear:hover {
  background: #eef2ff;
  color: #4f46e5;
}

.app-select-menu {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  right: 0;
  z-index: 120;
  border-radius: 14px;
  border: 1px solid #dbe1ea;
  background: #ffffff;
  box-shadow: 0 18px 36px rgba(15, 23, 42, 0.12);
  overflow-y: auto;
  padding: 6px;
}

.app-select-search-wrap {
  padding: 4px 4px 8px;
}

.app-select-search {
  width: 100%;
  height: 34px;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  padding: 0 10px;
  font-size: 13px;
}

.app-select-search:focus-visible {
  outline: none;
  border-color: #4f46e5;
  box-shadow: 0 0 0 2px rgba(79, 70, 229, 0.14);
}

.app-select-option {
  width: 100%;
  min-height: 36px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  border-radius: 10px;
  padding: 8px 10px;
  font-size: 13px;
  color: #334155;
  text-align: left;
}

.app-select-option:hover {
  background: #f1f5f9;
}

.app-select-option-highlighted {
  background: #eef2ff;
}

.app-select-option-selected {
  background: #eef2ff;
  color: #312e81;
  font-weight: 600;
}

.app-select-option-disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.app-select-empty {
  padding: 12px;
  text-align: center;
  font-size: 12px;
  color: #94a3b8;
}
</style>
