<script setup lang="ts">
import { useToast } from '@/composables/useToast'
const { toasts } = useToast()
</script>

<template>
  <Teleport to="body">
    <div class="fixed bottom-6 right-6 z-999 flex flex-col gap-2 pointer-events-none">
      <TransitionGroup 
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="opacity-0 translate-y-2"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="opacity-100 scale-100"
        leave-to-class="opacity-0 scale-95"
      >
        <div 
          v-for="toast in toasts" 
          :key="toast.id" 
          class="px-4 py-3 text-sm font-semibold border shadow-lg max-w-[320px] pointer-events-auto rounded-lg"
          :class="{
            'border-green-200 bg-green-50 text-green-800': toast.type === 'success',
            'border-red-200 bg-red-50 text-red-800': toast.type === 'error',
            'border-ui-border bg-white text-gray-900': toast.type === 'info' || !toast.type,
          }"
        >
          {{ toast.message }}
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
