<script lang="ts" setup>
import { computed, nextTick, onMounted, onUnmounted, ref, shallowRef } from 'vue'
import { type Line, Terminal } from './ansi'
import { visibleRange } from './virtual'

// maxLines is how many of the latest lines the panel keeps.
const maxLines = 10_000
// lineHeight is the height of every line in pixels: lines do not wrap, so
// the panel renders only the lines in view.
const lineHeight = 18
const lineHeightPx = `${lineHeight}px`
// overscan is how many lines are rendered beyond each edge of the view.
const overscan = 20

const term = new Terminal(maxLines)
const lines = shallowRef<Line[]>([])
const dropped = ref(0)
const panel = ref<HTMLElement>()
const scrollTop = ref(0)
const viewportHeight = ref(0)

const range = computed(() =>
  visibleRange(scrollTop.value, viewportHeight.value, lineHeight, lines.value.length, overscan),
)
const shown = computed(() => lines.value.slice(range.value.start, range.value.end))

const resize = new ResizeObserver(() => {
  if (panel.value) viewportHeight.value = panel.value.clientHeight
})

onMounted(() => {
  if (panel.value) resize.observe(panel.value)
})

onUnmounted(() => resize.disconnect())

function onScroll() {
  if (panel.value) scrollTop.value = panel.value.scrollTop
}

// append adds a chunk of script output. The panel follows the bottom unless
// the learner has scrolled up.
function append(text: string) {
  const el = panel.value
  const follow = !el || el.scrollHeight - el.scrollTop - el.clientHeight < 4
  term.write(text)
  lines.value = term.lines()
  dropped.value = term.dropped
  if (follow) {
    void nextTick(() => {
      const el = panel.value
      if (!el) return
      el.scrollTop = el.scrollHeight
      scrollTop.value = el.scrollTop
    })
  }
}

defineExpose({ append })
</script>

<template>
  <div ref="panel" class="output" @scroll="onScroll">
    <div class="content" :style="{ height: `${lines.length * lineHeight}px` }">
      <pre class="window" :style="{ transform: `translateY(${range.start * lineHeight}px)` }"><div
        v-for="(line, i) in shown"
        :key="dropped + range.start + i"
        v-memo="[line]"
        class="line"
      ><span v-for="(span, j) in line" :key="j" :class="span.cls" :style="span.css">{{ span.text }}</span></div></pre>
    </div>
  </div>
</template>

<style scoped>
.output {
  --fg-0: #000;
  --fg-1: #c01c28;
  --fg-2: #26a269;
  --fg-3: #a2734c;
  --fg-4: #12488b;
  --fg-5: #a347ba;
  --fg-6: #2aa1b3;
  --fg-7: #d0cfcc;
  --fg-8: #5e5c64;
  --fg-9: #f66151;
  --fg-10: #33d17a;
  --fg-11: #e9ad0c;
  --fg-12: #2a7bde;
  --fg-13: #c061cb;
  --fg-14: #33c7de;
  --fg-15: #fff;

  height: min(24rem, 50vh);
  overflow: auto;
  border: 1px solid GrayText;
}

.window {
  width: max-content;
  min-width: 100%;
  box-sizing: border-box;
  margin: 0;
  padding: 0 0.5rem;
  line-height: v-bind(lineHeightPx);
  tab-size: 4;
}

.line {
  height: v-bind(lineHeightPx);
  white-space: pre;
}

.bold {
  font-weight: bold;
}

.dim {
  opacity: 0.7;
}

.italic {
  font-style: italic;
}

.underline {
  text-decoration: underline;
}

.fg-0 {
  color: var(--fg-0);
}
.fg-1 {
  color: var(--fg-1);
}
.fg-2 {
  color: var(--fg-2);
}
.fg-3 {
  color: var(--fg-3);
}
.fg-4 {
  color: var(--fg-4);
}
.fg-5 {
  color: var(--fg-5);
}
.fg-6 {
  color: var(--fg-6);
}
.fg-7 {
  color: var(--fg-7);
}
.fg-8 {
  color: var(--fg-8);
}
.fg-9 {
  color: var(--fg-9);
}
.fg-10 {
  color: var(--fg-10);
}
.fg-11 {
  color: var(--fg-11);
}
.fg-12 {
  color: var(--fg-12);
}
.fg-13 {
  color: var(--fg-13);
}
.fg-14 {
  color: var(--fg-14);
}
.fg-15 {
  color: var(--fg-15);
}
.bg-0 {
  background-color: var(--fg-0);
}
.bg-1 {
  background-color: var(--fg-1);
}
.bg-2 {
  background-color: var(--fg-2);
}
.bg-3 {
  background-color: var(--fg-3);
}
.bg-4 {
  background-color: var(--fg-4);
}
.bg-5 {
  background-color: var(--fg-5);
}
.bg-6 {
  background-color: var(--fg-6);
}
.bg-7 {
  background-color: var(--fg-7);
}
.bg-8 {
  background-color: var(--fg-8);
}
.bg-9 {
  background-color: var(--fg-9);
}
.bg-10 {
  background-color: var(--fg-10);
}
.bg-11 {
  background-color: var(--fg-11);
}
.bg-12 {
  background-color: var(--fg-12);
}
.bg-13 {
  background-color: var(--fg-13);
}
.bg-14 {
  background-color: var(--fg-14);
}
.bg-15 {
  background-color: var(--fg-15);
}
</style>
