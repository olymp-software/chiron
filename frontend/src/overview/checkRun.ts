import { computed, ref } from 'vue'
import type { apperr, main } from '../../wailsjs/go/models'
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

// StartRun starts a script; CancelRun stops the run with the given id and
// returns after it has ended.
export type StartRun = () => Promise<main.StartCheckResult>
export type CancelRun = (runId: number) => Promise<main.ErrorResult>

// useCheckRun runs a script through startRun and passes its output to
// onOutput. Events of other runs are dropped; events that arrive before
// startRun returns the run id are held until it is known.
export function useCheckRun(
  startRun: StartRun,
  cancelRun: CancelRun,
  onOutput: (text: string) => void,
) {
  const exit = ref<CheckExit>()
  const error = ref<apperr.Error>()
  const cancelling = ref(false)
  const running = computed(() => !exit.value && !error.value)

  let runId: number | undefined
  let held: CheckEvent[] = []
  let offs: (() => void)[] = []
  let disposed = false
  // starting settles once startRun has answered, so cancel can wait for the
  // run id.
  let starting: Promise<void> = Promise.resolve()

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

  function start(): Promise<void> {
    starting = doStart()
    return starting
  }

  async function doStart() {
    if (disposed) {
      return
    }
    offs = [
      EventsOn('check:output', (output: CheckOutput) => handle({ output })),
      EventsOn('check:exit', (e: CheckExit) => handle({ exit: e })),
    ]
    const res = await startRun()
    if (res.error) {
      error.value = res.error
      dispose()
      return
    }
    runId = res.runId
    if (disposed) {
      void cancelRun(runId)
      return
    }
    const events = held
    held = []
    events.forEach(handle)
  }

  // cancel stops the script and reports whether it stopped. Pressed before
  // startRun has answered, it waits for the run id.
  async function cancel(): Promise<boolean> {
    cancelling.value = true
    await starting
    if (runId === undefined) {
      return true
    }
    const res = await cancelRun(runId)
    if (res.error) {
      error.value = res.error
      return false
    }
    return true
  }

  // dispose drops the event subscriptions and stops a script that is still
  // running, so leaving the view never leaves a script nobody can cancel.
  function dispose() {
    disposed = true
    offs.forEach((off) => off())
    offs = []
    if (runId !== undefined && running.value && !cancelling.value) {
      void cancelRun(runId)
    }
  }

  return { exit, error, cancelling, running, start, cancel, dispose }
}
