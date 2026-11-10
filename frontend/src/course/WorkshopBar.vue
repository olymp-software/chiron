<script lang="ts" setup>
import { ref } from 'vue'
import { OpenFolder } from '../../wailsjs/go/main/App'
import type { apperr } from '../../wailsjs/go/models'
import { ClipboardSetText } from '../../wailsjs/runtime/runtime'
import ErrorView from '../errors/ErrorView.vue'

const props = defineProps<{ courseKey: string; folder: string }>()

const copied = ref(false)
const error = ref<apperr.Error>()

async function copy() {
  copied.value = await ClipboardSetText(props.folder)
}

async function open() {
  const res = await OpenFolder(props.courseKey)
  error.value = res.error
}
</script>

<template>
  <div class="bar">
    <span>{{ $t('workshop.folder') }}</span>
    <code>{{ folder }}</code>
    <button @click="copy">{{ copied ? $t('workshop.copied') : $t('workshop.copy') }}</button>
    <button @click="open">{{ $t('workshop.openEditor') }}</button>
  </div>
  <ErrorView v-if="error" :error="error" />
</template>

<style scoped>
.bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}

code {
  flex: 1;
  min-width: 0;
  overflow-wrap: anywhere;
}
</style>
