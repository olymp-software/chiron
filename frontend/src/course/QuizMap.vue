<script lang="ts" setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Result } from '../quiz/quiz'

const props = defineProps<{ results: Result[]; open: number }>()

const emit = defineEmits<{ open: [index: number] }>()

const { t } = useI18n()

const correct = computed(() => props.results.filter((r) => r === 'correct').length)

const icons: Record<Result, string> = { unchecked: '', correct: '✔', incorrect: '✘' }

// labels are the screen-reader texts of the map, apart from the verdict
// shown under a checked question.
const labels: Record<Result, string> = {
  unchecked: 'quiz.mapUnchecked',
  correct: 'quiz.mapCorrect',
  incorrect: 'quiz.mapIncorrect',
}

function label(i: number): string {
  return t(labels[props.results[i]], { n: i + 1 })
}
</script>

<template>
  <div class="map">
    <nav :aria-label="$t('quiz.map')">
      <button
        v-for="(r, i) in results"
        :key="i"
        class="item"
        :class="r"
        :aria-label="label(i)"
        :aria-current="i === open ? 'step' : undefined"
        @click="emit('open', i)"
      >
        <span class="n">{{ i + 1 }}</span>
        <span class="icon" aria-hidden="true">{{ icons[r] }}</span>
      </button>
    </nav>
    <p class="counter" role="status">
      {{ $t('quiz.counter', { correct, total: results.length }) }}
    </p>
  </div>
</template>

<style scoped>
.map {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 1rem;
  align-items: center;
}

nav {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.item {
  display: inline-flex;
  gap: 0.125rem;
  align-items: center;
  justify-content: center;
  min-width: 2.5rem;
  border: 2px solid color-mix(in srgb, CanvasText 25%, transparent);
}

.item[aria-current] {
  border-color: CanvasText;
  font-weight: bold;
}

.item.correct {
  color: #26a269;
  border-color: #26a269;
}

.item.incorrect {
  color: #c01c28;
  border-color: #c01c28;
}

.item.correct[aria-current],
.item.incorrect[aria-current] {
  outline: 2px solid CanvasText;
  outline-offset: 1px;
}

.counter {
  margin: 0;
}
</style>
