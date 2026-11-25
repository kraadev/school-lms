<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
}>(), {
  placeholder: 'Tulis deskripsi dengan paragraf, daftar, atau penekanan teks...'
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const editor = ref<HTMLElement | null>(null)
const isEmpty = ref(!props.modelValue)

const syncEmptyState = () => {
  isEmpty.value = !(editor.value?.innerText.trim())
}

const emitContent = () => {
  if (!editor.value) return
  emit('update:modelValue', editor.value.innerHTML === '<br>' ? '' : editor.value.innerHTML)
  syncEmptyState()
}

const runCommand = (command: string, value?: string) => {
  editor.value?.focus()
  document.execCommand(command, false, value)
  emitContent()
}

const addLink = () => {
  const url = window.prompt('Masukkan alamat tautan')
  if (!url?.trim()) return
  const normalizedUrl = /^https?:\/\//i.test(url.trim()) ? url.trim() : `https://${url.trim()}`
  runCommand('createLink', normalizedUrl)
}

watch(() => props.modelValue, async (value) => {
  if (!editor.value || editor.value.innerHTML === value) return
  editor.value.innerHTML = value || ''
  await nextTick()
  syncEmptyState()
})

onMounted(() => {
  if (editor.value) editor.value.innerHTML = props.modelValue || ''
  syncEmptyState()
})
</script>

<template>
  <div class="rich-text-editor">
    <div class="rich-text-toolbar" role="toolbar" aria-label="Format teks">
      <button type="button" title="Paragraf" @click="runCommand('formatBlock', 'P')">¶</button>
      <button type="button" title="Judul kecil" @click="runCommand('formatBlock', 'H3')">H3</button>
      <button type="button" title="Tebal" @click="runCommand('bold')"><strong>B</strong></button>
      <button type="button" title="Miring" @click="runCommand('italic')"><em>I</em></button>
      <button type="button" title="Garis bawah" @click="runCommand('underline')"><u>U</u></button>
      <span class="toolbar-divider" />
      <button type="button" title="Daftar berpoin" @click="runCommand('insertUnorderedList')">• List</button>
      <button type="button" title="Daftar bernomor" @click="runCommand('insertOrderedList')">1. List</button>
      <button type="button" title="Tambah tautan" @click="addLink">Link</button>
    </div>
    <div
      ref="editor"
      class="rich-text-input"
      contenteditable="true"
      role="textbox"
      aria-multiline="true"
      :aria-label="placeholder"
      @input="emitContent"
      @blur="emitContent"
    />
    <span v-if="isEmpty" class="rich-text-placeholder">{{ placeholder }}</span>
  </div>
</template>

<style scoped>
.rich-text-editor {
  position: relative;
  overflow: hidden;
  border: 1px solid #dfe7e1;
  border-radius: 10px;
  background: #fff;
}
.rich-text-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  padding: 7px;
  border-bottom: 1px solid #edf2ee;
  background: #f8fbf9;
}
.rich-text-toolbar button {
  min-width: 30px;
  min-height: 28px;
  padding: 4px 7px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: #587265;
  cursor: pointer;
  font-size: 11px;
}
.rich-text-toolbar button:hover {
  background: #e5f1e8;
  color: #2d6247;
}
.toolbar-divider {
  width: 1px;
  height: 20px;
  margin: 0 3px;
  background: #dfe7e1;
}
.rich-text-input {
  min-height: 132px;
  padding: 13px 14px;
  color: var(--ink);
  font-size: 13px;
  line-height: 1.7;
  outline: none;
}
.rich-text-input:focus {
  box-shadow: inset 0 0 0 2px rgba(94, 182, 157, 0.18);
}
.rich-text-input :deep(p),
.rich-text-input :deep(h3),
.rich-text-input :deep(ul),
.rich-text-input :deep(ol) {
  margin: 0 0 8px;
}
.rich-text-input :deep(ul),
.rich-text-input :deep(ol) {
  padding-left: 22px;
}
.rich-text-input :deep(a) {
  color: #347b61;
  text-decoration: underline;
}
.rich-text-placeholder {
  position: absolute;
  top: 53px;
  left: 14px;
  color: #9aa8a0;
  font-size: 12px;
  pointer-events: none;
}
</style>
