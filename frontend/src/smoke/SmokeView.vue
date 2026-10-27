<script lang="ts" setup>
import { useI18n } from 'vue-i18n'
import type { Locale } from '../i18n/locale'
import GitPanel from './GitPanel.vue'
import MarkdownPanel from './MarkdownPanel.vue'
import ScriptPanel from './ScriptPanel.vue'

const { locale } = useI18n()
const locales: Locale[] = ['en', 'ru']
</script>

<template>
  <div class="smoke">
    <header>
      <h1>{{ $t('app.title') }} · {{ $t('smoke.title') }}</h1>
      <div role="group" :aria-label="$t('locale.label')">
        <button
          v-for="l in locales"
          :key="l"
          :aria-pressed="locale === l"
          :disabled="locale === l"
          @click="locale = l"
        >
          {{ $t(`locale.${l}`) }}
        </button>
      </div>
    </header>
    <ScriptPanel />
    <GitPanel />
    <MarkdownPanel />
  </div>
</template>

<style>
.smoke {
  display: grid;
  gap: 1rem;
  max-width: 960px;
  margin: 0 auto;
  padding: 1rem;
  text-align: left;
}

.smoke header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.smoke .panel {
  display: grid;
  gap: 0.5rem;
  padding: 1rem;
  border: 1px solid color-mix(in srgb, CanvasText 20%, transparent);
  border-radius: 8px;
}

.smoke textarea,
.smoke input,
.smoke pre {
  font-family: ui-monospace, Consolas, monospace;
}

.smoke label {
  display: grid;
  gap: 0.25rem;
}

.smoke pre {
  margin: 0;
  padding: 0.5rem;
  overflow-x: auto;
  white-space: pre-wrap;
}

.smoke .error {
  color: #d1242f;
}

.smoke button {
  justify-self: start;
}

/* Shiki dual themes: light colors inline, dark ones in --shiki-dark. */
@media (prefers-color-scheme: dark) {
  .shiki,
  .shiki span {
    color: var(--shiki-dark) !important;
    background-color: var(--shiki-dark-bg) !important;
  }
}
</style>
