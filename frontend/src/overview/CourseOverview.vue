<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { CourseOverview, EnrollCourse, SyncCourse } from '../../wailsjs/go/main/App'
import type { apperr, library } from '../../wailsjs/go/models'
import ErrorView from '../errors/ErrorView.vue'
import { courseImages } from '../markdown/images'
import MarkdownView from '../markdown/MarkdownView.vue'

const props = defineProps<{ courseKey: string }>()

const router = useRouter()

const overview = ref<library.Overview>()
// error is a failure to load the overview, to sync or to enroll, not a course
// error.
const error = ref<apperr.Error>()
const syncing = ref(false)
const enrolling = ref(false)

async function load() {
  const res = await CourseOverview(props.courseKey)
  error.value = res.error
  if (!res.error) {
    overview.value = res.overview
  }
}

async function sync() {
  syncing.value = true
  try {
    const res = await SyncCourse(props.courseKey)
    if (res.error) {
      error.value = res.error
      return
    }
    await load()
  } finally {
    syncing.value = false
  }
}

async function enroll() {
  enrolling.value = true
  try {
    const res = await EnrollCourse(props.courseKey)
    if (res.error) {
      error.value = res.error
      return
    }
    await router.push(`/courses/${props.courseKey}`)
  } finally {
    enrolling.value = false
  }
}

onMounted(load)
</script>

<template>
  <section class="overview">
    <button class="back" @click="router.push('/library')">← {{ $t('overview.back') }}</button>
    <p v-if="!overview && !error">{{ $t('overview.loading') }}</p>
    <template v-if="overview">
      <h1>{{ overview.title }}</h1>
      <dl>
        <dt>{{ $t('overview.source') }}</dt>
        <dd>
          <code>{{ overview.source }}</code>
        </dd>
        <dt>{{ $t('overview.key') }}</dt>
        <dd>
          <code>{{ overview.key }}</code>
        </dd>
      </dl>
      <div class="actions">
        <button v-if="overview.origin === 'local'" :disabled="syncing" @click="sync">
          {{ syncing ? $t('overview.syncing') : $t('overview.sync') }}
        </button>
        <button
          v-if="overview.enrolled"
          :disabled="!!overview.error || syncing"
          @click="router.push(`/courses/${courseKey}`)"
        >
          {{ $t('overview.continue') }}
        </button>
        <button v-else :disabled="!!overview.error || syncing || enrolling" @click="enroll">
          {{ $t('overview.enroll') }}
        </button>
      </div>
    </template>
    <ErrorView v-if="error" :error="error" />
    <template v-if="overview">
      <ErrorView v-if="overview.error" :error="overview.error" />
      <MarkdownView
        v-else-if="overview.text?.markdown"
        :source="overview.text.markdown"
        :resolve-image="courseImages(overview.key)"
      />
      <p v-else-if="overview.text?.plain" class="description">{{ overview.text.plain }}</p>
    </template>
  </section>
</template>

<style scoped>
.overview {
  display: grid;
  gap: 1rem;
}

.back {
  justify-self: start;
}

h1 {
  margin: 0;
}

.actions {
  display: flex;
  gap: 0.5rem;
}

dl {
  display: grid;
  grid-template-columns: max-content 1fr;
  gap: 0.25rem 1rem;
  margin: 0;
}

dd {
  margin: 0;
  overflow-wrap: anywhere;
}

.description {
  white-space: pre-wrap;
}
</style>
