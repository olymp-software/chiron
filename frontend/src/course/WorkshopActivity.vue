<script lang="ts" setup>
import { computed, nextTick, onUnmounted, ref, shallowRef, watch } from 'vue'
import { StartActivityCheck, StartControl, StopRun } from '../../wailsjs/go/main/App'
import type { library } from '../../wailsjs/go/models'
import ErrorView from '../errors/ErrorView.vue'
import type { ImageResolver } from '../markdown/images'
import MarkdownView from '../markdown/MarkdownView.vue'
import CheckModal from '../overview/CheckModal.vue'
import { useCheckRun } from '../overview/checkRun'
import OutputPanel from '../scripts/OutputPanel.vue'

const props = defineProps<{
  courseKey: string
  activity: library.Activity
  workshop: library.WorkshopActivity
  resolveImage: ImageResolver
}>()

const emit = defineEmits<{ running: [running: boolean] }>()

type Run = ReturnType<typeof useCheckRun>

const showSolution = ref(false)
const checking = ref(false)
// runs is the latest run of each control, by control index.
const runs = shallowRef<(Run | undefined)[]>([])
// generations counts the starts of each control: a new start gets a new, empty
// output panel.
const generations = ref<number[]>([])
const selected = ref<number>()
const panels = new Map<number, InstanceType<typeof OutputPanel>>()

const controls = computed(() => props.workshop.controls ?? [])
const started = computed(() => controls.value.flatMap((_, i) => (runs.value[i] ? [i] : [])))
const running = computed(() => runs.value.some((r) => r?.running.value))
const shownRun = computed(() =>
  selected.value === undefined ? undefined : runs.value[selected.value],
)

watch(running, (r) => emit('running', r))

function isRunning(i: number): boolean {
  return !!runs.value[i]?.running.value
}

// blocked tells that another control of the same exclusion group is running.
function blocked(i: number): boolean {
  const group = controls.value[i].group
  return controls.value.some((c, j) => j !== i && c.group === group && isRunning(j))
}

async function start(i: number) {
  runs.value[i]?.dispose()
  panels.delete(i)
  const { courseKey, activity } = props
  const run = useCheckRun(
    () => StartControl(courseKey, activity.slug, i),
    StopRun,
    (text) => panels.get(i)?.append(text),
  )
  generations.value[i] = (generations.value[i] ?? 0) + 1
  const next = [...runs.value]
  next[i] = run
  runs.value = next
  selected.value = i
  // The panel must exist before the first output arrives.
  await nextTick()
  await run.start()
}

function setPanel(i: number, panel: unknown) {
  if (panel) panels.set(i, panel as InstanceType<typeof OutputPanel>)
}

onUnmounted(() => runs.value.forEach((r) => r?.dispose()))
</script>

<template>
  <template v-for="(b, i) in activity.theoryBlocks" :key="i">
    <MarkdownView v-if="b.kind === 'markdown'" :source="b.content" :resolve-image="resolveImage" />
  </template>
  <div v-if="workshop.solution" class="solution">
    <button v-if="!showSolution" @click="showSolution = true">{{ $t('workshop.solution') }}</button>
    <MarkdownView v-else :source="workshop.solution" :resolve-image="resolveImage" />
  </div>
  <div v-if="workshop.hasCheck || controls.length" class="controls">
    <button v-if="workshop.hasCheck" class="check" @click="checking = true">
      {{ $t('workshop.check') }}
    </button>
    <template v-for="(c, i) in controls" :key="i">
      <button
        v-if="isRunning(i)"
        class="stop"
        :disabled="runs[i]?.cancelling.value"
        @click="runs[i]?.cancel()"
      >
        {{ $t('workshop.stop', { label: c.label }) }}
      </button>
      <button v-else class="control" :disabled="blocked(i)" @click="start(i)">{{ c.label }}</button>
    </template>
  </div>
  <div v-if="started.length" class="outputs">
    <div role="tablist">
      <button
        v-for="i in started"
        :key="i"
        role="tab"
        :aria-selected="selected === i"
        @click="selected = i"
      >
        {{ controls[i].label }}
      </button>
    </div>
    <OutputPanel
      v-for="i in started"
      v-show="selected === i"
      :key="`${i}-${generations[i]}`"
      :ref="(panel) => setPanel(i, panel)"
    />
    <ErrorView v-if="shownRun?.error.value" :error="shownRun.error.value" />
    <p v-else-if="shownRun?.exit.value?.stopped" class="result">{{ $t('check.stopped') }}</p>
    <p
      v-else-if="shownRun?.exit.value"
      class="result"
      :class="shownRun.exit.value.exitCode === 0 ? 'success' : 'failure'"
    >
      {{ shownRun.exit.value.exitCode === 0 ? '✔' : '✘' }}
      {{ $t('check.exitCode', { code: shownRun.exit.value.exitCode }) }}
    </p>
    <p v-else class="result">{{ $t('check.running') }}</p>
  </div>
  <CheckModal
    v-if="checking"
    :title="$t('workshop.check')"
    :start="() => StartActivityCheck(courseKey, activity.slug)"
    :cancel="StopRun"
    @closed="checking = false"
  />
</template>

<style scoped>
.controls,
[role='tablist'] {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.outputs {
  display: grid;
  gap: 0.5rem;
}

[role='tab'][aria-selected='true'] {
  font-weight: bold;
}

.result {
  margin: 0;
}

.success {
  color: #26a269;
}

.failure {
  color: #c01c28;
}
</style>
