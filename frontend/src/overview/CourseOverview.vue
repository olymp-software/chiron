<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import {
  CancelCheck,
  CheckVersions,
  CourseOverview,
  EnrollCourse,
  StartCheck,
  StartWorkshopControl,
  StopRun,
  SwitchVersion,
  SyncCourse,
} from '../../wailsjs/go/main/App'
import type { apperr, library } from '../../wailsjs/go/models'
import ErrorView from '../errors/ErrorView.vue'
import { courseImages } from '../markdown/images'
import MarkdownView from '../markdown/MarkdownView.vue'
import CheckModal from './CheckModal.vue'
import RemoveDialog from './RemoveDialog.vue'

const props = defineProps<{ courseKey: string }>()

const router = useRouter()
const { t } = useI18n()

const overview = ref<library.Overview>()
// error is a failure to load the overview, to sync, to switch or check
// versions or to enroll, not a course error.
const error = ref<apperr.Error>()
const syncing = ref(false)
const switching = ref(false)
const checkingVersions = ref(false)
// versionsResult tells what the last check for new versions found.
const versionsResult = ref('')
// busy is true while the course copy changes or is about to.
const busy = computed(() => syncing.value || switching.value)
const enrolling = ref(false)
const checking = ref(false)
// control is the index of the workshop control running in the modal.
const control = ref<number>()
// removing is true while the removal dialog is open.
const removing = ref(false)

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

async function switchVersion(version: string) {
  error.value = undefined
  versionsResult.value = ''
  switching.value = true
  try {
    const res = await SwitchVersion(props.courseKey, version)
    // A failed switch may still have changed the clone, so the overview is
    // read again either way.
    await load()
    if (res.error) {
      error.value = res.error
    }
  } finally {
    switching.value = false
  }
}

async function checkVersions() {
  error.value = undefined
  versionsResult.value = ''
  checkingVersions.value = true
  try {
    const res = await CheckVersions(props.courseKey)
    if (res.error) {
      error.value = res.error
      return
    }
    versionsResult.value = res.new?.length
      ? t('overview.newVersions', { versions: res.new.join(', ') })
      : t('overview.noNewVersions')
    await load()
  } finally {
    checkingVersions.value = false
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

function removeFailed(e: apperr.Error) {
  removing.value = false
  error.value = e
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
      <div v-if="overview.origin === 'repo'" class="versions">
        <label>
          {{ $t('overview.version') }}
          <select
            name="version"
            :value="overview.version"
            :disabled="switching || checkingVersions || enrolling"
            @change="switchVersion(($event.target as HTMLSelectElement).value)"
          >
            <option v-if="!overview.version" value="" disabled>
              {{ $t('overview.noVersion') }}
            </option>
            <option v-for="v in overview.versions ?? []" :key="v" :value="v">{{ v }}</option>
          </select>
        </label>
        <span v-if="switching">{{ $t('overview.switching') }}</span>
        <button
          class="check-versions"
          :disabled="switching || checkingVersions || enrolling"
          @click="checkVersions"
        >
          {{ checkingVersions ? $t('overview.checkingVersions') : $t('overview.checkVersions') }}
        </button>
        <span v-if="versionsResult" class="versions-result" aria-live="polite">
          {{ versionsResult }}
        </span>
      </div>
      <div class="actions">
        <button v-if="overview.origin === 'local'" :disabled="syncing" @click="sync">
          {{ syncing ? $t('overview.syncing') : $t('overview.sync') }}
        </button>
        <button v-if="overview.check" :disabled="busy" @click="checking = true">
          {{ $t('overview.checkEnvironment') }}
        </button>
        <button
          v-for="(label, i) in overview.controls ?? []"
          :key="i"
          :disabled="busy"
          @click="control = i"
        >
          {{ label }}
        </button>
        <button
          v-if="overview.enrolled"
          :disabled="blocked || busy"
          @click="router.push(`/courses/${courseKey}`)"
        >
          {{ overview.completed ? $t('overview.review') : $t('overview.continue') }}
        </button>
        <button v-else :disabled="blocked || busy || enrolling" @click="enroll">
          {{ $t('overview.enroll') }}
        </button>
      </div>
      <div class="removal">
        <button class="remove" :disabled="busy || enrolling" @click="removing = true">
          {{ $t('remove.button') }}
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
      <RemoveDialog
        v-if="removing"
        :course-key="courseKey"
        :title="overview.title"
        :workshop="overview.type === 'workshop'"
        @cancel="removing = false"
        @failed="removeFailed"
        @removed="router.replace('/library')"
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

.actions,
.versions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.versions {
  align-items: center;
}

.versions label {
  display: flex;
  align-items: center;
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

.removal {
  display: flex;
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
