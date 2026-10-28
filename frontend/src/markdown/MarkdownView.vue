<script lang="ts" setup>
import { computed, ref, watchEffect } from 'vue'
import { useI18n } from 'vue-i18n'
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'
import { calloutTypes, type CalloutTitles } from './dialect'
import type { ImageResolver } from './images'
import { externalLink } from './links'
import { renderMarkdown } from './markdown'

const props = defineProps<{ source: string; resolveImage?: ImageResolver }>()

const { t } = useI18n()

const calloutTitles = computed(
  () => Object.fromEntries(calloutTypes.map((c) => [c, t(`callout.${c}`)])) as CalloutTitles,
)

const html = ref('')

watchEffect(async () => {
  html.value = await renderMarkdown(props.source, {
    resolveImage: props.resolveImage,
    calloutTitles: calloutTitles.value,
  })
})

// onClick keeps the app window in place: external links open in the
// browser, any other link does nothing.
function onClick(e: MouseEvent) {
  const a = (e.target as Element | null)?.closest('a')
  if (!a) {
    return
  }
  e.preventDefault()
  const url = externalLink(a.getAttribute('href') ?? '')
  if (url) {
    BrowserOpenURL(url)
  }
}
</script>

<template>
  <!-- renderMarkdown escapes raw HTML (markdown-it html: false). -->
  <!-- eslint-disable-next-line vue/no-v-html -->
  <div class="markdown" @click="onClick" v-html="html"></div>
</template>

<style scoped>
.markdown :deep(img) {
  max-width: 100%;
}

.markdown :deep(table) {
  border-collapse: collapse;
}

.markdown :deep(th),
.markdown :deep(td) {
  padding: 0.25rem 0.5rem;
  border: 1px solid color-mix(in srgb, CanvasText 25%, transparent);
}

.markdown :deep(.task-list-item) {
  list-style: none;
}

.markdown :deep(.callout) {
  --callout-color: #0969da;
  margin: 1rem 0;
  padding: 0.5rem 1rem;
  border-left: 4px solid var(--callout-color);
  background: color-mix(in srgb, var(--callout-color) 8%, transparent);
}

.markdown :deep(.callout-title) {
  margin: 0 0 0.25rem;
  font-weight: bold;
  color: var(--callout-color);
}

.markdown :deep(.callout > :last-child) {
  margin-bottom: 0;
}

.markdown :deep(.callout-tip) {
  --callout-color: #1a7f37;
}

.markdown :deep(.callout-warning) {
  --callout-color: #9a6700;
}

.markdown :deep(.callout-task) {
  --callout-color: #8250df;
}
</style>
