import { beforeEach, describe, expect, it, vi } from 'vitest'
import { CancelCheck, StartCheck } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { useCheckRun } from './checkRun'

vi.mock('../../wailsjs/go/main/App', () => ({
  StartCheck: vi.fn(),
  CancelCheck: vi.fn(),
}))
vi.mock('../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
}))

const startCheck = vi.mocked(StartCheck)
const cancelCheck = vi.mocked(CancelCheck)
const eventsOn = vi.mocked(EventsOn)

// listeners are the current event subscriptions by event name.
let listeners: Map<string, (data: unknown) => void>
const emit = (name: string, data: unknown) => listeners.get(name)?.(data)

beforeEach(() => {
  vi.resetAllMocks()
  listeners = new Map()
  eventsOn.mockImplementation((name, callback) => {
    listeners.set(name, callback)
    return () => listeners.delete(name)
  })
})

function started(runId: number) {
  startCheck.mockResolvedValue({ runId } as main.StartCheckResult)
}

describe('useCheckRun', () => {
  it('keeps only output and exit of its own run', async () => {
    started(7)
    const output: string[] = []
    const run = useCheckRun('go-1', (t) => output.push(t))

    await run.start()
    emit('check:output', { runId: 6, text: 'old run\n' })
    emit('check:output', { runId: 7, text: 'mine\n' })
    emit('check:exit', { runId: 6, exitCode: 1, stopped: false })

    expect(startCheck).toHaveBeenCalledExactlyOnceWith('go-1')
    expect(output).toEqual(['mine\n'])
    expect(run.running.value).toBe(true)

    emit('check:exit', { runId: 7, exitCode: 0, stopped: false })

    expect(run.running.value).toBe(false)
    expect(run.exit.value).toEqual({ runId: 7, exitCode: 0, stopped: false })
  })

  it('keeps events that arrive before the run id is known', async () => {
    const output: string[] = []
    startCheck.mockImplementation(async () => {
      emit('check:output', { runId: 3, text: 'early\n' })
      emit('check:output', { runId: 2, text: 'other\n' })
      emit('check:exit', { runId: 3, exitCode: 1, stopped: false })
      return { runId: 3 } as main.StartCheckResult
    })
    const run = useCheckRun('go-1', (t) => output.push(t))

    await run.start()

    expect(output).toEqual(['early\n'])
    expect(run.exit.value?.exitCode).toBe(1)
  })

  it('cancels the running check', async () => {
    started(1)
    cancelCheck.mockResolvedValue({} as main.ErrorResult)
    const run = useCheckRun('go-1', () => {})
    await run.start()

    expect(await run.cancel()).toBe(true)

    expect(cancelCheck).toHaveBeenCalledExactlyOnceWith('go-1')
    expect(run.cancelling.value).toBe(true)
    expect(run.error.value).toBeUndefined()
  })

  it('reports a failed cancel', async () => {
    started(1)
    cancelCheck.mockResolvedValue({ error: { code: 'internal' } } as main.ErrorResult)
    const run = useCheckRun('go-1', () => {})
    await run.start()

    expect(await run.cancel()).toBe(false)

    expect(run.error.value?.code).toBe('internal')
    expect(run.running.value).toBe(false)
  })

  it('can be closed after a start error', async () => {
    startCheck.mockResolvedValue({
      runId: 0,
      error: { code: 'checkRunning' },
    } as main.StartCheckResult)
    const run = useCheckRun('go-1', () => {})

    await run.start()

    expect(run.running.value).toBe(false)
    expect(run.error.value?.code).toBe('checkRunning')
    expect(listeners.size).toBe(0)
  })

  it('unsubscribes and stops the running check on dispose', async () => {
    started(1)
    const run = useCheckRun('go-1', () => {})
    await run.start()

    run.dispose()

    expect(listeners.size).toBe(0)
    expect(cancelCheck).toHaveBeenCalledExactlyOnceWith('go-1')
  })

  it('does not cancel a finished check on dispose', async () => {
    started(1)
    const run = useCheckRun('go-1', () => {})
    await run.start()
    emit('check:exit', { runId: 1, exitCode: 0, stopped: false })

    run.dispose()

    expect(cancelCheck).not.toHaveBeenCalled()
  })

  it('stops a check disposed before StartCheck returns', async () => {
    let resolve: (r: main.StartCheckResult) => void = () => {}
    startCheck.mockReturnValue(new Promise((r) => (resolve = r)))
    const run = useCheckRun('go-1', () => {})
    const starting = run.start()

    run.dispose()
    resolve({ runId: 1 } as main.StartCheckResult)
    await starting

    expect(cancelCheck).toHaveBeenCalledExactlyOnceWith('go-1')
  })
})
