<script lang="ts" setup>
import { computed, nextTick, ref, watch } from 'vue'
import type { library } from '../../wailsjs/go/models'
import type { ImageResolver } from '../markdown/images'
import MarkdownView from '../markdown/MarkdownView.vue'
import { freshAnswers, isCorrect, result, type Answer } from '../quiz/quiz'
import QuizMap from './QuizMap.vue'
import QuizQuestion from './QuizQuestion.vue'

const props = defineProps<{
  theoryBlocks: library.TheoryBlock[]
  quiz: library.Quiz
  resolveImage: ImageResolver
}>()

// The whole quiz state lives here, so switching screens keeps it; a new
// component, mounted for every opened activity, starts fresh.
const screen = ref<'theory' | 'quiz'>('theory')
const started = ref(false)
const open = ref(0)
const answers = ref<Answer[]>(freshAnswers(props.quiz))

const questions = computed(() => props.quiz.questions)
const results = computed(() => answers.value.map(result))

// shown is the element of the screen on display.
const shown = ref<HTMLElement>()

// A new screen starts at its top with focus on it: the button that switched
// screens is gone, and the old scroll offset belongs to the other screen.
watch(screen, async () => {
  await nextTick()
  shown.value?.scrollIntoView?.({ block: 'start' })
  shown.value?.focus({ preventScroll: true })
})

function showQuiz() {
  started.value = true
  screen.value = 'quiz'
}

function select(selected: number[]) {
  const a = answers.value[open.value]
  if (!a.checked) {
    a.selected = selected
  }
}

function check() {
  const a = answers.value[open.value]
  if (a.checked || !a.selected.length) {
    return
  }
  a.correct = isCorrect(questions.value[open.value], a.selected)
  a.checked = true
}

function startOver() {
  answers.value = freshAnswers(props.quiz)
  open.value = 0
}
</script>

<template>
  <div v-if="screen === 'theory'" ref="shown" class="theory" tabindex="-1">
    <template v-for="(b, i) in theoryBlocks" :key="i">
      <MarkdownView
        v-if="b.kind === 'markdown'"
        :source="b.content"
        :resolve-image="resolveImage"
      />
    </template>
    <button class="start" @click="showQuiz">
      {{ $t(started ? 'quiz.back' : 'quiz.start') }}
    </button>
  </div>
  <div v-else ref="shown" class="quiz" tabindex="-1">
    <div class="bar">
      <QuizMap :results="results" :open="open" @open="(i: number) => (open = i)" />
      <div class="actions">
        <button class="to-theory" @click="screen = 'theory'">{{ $t('quiz.theory') }}</button>
        <button class="start-over" @click="startOver">{{ $t('quiz.startOver') }}</button>
      </div>
    </div>
    <QuizQuestion
      :key="open"
      :question="questions[open]"
      :answer="answers[open]"
      :n="open + 1"
      :total="questions.length"
      :resolve-image="resolveImage"
      @select="select"
      @check="check"
    />
    <div class="nav">
      <button class="previous" :disabled="open === 0" @click="open--">
        {{ $t('quiz.previous') }}
      </button>
      <button class="next-question" :disabled="open === questions.length - 1" @click="open++">
        {{ $t('quiz.nextQuestion') }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.theory,
.quiz {
  display: grid;
  gap: 1rem;
  scroll-margin-top: 1rem;
}

.theory:focus,
.quiz:focus {
  outline: none;
}

.theory .start {
  justify-self: start;
}

.bar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 1rem;
  align-items: center;
  justify-content: space-between;
}

.actions,
.nav {
  display: flex;
  gap: 0.5rem;
}

.nav {
  justify-content: space-between;
}
</style>
