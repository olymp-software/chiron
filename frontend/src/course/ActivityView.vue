<script lang="ts" setup>
import { computed, watchEffect } from 'vue'
import { useRouter } from 'vue-router'
import type { library } from '../../wailsjs/go/models'

const props = defineProps<{ courseKey: string; slug: string; tree: library.Tree }>()

const router = useRouter()

const activity = computed(() =>
  props.tree.sections.flatMap((s) => s.activities).find((a) => a.slug === props.slug),
)

// An activity that is not in the course opens as the current activity; the
// app redirect replaces the requested entry.
watchEffect(() => {
  if (!activity.value && props.tree.current !== props.slug) {
    router.replace(`/courses/${props.courseKey}/activities/${props.tree.current}`)
  }
})
</script>

<template>
  <article v-if="activity" class="activity">
    <h1>{{ activity.title }}</h1>
  </article>
</template>

<style scoped>
h1 {
  margin: 0;
}
</style>
