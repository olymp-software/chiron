// @vitest-environment jsdom
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter, RouterView, type Router } from 'vue-router'
import {
  Activity,
  CompleteActivity,
  StartActivityCheck,
  StartControl,
  StopActivityRuns,
  StopRun,
} from '../../wailsjs/go/main/App'
import { library, main } from '../../wailsjs/go/models'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import en from '../i18n/en.json'
import ActivityView from './ActivityView.vue'

vi.mock('../../wailsjs/go/main/App', () => ({
  Activity: vi.fn(),
  CompleteActivity: vi.fn(),
  StartActivityCheck: vi.fn(),
  StartControl: vi.fn(),
  StopRun: vi.fn(),
  StopActivityRuns: vi.fn(),
}))
vi.mock('../../wailsjs/runtime/runtime', () => ({ EventsOn: vi.fn() }))

const loadActivity = vi.mocked(Activity)
const completeActivity = vi.mocked(CompleteActivity)
const startControl = vi.mocked(StartControl)
const stopActivityRuns = vi.mocked(StopActivityRuns)

// listeners are the event subscriptions by event name.
let listeners: Map<string, Set<(data: unknown) => void>>
const emitEvent = (name: string, data: unknown) => listeners.get(name)?.forEach((l) => l(data))
const output = (runId: number, text: string) => emitEvent('check:output', { runId, text })
const exited = (runId: number) => emitEvent('check:exit', { runId, exitCode: 0, stopped: false })

const servers = [
  { label: 'One', group: '' },
  { label: 'Two', group: '' },
  { label: 'Server', group: 'server' },
]

function activity(workshop: object): library.Activity {
  return library.Activity.createFrom({
    slug: 'hello',
    title: 'Hello',
    dir: 'basics/hello',
    type: '',
    theoryBlocks: [{ kind: 'markdown', content: '# Hi' }],
    workshop: { hasCheck: false, controls: [], ...workshop },
    last: false,
    courseCompleted: false,
  })
}

function loads(workshop: object) {
  loadActivity.mockResolvedValue(main.ActivityResult.createFrom({ activity: activity(workshop) }))
}

const workshopTree = library.Tree.createFrom({
  title: 'Rollouts',
  type: 'workshop',
  current: 'hello',
  completed: false,
  sections: [],
})

let router: Router
let runId: number

beforeEach(async () => {
  vi.resetAllMocks()
  listeners = new Map()
  HTMLDialogElement.prototype.showModal = vi.fn()
  HTMLDialogElement.prototype.close = vi.fn()
  vi.stubGlobal(
    'ResizeObserver',
    class {
      observe() {}
      disconnect() {}
    },
  )
  vi.mocked(EventsOn).mockImplementation((name, callback) => {
    const set = listeners.get(name) ?? new Set()
    listeners.set(name, set)
    set.add(callback)
    return () => set.delete(callback)
  })
  runId = 0
  startControl.mockImplementation(async () => main.StartCheckResult.createFrom({ runId: ++runId }))
  vi.mocked(StopRun).mockResolvedValue(main.ErrorResult.createFrom({}))
  stopActivityRuns.mockResolvedValue(main.ErrorResult.createFrom({}))
  completeActivity.mockResolvedValue(
    main.CompleteResult.createFrom({ next: 'loops', tree: workshopTree }),
  )
  router = createRouter({
    history: createMemoryHistory(),
    routes: [
      {
        path: '/courses/:key/activities/:slug',
        component: ActivityView,
        props: (route) => ({
          courseKey: route.params.key,
          slug: route.params.slug,
          tree: workshopTree,
        }),
      },
      { path: '/library', component: { render: () => null } },
    ],
  })
  await router.push('/courses/go-1/activities/hello')
})

async function mountActivity(): Promise<VueWrapper> {
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })
  const w = mount(RouterView, { global: { plugins: [i18n, router] }, attachTo: document.body })
  await flushPromises()
  return w
}

function button(w: VueWrapper, label: string) {
  const b = w.findAll('button').find((x) => x.text() === label)
  if (!b) throw new Error(`no button ${label}`)
  return b
}

function hasButton(w: VueWrapper, label: string) {
  return w.findAll('button').some((x) => x.text() === label)
}

async function press(w: VueWrapper, label: string) {
  await button(w, label).trigger('click')
  await flushPromises()
}

// visiblePanels returns the text of the output panels that are shown.
function visiblePanels(w: VueWrapper): string[] {
  return w
    .findAll('.output')
    .filter((p) => (p.element as HTMLElement).style.display !== 'none')
    .map((p) => p.text())
}

describe('Solution', () => {
  it('is hidden until the learner presses Solution', async () => {
    loads({ solution: '# Undo it' })
    const w = await mountActivity()

    await vi.waitFor(() => expect(w.text()).toContain('Hi'))
    expect(w.text()).not.toContain('Undo it')

    await press(w, 'Solution')

    await vi.waitFor(() => expect(w.text()).toContain('Undo it'))
    expect(hasButton(w, 'Solution')).toBe(false)
  })

  it('has no button without a solution', async () => {
    loads({})

    const w = await mountActivity()

    expect(hasButton(w, 'Solution')).toBe(false)
  })
})

describe('Check', () => {
  it('has no button without check', async () => {
    loads({})

    const w = await mountActivity()

    expect(hasButton(w, 'Check')).toBe(false)
  })

  it('runs in the modal and keeps nothing after it closes', async () => {
    loads({ hasCheck: true })
    vi.mocked(StartActivityCheck).mockResolvedValue(main.StartCheckResult.createFrom({ runId: 5 }))
    const w = await mountActivity()

    await press(w, 'Check')
    output(5, 'failed\n')
    emitEvent('check:exit', { runId: 5, exitCode: 1, stopped: false })
    await flushPromises()

    expect(vi.mocked(StartActivityCheck)).toHaveBeenCalledExactlyOnceWith('go-1', 'hello')
    expect(w.get('dialog').text()).toContain('Exit code 1')

    await button(w, 'OK').trigger('click')
    await flushPromises()

    expect(w.find('dialog').exists()).toBe(false)
    expect(w.text()).not.toContain('Exit code 1')
  })
})

describe('Controls', () => {
  it('disables the controls of the running group only', async () => {
    loads({ controls: servers })
    const w = await mountActivity()

    await press(w, 'One')

    expect(hasButton(w, 'Stop: One')).toBe(true)
    expect(button(w, 'Two').attributes('disabled')).toBeDefined()
    expect(button(w, 'Server').attributes('disabled')).toBeUndefined()

    exited(1)
    await flushPromises()

    expect(button(w, 'Two').attributes('disabled')).toBeUndefined()
  })

  it('shows one output per control and switches between them', async () => {
    loads({ controls: servers })
    const w = await mountActivity()
    await press(w, 'One')
    await press(w, 'Server')

    output(1, 'from one\n')
    output(2, 'from server\n')
    await flushPromises()

    expect(visiblePanels(w)).toEqual(['from server'])

    await w.get('[role=tab]').trigger('click')

    expect(w.get('[role=tab]').text()).toBe('One')
    expect(visiblePanels(w)).toEqual(['from one'])
  })

  it('keeps the output of a finished control until it starts again', async () => {
    loads({ controls: servers })
    const w = await mountActivity()
    await press(w, 'One')
    output(1, 'first run\n')
    exited(1)
    await flushPromises()

    expect(visiblePanels(w)).toEqual(['first run'])

    await press(w, 'One')

    expect(visiblePanels(w)).toEqual([''])
    output(2, 'second run\n')
    await flushPromises()
    expect(visiblePanels(w)).toEqual(['second run'])
  })

  it('stops a running control', async () => {
    loads({ controls: servers })
    const w = await mountActivity()
    await press(w, 'One')

    await press(w, 'Stop: One')

    expect(vi.mocked(StopRun)).toHaveBeenCalledExactlyOnceWith(1)
  })

  it('stops a control pressed Stop on before it has started', async () => {
    loads({ controls: servers })
    let answer: (r: main.StartCheckResult) => void = () => {}
    startControl.mockReturnValue(new Promise((resolve) => (answer = resolve)))
    const w = await mountActivity()
    await press(w, 'One')

    await press(w, 'Stop: One')
    expect(vi.mocked(StopRun)).not.toHaveBeenCalled()

    answer(main.StartCheckResult.createFrom({ runId: 9 }))
    await flushPromises()

    expect(vi.mocked(StopRun)).toHaveBeenCalledExactlyOnceWith(9)
  })
})

describe('Leaving the activity', () => {
  it('completes without asking when nothing runs', async () => {
    loads({ controls: servers })
    const w = await mountActivity()

    await press(w, 'Next')

    expect(w.find('dialog').exists()).toBe(false)
    expect(completeActivity).toHaveBeenCalledExactlyOnceWith('go-1', 'hello')
    expect(router.currentRoute.value.params.slug).toBe('loops')
  })

  it('confirmed Next stops the scripts, completes and opens the next activity', async () => {
    loads({ controls: servers })
    const w = await mountActivity()
    await press(w, 'One')

    await press(w, 'Next')
    expect(completeActivity).not.toHaveBeenCalled()

    await press(w, 'Stop and leave')

    expect(stopActivityRuns).toHaveBeenCalledExactlyOnceWith('go-1')
    expect(completeActivity).toHaveBeenCalledExactlyOnceWith('go-1', 'hello')
    expect(router.currentRoute.value.params.slug).toBe('loops')
    expect(w.find('dialog').exists()).toBe(false)
  })

  it('declined Next stays and keeps the scripts running', async () => {
    loads({ controls: servers })
    const w = await mountActivity()
    await press(w, 'One')

    await press(w, 'Next')
    await press(w, 'Stay')

    expect(stopActivityRuns).not.toHaveBeenCalled()
    expect(completeActivity).not.toHaveBeenCalled()
    expect(router.currentRoute.value.params.slug).toBe('hello')
    expect(hasButton(w, 'Stop: One')).toBe(true)
  })

  it('asks before opening another activity', async () => {
    loads({ controls: servers })
    const w = await mountActivity()
    await press(w, 'One')

    const going = router.push('/courses/go-1/activities/maps')
    await flushPromises()
    await press(w, 'Stay')
    await going

    expect(router.currentRoute.value.params.slug).toBe('hello')
  })

  it('asks before leaving for the library', async () => {
    loads({ controls: servers })
    const w = await mountActivity()
    await press(w, 'One')

    const going = router.push('/library')
    await flushPromises()
    await press(w, 'Stop and leave')
    await going

    expect(stopActivityRuns).toHaveBeenCalledExactlyOnceWith('go-1')
    expect(router.currentRoute.value.path).toBe('/library')
  })

  it('shows Finish workshop on the last activity', async () => {
    loadActivity.mockResolvedValue(
      main.ActivityResult.createFrom({
        activity: { ...activity({}), last: true },
      }),
    )

    const w = await mountActivity()

    expect(w.get('footer button').text()).toBe('Finish workshop')
  })
})
