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
    <main>
      <!-- Keyed by the course so that switching courses reloads the screen. -->
      <RouterView v-slot="{ Component, route }">
        <component :is="Component" :key="String(route.params.key ?? '')" />
      </RouterView>
    </main>
  </div>
</template>

<style scoped>
.app {
  display: grid;
  gap: 1rem;
  max-width: 960px;
  margin: 0 auto;
  padding: 1rem;
}

header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.app-title {
  font-weight: bold;
}
</style>
