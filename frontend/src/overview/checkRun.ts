import { computed, ref } from 'vue'
import { CancelCheck, StartCheck } from '../../wailsjs/go/main/App'
import type { apperr } from '../../wailsjs/go/models'
import { EventsOn } from '../../wailsjs/runtime/runtime'

// Payloads of the check:output and check:exit events (app.go).
export interface CheckOutput {
  runId: number
  text: string
}

export interface CheckExit {
  runId: number
  exitCode: number
  stopped: boolean
}

type CheckEvent = { output: CheckOutput } | { exit: CheckExit }

// useCheckRun runs the environment check of a course and passes its output to
// onOutput. Events of other runs are dropped; events that arrive before
// StartCheck returns the run id are held until it is known.
export function useCheckRun(courseKey: string, onOutput: (text: string) => void) {
  const exit = ref<CheckExit>()
  const error = ref<apperr.Error>()
  const cancelling = ref(false)
  const running = computed(() => !exit.value && !error.value)

  let runId: number | undefined
  let held: CheckEvent[] = []
  let offs: (() => void)[] = []

  function handle(e: CheckEvent) {
    if (runId === undefined) {
      held.push(e)
      return
    }
    if ('output' in e) {
      if (e.output.runId === runId) onOutput(e.output.text)
    } else if (e.exit.runId === runId) {
      exit.value = e.exit
    }
  }

  async function start() {
    offs = [
      EventsOn('check:output', (output: CheckOutput) => handle({ output })),
      EventsOn('check:exit', (e: CheckExit) => handle({ exit: e })),
    ]
    const res = await StartCheck(courseKey)
    if (res.error) {
      error.value = res.error
      dispose()
      return
    }
    runId = res.runId
    const events = held
    held = []
    events.forEach(handle)
  }

  // cancel stops the check and reports whether it stopped; CancelCheck
  // returns after the run has ended.
  async function cancel(): Promise<boolean> {
    cancelling.value = true
    const res = await CancelCheck(courseKey)
    if (res.error) {
      error.value = res.error
      return false
    }
    return true
  }

  function dispose() {
    offs.forEach((off) => off())
    offs = []
  }

  return { exit, error, cancelling, running, start, cancel, dispose }
}
