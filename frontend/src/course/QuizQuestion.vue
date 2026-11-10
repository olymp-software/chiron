<script lang="ts" setup>
import { useId } from 'vue'
import { useI18n } from 'vue-i18n'
import type { library } from '../../wailsjs/go/models'
import type { ImageResolver } from '../markdown/images'
import MarkdownView from '../markdown/MarkdownView.vue'
import type { Answer } from '../quiz/quiz'

const props = defineProps<{
  question: library.Question
  answer: Answer
  n: number
  total: number
  resolveImage: ImageResolver
}>()

const emit = defineEmits<{ select: [selected: number[]]; check: [] }>()

const { t } = useI18n()

// name groups the radio buttons of this question.
const name = useId()

function selected(i: number): boolean {
  return props.answer.selected.includes(i)
}

function toggle(i: number) {
  if (props.answer.checked) {
    return
  }
  if (!props.question.multiple) {
    emit('select', [i])
  } else if (selected(i)) {
    emit(
      'select',
      props.answer.selected.filter((s) => s !== i),
    )
  } else {
    emit(
      'select',
      [...props.answer.selected, i].sort((a, b) => a - b),
    )
  }
}

// A checked question shows two independent things per option: a green
// background on every correct option (rightOption), and a verdict icon on
// every selected one. A missed correct option is green without an icon.
function rightOption(i: number): boolean {
  return props.answer.checked && props.question.options[i].correct
}

function verdict(i: number): 'correct' | 'incorrect' | undefined {
  if (!props.answer.checked || !selected(i)) {
    return undefined
  }
  return props.question.options[i].correct ? 'correct' : 'incorrect'
}

// optionStatus tells screen readers what the colour and icon show.
function optionStatus(i: number): string {
  const v = verdict(i)
  if (v) {
    return t(v === 'correct' ? 'quiz.chosenCorrect' : 'quiz.chosenIncorrect')
  }
  return rightOption(i) ? t('quiz.missed') : ''
}
</script>

<template>
  <section class="question">
    <p class="position">{{ $t('quiz.position', { n, total }) }}</p>
    <MarkdownView v-if="question.prompt" :source="question.prompt" :resolve-image="resolveImage" />
    <ul class="options">
      <li
        v-for="(o, i) in question.options"
        :key="i"
        class="option"
        :class="{ right: rightOption(i) }"
      >
        <label>
          <input
            :type="question.multiple ? 'checkbox' : 'radio'"
            :name="name"
            :checked="selected(i)"
            :disabled="answer.checked"
            @change="toggle(i)"
          />
          <MarkdownView class="text" :source="o.text" :resolve-image="resolveImage" />
          <span v-if="verdict(i)" class="verdict" :class="verdict(i)" aria-hidden="true">
            {{ verdict(i) === 'correct' ? '✔' : '✘' }}
          </span>
          <span v-if="optionStatus(i)" class="hidden">{{ optionStatus(i) }}</span>
        </label>
        <MarkdownView
          v-if="answer.checked && selected(i) && o.feedback"
          class="feedback"
          :source="o.feedback"
          :resolve-image="resolveImage"
        />
      </li>
    </ul>
    <button v-if="!answer.checked" :disabled="!answer.selected.length" @click="emit('check')">
      {{ $t('quiz.check') }}
    </button>
    <template v-else>
      <p class="result" :class="answer.correct ? 'correct' : 'incorrect'" role="status">
        {{ $t(answer.correct ? 'quiz.correct' : 'quiz.incorrect') }}
      </p>
      <MarkdownView
        v-if="question.explanation"
        class="explanation"
        :source="question.explanation"
        :resolve-image="resolveImage"
      />
    </template>
  </section>
</template>

<style scoped>
.question {
  display: grid;
  gap: 0.75rem;
  justify-items: start;
}

.position,
.result {
  margin: 0;
}

.position {
  font-weight: bold;
}

.options {
  display: grid;
  gap: 0.5rem;
  margin: 0;
  padding: 0;
  list-style: none;
  justify-self: stretch;
}

.option {
  position: relative;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
}

.option.right {
  background: color-mix(in srgb, #26a269 20%, transparent);
}

.option label {
  display: flex;
  gap: 0.5rem;
  align-items: baseline;
}

.text :deep(p) {
  margin: 0;
}

.feedback {
  margin-left: 1.5rem;
  font-style: italic;
}

.feedback :deep(p) {
  margin: 0.25rem 0 0;
}

.verdict {
  font-weight: bold;
}

.verdict.correct,
.result.correct {
  color: #26a269;
}

.verdict.incorrect,
.result.incorrect {
  color: #c01c28;
}

.hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}

.result {
  font-weight: bold;
}
</style>
