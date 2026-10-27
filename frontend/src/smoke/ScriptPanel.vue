<script lang="ts" setup>
import { ref } from 'vue'
import { SmokeRunScript } from '../../wailsjs/go/main/App'
import type { runtime } from '../../wailsjs/go/models'

// Guaranteed utilities plus one external command (hostname) to show whether
// a console window flashes from the GUI exe.
const defaultScript = `mkdir -p demo/sub
echo "hello from Chiron" > demo/sub/greeting.txt
mkdir copy
cp -r demo copy
cat copy/demo/sub/greeting.txt
touch copy/demo/sub/empty
mv copy/demo/sub/empty copy/demo/sub/moved
rm -rf demo
hostname
echo "to stderr" >&2
`

const script = ref(defaultScript)
const running = ref(false)
const result = ref<runtime.Result>()
const error = ref('')

async function run() {
  running.value = true
  result.value = undefined
  error.value = ''
  try {
    result.value = await SmokeRunScript(script.value)
  } catch (e) {
    error.value = String(e)
  } finally {
    running.value = false
  }
}
</script>

<template>
  <section class="panel">
    <h2>{{ $t('smoke.script.title') }}</h2>
    <textarea v-model="script" rows="12" spellcheck="false"></textarea>
    <button :disabled="running" @click="run">
      {{ running ? $t('smoke.script.running') : $t('smoke.script.run') }}
    </button>
    <p v-if="error" class="error">{{ $t('smoke.error', { message: error }) }}</p>
    <template v-if="result">
      <p>{{ $t('smoke.script.exitCode', { code: result.ExitCode }) }}</p>
      <h3>{{ $t('smoke.script.stdout') }}</h3>
      <pre>{{ result.Stdout }}</pre>
      <h3>{{ $t('smoke.script.stderr') }}</h3>
      <pre>{{ result.Stderr }}</pre>
    </template>
  </section>
</template>
