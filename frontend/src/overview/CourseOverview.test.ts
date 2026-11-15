// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import {
  CancelStep,
  CheckVersions,
  CourseOverview as LoadOverview,
  RemoveCourse,
  StartWorkshopControl,
  SwitchVersion,
  SyncCourse,
} from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import en from '../i18n/en.json'
import CourseOverview from './CourseOverview.vue'

vi.mock('../../wailsjs/go/main/App', () => ({
  RemoveCourse: vi.fn(),
  CancelCheck: vi.fn(),
  CancelStep: vi.fn(),
  CheckVersions: vi.fn(),
  StopRun: vi.fn(),
  SwitchVersion: vi.fn(),
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

let router: ReturnType<typeof createRouter>

async function mountOverview() {
  router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { render: () => null } },
      { path: '/library', component: { render: () => null } },
    ],
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

function brokenCourse() {
  const res = overview(false)
  res.overview.enrolled = false
  res.overview.error = { code: 'manifestInvalid', params: {} } as never
  return res
}

it.each([
  ['a course', () => overview(false)],
  ['a workshop', () => workshop()],
  ['a course with a broken manifest', brokenCourse],
])('offers Remove for %s', async (_, res) => {
  loadOverview.mockResolvedValue(res())

  const w = await mountOverview()

  expect(w.get('.remove').text()).toBe('Remove')
})

function repoCourse(version: string, versions: string[]) {
  const res = overview(false)
  res.overview.origin = 'repo'
  res.overview.source = 'https://github.com/owner/go'
  res.overview.version = version
  res.overview.versions = versions
  return res
}

function versionOptions(w: Awaited<ReturnType<typeof mountOverview>>) {
  return w.get('select[name=version]').findAll('option')
}

it('shows the version controls of a repository course and no Sync', async () => {
  loadOverview.mockResolvedValue(repoCourse('v1.0.0', ['v1.1.0', 'v1.0.0']))

  const w = await mountOverview()

  expect(w.get<HTMLSelectElement>('select[name=version]').element.value).toBe('v1.0.0')
  expect(versionOptions(w).map((o) => o.text())).toEqual(['v1.1.0', 'v1.0.0'])
  expect(w.get('.check-versions').text()).toBe('Check for new versions')
  expect(buttons(w)).toEqual(['Continue'])
})

it('shows Sync and no version controls for a local-folder course', async () => {
  loadOverview.mockResolvedValue(overview(false))

  const w = await mountOverview()

  expect(w.find('.versions').exists()).toBe(false)
  expect(buttons(w)).toEqual(['Sync', 'Continue'])
})

it('switches the version and reloads the overview', async () => {
  loadOverview.mockResolvedValue(repoCourse('v1.0.0', ['v1.1.0', 'v1.0.0']))
  const w = await mountOverview()
  vi.mocked(SwitchVersion).mockResolvedValue(main.CourseResult.createFrom({}))
  loadOverview.mockResolvedValue(repoCourse('v1.1.0', ['v1.1.0', 'v1.0.0']))

  await w.get('select[name=version]').setValue('v1.1.0')
  await flushPromises()

  expect(SwitchVersion).toHaveBeenCalledExactlyOnceWith('go-1', 'v1.1.0')
  expect(loadOverview).toHaveBeenCalledTimes(2)
  expect(w.get<HTMLSelectElement>('select[name=version]').element.value).toBe('v1.1.0')
})

it('shows a switch error', async () => {
  loadOverview.mockResolvedValue(repoCourse('v1.0.0', ['v1.1.0', 'v1.0.0']))
  const w = await mountOverview()
  vi.mocked(SwitchVersion).mockResolvedValue(
    main.CourseResult.createFrom({ error: { code: 'gitNotFound', params: {} } }),
  )

  await w.get('select[name=version]').setValue('v1.1.0')
  await flushPromises()

  expect(w.text()).toContain(en.errors.gitNotFound)
})

it('disables the version controls while switching', async () => {
  loadOverview.mockResolvedValue(repoCourse('v1.0.0', ['v1.1.0', 'v1.0.0']))
  const w = await mountOverview()
  vi.mocked(SwitchVersion).mockReturnValue(new Promise(() => {}))

  await w.get('select[name=version]').setValue('v1.1.0')

  expect(w.get('select[name=version]').attributes('disabled')).toBeDefined()
  expect(w.get('.check-versions').attributes('disabled')).toBeDefined()
})

it.each([
  [['v1.2.0'], 'Found: v1.2.0'],
  [[], 'No new versions'],
])('checks for new versions: %j', async (found, message) => {
  loadOverview.mockResolvedValue(repoCourse('v1.0.0', ['v1.1.0', 'v1.0.0']))
  const w = await mountOverview()
  vi.mocked(CheckVersions).mockResolvedValue(main.CheckVersionsResult.createFrom({ new: found }))
  loadOverview.mockResolvedValue(repoCourse('v1.0.0', [...found, 'v1.1.0', 'v1.0.0']))

  await w.get('.check-versions').trigger('click')
  await flushPromises()

  expect(CheckVersions).toHaveBeenCalledExactlyOnceWith('go-1', expect.any(String))
  expect(w.get('.versions-result').text()).toBe(message)
  expect(versionOptions(w).map((o) => o.text())).toEqual([...found, 'v1.1.0', 'v1.0.0'])
  expect(w.get<HTMLSelectElement>('select[name=version]').element.value).toBe('v1.0.0')
})

it('shows a failed check for new versions', async () => {
  loadOverview.mockResolvedValue(repoCourse('v1.0.0', ['v1.0.0']))
  const w = await mountOverview()
  vi.mocked(CheckVersions).mockResolvedValue(
    main.CheckVersionsResult.createFrom({
      new: [],
      error: { code: 'repoUnreachable', params: {} },
    }),
  )

  await w.get('.check-versions').trigger('click')
  await flushPromises()

  expect(w.text()).toContain(en.errors.repoUnreachable)
  expect(w.find('.versions-result').exists()).toBe(false)
})

it('cancels a check for new versions and keeps the course usable meanwhile', async () => {
  loadOverview.mockResolvedValue(repoCourse('v1.0.0', ['v1.0.0']))
  const w = await mountOverview()
  let finish: (r: main.CheckVersionsResult) => void = () => {}
  vi.mocked(CheckVersions).mockReturnValue(new Promise((resolve) => (finish = resolve)))

  await w.get('.check-versions').trigger('click')

  expect(buttons(w)).toEqual(['Continue'])
  expect(w.get('.actions button').attributes('disabled')).toBeUndefined()

  await w.get('.cancel-check-versions').trigger('click')
  const stepID = vi.mocked(CheckVersions).mock.calls[0][1]
  expect(CancelStep).toHaveBeenCalledExactlyOnceWith(stepID)

  finish(main.CheckVersionsResult.createFrom({ new: [], cancelled: true }))
  await flushPromises()

  expect(w.find('.cancel-check-versions').exists()).toBe(false)
  expect(w.find('.versions-result').exists()).toBe(false)
  expect(w.find('.error').exists()).toBe(false)
})

it('stops a running check for new versions when the overview closes', async () => {
  loadOverview.mockResolvedValue(repoCourse('v1.0.0', ['v1.0.0']))
  const w = await mountOverview()
  vi.mocked(CheckVersions).mockReturnValue(new Promise(() => {}))
  await w.get('.check-versions').trigger('click')

  w.unmount()

  expect(CancelStep).toHaveBeenCalledExactlyOnceWith(vi.mocked(CheckVersions).mock.calls[0][1])
})

it('keeps the version controls with a broken manifest', async () => {
  const res = repoCourse('v1.2.0', ['v1.2.0', 'v1.0.0'])
  res.overview.enrolled = false
  res.overview.error = { code: 'manifestInvalid', params: {} } as never
  loadOverview.mockResolvedValue(res)

  const w = await mountOverview()

  expect(w.get('select[name=version]').attributes('disabled')).toBeUndefined()
  expect(w.get('.check-versions').attributes('disabled')).toBeUndefined()
  expect(w.get('.actions button').attributes('disabled')).toBeDefined()
})

async function confirmRemoval(w: Awaited<ReturnType<typeof mountOverview>>) {
  await w.get('.remove').trigger('click')
  await w.get('dialog input[name=keepProgress]').setValue(true)
  await w.get('dialog .confirm').trigger('click')
  await flushPromises()
}

it('removes and replaces the overview with the library', async () => {
  HTMLDialogElement.prototype.showModal = vi.fn()
  HTMLDialogElement.prototype.close = vi.fn()
  loadOverview.mockResolvedValue(overview(false))
  vi.mocked(RemoveCourse).mockResolvedValue(main.RemoveResult.createFrom({ left: [] }))
  const w = await mountOverview()
  const replace = vi.spyOn(router, 'replace')
  const push = vi.spyOn(router, 'push')

  await confirmRemoval(w)

  expect(RemoveCourse).toHaveBeenCalledExactlyOnceWith('go-1', true, false)
  expect(replace).toHaveBeenCalledExactlyOnceWith('/library')
  expect(push).not.toHaveBeenCalled()
})

it('shows the paths left before leaving', async () => {
  HTMLDialogElement.prototype.showModal = vi.fn()
  HTMLDialogElement.prototype.close = vi.fn()
  loadOverview.mockResolvedValue(workshop())
  vi.mocked(RemoveCourse).mockResolvedValue(
    main.RemoveResult.createFrom({ left: ['C:\\learner\\go-1'] }),
  )
  const w = await mountOverview()
  const replace = vi.spyOn(router, 'replace')

  await confirmRemoval(w)

  expect(w.get('dialog .left').text()).toBe('C:\\learner\\go-1')
  expect(replace).not.toHaveBeenCalled()

  await w.get('dialog .actions button').trigger('click')
  await flushPromises()

  expect(replace).toHaveBeenCalledExactlyOnceWith('/library')
})

it('shows a removal error and stays', async () => {
  HTMLDialogElement.prototype.showModal = vi.fn()
  HTMLDialogElement.prototype.close = vi.fn()
  loadOverview.mockResolvedValue(overview(false))
  vi.mocked(RemoveCourse).mockResolvedValue(
    main.RemoveResult.createFrom({ left: [], error: { code: 'libraryWriteFailed', params: {} } }),
  )
  const w = await mountOverview()
  const replace = vi.spyOn(router, 'replace')

  await confirmRemoval(w)

  expect(w.find('dialog').exists()).toBe(false)
  expect(w.text()).toContain(en.errors.libraryWriteFailed)
  expect(replace).not.toHaveBeenCalled()
})
