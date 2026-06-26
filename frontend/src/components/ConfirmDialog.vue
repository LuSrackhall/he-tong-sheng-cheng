<script setup lang="ts">
import { useEscapeKey } from '../composables/useEscapeKey'

const props = withDefaults(defineProps<{
  visible: boolean
  title: string
  message: string
  confirmText?: string
  variant?: 'danger' | 'default'
}>(), {
  confirmText: '确认',
  variant: 'default',
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
  confirm: []
  cancel: []
}>()

useEscapeKey(() => {
  if (props.visible) {
    emit('update:visible', false)
    emit('cancel')
  }
})

function onConfirm() {
  emit('update:visible', false)
  emit('confirm')
}

function onCancel() {
  emit('update:visible', false)
  emit('cancel')
}

function onOverlayClick() {
  onCancel()
}
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="modal-overlay" @click.self="onOverlayClick">
      <div
        class="modal-content"
        style="max-width: 400px;"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="'confirm-dialog-title-' + title"
      >
        <h3 :id="'confirm-dialog-title-' + title" style="margin-bottom: 12px;">{{ title }}</h3>
        <p style="margin-bottom: 20px; color: var(--color-text-secondary); font-size: 0.9375rem; line-height: 1.6;">{{ message }}</p>
        <div style="display: flex; gap: 8px; justify-content: flex-end;">
          <button class="btn btn-secondary" @click="onCancel">取消</button>
          <button
            :class="['btn', variant === 'danger' ? 'btn-danger' : 'btn-primary']"
            @click="onConfirm"
          >{{ confirmText }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
