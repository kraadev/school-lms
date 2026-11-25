<script setup lang="ts">
withDefaults(defineProps<{
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost'
  type?: 'button' | 'submit' | 'reset'
  disabled?: boolean
}>(), {
  variant: 'primary',
  type: 'button',
  disabled: false,
})
</script>

<template>
  <button class="app-button" :class="`app-button-${variant}`" :type="type" :disabled="disabled">
    <slot />
  </button>
</template>

<style scoped>
.app-button { display: inline-flex; align-items: center; justify-content: center; gap: 8px; border: 1px solid transparent; border-radius: 9px; padding: 11px 15px; cursor: pointer; font: inherit; font-size: 11px; font-weight: 700; transition: .2s ease; }
.app-button:disabled { cursor: wait; opacity: .55; }
.app-button-primary { background: var(--coral); color: white; }
.app-button-primary:hover { background: #d96552; }
.app-button-secondary { border-color: var(--line); background: white; color: var(--green); }
.app-button-secondary:hover { border-color: #bdd5c5; background: #f6fbf7; }
.app-button-danger { background: #fff0ec; color: #b45242; }
.app-button-ghost { background: transparent; color: var(--muted); }
</style>
