<script lang="ts" setup>
import { onMounted, ref } from 'vue'

const emit = defineEmits<{ confirm: []; cancel: [] }>()

const dialog = ref<HTMLDialogElement>()

onMounted(() => dialog.value?.showModal())
</script>

<template>
  <dialog ref="dialog" @cancel.prevent="emit('cancel')">
    <h2>{{ $t('workshop.leaveTitle') }}</h2>
    <p>{{ $t('workshop.leaveText') }}</p>
    <div class="actions">
      <button @click="emit('cancel')">{{ $t('workshop.leaveCancel') }}</button>
      <button @click="emit('confirm')">{{ $t('workshop.leaveConfirm') }}</button>
    </div>
  </dialog>
</template>

<style scoped>
dialog {
  max-width: 32rem;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
