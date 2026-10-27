<script lang="ts" setup>
import { ref } from 'vue'
import { SmokeListTags } from '../../wailsjs/go/main/App'

// A small public repository with tags.
const url = ref('https://github.com/google/uuid')
const cloning = ref(false)
const tags = ref<string[]>()
const error = ref('')

async function clone() {
  cloning.value = true
  tags.value = undefined
  error.value = ''
  try {
    tags.value = (await SmokeListTags(url.value)) ?? []
  } catch (e) {
    error.value = String(e)
  } finally {
    cloning.value = false
  }
}
</script>

<template>
  <section class="panel">
    <h2>{{ $t('smoke.git.title') }}</h2>
    <label>
      {{ $t('smoke.git.url') }}
      <input v-model="url" type="url" spellcheck="false" />
    </label>
    <button :disabled="cloning" @click="clone">
      {{ cloning ? $t('smoke.git.cloning') : $t('smoke.git.clone') }}
    </button>
    <p v-if="error" class="error">{{ $t('smoke.error', { message: error }) }}</p>
    <template v-if="tags">
      <p v-if="tags.length === 0">{{ $t('smoke.git.noTags') }}</p>
      <template v-else>
        <h3>{{ $t('smoke.git.tags', { count: tags.length }) }}</h3>
        <ul>
          <li v-for="tag in tags" :key="tag">{{ tag }}</li>
        </ul>
      </template>
    </template>
  </section>
</template>
