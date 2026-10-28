<script lang="ts" setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Locale } from './i18n/locale'
import LibraryView from './library/LibraryView.vue'
import CourseOverview from './overview/CourseOverview.vue'

type View = { name: 'library' } | { name: 'overview'; source: string }

const view = ref<View>({ name: 'library' })

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
      <LibraryView
        v-if="view.name === 'library'"
        @open="(source) => (view = { name: 'overview', source })"
      />
      <CourseOverview
        v-else
        :key="view.source"
        :source="view.source"
        @back="view = { name: 'library' }"
      />
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
