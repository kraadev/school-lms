<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  content?: string | null
}>()

const allowedTags = new Set(['P', 'BR', 'STRONG', 'B', 'EM', 'I', 'U', 'H3', 'UL', 'OL', 'LI', 'A'])
const allowedAttributes = new Set(['href', 'target', 'rel'])

const safeContent = computed(() => {
  const source = props.content || ''
  if (!source) return ''
  if (typeof window === 'undefined') {
    return source
      .replace(/<\/?(script|style|iframe|object|embed)[^>]*>/gi, '')
      .replace(/\s+on[a-z-]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)/gi, '')
      .replace(/\s+(href|src)\s*=\s*("|')?javascript:[^\s"'>]+("|')?/gi, '')
  }

  const documentFragment = new DOMParser().parseFromString(source, 'text/html')
  const sanitizeElement = (element: Element) => {
    if (!allowedTags.has(element.tagName)) {
      const parent = element.parentNode
      if (!parent) return
      const replacement = documentFragment.createDocumentFragment()
      while (element.firstChild) replacement.appendChild(element.firstChild)
      parent.replaceChild(replacement, element)
      return
    }
    Array.from(element.attributes).forEach((attribute) => {
      if (!allowedAttributes.has(attribute.name.toLowerCase())) element.removeAttribute(attribute.name)
    })
    if (element.tagName === 'A') {
      const href = element.getAttribute('href') || ''
      if (!/^https?:\/\//i.test(href)) element.removeAttribute('href')
      else {
        element.setAttribute('target', '_blank')
        element.setAttribute('rel', 'noopener noreferrer')
      }
    }
    Array.from(element.children).forEach(sanitizeElement)
  }
  Array.from(documentFragment.body.children).forEach(sanitizeElement)
  return documentFragment.body.innerHTML
})
</script>

<template>
  <div v-if="safeContent" class="rich-text-content" v-html="safeContent" />
  <span v-else class="rich-text-empty">Belum ada deskripsi.</span>
</template>

<style scoped>
.rich-text-content {
  color: inherit;
  font-size: 13px;
  line-height: 1.75;
  overflow-wrap: anywhere;
}
.rich-text-content :deep(p) {
  margin: 0 0 10px;
}
.rich-text-content :deep(p:last-child) {
  margin-bottom: 0;
}
.rich-text-content :deep(h3) {
  margin: 16px 0 8px;
  color: var(--ink);
  font-family: 'Poppins', sans-serif;
  font-size: 16px;
  line-height: 1.35;
}
.rich-text-content :deep(ul),
.rich-text-content :deep(ol) {
  margin: 8px 0 12px;
  padding-left: 24px;
}
.rich-text-content :deep(li) {
  margin: 4px 0;
}
.rich-text-content :deep(a) {
  color: #347b61;
  text-decoration: underline;
  text-underline-offset: 2px;
}
.rich-text-empty {
  color: var(--muted);
}
</style>
