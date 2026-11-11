<script lang="ts" setup>
import { useI18n } from 'vue-i18n'
import type { Locale } from './i18n/locale'

const { locale } = useI18n()
const locales: Locale[] = ['en', 'ru']
</script>

<template>
  <div class="app">
    <header>
      <span class="app-title">{{ $t('app.title') }}</span>
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
    <!-- main scrolls, not the window, so the header stays in view; a screen
         that fills the page (the course screen) scrolls its own parts. -->
    <main>
      <div class="page">
        <!-- Keyed by the course so that switching courses reloads the screen. -->
        <RouterView v-slot="{ Component, route }">
          <component :is="Component" :key="String(route.params.key ?? '')" />
        </RouterView>
      </div>
    </main>
  </div>
</template>

<style scoped>
.app {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  height: 100vh;
}

header,
.page {
  box-sizing: border-box;
  width: 100%;
  max-width: 960px;
  margin: 0 auto;
  padding: 0 1rem;
}

header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding-top: 1rem;
}

main {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

/* A screen taller than the page overflows it and main scrolls; the course
   screen takes exactly the page height. */
.page {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding-bottom: 1rem;
}

.app-title {
  font-weight: bold;
}
</style>
