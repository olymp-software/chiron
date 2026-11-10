<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  CancelCheck,
  CourseOverview,
  EnrollCourse,
  StartCheck,
  StartWorkshopControl,
  StopRun,
  SyncCourse,
} from '../../wailsjs/go/main/App'
import type { apperr, library } from '../../wailsjs/go/models'
import ErrorView from '../errors/ErrorView.vue'
import { courseImages } from '../markdown/images'
import MarkdownView from '../markdown/MarkdownView.vue'
import CheckModal from './CheckModal.vue'

const props = defineProps<{ courseKey: string }>()

const router = useRouter()

const overview = ref<library.Overview>()
// error is a failure to load the overview, to sync or to enroll, not a course
// error.
const error = ref<apperr.Error>()
const syncing = ref(false)
const enrolling = ref(false)
const checking = ref(false)
// control is the index of the workshop control running in the modal.
const control = ref<number>()

// blocked is true while Enroll and Continue are unavailable: the course is
// broken or its environment check has not passed for the current check file.
const blocked = computed(
  () => !!overview.value?.error || (!!overview.value?.check && overview.value.check !== 'passed'),
)

async function checkClosed() {
  checking.value = false
  await load()
}

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
      <p v-if="overview.type === 'workshop'" class="mark">{{ $t('overview.workshop') }}</p>
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
        <button v-if="overview.check" :disabled="syncing" @click="checking = true">
          {{ $t('overview.checkEnvironment') }}
        </button>
        <button
          v-for="(label, i) in overview.controls ?? []"
          :key="i"
          :disabled="syncing"
          @click="control = i"
        >
          {{ label }}
        </button>
        <button
          v-if="overview.enrolled"
          :disabled="blocked || syncing"
          @click="router.push(`/courses/${courseKey}`)"
        >
          {{ overview.completed ? $t('overview.review') : $t('overview.continue') }}
        </button>
        <button v-else :disabled="blocked || syncing || enrolling" @click="enroll">
          {{ $t('overview.enroll') }}
        </button>
      </div>
      <p v-if="overview.completed" class="completed" role="status">
        {{ $t('overview.completed') }}
      </p>
      <p v-if="overview.check" class="check-status" :class="overview.check">
        {{ $t(`overview.checkStatus.${overview.check}`) }}
      </p>
      <CheckModal
        v-if="checking"
        :title="$t('check.title')"
        :start="() => StartCheck(courseKey)"
        :cancel="() => CancelCheck(courseKey)"
        @closed="checkClosed"
      />
      <CheckModal
        v-if="control !== undefined"
        :title="overview.controls?.[control] ?? ''"
        :start="() => StartWorkshopControl(courseKey, control ?? 0)"
        :cancel="StopRun"
        @closed="control = undefined"
      />
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

.mark {
  margin: 0;
  opacity: 0.7;
}

.actions {
  display: flex;
  flex-wrap: wrap;
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

.check-status,
.completed {
  margin: 0;
}

.completed {
  font-weight: bold;
}

.check-status.passed {
  color: #26a269;
}

.check-status.changed {
  color: #c01c28;
}

.description {
  white-space: pre-wrap;
}
</style>
