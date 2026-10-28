<script lang="ts" setup>
import { useI18n } from 'vue-i18n'
import type { apperr } from '../../wailsjs/go/models'
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'

defineProps<{ error: apperr.Error }>()

const { t } = useI18n()

const releasesURL = 'https://github.com/olymp-software/chiron/releases'

// message translates a coded error or problem; codes are defined in Go.
function message(item: { code: string; params?: Record<string, string> }): string {
  return t(`errors.${item.code}`, item.params ?? {})
}
</script>

<template>
  <div class="error" role="alert">
    <p>{{ message(error) }}</p>
    <ul v-if="error.problems?.length">
      <li v-for="(p, i) in error.problems" :key="i">
        <code v-if="p.path">{{ p.path }}</code
        ><span v-if="p.path">: </span>{{ message(p) }}
      </li>
    </ul>
    <button v-if="error.code === 'unsupportedSchemaVersion'" @click="BrowserOpenURL(releasesURL)">
      {{ t('app.releases') }}
    </button>
  </div>
</template>

<style scoped>
.error {
  color: #d1242f;
}

.error p {
  margin: 0;
}

.error ul {
  margin: 0.5rem 0 0;
  padding-left: 1.25rem;
}
</style>
