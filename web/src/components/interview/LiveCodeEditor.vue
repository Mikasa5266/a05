<script setup>
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import * as monaco from 'monaco-editor'

const SUPPORTED_LANGUAGES = ['javascript', 'typescript', 'python', 'go', 'java', 'cpp', 'c', 'json']

const props = defineProps({
  visible: {
    type: Boolean,
    default: false,
  },
  modelValue: {
    type: String,
    default: '',
  },
  question: {
    type: Object,
    default: () => ({}),
  },
  language: {
    type: String,
    default: 'javascript',
  },
  submitting: {
    type: Boolean,
    default: false,
  },
  readOnly: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:modelValue', 'language-change', 'submit', 'close'])

const editorHost = ref(null)
let editor = null
let suppressEmit = false
let monacoThemeReady = false

const questionTitle = computed(() => {
  const raw = String(props.question?.title || '').trim()
  return raw || '现场编程题'
})

const questionContent = computed(() => {
  const raw = String(props.question?.content || '').trim()
  return raw || '请在编辑器中完成核心逻辑，并在代码里保留关键注释。'
})

const normalizedLanguage = computed(() => {
  const raw = String(props.language || 'javascript').trim().toLowerCase()
  return SUPPORTED_LANGUAGES.includes(raw) ? raw : 'javascript'
})

const canSubmit = computed(() => {
  if (props.readOnly || props.submitting) return false
  return String(props.modelValue || '').trim().length > 0
})

function ensureEditor() {
  if (editor || !editorHost.value) return

  if (!monacoThemeReady) {
    monaco.editor.defineTheme('interview-night', {
      base: 'vs-dark',
      inherit: true,
      rules: [
        { token: 'comment', foreground: '6A9955' },
        { token: 'keyword', foreground: '4EC9B0' },
        { token: 'number', foreground: 'B5CEA8' },
        { token: 'string', foreground: 'CE9178' },
        { token: 'type', foreground: '4FC1FF' },
      ],
      colors: {
        'editor.background': '#0f1723',
        'editorLineNumber.foreground': '#5B6B7A',
        'editorLineNumber.activeForeground': '#A9BDD2',
        'editor.selectionBackground': '#1c3550',
      },
    })
    monacoThemeReady = true
  }

  editor = monaco.editor.create(editorHost.value, {
    value: String(props.modelValue || ''),
    language: normalizedLanguage.value,
    theme: 'interview-night',
    automaticLayout: true,
    fontSize: 14,
    fontFamily: 'JetBrains Mono, Fira Code, Cascadia Code, Consolas, monospace',
    fontLigatures: true,
    minimap: { enabled: false },
    scrollBeyondLastLine: false,
    lineNumbers: 'on',
    lineNumbersMinChars: 3,
    glyphMargin: false,
    folding: true,
    guides: {
      indentation: true,
      bracketPairs: true,
    },
    bracketPairColorization: {
      enabled: true,
    },
    wordWrap: 'on',
    readOnly: props.readOnly,
  })

  editor.onDidChangeModelContent(() => {
    if (suppressEmit) return
    emit('update:modelValue', editor.getValue())
  })
}

function syncEditorValue(nextValue) {
  if (!editor) return
  const normalized = String(nextValue || '')
  if (editor.getValue() === normalized) return

  suppressEmit = true
  editor.setValue(normalized)
  suppressEmit = false
}

function syncEditorLanguage(nextLanguage) {
  if (!editor) return
  const model = editor.getModel()
  if (!model) return
  monaco.editor.setModelLanguage(model, nextLanguage)
}

function handleLanguageChange(event) {
  const nextLanguage = String(event?.target?.value || 'javascript').trim().toLowerCase()
  emit('language-change', SUPPORTED_LANGUAGES.includes(nextLanguage) ? nextLanguage : 'javascript')
}

watch(
  () => props.visible,
  async (visible) => {
    if (!visible) return
    await nextTick()
    ensureEditor()
    syncEditorValue(props.modelValue)
    syncEditorLanguage(normalizedLanguage.value)
    editor?.updateOptions({ readOnly: props.readOnly })
  },
  { immediate: true }
)

watch(
  () => props.modelValue,
  (nextValue) => {
    syncEditorValue(nextValue)
  }
)

watch(
  () => normalizedLanguage.value,
  (nextLanguage) => {
    syncEditorLanguage(nextLanguage)
  }
)

watch(
  () => props.readOnly,
  (readOnly) => {
    editor?.updateOptions({ readOnly })
  }
)

onBeforeUnmount(() => {
  if (editor) {
    editor.dispose()
    editor = null
  }
})
</script>

<template>
  <section v-if="visible" class="code-editor-shell">
    <div class="code-editor-head">
      <div>
        <p class="code-editor-eyebrow">Live Coding</p>
        <h3 class="code-editor-title">{{ questionTitle }}</h3>
      </div>
      <div class="code-editor-tools">
        <label class="tool-label" for="code-language">语言</label>
        <select id="code-language" class="language-select" :value="normalizedLanguage" @change="handleLanguageChange">
          <option value="javascript">JavaScript</option>
          <option value="typescript">TypeScript</option>
          <option value="python">Python</option>
          <option value="go">Go</option>
          <option value="java">Java</option>
          <option value="cpp">C++</option>
          <option value="c">C</option>
          <option value="json">JSON</option>
        </select>
      </div>
    </div>

    <p class="code-editor-question">{{ questionContent }}</p>

    <div ref="editorHost" class="editor-host" />

    <div class="code-editor-footer">
      <p v-if="readOnly" class="footer-hint">当前面试尚未开始，暂不可提交评分。</p>
      <button type="button" class="btn-minor" @click="emit('close')">收起</button>
      <button type="button" class="btn-major" :disabled="!canSubmit" @click="emit('submit')">
        {{ submitting ? '提交中...' : '提交 mock 评分' }}
      </button>
    </div>
  </section>
</template>

<style scoped>
.code-editor-shell {
  margin-top: 16px;
  border-radius: 18px;
  border: 1px solid rgba(148, 201, 255, 0.28);
  background: linear-gradient(145deg, rgba(8, 21, 42, 0.92), rgba(19, 41, 70, 0.88));
  padding: 14px;
}

.code-editor-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.code-editor-eyebrow {
  margin: 0;
  font-size: 11px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: rgba(198, 227, 255, 0.72);
}

.code-editor-title {
  margin: 6px 0 0;
  font-size: 16px;
  font-weight: 600;
  color: #f1f8ff;
}

.code-editor-tools {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tool-label {
  font-size: 12px;
  color: rgba(227, 243, 255, 0.85);
}

.language-select {
  border-radius: 10px;
  border: 1px solid rgba(159, 214, 255, 0.35);
  background: rgba(11, 29, 53, 0.85);
  color: #f4fbff;
  padding: 6px 10px;
  font-size: 12px;
}

.code-editor-question {
  margin: 12px 0;
  font-size: 13px;
  line-height: 1.5;
  color: rgba(233, 246, 255, 0.9);
  white-space: pre-wrap;
}

.editor-host {
  height: 280px;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid rgba(159, 214, 255, 0.24);
}

.code-editor-footer {
  margin-top: 12px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

.footer-hint {
  margin: 0;
  margin-right: auto;
  font-size: 12px;
  color: rgba(255, 209, 158, 0.94);
}

.btn-minor,
.btn-major {
  border-radius: 10px;
  padding: 8px 12px;
  font-size: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border: 1px solid transparent;
}

.btn-minor {
  background: rgba(31, 71, 118, 0.4);
  border-color: rgba(162, 210, 255, 0.35);
  color: rgba(241, 249, 255, 0.95);
}

.btn-major {
  background: linear-gradient(135deg, rgba(95, 178, 255, 0.75), rgba(62, 144, 236, 0.8));
  border-color: rgba(197, 232, 255, 0.5);
  color: #f8fcff;
}

.btn-major:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 767px) {
  .code-editor-head {
    flex-direction: column;
  }

  .editor-host {
    height: 240px;
  }

  .code-editor-footer {
    flex-wrap: wrap;
  }

  .btn-minor,
  .btn-major {
    flex: 1;
  }
}
</style>
