<script lang="ts" setup>
import type { library } from '../../wailsjs/go/models'

defineProps<{ courseKey: string; tree: library.Tree }>()
</script>

<template>
  <nav class="tree" :aria-label="tree.title">
    <section v-for="s in tree.sections" :key="s.slug">
      <h2>{{ s.title }}</h2>
      <ul>
        <li v-for="a in s.activities" :key="a.slug">
          <!-- The open activity gets router-link-exact-active and aria-current. -->
          <RouterLink :to="`/courses/${courseKey}/activities/${a.slug}`">{{ a.title }}</RouterLink>
        </li>
      </ul>
    </section>
  </nav>
</template>

<style scoped>
.tree {
  display: grid;
  gap: 1rem;
}

h2 {
  margin: 0 0 0.25rem;
  font-size: 1rem;
}

ul {
  display: grid;
  gap: 0.25rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

a {
  display: block;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  color: inherit;
  text-decoration: none;
}

a.router-link-exact-active {
  background: rgba(127, 127, 127, 0.25);
  font-weight: bold;
}
</style>
