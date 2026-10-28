<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { AddLocalCourse, ListCourses, PickFolder } from '../../wailsjs/go/main/App'
import type { apperr, library } from '../../wailsjs/go/models'
import ErrorView from '../errors/ErrorView.vue'
import ScriptWarning from './ScriptWarning.vue'

const emit = defineEmits<{ open: [source: string] }>()

const { t } = useI18n()

const courses = ref<library.Entry[]>([])
const error = ref<apperr.Error>()
// pendingPath is the picked folder waiting for the script warning.
const pendingPath = ref('')
const adding = ref(false)

onMounted(async () => {
  const res = await ListCourses()
  courses.value = res.courses ?? []
  error.value = res.error
})

async function pickFolder() {
  error.value = undefined
  pendingPath.value = await PickFolder(t('library.pickFolder'))
}

async function add() {
  const path = pendingPath.value
  pendingPath.value = ''
  adding.value = true
  try {
    const res = await AddLocalCourse(path)
    if (res.error) {
      error.value = res.error
      return
    }
    emit('open', res.course.source)
  } finally {
    adding.value = false
  }
}
</script>

<template>
  <section class="library">
    <div class="heading">
      <h1>{{ $t('library.title') }}</h1>
      <button :disabled="adding" @click="pickFolder">
        {{ adding ? $t('library.adding') : $t('library.add') }}
      </button>
    </div>
    <ErrorView v-if="error" :error="error" />
    <p v-if="!courses.length && !error">{{ $t('library.empty') }}</p>
    <ul v-else>
      <li v-for="c in courses" :key="c.key">
        <button class="course" @click="emit('open', c.source)">{{ c.title }}</button>
      </li>
    </ul>
    <ScriptWarning v-if="pendingPath" @confirm="add" @cancel="pendingPath = ''" />
  </section>
</template>

<style scoped>
.library {
  display: grid;
  gap: 1rem;
}

.heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

h1 {
  margin: 0;
}

ul {
  display: grid;
  gap: 0.5rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.course {
  width: 100%;
  padding: 0.75rem 1rem;
  text-align: left;
}
</style>
