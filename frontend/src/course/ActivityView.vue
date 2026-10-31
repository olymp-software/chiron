<script lang="ts" setup>
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Activity as LoadActivity, CompleteActivity } from '../../wailsjs/go/main/App'
import type { apperr, library } from '../../wailsjs/go/models'
import ErrorView from '../errors/ErrorView.vue'
import { activityImages } from '../markdown/images'
import MarkdownView from '../markdown/MarkdownView.vue'

const props = defineProps<{ courseKey: string; slug: string; tree: library.Tree }>()

const emit = defineEmits<{ tree: [tree: library.Tree] }>()

const router = useRouter()

const activity = ref<library.Activity>()
const error = ref<apperr.Error>()
const completing = ref(false)
const completeError = ref<apperr.Error>()

// loads numbers activity loads, so a slow answer for an activity the learner
// has already left is dropped.
let loads = 0

watch(
  () => [props.courseKey, props.slug],
  async () => {
    const load = ++loads
    activity.value = undefined
    error.value = undefined
    completeError.value = undefined
    const res = await LoadActivity(props.courseKey, props.slug)
    if (load !== loads) {
      return
    }
    // A locked activity or one that is not in the course opens as the current
    // activity; the app redirect replaces the requested entry.
    if (res.error?.code === 'activityUnavailable' && props.tree.current !== props.slug) {
      router.replace(`/courses/${props.courseKey}/activities/${props.tree.current}`)
      return
    }
    if (res.error) {
      error.value = res.error
      return
    }
    activity.value = res.activity
  },
  { immediate: true },
)

// complete is Next and Finish course: Go completes the activity and returns
// the activity to open next, "" after the last one.
async function complete() {
  const a = activity.value
  if (!a || completing.value) {
    return
  }
  completing.value = true
  completeError.value = undefined
  try {
    const res = await CompleteActivity(props.courseKey, a.slug)
    if (res.error) {
      completeError.value = res.error
      return
    }
    emit('tree', res.tree)
    if (res.next) {
      router.push(`/courses/${props.courseKey}/activities/${res.next}`)
    } else {
      a.completed = true
      a.courseCompleted = res.tree.completed
    }
  } finally {
    completing.value = false
  }
}
</script>

<template>
  <ErrorView v-if="error" :error="error" />
  <p v-else-if="!activity">{{ $t('course.loading') }}</p>
  <article v-else class="activity">
    <h1>{{ activity.title }}</h1>
    <ErrorView v-if="activity.error" :error="activity.error" />
    <template v-else>
      <template v-for="(b, i) in activity.theoryBlocks" :key="i">
        <MarkdownView
          v-if="b.kind === 'markdown'"
          :source="b.content"
          :resolve-image="activityImages(courseKey, activity.dir)"
        />
      </template>
    </template>
    <footer>
      <p v-if="activity.last && activity.courseCompleted" class="completed" role="status">
        {{ $t('course.completed') }}
      </p>
      <button v-else :disabled="completing" @click="complete">
        {{ activity.last ? $t('course.finish') : $t('course.next') }}
      </button>
      <ErrorView v-if="completeError" :error="completeError" />
    </footer>
  </article>
</template>

<style scoped>
.activity {
  display: grid;
  gap: 1rem;
}

h1 {
  margin: 0;
}

footer {
  display: grid;
  justify-items: end;
  gap: 0.5rem;
}

.completed {
  margin: 0;
  font-weight: bold;
}
</style>
