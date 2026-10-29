<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { CourseTree as LoadTree, ListCourses } from '../../wailsjs/go/main/App'
import type { apperr, library } from '../../wailsjs/go/models'
import ErrorView from '../errors/ErrorView.vue'
import CourseTree from './CourseTree.vue'

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

function switchCourse(e: Event) {
  router.push(`/courses/${(e.target as HTMLSelectElement).value}`)
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
    <div v-if="tree" class="body">
      <CourseTree :course-key="courseKey" :tree="tree" />
      <RouterView v-slot="{ Component }">
        <component :is="Component" :tree="tree" />
      </RouterView>
    </div>
  </section>
</template>

<style scoped>
.course {
  display: grid;
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
  gap: 1rem;
  align-items: start;
}
</style>
