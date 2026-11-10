// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import {
  CourseOverview as LoadOverview,
  StartWorkshopControl,
  SyncCourse,
} from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import en from '../i18n/en.json'
import CourseOverview from './CourseOverview.vue'

vi.mock('../../wailsjs/go/main/App', () => ({
  CancelCheck: vi.fn(),
  StopRun: vi.fn(),
  CourseOverview: vi.fn(),
  EnrollCourse: vi.fn(),
  StartCheck: vi.fn(),
  StartWorkshopControl: vi.fn(),
  SyncCourse: vi.fn(),
}))
vi.mock('../../wailsjs/runtime/runtime', () => ({ EventsOn: vi.fn() }))

const loadOverview = vi.mocked(LoadOverview)

function overview(completed: boolean) {
  return main.OverviewResult.createFrom({
    overview: {
      origin: 'local',
      source: '/src/go',
      key: 'go-1',
      title: 'Go',
      enrolled: true,
      completed,
      text: {},
      check: '',
    },
  })
}

beforeEach(() => {
  vi.resetAllMocks()
})

async function mountOverview() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/', component: { render: () => null } }],
  })
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })
  const w = mount(CourseOverview, {
    props: { courseKey: 'go-1' },
    global: { plugins: [i18n, router] },
  })
  await flushPromises()
  return w
}

function buttons(w: Awaited<ReturnType<typeof mountOverview>>): string[] {
  return w.findAll('.actions button').map((b) => b.text())
}

it('shows a completed course with Review', async () => {
  loadOverview.mockResolvedValue(overview(true))

  const w = await mountOverview()

  expect(w.get('[role=status]').text()).toBe('Course completed')
  expect(buttons(w)).toEqual(['Sync', 'Review'])
})

it('shows an enrolled course with Continue', async () => {
  loadOverview.mockResolvedValue(overview(false))

  const w = await mountOverview()

  expect(w.find('[role=status]').exists()).toBe(false)
  expect(buttons(w)).toEqual(['Sync', 'Continue'])
})

function workshop(controls?: string[]) {
  const res = overview(false)
  res.overview.type = 'workshop'
  res.overview.controls = controls
  return res
}

it('marks a workshop', async () => {
  loadOverview.mockResolvedValue(workshop())

  const w = await mountOverview()

  expect(w.get('.mark').text()).toBe('Workshop')
})

it('does not mark a course', async () => {
  loadOverview.mockResolvedValue(overview(false))

  const w = await mountOverview()

  expect(w.find('.mark').exists()).toBe(false)
})

it('runs a workshop control in the modal and keeps nothing after it closes', async () => {
  HTMLDialogElement.prototype.showModal = vi.fn()
  HTMLDialogElement.prototype.close = vi.fn()
  vi.stubGlobal(
    'ResizeObserver',
    class {
      observe() {}
      disconnect() {}
    },
  )
  const handlers = new Map<string, (data: unknown) => void>()
  vi.mocked(EventsOn).mockImplementation((name, callback) => {
    handlers.set(name, callback)
    return () => handlers.delete(name)
  })
  vi.mocked(StartWorkshopControl).mockResolvedValue(main.StartCheckResult.createFrom({ runId: 4 }))
  loadOverview.mockResolvedValue(workshop(['Check leftovers']))
  const w = await mountOverview()

  await w.findAll('.actions button')[1].trigger('click')
  await flushPromises()
  handlers.get('check:exit')?.({ runId: 4, exitCode: 0, stopped: false })
  await flushPromises()

  expect(StartWorkshopControl).toHaveBeenCalledExactlyOnceWith('go-1', 0)
  expect(w.get('dialog h2').text()).toBe('Check leftovers')

  await w.get('dialog .actions button').trigger('click')
  await flushPromises()

  expect(w.find('dialog').exists()).toBe(false)
})

it('updates the completed status after Sync', async () => {
  loadOverview.mockResolvedValue(overview(true))
  const w = await mountOverview()
  vi.mocked(SyncCourse).mockResolvedValue(main.CourseResult.createFrom({}))
  loadOverview.mockResolvedValue(overview(false))

  await w.get('.actions button').trigger('click')
  await flushPromises()

  expect(w.find('[role=status]').exists()).toBe(false)
  expect(buttons(w)).toEqual(['Sync', 'Continue'])
})
