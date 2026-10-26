<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { renderMarkdown } from '../markdown/markdown'
import sample from './sample.md?raw'

const html = ref('')
const error = ref('')

onMounted(async () => {
  try {
    html.value = await renderMarkdown(sample)
  } catch (e) {
    error.value = String(e)
  }
})
</script>

<template>
  <section class="panel">
    <h2>{{ $t('smoke.markdown.title') }}</h2>
    <p v-if="error" class="error">{{ $t('smoke.error', { message: error }) }}</p>
    <!-- renderMarkdown escapes raw HTML (markdown-it html: false). -->
    <!-- eslint-disable-next-line vue/no-v-html -->
    <div class="markdown" v-html="html"></div>
  </section>
</template>
