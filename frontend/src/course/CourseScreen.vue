<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { CourseTree as LoadTree, ListCourses } from '../../wailsjs/go/main/App'
import type { apperr, library } from '../../wailsjs/go/models'
import ErrorView from '../errors/ErrorView.vue'
import CourseTree from './CourseTree.vue'
import WorkshopBar from './WorkshopBar.vue'

const props = defineProps<{ courseKey: string }>()

const router = useRouter()

const tree = ref<library.Tree>()
const courses = ref<library.Course[]>([])
const error = ref<apperr.Error>()

const enrolled = computed(() => courses.value.filter((c) => c.enrolled))

onMounted(async () => {
  const [treeRes, listRes] = await Promise.all([LoadTree(props.courseKey), ListCourses()])
  error.value = treeRes.error ?? listRes.error
  if (!treeRes.error) {
    tree.value = treeRes.tree
  }
  courses.value = listRes.courses ?? []
})

// switchCourse opens the picked course; if the learner stays (a leave guard
// declined), the select shows the course on the screen again.
async function switchCourse(e: Event) {
  const select = e.target as HTMLSelectElement
  if (await router.push(`/courses/${select.value}`)) {
    select.value = props.courseKey
  }
}
</script>

<template>
  <section class="course">
    <header>
      <button @click="router.push('/library')">{{ $t('course.library') }}</button>
      <select
        :aria-label="$t('course.switch')"
        :value="courseKey"
        :disabled="!enrolled.length"
        @change="switchCourse"
      >
        <option v-for="c in enrolled" :key="c.key" :value="c.key">{{ c.title }}</option>
      </select>
      <button @click="router.push(`/library/${courseKey}`)">{{ $t('course.overview') }}</button>
    </header>
    <ErrorView v-if="error" :error="error" />
    <p v-else-if="!tree">{{ $t('course.loading') }}</p>
    <WorkshopBar
      v-if="tree?.workshopFolder"
      :course-key="courseKey"
      :folder="tree.workshopFolder"
    />
    <div v-if="tree" class="body">
      <CourseTree :course-key="courseKey" :tree="tree" />
      <RouterView v-slot="{ Component }">
        <component :is="Component" :tree="tree" @tree="(t: library.Tree) => (tree = t)" />
      </RouterView>
    </div>
  </section>
</template>

<style scoped>
/* The screen fills the page; the course tree and the activity scroll
   separately so the header and the other column stay in view. */
.course {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  gap: 1rem;
}

header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

select {
  flex: 1;
  min-width: 0;
}

.body {
  display: grid;
  grid-template-columns: minmax(12rem, 1fr) 3fr;
  grid-template-rows: minmax(0, 1fr);
  flex: 1;
  min-height: 0;
  gap: 1rem;
}

/* Root elements of the tree and the activity carry this scope id. They keep
   their content height, capped by the row: a stretched grid root would
   share the extra height among its rows and grow its buttons. */
.body > * {
  align-self: start;
  max-height: 100%;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
}
</style>
