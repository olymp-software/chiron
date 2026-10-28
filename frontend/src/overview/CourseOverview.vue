<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { CourseOverview, SyncCourse } from '../../wailsjs/go/main/App'
import type { apperr, library } from '../../wailsjs/go/models'
import ErrorView from '../errors/ErrorView.vue'
import { courseImages } from '../markdown/images'
import MarkdownView from '../markdown/MarkdownView.vue'

const props = defineProps<{ source: string }>()
const emit = defineEmits<{ back: [] }>()

const overview = ref<library.Overview>()
// error is a failure to load the overview or to sync, not a course error.
const error = ref<apperr.Error>()
const syncing = ref(false)

async function load() {
  const res = await CourseOverview(props.source)
  error.value = res.error
  if (!res.error) {
    overview.value = res.overview
  }
}

async function sync() {
  syncing.value = true
  try {
    const res = await SyncCourse(props.source)
    if (res.error) {
      error.value = res.error
      return
    }
    await load()
  } finally {
    syncing.value = false
  }
}

onMounted(load)
</script>

<template>
  <section class="overview">
    <button class="back" @click="emit('back')">← {{ $t('overview.back') }}</button>
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
      <div v-if="overview.origin === 'local'">
        <button :disabled="syncing" @click="sync">
          {{ syncing ? $t('overview.syncing') : $t('overview.sync') }}
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
