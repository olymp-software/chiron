<script lang="ts" setup>
import { onMounted, onUnmounted, ref } from 'vue'
import ErrorView from '../errors/ErrorView.vue'
import OutputPanel from '../scripts/OutputPanel.vue'
import { useCheckRun } from './checkRun'

const props = defineProps<{ courseKey: string }>()
const emit = defineEmits<{ closed: [] }>()

const dialog = ref<HTMLDialogElement>()
const panel = ref<InstanceType<typeof OutputPanel>>()

const { exit, error, cancelling, running, start, cancel, dispose } = useCheckRun(
  props.courseKey,
  (text) => panel.value?.append(text),
)

function close() {
  dialog.value?.close()
  emit('closed')
}

// cancelAndClose stops the check and closes the modal; if stopping fails,
// the modal stays open with the error and OK.
async function cancelAndClose() {
  if (await cancel()) close()
}

onMounted(() => {
  dialog.value?.showModal()
  void start()
})

onUnmounted(dispose)
</script>

<template>
  <dialog ref="dialog" @cancel.prevent>
    <h2>{{ $t('check.title') }}</h2>
    <OutputPanel ref="panel" />
    <ErrorView v-if="error" :error="error" />
    <div class="actions">
      <p v-if="exit?.stopped" class="result">{{ $t('check.stopped') }}</p>
      <p v-else-if="exit" class="result" :class="exit.exitCode === 0 ? 'success' : 'failure'">
        {{ exit.exitCode === 0 ? '✔' : '✘' }} {{ $t('check.exitCode', { code: exit.exitCode }) }}
      </p>
      <p v-else-if="running" class="result">{{ $t('check.running') }}</p>
      <button v-if="running" :disabled="cancelling" @click="cancelAndClose">
        {{ $t('check.cancel') }}
      </button>
      <button v-else @click="close">{{ $t('check.ok') }}</button>
    </div>
  </dialog>
</template>

<style scoped>
dialog {
  width: min(56rem, 90vw);
}

.actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.result {
  margin: 0 auto 0 0;
}

.success {
  color: #26a269;
}

.failure {
  color: #c01c28;
}
</style>
