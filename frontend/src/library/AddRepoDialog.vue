<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { AddRepoCourse, CancelAdd, RemoteVersions } from '../../wailsjs/go/main/App'
import type { apperr } from '../../wailsjs/go/models'
import ErrorView from '../errors/ErrorView.vue'

const emit = defineEmits<{ close: [] }>()

const router = useRouter()

const dialog = ref<HTMLDialogElement>()
// step is the part of the add the dialog shows: the address, the version
// choice with the script warning, or the running clone.
const step = ref<'address' | 'version' | 'cloning'>('address')
const address = ref('')
// listing is true while the versions of the address are listed.
const listing = ref(false)
const source = ref('')
const versions = ref<string[]>([])
const version = ref('')
// error is the failure of the step the dialog shows.
const error = ref<apperr.Error>()

// stepID identifies the running listing or clone, so CancelAdd stops exactly
// that step even if it reaches Go before the step itself. A step whose id is
// no longer current was cancelled; its result is ignored.
let stepID = ''
let steps = 0

function newStep(): string {
  steps++
  stepID = `${Date.now()}-${steps}`
  return stepID
}

onMounted(() => dialog.value?.showModal())

async function next() {
  error.value = undefined
  listing.value = true
  const id = newStep()
  try {
    const res = await RemoteVersions(address.value, id)
    if (res.cancelled || id !== stepID) {
      return
    }
    if (res.error) {
      error.value = res.error
      return
    }
    if (res.remote.existing) {
      emit('close')
      await router.push(`/library/${res.remote.existing.key}`)
      return
    }
    source.value = res.remote.source
    versions.value = res.remote.versions
    version.value = versions.value[0] ?? ''
    step.value = 'version'
  } finally {
    listing.value = false
  }
}

async function add() {
  error.value = undefined
  step.value = 'cloning'
  const id = newStep()
  try {
    const res = await AddRepoCourse(source.value, version.value, id)
    if (res.cancelled || res.error) {
      error.value = res.error
      return
    }
    emit('close')
    await router.push(`/library/${res.course.key}`)
  } finally {
    // Back to the version choice after a cancel, an error or a failed call.
    if (step.value === 'cloning') {
      step.value = 'version'
    }
  }
}

function back() {
  error.value = undefined
  step.value = 'address'
}

// cancel stops a running clone and returns to the version choice once it has
// stopped; otherwise it stops a running listing and closes the dialog.
async function cancel() {
  if (step.value === 'cloning') {
    await CancelAdd(stepID)
    return
  }
  if (listing.value) {
    const id = stepID
    stepID = ''
    await CancelAdd(id)
  }
  emit('close')
}
</script>

<template>
  <dialog ref="dialog" @cancel.prevent="cancel">
    <h2>{{ $t('addRepo.title') }}</h2>
    <form v-if="step === 'address'" @submit.prevent="next">
      <label>
        {{ $t('addRepo.address') }}
        <input
          v-model.trim="address"
          name="address"
          type="text"
          autocomplete="off"
          spellcheck="false"
          :disabled="listing"
        />
      </label>
      <p class="hint">{{ $t('addRepo.addressHint') }}</p>
      <ErrorView v-if="error" :error="error" />
      <div class="actions">
        <button type="button" @click="cancel">{{ $t('addRepo.cancel') }}</button>
        <button type="submit" class="next" :disabled="!address || listing">
          {{ listing ? $t('addRepo.listing') : $t('addRepo.next') }}
        </button>
      </div>
    </form>
    <form v-else-if="step === 'version'" @submit.prevent="add">
      <p>
        <code>{{ source }}</code>
      </p>
      <label>
        {{ $t('addRepo.version') }}
        <select v-model="version" name="version">
          <option v-for="v in versions" :key="v" :value="v">{{ v }}</option>
        </select>
      </label>
      <p class="warning">{{ $t('scriptWarning.text') }}</p>
      <ErrorView v-if="error" :error="error" />
      <div class="actions">
        <button type="button" @click="back">{{ $t('addRepo.back') }}</button>
        <button type="button" @click="cancel">{{ $t('addRepo.cancel') }}</button>
        <button type="submit" class="add">{{ $t('addRepo.add') }}</button>
      </div>
    </form>
    <template v-else>
      <p role="status">{{ $t('addRepo.cloning') }}</p>
      <div class="actions">
        <button type="button" class="cancel" @click="cancel">{{ $t('addRepo.cancel') }}</button>
      </div>
    </template>
  </dialog>
</template>

<style scoped>
dialog {
  width: min(36rem, 90vw);
}

form {
  display: grid;
  gap: 0.75rem;
}

label {
  display: grid;
  gap: 0.25rem;
}

p {
  margin: 0;
}

code {
  overflow-wrap: anywhere;
}

.hint {
  opacity: 0.7;
  font-size: 0.875rem;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
