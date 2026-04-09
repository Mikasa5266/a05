<template>
  <component
    :is="tag"
    class="question-markdown"
    :class="{
      'question-markdown-inline': inline,
      'question-markdown-empty': !normalizedSource,
    }"
    v-html="sanitizedHtml"
  />
</template>

<script setup>
import { computed } from "vue";
import DOMPurify from "dompurify";
import { Marked, Renderer } from "marked";

const props = defineProps({
  source: {
    type: [String, Number],
    default: "",
  },
  tag: {
    type: String,
    default: "div",
  },
  inline: {
    type: Boolean,
    default: false,
  },
  emptyText: {
    type: String,
    default: "",
  },
});

const escapeHtml = (value = "") =>
  String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");

const escapeAttribute = (value = "") =>
  escapeHtml(value).replaceAll("`", "&#96;");

const unwrapOuterFence = (text = "") => {
  const trimmed = String(text || "").trim();
  const match = trimmed.match(
    /^```(?:json|markdown|md|text|txt)?\s*([\s\S]*?)\s*```$/i,
  );
  if (!match) {
    return trimmed;
  }
  return String(match[1] || "").trim();
};

const normalizeMarkdownSource = (value) => {
  let text = String(value ?? "");
  text = text.replaceAll("\uFEFF", "");
  text = text.replaceAll("\uFFFD", "");
  text = text.replace(/\r\n?/g, "\n");
  if (!text.includes("\n") && text.includes("\\n")) {
    text = text.replace(/\\n/g, "\n");
  }
  text = unwrapOuterFence(text);
  text = text.replace(/[^\S\n]+\n/g, "\n");
  text = text.replace(/[\u0000-\u0008\u000B\u000C\u000E-\u001F\u007F]/g, "");
  return text.trim();
};

const renderer = new Renderer();

renderer.code = ({ text, lang }) => {
  const normalizedLanguage = String(lang || "").trim().toLowerCase();
  const languageClass = normalizedLanguage
    ? `language-${escapeAttribute(normalizedLanguage)}`
    : "language-plain";
  const languageBadge = normalizedLanguage
    ? ` data-lang="${escapeAttribute(normalizedLanguage)}"`
    : "";
  return `<pre class="question-code-block"${languageBadge}><code class="${languageClass}">${escapeHtml(text)}</code></pre>`;
};

const markdown = new Marked({
  async: false,
  breaks: true,
  gfm: true,
  renderer,
});

const normalizedSource = computed(() => normalizeMarkdownSource(props.source));

const renderedHtml = computed(() => {
  if (!normalizedSource.value) {
    if (!props.emptyText) {
      return "";
    }
    return `<p>${escapeHtml(props.emptyText)}</p>`;
  }

  try {
    if (props.inline) {
      return markdown.parseInline(normalizedSource.value);
    }
    return markdown.parse(normalizedSource.value);
  } catch {
    return `<p>${escapeHtml(normalizedSource.value)}</p>`;
  }
});

const sanitizedHtml = computed(() =>
  DOMPurify.sanitize(renderedHtml.value, {
    USE_PROFILES: { html: true },
    ADD_ATTR: ["class", "target", "rel", "data-lang"],
  }),
);
</script>

<style scoped>
.question-markdown {
  color: inherit;
  line-height: 1.8;
  word-break: break-word;
}

.question-markdown-inline :deep(p) {
  display: inline;
  margin: 0;
}

.question-markdown :deep(p:first-child) {
  margin-top: 0;
}

.question-markdown :deep(p:last-child) {
  margin-bottom: 0;
}

.question-markdown :deep(p) {
  margin: 0 0 0.9rem;
}

.question-markdown :deep(ul),
.question-markdown :deep(ol) {
  margin: 0 0 0.9rem;
  padding-left: 1.1rem;
}

.question-markdown :deep(li + li) {
  margin-top: 0.35rem;
}

.question-markdown :deep(strong) {
  color: #0f172a;
}

.question-markdown :deep(a) {
  color: #0369a1;
  text-decoration: underline;
}

.question-markdown :deep(code:not(.language-plain)) {
  font-family: "JetBrains Mono", "Fira Code", Consolas, monospace;
}

.question-markdown :deep(code) {
  border-radius: 0.5rem;
  background: rgba(148, 163, 184, 0.14);
  padding: 0.14rem 0.38rem;
  font-size: 0.92em;
}

.question-markdown :deep(.question-code-block) {
  position: relative;
  overflow-x: auto;
  margin: 0 0 1rem;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 1.25rem;
  background:
    radial-gradient(circle at top left, rgba(14, 165, 233, 0.14), transparent 38%),
    linear-gradient(180deg, #0f172a 0%, #111827 100%);
  padding: 1.2rem 1rem 1rem;
  color: #e2e8f0;
}

.question-markdown :deep(.question-code-block[data-lang]::before) {
  content: attr(data-lang);
  position: absolute;
  top: 0.85rem;
  right: 0.95rem;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.18);
  padding: 0.12rem 0.5rem;
  font-size: 0.68rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: rgba(226, 232, 240, 0.92);
}

.question-markdown :deep(.question-code-block code) {
  display: block;
  min-width: max-content;
  background: transparent;
  padding: 0;
  color: inherit;
  font-size: 0.86rem;
  line-height: 1.7;
}

.question-markdown-empty {
  color: #64748b;
}
</style>
