<script lang="ts" setup>
import { onMounted, ref } from 'vue'

const emit = defineEmits<{ confirm: []; cancel: [] }>()

const dialog = ref<HTMLDialogElement>()

onMounted(() => dialog.value?.showModal())
</script>

<template>
  <dialog ref="dialog" @cancel.prevent="emit('cancel')">
    <h2>{{ $t('scriptWarning.title') }}</h2>
    <p>{{ $t('scriptWarning.text') }}</p>
    <div class="actions">
      <button @click="emit('cancel')">{{ $t('scriptWarning.cancel') }}</button>
      <button @click="emit('confirm')">{{ $t('scriptWarning.confirm') }}</button>
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
