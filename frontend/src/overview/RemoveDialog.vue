<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { RemoveCourse } from '../../wailsjs/go/main/App'
import type { apperr } from '../../wailsjs/go/models'

const props = defineProps<{ courseKey: string; title: string; workshop: boolean }>()
// removed: the library item is gone (after OK when files were left);
// failed: nothing was removed.
const emit = defineEmits<{ cancel: []; removed: []; failed: [error: apperr.Error] }>()

const dialog = ref<HTMLDialogElement>()
const keepProgress = ref(false)
const keepFiles = ref(false)
const removing = ref(false)
// left holds the paths that could not be deleted, once removal is done.
const left = ref<string[]>()

const variant = computed(() => (props.workshop ? 'workshop' : 'course'))

function close() {
  dialog.value?.close()
}

function cancel() {
  if (removing.value || left.value) return
  close()
  emit('cancel')
}

async function remove() {
  removing.value = true
  try {
    const res = await RemoveCourse(props.courseKey, keepProgress.value, keepFiles.value)
    if (res.error) {
      close()
      emit('failed', res.error)
      return
    }
    if (res.left?.length) {
      left.value = res.left
      return
    }
    close()
    emit('removed')
  } finally {
    removing.value = false
  }
}

function ok() {
  close()
  emit('removed')
}

onMounted(() => dialog.value?.showModal())
</script>

<template>
  <dialog ref="dialog" @cancel.prevent="cancel">
    <template v-if="left">
      <h2>{{ $t('remove.leftTitle') }}</h2>
      <p>{{ $t(`remove.leftText.${variant}`) }}</p>
      <ul class="left">
        <li v-for="path in left" :key="path">
          <code>{{ path }}</code>
        </li>
      </ul>
      <div class="actions">
        <button @click="ok">{{ $t('remove.ok') }}</button>
      </div>
    </template>
    <template v-else>
      <h2>{{ $t(`remove.title.${variant}`, { title }) }}</h2>
      <p>{{ $t(`remove.text.${variant}`) }}</p>
      <label>
        <input v-model="keepProgress" type="checkbox" name="keepProgress" :disabled="removing" />
        {{ $t('remove.keepProgress') }}
      </label>
      <label>
        <input v-model="keepFiles" type="checkbox" name="keepFiles" :disabled="removing" />
        {{ $t(`remove.keepFiles.${variant}`) }}
      </label>
      <div class="actions">
        <button class="cancel" :disabled="removing" @click="cancel">
          {{ $t('remove.cancel') }}
        </button>
        <button class="confirm" :disabled="removing" @click="remove">
          {{ removing ? $t('remove.removing') : $t('remove.confirm') }}
        </button>
      </div>
    </template>
  </dialog>
</template>

<style scoped>
dialog {
  max-width: 36rem;
}

label {
  display: block;
  margin: 0.25rem 0;
}

.left {
  overflow-wrap: anywhere;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 1rem;
}
</style>
