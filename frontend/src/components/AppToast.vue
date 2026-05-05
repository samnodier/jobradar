<script setup lang="ts">
import { useToast } from '@/composables/useToast'
const { toasts } = useToast()
</script>

<template>
  <Teleport to="body">
    <div class="toast-stack">
      <TransitionGroup name="toast">
        <div v-for="toast in toasts" :key="toast.id" class="toast" :class="`toast--${toast.type}`">
          {{ toast.message }}
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-stack {
  position: fixed;
  bottom: var(--spacing-6);
  right: var(--spacing-6);
  z-index: 999;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  pointer-events: none;
}

.toast {
  padding: var(--spacing-3) var(--spacing-4);
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  background: var(--color-bg-primary);
  border: 1px solid var(--color-border);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  color: var(--color-text-primary);
  max-width: 320px;
}

.toast--success {
  border-color: #86efac;
  background: #f0fdf4;
  color: #15803d;
}
.toast--error {
  border-color: #fca5a5;
  background: #fef2f2;
  color: #b91c1c;
}
.toast--info {
  border-color: var(--color-border);
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.2s ease;
}
.toast-enter-from {
  opacity: 0;
  transform: translateY(8px);
}
.toast-leave-to {
  opacity: 0;
  transform: translateY(8px);
}
</style>
